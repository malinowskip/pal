package llm_provider

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type OpenAILLMProvider struct {
	apiKey string
	model  string
}

type payload struct {
	Model    string
	Messages []message
	Stream   bool
}

type message struct {
	Role    string
	Content string
}

// Quality-of-life function to support short-hand model names for OpenAI.
func resolveOpenaiModel(input string) string {
	if input == "4o-mini" {
		return "gpt-4o-mini"
	}

	if input == "4o" {
		return "gpt-4o"
	}

	return input
}

func NewOpenAILLMProvider(apiKey string, model string) *OpenAILLMProvider {
	return &OpenAILLMProvider{
		apiKey: apiKey,
		model:  resolveOpenaiModel(model),
	}
}

func (p *OpenAILLMProvider) GetCompletion(
	fullSystemMessage string,
	messages []Message,
	handleTokens func(tokens string) error,
) error {
	client := openai.NewClient(p.apiKey)

	finalMessages := buildMessages(fullSystemMessage, messages)

	request := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: finalMessages,
		Stream:   true,
	}

	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		request,
	)

	if err != nil {
		return fmt.Errorf("Unsuccessful request to the OpenAI API: %v", err)
	}

	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			fmt.Printf("\nStream error: %v\n", err)
			return err
		}

		text := response.Choices[0].Delta.Content
		handleTokens(text)
	}
}

func buildMessages(
	fullSystemMessage string,
	messages []Message,
) []openai.ChatCompletionMessage {
	var finalMessages []openai.ChatCompletionMessage

	finalMessages = append(finalMessages, openai.ChatCompletionMessage{Role: "system", Content: fullSystemMessage})

	for _, m := range messages {
		finalMessages = append(finalMessages, openai.ChatCompletionMessage{Role: m.Role, Content: m.Content})
	}

	return finalMessages
}
