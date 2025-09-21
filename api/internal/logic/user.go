package logic

import (
	"cc.tim/client/kafka"
	pkg2 "cc.tim/client/pkg"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"cc.tim/client/config"
	"cc.tim/client/db"
	"cc.tim/client/model"
	r "cc.tim/client/redis"
	"github.com/go-redis/redis/v8"
)

// GetCaptchaLogic 获取邮箱验证码
func GetCaptchaLogic(ip, email string) (*pkg2.Response, error) {
	countKey := fmt.Sprintf(model.CapChaCountKEY, ip)
	count, err := r.RedisClient.Get(r.Ctx, countKey).Int()
	if err != nil && err != redis.Nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "请求过多", nil), err
	}

	if count >= 5 {
		return pkg2.NewResponse(http.StatusTooManyRequests, "当前ip请求过多", nil), nil
	}

	captcha := pkg2.GenerateCaptcha(config.Config.Captcha.Length)
	body := fmt.Sprintf(config.Config.Captcha.Template, captcha)

	emailSender := pkg2.NewSender()
	if err := emailSender.Sender(email, body); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取失败", nil), err
	}
	captchakey := fmt.Sprintf(model.CaptChaKEY, email)
	r.RedisClient.Set(r.Ctx, captchakey, captcha, 10*time.Minute)
	pipe := r.RedisClient.TxPipeline()
	pipe.Incr(r.Ctx, countKey)
	pipe.Expire(r.Ctx, countKey, 24*time.Hour)
	if _, err := pipe.Exec(r.Ctx); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取失败", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "发送成功", nil), nil
}

// RegisterLogic 用户注册
func RegisterLogic(ip, email, password, captcha string) (*pkg2.Response, error) {

	captchakey := fmt.Sprintf(model.CaptChaKEY, email)
	val, err := r.RedisClient.Get(r.Ctx, captchakey).Result()
	if err != nil {
		if err == redis.Nil {
			return pkg2.NewResponse(http.StatusBadRequest, "验证码已过期或不存在", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "注册失败", nil), err
	}
	if val != captcha {
		return pkg2.NewResponse(http.StatusBadRequest, "验证码错误", nil), nil
	}

	var ok bool
	err = db.MysqlDb.QueryRow("SELECT EXISTS(SELECT 1 FROM user WHERE email=?)", email).Scan(&ok)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "注册失败", nil), err
	}
	if ok {
		return pkg2.NewResponse(http.StatusBadRequest, "邮箱已存在", nil), err
	}

	passwordHash, err := pkg2.HashPassword(password)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "注册失败", nil), err
	}

	username := fmt.Sprintf("新用户%04d", rand.Intn(10000))
	userid := uint64(pkg2.SnowflakeNode.Generate().Int64())

	rr := kafka.NewInstanceMysql()

	if err := rr.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("INSERT INTO user (id,name,email,password,introduction,avatar) VALUES (?,?,?,?,?,?)", userid, username, email, passwordHash, "该人未写简介", "default.jpg"),
	}); err != nil {
		log.Println(err)
		return pkg2.NewResponse(http.StatusInternalServerError, "注册失败", nil), err
	}
	countKey := fmt.Sprintf(model.CapChaCountKEY, ip)

	// 清理 Redis
	r.RedisClient.Del(r.Ctx, captchakey)
	r.RedisClient.Del(r.Ctx, countKey)

	return pkg2.NewResponse(http.StatusOK, "注册成功", map[string]interface{}{
		"user_email": email,
		"user_name":  username,
	}), err
}

// LoginLogic 用户登录
func LoginLogic(email, password, ip, os, deviceTag string) (*pkg2.Response, error) {
	var (
		userID   uint64
		hashPwd  string
		deviceID uint64
		isNew    bool
	)

	if err := db.MysqlDb.QueryRow(
		"SELECT id, password FROM user WHERE email=?", email,
	).Scan(&userID, &hashPwd); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pkg2.NewResponse(http.StatusBadRequest, "邮箱或密码错误", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "登录失败", nil), err
	}

	if !pkg2.CheckPassword(password, hashPwd) {
		return pkg2.NewResponse(http.StatusBadRequest, "邮箱或密码错误", nil), nil
	}

	userKey := fmt.Sprintf(model.UserDeviceKEY, userID)

	count, err := r.RedisClient.SCard(r.Ctx, userKey).Result()
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "登录失败", nil), err
	}

	if count >= 3 {
		// Redis 里已经 3 个设备，必须检查该设备是否存在
		err = db.MysqlDb.QueryRow(
			"SELECT id FROM device WHERE device_tag = ? AND user_id = ?",
			deviceTag, userID,
		).Scan(&deviceID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return pkg2.NewResponse(http.StatusBadRequest, "登陆设备过多", nil), nil
			}
			return pkg2.NewResponse(http.StatusInternalServerError, "登录失败", nil), err
		}

		exists, err := r.RedisClient.SIsMember(r.Ctx, userKey, deviceID).Result()
		if err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "登录失败", nil), err
		}
		if !exists {
			return pkg2.NewResponse(http.StatusBadRequest, "登陆设备过多", nil), nil
		}
	} else {
		// 设备数不足 3，需要先查数据库，看是否存在该设备
		err = db.MysqlDb.QueryRow(
			"SELECT id FROM device WHERE device_tag = ? AND user_id = ?",
			deviceTag, userID,
		).Scan(&deviceID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// 新设备，生成新 ID
				deviceID = uint64(pkg2.SnowflakeNode.Generate().Int64())
				isNew = true
			} else {
				return pkg2.NewResponse(http.StatusInternalServerError, "登录失败", nil), err
			}
		}

		// 把 deviceID 添加到 Redis
		if _, err := r.RedisClient.SAdd(r.Ctx, userKey, deviceID).Result(); err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "登录失败", nil), err
		}
	}

	// 设置 Redis 过期时间
	r.RedisClient.Expire(r.Ctx, userKey, 3*24*time.Hour)

	// 4. Kafka 记录更新/插入
	t := kafka.NewInstanceMysql()
	if isNew {
		t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg(
			"INSERT INTO device (id, user_id, ip, os, device_tag, login_time) VALUES (?,?,?,?,?,?)",
			deviceID, userID, ip, os, deviceTag, time.Now(),
		)})
	} else {
		t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg(
			"UPDATE device SET login_time=? WHERE id=?",
			time.Now(), deviceID,
		)})
	}

	token, _, err := pkg2.GenerateLoginToken(email, userID, deviceID)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "登录失败，请重试", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "登录成功", map[string]interface{}{
		"user_email": email,
		"token":      token,
	}), nil
}
