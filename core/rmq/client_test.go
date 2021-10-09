package rmq

import (
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"testing"
)

func TestMessageWrapper_Send(t *testing.T) {
	conf := ClientConfig{
		Service:     "rmq-test",
		NameServers: []string{"xxx.com"},
		Topic:       "topic-test",
	}

	err := InitRmq("rmq-test", conf)
	if err != nil {
		t.Fatal(err)
	}

	if err := StartProducer("rmq-test"); err != nil {
		t.Fatal(err)
	}

	msg := messageWrapper{
		msg: &primitive.Message{
			Topic: "test",
			Body:  []byte("Hello RocketMQ Go Client!"),
		},
		client: rmqServices["rmq-test"],
	}

	msg.WithTag("tag-test")

	_, err = msg.Send()
	if err != nil {
		t.Fatal(err)
	}
}
