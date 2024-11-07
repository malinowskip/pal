package llm_provider

import (
	"testing"
)

func TestTestLLMProvider(t *testing.T) {
	provider := TestLLMProvider{}

	fullSystemMessage := "Hello, world!"
	var messages []Message

	var receivedMessage string

	provider.GetCompletion(fullSystemMessage, messages, func(tokens string) error {
		receivedMessage = receivedMessage + tokens
		return nil
	})

	if receivedMessage != TestProviderExpectedMessage {
		t.Error("The message received does not match the expected message.")
	}
}
