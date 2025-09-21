package db

import (
	"cc.tim/client/config"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var MessageCollection *mongo.Collection

func InitMongo() {
	mongoConf := config.Config.MongoDB

	uri := fmt.Sprintf("mongodb://%s:%d", mongoConf.Host, mongoConf.Port)
	if mongoConf.Username != "" && mongoConf.Password != "" {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%d", mongoConf.Username, mongoConf.Password, mongoConf.Host, mongoConf.Port)
	}

	clientOptions := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(mongoConf.Timeout)*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatalf("MongoDB 连接失败: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB 测试连接失败: %v", err)
	}
	MessageCollection = client.Database(mongoConf.Database).Collection(config.Config.MongoDB.Database)
}
