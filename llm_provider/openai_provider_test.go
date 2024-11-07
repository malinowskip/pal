package llm_provider

import (
	"testing"
)

func TestOpenAILLMProviderCreation(t *testing.T) {
	apikey := "key"
	model := "gpt-4o-mini"
	provider := NewOpenAILLMProvider(apikey, model)

	if provider.apiKey != apikey {
		t.Error("The provider is missing the api key")
	}

	if provider.model != model {
		t.Error("The provider is missing the model name")
	}

}

func TestOpenaiModelResolution(t *testing.T) {
	mappings := map[string]string{
		"gpt-4o-mini":              "gpt-4o-mini",
		"gpt-4o":                   "gpt-4o",
		"4o":                       "gpt-4o",
		"4o-mini":                  "gpt-4o-mini",
		"should-be-passed-through": "should-be-passed-through",
	}

	for input, expectedOutput := range mappings {
		if resolveOpenaiModel(input) != expectedOutput {
			t.Errorf("%s should be resolved to %s", input, expectedOutput)
		}
	}
}
