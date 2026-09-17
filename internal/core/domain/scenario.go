package domain

import "time"

// ErrorInjection describes an error the mock should return instead of a normal response.
type ErrorInjection struct {
	StatusCode int
	Message    string
	Type       string // provider-agnostic error category, e.g. "rate_limit", "invalid_request"
}

// Scenario describes how the mock should behave for a given model/matcher.
type Scenario struct {
	Key string

	// Fixed content returned for chat completions. If empty, a default canned reply is used.
	ResponseText string

	// LatencyMs is injected before responding (or before the first stream chunk).
	LatencyMs int

	// TokensPerSecond simulates generation speed for streaming responses; 0 means "as fast as possible".
	TokensPerSecond int

	// Error, if non-nil, makes the use case return this error instead of a normal response.
	Error *ErrorInjection

	// Usage overrides the computed token usage when non-zero.
	Usage TokenUsage
}

// ScenarioKey identifies which scenario to resolve for a request.
type ScenarioKey struct {
	Model     string
	Operation string // "chat", "embedding", etc.
}

// Now is a small helper type alias kept in the domain so usecases don't need to import "time" callers'
// concrete clock implementations; usecases depend on the Clock port instead.
type Duration = time.Duration
