package domain

import "time"

// Event is a generic representation of the CloudEvents-shaped usage event LiteLLM's OpenMeter
// integration POSTs on every successful call (see litellm/integrations/openmeter.py:
// OpenMeterLogger._common_logic). "Data" is kept loose (map) since it's just
// {"model", "cost", "prompt_tokens", "completion_tokens", "total_tokens"} today but is not
// part of the CloudEvents envelope contract this mock actually needs to be faithful to.
type Event struct {
	ID         string         `json:"id" bson:"_id"`
	SpecVersion string        `json:"specversion" bson:"specversion"`
	Type       string         `json:"type" bson:"type"`
	Source     string         `json:"source" bson:"source"`
	Subject    string         `json:"subject" bson:"subject"`
	Time       time.Time      `json:"time" bson:"time"`
	Data       map[string]any `json:"data" bson:"data"`
	ReceivedAt time.Time      `json:"received_at" bson:"received_at"`
}
