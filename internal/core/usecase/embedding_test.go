package usecase

import (
	"context"
	"reflect"
	"testing"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func TestEmbed_Deterministic(t *testing.T) {
	repo := newFakeScenarioRepo()
	emb := NewEmbedding(repo, newFakeClock())

	req := domain.EmbeddingRequest{Model: "text-embedding-3-small", Input: []string{"hello world"}}
	resp1, err := emb.Embed(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp2, err := emb.Embed(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(resp1.Embeddings[0].Vector, resp2.Embeddings[0].Vector) {
		t.Error("expected deterministic embeddings for identical input")
	}
	if len(resp1.Embeddings[0].Vector) != emb.Dims {
		t.Errorf("expected %d dims, got %d", emb.Dims, len(resp1.Embeddings[0].Vector))
	}
}

func TestEmbed_InvalidRequest(t *testing.T) {
	emb := NewEmbedding(newFakeScenarioRepo(), newFakeClock())
	_, err := emb.Embed(context.Background(), domain.EmbeddingRequest{})
	if err != domain.ErrInvalidRequest {
		t.Errorf("expected ErrInvalidRequest, got %v", err)
	}
}
