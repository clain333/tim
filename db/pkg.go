package main

import (
	"cc.tim/client/db"
	"cc.tim/client/model"
	"cc.tim/client/pkg"
	"cc.tim/client/redis"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func getdeviceid(conversationID uint64) ([]uint64, error) {
	conversationkey := fmt.Sprintf(model.ConversationKEY, conversationID)
	var userid []uint64
	count := redis.RedisClient.Exists(redis.Ctx, conversationkey).Val()
	if count == 0 {
		err := db.MysqlDb.QueryRow("SELECT 1 FROM friend WHERE conversation_id = ?", conversationID).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if err == nil {
			var (
				minid uint64
				maxid uint64
			)
			db.MysqlDb.QueryRow("SELECT min_id,max_id FROM friend WHERE conversation_id = ?", conversationID).Scan(&minid, &maxid)
			userid = append(userid, minid, maxid)
		} else {
			fmt.Println(conversationID)
			rows, err := db.MysqlDb.Query("SELECT g.user_id FROM group_member g JOIN chat_group c ON g.group_id = c.id WHERE c.conversation_id = ?", conversationID)
			if err != nil {
				return nil, err
			}
			defer rows.Close()
			for rows.Next() {
				var uid uint64
				rows.Scan(&uid)
				userid = append(userid, uid)
			}
		}
		if len(userid) == 0 {
			return nil, errors.New("no user found")
		}
		for _, uid := range userid {
			redis.RedisClient.SAdd(redis.Ctx, conversationkey, uid)
		}
		redis.RedisClient.Expire(redis.Ctx, conversationkey, 24*time.Hour)
	} else {
		useridstr, err := redis.RedisClient.SMembers(redis.Ctx, conversationkey).Result()
		if err != nil {
			return nil, err
		}
		for _, str := range useridstr {
			userid = append(userid, pkg.StrTurnUint(str))
		}
		redis.RedisClient.Expire(redis.Ctx, conversationkey, 24*time.Hour)
	}
	if len(userid) == 0 {
		return nil, errors.New("数据库为空")
	}
	placeholders := make([]string, len(userid))
	args := make([]interface{}, len(userid))
	for i, id := range userid {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("SELECT id FROM device WHERE user_id IN (%s)", strings.Join(placeholders, ","))

	rows, err := db.MysqlDb.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deviceIDs []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		deviceIDs = append(deviceIDs, id)
	}
	return deviceIDs, nil
}
