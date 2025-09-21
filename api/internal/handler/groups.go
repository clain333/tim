package handler

import (
	"cc.tim/client/logger"
	pkg2 "cc.tim/client/pkg"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"im.api/internal/logic"
	"net/http"
)

func GetGroupByNumberHandler(c *gin.Context) {
	groupNumber := c.Query("number")
	if groupNumber == "" {
		c.JSON(http.StatusBadRequest, "number不能为空")
		return
	}

	response, err := logic.GetGroupByNumberLogic(groupNumber)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func CreateGroupHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusUnauthorized, "name不能为空")
		return
	}
	response, err := logic.CreateGroupLogic(userID.(uint64), name)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func GetUserGroupsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	response, err := logic.GetUserGroupsLogic(userID.(uint64))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func SendJoinGroupRequestHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}

	response, err := logic.SendJoinGroupRequestLogic(userID.(uint64), pkg2.StrTurnUint(groupID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func LeaveGroupHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}

	response, err := logic.LeaveGroupLogic(userID.(uint64), pkg2.StrTurnUint(groupID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func GetGroupInfoHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, "未授权访问")
		return
	}

	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}

	response, err := logic.GetGroupInfoLogic(userID.(uint64), pkg2.StrTurnUint(groupID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func UpdateGroupInfoHandler(c *gin.Context) {
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}
	var req logic.GroupUpdate
	c.ShouldBindJSON(&req)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, "name不能为空")
		return
	}
	response, err := logic.UpdateGroupInfoLogic(pkg2.StrTurnUint(groupID), req)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func UpdateGroupAvatarHandler(c *gin.Context) {
	groupID := c.Query("group_id")

	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, "头像上传失败")
		return

	}

	response, err := logic.UpdateGroupAvatarLogic(pkg2.StrTurnUint(groupID), c, file)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func HandleJoinGroupRequestHandler(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, "user_id不能为空")
		return
	}
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}
	action := c.Query("action")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "action不能为空")
		return
	}

	response, err := logic.HandleJoinGroupRequestLogic(pkg2.StrTurnUint(userID), pkg2.StrTurnUint(groupID), int8(pkg2.StrTurnUint(action)))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func KickGroupMemberHandler(c *gin.Context) {
	groupID := c.Query("group_id")
	userID := c.Query("user_id")
	if groupID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, "group_id或user_id不能为空")
		return
	}

	response, err := logic.KickGroupMemberLogic(pkg2.StrTurnUint(groupID), pkg2.StrTurnUint(userID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func DissolveGroupHandler(c *gin.Context) {
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}

	response, err := logic.DissolveGroupLogic(pkg2.StrTurnUint(groupID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func GetGroupsRequestsHandler(c *gin.Context) {
	groupID := c.Query("group_id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, "group_id不能为空")
		return
	}
	response, err := logic.GetGroupsRequestsLogic(pkg2.StrTurnUint(groupID))
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}
