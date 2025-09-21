package logic

import (
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	"cc.tim/client/model"
	pkg2 "cc.tim/client/pkg"
	"cc.tim/client/redis"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

type FriendRequest struct {
	Id     uint64 `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Email  string `json:"email"`
}

type FriendList struct {
	Id     uint64 `json:"friend_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type Friend struct {
	ConversationID uint64 `json:"conversation_id"`
	FriendID       uint64 `json:"friend_id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Avatar         string `json:"avatar"`
	Introduction   string `json:"introduction"`
}

// ----------------- 搜索用户 -----------------
func SearchUserLogic(email string) (*pkg2.Response, error) {
	var user FriendList
	err := db.MysqlDb.QueryRow(
		"SELECT id, name, avatar FROM user WHERE email = ?",
		email,
	).Scan(&user.Id, &user.Name, &user.Avatar)

	if err == sql.ErrNoRows {
		return pkg2.NewResponse(http.StatusOK, "用户不存在", nil), nil
	} else if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "查询失败", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "查询成功", user), nil
}

// ----------------- 发送好友请求 -----------------
func SendFriendRequestLogic(userID, friendID uint64) (*pkg2.Response, error) {

	minID, maxID := pkg2.Cpnum(userID, friendID)

	// 检查是否已经是好友
	var exists bool
	err := db.MysqlDb.QueryRow("SELECT EXISTS(SELECT 1 FROM friend WHERE min_id=? AND max_id=?)", minID, maxID).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return pkg2.NewResponse(http.StatusInternalServerError, "发送好友请求失败", nil), err
	}
	if exists {
		return pkg2.NewResponse(http.StatusBadRequest, "你和对方已经是好友", nil), nil
	}

	// 检查是否已经发送过请求
	err = db.MysqlDb.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_request WHERE send_id=? AND receive_id=?)", userID, friendID).Scan(&exists)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "发送好友请求失败", nil), err
	}
	if exists {
		return pkg2.NewResponse(http.StatusBadRequest, "已发送好友请求", nil), nil
	}
	err = db.MysqlDb.QueryRow("SELECT EXISTS(SELECT 1 FROM friend_request WHERE send_id=? AND receive_id=?)", friendID, userID).Scan(&exists)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "发送好友请求失败", nil), err
	}
	if exists {
		return pkg2.NewResponse(http.StatusBadRequest, "对方已发送好友请求给你，请先处理", nil), nil
	}

	// 异步插入好友请求
	t := kafka.NewInstanceMysql()
	err = t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg(
		"INSERT INTO friend_request(send_id, receive_id) VALUES (?, ?)",
		userID, friendID,
	)})
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "发送好友请求失败", nil), nil
	}
	k := kafka.NewInstanceMsg()
	var m = model.KafkaSystemMsg{
		MsgType: model.SYSTEM,
		Action:  "有人发送好友请求",
		UserId:  friendID,
	}
	bb, _ := json.Marshal(&m)
	k.SendMessage(bb)
	return pkg2.NewResponse(http.StatusOK, "好友请求发送成功", nil), nil
}

// ----------------- 获取好友请求列表 -----------------
func GetFriendRequestsLogic(userID uint64) (*pkg2.Response, error) {
	rows, err := db.MysqlDb.Query(`
		SELECT u.id, u.name, u.avatar, u.email
		FROM friend_request f
		JOIN user u ON f.send_id = u.id
		WHERE f.receive_id = ?`, userID)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取好友请求列表失败", nil), err
	}
	defer rows.Close()

	var requests []FriendRequest
	for rows.Next() {
		var fr FriendRequest
		if err := rows.Scan(&fr.Id, &fr.Name, &fr.Avatar, &fr.Email); err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "获取好友请求列表失败", nil), err
		}
		requests = append(requests, fr)
	}

	if err := rows.Err(); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取好友请求列表失败", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "获取好友请求列表成功", requests), nil
}

// ----------------- 处理好友请求 -----------------
func HandleFriendRequestLogic(userID, friendID uint64, action int8) (*pkg2.Response, error) {
	var requestID uint64
	err := db.MysqlDb.QueryRow(
		"SELECT id FROM friend_request WHERE send_id = ? AND receive_id = ?",
		friendID, userID,
	).Scan(&requestID)

	if err != nil {
		if err == sql.ErrNoRows {
			return pkg2.NewResponse(http.StatusBadRequest, "没有该好友请求", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "处理失败", nil), err
	}

	switch action {
	case 1:
		friendRecordID := uint64(pkg2.SnowflakeNode.Generate().Int64())
		conversationID := uint64(pkg2.SnowflakeNode.Generate().Int64())
		minID, maxID := pkg2.Cpnum(userID, friendID)
		t := kafka.NewInstanceMysql()
		err = t.SendMysqlMessage([]*model.KafkaMysqlMsg{
			model.NewKafkaMysqlMsg("INSERT INTO friend(id, conversation_id,min_id, max_id) VALUES(?,?,?,?)", friendRecordID, conversationID, minID, maxID),
			model.NewKafkaMysqlMsg("INSERT INTO new_seq (conversation_id) VALUES (?)", conversationID),
			model.NewKafkaMysqlMsg("DELETE FROM friend_request WHERE id=?", requestID),
		})
		if err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "处理失败", nil), err
		}
		k := kafka.NewInstanceMsg()
		var m = model.KafkaSystemMsg{
			MsgType: model.SYSTEM,
			Action:  "有人同意了你的好友请求",
			UserId:  friendID,
		}
		bb, _ := json.Marshal(&m)
		k.SendMessage(bb)
		return pkg2.NewResponse(http.StatusOK, "好友请求已接受", nil), nil

	case 2: // 拒绝
		t := kafka.NewInstanceMysql()
		t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg("DELETE FROM friend_request WHERE id=?", requestID)})
		k := kafka.NewInstanceMsg()
		var m = model.KafkaSystemMsg{
			MsgType: model.SYSTEM,
			Action:  "有人拒绝了你的好友请求",
			UserId:  friendID,
		}
		bb, _ := json.Marshal(&m)
		k.SendMessage(bb)
		return pkg2.NewResponse(http.StatusOK, "好友请求已拒绝", nil), nil

	default:
		return pkg2.NewResponse(http.StatusBadRequest, "无效操作", nil), nil
	}
}

// ----------------- 删除好友 -----------------
func DeleteFriendLogic(userID, friendID uint64) (*pkg2.Response, error) {
	minID, maxID := pkg2.Cpnum(userID, friendID)
	var conversationid uint64
	err := db.MysqlDb.QueryRow("SELECT conversation_id FROM friend WHERE min_id=? AND max_id=?", minID, maxID).Scan(&conversationid)
	if err != nil && err != sql.ErrNoRows {
		return pkg2.NewResponse(http.StatusInternalServerError, "删除好友失败", nil), err
	}
	if err == sql.ErrNoRows {
		return pkg2.NewResponse(http.StatusBadRequest, "对方不是你的好友", nil), err
	}
	t := kafka.NewInstanceMysql()
	err = t.SendMysqlMessage([]*model.KafkaMysqlMsg{model.NewKafkaMysqlMsg("DELETE FROM friend WHERE min_id = ? AND max_id = ?", minID, maxID)})
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "删除好友失败", nil), err
	}
	k := kafka.NewInstanceDelete()
	var c = struct {
		ConversationID uint64 `json:"conversation_id"`
	}{
		ConversationID: conversationid,
	}
	conversationkey := fmt.Sprintf(model.ConversationKEY, conversationid)
	redis.RedisClient.Del(redis.Ctx, conversationkey)
	b, err := json.Marshal(&c)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "删除好友失败", nil), err
	}
	k.SendMessage(b)
	return pkg2.NewResponse(http.StatusOK, "好友已删除", nil), nil
}

// ----------------- 获取好友列表 -----------------
func GetFriendListLogic(userID uint64) (*pkg2.Response, error) {
	query := `
		SELECT u.id, u.name, u.avatar
		FROM friend f
		JOIN user u ON u.id = CASE WHEN f.min_id=? THEN f.max_id ELSE f.min_id END
		WHERE ? IN (f.min_id, f.max_id);
	`
	rows, err := db.MysqlDb.Query(query, userID, userID)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "查询好友列表失败", nil), err
	}
	defer rows.Close()

	var friends []FriendList
	for rows.Next() {
		var f FriendList
		if err := rows.Scan(&f.Id, &f.Name, &f.Avatar); err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "扫描好友数据失败", nil), err
		}
		friends = append(friends, f)
	}

	return pkg2.NewResponse(http.StatusOK, "获取好友列表成功", friends), nil
}

// ----------------- 获取好友信息 -----------------
func GetFriendInfoLogic(userID, friendID uint64) (*pkg2.Response, error) {
	minID, maxID := pkg2.Cpnum(userID, friendID)

	sqlStr := `
		SELECT f.conversation_id, u.id, u.name, u.email, u.avatar, u.introduction
		FROM friend f
		JOIN user u ON u.id= ?
		WHERE f.min_id=? AND f.max_id=?;
	`
	var f Friend
	err := db.MysqlDb.QueryRow(sqlStr, friendID, minID, maxID).Scan(&f.ConversationID, &f.FriendID, &f.Name, &f.Email, &f.Avatar, &f.Introduction)
	if err != nil {
		if err == sql.ErrNoRows {
			return pkg2.NewResponse(http.StatusBadRequest, "好友不存在", nil), nil
		}
		return pkg2.NewResponse(http.StatusInternalServerError, "获取失败", nil), err
	}

	return pkg2.NewResponse(http.StatusOK, "获取成功", f), nil
}
