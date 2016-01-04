package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"strings"
)

type VersionCommand struct {
	Ui     cli.ColoredUi
	GitSha string
}

func (c *VersionCommand) Help() string {
	helpText := `Usage: pill version`
	return strings.TrimSpace(helpText)
}

func (c *VersionCommand) Run(args []string) int {
	c.Ui.Info(fmt.Sprintf("Version: %s", c.GitSha))
	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "outputs version of the build"
}
