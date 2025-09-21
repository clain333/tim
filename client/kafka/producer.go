package kafka

import (
	"cc.tim/client/logger"
	"cc.tim/client/model"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net"
	"time"

	config2 "cc.tim/client/config"
	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	Topic    string
}

var (
	prod sarama.SyncProducer
)

// InitProducer 初始化 Kafka 生产者（单例模式）
func InitProducer() {

	brokers := []string{config2.Config.Kafka.Brokers}
	for _, broker := range brokers {
		conn, err := net.DialTimeout("tcp", broker, 2*time.Second)
		if err != nil {
			panic(fmt.Sprintf("无法连接 Kafka broker %s: %v", broker, err))
		}
		conn.Close()
	}
	config := sarama.NewConfig()
	config.Version = sarama.V4_0_0_0
	config.Net.SASL.Enable = true
	config.Net.SASL.User = config2.Config.Kafka.User
	config.Net.SASL.Password = config2.Config.Kafka.Password
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.TLS.Enable = false
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	var err error
	prod, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(fmt.Sprintf("初始化 Kafka 生产者失败: %v", err))
	}

}

func NewInstanceMysql() *Producer {
	return &Producer{
		producer: prod,
		Topic:    config2.Config.Kafka.Topics["mysql"],
	}
}
func NewInstanceMsg() *Producer {
	return &Producer{
		producer: prod,
		Topic:    config2.Config.Kafka.Topics["msg"],
	}
}

func NewInstanceConn() *Producer {
	return &Producer{
		producer: prod,
		Topic:    config2.Config.Kafka.Topics["conn"],
	}
}

func NewInstanceOffline() *Producer {
	return &Producer{
		producer: prod,
		Topic:    config2.Config.Kafka.Topics["offline"],
	}
}

func NewInstanceDelete() *Producer {
	return &Producer{
		producer: prod,
		Topic:    config2.Config.Kafka.Topics["delete"],
	}
}

// SendMessage 发送消息到指定 Topic
func (p *Producer) SendMessage(value []byte) error {
	if len(value) == 0 {
		return fmt.Errorf("Kafka 消息内容不能为空")
	}

	msg := &sarama.ProducerMessage{
		Topic: p.Topic,
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (p *Producer) SendMysqlMessage(value []*model.KafkaMysqlMsg) error {
	if value == nil {
		return fmt.Errorf("Kafka 消息内容不能为空")
	}
	msgs, err := json.Marshal(value)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: p.Topic,
		Value: sarama.ByteEncoder(msgs),
	}
	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
		return err
	}
	return nil
}
