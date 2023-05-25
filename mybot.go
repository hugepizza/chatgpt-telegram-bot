package main

import (
	tts "cloud.google.com/go/texttospeech/apiv1"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sashabaranov/go-openai"
	"github.com/u2takey/go-utils/uuid"
	"google.golang.org/api/option"
	"log"
	"os"
	"time"
)

type MyBot struct {
	*tgbotapi.BotAPI
	ttsCLi    *tts.Client
	openaiCli *openai.Client

	users map[int64]*User
}

var cfg struct {
	TelegramAPIToken                    string  `env:"TELEGRAM_APITOKEN,required"`
	OpenAIAPIKey                        string  `env:"OPENAI_API_KEY,required"`
	ModelTemperature                    float32 `env:"MODEL_TEMPERATURE" envDefault:"1.0"`
	AllowedTelegramID                   []int64 `env:"ALLOWED_TELEGRAM_ID" envSeparator:","`
	ConversationIdleTimeoutSeconds      int     `env:"CONVERSATION_IDLE_TIMEOUT_SECONDS" envDefault:"86400"`
	NotifyUserOnConversationIdleTimeout bool    `env:"NOTIFY_USER_ON_CONVERSATION_IDLE_TIMEOUT" envDefault:"true"`
	GoogleTTSCertFile                   string  `env:"GOOGLE_TTS_Cert_File,required"`
}

func NewMyBot() (*MyBot, error) {
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramAPIToken)
	if err != nil {
		return nil, err
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	ttsCli, err := tts.NewClient(context.Background(), option.WithCredentialsFile(cfg.GoogleTTSCertFile))
	if err != nil {
		return nil, err
	}

	openaiCli := openai.NewClient(cfg.OpenAIAPIKey)
	if err != nil {
		return nil, err
	}
	return &MyBot{
		BotAPI:    bot,
		ttsCLi:    ttsCli,
		openaiCli: openaiCli,
		users:     map[int64]*User{},
	}, nil
}

func (bot *MyBot) Run() {
	_, _ = bot.Request(tgbotapi.NewSetMyCommands([]tgbotapi.BotCommand{
		{
			Command:     "new",
			Description: "restart",
		},
	}...))
	// check user context expiration every 5 seconds
	go func() {
		for {
			for userID, user := range bot.users {
				cleared := bot.clearUserContextIfExpires(userID)
				if cleared {
					lastMessage := user.LatestMessage
					if cfg.NotifyUserOnConversationIdleTimeout {
						msg := tgbotapi.NewEditMessageText(userID, lastMessage.MessageID, lastMessage.Text+"\n\nContext cleared due to inactivity.")
						_ = bot.send(msg)
					}
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if err := bot.response(update); err != nil {
			log.Printf("error update, %v \n", err)
		}
	}
}

func (bot *MyBot) response(update tgbotapi.Update) error {
	if update.Message == nil { // ignore any non-Message updates
		return errors.New("empty message")
	}

	err := bot.send(tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping))
	if err != nil {
		// Sending chat action returns bool value, which causes `Send` to return unmarshal error.
		// So we need to check if it's an unmarshal error and ignore it.
		var unmarshalError *json.UnmarshalTypeError
		if !errors.As(err, &unmarshalError) {
			log.Print(err)
		}
	}

	if len(cfg.AllowedTelegramID) != 0 {
		var userAllowed bool
		for _, allowedID := range cfg.AllowedTelegramID {
			if allowedID == update.Message.Chat.ID {
				userAllowed = true
			}
		}
		if !userAllowed {
			_, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You are not allowed to use this bot. User ID: %d", update.Message.Chat.ID)))
			if err != nil {
				log.Print(err)
			}
			return err
		}
	}

	if update.Message.IsCommand() { // ignore any non-command Messages
		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "new":
			bot.resetUser(update.Message.From.ID)
			msg.Text = "OK, let's start a new conversation."
		default:
			msg.Text = "I don't know this command."
		}

		if _, err := bot.Send(msg); err != nil {
			log.Print(err)
		}
	} else {
		questionText := update.Message.Text
		voice := false
		if update.Message.Voice != nil && update.Message.Voice.FileID != "" {
			questionText, _, err = a2t(bot.BotAPI, bot.openaiCli, update.Message.Voice.FileID)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to trans audio to text")
				_, _ = bot.Send(msg)
				return err
			}
			voice = true
		}

		answerText, contextTrimmed, err := bot.handleUserPrompt(update.Message.From.ID, questionText)
		if err != nil {
			log.Print(err)
			err = bot.send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			if err != nil {
				log.Print(err)
			}
		} else {
			if !voice {
				err = bot.send(tgbotapi.NewMessage(update.Message.Chat.ID, answerText))
				if err != nil {
					log.Print(err)
				}
			} else {
				speech, err := t2s(bot.ttsCLi, answerText)
				if err != nil {
					return err
				}
				inputFile := tgbotapi.FileBytes{
					Name:  uuid.NewUUID(),
					Bytes: speech,
				}
				if err := bot.send(tgbotapi.NewVoice(update.Message.Chat.ID, inputFile)); err != nil {
					return err
				}
			}

			if contextTrimmed {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Context trimmed.")
				msg.DisableNotification = true
				err = bot.send(msg)
				if err != nil {
					log.Print(err)
				}
			}
		}
	}
	return nil
}

func (bot *MyBot) send(c tgbotapi.Chattable) error {
	msg, err := bot.Send(c)
	if err == nil {
		bot.users[msg.Chat.ID].LatestMessage = msg
	}

	return err
}

func (bot *MyBot) clearUserContextIfExpires(userID int64) bool {
	user := bot.users[userID]
	if user != nil &&
		user.LastActiveTime.Add(time.Duration(cfg.ConversationIdleTimeoutSeconds)*time.Second).Before(time.Now()) {
		bot.resetUser(userID)
		return true
	}

	return false
}

func (bot *MyBot) resetUser(userID int64) {
	delete(bot.users, userID)
}

func (bot *MyBot) handleUserPrompt(userID int64, msg string) (string, bool, error) {
	bot.clearUserContextIfExpires(userID)

	if _, ok := bot.users[userID]; !ok {
		bot.users[userID] = &User{
			TelegramID:     userID,
			LastActiveTime: time.Now(),
			HistoryMessage: []openai.ChatCompletionMessage{},
		}
	}

	bot.users[userID].HistoryMessage = append(bot.users[userID].HistoryMessage, openai.ChatCompletionMessage{
		Role:    "user",
		Content: msg,
	})
	bot.users[userID].LastActiveTime = time.Now()

	c := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Temperature: cfg.ModelTemperature,
		TopP:        1,
		N:           1,
		// PresencePenalty:  0.2,
		// FrequencyPenalty: 0.2,
		Messages: bot.users[userID].HistoryMessage,
	}

	fmt.Println(req)

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Print(err)
		bot.users[userID].HistoryMessage = bot.users[userID].HistoryMessage[:len(bot.users[userID].HistoryMessage)-1]
		return "", false, err
	}

	answer := resp.Choices[0].Message

	bot.users[userID].HistoryMessage = append(bot.users[userID].HistoryMessage, answer)

	var contextTrimmed bool
	if resp.Usage.TotalTokens > 3500 {
		bot.users[userID].HistoryMessage = bot.users[userID].HistoryMessage[1:]
		contextTrimmed = true
	}

	return answer.Content, contextTrimmed, nil
}
