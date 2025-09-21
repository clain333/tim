package logic

import (
	"cc.tim/client/config"
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	"cc.tim/client/model"
	pkg2 "cc.tim/client/pkg"
	r "cc.tim/client/redis"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"im.api/internal/ws"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

// ------------------------ 用户模块 ------------------------

func GetUserInfoLogic(userid uint64) (*pkg2.Response, error) {
	user := struct {
		ID           uint64 `json:"id"`
		Name         string `json:"name"`
		Email        string `json:"email,omitempty"`
		Avatar       string `json:"avatar,omitempty"`
		Introduction string `json:"introduction,omitempty"`
		CreateTime   string `json:"create_time"`
		UpdateTime   string `json:"update_time"`
	}{}
	var CreateTimes time.Time
	var UpdateTimes time.Time
	err := db.MysqlDb.QueryRow(
		"SELECT id,name,email,avatar,introduction,create_time,update_time FROM user WHERE id=?",
		userid,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Avatar, &user.Introduction, &CreateTimes, &UpdateTimes)

	if errors.Is(err, sql.ErrNoRows) {
		return pkg2.NewResponse(http.StatusBadRequest, "用户不存在", nil), nil
	}
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), err
	}
	user.CreateTime = pkg2.TimeTrunStr(CreateTimes)
	user.UpdateTime = pkg2.TimeTrunStr(UpdateTimes)
	return pkg2.NewResponse(http.StatusOK, "获取成功", user), nil
}

func UpdateUserInfoLogic(userid uint64, name, introduction string) (*pkg2.Response, error) {
	t := kafka.NewInstanceMysql()
	err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("UPDATE user SET name=?,introduction=? WHERE id=?", name, introduction, userid),
	})
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "更新失败", nil), err
	}
	return pkg2.NewResponse(http.StatusOK, "更新成功", nil), nil
}

func UploadUserAvatarLogic(c *gin.Context, f *multipart.FileHeader, userid uint64) (*pkg2.Response, error) {
	if f.Size > config.Config.Avatar.MaxSize {
		return pkg2.NewResponse(http.StatusBadRequest, "文件过大", nil), nil
	}

	ext := filepath.Ext(f.Filename)
	allowed := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}
	if !allowed[ext] {
		return pkg2.NewResponse(http.StatusBadRequest, "不支持的文件格式", nil), nil
	}

	filename := fmt.Sprintf("%d%s", userid, ext)
	dst := filepath.Clean(filepath.Join(config.Config.Avatar.Src, filename))

	if err := c.SaveUploadedFile(f, dst); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "保存失败", nil), err
	}

	t := kafka.NewInstanceMysql()
	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("UPDATE user SET avatar=? WHERE id=?", filename, userid),
	}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "更新失败", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "上传成功", map[string]string{"avatar": filename}), nil
}

func ChangePasswordLogic(userid uint64, oldPassword, newPassword string) (*pkg2.Response, error) {
	var hpassword string
	if err := db.MysqlDb.QueryRow("SELECT password FROM user WHERE id=?", userid).Scan(&hpassword); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return pkg2.NewResponse(http.StatusUnauthorized, "用户不存在", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "修改失败", nil), err
	}

	if !pkg2.CheckPassword(oldPassword, hpassword) {
		return pkg2.NewResponse(http.StatusBadRequest, "旧密码错误", nil), nil
	}

	newHash, err := pkg2.HashPassword(newPassword)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "修改失败", nil), err
	}

	// 先写数据库，再清理 Redis
	t := kafka.NewInstanceMysql()
	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("UPDATE user SET password=? WHERE id=?", newHash, userid),
	}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "更新失败", nil), err
	}

	// 清除 Redis 登录态
	userkey := fmt.Sprintf(model.UserDeviceKEY, userid)
	deviceIDstr, _ := r.RedisClient.SMembers(r.Ctx, userkey).Result()
	for _, d := range deviceIDstr {
		deviceID := pkg2.StrTurnUint(d)
		r.RedisClient.Del(r.Ctx, fmt.Sprintf(model.DeviceKEY, deviceID))
		if conn, ok := ws.Clients.Load(deviceID); ok {
			conn.(*websocket.Conn).Close()
			ws.Clients.Delete(deviceID)
		}
	}
	r.RedisClient.Del(r.Ctx, userkey)

	return pkg2.NewResponse(http.StatusOK, "密码修改成功", nil), nil
}

func GetDevicesLogic(userid, deviceid uint64) (*pkg2.Response, error) {
	userkey := fmt.Sprintf(model.UserDeviceKEY, userid)
	result, err := r.RedisClient.SMembers(r.Ctx, userkey).Result()
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取失败", nil), err
	}

	devices := make([]struct {
		DeviceID      uint64 `json:"device_id"`
		LastLoginTime string `json:"last_login_time"`
		IP            string `json:"ip"`
		OS            string `json:"os"`
		DeviceTag     string `json:"device_tag"`
		IsCurrent     bool   `json:"is_current"`
	}, 0, len(result))

	for _, idStr := range result {
		id := pkg2.StrTurnUint(idStr)
		var ip, os, deviceTag string
		var loginTime time.Time
		if err := db.MysqlDb.QueryRow(
			"SELECT ip,os,device_tag,login_time FROM device WHERE id=?",
			id,
		).Scan(&ip, &os, &deviceTag, &loginTime); err == nil {
			devices = append(devices, struct {
				DeviceID      uint64 `json:"device_id"`
				LastLoginTime string `json:"last_login_time"`
				IP            string `json:"ip"`
				OS            string `json:"os"`
				DeviceTag     string `json:"device_tag"`
				IsCurrent     bool   `json:"is_current"`
			}{
				DeviceID:      id,
				LastLoginTime: pkg2.TimeTrunStr(loginTime),
				IP:            ip,
				OS:            os,
				DeviceTag:     deviceTag,
				IsCurrent:     id == deviceid,
			})
		}
	}

	return pkg2.NewResponse(http.StatusOK, "获取成功", devices), nil
}

func LogoutDeviceLogic(userid, deviceid uint64) (*pkg2.Response, error) {
	userkey := fmt.Sprintf(model.UserDeviceKEY, userid)
	devicekey := fmt.Sprintf(model.DeviceKEY, deviceid)

	removed, err := r.RedisClient.SRem(r.Ctx, userkey, deviceid).Result()
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "退出失败", nil), err
	}
	if removed == 0 {
		return pkg2.NewResponse(http.StatusBadRequest, "设备不存在", nil), nil
	}

	r.RedisClient.Del(r.Ctx, devicekey)
	if value, ok := ws.Clients.Load(deviceid); ok {
		value.(*websocket.Conn).Close()
		ws.Clients.Delete(deviceid)
	}

	return pkg2.NewResponse(http.StatusOK, "设备退出成功", nil), nil
}
