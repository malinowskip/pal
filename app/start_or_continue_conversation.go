package app

import (
	"errors"
	"fmt"
	"github.com/malinowskip/pal/config"
	"github.com/malinowskip/pal/documents"
	"github.com/malinowskip/pal/llm_provider"
	"github.com/malinowskip/pal/persistence"
	"github.com/malinowskip/pal/util"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"
)

// This command allows the user to talk to an LLM about their project by
// starting a new conversation or continuing an existing one.
func StartOrContinueConversation(c *cli.Context) error {
	// The user must provide a message from up to two sources: stdin or --message
	// argument. If both are provided, they will be concatenated.
	userMessage, err := fetchUserMessage(c)
	if err != nil {
		return err
	}

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

	// After initialization, the LLM provider should be ready to generate
	// completions. However, the initialization itself doesn’t send any external
	// requests yet, so potential errors might be returned later on, when we
	// request a chat completion (e.g. if the user provides an invalid API key).
	provider, err := llm_provider.ResolveFromConfig(&finalConfig)
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

	// Exit if the context is too long.
	if err = checkContextLength(context, finalConfig.MaxContextLength); err != nil {
		return err
	}

	// System message followed by the context string.
	fullSystemMessage := fmt.Sprintf("%s\n\n%s", finalConfig.SystemMessage, context)

	// Initialize database connection for saving and retrieving conversations from
	// the local database.
	db, err := persistence.StartClient(projectPath)
	if err != nil {
		return err
	}

	// PATH 1: start a new conversation
	//
	// If the --continue flag is not set or if there are no conversations in the
	// database, we will start a new conversation and store its contents in the
	// database.
	startNewConversation := func() error {
		// Since this is a new conversation, only one message will be passed to the
		// LLM.
		messages := []llm_provider.Message{
			{Role: "user", Content: userMessage},
		}

		// Nil pointer to the assistant’s upcoming reply in the database. It will be
		// initiated only after the first batch of tokens is received.
		var dbAssistantReply *persistence.Message

		err = provider.GetCompletion(fullSystemMessage, messages, func(tokens string) error {
			if dbAssistantReply == nil {
				dbConversation, err := db.InitializeConversation()
				if err != nil {
					return err
				}
				_, err = db.InsertMessageIntoConversation(
					dbConversation.Id,
					"user",
					userMessage,
				)
				if err != nil {
					return err
				}
				assistantMsg, err := db.InsertMessageIntoConversation(
					dbConversation.Id,
					"assistant",
					"",
				)
				if err != nil {
					return err
				}
				dbAssistantReply = &assistantMsg
			}

			fmt.Print(tokens)
			if err = db.WriteToMessage(dbAssistantReply.Id, tokens); err != nil {
				return err
			}

			return nil
		})

		return err
	}

	// PATH 2: continue an existing conversation.
	//
	// If the --continue flag is set, attempt to fetch the most recent conversation
	// from the database and continue it. Otherwise, fall back to
	// `startNewConversation`.
	continueLastConversation := func() error {
		// Attempt to retrieve the most recent converastion from the database or just
		// start a new conversation on error.
		recentConversation, err := db.FetchRecentConversation()
		if err != nil {
			return startNewConversation()
		}

		// Messages to be sent to the LLM. This will include existing messages in the
		// conversation, followed by the current message.
		var messages []llm_provider.Message

		// Include messages retrieved from the database.
		for _, m := range recentConversation.Messages {
			messages = append(messages, llm_provider.Message{Role: m.Role, Content: m.Content})
		}

		// Include the current message.
		messages = append(messages, llm_provider.Message{Role: "user", Content: userMessage})

		// Nil pointer to the assistant’s upcoming reply in the database. It will be
		// initiated only after the first batch of tokens is received.
		var dbAssistantReply *persistence.Message

		err = provider.GetCompletion(fullSystemMessage, messages, func(tokens string) error {
			if dbAssistantReply == nil {
				_, err = db.InsertMessageIntoConversation(recentConversation.Id, "user", userMessage)
				if err != nil {
					return err
				}

				assistantMsg, err := db.InsertMessageIntoConversation(
					recentConversation.Id,
					"assistant",
					"",
				)
				if err != nil {
					return err
				}
				dbAssistantReply = &assistantMsg
			}

			fmt.Print(tokens)
			if err = db.WriteToMessage(dbAssistantReply.Id, tokens); err != nil {
				return err
			}

			return nil
		})

		return err
	}

	if c.Bool("continue") == true {
		err = continueLastConversation()
	} else {
		err = startNewConversation()
	}

	if err != nil {
		return err
	}

	if finalConfig.MaxConversationHistory > -1 {
		if err = db.PruneOldConversations(finalConfig.MaxConversationHistory); err != nil {
			return err
		}
	}

	return err
}

// Fetch the message provided by the user. The message might come from up to two
// sources: the --message flag and stdin.
func fetchUserMessage(c *cli.Context) (string, error) {
	// Contains up to two elements: message from stdin and message from the
	// --message argument.
	var messageComponents []string

	stdinText, err := util.ReadStdin()

	if err != nil {
		return "", err
	}

	if len(stdinText) > 0 {
		messageComponents = append(messageComponents, stdinText)
	}

	messageFromArgs := c.Args().First()

	if len(messageFromArgs) > 0 {
		messageComponents = append(messageComponents, messageFromArgs)
	}

	if len(stdinText) == 0 && len(messageFromArgs) == 0 {
		return "", fmt.Errorf("The message cannot be empty.")
	}

	finalMessage := strings.Join(messageComponents, "\n\n")

	return finalMessage, nil
}

func resolveFinalConfig(projectPath string) (config.Config, error) {
	userConfig, err := fetchUserConfig(projectPath)

	if err != nil {
		return userConfig, err
	}

	finalConfig, err := config.ResolveConfig(&userConfig)

	if err != nil {
		return finalConfig, errors.Join(fmt.Errorf("The config is invalid."), err)
	}

	return finalConfig, nil
}

func checkContextLength(input string, maxLength int) error {
	if len(input) > maxLength {
		return fmt.Errorf(
			`Context length (%d) exceeds the maximum permitted context (%d), configurable by setting the "max-context-length" configuration setting (counted in characters).`,
			len(input),
			maxLength,
		)
	}

	return nil
}
