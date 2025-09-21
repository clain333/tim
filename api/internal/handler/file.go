package handler

import (
	"cc.tim/client/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"im.api/internal/logic"
	"net/http"
)

func CheckFileHashHandler(c *gin.Context) {
	var req struct {
		Hash     string `json:"hash"`
		Filename string `json:"filename"`
	}
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, "结构错误为空")
		return
	}
	if req.Hash == "" {
		c.JSON(http.StatusBadRequest, "hash不能为空")
		return
	}
	if req.Filename == "" {
		c.JSON(http.StatusBadRequest, "filename不能为空")
		return
	}
	response, err := logic.CheckFileHashLogic(req.Hash, req.Filename)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

func UploadFileHandler(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, "file为空")
		return
	}

	filename := c.PostForm("filename")

	if filename == "" {
		c.JSON(http.StatusBadRequest, "filename不能为空")
		return
	}
	response, err := logic.UploadFileLogic(filename, fileHeader)
	if err != nil {
		logger.Error("发生错误", zap.Error(err))
	}
	c.JSON(response.Code, response)
}

// GetFileHandler 处理获取文件的请求
func GetFileHandler(c *gin.Context) {
	fileID := c.Query("fileId")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, "file_id为空")
		return
	}
	logic.GetFileLogic(fileID, c)

}
