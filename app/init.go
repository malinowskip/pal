package app

import (
	"fmt"
	"github.com/malinowskip/pal/util"
	"path"

	"github.com/urfave/cli/v2"
)

// This command initializes a minimal config file in the project directory.
func InitProject(c *cli.Context) error {
	projectPath := c.Path("project-path")
	requestedProvider := c.Args().First()

	if projectPath == "" {
		return fmt.Errorf("Failed to initialize project. Missing project path.")
	}

	_, err := initConfigFile(projectPath, requestedProvider)

	if err != nil {
		return err
	}

	fmt.Println("Initialization successful! Please review the generated config file:")
	fmt.Printf("  pal.toml\n\n")

	if util.FileExists(path.Join(projectPath, ".gitignore")) {
		fmt.Println("Consider adding .pal to your .gitignore file:")
		fmt.Printf("  echo \".pal\" >> %s\n\n", path.Join(projectPath, ".gitignore"))
	}

	fmt.Println("Please run the following command to analyze your expected token usage:")
	fmt.Printf("  %s --path %s analyze\n", c.App.Name, projectPath)

	return nil
}
