package vertex

import "google.golang.org/genai"

// generateContentRequest is a hand-rolled mirror of Vertex's generateContent/streamGenerateContent
// REST request body, built from genai's own Content/Part types (which have proper JSON tags) plus
// a small hand-rolled generationConfig, since genai.GenerateContentConfig itself is shaped for the
// client SDK's internal call signature rather than the literal wire body.
type generateContentRequest struct {
	Contents          []*genai.Content  `json:"contents"`
	SystemInstruction *genai.Content    `json:"systemInstruction,omitempty"`
	GenerationConfig  *generationConfig `json:"generationConfig,omitempty"`
}

type generationConfig struct {
	Temperature     *float32 `json:"temperature,omitempty"`
	MaxOutputTokens int32    `json:"maxOutputTokens,omitempty"`
}

type countTokensRequest struct {
	Contents []*genai.Content `json:"contents"`
}

type countTokensResponse struct {
	TotalTokens int32 `json:"totalTokens"`
}
