package rocketmq

import (
	"context"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	rocketmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type ProducerConfig struct {
	Endpoint  []string
	AccessKey string
	SecretKey string
	Namespace string

	Retry       int
	SendTimeout time.Duration

	Topic string
	Tags  []string

	interceptors []primitive.Interceptor

	SlowSendDuration time.Duration
}

type Producer struct {
	producer rocketmq.Producer
	config   *ProducerConfig
}

func NewProducer(config *ProducerConfig) *Producer {
	var credentials = primitive.Credentials{
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
	}

	producer, err := rocketmq.NewProducer(
		producer.WithNameServer(config.Endpoint),
		producer.WithRetry(config.Retry),
		producer.WithSendMsgTimeout(config.SendTimeout),
		producer.WithCredentials(credentials),
		producer.WithNamespace(config.Namespace),
		producer.WithInterceptor(producerMetricInterceptor(config)),
		producer.WithInterceptor(config.interceptors...),
	)

	if err != nil {
		panic(fmt.Sprintf("new producer fail, err msg: %s", err.Error()))
	}

	p := &Producer{
		producer: producer,
		config:   config,
	}

	return p
}

func (p *Producer) Start() error {
	return p.producer.Start()
}

func (p *Producer) Shutdown() error {
	return p.producer.Shutdown()
}

func (p *Producer) SendSyncMsg(ctx context.Context, content string) error {
	msgs := getSendMsgs(p.config.Topic, p.config.Tags, content)
	_, err := p.producer.SendSync(ctx, msgs...)
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) SendAsyncMsg(ctx context.Context, content string, callback func(context.Context, *primitive.SendResult, error)) error {
	msgs := getSendMsgs(p.config.Topic, p.config.Tags, content)
	err := p.producer.SendAsync(ctx, callback, msgs...)
	if err != nil {
		return err
	}

	return nil
}

func getSendMsgs(topic string, tags []string, content string) []*primitive.Message {
	var msgs []*primitive.Message

	if 0 == len(tags) {
		msgs = make([]*primitive.Message, 1)
		msgs[0] = primitive.NewMessage(topic, []byte(content)).
			WithKeys([]string{xstring.RandomString(16)})
	} else {
		msgs = make([]*primitive.Message, len(tags))
		for index, tag := range tags {
			msg := primitive.NewMessage(topic, []byte(content)).
				WithTag(tag).WithKeys([]string{xstring.RandomString(16)})
			msgs[index] = msg
		}
	}

	return msgs
}
