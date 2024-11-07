package app

import (
	"os"
	"pal/config"
	"pal/llm_provider"
	"pal/persistence"
	"pal/testutil"
	"path"
	"testing"
)

func TestHappyPath(t *testing.T) {
	projectPath, db := instantiateEnvironment(t)

	// First run.
	err := Run([]string{"pal", "--path", projectPath, "Hello"})
	if err != nil {
		t.Error(err)
	}

	// The database should log the user’s message and the message returned by the
	// assistant.
	convo, err := db.FetchRecentConversation()
	if err != nil {
		t.Error(err)
	}

	expectedConvo := persistence.Conversation{Id: 1, Messages: []persistence.Message{
		{Id: 1, Role: "user", Content: "Hello"},
		{Id: 2, Role: "assistant", Content: llm_provider.TestProviderExpectedMessage},
	}}

	testutil.AssertDeepEquals(t, convo, expectedConvo)

	// Second run. Continue the existing conversation.
	if err = Run([]string{"pal", "--path", projectPath, "--continue", "Hello again"}); err != nil {
		t.Error(err)
	}

	// We should get the same conversation, with two additional messages.
	convo, err = db.FetchRecentConversation()
	if err != nil {
		t.Error(err)
	}

	expectedConvo.Messages = append(expectedConvo.Messages, persistence.Message{
		Id:      3,
		Role:    "user",
		Content: "Hello again",
	})

	expectedConvo.Messages = append(expectedConvo.Messages, persistence.Message{
		Id:      4,
		Role:    "assistant",
		Content: llm_provider.TestProviderExpectedMessage,
	})

	testutil.AssertDeepEquals(t, convo, expectedConvo)
}

func TestExitsEarlyIfNoConfigFileInProject(t *testing.T) {
	projectPath, _ := instantiateEnvironment(t)
	configPath := path.Join(projectPath, "pal.toml")
	if err := os.Remove(configPath); err != nil {
		t.Error(err)
	}

	err := Run([]string{"pal", "--path", projectPath, "hello"})
	testutil.AssertDeepEquals(t, err.Error(), "Failed to open config file.")
}

func TestExitsEarlyIfContextExceedsLimit(t *testing.T) {
	projectPath, _ := instantiateEnvironment(t)
	conf, err := config.ResolveConfig(&config.Config{
		Provider:         "testing",
		MaxContextLength: 10,
	})

	if err != nil {
		t.Error(err)
	}

	if err := saveConfigToFile(projectPath, conf); err != nil {
		t.Error(err)
	}

	testFilePath := path.Join(projectPath, "README.md")
	if err = os.WriteFile(testFilePath, []byte("Hello, world!"), 0755); err != nil {
		t.Error(err)
	}

	err = Run([]string{"pal", "--path", projectPath, "hello"})

	if err == nil {
		t.Error("Should throw an error if context is too long.")
	}
}

func TestPrunesOldConversations(t *testing.T) {
	projectPath, client := instantiateEnvironment(t)
	conf, err := config.ResolveConfig(&config.Config{
		Provider:               "testing",
		MaxConversationHistory: 1,
	})

	if err != nil {
		t.Error(err)
	}

	if err := saveConfigToFile(projectPath, conf); err != nil {
		t.Error(err)
	}

	for i := 1; i <= 10; i++ {
		client.InitializeConversation()
	}

	if err = Run([]string{"pal", "--path", projectPath, "hello"}); err != nil {
		t.Error(err)
	}

	var count int
	client.Conn.QueryRow("select count(*) from conversations").Scan(&count)

	if count != 1 {
		t.Errorf("After prunning, the conversation count should be 1 (actual count: %d).", count)
	}
}

func instantiateEnvironment(t *testing.T) (string, persistence.DatabaseClient) {
	projectPath := t.TempDir()

	conf, err := config.ResolveConfig(&config.Config{Provider: "testing"})
	if err != nil {
		t.Error(err)
	}

	if err := saveConfigToFile(projectPath, conf); err != nil {
		t.Error(err)
	}

	client, err := persistence.StartClient(projectPath)
	if err != nil {
		t.Error(err)
	}

	return projectPath, client
}

func saveConfigToFile(projectPath string, conf config.Config) error {
	toml, err := conf.ToToml()
	if err != nil {
		return err
	}

	configPath := path.Join(projectPath, "pal.toml")

	if err = os.WriteFile(configPath, []byte(toml), 0755); err != nil {
		return err
	}

	return nil
}
