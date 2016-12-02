package command

import (
	"archive/zip"
	"bufio"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/meirf/gopart"
	"github.com/mitchellh/cli"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type ExtractCommand struct {
	Ui  cli.ColoredUi
	Key string
}

func (c *ExtractCommand) Help() string {
	helpText := `Usage: pill extract {pathToDir} {path to file containing SN to check}`
	return strings.TrimSpace(helpText)
}

type Extracter struct {
	sn        map[string]bool
	fileInfos []os.FileInfo
	path      string
	key       string
	Ui        cli.ColoredUi
	deviceIds []Pair
}

type Pair struct {
	DeviceId string
	Metadata string
}

type Conflict struct {
	Stored string
	Sent   string
}

func (e *Extracter) Run() error {
	for _, fileInfo := range e.fileInfos {
		if strings.HasSuffix(fileInfo.Name(), "zip") {
			path := path.Join(e.path, fileInfo.Name())

			err := e.extract(path, e.key)
			if err != nil {
				return err
			}
		}
	}

	if len(e.sn) > 0 {
		for sn, _ := range e.sn {
			e.Ui.Error(fmt.Sprintf("missing: %s", sn))
		}
	}

	return nil
}

func (c *ExtractCommand) Run(args []string) int {
	fileInfos, err := ioutil.ReadDir(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}

	file, err := os.Open(args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	sns := make(map[string]bool, 0)
	for scanner.Scan() {
		sns[scanner.Text()] = true
	}

	fmt.Println("SNS = ", len(sns))
	if err := scanner.Err(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	extracter := &Extracter{
		sn:        sns,
		fileInfos: fileInfos,
		path:      args[0],
		key:       c.Key,
		Ui:        c.Ui,
	}

	err = extracter.Run()
	if err != nil {
		c.Ui.Error(err.Error())
		return 1
	}

	c.Ui.Info(fmt.Sprintf("%d device ids", len(extracter.deviceIds)))

	sess, err := session.NewSession()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("failed to create session: %s", err))
		return 1
	}
	config := &aws.Config{
		Region: aws.String("us-east-1"),
	}
	svc := dynamodb.New(sess, config)

	partitionSize := 99

	allDeviceIds := make(map[string]Conflict)
	for _, pair := range extracter.deviceIds {
		allDeviceIds[pair.DeviceId] = Conflict{Sent: pair.Metadata}
	}

	for idxRange := range gopart.Partition(len(extracter.deviceIds), partitionSize) {
		deviceIds := extracter.deviceIds[idxRange.Low:idxRange.High]
		keys := make([]map[string]*dynamodb.AttributeValue, 0)
		fmt.Printf("Range: %d, %d\n", idxRange.Low, idxRange.High)
		for _, pair := range deviceIds {
			m := map[string]*dynamodb.AttributeValue{
				"device_id": {
					S: aws.String(pair.DeviceId),
				},
			}
			keys = append(keys, m)
		}
		params := &dynamodb.BatchGetItemInput{
			RequestItems: map[string]*dynamodb.KeysAndAttributes{ // Required
				"pill_key_store": { // Required
					Keys: keys,
					AttributesToGet: []*string{
						aws.String("device_id"),
						aws.String("metadata"),
					},
				},
			},
		}

		resp, err := svc.BatchGetItem(params)
		if err != nil {
			c.Ui.Error(err.Error())
			return 1
		}

		table, _ := resp.Responses["pill_key_store"]
		fmt.Println("Got results", len(table))
		for _, item := range table {
			pillId := aws.StringValue(item["device_id"].S)
			conflict, _ := allDeviceIds[pillId]
			conflict.Stored = aws.StringValue(item["metadata"].S)
			allDeviceIds[pillId] = conflict
		}
	}

	for serial, _ := range extracter.sn {
		c.Ui.Warn(serial)
	}

	fmt.Println("len alldevice", len(allDeviceIds))
	fmt.Println("pill\tstored\tsent")
	for pillId, conflict := range allDeviceIds {
		if conflict.Stored == "" {
			fmt.Printf("%s\t%s\t%s\n", pillId, conflict.Sent, "not provisioned")
		} else {
			fmt.Printf("%s\t%s\t%s\n", pillId, conflict.Stored, conflict.Sent)
		}
	}
	return 0

}

func (c *ExtractCommand) Synopsis() string {
	return "extracts all device ids for zip files in given directory and compares missing SN"
}

func (e *Extracter) extract(archive, key string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		fname := file.FileInfo().Name()

		_, found := e.sn[fname]
		if !found {
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		buff, err := ioutil.ReadAll(fileReader)
		if err != nil {
			return err
		}

		blob, parseErr := parse(buff, key)
		if parseErr != nil {
			return parseErr
		}
		pair := Pair{
			DeviceId: blob.DeviceId,
			Metadata: fname,
		}
		e.deviceIds = append(e.deviceIds, pair)

		delete(e.sn, fname)
	}
	return nil
}
