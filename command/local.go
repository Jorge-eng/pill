package command

import (
	"bufio"
	"fmt"
	"github.com/mitchellh/cli"
	"log"
	"os"
	"strings"
)

type LocalCommand struct {
	Ui cli.ColoredUi
}

func (c *LocalCommand) Help() string {
	return "local"
}

func (c *LocalCommand) Synopsis() string {
	return "does something locally"
}

func (c *LocalCommand) Run(args []string) int {

	redshift := make(map[string][]string, 0)

	file, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, "|")

		if len(parts) != 2 {
			log.Fatalf("bad: %s\n", line)
		}

		sn := parts[1]
		deviceId := parts[0]
		if sn == "" || len(sn) != len("90500007A01151701858") {
			continue
		}

		v, found := redshift[sn]
		if !found {
			v = make([]string, 0)
		}
		v = append(v, deviceId)
		redshift[sn] = v
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("No scanning error for redshift")

	checkFile, err := os.Open(args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer checkFile.Close()

	checkScanner := bufio.NewScanner(checkFile)
	for checkScanner.Scan() {
		sn := strings.TrimSpace(checkScanner.Text())
		if sn == "" {
			continue
		}
		deviceIds, found := redshift[sn]
		if !found {
			fmt.Println("Not found", sn, deviceIds)
		}
	}

	if err := checkScanner.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(redshift))
	return 0
}
