package main

import (
	"cc.tim/client/db"
	"cc.tim/client/kafka"
	model2 "cc.tim/client/model"
	"cc.tim/client/pkg"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func CMysqlCommit() {
	handler := func(body []byte) error {
		var msgs []model2.KafkaMysqlMsg
		if err := json.Unmarshal(body, &msgs); err != nil {
			return err
		}
		tx, err := db.MysqlDb.Begin()
		if err != nil {
			return err
		}

		for _, msg := range msgs {
			if _, err := tx.Exec(msg.Sqls, msg.Args...); err != nil {
				_ = tx.Rollback()
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}
		return nil
	}
	kafka.Consume("mysql_topic", handler)
}

func CMsg() {
	handler := func(body []byte) error {
		var inmsg model2.KafkaINMsg
		var outmsg model2.KafkaOUtMsg

		json.Unmarshal(body, &inmsg)
		var err error
		var deviceid []uint64
		if inmsg.MsgType != 3 {
			deviceid, err = getdeviceid(inmsg.ConversationID)
			if err != nil {
				return err
			}
		}
		switch inmsg.MsgType {
		case model2.TEXT:
			var t model2.TextMessage
			tx, err := db.MysqlDb.Begin()
			if err != nil {
				return err
			}
			var seq uint64
			err = tx.QueryRow("SELECT seq FROM new_seq WHERE conversation_id = ?", inmsg.ConversationID).Scan(&seq)
			if err != nil {
				_ = tx.Rollback()
				return err
			}

			t.Content = inmsg.Content
			t.Seq = seq + 1
			t.FromUser = inmsg.FromUser
			t.ConversationID = inmsg.ConversationID
			t.CreatedAt = inmsg.CreatedAt
			t.MsgType = model2.TEXT
			result, err := db.MessageCollection.InsertOne(context.TODO(), &t)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			_, err = tx.Exec("UPDATE new_seq SET seq = ? WHERE conversation_id = ?", seq+1, inmsg.ConversationID)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			err = tx.Commit()
			if err != nil {
				db.MessageCollection.DeleteOne(context.TODO(), bson.M{"_id": result.InsertedID.(primitive.ObjectID)})
				return err
			}
			outmsg.FromUser = inmsg.FromUser
			outmsg.ConversationID = inmsg.ConversationID
			outmsg.MsgType = model2.TEXT
			outmsg.Content = inmsg.Content
			outmsg.CreatedAt = inmsg.CreatedAt
			outmsg.SendDevice = deviceid
			newmsg, _ := pkg.RemoveFields(outmsg, "FileID", "FileName", "Size")
			k := kafka.NewInstanceConn()
			bodys, err := json.Marshal(&newmsg)
			err = k.SendMessage(bodys)
			if err != nil {
				return err
			}
		case model2.FILE:
			var oname string
			var size uint64
			err = db.MysqlDb.QueryRow("SELECT original_filename,file_size FROM files WHERE id = ?", inmsg.FileID).Scan(&oname, &size)
			if err != nil {
				return err
			}
			var t model2.FileMessage
			tx, err := db.MysqlDb.Begin()
			if err != nil {
				return err
			}
			var seq uint64
			err = tx.QueryRow("SELECT seq FROM new_seq WHERE conversation_id = ?", inmsg.ConversationID).Scan(&seq)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			t.Size = size
			t.FileID = inmsg.FileID
			t.FileName = oname
			t.Seq = seq + 1
			t.FromUser = inmsg.FromUser
			t.ConversationID = inmsg.ConversationID
			t.CreatedAt = inmsg.CreatedAt
			t.MsgType = model2.FILE
			result, err := db.MessageCollection.InsertOne(context.TODO(), &t)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			_, err = tx.Exec("UPDATE new_seq SET seq = ? WHERE conversation_id = ?", seq+1, inmsg.ConversationID)
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			err = tx.Commit()
			if err != nil {
				db.MessageCollection.DeleteOne(context.TODO(), bson.M{"_id": result.InsertedID.(primitive.ObjectID)})
				return err
			}
			outmsg.FromUser = inmsg.FromUser
			outmsg.ConversationID = inmsg.ConversationID
			outmsg.MsgType = model2.FILE
			outmsg.FileName = oname
			outmsg.FileID = inmsg.FileID
			outmsg.Size = size
			outmsg.CreatedAt = inmsg.CreatedAt
			outmsg.SendDevice = deviceid
			newmsg, _ := pkg.RemoveFields(outmsg, "Content")
			k := kafka.NewInstanceConn()
			bodys, err := json.Marshal(&newmsg)
			err = k.SendMessage(bodys)
			if err != nil {
				return err
			}
		case model2.SYSTEM:
			var m model2.KafkaSystemINMsg
			json.Unmarshal(body, &m)
			fmt.Println("系统信息")
			fmt.Println(m)
			rows, err := db.MysqlDb.Query("SELECT id FROM device WHERE user_id = ?", m.UserId)
			if err != nil {
				return err
			}
			defer rows.Close()
			var deviceids []uint64
			for rows.Next() {
				var id uint64
				if err := rows.Scan(&id); err != nil {
					return err
				}
				deviceids = append(deviceids, id)
			}
			var t model2.SystemMessage
			var mm model2.KafkaSystemOUtMsg

			t.UserID = m.UserId
			t.Action = m.Action
			t.CreatedAt = inmsg.CreatedAt
			t.MsgType = model2.SYSTEM
			_, err = db.MessageCollection.InsertOne(context.TODO(), &t)
			if err != nil {
				return err
			}
			mm.MsgType = model2.SYSTEM
			mm.CreatedAt = inmsg.CreatedAt
			mm.SendDevice = deviceids
			mm.Action = m.Action
			k := kafka.NewInstanceConn()
			bodys, err := json.Marshal(&mm)
			err = k.SendMessage(bodys)
			if err != nil {
				return err
			}
		}
		return nil
	}
	kafka.Consume("msg_topic", handler)
}

func COffline() {
	handler := func(body []byte) error {
		var m model2.KafkaofflineSeq
		if err := json.Unmarshal(body, &m); err != nil {
			return err
		}

		var errs []error
		tx, err := db.MysqlDb.Begin()
		if err != nil {
			return err
		}
		for _, deviceID := range m.DeivceId {
			var seq uint64
			var ok bool
			tx.QueryRow("SELECT EXISTS(SELECT 1 FROM offline WHERE device_id = ?)", deviceID).Scan(&ok)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if ok {
				continue
			}
			err = tx.QueryRow(`SELECT seq FROM new_seq WHERE conversation_id = ?`, m.ConversationID).Scan(&seq)
			if err != nil && err != sql.ErrNoRows {
				_ = tx.Rollback()
				errs = append(errs, err)
				continue
			}
			if seq == 0 {
				seq = 1
			}
			_, err = tx.Exec("INSERT INTO offline (device_id, seq, conversation_id) VALUES (?,?,?)", deviceID, seq, m.ConversationID)
			if err != nil {
				_ = tx.Rollback()
				errs = append(errs, err)
				continue
			}

		}
		if err := tx.Commit(); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("multiple errors: %v", errs)
		}
		return nil
	}

	kafka.Consume("offline_topic", handler)
}

func CDelete() {
	handler := func(body []byte) error {
		var m struct {
			ConversationID uint64 `json:"conversation_id"`
		}
		err := json.Unmarshal(body, &m)
		if err != nil {
			return err
		}
		filter := bson.M{"conversation_id": m.ConversationID}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err = db.MessageCollection.DeleteMany(ctx, filter)
		if err != nil {
			return err
		}
		_, err = db.MysqlDb.Exec("DELETE FROM new_seq WHERE conversation_id = ?", m.ConversationID)
		_, err = db.MysqlDb.Exec("DELETE FROM offline WHERE conversation_id = ?", m.ConversationID)
		if err != nil {
			return err
		}
		return nil

	}

	kafka.Consume("delete_topic", handler)
}
