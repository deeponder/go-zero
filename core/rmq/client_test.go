package rmq

import (
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"testing"
)

func TestMessageWrapper_Send(t *testing.T) {
	conf := ClientConfig{
		Service:     "rmq-test",
		NameServers: []string{"192.168.50.41:9876"},
		Topic:       "ablogs",
		Group:       "ab",
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
			Topic: "ablogs",
			Body:  []byte("Hello RocketMQ, From go-zero"),
		},
		client: rmqServices["rmq-test"],
	}

	msg.WithTag("tag-test")

	_, err = msg.Send()
	if err != nil {
		t.Fatal(err)
	}
}
