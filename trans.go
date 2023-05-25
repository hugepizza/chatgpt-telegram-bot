package main

import (
	tts "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/u2takey/go-utils/uuid"
	"io"
	"net/http"
	"os"
)

func t2s(cli *tts.Client, text string) ([]byte, error) {
	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := cli.SynthesizeSpeech(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.GetAudioContent(), nil
}

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
