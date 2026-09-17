package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

// Chat implements ports.ChatCompletionUseCase. It is pure business logic: no JSON, no HTTP,
// no provider SDKs. It only depends on the ports it needs (ScenarioRepository, Clock).
type Chat struct {
	Scenarios ports.ScenarioRepository
	Clock     ports.Clock
}

func NewChat(scenarios ports.ScenarioRepository, clock ports.Clock) *Chat {
	return &Chat{Scenarios: scenarios, Clock: clock}
}

const defaultCannedReply = "This is a mocked response."

func (c *Chat) Complete(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	if req.Model == "" || len(req.Messages) == 0 {
		return domain.ChatResponse{}, domain.ErrInvalidRequest
	}

	scenario, err := c.Scenarios.Resolve(ctx, domain.ScenarioKey{Model: req.Model, Operation: "chat"})
	if err != nil {
		return domain.ChatResponse{}, err
	}

	if scenario.LatencyMs > 0 {
		c.Clock.Sleep(time.Duration(scenario.LatencyMs) * time.Millisecond)
	}

	if scenario.Error != nil {
		return domain.ChatResponse{}, &domain.InjectedError{Injection: *scenario.Error}
	}

	text := scenario.ResponseText
	if text == "" {
		text = defaultCannedReply
	}

	usage := scenario.Usage
	if usage.TotalTokens == 0 {
		usage = estimateUsage(req, text)
	}

	return domain.ChatResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", c.Clock.Now().UnixNano()),
		Model:   req.Model,
		Created: c.Clock.Now().Unix(),
		Choices: []domain.Choice{{
			Index: 0,
			Message: domain.ChatMessage{
				Role:    domain.RoleAssistant,
				Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: text}},
			},
			FinishReason: domain.FinishStop,
		}},
		Usage: usage,
	}, nil
}

func (c *Chat) Stream(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamChunk, error) {
	if req.Model == "" || len(req.Messages) == 0 {
		return nil, domain.ErrInvalidRequest
	}

	scenario, err := c.Scenarios.Resolve(ctx, domain.ScenarioKey{Model: req.Model, Operation: "chat"})
	if err != nil {
		return nil, err
	}

	if scenario.LatencyMs > 0 {
		c.Clock.Sleep(time.Duration(scenario.LatencyMs) * time.Millisecond)
	}

	if scenario.Error != nil {
		return nil, &domain.InjectedError{Injection: *scenario.Error}
	}

	text := scenario.ResponseText
	if text == "" {
		text = defaultCannedReply
	}
	words := splitWords(text)

	id := fmt.Sprintf("chatcmpl-%d", c.Clock.Now().UnixNano())
	created := c.Clock.Now().Unix()

	ch := make(chan domain.StreamChunk)
	go func() {
		defer close(ch)

		interDelay := time.Duration(0)
		if scenario.TokensPerSecond > 0 {
			interDelay = time.Second / time.Duration(scenario.TokensPerSecond)
		}

		for i, w := range words {
			select {
			case <-ctx.Done():
				return
			default:
			}
			chunk := domain.StreamChunk{ID: id, Model: req.Model, Created: created, Index: 0}
			if i == 0 {
				chunk.DeltaRole = domain.RoleAssistant
			}
			chunk.DeltaContent = w
			ch <- chunk
			if interDelay > 0 {
				c.Clock.Sleep(interDelay)
			}
		}

		usage := scenario.Usage
		if usage.TotalTokens == 0 {
			usage = estimateUsage(req, text)
		}
		ch <- domain.StreamChunk{
			ID: id, Model: req.Model, Created: created, Index: 0,
			FinishReason: domain.FinishStop,
			Usage:        &usage,
		}
	}()

	return ch, nil
}

func splitWords(text string) []string {
	var out []string
	cur := ""
	for _, r := range text {
		cur += string(r)
		if r == ' ' {
			out = append(out, cur)
			cur = ""
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	if len(out) == 0 {
		out = []string{text}
	}
	return out
}

func estimateUsage(req domain.ChatRequest, responseText string) domain.TokenUsage {
	prompt := 0
	for _, m := range req.Messages {
		prompt += len(m.Text()) / 4
	}
	completion := len(responseText) / 4
	if completion == 0 {
		completion = 1
	}
	return domain.TokenUsage{
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      prompt + completion,
	}
}
