package api

import (
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
	"testing"
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
	_, err := client.createDownloader()
	if err != nil {
		t.Fatalf(err.Error())
	}
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
