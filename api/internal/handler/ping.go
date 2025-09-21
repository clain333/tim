package handler

import (
	"net/http"
	"im.api/internal/logic"
	
	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	response := logic.PingLogic()
	c.JSON(http.StatusOK, response)
}