package llm_provider

import (
	"testing"
)

func TestAnthropicLLMProviderCreation(t *testing.T) {
	apikey := "key"
	model := "claude-3-5-haiku-latest"
	provider := NewAnthropicLLMProvider(apikey, model)

	if provider.apiKey != apikey {
		t.Error("The provider is missing the api key")
	}

	if provider.model != model {
		t.Error("The provider is missing the model name")
	}

}

func TestAnthropicModelResolution(t *testing.T) {
	mappings := map[string]string{
		"should-be-passed-through": "should-be-passed-through",
		"opus":                     "claude-3-opus-latest",
		"sonnet":                   "claude-3-5-sonnet-latest",
		"haiku":                    "claude-3-5-haiku-latest",
	}

	for input, expectedOutput := range mappings {
		if resolveAnthropicModel(input) != expectedOutput {
			t.Errorf("%s should be resolved to %s", input, expectedOutput)
		}
	}
}
