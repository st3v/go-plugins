package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/micro/go-micro/broker"
	"github.com/micro/go-micro/cmd"
	"golang.org/x/net/context"
	"time"
)

const (
	defaultMaxMessages       = 1
	defaultVisibilityTimeout = 3
	defaultWaitSeconds       = 10
)

// Amazon SQS Broker
type sqsBroker struct {
	session *session.Session
	svc     *sqs.SQS
	options broker.Options
}

// A subscriber (poller) to an SQS queue
type subscriber struct {
	options broker.SubscribeOptions
	svc     *sqs.SQS
	URL     string
	exit    chan bool
}

// A wrapper around a message published on an SQS queue and delivered via subscriber
type publication struct {
	sMessage *sqs.Message
	svc      *sqs.SQS
	m        *broker.Message
	URL      string
}

func init() {
	cmd.DefaultBrokers["sqs"] = NewBroker
}

// run is designed to run as a goroutine and poll SQS for new messages. Note that it's possible to receive
// more than one message from a single poll depending on the options configured for the plugin
func (s *subscriber) run(hdlr broker.Handler) {
	for {
		select {
		case <-s.exit:
			return
		default:
			result, err := s.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:            &s.URL,
				MaxNumberOfMessages: s.getMaxMessages(),
				VisibilityTimeout:   s.getVisibilityTimeout(),
				WaitTimeSeconds:     s.getWaitSeconds(),
				AttributeNames: aws.StringSlice([]string{
					"SentTimestamp", // TODO: not currently exposing this to plugin users
				}),
				MessageAttributeNames: aws.StringSlice([]string{
					"All",
				}),
			})

			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			if len(result.Messages) == 0 {
				time.Sleep(time.Second)
				continue
			}

			for _, sm := range result.Messages {
				s.handleMessage(sm, hdlr)
			}
		}
	}
}

func (s *subscriber) getMaxMessages() *int64 {
	if v := s.options.Context.Value(maxMessagesKey{}); v != nil {
		v2 := v.(int64)
		return aws.Int64(v2)
	}
	return aws.Int64(defaultMaxMessages)
}

func (s *subscriber) getVisibilityTimeout() *int64 {
	if v := s.options.Context.Value(visiblityTimeoutKey{}); v != nil {
		v2 := v.(int64)
		return aws.Int64(v2)
	}
	return aws.Int64(defaultVisibilityTimeout)
}

func (s *subscriber) getWaitSeconds() *int64 {
	if v := s.options.Context.Value(waitTimeSecondsKey{}); v != nil {
		v2 := v.(int64)
		return aws.Int64(v2)
	}
	return aws.Int64(defaultWaitSeconds)
}

func (s *subscriber) handleMessage(msg *sqs.Message, hdlr broker.Handler) {
	m := &broker.Message{
		Header: buildMessageHeader(msg.MessageAttributes),
		Body:   []byte(*msg.Body),
	}

	p := &publication{
		sMessage: msg,
		m:        m,
		URL:      s.URL,
		svc:      s.svc,
	}

	if err := hdlr(p); err == nil {
		if s.options.AutoAck {
			p.Ack()
		}
	}
}

func (s *subscriber) Options() broker.SubscribeOptions {
	return s.options
}

func (s *subscriber) Topic() string {
	return s.URL
}

func (s *subscriber) Unsubscribe() error {
	select {
	case <-s.exit:
		return nil
	default:
		close(s.exit)
		return nil
	}
}

func (p *publication) Ack() error {
	_, err := p.svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &p.URL,
		ReceiptHandle: p.sMessage.ReceiptHandle,
	})
	return err
}

func (p *publication) Topic() string {
	return p.URL
}

func (p *publication) Message() *broker.Message {
	return p.m
}

func (b *sqsBroker) Options() broker.Options {
	return b.options
}

func (b *sqsBroker) Address() string {
	return ""
}

// Connect does nothing as AWS does all queue operations in a single shot with no persistent connection
func (b *sqsBroker) Connect() error {
	return nil
}

// Disconnect does nothing as there's no live connection to terminate
func (b *sqsBroker) Disconnect() error {
	return nil
}

// Init initializes a broker and configures an AWS session and SQS struct
func (b *sqsBroker) Init(opts ...broker.Option) error {

	for _, o := range opts {
		o(&b.options)
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := sqs.New(sess)
	b.svc = svc
	b.session = sess

	return nil
}

// Publish publishes a message via SQS
func (b *sqsBroker) Publish(queueURL string, msg *broker.Message, opts ...broker.PublishOption) error {
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(msg.Body[:])),
		QueueUrl:    &queueURL,
	}
	input.MessageAttributes = copyMessageHeader(msg)
	input.MessageDeduplicationId = b.generateDedupID(msg)
	input.MessageGroupId = b.generateGroupID(msg)

	_, err := b.svc.SendMessage(input)

	if err != nil {
		return err
	}

	// Broker interfaces don't let us do anything with message ID or sequence number
	return nil
}

// Subscribe subsribes to an SQS queue, starting a goroutine to poll for messages
func (b *sqsBroker) Subscribe(queueURL string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	options := broker.SubscribeOptions{
		AutoAck: true,
		Queue:   queueURL,
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	subscriber := &subscriber{
		options: options,
		URL:     queueURL,
		svc:     b.svc,
		exit:    make(chan bool),
	}
	go subscriber.run(h)

	return subscriber, nil
}

// String returns the name of the broker plugin
func (b *sqsBroker) String() string {
	return "sqs"
}

func copyMessageHeader(m *broker.Message) (attribs map[string]*sqs.MessageAttributeValue) {
	attribs = make(map[string]*sqs.MessageAttributeValue)
	for k, v := range m.Header {
		attribs[k] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(v),
		}
	}
	return attribs
}

func buildMessageHeader(attribs map[string]*sqs.MessageAttributeValue) map[string]string {
	res := make(map[string]string)

	for k, v := range attribs {
		res[k] = *v.StringValue
	}
	return res
}

func (b *sqsBroker) generateGroupID(m *broker.Message) *string {
	raw := b.options.Context.Value(groupIdFunctionKey{})
	if raw != nil {
		s := raw.(StringFromMessageFunc)(m)
		return &s
	}
	return nil
}

func (b *sqsBroker) generateDedupID(m *broker.Message) *string {
	raw := b.options.Context.Value(dedupFunctionKey{})
	if raw != nil {
		s := raw.(StringFromMessageFunc)(m)
		return &s
	}
	return nil
}

// NewBroker creates a new broker with options
func NewBroker(opts ...broker.Option) broker.Broker {
	options := broker.Options{
		Context: context.Background(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &sqsBroker{
		options: options,
	}
}
