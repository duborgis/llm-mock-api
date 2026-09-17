package contract

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func newBedrockClient(baseURL string) *bedrockruntime.Client {
	return bedrockruntime.New(bedrockruntime.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(baseURL),
		Credentials:  awscreds.NewStaticCredentialsProvider("AKIAFAKE", "secretfake", ""),
	})
}

func TestBedrockContract_Converse(t *testing.T) {
	ts, store := newTestServer()
	defer ts.Close()

	modelID := "anthropic.claude-3-5-sonnet-20240620-v1:0"
	_ = store.Put(context.Background(), domain.Scenario{
		Key:          modelID + "|chat",
		ResponseText: "hello from bedrock mock",
	})

	client := newBedrockClient(ts.URL)
	out, err := client.Converse(context.Background(), &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelID),
		Messages: []types.Message{{
			Role:    types.ConversationRoleUser,
			Content: []types.ContentBlock{&types.ContentBlockMemberText{Value: "hi"}},
		}},
	})
	if err != nil {
		t.Fatalf("bedrockruntime Converse call failed: %v", err)
	}

	msg, ok := out.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		t.Fatalf("expected message output variant, got %T", out.Output)
	}
	textBlock, ok := msg.Value.Content[0].(*types.ContentBlockMemberText)
	if !ok {
		t.Fatalf("expected text content block, got %T", msg.Value.Content[0])
	}
	if textBlock.Value != "hello from bedrock mock" {
		t.Errorf("expected configured content, got %q", textBlock.Value)
	}
}

func TestBedrockContract_InvokeModel(t *testing.T) {
	ts, store := newTestServer()
	defer ts.Close()

	modelID := "anthropic.claude-3-5-sonnet-20240620-v1:0"
	_ = store.Put(context.Background(), domain.Scenario{
		Key:          modelID + "|chat",
		ResponseText: "hello from invoke",
	})

	client := newBedrockClient(ts.URL)
	body := []byte(`{"anthropic_version":"bedrock-2023-05-31","max_tokens":100,"messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]}`)
	out, err := client.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		ContentType: aws.String("application/json"),
		Body:        body,
	})
	if err != nil {
		t.Fatalf("bedrockruntime InvokeModel call failed: %v", err)
	}
	if len(out.Body) == 0 {
		t.Error("expected non-empty response body")
	}
}
