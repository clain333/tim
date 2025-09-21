package model

type Message struct {
	ID             string `json:"id" bson:"_id,omitempty"`
	ConversationID uint64 `json:"conversation_id" bson:"conversation_id"`
	FromUser       uint64 `json:"from_user" bson:"from_user"`
	Seq            uint64 `json:"seq" bson:"seq"`
	MsgType        uint8  `json:"msg_type" bson:"msg_type"` // 固定值 1
	Content        string `json:"content" bson:"content"`
	CreatedAt      string `json:"created_at" bson:"created_at"`
	FileID         uint64 `json:"file_id" bson:"file_id"`
	FileName       string `json:"file_name" bson:"file_name"`
	Size           uint64 `json:"size" bson:"size"`
}

// 1️⃣ 文本消息
type TextMessage struct {
	ID             string `json:"id" bson:"_id,omitempty"`
	ConversationID uint64 `json:"conversation_id" bson:"conversation_id"`
	FromUser       uint64 `json:"from_user" bson:"from_user"`
	Seq            uint64 `json:"seq" bson:"seq"`
	MsgType        uint8  `json:"msg_type" bson:"msg_type"` // 固定值 1
	Content        string `json:"content" bson:"content"`
	CreatedAt      string `json:"created_at" bson:"created_at"`
}

// 2️⃣ 文件消息
type FileMessage struct {
	ID             string `json:"id" bson:"_id,omitempty"`
	ConversationID uint64 `json:"conversation_id" bson:"conversation_id"`
	FromUser       uint64 `json:"from_user" bson:"from_user"`
	Seq            uint64 `json:"seq" bson:"seq"`
	MsgType        uint8  `json:"msg_type" bson:"msg_type"` // 固定值 2
	FileID         uint64 `json:"file_id" bson:"file_id"`
	FileName       string `json:"file_name" bson:"file_name"`
	Size           uint64 `json:"size" bson:"size"`
	CreatedAt      string `json:"created_at" bson:"created_at"`
}

// 3️⃣ 系统消息
type SystemMessage struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	UserID    uint64 `json:"user_id" bson:"user_id"`
	MsgType   uint8  `json:"msg_type" bson:"msg_type"` // 固定值 3
	Action    string `json:"action" bson:"action"`
	CreatedAt string `json:"created_at" bson:"created_at"`
}
