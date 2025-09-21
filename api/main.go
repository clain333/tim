package main

import (
	"cc.tim/client/config"
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	"cc.tim/client/logger"
	"cc.tim/client/pkg"
	"cc.tim/client/redis"
	"im.api/internal/consumer"
	"im.api/internal/router"
	"strconv"
	"time"
)

func main() {
	config.Init("../config.yaml")
	logger.InitLogger("error.log")
	defer logger.Sync()
	pkg.Init()
	redis.InitRedis()
	db.InitMysql()
	db.InitMongo()
	kafka.InitProducer()
	kafka.InitKafkaConsumer()
	go consumer.CConn()
	r := router.InitRouter()

	port := config.Config.Server.Port
	if port == 0 {
		port = 8080
	}
	_, err := time.LoadLocation(config.Config.MySQL.Loc)
	if err != nil {
		panic(err)
	}
	r.Run(":" + strconv.Itoa(port))
}
