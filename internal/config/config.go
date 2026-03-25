package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        int
	DataDir     string
	LLMProvider string
	GeminiKey   string
	OpenAIKey   string
}

func Load() (Config, error) {
	port, err := intEnv("OMNI_PORT", 8080)
	if err != nil {
		return Config{}, fmt.Errorf("OMNI_PORT: %w", err)
	}

	cfg := Config{
		Port:        port,
		DataDir:     stringEnv("OMNI_DATA_DIR", "/data/omni"),
		LLMProvider: stringEnv("LLM_PROVIDER", "gemini"),
		GeminiKey:   stringEnv("GEMINI_API_KEY", ""),
		OpenAIKey:   stringEnv("OPENAI_API_KEY", ""),
	}

	return cfg, nil
}

func (c Config) DBPath() string  { return c.DataDir + "/omni.db" }
func (c Config) VecPath() string { return c.DataDir + "/vectors.bin" }
func (c Config) DocsDir() string { return c.DataDir + "/documents" }

func stringEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func intEnv(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q", v)
	}
	return n, nil
}
