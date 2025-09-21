package logic

import (
	"cc.tim/client/config"
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	"cc.tim/client/model"
	pkg2 "cc.tim/client/pkg"
	"cc.tim/client/redis"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type GroupRequest struct {
	Id     uint64 `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Email  string `json:"email"`
}

type GroupingList struct {
	Id     uint64 `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Grouping struct {
	Id             uint64         `json:"id"`
	ConversationId uint64         `json:"conversation_id"`
	Name           string         `json:"name"`
	MemberCount    uint64         `json:"member_count"`
	Description    string         `json:"description"`
	Avatar         string         `json:"avatar"`
	GroupOwner     uint64         `json:"group_owner"`
	Members        []GroupingList `json:"members"`
}

type GroupUpdate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetGroupByNumberLogic 根据群号码查询群信息
func GetGroupByNumberLogic(number string) (*pkg2.Response, error) {
	var g GroupingList
	err := db.MysqlDb.QueryRow(`SELECT id, name, avatar_url FROM chat_group WHERE number = ?`, number).
		Scan(&g.Id, &g.Name, &g.Avatar)
	if err != nil {
		if err == sql.ErrNoRows {
			return pkg2.NewResponse(http.StatusOK, "找不到该群组", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "查询群组失败", nil), err
	}
	return pkg2.NewResponse(http.StatusOK, "获取群信息成功", g), nil
}

// CreateGroupLogic 创建群组
func CreateGroupLogic(userID uint64, name string) (*pkg2.Response, error) {
	description := "该群未写简介"
	avatar := "default.jpg"

	groupID := uint64(pkg2.SnowflakeNode.Generate().Int64())
	conversationID := uint64(pkg2.SnowflakeNode.Generate().Int64())
	memberID := uint64(pkg2.SnowflakeNode.Generate().Int64())

	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(90000) + 10000

	t := kafka.NewInstanceMysql()

	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("INSERT INTO chat_group (id, number, conversation_id, name, group_owner, avatar_url, description, member_count) VALUES (?,?,?,?,?,?,?,?)", groupID, strconv.Itoa(number), conversationID, name, userID, avatar, description, 1),
		model.NewKafkaMysqlMsg("INSERT INTO group_member (id, group_id, user_id, role) VALUES (?,?,?,?)", memberID, groupID, userID, 1),
		model.NewKafkaMysqlMsg("INSERT INTO new_seq (conversation_id) VALUES (?)", conversationID),
	}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "群组创建失败", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "群组创建成功", map[string]interface{}{
		"name":   name,
		"number": strconv.Itoa(number),
	}), nil
}

// GetUserGroupsLogic 获取用户加入的群列表
func GetUserGroupsLogic(userID uint64) (*pkg2.Response, error) {
	rows, err := db.MysqlDb.Query(`
		SELECT cg.id, cg.name, cg.avatar_url
		FROM group_member gm 
		JOIN chat_group cg ON cg.id = gm.group_id 
		WHERE gm.user_id = ?`, userID)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取群组失败", nil), err
	}
	defer rows.Close()

	var groups []GroupingList
	for rows.Next() {
		var g GroupingList
		if err := rows.Scan(&g.Id, &g.Name, &g.Avatar); err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "解析群组失败", nil), err
		}
		groups = append(groups, g)
	}
	if len(groups) == 0 {
		return pkg2.NewResponse(http.StatusOK, "用户未加入任何群组", nil), nil
	}
	return pkg2.NewResponse(http.StatusOK, "获取用户群组列表成功", groups), nil
}

// SendJoinGroupRequestLogic 发送加入群请求
func SendJoinGroupRequestLogic(userID, groupID uint64) (*pkg2.Response, error) {
	var exists bool
	err := db.MysqlDb.QueryRow("SELECT EXISTS(SELECT 1 FROM group_member WHERE user_id = ? AND group_id = ?)", userID, groupID).Scan(&exists)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), err
	}
	if exists {
		return pkg2.NewResponse(http.StatusOK, "你已加入群组", nil), nil
	}
	err = db.MysqlDb.QueryRow("SELECT EXISTS(SELECT 1 FROM group_join_request WHERE user_id = ? AND group_id = ?)", userID, groupID).Scan(&exists)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), nil
	}
	if exists {
		return pkg2.NewResponse(http.StatusOK, "已经发送过入群申请", nil), nil

	}

	t := kafka.NewInstanceMysql()
	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg(
		"INSERT INTO group_join_request (id, group_id, user_id) VALUES (?,?,?)",
		uint64(pkg2.SnowflakeNode.Generate().Int64()), groupID, userID,
	)}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "发送请求失败", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "加入群请求已发送", nil), nil
}

