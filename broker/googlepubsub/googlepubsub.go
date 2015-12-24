/*
	Google cloud pubsub broker
*/

package googlepubsub

import (
	"sync"
	"time"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/cmd"
	"github.com/pborman/uuid"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/pubsub"

	"golang.org/x/net/context"
)

type pubsubBroker struct {
	ctx context.Context
}

// A pubsub subscriber that manages handling of messages
type subscriber struct {
	opts  broker.SubscribeOptions
	topic string
	ctx   context.Context

	once sync.Once
	exit chan bool
}

// A single publication received by a handler
type publication struct {
	pm    *pubsub.Message
	m     *broker.Message
	ctx   context.Context
	topic string
	sub   string
}

var (
	Key       []byte // Google Developers Console JSON Key
	ProjectID string // Google Developers Console Project ID
	PullNum   = 1    // Number of messages to pull at any given attempt
)

func init() {
	cmd.Brokers["googlepubsub"] = NewBroker
}

func (s *subscriber) run(hdlr broker.Handler) {
	for {
		select {
		case <-s.exit:
			return
		default:
			messages, err := pubsub.PullWait(s.ctx, s.opts.Queue, PullNum)
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			for _, pm := range messages {
				// create broker message
				m := &broker.Message{
					Header: pm.Attributes,
					Body:   pm.Data,
				}

				// create publication
				p := &publication{
					pm:    pm,
					m:     m,
					ctx:   s.ctx,
					topic: s.topic,
					sub:   s.opts.Queue,
				}

				// If the error is nil lets check if we should auto ack
				if err := hdlr(p); err == nil {
					// auto ack?
					if s.opts.AutoAck {
						p.Ack()
					}
				}
			}
		}
	}
}

func (s *subscriber) Config() broker.SubscribeOptions {
	return s.opts
}

func (s *subscriber) Topic() string {
	return s.topic
}

func (s *subscriber) Unsubscribe() error {
	s.once.Do(func() {
		close(s.exit)
	})
	return pubsub.DeleteSub(s.ctx, s.opts.Queue)
}

func (p *publication) Ack() error {
	return pubsub.Ack(p.ctx, p.sub, p.pm.ID)
}

func (p *publication) Topic() string {
	return p.topic
}

func (p *publication) Message() *broker.Message {
	return p.m
}

func (b *pubsubBroker) Address() string {
	return ""
}

func (b *pubsubBroker) Connect() error {
	return nil
}

func (b *pubsubBroker) Disconnect() error {
	return nil
}

func (b *pubsubBroker) Init(opts ...broker.Option) error {
	conf, err := google.JWTConfigFromJSON(Key, pubsub.ScopeCloudPlatform, pubsub.ScopePubSub)
	if err != nil {
		return err
	}

	b.ctx = cloud.NewContext(ProjectID, conf.Client(oauth2.NoContext))
	return nil
}

func (b *pubsubBroker) Publish(topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	exists, err := pubsub.TopicExists(b.ctx, topic)
	if err != nil {
		return err
	}

	if !exists {
		if err := pubsub.CreateTopic(b.ctx, topic); err != nil {
			return err
		}
	}

	m := &pubsub.Message{
		ID:         uuid.NewUUID().String(),
		AckID:      uuid.NewUUID().String(),
		Data:       msg.Body,
		Attributes: msg.Header,
	}

	_, err = pubsub.Publish(b.ctx, topic, m)
	return err
}

func (b *pubsubBroker) Subscribe(topic string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	opt := broker.SubscribeOptions{
		AutoAck: true,
		Queue:   uuid.NewUUID().String(),
	}

	for _, o := range opts {
		o(&opt)
	}

	exists, err := pubsub.SubExists(b.ctx, opt.Queue)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err := pubsub.CreateSub(b.ctx, opt.Queue, topic, time.Duration(0), ""); err != nil {
			return nil, err
		}
	}

	var once sync.Once
	subscriber := &subscriber{
		opts:  opt,
		topic: topic,
		ctx:   b.ctx,
		once:  once,
		exit:  make(chan bool),
	}

	go subscriber.run(h)

	return subscriber, nil
}

func (b *pubsubBroker) String() string {
	return "googlepubsub"
}

func NewBroker(addrs []string, opt ...broker.Option) broker.Broker {
	conf, _ := google.JWTConfigFromJSON(Key, pubsub.ScopeCloudPlatform, pubsub.ScopePubSub)
	ctx := cloud.NewContext(ProjectID, conf.Client(oauth2.NoContext))

	return &pubsubBroker{
		ctx: ctx,
	}
}
