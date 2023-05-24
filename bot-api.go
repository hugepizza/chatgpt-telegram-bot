package main

import (
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func transcribeVoice(msg *tgbotapi.Voice) (string, error) {
	if msg == nil {
		return "", errors.New("no voice")
	}
	return "", nil
}
