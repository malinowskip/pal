package app

import (
	"os"
	"pal/config"
	"pal/documents"
	"path"
	"testing"
)

func TestAssembleContextString(t *testing.T) {
	var docs []documents.Document
	docs = append(docs, documents.Document{Path: "README.md", Content: "Hello, world!\n"})

	expected := "<documents>\n<document>\n<source>README.md</source>\n<document_content>\nHello, world!\n</document_content>\n</document>\n</documents>"
	context, _ := assembleContextString(&docs)

	if context != expected {
		t.Errorf(
			"Context built does not match expected context.\n\nBuilt:\n\n%v\n\nExpected:\n\n%v",
			context,
			expected,
		)
	}
}

func TestFetchUserConfig(t *testing.T) {
	projectPath := t.TempDir()
	configFile, _ := os.Create(path.Join(projectPath, "pal.toml"))
	conf := config.DefaultConfig()
	conf.SystemMessage = "yes"
	toml, _ := conf.ToToml()
	configFile.WriteString(toml)
	resolvedConf, _ := fetchUserConfig(projectPath)

	if resolvedConf.SystemMessage != "yes" {
		t.Error("Failed to fetch config from the filesystem.")
	}
}

func TestInitConfigFile(t *testing.T) {
	t.Run("Happy path.", func(t *testing.T) {
		projectPath := t.TempDir()
		file, _ := initConfigFile(projectPath, "")
		byteContents, _ := os.ReadFile(file.Name())
		contents := string(byteContents)

		expectedConfig := config.Config{
			Provider:         "openai",
			Exclude:          []string{"pal.toml"},
			MaxContextLength: config.DefaultConfig().MaxContextLength,
			MaxFileSize:      config.DefaultConfig().MaxFileSize,
			Openai: config.OpenaiConfig{
				ApiKeyEnv: "OPENAI_API_KEY",
				Model:     "gpt-4o-mini",
			},
			Anthropic: config.AnthropicConfig{
				ApiKeyEnv: "ANTHROPIC_API_KEY",
				Model:     "claude-3-5-haiku-latest",
			},
		}

		expected, _ := expectedConfig.ToToml()

		if contents != expected {
			t.Errorf("Generated config does not match the expected config. Generated: \n%s\n\nExpected: \n%s\n", contents, expected)
		}
	})

	t.Run("Fails if config file already exists.", func(t *testing.T) {
		projectPath := t.TempDir()
		initConfigFile(projectPath, "")

		_, err := initConfigFile(projectPath, "")

		if err == nil {
			t.Error("InitConfig succeeded even though the config file already existed.")
		}
	})

}
