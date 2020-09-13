package kafka

import "github.com/Shopify/sarama"

type TraceInterceptor struct {
	TraceID string
}

func (ti *TraceInterceptor) OnSend(message *sarama.ProducerMessage) {

}
