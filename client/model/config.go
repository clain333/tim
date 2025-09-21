package model

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Ws      ServerConfig  `yaml:"ws"`
	MySQL   MySQLConfig   `yaml:"mysql"`
	Email   EmailConfig   `yaml:"email"`
	Captcha CaptchaConfig `yaml:"captcha"`
	Redis   RedisConfig   `yaml:"redis"`
	Jwt     Jwt           `yaml:"jwt"`
	Avatar  Avatar        `yaml:"avatar"`
	Kafka   Kafka         `yaml:"kafka"`
	MongoDB MongoConfig   `yaml:"mongodb"`
	Upload  FileSize      `yaml:"upload"`
}
type FileSize struct {
	Maxsize int64 `yaml:"maxsize"`
}
type MongoConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Timeout  int    `yaml:"timeout"`
}
type Kafka struct {
	Brokers          string            `yaml:"brokers"`            // Kafka 集群地址
	Topics           map[string]string `yaml:"topics"`             // 对应原 RabbitMQ 的队列
	GroupID          string            `yaml:"group_id"`           // 消费者组
	EnableAutoCommit bool              `yaml:"enable_auto_commit"` // 是否自动提交 Offset
	AutoOffsetReset  string            `yaml:"auto_offset_reset"`  // 消费起始位置
	User             string            `yaml:"user"`
	Password         string            `yaml:"password"`
}

type Avatar struct {
	Src     string `yaml:"src"`
	MaxSize int64  `yaml:"maxSize"`
}
type Jwt struct {
	Secret string `yaml:"secret"`
}
type ServerConfig struct {
	Port int `yaml:"port"`
}

type MySQLConfig struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
	DBName    string `yaml:"dbname"`
	Charset   string `yaml:"charset"`
	ParseTime bool   `yaml:"parseTime"`
	Loc       string `yaml:"loc"`
}

type EmailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type CaptchaConfig struct {
	Template string `yaml:"template"`
	Length   int    `yaml:"length"`
	Subject  string `yaml:"subject"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
