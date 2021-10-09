package rmq

import (
	"context"
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
	"hash/fnv"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

type queueSelectorByShardingKey struct{}

func (q *queueSelectorByShardingKey) Select(msg *primitive.Message, queues []*primitive.MessageQueue) *primitive.MessageQueue {
	if msg.GetShardingKey() == "" {
		return queues[rand.Int()%len(queues)]
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(msg.GetShardingKey()))
	return queues[h.Sum32()%uint32(len(queues))]
}

func newProducer(ak, sk string, instance, group string, nsDomain string, retry int, timeout time.Duration) (*rmqProducer, error) {
	options := []producer.Option{
		producer.WithInstanceName(instance + "-" + strconv.Itoa(os.Getpid()) + "-producer"),
		producer.WithGroupName(group),
		producer.WithNameServerDomain(nsDomain),
		producer.WithRetry(retry),
		producer.WithQueueSelector(&queueSelectorByShardingKey{}),
	}
	if ak != "" && sk != "" {
		options = append(options, producer.WithCredentials(primitive.Credentials{
			AccessKey: ak,
			SecretKey: sk,
		}))
	}
	if timeout != 0 {
		options = append(options, producer.WithSendMsgTimeout(timeout))
	}
	prod, err := rocketmq.NewProducer(options...)
	if err != nil {
		logx.Errorf("failed to create producer, ns:%s, err:%s", nsDomain, err.Error())
		return nil, err
	}

	return &rmqProducer{
		producer: prod,
		started:  false,
	}, nil
}

type rmqProducer struct {
	producer rocketmq.Producer
	started  bool
}

func (p *rmqProducer) start() error {
	err := p.producer.Start()
	if err != nil {
		logx.Errorf("failed to start consumer, err:%s", err.Error())
		return err
	}

	return nil
}

func (p *rmqProducer) stop() error {
	return p.producer.Shutdown()
}

func (p *rmqProducer) SendMessage(msgs ...*primitive.Message) (string, string, string, error) {
	res, err := p.producer.SendSync(context.Background(), msgs...)
	if err != nil {
		logx.Errorf("failed to send messages:%s", err.Error())
		return "", "", "", err
	}
	return res.MessageQueue.String(), res.MsgID, res.OffsetMsgID, err
}
