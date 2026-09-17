package contract

import (
	"context"
	"testing"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func TestOpenAIContract_ChatCompletion(t *testing.T) {
	ts, store := newTestServer()
	defer ts.Close()

	_ = store.Put(context.Background(), domain.Scenario{
		Key:          "gpt-4o-mini|chat",
		ResponseText: "hello from the mock",
	})

	client := openai.NewClient(option.WithBaseURL(ts.URL+"/v1/"), option.WithAPIKey("test-key"))

	resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: "gpt-4o-mini",
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("hi there"),
		},
	})
	if err != nil {
		t.Fatalf("openai-go client call failed: %v", err)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "hello from the mock" {
		t.Errorf("expected configured content, got %q", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("expected finish reason stop, got %q", resp.Choices[0].FinishReason)
	}
}

func TestOpenAIContract_ListModels(t *testing.T) {
	ts, _ := newTestServer()
	defer ts.Close()

	client := openai.NewClient(option.WithBaseURL(ts.URL+"/v1/"), option.WithAPIKey("test-key"))
	page, err := client.Models.List(context.Background())
	if err != nil {
		t.Fatalf("openai-go client call failed: %v", err)
	}
	if len(page.Data) == 0 {
		t.Error("expected at least one model")
	}
}

func TestOpenAIContract_Embeddings(t *testing.T) {
	ts, _ := newTestServer()
	defer ts.Close()

	client := openai.NewClient(option.WithBaseURL(ts.URL+"/v1/"), option.WithAPIKey("test-key"))
	resp, err := client.Embeddings.New(context.Background(), openai.EmbeddingNewParams{
		Model: "text-embedding-3-small",
		Input: openai.EmbeddingNewParamsInputUnion{OfString: openai.String("hello world")},
	})
	if err != nil {
		t.Fatalf("openai-go client call failed: %v", err)
	}
	if len(resp.Data) != 1 || len(resp.Data[0].Embedding) == 0 {
		t.Errorf("expected non-empty embedding vector, got %+v", resp.Data)
	}
}
