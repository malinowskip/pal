package app

import (
	"fmt"
	"pal/config"

	"github.com/urfave/cli/v2"
)

// This command resolves the final configuration (defaults + user’s overrides)
// and prints them to the console.
func PrintConfig(c *cli.Context) error {
	projectPath := c.Path("project-path")

	userConfig, err := fetchUserConfig(projectPath)

	if err != nil {
		return fmt.Errorf("Failed to read the contents of the config file.")
	}

	finalConfig, err := config.ResolveConfig(&userConfig)

	if err != nil {
		return err
	}

	tomlString, err := finalConfig.ToToml()

	if err != nil {
		return fmt.Errorf("Failed to encode config as TOML.")
	}

	fmt.Println(tomlString)

	return nil
}
