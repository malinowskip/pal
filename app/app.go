package app

import (
	"errors"
	"fmt"
	"os"
	"pal/config"
	"pal/constants"
	"pal/documents"
	"pal/util"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

// Entry point for the CLI application. It should be executed in `main.go`, with
// `os.Args` passed as an argument.
func Run(args []string) error {
	return app.Run(args)
}

var app = &cli.App{
	Name:  constants.AppName,
	Usage: "Talk to an LLM about your source code and documentation",
	Flags: []cli.Flag{
		&cli.PathFlag{
			Name:    "project-path",
			Usage:   "Specify the path to the project’s root directory",
			Value:   ".",
			Aliases: []string{"p", "path"},
		},
		&cli.BoolFlag{
			Name:    "continue",
			Usage:   "Continue the most recent conversation",
			Value:   false,
			Aliases: []string{"c"},
		},
	},
	Action: StartOrContinueConversation,
	Commands: []*cli.Command{
		{
			Name:   "init",
			Usage:  "Initializes a new project",
			Action: InitProject,
		},
		{
			Name:   "config",
			Usage:  "Resolves and prints final configuration",
			Action: PrintConfig,
		},
		{
			Name:   "analyze",
			Usage:  "Prints useful information on context length",
			Action: Analyze,
		},
	},
}

// Formats documents into a single string organized using XML tags – as
// recommended by Anthropic in “[Long context prompting tips]”. The output
// can be included in the conversation as context.
//
// [Long context prompting tips]: https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/use-xml-tags
func assembleContextString(docs *[]documents.Document) (string, error) {
	var output strings.Builder

	output.WriteString("<documents>\n")

	for _, doc := range *docs {
		output.WriteString("<document>\n")
		output.WriteString(fmt.Sprintf("<source>%s</source>\n", doc.Path))
		output.WriteString("<document_content>\n")
		output.WriteString(doc.Content)
		output.WriteString("</document_content>\n")
		output.WriteString("</document>\n")
	}

	output.WriteString("</documents>")
	return output.String(), nil
}

// Loads and parses the `pal.toml` configuration file from the given project
// path. Returns the parsed Config and any errors encountered while reading or
// parsing the file.
func fetchUserConfig(projectPath string) (config.Config, error) {
	path := path.Join(projectPath, "pal.toml")

	configFile, err := os.Open(path)

	if err != nil {
		return config.Config{}, fmt.Errorf("Failed to open config file.")
	}

	configToml, err := os.ReadFile(configFile.Name())

	if err != nil {
		return config.Config{}, fmt.Errorf("Failed to read config file.")
	}

	parsedConfig, err := config.ConfigFromToml(string(configToml))

	if err != nil {
		return config.Config{}, fmt.Errorf("Invalid config file.")
	}

	return parsedConfig, nil
}

// Initializes a config file with a minimal set of options.
func initConfigFile(projectPath string, requestedProvider string) (*os.File, error) {
	filePath := path.Join(projectPath, "pal.toml")

	if util.FileExists(filePath) {
		return &os.File{}, fmt.Errorf("Failed to initialize config file. The path already exists.")
	}

	file, err := os.Create(filePath)

	if err != nil {
		return file, errors.Join(fmt.Errorf("Failed to create config file."), err)
	}

	defaultConfig := config.DefaultConfig()

	conf := config.Config{
		Exclude:          defaultConfig.Exclude,
		MaxFileSize:      defaultConfig.MaxFileSize,
		MaxContextLength: defaultConfig.MaxContextLength,
		Openai:           defaultConfig.Openai,
		Anthropic:        defaultConfig.Anthropic,
	}

	if requestedProvider == "anthropic" {
		conf.Provider = "anthropic"
	} else if requestedProvider == "openai" {
		conf.Provider = "openai"
	} else {
		conf.Provider = "openai"
	}

	defaultConfigString, err := conf.ToToml()

	if err != nil {
		return file, fmt.Errorf("Failed to encode default config as TOML.")
	}

	_, err = file.WriteString(defaultConfigString)

	if err != nil {
		return file, fmt.Errorf("Failed to write to config file.")
	}

	return file, err
}
