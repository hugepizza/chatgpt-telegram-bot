package tts

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AzureTTS struct {
	*http.Client
	key, region string
}

func NewAzureTTS(key, region string) (*AzureTTS, error) {
	return &AzureTTS{
		Client: &http.Client{Timeout: time.Second * 15},
		key:    key,
		region: region,
	}, nil
}

func (tts *AzureTTS) T2S(text string) ([]byte, int, error) {
	url := fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", tts.region)
	postData := fmt.Sprintf(`<speak version='1.0' xml:lang='en-US'>
    <voice xml:lang='en-US' xml:gender='Female' name='en-US-JennyNeural'>
        %s
    </voice>
</speak>`, text)
	req, err := http.NewRequest("POST", url, strings.NewReader(postData))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Ocp-Apim-Subscription-Key", tts.key)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", "audio-16khz-128kbitrate-mono-mp3")
	req.Header.Set("User-Agent", "ph")
	resp, err := tts.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("%d", resp.StatusCode)
	}
	voiceData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return voiceData, 0, nil
}
