package api

import (
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"os"
	"testing"
)

func TestPullTask(t *testing.T) {
	client := BkRepoClient{}
	client.Args = &object.Arguments{
		Url:              os.Getenv("URL"),
		Token:            os.Getenv("TOKEN"),
		ExecutionCluster: os.Getenv("EXECUTION_CLUSTER"),
		PullRetry:        -1,
	}
	toolInput, err := client.pullToolInput()
	if err != nil {
		fmt.Println(err.Error())
	}

	if toolInput != nil {
		fmt.Println(toolInput.TaskId)
	}
}
