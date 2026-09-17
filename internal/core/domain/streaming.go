package domain

// StreamChunk is one generic incremental piece of a streamed chat completion.
type StreamChunk struct {
	ID           string
	Model        string
	Created      int64
	Index        int
	DeltaRole    Role
	DeltaContent string
	FinishReason FinishReason // empty until the final chunk
	Usage        *TokenUsage  // set only on the final chunk, if the provider reports usage
}