// LeaveGroupLogic 退出群组
func LeaveGroupLogic(userID, groupID uint64) (*pkg2.Response, error) {
	var role int
	err := db.MysqlDb.QueryRow("SELECT role FROM group_member WHERE user_id = ? AND group_id = ?", userID, groupID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return pkg2.NewResponse(http.StatusBadRequest, "你不在该群组里", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), err
	}
	if role == 1 {
		return pkg2.NewResponse(http.StatusBadRequest, "群主只能解散群组", nil), nil
	}
	t := kafka.NewInstanceMysql()
	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("DELETE FROM group_member WHERE user_id = ? AND group_id = ?", userID, groupID),
		model.NewKafkaMysqlMsg("UPDATE chat_group SET member_count = member_count - 1 WHERE id = ?", groupID),
	}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "退出群组失败", nil), err
	}
	var conversationID uint64
	db.MysqlDb.QueryRow("SELECT conversation_id FROM chat_group WHERE id = ?", groupID).Scan(&conversationID)
	conversationkey := fmt.Sprintf(model.ConversationKEY, conversationID)
	redis.RedisClient.SRem(redis.Ctx, conversationkey, userID)
	return pkg2.NewResponse(http.StatusOK, "已成功退出群组", nil), nil
}

// 查询群组详细信息
func GetGroupInfoLogic(userID, groupID uint64) (*pkg2.Response, error) {
	var group Grouping

	// 1. 查询群信息（同时检查用户是否在群里）
	err := db.MysqlDb.QueryRow(`
        SELECT 
            cg.id, cg.conversation_id, cg.name, cg.member_count,
            cg.description, cg.avatar_url, cg.group_owner
        FROM chat_group cg
        JOIN group_member gm ON gm.group_id = cg.id
        WHERE gm.group_id = ? AND gm.user_id = ?`,
		groupID, userID).
		Scan(&group.Id, &group.ConversationId, &group.Name, &group.MemberCount,
			&group.Description, &group.Avatar, &group.GroupOwner)

	if err != nil {
		if err == sql.ErrNoRows {
			return pkg2.NewResponse(http.StatusBadRequest, "你不在该群组或群组不存在", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "查询群组信息失败", nil), err
	}

	// 2. 查询群成员列表
	rows, err := db.MysqlDb.Query(`
        SELECT u.id, u.name, u.avatar
        FROM group_member gm
        JOIN user u ON gm.user_id = u.id
        WHERE gm.group_id = ?`, groupID)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "查询群成员失败", nil), err
	}
	defer rows.Close()

	var members []GroupingList
	for rows.Next() {
		var m GroupingList
		if err := rows.Scan(&m.Id, &m.Name, &m.Avatar); err != nil {
			log.Printf("GetGroupInfoLogic members scan error: %v", err)
			continue
		}
		members = append(members, m)
	}
	group.Members = members

	return pkg2.NewResponse(http.StatusOK, "获取群组信息成功", group), nil
}

// UpdateGroupInfoLogic 更新群信息
func UpdateGroupInfoLogic(groupID uint64, req GroupUpdate) (*pkg2.Response, error) {
	t := kafka.NewInstanceMysql()
	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg(
		"UPDATE chat_group SET name = ?, description = ? WHERE id = ?",
		req.Name, req.Description, groupID,
	)}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "更新群信息失败", nil), err
	}
	return pkg2.NewResponse(http.StatusOK, "群信息更新成功", req), nil
}

// UpdateGroupAvatarLogic 更新群头像
func UpdateGroupAvatarLogic(groupID uint64, c *gin.Context, file *multipart.FileHeader) (*pkg2.Response, error) {
	if file.Size > config.Config.Avatar.MaxSize {
		return pkg2.NewResponse(http.StatusBadRequest, "文件大小超限", nil), nil
	}

	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}
	if !allowed[ext] {
		return pkg2.NewResponse(http.StatusBadRequest, "不支持的文件格式", nil), nil
	}

	filename := fmt.Sprintf("%d%s", groupID, ext)
	dst := filepath.Join(config.Config.Avatar.Src, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "保存头像失败", nil), err
	}

	t := kafka.NewInstanceMysql()
	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg("UPDATE chat_group SET avatar_url = ? WHERE id = ?", filename, groupID)}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "更新头像失败", nil), err
	}
	return pkg2.NewResponse(http.StatusOK, "群头像更新成功", map[string]string{"avatar": filename}), nil
}

