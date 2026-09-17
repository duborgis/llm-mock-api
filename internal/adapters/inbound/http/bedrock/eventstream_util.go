package bedrock

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"
)

// writeEventStreamMessage encodes one AWS eventstream binary frame carrying eventType/payload,
// exactly as bedrockruntime's real ConverseStream/InvokeModelWithResponseStream responses do.
// This gives genuine wire-format fidelity (not a simplified framing) using AWS's own
// eventstream encoder, which the Go SDK's ConverseStream/InvokeModelWithResponseStream
// readers decode with the same package.
func writeEventStreamMessage(w http.ResponseWriter, enc *eventstream.Encoder, eventType string, payload []byte) error {
	msg := eventstream.Message{
		Headers: eventstream.Headers{
			{Name: ":message-type", Value: eventstream.StringValue("event")},
			{Name: ":event-type", Value: eventstream.StringValue(eventType)},
			{Name: ":content-type", Value: eventstream.StringValue("application/json")},
		},
		Payload: payload,
	}
	if err := enc.Encode(w, msg); err != nil {
		return err
	}
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}
