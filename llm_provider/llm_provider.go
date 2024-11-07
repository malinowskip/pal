package llm_provider

import (
	"os"
	"pal/config"
)

// A provider should act as a proxy to some LLM provider, such as "openai",
// "anthropic" or different. Its only task is to implement the GetCompletion
// function.
type LLMProvider interface {
	// Get a completion from an LLM. The function should pass the system message
	// (already including the context) to the LLM.
	GetCompletion(
		fullSystemMessage string,
		messages []Message,
		handleTokens func(tokens string) error,
	) error
}

type Message struct {
	Role    string
	Content string
}

func ResolveFromConfig(conf *config.Config) (LLMProvider, error) {
	var llmProvider LLMProvider
	var err error

	if conf.Provider == "testing" {
		llmProvider = &TestLLMProvider{}
	}

	if conf.Provider == "openai" {
		apiKey := os.Getenv(conf.Openai.ApiKeyEnv)
		model := conf.Openai.Model
		llmProvider = NewOpenAILLMProvider(apiKey, model)
	}

	if conf.Provider == "anthropic" {
		apiKey := os.Getenv(conf.Anthropic.ApiKeyEnv)
		model := conf.Anthropic.Model
		llmProvider = NewAnthropicLLMProvider(apiKey, model)
	}

	return llmProvider, err
}
