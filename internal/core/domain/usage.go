package domain

// TokenUsage tracks generic token accounting for a completion or embedding call.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}
