package middleware

import (
	"cc.tim/client/db"
	"cc.tim/client/model"
	pkg2 "cc.tim/client/pkg"
	r "cc.tim/client/redis"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"strings"
	"time"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{"message": "请求未携带 token"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusOK, gin.H{"message": "请求 token 格式错误"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := pkg2.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "无效或过期的 token"})
			c.Abort()
			return
		}

		// Redis 验证

		ctx := c.Request.Context()
		sessionkey := fmt.Sprintf(model.DeviceKEY, claims.DeviceId)
		_, err = r.RedisClient.Get(ctx, sessionkey).Result()
		if err != nil {
			if err == redis.Nil {
				c.JSON(http.StatusOK, gin.H{"message": "未连接"})
				c.Abort()
				return
			} else {
				c.JSON(http.StatusOK, gin.H{"message": "token 已失效,请重新登录"})
				c.Abort()
				return
			}
		}
		r.RedisClient.Expire(ctx, sessionkey, 12*time.Hour)
		c.Set("user_id", claims.UserId)
		c.Set("device_id", claims.DeviceId)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

func WSJWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{"message": "请求未携带 token"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusOK, gin.H{"message": "请求 token 格式错误"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := pkg2.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "无效或过期的 token"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserId)
		c.Set("device_id", claims.DeviceId)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, "未授权访问")
			c.Abort()
			return
		}

		uid, ok := userID.(uint64)
		if !ok {
			c.JSON(http.StatusUnauthorized, "用户ID类型错误")
			c.Abort()
			return
		}

		groupID := c.Query("group_id")
		if groupID == "" {
			c.JSON(http.StatusBadRequest, "group_id不能为空")
			c.Abort()
			return
		}

		var dummy bool
		err := db.MysqlDb.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM chat_group WHERE id = ? AND group_owner = ? )",
			pkg2.StrTurnUint(groupID),
			uid,
		).Scan(&dummy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "数据库查询失败")
			c.Abort()
			return
		}
		if !dummy {
			c.JSON(http.StatusUnauthorized, "权限不够")
			c.Abort()
			return
		}

		c.Next()
	}
}
