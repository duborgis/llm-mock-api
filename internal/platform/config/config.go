package config

import "os"

// Config holds env-var driven server configuration.
type Config struct {
	HTTPPort     string
	AdminPort    string
	ScenarioFile string
}

func Load() Config {
	return Config{
		HTTPPort:     getenv("MOCK_HTTP_PORT", "8080"),
		AdminPort:    getenv("MOCK_ADMIN_PORT", "8081"),
		ScenarioFile: getenv("MOCK_SCENARIO_FILE", ""),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
