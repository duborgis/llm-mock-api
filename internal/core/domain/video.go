package domain

// VideoRequest is the generic domain shape for a Sora-style video generation job, mirrored
// from openai-python's VideoCreateParams (openai.types.video_create_params) since the Go SDK
// (openai-go v1.12.0) has no video support at all yet — see internal/core/usecase/video.go.
type VideoRequest struct {
	Model   string
	Prompt  string
	Seconds string
	Size    string
}

// VideoError mirrors openai-python's VideoCreateError.
type VideoError struct {
	Code    string
	Message string
}

// Video mirrors openai-python's Video type (openai.types.video): a generation job, not an
// immediate result — real Sora jobs are polled via GET until Status reaches "completed".
type Video struct {
	ID          string
	Object      string
	Status      string // "queued", "in_progress", "completed", "failed"
	Progress    int
	Model       string
	Prompt      string
	Seconds     string
	Size        string
	CreatedAt   int64
	CompletedAt *int64
	ExpiresAt   *int64
	Error       *VideoError
}
