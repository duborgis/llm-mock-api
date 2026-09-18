package domain

import "time"

// RawResponse is the full, provider-shaped response body LiteLLM got back from the mock for
// a single call — captured via a custom LiteLLM callback (litellm/custom_callback.py) reading
// kwargs["original_response"], not anything OpenMeter's real API knows about. Kept alongside
// the OpenMeter events because it lands in the same Mongo instance for the same reason: easy
// inspection of what actually happened on a call, without digging through container logs.
type RawResponse struct {
	CallID     string    `json:"call_id" bson:"_id"`
	Model      string    `json:"model" bson:"model"`
	Route      string    `json:"route" bson:"route"`
	Response   any       `json:"raw_response" bson:"raw_response"`
	ReceivedAt time.Time `json:"received_at" bson:"received_at"`
}
