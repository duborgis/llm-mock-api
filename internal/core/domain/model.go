package domain

// ModelInfo is a generic description of a model exposed by the catalog.
type ModelInfo struct {
	ID        string
	Owner     string
	Created   int64
	Family    string   // e.g. "gpt", "claude", "nova", "llama", "mistral", "gpt-oss", "gemini", "embedding"
	Providers []string // which mock surfaces list this model: "openai", "bedrock", "vertex"
}

// HasProvider reports whether this model should be listed under the given provider surface.
func (m ModelInfo) HasProvider(provider string) bool {
	for _, p := range m.Providers {
		if p == provider {
			return true
		}
	}
	return false
}
