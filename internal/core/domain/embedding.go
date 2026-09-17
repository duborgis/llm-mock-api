package domain

// EmbeddingRequest is a generic request for one or more embedding vectors.
type EmbeddingRequest struct {
	Model string
	Input []string
}

// Embedding is a single generated vector, with its position in the input batch.
type Embedding struct {
	Index     int
	Vector    []float64
	InputText string
}

// EmbeddingResponse is the generic embedding result.
type EmbeddingResponse struct {
	Model      string
	Embeddings []Embedding
	Usage      TokenUsage
}
