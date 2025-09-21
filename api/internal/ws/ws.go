package ws

import (
	"cc.tim/client/model"
	"cc.tim/client/redis"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

var Clients sync.Map

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 10 * time.Second,
}

func WsHandler(c *gin.Context) {
	deviceid, _ := c.Get("device_id")
	if _, ok := Clients.Load(deviceid); ok {
		c.JSON(http.StatusBadRequest, "该设备已在其他地方登录")
		return
	}
	userid, _ := c.Get("user_id")
	userkey := fmt.Sprintf(model.UserDeviceKEY, userid)
	exists, err := redis.RedisClient.SIsMember(redis.Ctx, userkey, deviceid).Result()
	if !exists {
		c.JSON(http.StatusBadRequest, "请先登录")
		return
	}
	sessionkey := fmt.Sprintf(model.DeviceKEY, deviceid)
	redis.RedisClient.Set(redis.Ctx, sessionkey, "1", 12*time.Hour)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, "请求头错误")
		return
	}
	defer conn.Close()

	conn.SetReadLimit(256 * 1024)
	pingInterval := 30 * time.Second
	Clients.Store(deviceid, conn)

	ticker := time.NewTicker(pingInterval)
	for range ticker.C {
		err := conn.WriteMessage(1, []byte{1})
		if err != nil {
			Clients.Delete(deviceid)
			conn.Close()
			return
		}
	}
}
