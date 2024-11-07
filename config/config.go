// Configuration options available to the user. They should be defined in a
// config file in TOML format in the root directory of a project. The config
// file must exist, but all options are optional.

package config

import (
	_ "embed"
	"errors"
	"fmt"
	"slices"

	"github.com/dustin/go-humanize"
	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	// LLM provider. Either `openai`, `anthropic` or `testing`.
	Provider string `toml:"provider,omitempty"`
	// The system message is dynamically added to each request, followed by the
	// context string.
	SystemMessage string `toml:"system-message,multiline,omitempty"`
	// Additional .gitignore patterns for files that should be excluded from the
	// context sent to the LLM.
	Exclude []string `toml:"exclude"`
	// Maximum length of the full context string.
	MaxContextLength int `toml:"max-context-length,omitempty"`
	// Files beyond this limit will be ignored. This value should be defined using
	// SI notation, e.g. 20KB.
	MaxFileSize string `toml:"max-file-size,omitempty"`
	// Conversations beyond this limit will be pruned from the database. -1 can be
	// set to ignore this option.
	MaxConversationHistory int `toml:"max-conversation-history,omitempty"`
	// Openai configuration. Will be used if the `openai` provider is set.
	Openai OpenaiConfig `toml:"openai,omitempty"`
	// Anthropic configuration. Will be used if the `anthropic` provider is set.
	Anthropic AnthropicConfig `toml:"anthropic,omitempty"`
}

type OpenaiConfig struct {
	ApiKeyEnv string `toml:"api-key-env"`
	Model     string `toml:"model"`
}

type AnthropicConfig struct {
	ApiKeyEnv string `toml:"api-key-env"`
	Model     string `toml:"model"`
}

// Provides basic validation.
func (c *Config) Validate() error {
	supportedProviders := []string{"openai", "anthropic", "testing"}

	var errorBag error

	if !slices.Contains(supportedProviders, c.Provider) {
		errorBag = errors.Join(errorBag, fmt.Errorf(`%s is not a supported value for the "%s" configuration value.`, c.Provider, "provider"))
	}

	if c.SystemMessage == "" {
		errorBag = errors.Join(errorBag, fmt.Errorf(`Missing  "system-message" configuration value.`))
	}

	if _, err := humanize.ParseBytes(c.MaxFileSize); err != nil {
		errorBag = errors.Join(errorBag, fmt.Errorf(`Incorrect string representation of bytes for "%s" configuration value.`, "max-file-size"))
	}

	return errorBag
}

func ConfigFromToml(input string) (Config, error) {
	var cfg Config

	err := toml.Unmarshal([]byte(input), &cfg)

	if err != nil {
		return cfg, errors.New("Failed to parse TOML config string.")
	}

	return cfg, nil
}

func (c *Config) ToToml() (string, error) {
	tomlString, err := toml.Marshal(c)

	if err != nil {
		return "", err
	}

	return string(tomlString), nil
}

//go:embed default-system-message.md
var defaultSystemMessage string

func DefaultConfig() Config {
	return Config{
		Provider:               "openai",
		SystemMessage:          defaultSystemMessage,
		Exclude:                []string{"pal.toml"},
		MaxContextLength:       100_000,
		MaxFileSize:            "20KB",
		MaxConversationHistory: 100,
		Openai: OpenaiConfig{
			ApiKeyEnv: "OPENAI_API_KEY",
			Model:     "gpt-4o-mini",
		},
		Anthropic: AnthropicConfig{
			ApiKeyEnv: "ANTHROPIC_API_KEY",
			Model:     "claude-3-5-haiku-latest",
		},
	}
}

// ResolveConfig merges the default configuration with any overrides provided by
// the user in the `pal.toml` file.
//
// It first creates a copy of the default configuration, and then applies any
// non-zero values from the user-provided overrides. This ensures that the
// returned configuration always contains a complete set of valid options, with
// the user’s overrides taking precedence.
//
// The function also performs basic validation on the resolved configuration.
//
// Returns the final, resolved configuration, or an error if any validation
// checks fail.
func ResolveConfig(overrides *Config) (Config, error) {
	conf := DefaultConfig()

	if overrides.SystemMessage != "" {
		conf.SystemMessage = overrides.SystemMessage
	}

	if overrides.Exclude != nil {
		conf.Exclude = overrides.Exclude
	}

	if overrides.Provider != "" {
		conf.Provider = overrides.Provider
	}

	if overrides.MaxContextLength > 0 {
		conf.MaxContextLength = overrides.MaxContextLength
	}

	if overrides.MaxFileSize != "" {
		conf.MaxFileSize = overrides.MaxFileSize
	}

	if overrides.Openai.ApiKeyEnv != "" {
		conf.Openai.ApiKeyEnv = overrides.Openai.ApiKeyEnv
	}

	if overrides.Openai.Model != "" {
		conf.Openai.Model = overrides.Openai.Model
	}

	if overrides.Anthropic.ApiKeyEnv != "" {
		conf.Anthropic.ApiKeyEnv = overrides.Anthropic.ApiKeyEnv
	}

	if overrides.Anthropic.Model != "" {
		conf.Anthropic.Model = overrides.Anthropic.Model
	}

	if overrides.MaxConversationHistory != 0 {
		conf.MaxConversationHistory = overrides.MaxConversationHistory
	}

	err := conf.Validate()

	return conf, err
}
