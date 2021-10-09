// Package rmq 提供了访问RocketMQ服务的能力
package rmq

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

const prefix = "@@rmq."

var (
	// ErrRmqSvcConfigInvalid 服务配置无效
	ErrRmqSvcConfigInvalid = fmt.Errorf("requested rmq service is not correctly configured")
	// ErrRmqSvcNotRegiestered 服务尚未被注册
	ErrRmqSvcNotRegiestered = fmt.Errorf("requested rmq service is not registered")
	// ErrRmqSvcInvalidOperation 当前操作无效
	ErrRmqSvcInvalidOperation = fmt.Errorf("requested rmq service is not suitable for current operation")
)

var (
	rmqServices   = make(map[string]*client)
	rmqServicesMu sync.Mutex
)

// MessageCallback 定义业务方接收消息的回调接口
type MessageCallback func(msg Message) error

func (conf *ClientConfig) checkConfig() error {

	if conf.Group == "" {
		return ErrRmqSvcConfigInvalid
	}
	if conf.Topic == "" {
		return ErrRmqSvcConfigInvalid
	}
	if len(conf.NameServers) == 0 {
		return ErrRmqSvcConfigInvalid
	}
	return nil
}

func InitRmq(service string, config ClientConfig) (err error) {
	if err = config.checkConfig(); err != nil {
		return err
	}

	clnt := &client{
		ClientConfig: &config,
	}
	rmqServicesMu.Lock()
	defer rmqServicesMu.Unlock()

	err = clnt.startNamingHandler()
	if err != nil {
		return err
	}

	rmqServices[service] = clnt
	return nil
}

// StartProducer 启动指定已注册的RocketMQ生产服务
func StartProducer(service string) error {
	if client, ok := rmqServices[service]; ok {
		client.mu.Lock()
		defer client.mu.Unlock()
		if client.producer != nil {
			return ErrRmqSvcInvalidOperation
		}
		var err error
		var nsDomain string
		nsDomain, err = client.getNameserverDomain()
		if err != nil {
			return err
		}
		client.producer, err = newProducer(
			client.ClientConfig.Auth.AccessKey, client.ClientConfig.Auth.SecretKey,
			service, client.ClientConfig.Group, nsDomain,
			client.ClientConfig.Retry, time.Duration(client.ClientConfig.Timeout)*time.Millisecond)
		if err != nil {
			return err
		}
		return client.producer.start()
	}
	return ErrRmqSvcNotRegiestered
}

// StopProducer 停止指定已注册的RocketMQ生产服务
func StopProducer(service string) error {
	if client, ok := rmqServices[service]; ok {
		client.mu.Lock()
		defer client.mu.Unlock()
		if client.producer == nil {
			return ErrRmqSvcInvalidOperation
		}
		err := client.producer.stop()
		client.producer = nil
		return err
	}
	return ErrRmqSvcNotRegiestered
}

func call(fn MessageCallback, m *primitive.MessageExt) (err error) {
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]

			info, _ := json.Marshal(map[string]interface{}{
				"time":   time.Now().Format("2006-01-02 15:04:05"),
				"level":  "error",
				"module": "stack",
			})
			log.Printf("%s\n-------------------stack-start-------------------\n%v\n%s\n-------------------stack-end-------------------\n", string(info), r, buf)
		}
	}()

	err = fn(&messageWrapper{
		msg:      &m.Message,
		offsetID: m.OffsetMsgId,
	})
	if err != nil {
		logx.Errorf("failed to consume message:%s", err.Error())
	}

	return err
}

// StartConsumer 启动指定已注册的RocketMQ消费服务， 同时指定要消费的消息标签，以及消费回调
func StartConsumer(service string, tags []string, callback MessageCallback) error {
	if _, exist := rmqServices[service]; !exist {
		return ErrRmqSvcNotRegiestered
	}
	client := rmqServices[service]
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.pushConsumer != nil || callback == nil {
		return ErrRmqSvcInvalidOperation
	}
	var err error
	var nsDomain string
	nsDomain, err = client.getNameserverDomain()
	if err != nil {
		logx.Errorf("invalid consumer nameServer, err:%s", err.Error())
		return err
	}

	cb := func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, m := range msgs {
			if ctx.Err() != nil {
				logx.Errorf("stop consume cause ctx cancelled, err:%s", err.Error())
				return consumer.SuspendCurrentQueueAMoment, ctx.Err()
			}
			if err := call(callback, m); err != nil {
				return consumer.SuspendCurrentQueueAMoment, nil
			}
		}
		return consumer.ConsumeSuccess, nil
	}

	client.pushConsumer, err = newPushConsumer(
		client.ClientConfig.Auth.AccessKey,
		client.ClientConfig.Auth.SecretKey,
		service,
		client.ClientConfig.Group,
		client.ClientConfig.Topic,
		client.ClientConfig.Broadcast,
		client.ClientConfig.Orderly,
		client.ClientConfig.Retry,
		tags,
		nsDomain,
		cb)
	if err != nil {
		logx.Errorf("create new consumer error:%s", err.Error())
		return err
	}
	return client.pushConsumer.start()
}

// StopConsumer 停止指定已注册的RocketMQ消费服务
func StopConsumer(service string) error {
	if client, exist := rmqServices[service]; exist {
		client.mu.Lock()
		defer client.mu.Unlock()
		if client.pushConsumer == nil {
			return ErrRmqSvcInvalidOperation
		}
		err := client.pushConsumer.stop()
		client.pushConsumer = nil
		return err
	}
	return ErrRmqSvcNotRegiestered
}

// NewMessage 创建一条新的消息
func NewMessage(service string, content []byte) (Message, error) {
	if client, exist := rmqServices[service]; exist {
		return &messageWrapper{
			client: client,
			msg:    primitive.NewMessage(client.ClientConfig.Topic, content),
		}, nil
	}
	return nil, ErrRmqSvcNotRegiestered
}

var consumers []string

func Use(service string, tags []string, handler MessageCallback) {
	if err := StartConsumer(service, tags, handler); err != nil {
		panic("Start consumer  error: " + err.Error())
	}
	consumers = append(consumers, service)
}

func StopRocketMqConsume() {
	for _, svc := range consumers {
		_ = StopConsumer(svc)
	}
}
