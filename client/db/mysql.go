package db

import (
	"cc.tim/client/config"
	"database/sql"
	"fmt"
	_ "gorm.io/driver/mysql"
	"log"
	"net/url"
	"time"
)

var MysqlDb *sql.DB

func InitMysql() {
	// 拼接 DSN
	loc := url.QueryEscape(config.Config.MySQL.Loc)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		config.Config.MySQL.Username,
		config.Config.MySQL.Password,
		config.Config.MySQL.Host,
		config.Config.MySQL.Port,
		config.Config.MySQL.DBName,
		config.Config.MySQL.Charset,
		config.Config.MySQL.ParseTime,
		loc,
	)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 设置连接池
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	// 检查连接是否可用
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库不可用: %v", err)
	}

	MysqlDb = db
}
