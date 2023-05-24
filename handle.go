package main

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gogpt "github.com/sashabaranov/go-openai"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/u2takey/go-utils/uuid"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func downloadFile(url string) (*os.File, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp("", "tg-gpt-*.oga")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	_, err = file.Write(bs)
	if err != nil {
		return nil, err
	}
	mp3Name := fmt.Sprintf("%s.mp3", uuid.NewUUID())
	if err := ffmpeg.Input(file.Name()).Output(mp3Name).OverWriteOutput().Run(); err != nil {
		return nil, err
	}
	return os.Open(mp3Name)
}

func a2t(bot *tgbotapi.BotAPI, fileId string) (string, bool, error) {
	fileUrl, _ := bot.GetFileDirectURL(fileId)
	file, err := downloadFile(fileUrl)
	if err != nil {
		return "", false, err
	}
	defer os.Remove(file.Name())
	c := gogpt.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	req := gogpt.AudioRequest{
		Model:    gogpt.Whisper1,
		FilePath: file.Name(),
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return "", false, err
	}
	return resp.Text, true, err
}

func handleUserPrompt(userID int64, msg string) (string, bool, error) {
	clearUserContextIfExpires(userID)

	if _, ok := users[userID]; !ok {
		users[userID] = &User{
			TelegramID:     userID,
			LastActiveTime: time.Now(),
			HistoryMessage: []gogpt.ChatCompletionMessage{},
		}
	}

	users[userID].HistoryMessage = append(users[userID].HistoryMessage, gogpt.ChatCompletionMessage{
		Role:    "user",
		Content: msg,
	})
	users[userID].LastActiveTime = time.Now()

	c := gogpt.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	req := gogpt.ChatCompletionRequest{
		Model:       gogpt.GPT3Dot5Turbo,
		Temperature: cfg.ModelTemperature,
		TopP:        1,
		N:           1,
		// PresencePenalty:  0.2,
		// FrequencyPenalty: 0.2,
		Messages: users[userID].HistoryMessage,
	}

	fmt.Println(req)

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Print(err)
		users[userID].HistoryMessage = users[userID].HistoryMessage[:len(users[userID].HistoryMessage)-1]
		return "", false, err
	}

	answer := resp.Choices[0].Message

	users[userID].HistoryMessage = append(users[userID].HistoryMessage, answer)

	var contextTrimmed bool
	if resp.Usage.TotalTokens > 3500 {
		users[userID].HistoryMessage = users[userID].HistoryMessage[1:]
		contextTrimmed = true
	}

	return answer.Content, contextTrimmed, nil
}