// HandleJoinGroupRequestLogic 处理加入群请求
func HandleJoinGroupRequestLogic(userID, groupID uint64, action int8) (*pkg2.Response, error) {
	t := kafka.NewInstanceMysql()
	var msgs []*model.KafkaMysqlMsg

	// 删除加入请求
	msgs = append(msgs, model.NewKafkaMysqlMsg("DELETE FROM group_join_request WHERE user_id = ? AND group_id = ?", userID, groupID))

	switch action {
	case 1: // 接受
		msgs = append(msgs, model.NewKafkaMysqlMsg("INSERT INTO group_member (id, group_id, user_id, role) VALUES (?,?,?,?)", uint64(pkg2.SnowflakeNode.Generate().Int64()), groupID, userID, 0),
			model.NewKafkaMysqlMsg("UPDATE chat_group SET member_count = member_count+1 WHERE id = ?", groupID))
	case 2: // 拒绝，不做额外操作
	default:
		return pkg2.NewResponse(http.StatusBadRequest, "无效操作", nil), nil
	}

	if err := t.SendMysqlMessage(msgs); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "处理请求失败", nil), err
	}

	msg := "加入群请求已拒绝"
	if action == 1 {
		msg = "加入群请求已接受"
	}
	return pkg2.NewResponse(http.StatusOK, msg, nil), nil
}

// KickGroupMemberLogic 踢出群成员
func KickGroupMemberLogic(groupID, userID uint64) (*pkg2.Response, error) {
	var role int
	err := db.MysqlDb.QueryRow("SELECT role FROM group_member WHERE group_id = ? AND user_id = ?", groupID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return pkg2.NewResponse(http.StatusBadRequest, "用户不在群组里", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), err
	}
	if role == 1 {
		return pkg2.NewResponse(http.StatusBadRequest, "不能踢出群主", nil), nil
	}

	t := kafka.NewInstanceMysql()
	if err := t.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("DELETE FROM group_member WHERE user_id = ? AND group_id = ?", userID, groupID),
		model.NewKafkaMysqlMsg("UPDATE chat_group SET member_count = member_count - 1 WHERE id = ?", groupID),
	}); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "踢出群组失败", nil), err
	}
	var conversationID uint64
	db.MysqlDb.QueryRow("SELECT conversation_id FROM chat_group WHERE id = ?", groupID).Scan(&conversationID)
	conversationkey := fmt.Sprintf(model.ConversationKEY, conversationID)
	redis.RedisClient.SRem(redis.Ctx, conversationkey, userID)
	return pkg2.NewResponse(http.StatusOK, "成员已被踢出群组", nil), nil
}

// DissolveGroupLogic 解散群组
func DissolveGroupLogic(groupID uint64) (*pkg2.Response, error) {
	t := kafka.NewInstanceMysql()
	msgs := []*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("DELETE FROM group_member WHERE group_id = ?", groupID),
		model.NewKafkaMysqlMsg("DELETE FROM chat_group WHERE id = ?", groupID),
	}
	var conversationid uint64
	err := db.MysqlDb.QueryRow("SELECT conversation_id FROM chat_group WHERE id =?", groupID).Scan(&conversationid)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "删除好友失败", nil), err

	}
	var c = struct {
		ConversationID uint64 `json:"conversation_id"`
	}{
		ConversationID: conversationid,
	}
	b, err := json.Marshal(&c)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "删除好友失败", nil), err
	}
	k := kafka.NewInstanceDelete()
	conversationkey := fmt.Sprintf(model.ConversationKEY, conversationid)
	redis.RedisClient.Del(redis.Ctx, conversationkey)
	err = k.SendMessage(b)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "解散群组失败", nil), err
	}
	if err := t.SendMysqlMessage(msgs); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "解散群组失败", nil), err
	}
	return pkg2.NewResponse(http.StatusOK, "群组已解散", nil), nil
}

// GetGroupsRequestsLogic 获取群加入请求列表
func GetGroupsRequestsLogic(groupID uint64) (*pkg2.Response, error) {
	rows, err := db.MysqlDb.Query(`
		SELECT u.id, u.name, u.email, u.avatar 
		FROM group_join_request g 
		JOIN user u ON g.user_id = u.id 
		WHERE g.group_id = ?`, groupID)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取请求失败", nil), err
	}
	defer rows.Close()

	var list []GroupRequest
	for rows.Next() {
		var r GroupRequest
		if err := rows.Scan(&r.Id, &r.Name, &r.Email, &r.Avatar); err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "解析请求失败", nil), err
		}
		list = append(list, r)
	}
	if len(list) == 0 {
		return pkg2.NewResponse(http.StatusOK, "当前无加入请求", nil), nil
	}
	return pkg2.NewResponse(http.StatusOK, "获取成功", list), nil
}
