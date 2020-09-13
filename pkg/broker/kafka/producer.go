package kafka

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/stats/metric"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/Shopify/sarama"
)

type ProducerConfig struct {
	Addr  []string
	Topic []string

	EnableSASLAuth bool
	SASLMechanism  string
	SASLUser       string
	SASLPassword   string
	SASLHandshake  bool

	DialTimeout      time.Duration
	SlowSendDuration time.Duration

	EnableReturnSuccess bool

	ClientID string
}

type SyncProducer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig

	interceptors []sarama.ProducerInterceptor
}

func NewSyncProducer(config *ProducerConfig) *SyncProducer {
	sconfig := newKafkaProducerConfig(config)
	producer, err := sarama.NewSyncProducer(config.Addr, sconfig)
	if err != nil {
		panic(fmt.Sprintf("create kafka sync producer fail, err msg:%s", err.Error()))
	}

	sp := &SyncProducer{
		producer: producer,
		config:   config,
	}

	return sp
}

func newKafkaProducerConfig(config *ProducerConfig) *sarama.Config {
	sconfig := sarama.NewConfig()

	sconfig.Net.SASL.Enable = config.EnableSASLAuth
	sconfig.Net.SASL.Mechanism = sarama.SASLMechanism(config.SASLMechanism)
	sconfig.Net.SASL.User = config.SASLUser
	sconfig.Net.SASL.Password = config.SASLPassword
	sconfig.Net.SASL.Handshake = config.SASLHandshake
	sconfig.Net.DialTimeout = config.DialTimeout
	sconfig.Version = sarama.V0_10_2_1
	sconfig.Producer.Return.Successes = true

	if "" != config.ClientID {
		sconfig.ClientID = config.ClientID
	}

	return sconfig
}

func (sp *SyncProducer) SendSyncMsg(ctx context.Context, content string) error {
	for _, topic := range sp.config.Topic {
		now := time.Now()
		key := xstring.RandomString(16)
		msg := &sarama.ProducerMessage{
			Topic:     topic,
			Key:       sarama.StringEncoder(key),
			Value:     sarama.StringEncoder(content),
			Timestamp: time.Now(),
		}

		partition, offset, err := sp.producer.SendMessage(msg)

		var errmsg string
		if err != nil {
			errmsg = err.Error()
		}
		duration := time.Since(now).Seconds()
		fields := make([]log.Field, 0, 7)
		fields = append(
			fields,
			log.Any("topic", sp.config.Topic),
			log.String("key", key),
			log.String("content", content),
			log.Int32("partition", partition),
			log.Int64("offset", offset),
			log.Float64("duration", duration),
			log.String("error", errmsg),
		)

		if err != nil {
			log.Error(ctx, "kafka produce fail", fields...)
			metric.KafkaClientHandleCounter.Inc("unknown", "kafka", topic, "produce", "fail")
			metric.KafkaClientReqDuration.Observe(duration, "unknown", "kafka", topic, "produce")
		} else {
			log.Info(ctx, "kafka produce success", fields...)
			metric.KafkaClientHandleCounter.Inc(strconv.Itoa(int(partition)), "kafka", topic, "produce", "success")
			metric.KafkaClientReqDuration.Observe(duration, strconv.Itoa(int(partition)), "kafka", topic, "produce")
		}

		if sp.config.SlowSendDuration > 0 && time.Since(now) > sp.config.SlowSendDuration {
			log.Warn(ctx, "kafka produce slow", fields...)
		}
	}

	return nil
}

func (sp *SyncProducer) Close() error {
	return sp.producer.Close()
}
