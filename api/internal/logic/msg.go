package logic

import (
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	"cc.tim/client/model"
	pkg2 "cc.tim/client/pkg"
	"cc.tim/client/redis"
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"sort"
	"time"
)

func OfflineMsgLogic(deviceid uint64) (*pkg2.Response, error) {
	type offl struct {
		conversationid uint64
		seq            uint64
	}
	var offlines []offl

	rows, err := db.MysqlDb.Query("SELECT conversation_id, seq FROM offline WHERE device_id = ?", deviceid)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取失败", nil), err
	}
	defer rows.Close()

	for rows.Next() {
		var cid, seq uint64
		if err := rows.Scan(&cid, &seq); err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "获取失败", nil), err
		}
		offlines = append(offlines, offl{conversationid: cid, seq: seq})
	}

	t := kafka.NewInstanceMysql()
	err = t.SendMysqlMessage([]*model.KafkaMysqlMsg{
		model.NewKafkaMysqlMsg("DELETE FROM offline WHERE device_id = ?", deviceid),
	})
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取失败", nil), err
	}

	if len(offlines) == 0 {
		return pkg2.NewResponse(http.StatusOK, "无离线消息", nil), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 用于最终返回的结构
	var Msgs []map[string]interface{}

	for _, v := range offlines {
		// 查询文本消息
		textFilter := bson.M{
			"conversation_id": v.conversationid,
			"seq":             bson.M{"$gte": v.seq},
		}
		cursor, err := db.MessageCollection.Find(ctx, textFilter)
		if err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "获取文本消息失败", nil), err
		}
		var results []model.Message
		if err := cursor.All(ctx, &results); err != nil {
			return pkg2.NewResponse(http.StatusInternalServerError, "获取文本消息失败", nil), err
		}
		cursor.Close(ctx)

		for _, msg := range results {
			if msg.MsgType == model.TEXT {
				Msgs = append(Msgs, map[string]interface{}{
					"conversation_id": msg.ConversationID,
					"from_user":       msg.FromUser,
					"msg_type":        msg.MsgType,
					"created_at":      msg.CreatedAt,
					"content":         msg.Content,
				})
			} else {
				Msgs = append(Msgs, map[string]interface{}{
					"conversation_id": msg.ConversationID,
					"from_user":       msg.FromUser,
					"msg_type":        msg.MsgType,
					"content":         msg.Content,
					"created_at":      msg.CreatedAt,
					"file_id":         msg.FileID,
					"file_name":       msg.FileName,
					"size":            msg.Size,
				})
			}

		}
	}
	result := map[string]interface{}{
		"messages": Msgs,
	}

	return pkg2.NewResponse(http.StatusOK, "获取成功", result), nil
}

func MsgLogic(conversationid, num, page uint64) (*pkg2.Response, error) {
	if num == 0 {
		num = 20 // 默认每页 20 条
	}
	if page == 0 {
		page = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 一次查所有消息（文本+文件+系统）
	filter := bson.M{
		"conversation_id": conversationid,
	}
	cursor, err := db.MessageCollection.Find(ctx, filter)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "获取消息失败", nil), err
	}

	var results []model.Message
	if err := cursor.All(ctx, &results); err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "解码消息失败", nil), err
	}
	cursor.Close(ctx)

	// 统一组装
	var allMessages []map[string]interface{}
	for _, msg := range results {
		switch msg.MsgType {
		case model.TEXT: // 文本消息
			allMessages = append(allMessages, map[string]interface{}{
				"seq":             msg.Seq,
				"msg_type":        msg.MsgType,
				"conversation_id": msg.ConversationID,
				"from_user":       msg.FromUser,
				"content":         msg.Content,
				"created_at":      msg.CreatedAt,
			})
		case model.FILE: // 文件消息
			allMessages = append(allMessages, map[string]interface{}{
				"seq":             msg.Seq,
				"msg_type":        msg.MsgType,
				"conversation_id": msg.ConversationID,
				"from_user":       msg.FromUser,
				"file_id":         msg.FileID,
				"file_name":       msg.FileName,
				"size":            msg.Size,
				"created_at":      msg.CreatedAt,
			})
		}
	}

	// 按 seq 排序
	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i]["seq"].(uint64) < allMessages[j]["seq"].(uint64)
	})

	start := int((page - 1) * num)
	end := start + int(num)
	if start >= len(allMessages) {
		return pkg2.NewResponse(http.StatusOK, "没有更多消息", map[string]interface{}{
			"page":     page,
			"num":      num,
			"total":    len(allMessages),
			"messages": []interface{}{},
		}), nil
	}
	if end > len(allMessages) {
		end = len(allMessages)
	}

	result := map[string]interface{}{
		"page":     page,
		"num":      num,
		"total":    len(allMessages),
		"messages": allMessages[start:end],
	}

	return pkg2.NewResponse(http.StatusOK, "获取成功", result), nil
}

func SendMsgLogic(userid uint64, msg model.KafkaINMsg) (*pkg2.Response, error) {
	if msg.FromUser != userid {
		return pkg2.NewResponse(http.StatusBadRequest, "发送失败", nil), nil
	}
	if msg.MsgType != model.TEXT && msg.MsgType != model.FILE {
		return pkg2.NewResponse(http.StatusBadRequest, "发送失败", nil), nil
	}
	var ok bool
	err := db.MysqlDb.QueryRow("SELECT EXISTS(SELECT 1 FROM new_seq WHERE conversation_id = ?)", msg.ConversationID).Scan(&ok)
	if err != nil {
		return pkg2.NewResponse(http.StatusBadRequest, "发送失败", nil), nil
	}
	if !ok {
		return pkg2.NewResponse(http.StatusBadRequest, "发送失败", nil), nil
	}

	switch msg.MsgType {
	case model.TEXT:
		if msg.Content == "" {
			return pkg2.NewResponse(http.StatusBadRequest, "发送失败,msg不能为空", nil), nil
		}
	case model.FILE:
		if msg.FileID == 0 {
			return pkg2.NewResponse(http.StatusBadRequest, "发送失败,file_id不能为空", nil), nil
		}
		filekey := fmt.Sprintf(model.FileKEY, msg.FileID)
		count := redis.RedisClient.Exists(redis.Ctx, filekey).Val()
		if count == 0 {
			return pkg2.NewResponse(http.StatusBadRequest, "发送失败", nil), nil
		}
		redis.RedisClient.Del(redis.Ctx, filekey)
	}
	f := kafka.NewInstanceMsg()
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	err = f.SendMessage(body)
	if err != nil {
		return pkg2.NewResponse(http.StatusInternalServerError, "发送失败", nil), nil
	}
	return pkg2.NewResponse(http.StatusOK, "发送成功", nil), nil
}
