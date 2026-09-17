package usecase

import (
	"context"
	"testing"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func TestChatComplete_DefaultCannedReply(t *testing.T) {
	repo := newFakeScenarioRepo()
	chat := NewChat(repo, newFakeClock())

	resp, err := chat.Complete(context.Background(), domain.ChatRequest{
		Model:    "gpt-4o-mini",
		Messages: []domain.ChatMessage{{Role: domain.RoleUser, Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: "hi"}}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Text() != defaultCannedReply {
		t.Errorf("expected default canned reply, got %q", resp.Choices[0].Message.Text())
	}
	if resp.Choices[0].FinishReason != domain.FinishStop {
		t.Errorf("expected stop finish reason, got %q", resp.Choices[0].FinishReason)
	}
}

func TestChatComplete_ScenarioFixedResponse(t *testing.T) {
	repo := newFakeScenarioRepo()
	_ = repo.Put(context.Background(), domain.Scenario{Key: "gpt-4o|chat", ResponseText: "custom reply"})
	chat := NewChat(repo, newFakeClock())

	resp, err := chat.Complete(context.Background(), domain.ChatRequest{
		Model:    "gpt-4o",
		Messages: []domain.ChatMessage{{Role: domain.RoleUser, Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: "hi"}}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Choices[0].Message.Text() != "custom reply" {
		t.Errorf("expected custom reply, got %q", resp.Choices[0].Message.Text())
	}
}

func TestChatComplete_ErrorInjection(t *testing.T) {
	repo := newFakeScenarioRepo()
	_ = repo.Put(context.Background(), domain.Scenario{
		Key:   "broken-model|chat",
		Error: &domain.ErrorInjection{StatusCode: 429, Message: "rate limited", Type: "rate_limit"},
	})
	chat := NewChat(repo, newFakeClock())

	_, err := chat.Complete(context.Background(), domain.ChatRequest{
		Model:    "broken-model",
		Messages: []domain.ChatMessage{{Role: domain.RoleUser, Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: "hi"}}}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	injected, ok := err.(*domain.InjectedError)
	if !ok {
		t.Fatalf("expected *domain.InjectedError, got %T", err)
	}
	if injected.Injection.StatusCode != 429 {
		t.Errorf("expected status 429, got %d", injected.Injection.StatusCode)
	}
}

func TestChatComplete_InvalidRequest(t *testing.T) {
	chat := NewChat(newFakeScenarioRepo(), newFakeClock())
	_, err := chat.Complete(context.Background(), domain.ChatRequest{})
	if err != domain.ErrInvalidRequest {
		t.Errorf("expected ErrInvalidRequest, got %v", err)
	}
}

func TestChatStream_EmitsChunksAndFinalUsage(t *testing.T) {
	repo := newFakeScenarioRepo()
	_ = repo.Put(context.Background(), domain.Scenario{Key: "gpt-4o|chat", ResponseText: "hello there friend"})
	chat := NewChat(repo, newFakeClock())

	ch, err := chat.Stream(context.Background(), domain.ChatRequest{
		Model:    "gpt-4o",
		Messages: []domain.ChatMessage{{Role: domain.RoleUser, Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: "hi"}}}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var chunks []domain.StreamChunk
	for c := range ch {
		chunks = append(chunks, c)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected at least 2 chunks, got %d", len(chunks))
	}
	last := chunks[len(chunks)-1]
	if last.FinishReason != domain.FinishStop {
		t.Errorf("expected final chunk finish reason stop, got %q", last.FinishReason)
	}
	if last.Usage == nil {
		t.Error("expected final chunk to carry usage")
	}
	if chunks[0].DeltaRole != domain.RoleAssistant {
		t.Errorf("expected first chunk to carry assistant role delta, got %q", chunks[0].DeltaRole)
	}
}
