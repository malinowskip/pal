package llm_provider

type TestLLMProvider struct{}

const TestProviderExpectedMessage = "Hello, world!"

func (p *TestLLMProvider) GetCompletion(
	fullSystemMessage string,
	messages []Message,
	handleTokens func(tokens string) error,
) error {
	return handleTokens(TestProviderExpectedMessage)
}
