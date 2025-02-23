package api

import "encoding/json"

type ApiResponse struct {
	Ok          bool            `json:"ok"`
	Result      json.RawMessage `json:"result,omitempty"`
	ErrorCode   int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
}

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID   int                   `json:"message_id"`
	Sender      *User                 `json:"from,omitempty"`
	Date        int                   `json:"date"`
	Chat        *Chat                 `json:"chat"`
	Text        string                `json:"text,omitempty"`
	Entities    []MessageEntity       `json:"entities,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type User struct {
	ID       int64  `json:"id"`
	UserName string `json:"username,omitempty"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type CallbackQuery struct {
	ID              string   `json:"id"`
	Sender          *User    `json:"from"`
	Message         *Message `json:"message,omitempty"`
	InlineMessageID string   `json:"inline_message_id,omitempty"`
	Data            string   `json:"data,omitempty"`
}

type RequestUpdates struct {
	Offset         int      `json:"offset,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type SendMessage struct {
	ChatID      int64                 `json:"chat_id"`
	Text        string                `json:"text"`
	Entities    []MessageEntity       `json:"entities,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type EditMessageText struct {
	ChatID      int64                 `json:"chat_id"`
	MessageID   int                   `json:"message_id"`
	Text        string                `json:"text"`
	Entities    []MessageEntity       `json:"entities,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type AnswerCallbackQuery struct {
	CallbackQueryId string `json:"callback_query_id"`
	Text            string `json:"text,omitempty"`
	ShowAlert       bool   `json:"show_alert,omitempty"`
}
