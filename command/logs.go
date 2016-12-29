package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"path"
	"strings"
)

type LogsCommand struct {
	Ui cli.ColoredUi
}

func (c *LogsCommand) Help() string {
	helpText := `Usage: pill logs {dir} {deviceId}`
	return strings.TrimSpace(helpText)
}

func (c *LogsCommand) Run(args []string) int {
	fileInfos, err := ioutil.ReadDir(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}

	deviceId := args[1]
	c.Ui.Info("Checking for " + deviceId)
	c.Ui.Info(fmt.Sprintf("Checking in %d files", len(fileInfos)))

	i := 0
	for _, fileInfo := range fileInfos {

		if strings.HasSuffix(fileInfo.Name(), "zip") {
			path := path.Join(args[0], fileInfo.Name())
			res, err := checkLogs(path, deviceId)
			if err != nil {
				c.Ui.Error(fmt.Sprintf("%s %v", path, err))
			}
			if len(res) > 0 {
				for _, r := range res {
					c.Ui.Info(fmt.Sprintf("%s found in %s", r, fileInfo.Name()))
				}
			}
		}
		i += 1
	}
	c.Ui.Info(fmt.Sprintf("checked: %d files", i))
	return 0
}

func (c *LogsCommand) Synopsis() string {
	return "checks for sn in the given directory"
}
