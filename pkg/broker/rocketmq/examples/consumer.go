package main

import (
	"context"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/broker/rocketmq"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func main() {
	defer log.New(nil).Sync()

	config := &rocketmq.ConsumerConfig{
		Endpoint:  []string{"your_endpoint"},
		AccessKey: "your_access_key",
		SecretKey: "your_secret_key",
		Namespace: "your_namespace",
		Topic:     "your_topic",
		Gid:       "your_group_id",
		Tags:      []string{"go-rocketmq"},
	}

	consumer := rocketmq.NewPushConsumer(config)
	consumer.Subscribe(consumeMsg)

	consumer.Start()
	time.Sleep(time.Hour)
}

func consumeMsg(ctx context.Context, msg *primitive.MessageExt) error {
	fmt.Println("msg", msg.String())
	return nil
}
