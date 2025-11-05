// Package agent provides a deep-agent implementation supporting multiple LLM providers
package agent

import (
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// Provider defines the type of LLM provider
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderOllama Provider = "ollama"
)

// Config holds the configuration for the agent
type Config struct {
	Provider Provider
	Model    string
	APIKey   string
	BaseURL  string // For Ollama or custom endpoints
}

// NewAgent creates a new agent with the given configuration
func NewAgent(config Config) (*Agent, error) {
	if config.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	opts := []option.RequestOption{}

	switch config.Provider {
	case ProviderOpenAI:
		if config.APIKey == "" {
			return nil, fmt.Errorf("API key is required for OpenAI")
		}
		opts = append(opts, option.WithAPIKey(config.APIKey))

	case ProviderOllama:
		// Ollama uses OpenAI-compatible API endpoint
		if config.BaseURL == "" {
			config.BaseURL = "http://localhost:11434/v1"
		}
		opts = append(opts,
			option.WithBaseURL(config.BaseURL),
			option.WithAPIKey("ollama"), // Ollama doesn't require real API key
		)

	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}

	client := openai.NewClient(opts...)

	return &Agent{
		config: config,
		client: &client,
	}, nil
}
