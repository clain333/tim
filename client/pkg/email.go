package pkg

import (
	"cc.tim/client/config"
	"gopkg.in/gomail.v2"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	EMAIL = iota
)

var SENDERTYPE int = 0

func init() {
	sendtype := os.Getenv("CAPTCHA_SENDTYPE")
	switch sendtype {
	case "EMAIL":
		SENDERTYPE = EMAIL
	}

}

type SendCaptcha interface {
	Sender(to, captcha string) error
}

type SMTPEmailSender struct {
}

func NewSender() SendCaptcha {
	switch SENDERTYPE {
	case EMAIL:
		return &SMTPEmailSender{}
	}
	return nil
}

// SendCaptcha 发送验证码到邮箱
func (s *SMTPEmailSender) Sender(to, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.Config.Email.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", config.Config.Captcha.Subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.qq.com", 587, config.Config.Email.Username, config.Config.Email.Password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// GenerateCaptcha 根据配置生成验证码
func GenerateCaptcha(ii int) string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < ii; i++ {
		code += strconv.Itoa(rand.Intn(10))
	}
	return code
}
