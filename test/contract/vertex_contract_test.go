package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"google.golang.org/genai"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// TestVertexContract_GenerateContent uses a plain http.Client against the mock (Vertex's Go
// SDK support for custom base URLs is limited) and unmarshals the response into genai's own
// GenerateContentResponse struct — proving the JSON shape round-trips through Google's own
// generated types, which is the fidelity property that matters here.
func TestVertexContract_GenerateContent(t *testing.T) {
	ts, store := newTestServer()
	defer ts.Close()

	_ = store.Put(context.Background(), domain.Scenario{
		Key:          "gemini-1.5-pro|chat",
		ResponseText: "hello from vertex mock",
	})

	reqBody := map[string]any{
		"contents": []*genai.Content{{
			Role:  "user",
			Parts: []*genai.Part{{Text: "hi"}},
		}},
	}
	data, _ := json.Marshal(reqBody)

	url := ts.URL + "/v1/projects/my-project/locations/us-central1/publishers/google/models/gemini-1.5-pro:generateContent"
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("HTTP call failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var out genai.GenerateContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("genai.GenerateContentResponse failed to unmarshal response: %v", err)
	}
	if len(out.Candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(out.Candidates))
	}
	text := out.Candidates[0].Content.Parts[0].Text
	if text != "hello from vertex mock" {
		t.Errorf("expected configured content, got %q", text)
	}
}

func TestVertexContract_CountTokens(t *testing.T) {
	ts, _ := newTestServer()
	defer ts.Close()

	reqBody := map[string]any{
		"contents": []*genai.Content{{Role: "user", Parts: []*genai.Part{{Text: "hello world"}}}},
	}
	data, _ := json.Marshal(reqBody)
	url := ts.URL + "/v1/projects/my-project/locations/us-central1/publishers/google/models/gemini-1.5-pro:countTokens"
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("HTTP call failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
