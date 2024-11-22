# Pal

Pal is a command-line tool that facilitates conversations with language models
(LLMs) about your project’s source code and documentation.

Pal may be a good fit if you want to include a predetermined set of documents as
context for each request. By default, all text files in a project are sent to
the language model, except for those matching the provided “exclude”
patterns or excluded via `.gitignore` files.

To use Pal, you need to install the CLI tool and initialize a `pal.toml`
configuration file in your project's root directory. Then you can start
discussing your codebase with the LLM.

## Installation

Currently, the installation is supported only via the Go toolchain. The
following command will install `pal` on your system.

```sh
go install github.com/malinowskip/pal@latest
```

## Getting started

To get started, navigate to the root directory of your project and run the
following command to initialize your project:

```sh
pal init
```

This will create a `pal.toml` configuration file, as well as system files (for
keeping conversation history) in the local `.pal` directory.

### Verify the configuration file

In the `pal.toml` file, you’ll need to select a provider (either `openai` or
`anthropic`) and make sure you have stored the corresponding API key in the
environment variable specified in the config file. The default provider is `openai`.

Configuration is entirely optional and not necessary, unless you wish to
override the default options.

### Analyze context size and expected token usage

Finally, you should run the `pal analyze` command in your project. This will
provide an overview of the context that would be sent to the LLM, including:

- The total number of characters that would be included in the context (which
  you can use to estimate the number of input tokens).
- A list of the largest files that would be part of the context.

Running `pal analyze` can help you understand the size of the context and
identify any files that may be contributing significantly to the overall context
length. This information can be useful when configuring the
`max-context-length`, `max-file-size`, and `exclude` options in your `pal.toml`
file.

## Usage

You can start a conversation with the LLM by running the following command from
your project’s root directory:

```sh
pal "Describe the provided context in one paragraph"
```

To continue the most recent conversation, use the `-c` (or `--continue`) flag:

```sh
pal -c "Tell me more about the possible configuration options in one paragraph"
```

Alternatively, you can specify the project path using the `-p` (or
`--project-path`) flag:

```sh
pal -p path/to/your/project "Hello, world!"
```

## Managing the context size

By default, Pal will load all files in the project directory as context, **excluding**:

- Files matching patterns defined in the `exclude` configuration option.
- Files matching patterns defined in `.gitignore` files.
- Files whose size exceeds the `max-file-size` configuration option.
- Files that are not valid UTF-8 (such as images).
- The `.git` and `.pal` directories.

If the entire context exceeds the `max-context-length` configuration option, the
program will exit.

## Configuration

The `pal.toml` configuration file supports the following options:

- `provider`: The LLM provider to use, either `openai` or `anthropic` (default: `openai`).
- `system-message`: The initial system message. Context will be appended to it
  dynamically on each request. The default system message is defined
  [here](./config/default-system-message.md).
- `exclude`: A list of additional `.gitignore` glob patterns for paths to be excluded from the context.
- `max-context-length`: The maximum length (in characters) of the context sent to the LLM (default: `100000`).
- `max-file-size`: Files exceeding this size will be ignored (default: `20KB`).
- `max-conversation-history`: Older conversations beyond the specified limit
  will be pruned from the database (defualt: `100`). Can be set to `-1` to disable pruning.
- `openai.api-key-env`: The environment variable containing the OpenAI API key (default: `OPENAI_API_KEY`).
- `openai.model`: The OpenAI model to use (default: `gpt-4o-mini`).
- `anthropic.api-key-env`: The environment variable containing the Anthropic API key (default: `ANTHROPIC_API_KEY`).
- `anthropic.model`: The Anthropic model to use (default: `claude-3-5-haiku-latest`).

All options are optional and can be used to override the defaults.
