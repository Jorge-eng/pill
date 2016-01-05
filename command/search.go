package command

import (
	"fmt"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"path"
	"strings"
)

type SearchCommand struct {
	Ui  cli.ColoredUi
	Key string
}

func (c *SearchCommand) Help() string {
	helpText := `Usage: pill search {pathToDir} {deviceId}`
	return strings.TrimSpace(helpText)
}

func (c *SearchCommand) Run(args []string) int {
	fileInfos, err := ioutil.ReadDir(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}

	deviceId := args[1]
	c.Ui.Info("Searching for deviceId: " + deviceId)

	found := false
	for _, fileInfo := range fileInfos {
		path := path.Join(args[0], fileInfo.Name())

		if strings.HasSuffix(fileInfo.Name(), "zip") {
			fname, err := search(path, deviceId, c.Key)
			if err != nil {
				c.Ui.Error(fmt.Sprintf("%s: %v", fileInfo.Name(), err))
				continue
			}

			if fname != "" {
				c.Ui.Info(fmt.Sprintf("zip file: %s", fileInfo.Name()))
				c.Ui.Info(fmt.Sprintf("blob name: %s", fname))
				c.Ui.Info(fmt.Sprintf("device_id: %s", deviceId))
				found = true
			}
		}
	}

	if !found {
		c.Ui.Warn(fmt.Sprintf("%s NOT FOUND", deviceId))
	}

	return 0
}

func (c *SearchCommand) Synopsis() string {
	return "search for deviceId in all zip files in given directory"
}
