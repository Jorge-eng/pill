package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"strings"
)

type SingleCommand struct {
	Ui cli.ColoredUi
}

func (c *SingleCommand) Help() string {
	helpText := `Usage: pill singe {pathToZipFile}`
	return strings.TrimSpace(helpText)
}

func (c *SingleCommand) Run(args []string) int {
	zipFileName := args[0]
	if !strings.HasSuffix(zipFileName, "zip") {
		c.Ui.Warn("Not valid zip file")
	}

	failedUploads, err := process(zipFileName)
	if len(failedUploads) > 0 {
		for _, failed := range failedUploads {
			c.Ui.Warn(failed)
		}
	}

	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}

	c.Ui.Info(fmt.Sprintf("Processed: %s", zipFileName))
	return 0
}

func (c *SingleCommand) Synopsis() string {
	return "processes the given zip file"
}
