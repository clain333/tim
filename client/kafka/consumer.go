package kafka

import (
	config2 "cc.tim/client/config"
	"cc.tim/client/logger"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"log"
)

var consumer sarama.Consumer

type ConsumerHand func(b []byte) error

func InitKafkaConsumer() error {

	config := sarama.NewConfig()
	config.Version = sarama.V4_0_0_0
	config.Net.SASL.Enable = true
	config.Net.SASL.User = config2.Config.Kafka.User
	config.Net.SASL.Password = config2.Config.Kafka.Password
	config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	config.Net.TLS.Enable = false
	brokers := []string{config2.Config.Kafka.Brokers}
	var err error
	consumer, err = sarama.NewConsumer(brokers, config)
	if err != nil {
		return err
	}
	return nil

}

func Consume(topic string, handler ConsumerHand) {
	partitions, err := consumer.Partitions(topic)

	if err != nil {
		log.Fatalf("无法获取分区: %v", err)
	}
	for _, partition := range partitions {
		pc, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("无法消费分区 %d: %v", partition, err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()
			for {
				select {
				case msg := <-pc.Messages():
					if err := handler(msg.Value); err != nil {
						logger.Error("发生错误", zap.Error(err))
					}
				case err := <-pc.Errors():
					logger.Error("发生错误", zap.Error(err))
				}
			}
		}(pc)
	}

}
