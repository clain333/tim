package model

import "encoding/json"

type SafeArgs []interface{}

func (a *SafeArgs) UnmarshalJSON(data []byte) error {
	var raw []interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for _, v := range raw {
		switch val := v.(type) {
		case float64:
			// 这里转成 int64，避免 e+18 精度丢失
			*a = append(*a, int64(val))
		default:
			*a = append(*a, val)
		}
	}
	return nil
}

type KafkaMysqlMsg struct {
	Sqls string   `json:"sqls"`
	Args SafeArgs `json:"args"`
}

func NewKafkaMysqlMsg(sql string, args ...interface{}) *KafkaMysqlMsg {
	return &KafkaMysqlMsg{sql, args}
}

type KafkaINMsg struct {
	MsgType        uint8  `json:"msg_type"`
	ConversationID uint64 `json:"conversation_id"`
	FromUser       uint64 `json:"from_user"`
	FileID         uint64 `json:"file_id"`
	Content        string `json:"content"`
}
type KafkaOUtMsg struct {
	SendDevice     []uint64 `json:"send_device"`
	ConversationID uint64   `json:"conversation_id"`
	FromUser       uint64   `json:"from_user"`
	MsgType        uint8    `json:"msg_type" `
	FileID         uint64   `json:"file_id" `
	FileName       string   `json:"file_name"`
	Size           uint64   `json:"size" `
	Content        string   `json:"content"`
	CreatedAt      string   `json:"created_at"`
}

type KafkaSystemINMsg struct {
	UserId uint64 `json:"user_id"`
	Action string `json:"action"`
}
type KafkaSystemOUtMsg struct {
	SendDevice []uint64 `json:"send_device"`
	Userid     uint64   `json:"user_id"`
	MsgType    uint8    `json:"msg_type" ` // 固定值 3
	Action     string   `json:"action" `
	CreatedAt  string   `json:"created_at"`
}

type KafkaofflineSeq struct {
	DeivceId       []uint64 `json:"deivce_id"`
	ConversationID uint64   `json:"conversation_id"`
}

type KafkaSystemMsg struct {
	MsgType uint8  `json:"msg_type"`
	UserId  uint64 `json:"user_id"`
	Action  string `json:"action"`
}
