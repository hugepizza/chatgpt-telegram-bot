package main

import (
	"bytes"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	mp3 "github.com/hajimehoshi/go-mp3"
	openai "github.com/sashabaranov/go-openai"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/u2takey/go-utils/uuid"
	"io"
	"math"
	"net/http"
	"os"
	"path"
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

	os.TempDir()
	mp3Name := path.Join(os.TempDir(), fmt.Sprintf("/%s.mp3", uuid.NewUUID()))
	if err := ffmpeg.Input(file.Name()).Output(mp3Name).OverWriteOutput().Run(); err != nil {
		return nil, err
	}
	return os.Open(mp3Name)
}

func a2t(bot *tgbotapi.BotAPI, openaiCli *openai.Client, fileId string) (string, bool, error) {
	fileUrl, _ := bot.GetFileDirectURL(fileId)
	file, err := downloadFile(fileUrl)
	if err != nil {
		return "", false, err
	}
	defer os.Remove(file.Name())
	ctx := context.Background()
	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: file.Name(),
	}
	resp, err := openaiCli.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return "", false, err
	}
	return resp.Text, true, err
}

func GetAudioDuration(data []byte) (int, error) {
	d, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	samples := d.Length() / 4 // simpleSize = 4
	return int(math.Ceil(float64(samples) / float64(d.SampleRate()))), nil
}
