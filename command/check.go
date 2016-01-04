package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"path"
	"strings"
)

type CheckCommand struct {
	Ui  cli.ColoredUi
	Key string
}

func (c *CheckCommand) Help() string {
	helpText := `Usage: pill check {dir} {snToFind}`
	return strings.TrimSpace(helpText)
}

func (c *CheckCommand) Run(args []string) int {
	fileInfos, err := ioutil.ReadDir(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}

	snToFind := args[1]
	c.Ui.Info("Checking for " + snToFind)

	for _, fileInfo := range fileInfos {

		if strings.HasSuffix(fileInfo.Name(), "zip") {
			path := path.Join(args[0], fileInfo.Name())
			err := check(path, snToFind, c.Key)
			if err != nil {
				c.Ui.Error(fmt.Sprintf("%s %v", path, err))
			}
		}
	}
	return 0
}

func (c *CheckCommand) Synopsis() string {
	return "processes all zip files in current directory"
}
