package main

import (
	"github.com/hello/pill/command"
	"github.com/mitchellh/cli"
	"os"
	"os/signal"
)

// Commands is the mapping of all the available cli commands
var Commands map[string]cli.CommandFactory

var (
	UiColorBlack = cli.UiColor{37, false}
)

func init() {
	cui := cli.ColoredUi{
		InfoColor:  cli.UiColorGreen,
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		Ui: &cli.BasicUi{
			Writer: os.Stdout,
			Reader: os.Stdin,
		},
	}

	Commands = map[string]cli.CommandFactory{

		"download": func() (cli.Command, error) {
			return &command.DownloadCommand{
				Ui: cui,
			}, nil
		},
		"process": func() (cli.Command, error) {
			return &command.ProcessCommand{
				Ui: cui,
			}, nil
		},
		"single": func() (cli.Command, error) {
			return &command.SingleCommand{
				Ui: cui,
			}, nil
		},
		"local": func() (cli.Command, error) {
			return &command.LocalCommand{
				Ui: cui,
			}, nil
		},
		"check": func() (cli.Command, error) {
			return &command.CheckCommand{
				Ui:  cui,
				Key: FactoryKey,
			}, nil
		},
		"encrypt": func() (cli.Command, error) {
			return &command.EncryptCommand{
				Ui:  cui,
				Key: FactoryKey,
			}, nil
		},
		"search": func() (cli.Command, error) {
			return &command.SearchCommand{
				Ui:  cui,
				Key: FactoryKey,
			}, nil
		},
		"clean": func() (cli.Command, error) {
			return &command.CleanCommand{
				Ui: cui,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Ui:     cui,
				GitSha: GitSha,
			}, nil
		},
	}
}

// makeShutdownCh returns a channel that can be used for shutdown
// notifications for commands. This channel will send a message for every
// interrupt received.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
