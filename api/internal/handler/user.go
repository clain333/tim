package handler

import (
	"cc.tim/client/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"im.api/internal/logic"
	"net/http"

	"github.com/mssola/user_agent"
)

// RegisterHandler 处理注册请求
func RegisterHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Captcha  string `json:"captcha"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, "请求参数错误")
		return
	}
	if req.Captcha == "" {
		c.JSON(http.StatusBadRequest, "captcha不能为空")
		return
	}
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, "email不能为空")
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, "密码不能少于6位数")
		return
	}
	ip := c.ClientIP()
	response, err := logic.RegisterLogic(ip, req.Email, req.Password, req.Captcha)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// LoginHandler 处理登录请求
func LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, "请求参数错误")
		return
	}
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, "email不能为空")
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, "密码不能少于6位数")
		return
	}
	ua := c.GetHeader("User-Agent")
	uas := user_agent.New(ua)
	os := uas.OS()
	ip := c.ClientIP()

	if os == "" {
		os = "unknown"
	}
	deviceTAG := c.GetHeader("X-Device-TAG")
	if deviceTAG == "" {
		c.JSON(http.StatusBadRequest, "X-Device-TAG请求头不能为空")
		return
	}
	response, err := logic.LoginLogic(req.Email, req.Password, ip, os, deviceTAG)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// GetCaptchaHandler 处理获取验证码请求
func GetCaptchaHandler(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, "email不能为空")
		return
	}
	ip := c.ClientIP()
	response, err := logic.GetCaptchaLogic(ip, email)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}
