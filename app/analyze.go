package app

import (
	"errors"
	"fmt"
	"github.com/malinowskip/pal/documents"
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// This command outputs information on the context that would be sent to the LLM
// based on the current configuration. Its purpose is to help the user estimate
// expected token usage.
func Analyze(c *cli.Context) error {
	// Root path of the project. If not set by the user, it will be set to the
	// current directory, i.e. ".".
	projectPath := c.Path("project-path")
	if projectPath == "" {
		return fmt.Errorf("The project path may not be empty.")
	}

	// Default config values overridden by any values defined by the user in
	// `pal.toml`.
	finalConfig, err := resolveFinalConfig(projectPath)
	if err != nil {
		return err
	}

	// Maximum size of documents to be included in the context. In the config, this
	// value is specified using SI notation, e.g. “10K”, so it needs to be
	// converted to bytes.
	maxFileSize, err := humanize.ParseBigBytes(finalConfig.MaxFileSize)
	if err != nil {
		return errors.Join(fmt.Errorf("The config is invalid."), err)
	}

	// Load all project documents that will be included in the context.
	documents, err := documents.LoadDocuments(
		projectPath,
		finalConfig.Exclude,
		maxFileSize.Int64(),
	)
	if err != nil {
		return err
	}

	// Concatenate all documents into a single string that will be passed to the
	// LLM at the end of the system message.
	context, err := assembleContextString(&documents)
	if err != nil {
		return err
	}

	p := message.NewPrinter(language.English)

	p.Println(
		"Number of documents that would be included in the context:",
	)

	p.Printf("  %d\n\n", len(documents))

	p.Println(
		"Full context string length:",
	)

	p.Printf(
		"  %d characters\n\n",
		len(context),
	)

	sort.Slice(documents, func(i, j int) bool {
		return len(documents[i].Content) > len(documents[j].Content)
	})

	p.Println("Fifteen largest documents:")

	for i, d := range documents {
		if i > 14 {
			break
		}
		fmt.Printf("  %s\n", d.Path)
	}

	return nil
}
