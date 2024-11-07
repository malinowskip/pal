package llm_provider

import (
	"context"

	"github.com/liushuangls/go-anthropic/v2"
)

type AnthropicLLMProvider struct {
	apiKey string
	model  string
}

// Quality-of-life function to support short-hand model names for Anthropic.
func resolveAnthropicModel(input string) string {
	if input == "opus" {
		return "claude-3-opus-latest"
	}

	if input == "sonnet" {
		return "claude-3-5-sonnet-latest"
	}

	if input == "haiku" {
		return "claude-3-5-haiku-latest"
	}

	return input
}

func NewAnthropicLLMProvider(apiKey string, model string) *AnthropicLLMProvider {
	return &AnthropicLLMProvider{
		apiKey: apiKey,
		model:  resolveAnthropicModel(model),
	}
}

func (p *AnthropicLLMProvider) GetCompletion(
	fullSystemMessage string,
	messages []Message,
	handleTokens func(tokens string) error,
) error {
	client := anthropic.NewClient(p.apiKey, anthropic.WithBetaVersion(anthropic.BetaPromptCaching20240731))

	request := anthropic.MessagesStreamRequest{
		MessagesRequest: anthropic.MessagesRequest{
			Model: anthropic.Model(p.model),
			MultiSystem: []anthropic.MessageSystemPart{
				{
					Type: "text",
					Text: fullSystemMessage,
					CacheControl: &anthropic.MessageCacheControl{
						Type: anthropic.CacheControlTypeEphemeral,
					},
				},
			},
			MaxTokens: 1000,
		},
		OnContentBlockDelta: func(data anthropic.MessagesEventContentBlockDeltaData) {
			handleTokens(*data.Delta.Text)
		},
	}

	for _, message := range messages {
		if message.Role == "user" {
			request.MessagesRequest.Messages = append(
				request.MessagesRequest.Messages,
				anthropic.NewUserTextMessage(message.Content),
			)
		} else {
			request.MessagesRequest.Messages = append(
				request.MessagesRequest.Messages,
				anthropic.NewAssistantTextMessage(message.Content),
			)
		}
	}

	_, err := client.CreateMessagesStream(context.Background(), request)

	return err
}
