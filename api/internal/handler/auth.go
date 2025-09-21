package handler

import (
	"cc.tim/client/logger"
	pkg2 "cc.tim/client/pkg"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"im.api/internal/logic"
	"net/http"
)

// 获取用户信息
func UserInfoHandler(c *gin.Context) {
	userID, ok1 := c.Get("user_id")

	if !ok1 {
		c.JSON(http.StatusOK, "获取失败")
		return
	}
	response, err := logic.GetUserInfoLogic(userID.(uint64))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// 更新用户信息
func UpdateUserInfoHandler(c *gin.Context) {
	var req struct {
		Name         string `json:"name"`
		Introduction string `json:"introduction"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, "请求参数错误")
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, "name不能为空")
		return
	}
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusBadRequest, "获取失败")
		return
	}

	response, err := logic.UpdateUserInfoLogic(userID.(uint64), req.Name, req.Introduction)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// 上传用户头像
func UploadUserAvatarHandler(c *gin.Context) {
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, "头像上传失败")
		return

	}
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusBadRequest, "头像上传失败")
		return
	}

	response, err := logic.UploadUserAvatarLogic(c, file, userID.(uint64))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// ChangePasswordHandler 处理修改密码请求
func ChangePasswordHandler(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, "请求参数错误")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, "旧密码和新密码不能为空")
		return
	}
	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, "旧密码和新密码相同")
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, "新密码长度要大于6位数")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	response, err := logic.ChangePasswordLogic(userID.(uint64), req.OldPassword, req.NewPassword)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// GetDevicesHandler 获取用户登录过的设备列表
func GetDevicesHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}
	DeviceID, exists := c.Get("device_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	response, err := logic.GetDevicesLogic(userID.(uint64), DeviceID.(uint64))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// LogoutDeviceHandler 退出指定设备
func LogoutDeviceHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	deviceID := c.Query("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, "设备ID不能为空")
		return
	}

	response, err := logic.LogoutDeviceLogic(userID.(uint64), pkg2.StrTurnUint(deviceID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}
