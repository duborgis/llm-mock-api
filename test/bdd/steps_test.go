//go:build bdd

// Package bdd drives real HTTP requests against the docker-compose stack (LiteLLM +
// OpenMeter mock) to verify that usage events are actually emitted end-to-end, for every
// route the mock implements. Run with: go test -tags bdd ./test/bdd/...
// Requires `make litellm-up` (or `docker compose up -d`) to already be running.
package bdd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var (
	litellmBaseURL   = envOr("LITELLM_BASE_URL", "http://localhost:4000")
	openmeterBaseURL = envOr("OPENMETER_BASE_URL", "http://localhost:9010")
	litellmMasterKey = envOr("LITELLM_MASTER_KEY", "sk-litellm-local-test")
)

type openmeterEvent struct {
	ID   string         `json:"id"`
	Type string         `json:"type"`
	Time time.Time      `json:"time"`
	Data map[string]any `json:"data"`
}

type rawResponse struct {
	CallID     string    `json:"call_id"`
	Model      string    `json:"model"`
	Route      string    `json:"route"`
	ReceivedAt time.Time `json:"received_at"`
}

type suiteState struct {
	scenarioStart time.Time
	lastErr       error
}

func (s *suiteState) reachable(ctx context.Context) error {
	s.scenarioStart = time.Now().UTC()

	resp, err := http.Get(litellmBaseURL + "/health/liveliness")
	if err != nil {
		return fmt.Errorf("litellm not reachable at %s: %w", litellmBaseURL, err)
	}
	resp.Body.Close()

	resp, err = http.Get(openmeterBaseURL + "/api/v1/events?limit=1")
	if err != nil {
		return fmt.Errorf("openmeter mock not reachable at %s: %w", openmeterBaseURL, err)
	}
	resp.Body.Close()
	return nil
}

func (s *suiteState) postToLiteLLM(path string, body map[string]any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, litellmBaseURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+litellmMasterKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("%s returned status %d", path, resp.StatusCode)
	}
	return nil
}

func (s *suiteState) requestChatCompletion(model, prompt string) error {
	s.lastErr = s.postToLiteLLM("/v1/chat/completions", map[string]any{
		"model":    model,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"user":     "bdd-tests@example.com",
	})
	return s.lastErr
}

func (s *suiteState) requestResponse(model, prompt string) error {
	s.lastErr = s.postToLiteLLM("/v1/responses", map[string]any{
		"model": model,
		"input": prompt,
		"user":  "bdd-tests@example.com",
	})
	return s.lastErr
}

func (s *suiteState) requestImageGeneration(model, prompt string) error {
	s.lastErr = s.postToLiteLLM("/v1/images/generations", map[string]any{
		"model":  model,
		"prompt": prompt,
		"user":   "bdd-tests@example.com",
	})
	return s.lastErr
}

func (s *suiteState) requestVideoGeneration(model, prompt string) error {
	s.lastErr = s.postToLiteLLM("/v1/videos", map[string]any{
		"model":  model,
		"prompt": prompt,
		"user":   "bdd-tests@example.com",
	})
	return s.lastErr
}

func (s *suiteState) requestSpeech(model, input string) error {
	s.lastErr = s.postToLiteLLM("/v1/audio/speech", map[string]any{
		"model": model,
		"input": input,
		"voice": "alloy",
		"user":  "bdd-tests@example.com",
	})
	return s.lastErr
}

// requestTranscription posts a tiny fake audio file as multipart/form-data, the wire format
// /v1/audio/transcriptions actually expects (unlike every other route here, which is JSON).
func (s *suiteState) requestTranscription(model string) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("model", model); err != nil {
		return err
	}
	if err := writer.WriteField("user", "bdd-tests@example.com"); err != nil {
		return err
	}
	part, err := writer.CreateFormFile("file", "fake.wav")
	if err != nil {
		return err
	}
	if _, err := part.Write([]byte("not real audio, just bytes for the mock to accept")); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, litellmBaseURL+"/v1/audio/transcriptions", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+litellmMasterKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.lastErr = err
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		s.lastErr = fmt.Errorf("/v1/audio/transcriptions returned status %d", resp.StatusCode)
		return s.lastErr
	}
	s.lastErr = nil
	return nil
}

