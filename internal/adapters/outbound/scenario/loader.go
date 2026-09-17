package scenario

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// fileScenario mirrors domain.Scenario in a YAML-friendly shape.
type fileScenario struct {
	Model           string `yaml:"model"`
	Operation       string `yaml:"operation"`
	ResponseText    string `yaml:"response_text"`
	LatencyMs       int    `yaml:"latency_ms"`
	TokensPerSecond int    `yaml:"tokens_per_second"`
	Error           *struct {
		StatusCode int    `yaml:"status_code"`
		Message    string `yaml:"message"`
		Type       string `yaml:"type"`
	} `yaml:"error"`
}

type fileConfig struct {
	Scenarios []fileScenario `yaml:"scenarios"`
}

// LoadFile reads a YAML scenario definition file and seeds the given store with its contents.
// Path may be empty, in which case LoadFile is a no-op (the store keeps its benign defaults).
func LoadFile(ctx context.Context, path string, store *Store) error {
	if path == "" {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cfg fileConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return err
	}
	for _, fs := range cfg.Scenarios {
		op := fs.Operation
		if op == "" {
			op = "chat"
		}
		sc := domain.Scenario{
			Key:             Key(fs.Model, op),
			ResponseText:    fs.ResponseText,
			LatencyMs:       fs.LatencyMs,
			TokensPerSecond: fs.TokensPerSecond,
		}
		if fs.Error != nil {
			sc.Error = &domain.ErrorInjection{
				StatusCode: fs.Error.StatusCode,
				Message:    fs.Error.Message,
				Type:       fs.Error.Type,
			}
		}
		if err := store.Put(ctx, sc); err != nil {
			return err
		}
	}
	return nil
}
