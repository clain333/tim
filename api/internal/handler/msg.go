package handler

import (
	"cc.tim/client/logger"
	"cc.tim/client/model"
	pkg2 "cc.tim/client/pkg"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"im.api/internal/logic"
	"net/http"
)

func OfflineMsgHandler(c *gin.Context) {
	deviceid, ok := c.Get("device_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}
	response, err := logic.OfflineMsgLogic(deviceid.(uint64))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}
func MsgHandler(c *gin.Context) {
	cid := c.Query("conversation_id")
	num := c.Query("num")
	page := c.Query("page")
	response, err := logic.MsgLogic(pkg2.StrTurnUint(cid), pkg2.StrTurnUint(num), pkg2.StrTurnUint(page))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func SendMsgHandler(c *gin.Context) {
	var msg model.KafkaINMsg
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, "结构体格式不正确")
	}
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}
	response, err := logic.SendMsgLogic(userID.(uint64), msg)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}
