package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gogpt "github.com/sashabaranov/go-openai"
	"time"
)

type User struct {
	TelegramID     int64
	LastActiveTime time.Time
	HistoryMessage []gogpt.ChatCompletionMessage
	LatestMessage  tgbotapi.Message
}

var users = make(map[int64]*User)
