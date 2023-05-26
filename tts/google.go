package tts

import (
	google "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"google.golang.org/api/option"
)

type GoogleTTS struct {
	*google.Client
}

func NewGoogleTTS(certFile string) (*GoogleTTS, error) {
	client, err := google.NewClient(context.Background(), option.WithCredentialsFile(certFile))
	if err != nil {
		panic(err)
	}
	return &GoogleTTS{Client: client}, nil
}

func (tts *GoogleTTS) T2S(text string) ([]byte, int, error) {
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
	resp, err := tts.SynthesizeSpeech(context.Background(), &req)
	if err != nil {
		return nil, 0, err
	}
	return resp.GetAudioContent(), 0, nil
}
