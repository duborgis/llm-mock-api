package usecase

import (
	"bytes"
	"context"
	"encoding/binary"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

// silentWAV builds a real, valid mono 16-bit PCM WAV file containing a fraction of a second
// of silence. The mock never synthesizes actual speech — this lets callers round-trip a
// syntactically valid audio payload, regardless of the response_format they asked for.
func silentWAV() []byte {
	const sampleRate = 8000
	samples := make([]byte, sampleRate/10*2) // 100ms of 16-bit silence

	buf := &bytes.Buffer{}
	buf.WriteString("RIFF")
	binary.Write(buf, binary.LittleEndian, uint32(36+len(samples)))
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	binary.Write(buf, binary.LittleEndian, uint32(16)) // fmt chunk size
	binary.Write(buf, binary.LittleEndian, uint16(1))  // PCM
	binary.Write(buf, binary.LittleEndian, uint16(1))  // mono
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate*2)) // byte rate
	binary.Write(buf, binary.LittleEndian, uint16(2))            // block align
	binary.Write(buf, binary.LittleEndian, uint16(16))           // bits per sample
	buf.WriteString("data")
	binary.Write(buf, binary.LittleEndian, uint32(len(samples)))
	buf.Write(samples)
	return buf.Bytes()
}

// audioContentTypes maps openai-go's AudioSpeechNewParamsResponseFormat values to their
// serving content type.
var audioContentTypes = map[string]string{
	"mp3":  "audio/mpeg",
	"opus": "audio/opus",
	"aac":  "audio/aac",
	"flac": "audio/flac",
	"wav":  "audio/wav",
	"pcm":  "audio/pcm",
}

// Audio implements ports.AudioTranscriptionUseCase and ports.AudioSpeechUseCase.
type Audio struct {
	Scenarios ports.ScenarioRepository
}

func NewAudio(scenarios ports.ScenarioRepository) *Audio {
	return &Audio{Scenarios: scenarios}
}

func (a *Audio) Transcribe(ctx context.Context, req domain.TranscriptionRequest) (domain.TranscriptionResponse, error) {
	if req.Model == "" {
		return domain.TranscriptionResponse{}, domain.ErrInvalidRequest
	}

	scenario, err := a.Scenarios.Resolve(ctx, domain.ScenarioKey{Model: req.Model, Operation: "audio_transcription"})
	if err != nil {
		return domain.TranscriptionResponse{}, err
	}
	if scenario.Error != nil {
		return domain.TranscriptionResponse{}, &domain.InjectedError{Injection: *scenario.Error}
	}

	return domain.TranscriptionResponse{Text: "This is a mocked transcription."}, nil
}

func (a *Audio) Synthesize(ctx context.Context, req domain.SpeechRequest) (domain.SpeechResponse, error) {
	if req.Model == "" || req.Input == "" {
		return domain.SpeechResponse{}, domain.ErrInvalidRequest
	}

	scenario, err := a.Scenarios.Resolve(ctx, domain.ScenarioKey{Model: req.Model, Operation: "audio_speech"})
	if err != nil {
		return domain.SpeechResponse{}, err
	}
	if scenario.Error != nil {
		return domain.SpeechResponse{}, &domain.InjectedError{Injection: *scenario.Error}
	}

	contentType, ok := audioContentTypes[req.ResponseFormat]
	if !ok {
		contentType = audioContentTypes["mp3"]
	}

	return domain.SpeechResponse{Audio: silentWAV(), ContentType: contentType}, nil
}
