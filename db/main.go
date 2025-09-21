package main

import (
	"cc.tim/client/config"
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	"cc.tim/client/logger"
	"cc.tim/client/redis"
	"log"
)

func main() {
	config.Init("../config.yaml")
	logger.InitLogger("error.log")
	defer logger.Sync()
	kafka.InitKafkaConsumer()
	redis.InitRedis()
	db.InitMongo()
	kafka.InitProducer()
	db.InitMysql()
	go CMysqlCommit()
	go CMsg()
	go COffline()
	go CDelete()
	log.Println("consumer启动成功")
	select {}
}
