package rmq

import (
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
	"testing"
)

func cb(msg Message) error {
	logx.Infof(string(msg.GetContent()))

	return nil
}

func TestStartConsumer(t *testing.T) {
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

	err = StartConsumer("rmq-test", []string{"tag-test"}, cb)
	if err != nil {
		t.Fatal(err)
	}

}
