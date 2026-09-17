package domain

// Role identifies who authored a ChatMessage.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ContentPartType distinguishes text from image content within a message.
type ContentPartType string

const (
	ContentPartText  ContentPartType = "text"
	ContentPartImage ContentPartType = "image"
)

// ContentPart is one piece of a (possibly multimodal) message.
type ContentPart struct {
	Type     ContentPartType
	Text     string
	ImageURL string
}

// ChatMessage is a single generic message in a conversation.
type ChatMessage struct {
	Role    Role
	Content []ContentPart
	Name    string
}

// Text returns the concatenation of all text parts, convenience for simple cases.
func (m ChatMessage) Text() string {
	out := ""
	for _, p := range m.Content {
		if p.Type == ContentPartText {
			out += p.Text
		}
	}
	return out
}

// ChatRequest is the generic, provider-agnostic chat completion request.
type ChatRequest struct {
	Model       string
	Messages    []ChatMessage
	Temperature float64
	MaxTokens   int
	Stream      bool
}

// FinishReason enumerates why generation stopped.
type FinishReason string

const (
	FinishStop          FinishReason = "stop"
	FinishLength        FinishReason = "length"
	FinishContentFilter FinishReason = "content_filter"
	FinishError         FinishReason = "error"
)

// Choice is one generated completion candidate.
type Choice struct {
	Index        int
	Message      ChatMessage
	FinishReason FinishReason
}

// ChatResponse is the generic chat completion result.
type ChatResponse struct {
	ID      string
	Model   string
	Created int64
	Choices []Choice
	Usage   TokenUsage
}
