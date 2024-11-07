package config

import (
	"pal/testutil"
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	tomlString := `
		exclude = ["some-file.md"]
	`

	conf, _ := ConfigFromToml(tomlString)

	testutil.AssertDeepEquals(t, conf.Exclude, []string{"some-file.md"})
}

func TestEncodeConfig(t *testing.T) {
	conf := Config{
		Exclude: []string{"foo.md"},
	}

	encoded, _ := conf.ToToml()

	decoded, _ := ConfigFromToml(encoded)

	testutil.AssertDeepEquals(t, conf, decoded)
}

func TestDefaultConfig(t *testing.T) {

	conf := DefaultConfig()

	testutil.AssertDeepEquals(t, conf.Exclude, []string{"pal.toml"})
	testutil.AssertDeepEquals(t, conf.Provider, "openai")
	testutil.AssertDeepEquals(t, conf.MaxContextLength, 100_000)
	testutil.AssertDeepEquals(t, conf.MaxFileSize, "20KB")
	testutil.AssertDeepEquals(t, conf.MaxConversationHistory, 100)
	testutil.AssertDeepEquals(t, conf.Openai.ApiKeyEnv, "OPENAI_API_KEY")
	testutil.AssertDeepEquals(t, conf.Openai.Model, "gpt-4o-mini")
	testutil.AssertDeepEquals(t, conf.Anthropic.Model, "claude-3-5-haiku-latest")
	testutil.AssertDeepEquals(t, conf.Anthropic.ApiKeyEnv, "ANTHROPIC_API_KEY")

	t.Run("System message", func(t *testing.T) {
		if len(conf.SystemMessage) == 0 {
			t.Errorf("Missing system message.")
		}
	})
}

func TestResolveConfig(t *testing.T) {

	testOverride := func(t *testing.T, fieldName string, value any) {
		var overrides Config
		reflectedValue := reflect.ValueOf(value)

		reflectedOverrides := reflect.ValueOf(&overrides).Elem()
		reflectedOverrides.FieldByName(fieldName).Set(reflectedValue)

		FinalConfig, _ := ResolveConfig(&overrides)
		finalConfigReflected := reflect.ValueOf(FinalConfig)
		overridenValue := finalConfigReflected.FieldByName(fieldName)

		testutil.AssertDeepEquals(t, reflectedValue.Interface(), overridenValue.Interface())
	}

	testOverride(t, "SystemMessage", "override")
	testOverride(t, "Exclude", []string{"hello"})
	testOverride(t, "Provider", "hello")
	testOverride(t, "Openai", OpenaiConfig{ApiKeyEnv: "hello", Model: "hello"})
	testOverride(t, "Anthropic", AnthropicConfig{ApiKeyEnv: "hello", Model: "hello"})
	testOverride(t, "MaxFileSize", "5KB")
	testOverride(t, "MaxConversationHistory", 5)

	t.Run("Returns default config if overrides are empty.", func(t *testing.T) {
		overrides := Config{}

		FinalConfig, _ := ResolveConfig(&overrides)

		testutil.AssertDeepEquals(t, FinalConfig, DefaultConfig())
	})
}

func TestConfigValidation(t *testing.T) {
	t.Run("Default config is valid", func(t *testing.T) {
		conf := DefaultConfig()
		err := conf.Validate()

		if err != nil {
			t.Errorf("The default config should be valid.")
		}
	})

	t.Run("Incorrect provider", func(t *testing.T) {
		values := []string{"", "some-unsupported-model"}

		for _, value := range values {
			conf := DefaultConfig()
			conf.Provider = value
			if conf.Validate() == nil {
				t.Errorf("%s is not a valid value for the %s field.", value, "Provider")
			}
		}
	})

	t.Run("Missing system message", func(t *testing.T) {
		conf := DefaultConfig()
		conf.SystemMessage = ""
		if conf.Validate() == nil {
			t.Errorf("%s is not a valid value for the %s field.", "", "SystemMessage")
		}
	})

	t.Run("Incorrect MaxFileSize notation", func(t *testing.T) {
		conf := DefaultConfig()
		conf.MaxFileSize = "10XYZ"
		if conf.Validate() == nil {
			t.Errorf("%s is not a valid value for the %s field.", conf.MaxFileSize, "MaxFileSize")
		}
	})
}
