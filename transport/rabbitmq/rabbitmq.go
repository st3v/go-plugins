package rabbitmq

import (
	"fmt"
	"io"
	"sync"
	"time"

	"errors"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/streadway/amqp"

	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/transport"
)

const (
	directReplyQueue = "amq.rabbitmq.reply-to"
)

type rmqtport struct {
	conn  *rabbitMQConn
	addrs []string
	opts  transport.Options

	once    sync.Once
	replyTo string

	sync.Mutex
	inflight map[string]chan amqp.Delivery
}

type rmqtportClient struct {
	rt    *rmqtport
	addr  string
	corId string
	reply chan amqp.Delivery
}

type rmqtportSocket struct {
	conn  *rabbitMQConn
	d     *amqp.Delivery
	close chan bool
	once  sync.Once
	sync.Mutex
	r  chan *amqp.Delivery
	bl []*amqp.Delivery
}

type rmqtportListener struct {
	conn *rabbitMQConn
	addr string

	sync.RWMutex
	so map[string]*rmqtportSocket
}

func init() {
	cmd.DefaultTransports["rabbitmq"] = NewTransport
}

func (r *rmqtportClient) Send(m *transport.Message) error {
	if !r.rt.conn.IsConnected() {
		return errors.New("Not connected to AMQP")
	}

	headers := amqp.Table{}
	for k, v := range m.Header {
		headers[k] = v
	}

	message := amqp.Publishing{
		CorrelationId: r.corId,
		Timestamp:     time.Now().UTC(),
		Body:          m.Body,
		ReplyTo:       r.rt.replyTo,
		Headers:       headers,
	}

	if err := r.rt.conn.Publish(DefaultExchange, r.addr, message); err != nil {
		return err
	}

	return nil
}

func (r *rmqtportClient) Recv(m *transport.Message) error {
	select {
	case d := <-r.reply:
		mr := &transport.Message{
			Header: make(map[string]string),
			Body:   d.Body,
		}

		for k, v := range d.Headers {
			mr.Header[k] = fmt.Sprintf("%v", v)
		}

		*m = *mr
		return nil
	case <-time.After(time.Second * 10):
		return errors.New("timed out")
	}
}

func (r *rmqtportClient) Close() error {
	r.rt.popReq(r.corId)
	return nil
}

func (r *rmqtportSocket) Recv(m *transport.Message) error {
	if m == nil {
		return errors.New("message passed in is nil")
	}

	d, ok := <-r.r
	if !ok {
		return io.EOF
	}

	r.Lock()
	if len(r.bl) > 0 {
		select {
		case r.r <- r.bl[0]:
			r.bl = r.bl[1:]
		default:
		}
	}
	r.Unlock()

	mr := &transport.Message{
		Header: make(map[string]string),
		Body:   d.Body,
	}

	for k, v := range d.Headers {
		mr.Header[k] = fmt.Sprintf("%v", v)
	}

	*m = *mr
	return nil
}

func (r *rmqtportSocket) Send(m *transport.Message) error {
	msg := amqp.Publishing{
		CorrelationId: r.d.CorrelationId,
		Timestamp:     time.Now().UTC(),
		Body:          m.Body,
		Headers:       amqp.Table{},
	}

	for k, v := range m.Header {
		msg.Headers[k] = v
	}
	return r.conn.Publish("", r.d.ReplyTo, msg)
}

func (r *rmqtportSocket) Close() error {
	r.once.Do(func() {
		close(r.close)
	})
	return nil
}

func (r *rmqtportListener) Addr() string {
	return r.addr
}

func (r *rmqtportListener) Close() error {
	r.conn.Close()
	return nil
}

func (r *rmqtportListener) Accept(fn func(transport.Socket)) error {
	deliveries, err := r.conn.Consume(r.addr)
	if err != nil {
		return err
	}

	for d := range deliveries {
		r.RLock()
		sock, ok := r.so[d.CorrelationId]
		r.RUnlock()
		if !ok {
			var once sync.Once
			sock = &rmqtportSocket{
				d:     &d,
				r:     make(chan *amqp.Delivery, 1),
				conn:  r.conn,
				once:  once,
				close: make(chan bool, 1),
			}
			r.Lock()
			r.so[sock.d.CorrelationId] = sock
			r.Unlock()

			go func() {
				<-sock.close
				r.Lock()
				delete(r.so, sock.d.CorrelationId)
				r.Unlock()
			}()

			go fn(sock)
		}

		select {
		case <-sock.close:
			continue
		default:
		}

		sock.Lock()
		sock.bl = append(sock.bl, &d)
		select {
		case sock.r <- sock.bl[0]:
			sock.bl = sock.bl[1:]
		default:
		}
		sock.Unlock()
	}

	return nil
}

func (r *rmqtport) putReq(id string) chan amqp.Delivery {
	r.Lock()
	ch := make(chan amqp.Delivery, 1)
	r.inflight[id] = ch
	r.Unlock()
	return ch
}

func (r *rmqtport) getReq(id string) chan amqp.Delivery {
	r.Lock()
	defer r.Unlock()
	if ch, ok := r.inflight[id]; ok {
		return ch
	}
	return nil
}

func (r *rmqtport) popReq(id string) {
	r.Lock()
	defer r.Unlock()
	if _, ok := r.inflight[id]; ok {
		delete(r.inflight, id)
	}
}

func (r *rmqtport) init() {
	<-r.conn.Init(r.opts.Secure, r.opts.TLSConfig)
	if err := r.conn.Channel.DeclareReplyQueue(r.replyTo); err != nil {
		return
	}
	deliveries, err := r.conn.Channel.ConsumeQueue(r.replyTo)
	if err != nil {
		return
	}
	go func() {
		for delivery := range deliveries {
			go r.handle(delivery)
		}
	}()
}

func (r *rmqtport) handle(delivery amqp.Delivery) {
	ch := r.getReq(delivery.CorrelationId)
	if ch == nil {
		return
	}
	ch <- delivery
}

func (r *rmqtport) Dial(addr string, opts ...transport.DialOption) (transport.Client, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	r.once.Do(r.init)

	return &rmqtportClient{
		rt:    r,
		addr:  addr,
		corId: id.String(),
		reply: r.putReq(id.String()),
	}, nil
}

func (r *rmqtport) Listen(addr string, opts ...transport.ListenOption) (transport.Listener, error) {
	if len(addr) == 0 || addr == ":0" {
		id, err := uuid.NewV4()
		if err != nil {
			return nil, err
		}
		addr = id.String()
	}

	conn := newRabbitMQConn("", r.addrs)
	<-conn.Init(r.opts.Secure, r.opts.TLSConfig)

	return &rmqtportListener{
		addr: addr,
		conn: conn,
		so:   make(map[string]*rmqtportSocket),
	}, nil
}

func (r *rmqtport) String() string {
	return "rabbitmq"
}

func NewTransport(addrs []string, opts ...transport.Option) transport.Transport {
	var options transport.Options
	for _, o := range opts {
		o(&options)
	}

	return &rmqtport{
		opts:     options,
		conn:     newRabbitMQConn("", addrs),
		addrs:    addrs,
		replyTo:  directReplyQueue,
		inflight: make(map[string]chan amqp.Delivery),
	}
}
