package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"path"
	"strings"
)

type ProcessCommand struct {
	Ui cli.ColoredUi
}

func (c *ProcessCommand) Help() string {
	helpText := `Usage: pill process {pathToDir}`
	return strings.TrimSpace(helpText)
}

func (c *ProcessCommand) Run(args []string) int {
	fileInfos, err := ioutil.ReadDir(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}

	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), "zip") {
			path := path.Join(args[0], fileInfo.Name())
			failedUploads, err := process(path)
			if len(failedUploads) > 0 {
				for _, failed := range failedUploads {
					c.Ui.Warn(failed)
				}
			}

			if err != nil {
				c.Ui.Error(fmt.Sprintf("%v", err))
				return 1
			}
			c.Ui.Info(path)
		}
	}
	return 0
}

func (c *ProcessCommand) Synopsis() string {
	return "processes all zip files in given directory"
}
