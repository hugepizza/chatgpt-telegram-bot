package main

import (
	"github.com/caarlos0/env/v7"
	ttsp "github.com/hugepizza/chatgpt-telegram-bot/tts"
)

var cfg struct {
	TelegramAPIToken                    string  `env:"TELEGRAM_APITOKEN,required"`
	OpenAIAPIKey                        string  `env:"OPENAI_API_KEY,required"`
	ModelTemperature                    float32 `env:"MODEL_TEMPERATURE" envDefault:"1.0"`
	AllowedTelegramID                   []int64 `env:"ALLOWED_TELEGRAM_ID" envSeparator:","`
	ConversationIdleTimeoutSeconds      int     `env:"CONVERSATION_IDLE_TIMEOUT_SECONDS" envDefault:"86400"`
	NotifyUserOnConversationIdleTimeout bool    `env:"NOTIFY_USER_ON_CONVERSATION_IDLE_TIMEOUT" envDefault:"true"`
	AzureTTSKey                         string  `env:"Azure_TTS_Key"`
	AzureTTSRegion                      string  `env:"Azure_TTS_Region"`
	GoogleTTSCertFile                   string  `env:"GOOGLE_TTS_Cert_File"`
}

func main() {
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	var (
		err error
		tts ttsp.Client
	)
	if cfg.GoogleTTSCertFile != "" {
		tts, err = ttsp.NewGoogleTTS(cfg.GoogleTTSCertFile)
	} else if cfg.AzureTTSKey != "" && cfg.AzureTTSRegion != "" {
		tts, err = ttsp.NewAzureTTS(cfg.AzureTTSKey, cfg.AzureTTSRegion)
	} else {
		panic("need a tts client")
	}
	if err != nil {
		panic(err)
	}

	bot, err := NewMyBot(tts)
	if err != nil {
		panic(err)
	}
	bot.Run()
}
