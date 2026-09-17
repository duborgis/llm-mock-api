package openai

import (
	"encoding/json"
	"net/http"
)

// moderations is a minimal stub (deferred surface, see README roadmap): it always reports
// the content as not flagged, using OpenAI's documented moderation response envelope shape.
func (h *Handler) moderations(w http.ResponseWriter, r *http.Request) {
	type result struct {
		Flagged    bool               `json:"flagged"`
		Categories map[string]bool    `json:"categories"`
		Scores     map[string]float64 `json:"category_scores"`
	}
	resp := struct {
		ID      string   `json:"id"`
		Model   string   `json:"model"`
		Results []result `json:"results"`
	}{
		ID:    "modr-mock",
		Model: "text-moderation-mock",
		Results: []result{{
			Flagged:    false,
			Categories: map[string]bool{},
			Scores:     map[string]float64{},
		}},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
