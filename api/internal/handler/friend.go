package handler

import (
	"cc.tim/client/logger"
	pkg2 "cc.tim/client/pkg"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"im.api/internal/logic"
	"net/http"
)

func SendFriendRequestHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	friendID := c.Query("friend_id")
	if friendID == "" {
		c.JSON(http.StatusBadRequest, "好友ID不能为空")
		return
	}

	response, err := logic.SendFriendRequestLogic(userID.(uint64), pkg2.StrTurnUint(friendID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func HandleFriendRequestHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	requestID := c.Query("friend_id")
	action := c.Query("action")
	if requestID == "" || action == "" {
		c.JSON(http.StatusBadRequest, "请求ID或操作不能为空")
		return
	}

	response, err := logic.HandleFriendRequestLogic(userID.(uint64), pkg2.StrTurnUint(requestID), int8(pkg2.StrTurnUint(action)))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func GetFriendRequestsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	response, err := logic.GetFriendRequestsLogic(userID.(uint64))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func GetFriendListHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	response, err := logic.GetFriendListLogic(userID.(uint64))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func DeleteFriendHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	friendID := c.Query("friend_id")
	if friendID == "" {
		c.JSON(http.StatusBadRequest, "好友ID不能为空")
		return
	}

	response, err := logic.DeleteFriendLogic(userID.(uint64), pkg2.StrTurnUint(friendID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func GetFriendInfoHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	friendID := c.Query("friend_id")
	if friendID == "" {
		c.JSON(http.StatusBadRequest, "好友ID不能为空")
		return
	}

	response, err := logic.GetFriendInfoLogic(userID.(uint64), pkg2.StrTurnUint(friendID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func SearchUserHandler(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, "email不能为空")
		return
	}

	response, err := logic.SearchUserLogic(email)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}
