package domain

// TranscriptionRequest is the generic domain shape for speech-to-text, decoupled from the
// SDK's multipart request struct.
type TranscriptionRequest struct {
	Model    string
	Filename string
	Language string
	Prompt   string
}

// TranscriptionResponse is the generic domain shape for a speech-to-text result.
type TranscriptionResponse struct {
	Text string
}

// SpeechRequest is the generic domain shape for text-to-speech.
type SpeechRequest struct {
	Model          string
	Input          string
	Voice          string
	ResponseFormat string
}

// SpeechResponse is the generic domain shape for a text-to-speech result: raw audio bytes
// plus the content type they should be served as.
type SpeechResponse struct {
	Audio       []byte
	ContentType string
}
