package main

import (
	"context"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/waterdrop/pkg/mq/rocketmq"
)

func main() {
	defer log.New(nil).Sync()

	config := &rocketmq.ProducerConfig{
		Endpoint:    []string{"your_endpoint"},
		AccessKey:   "your_access_key",
		SecretKey:   "your_secret_key",
		Namespace:   "your_namespace",
		Retry:       1,
		SendTimeout: time.Second,
		Topic:       "your_topic",
		Tags:        []string{"go-rocketmq"},
	}

	p := rocketmq.NewProducer(config)
	p.Start()

	for i := 0; i < 100; i++ {
		p.SendSyncMsg(context.Background(), "Hello RocketMQ Go Client!"+xstring.RandomString(16))
	}

	p.Shutdown()
}