// eventForModelWithin polls the OpenMeter mock's event list until a litellm_tokens event
// for the given model, timestamped after this scenario started, shows up (or the deadline hits).
func (s *suiteState) eventForModelWithin(model string, seconds int) error {
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	var lastSeen []openmeterEvent

	for time.Now().Before(deadline) {
		resp, err := http.Get(openmeterBaseURL + "/api/v1/events?limit=20")
		if err != nil {
			return err
		}
		var body struct {
			Events []openmeterEvent `json:"events"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()
		lastSeen = body.Events

		for _, e := range body.Events {
			if e.Time.Before(s.scenarioStart) {
				continue
			}
			if got, _ := e.Data["model"].(string); got == model {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("no litellm_tokens event for model %q appeared within %ds (last seen: %d events)", model, seconds, len(lastSeen))
}

// rawResponseForModelWithin polls openmeter-mock's raw-responses list (populated by our own
// custom LiteLLM callback, litellm/custom_callback.py) until an entry for the given model,
// received after this scenario started, shows up (or the deadline hits).
func (s *suiteState) rawResponseForModelWithin(model string, seconds int) error {
	deadline := time.Now().Add(time.Duration(seconds) * time.Second)
	var lastSeen []rawResponse

	for time.Now().Before(deadline) {
		resp, err := http.Get(openmeterBaseURL + "/api/v1/raw-responses?limit=20")
		if err != nil {
			return err
		}
		var body struct {
			Responses []rawResponse `json:"raw_responses"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()
		lastSeen = body.Responses

		for _, r := range body.Responses {
			if r.ReceivedAt.Before(s.scenarioStart) {
				continue
			}
			if r.Model == model {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("no raw response for model %q appeared within %ds (last seen: %d responses)", model, seconds, len(lastSeen))
}

// printEvents fetches the current events straight from the OpenMeter mock (backed by Mongo)
// and prints them, so a scenario run makes it easy to eyeball what actually landed in the DB.
func (s *suiteState) printEvents() error {
	resp, err := http.Get(openmeterBaseURL + "/api/v1/events?limit=20")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var body struct {
		Events []openmeterEvent `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	fmt.Printf("      --- %d event(s) in MongoDB (openmeter_mock.events) ---\n", len(body.Events))
	for _, e := range body.Events {
		data, _ := json.Marshal(e.Data)
		fmt.Printf("      [%s] id=%s type=%s data=%s\n", e.Time.Format(time.RFC3339), e.ID, e.Type, string(data))
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	s := &suiteState{}

	sc.Step(`^the LiteLLM proxy and the OpenMeter mock are reachable$`, s.reachable)
	sc.Step(`^I request a chat completion from LiteLLM for model "([^"]*)" with prompt "([^"]*)"$`, s.requestChatCompletion)
	sc.Step(`^I request a response from LiteLLM for model "([^"]*)" with prompt "([^"]*)"$`, s.requestResponse)
	sc.Step(`^I request an image generation from LiteLLM for model "([^"]*)" with prompt "([^"]*)"$`, s.requestImageGeneration)
	sc.Step(`^I request an audio transcription from LiteLLM for model "([^"]*)"$`, s.requestTranscription)
	sc.Step(`^I request audio speech from LiteLLM for model "([^"]*)" with input "([^"]*)"$`, s.requestSpeech)
	sc.Step(`^I request a video generation from LiteLLM for model "([^"]*)" with prompt "([^"]*)"$`, s.requestVideoGeneration)
	sc.Step(`^an OpenMeter event for model "([^"]*)" should appear within (\d+) seconds$`, s.eventForModelWithin)
	sc.Step(`^a raw response for model "([^"]*)" should appear within (\d+) seconds$`, s.rawResponseForModelWithin)
	sc.Step(`^I print the OpenMeter events stored in MongoDB$`, s.printEvents)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
