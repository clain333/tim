package consumer

import (
	"cc.tim/client/kafka"
	"cc.tim/client/model"
	"cc.tim/client/pkg"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"im.api/internal/ws"
)

func sendToDevices(devices []uint64, payload any) (offline []uint64, errs []error) {
	for _, device := range devices {
		if v, ok := ws.Clients.Load(device); ok {
			if err := v.(*websocket.Conn).WriteJSON(payload); err != nil {
				errs = append(errs, fmt.Errorf("device %d write error: %w", device, err))
				offline = append(offline, device)
			}
		} else {
			offline = append(offline, device)
		}
	}
	return
}

func pushOffline(conversationID uint64, devices []uint64) error {
	if len(devices) == 0 {
		return nil
	}
	k := kafka.NewInstanceOffline()
	body, err := json.Marshal(&model.KafkaofflineSeq{
		ConversationID: conversationID,
		DeivceId:       devices,
	})
	if err != nil {
		return fmt.Errorf("marshal offline msg error: %w", err)
	}
	k.SendMessage(body)
	return nil
}

func CConn() {
	handler := func(body []byte) error {
		var msg model.KafkaOUtMsg
		if err := json.Unmarshal(body, &msg); err != nil {
			return fmt.Errorf("unmarshal probe error: %w", err)
		}

		var errs []error

		switch msg.MsgType {
		case model.TEXT:
			newmsg, _ := pkg.RemoveFields(msg, "FileID", "FileName", "Size", "SendDevice")

			offline, es := sendToDevices(msg.SendDevice, newmsg)
			errs = append(errs, es...)
			if err := pushOffline(msg.ConversationID, offline); err != nil {
				errs = append(errs, err)
			}
		case model.FILE:
			newmsg, _ := pkg.RemoveFields(msg, "Content", "SendDevice")

			offline, es := sendToDevices(msg.SendDevice, newmsg)
			errs = append(errs, es...)
			if err := pushOffline(msg.ConversationID, offline); err != nil {
				errs = append(errs, err)
			}

		case model.SYSTEM:
			var m model.KafkaSystemOUtMsg
			if err := json.Unmarshal(body, &m); err != nil {
				return fmt.Errorf("unmarshal KafkaSystemOUtMsg error: %w", err)
			}
			newmsg, _ := pkg.RemoveFields(m, "SendDevice")

			_, es := sendToDevices(m.SendDevice, newmsg)
			errs = append(errs, es...)

		default:
			return fmt.Errorf("未知消息类型: %d", msg.MsgType)
		}

		if len(errs) > 0 {
			return fmt.Errorf("处理过程中出现错误: %v", errs)
		}
		return nil
	}

	kafka.Consume("conn_topic", handler)
}
