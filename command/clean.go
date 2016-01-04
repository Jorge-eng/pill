package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"os"
)

type CleanCommand struct {
	Ui cli.ColoredUi
}

func (c *CleanCommand) Help() string {
	return "warning: deletes zip directory in current directory"
}

func (c *CleanCommand) Synopsis() string {
	return "warning: deletes zip directory in current directory"
}

func (c *CleanCommand) Run(args []string) int {

	resp, err := c.Ui.Ask("This will delete zip/ in current dir. Are you sure? y/n")
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}
	if resp != "y" {
		c.Ui.Info("Doing nothing.")
		return 0
	}

	removeErr := os.RemoveAll(".zip")
	if removeErr != nil {
		c.Ui.Error(fmt.Sprintf("%v", removeErr))
		return 1
	}
	c.Ui.Warn("./zip deleted")
	return 0
}
