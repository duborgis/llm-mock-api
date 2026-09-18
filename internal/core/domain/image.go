package domain

// ImageRequest is a generic request to generate one or more images from a text prompt.
type ImageRequest struct {
	Model  string
	Prompt string
	N      int
	Size   string
	User   string
}

// GeneratedImage is a single generated image, encoded as base64 (the mock never renders
// real pixels — it returns a small deterministic placeholder so callers can round-trip
// the response shape without needing a real image model).
type GeneratedImage struct {
	B64JSON       string
	RevisedPrompt string
}

// ImageResponse is the generic image-generation result.
type ImageResponse struct {
	Created int64
	Images  []GeneratedImage
	Usage   TokenUsage
}
