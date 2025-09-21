package kafka

import (
	config2 "cc.tim/client/config"
	"testing"
)

func TestKafka(t *testing.T) {
	config2.Init("../../config.yaml")
	InitProducer()
	k := NewInstanceMysql()
	k.SendMessage([]byte("124324425"))
}
