package usecase

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

// placeholderMP4 is a syntactically minimal ISO base media file (a bare "ftyp" box) — unlike
// tiny1x1PNG/silentWAV it has no "mdat"/"moov" so it won't actually decode as a playable clip.
// A byte-perfect placeholder isn't worth the effort for a mock nobody plays back; it exists so
// GET .../content returns *something* shaped like an mp4 file, not an empty body.
func placeholderMP4() []byte {
	buf := &bytes.Buffer{}
	brand := []byte("isomiso2mp41")
	binary.Write(buf, binary.BigEndian, uint32(8+8+len(brand)))
	buf.WriteString("ftyp")
	buf.WriteString("isom")
	binary.Write(buf, binary.BigEndian, uint32(512))
	buf.Write(brand)
	return buf.Bytes()
}

// Video implements ports.VideoGenerationUseCase. Real Sora jobs run asynchronously (queued ->
// in_progress -> completed, polled over minutes); this mock has nothing to render, so it marks
// every job completed immediately and just keeps it around in memory for the inevitable GET.
type Video struct {
	Scenarios ports.ScenarioRepository
	Clock     ports.Clock

	mu     sync.Mutex
	jobs   map[string]domain.Video
	nextID int64
}

func NewVideo(scenarios ports.ScenarioRepository, clock ports.Clock) *Video {
	return &Video{Scenarios: scenarios, Clock: clock, jobs: make(map[string]domain.Video)}
}

func (v *Video) Create(ctx context.Context, req domain.VideoRequest) (domain.Video, error) {
	if req.Model == "" || req.Prompt == "" {
		return domain.Video{}, domain.ErrInvalidRequest
	}

	scenario, err := v.Scenarios.Resolve(ctx, domain.ScenarioKey{Model: req.Model, Operation: "video"})
	if err != nil {
		return domain.Video{}, err
	}
	if scenario.Error != nil {
		return domain.Video{}, &domain.InjectedError{Injection: *scenario.Error}
	}

	seconds := req.Seconds
	if seconds == "" {
		seconds = "4"
	}
	size := req.Size
	if size == "" {
		size = "720x1280"
	}

	now := v.Clock.Now().Unix()
	completed := now
	id := fmt.Sprintf("video_%d", atomic.AddInt64(&v.nextID, 1))
	job := domain.Video{
		ID:          id,
		Object:      "video",
		Status:      "completed",
		Progress:    100,
		Model:       req.Model,
		Prompt:      req.Prompt,
		Seconds:     seconds,
		Size:        size,
		CreatedAt:   now,
		CompletedAt: &completed,
	}

	v.mu.Lock()
	v.jobs[id] = job
	v.mu.Unlock()

	return job, nil
}

func (v *Video) Get(ctx context.Context, id string) (domain.Video, error) {
	v.mu.Lock()
	job, ok := v.jobs[id]
	v.mu.Unlock()
	if !ok {
		return domain.Video{}, domain.ErrNotFound
	}
	return job, nil
}

func (v *Video) Content(_ context.Context, id string) ([]byte, string, error) {
	v.mu.Lock()
	_, ok := v.jobs[id]
	v.mu.Unlock()
	if !ok {
		return nil, "", domain.ErrNotFound
	}
	return placeholderMP4(), "video/mp4", nil
}
