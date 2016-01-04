package command

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	getter "github.com/hashicorp/go-getter"
	"github.com/mitchellh/cli"
	"net/url"
	"strings"
	"time"
)

type DownloadCommand struct {
	Ui cli.ColoredUi
}

type DownloadInfo struct {
	s3Obj *s3.Object
	url   string
}

func (c *DownloadCommand) Help() string {
	helpText := `Usage: pill download {prefix}`
	return strings.TrimSpace(helpText)
}

func (c *DownloadCommand) Run(args []string) int {
	config := &aws.Config{
		Region: aws.String("us-east-1"),
	}
	s3client := s3.New(session.New(), config)

	lbi := &s3.ListObjectsInput{
		Bucket: aws.String("hello-jabil"),
		Prefix: aws.String(args[0]),
	}

	resp, err := s3client.ListObjects(lbi)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("%v", err))
		return 1
	}
	total := len(resp.Contents)
	numWorker := 10

	dlIn := make(chan DownloadInfo, 2)
	dlProgress := make(chan int, 2)
	dlOut := make(chan error, 1)
	dlDone := make(chan bool, 1)

	for i := 0; i < numWorker; i++ {
		go dl(i, dlIn, dlOut, dlDone, dlProgress)
	}

	go displayErr(c.Ui, dlOut)
	go displayProgress(c.Ui, dlProgress, total)

	c.Ui.Info(fmt.Sprintf("About to download %d files", total))

	for _, x := range resp.Contents {
		url := fmt.Sprintf("https://s3.amazonaws.com/hello-jabil/%s", *x.Key)
		dlIn <- DownloadInfo{s3Obj: x, url: url}
		time.Sleep(10 * time.Millisecond)
	}

	close(dlIn)

	for _ = range dlDone {
		numWorker -= 1
		if numWorker <= 0 {
			break
		}
	}

	close(dlOut)
	close(dlDone)
	time.Sleep(1 * time.Second)
	c.Ui.Info("All done")
	return 0
}

func (c *DownloadCommand) Synopsis() string {
	return "downloads all zip files starting with given prefix from s3"
}

func displayErr(ui cli.ColoredUi, in chan error) {
	for message := range in {
		ui.Error(fmt.Sprintf("%v", message))
	}
}

func displayProgress(ui cli.ColoredUi, in chan int, total int) {
	counter := 0
	start := int32(0)
	for _ = range in {
		counter += 1
		progress := int32(float64(counter) / float64(total) * 100)
		if progress != start && progress > 0 && progress%5 == 0 {
			ui.Info(fmt.Sprintf("%d%% (%d/%d)", progress, counter, total))
		}
		start = progress

	}
}

func dl(id int, in chan DownloadInfo, out chan error, done chan bool, progress chan int) {
	g := getter.S3Getter{}

	count := 0
	for message := range in {
		u, _ := url.Parse(message.url)
		err := g.GetFile(*message.s3Obj.Key, u)

		if err != nil {
			out <- err
		}
		count += 1
		progress <- 1
	}

	done <- true
}
