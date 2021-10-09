package rmq

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"gitlab.deepwisdomai.com/infra/go-zero/core/logx"
)

// auth 提供链接到Broker所需要的验证信息（按需配置）
type auth struct {
	AccessKey string `json:"ak,omitempty" yaml:"ak,omitempty"`
	SecretKey string `json:"sk,omitempty" yaml:"sk,omitempty"`
}

// ClientConfig 包含链接到RocketMQ服务所需要的各配置项
type ClientConfig struct {
	// 集群名字
	Service string `json:"-" yaml:"service"`
	// 提供名字服务器的地址列表，例如: [ "127.0.0.1:9876" ]
	NameServers []string `json:"nameservers" yaml:"nameservers"`
	// 生产/消费者组名称，各业务线间需要保持唯一
	Group string `json:"group" yaml:"group"`
	// 要消费/订阅的主题
	Topic string `json:"topic" yaml:"topic"`
	// 如果配置了ACL，需提供验证信息
	Auth auth `json:"auth" yaml:"auth"`
	// 是否是广播消费模式
	Broadcast bool `json:"broadcast" yaml:"broadcast"`
	// 是否是顺序消费模式
	Orderly bool `json:"orderly" yaml:"orderly"`
	// 生产失败时的重试次数
	Retry int `json:"retry" yaml:"retry"`
	// 生产超时时间
	Timeout int `json:"timeout" yaml:"timeout"`
}

// Client 为客户端主体结构
type client struct {
	*ClientConfig
	mu sync.RWMutex

	producer       *rmqProducer
	pushConsumer   *rmqPushConsumer
	namingListener net.Listener
}

func (c *client) startNamingHandler() error {
	var err error
	c.namingListener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		logx.Error("failed to create naming listener")
		return err
	}
	go func() {
		err = http.Serve(c.namingListener, c.createNamingHandler())
		logx.Errorf("naming handler stopped, err:%s", err.Error())
	}()
	return nil
}

func (c *client) createNamingHandler() http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		if len(c.ClientConfig.NameServers) > 0 {
			logx.Infof("try serve through static config, nameServer: %s", c.ClientConfig.NameServers)
			var firstItem = true
			for _, ns := range c.ClientConfig.NameServers {
				var parts = strings.Split(ns, ":")
				if len(parts) != 2 {
					logx.Errorf("invalid nameserver config: %s", ns)
					continue
				}
				var host = parts[0]
				var port = parts[1]
				// have to resolve the domain name to ips
				addrs, err := net.LookupHost(host)
				if err != nil {
					logx.Errorf("failed to lookup nameserver, host:%s", host)
					continue
				}
				for _, addr := range addrs {
					if !firstItem {
						_, err := resp.Write([]byte(";"))
						if err != nil {
							logx.Errorf("write response failed, err:%s", err.Error())
						}
					}
					_, err := resp.Write([]byte(addr + ":" + port))
					if err != nil {
						logx.Errorf("write response failed, err:%s", err.Error())
					}
					firstItem = false
				}
			}
			return
		}

		// no ns available
		resp.WriteHeader(http.StatusNotFound)
	}
}
func (c *client) getNameserverDomain() (string, error) {
	if c.namingListener != nil {
		return "http://" + c.namingListener.Addr().String(), nil
	}
	return "", ErrRmqSvcInvalidOperation
}

// DelayLevel 定义消息延迟发送的级别
type DelayLevel int

const (
	Second = DelayLevel(iota)
	Seconds5
	Seconds10
	Seconds30
	Minute1
	Minutes2
	Minutes3
	Minutes4
	Minutes5
	Minutes6
	Minutes7
	Minutes8
	Minutes9
	Minutes10
	Minutes20
	Minutes30
	Hour1
	Hours2
)

type rlogger struct {
	initOnce sync.Once
	verbose  bool
}

func (r *rlogger) Level(level string) {}

func (r *rlogger) isVerbose() bool {
	r.initOnce.Do(func() {
		if os.Getenv("RMQ_SDK_VERBOSE") != "" {
			r.verbose = true
		} else {
			r.verbose = false
		}
	})
	return r.verbose
}

// Message 消息提供的接口定义
type Message interface {
	WithTag(string) Message
	WithShard(string) Message
	WithDelay(DelayLevel) Message
	Send() (msgID string, err error)
	GetContent() []byte
	GetTag() string
	GetShard() string
	GetID() string
}

type messageWrapper struct {
	msg      *primitive.Message
	client   *client
	offsetID string
}

// WithTag 设置消息的标签Tag
func (m *messageWrapper) WithTag(tag string) Message {
	m.msg = m.msg.WithTag(tag)
	return m
}

// WithShard 设置消息的分片键
func (m *messageWrapper) WithShard(shard string) Message {
	m.msg = m.msg.WithShardingKey(shard)
	return m
}

// WithDelay 设置消息的延迟等级
func (m *messageWrapper) WithDelay(lvl DelayLevel) Message {
	m.msg = m.msg.WithDelayTimeLevel(int(lvl))
	return m
}

// Send 发送消息
func (m *messageWrapper) Send() (msgID string, err error) {
	if m.client == nil {
		logx.Errorf("client is not specified")
		return "", ErrRmqSvcInvalidOperation
	}
	m.client.mu.Lock()
	prod := m.client.producer
	m.client.mu.Unlock()
	if prod == nil {
		logx.Errorf("producer not started")
		return "", ErrRmqSvcInvalidOperation
	}
	queue, id, offset, err := m.client.producer.SendMessage(m.msg)
	if err != nil {
		logx.Errorf("failed to send message, err:%s, msg:%s", err.Error(), m.msg.String())
		return "", err
	}

	logx.Infof("sent message, msg:%s, queue:%s, msgid:%s, offsetid:%s", m.msg.String(), queue, id, offset)

	return offset, nil
}

// GetContent 获取消息体内容
func (m *messageWrapper) GetContent() []byte {
	return m.msg.Body
}

// GetTag 获取消息标签
func (m *messageWrapper) GetTag() string {
	return m.msg.GetTags()
}

// GetShard 获取消息分片键
func (m *messageWrapper) GetShard() string {
	return m.msg.GetShardingKey()
}

// GetID 获取消息ID
func (m *messageWrapper) GetID() string {
	return m.offsetID
}

type MessageBatch []Message

func (batch MessageBatch) Send() (msgID string, err error) {
	var msgs = make([]*primitive.Message, 0)
	for _, m := range batch {
		msgs = append(msgs, m.(*messageWrapper).msg)
	}
	if len(msgs) < 1 {
		return "", ErrRmqSvcInvalidOperation
	}

	queue, id, offset, err := batch[0].(*messageWrapper).client.producer.SendMessage(msgs...)

	if err != nil {
		logx.Errorf("failed to send message batch, err%s", err.Error())
		return "", err
	}

	logx.Infof("sent message batch, queue:%s, msgid:%s, offsetid:%s", queue, id, offset)

	return offset, nil
}
