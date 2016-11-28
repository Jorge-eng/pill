package command

import (
	"github.com/mitchellh/cli"
	"io/ioutil"
	"strings"
	// "os"
	"fmt"
	"encoding/hex"
)

type EncryptCommand struct {
	Ui  cli.ColoredUi
	Key string
}

func (c *EncryptCommand) Help() string {
	helpText := `Usage: pill upload {path to blob}`
	return strings.TrimSpace(helpText)
}

func (c *EncryptCommand) Run(args []string) int {
	file, _ := ioutil.ReadFile(args[0])

	
	// factoryKey, err := hex.DecodeString(c.Key)
	// if err != nil {
	// 	c.Ui.Error(fmt.Sprintf("%v", err))
	// 	return 0
	// }


	deviceId := hex.EncodeToString(file[:8])
	hardwareKey := hex.EncodeToString(file[32+128:32+128+16])
	c.Ui.Info(fmt.Sprintf("Device Id: %s", deviceId))
	c.Ui.Info(fmt.Sprintf("Hardware Key: %s", hardwareKey))
	return 0
}

func (c *EncryptCommand) Synopsis() string {
	return "uploads a single blob"
}
