package bedrock

// Hand-rolled mirrors of Bedrock's documented wire JSON shapes. See router.go for why these
// aren't the bedrockruntime SDK's own Go structs.

// converseContentBlock mirrors Bedrock's ContentBlock union (text-only variant, which covers
// the common case this mock targets).
type converseContentBlock struct {
	Text string `json:"text,omitempty"`
}

type converseMessage struct {
	Role    string                 `json:"role"`
	Content []converseContentBlock `json:"content"`
}

type converseInferenceConfig struct {
	MaxTokens   *int32   `json:"maxTokens,omitempty"`
	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"topP,omitempty"`
}

// converseRequestBody mirrors the ConverseInput wire body (modelId comes from the URL path).
type converseRequestBody struct {
	Messages        []converseMessage       `json:"messages"`
	System          []converseContentBlock  `json:"system,omitempty"`
	InferenceConfig converseInferenceConfig `json:"inferenceConfig,omitempty"`
}

type converseUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

type converseMetrics struct {
	LatencyMs int64 `json:"latencyMs"`
}

// converseResponseBody mirrors the ConverseOutput wire body.
type converseResponseBody struct {
	Output     converseOutputWrapper `json:"output"`
	StopReason string                `json:"stopReason"`
	Usage      converseUsage         `json:"usage"`
	Metrics    converseMetrics       `json:"metrics"`
}

type converseOutputWrapper struct {
	Message converseMessage `json:"message"`
}

// converseErrorBody mirrors Bedrock's smithy restjson1 error envelope.
type converseErrorBody struct {
	Message string `json:"message"`
}

// Claude-on-Bedrock InvokeModel body shapes (Anthropic Messages API format).
type claudeInvokeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type claudeInvokeMessage struct {
	Role    string                     `json:"role"`
	Content []claudeInvokeContentBlock `json:"content"`
}

type claudeInvokeRequestBody struct {
	AnthropicVersion string                `json:"anthropic_version"`
	Messages         []claudeInvokeMessage `json:"messages"`
	System           string                `json:"system,omitempty"`
	MaxTokens        int                   `json:"max_tokens"`
	Temperature      float64               `json:"temperature,omitempty"`
}

type claudeInvokeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type claudeInvokeResponseBody struct {
	ID           string                     `json:"id"`
	Type         string                     `json:"type"`
	Role         string                     `json:"role"`
	Model        string                     `json:"model"`
	Content      []claudeInvokeContentBlock `json:"content"`
	StopReason   string                     `json:"stop_reason"`
	StopSequence *string                    `json:"stop_sequence"`
	Usage        claudeInvokeUsage          `json:"usage"`
}
