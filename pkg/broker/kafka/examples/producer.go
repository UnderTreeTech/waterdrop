package main

import (
	"context"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/utils/xstring"

	"github.com/UnderTreeTech/waterdrop/pkg/broker/kafka"
	"github.com/UnderTreeTech/waterdrop/pkg/log"
)

func main() {
	defer log.New(nil).Sync()

	config := &kafka.ProducerConfig{
		Addr:           []string{"your_instance_addr"},
		Topic:          []string{"your_instance_topic"},
		EnableSASLAuth: true,
		SASLMechanism:  "your_sasl_mechanis", //PLAIN
		SASLUser:       "your_sasl_user",
		SASLPassword:   "your_sasl_password",
		SASLHandshake:  true,

		DialTimeout:         time.Second * 5,
		EnableReturnSuccess: true,
	}

	producer := kafka.NewSyncProducer(config)

	for i := 0; i < 100000; i++ {
		err := producer.SendSyncMsg(context.Background(), xstring.RandomString(16))
		if err != nil {
			fmt.Println("error", err.Error())
		}
	}

	producer.Close()
}
