package api

import (
	"context"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"net"
	"os"
	"testing"
	"time"
)

func TestPullTask(t *testing.T) {
	client := createClient()
	toolInput, err := client.pullToolInput()
	if err != nil {
		fmt.Println(err.Error())
	}

	if toolInput != nil {
		fmt.Println(toolInput.TaskId)
	}
}

func TestCreateDownloader(t *testing.T) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}
	downloaderClient := (&object.Arguments{}).CustomDownloaderHttpClientDialContext(dialContext)
	client := createClient()
	args := make([]object.Argument, 2)
	args = append(args, object.Argument{
		Type:  "STRING",
		Key:   util.ArgKeyDownloaderWorkerHeaders,
		Value: "X-Test-Header: test, X-Test-Header2: test2",
	})
	args = append(args, object.Argument{
		Type:  "NUMBER",
		Key:   util.ArgKeyDownloaderWorkerCount,
		Value: "2",
	})
	client.ToolInput = &object.ToolInput{
		ToolConfig: object.ToolConfig{Args: args},
	}
	downloader, err := client.createDownloader(downloaderClient)
	if err != nil {
		t.Fatalf(err.Error())
	}
	reader, err := downloader.Download(os.Getenv("ARTIFACT_URL"))
	if err != nil {
		fmt.Println(err.Error())
	} else {
		_ = reader.Close()
	}

	_ = os.RemoveAll(util.WorkDir)
}

func TestHeartbeat(t *testing.T) {
	client := createClient()
	client.Args.Heartbeat = 2
	client.ToolInput = &object.ToolInput{}
	client.ToolInput.TaskId = os.Getenv("TASK_ID")
	ctx, cancel := context.WithCancel(context.Background())
	go client.heartbeat(ctx)
	time.Sleep(10 * time.Second)
	cancel()
	time.Sleep(5 * time.Second)
}

func createClient() *BkRepoClient {
	client := BkRepoClient{}
	client.Args = &object.Arguments{
		Url:              os.Getenv("URL"),
		Token:            os.Getenv("TOKEN"),
		ExecutionCluster: os.Getenv("EXECUTION_CLUSTER"),
		PullRetry:        -1,
	}
	return &client
}
