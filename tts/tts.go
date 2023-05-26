package tts

type Client interface {
	// T2S content and duration
	T2S(text string) ([]byte, int, error)
}
