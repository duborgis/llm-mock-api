package domain

// ModelInfo is a generic description of a model exposed by the catalog.
type ModelInfo struct {
	ID      string
	Owner   string
	Created int64
	Family  string // e.g. "gpt", "claude", "titan", "llama", "gemini"
}
