package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// analystTemporaryPrefix 制品分析服务接口前缀
const analystTemporaryPrefix = "/api/analyst/api/temporary"

var client *BkRepoClient

// BkRepoClient 为分析任务的输入输出操作提供同一入口
type BkRepoClient struct {
	Args      *object.Arguments
	ToolInput *object.ToolInput
}

// Response 制品分析服务响应
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// StandardScanExecutorResult 分析结果
type StandardScanExecutorResult struct {
	Type       string             `json:"type"`
	ScanStatus object.TaskStatus  `json:"scanStatus"`
	Output     *object.ToolOutput `json:"output"`
}

// ReportResultRequest 分析结果上报请求
type ReportResultRequest struct {
	SubTaskId          string                      `json:"subTaskId"`
	ScanStatus         object.TaskStatus           `json:"scanStatus"`
	ScanExecutorResult *StandardScanExecutorResult `json:"scanExecutorResult"`
	Token              string                      `json:"token"`
}

// GetClient 获取BkRepoClient
func GetClient(args *object.Arguments) *BkRepoClient {
	if client == nil {
		client = &BkRepoClient{args, nil}
	}
	return client
}

// Start 开始分析
func (c *BkRepoClient) Start() (*object.ToolInput, error) {
	if c.ToolInput == nil {
		if err := c.initToolInput(); err != nil {
			return nil, err
		}
		toolInputJson, _ := json.Marshal(c.ToolInput)
		util.Info("init tool input success:\n %s", toolInputJson)

		// 是在线任务时，更新任务状态为执行中
		if c.Args.Online() {
			if err := c.updateSubtaskStatus(); err != nil {
				return nil, err
			}
			util.Info("update subtask status success")
		}
	}
	return c.ToolInput, nil
}

// Finish 分析结束
func (c *BkRepoClient) Finish(toolOutput *object.ToolOutput) {
	toolOutput.TaskId = client.ToolInput.TaskId
	if c.Args.Offline() {
		if err := util.WriteToFile(c.Args.OutputFilePath, toolOutput); err != nil {
			panic("Finish analyze failed: " + err.Error())
		}
	} else {
		url := c.Args.Url + analystTemporaryPrefix + "/scan/report"
		result := StandardScanExecutorResult{"standard", toolOutput.Status, toolOutput}
		reqBody, err := json.Marshal(
			ReportResultRequest{
				SubTaskId:          client.ToolInput.TaskId,
				ScanStatus:         toolOutput.Status,
				ScanExecutorResult: &result,
				Token:              client.Args.Token,
			},
		)
		if err != nil {
			panic("Finish analyze failed: " + err.Error())
		}
		req, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
		if err != nil {
			panic("Finish analyze failed: " + err.Error())
		}
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic("Report analysis result failed, taskId: " + toolOutput.TaskId + ", err: " + err.Error())
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			panic("Report analysis result failed, taskId: " + toolOutput.TaskId)
		}
	}
}

func (c *BkRepoClient) Failed(err error) {
	util.Error("analyze failed %s", err)
	output := object.NewFailedOutput(err)
	c.Finish(output)
}

// GenerateInputFile 生成待分析文件
func (c *BkRepoClient) GenerateInputFile() (*os.File, error) {
	return util.GenerateInputFile(c.ToolInput, &util.DefaultDownloader{})
}

// updateSubtaskStatus 更新任务状态为执行中
func (c *BkRepoClient) updateSubtaskStatus() error {
	url := c.Args.Url + analystTemporaryPrefix + "/scan/subtask/" + c.Args.TaskId + "/status?token=" + c.Args.Token + "&status=EXECUTING"
	request, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("更新扫描任务[" + c.Args.TaskId + "]状态失败, status: " + response.Status)
	}

	res := new(Response[bool])
	if err := json.NewDecoder(response.Body).Decode(res); err != nil {
		return err
	}

	if !res.Data {
		return errors.New("更新扫描任务[" + c.Args.TaskId + "]状态失败, msg: " +
			res.Message + "code: " + strconv.Itoa(res.Code))
	}
	return nil
}

// initToolInput 从本地加载input.json或从服务端拉取toolInput信息
func (c *BkRepoClient) initToolInput() error {
	if c.Args.Offline() {
		fileContent, err := os.ReadFile(c.Args.InputFilePath)
		if err != nil {
			return err
		}
		toolInput := new(object.ToolInput)
		if err := json.Unmarshal(fileContent, toolInput); err != nil {
			return err
		}
		c.ToolInput = toolInput
	} else if c.Args.TaskId != "" {
		var err error
		if c.ToolInput, err = c.fetchToolInput(); err != nil {
			return err
		}
	} else {
		var err error
		if c.ToolInput, err = c.pullToolInput(); err != nil {
			return err
		}
		c.Args.TaskId = c.ToolInput.TaskId
	}
	return nil
}

// fetchToolInput 从制品分析服务拉取工具输入
func (c *BkRepoClient) fetchToolInput() (*object.ToolInput, error) {
	url := c.Args.Url + analystTemporaryPrefix + "/scan/subtask/" + c.Args.TaskId + "/input?token=" + c.Args.Token
	return c.doFetchToolInput(url)
}

// pullTooInput 从制品分析服务拉取工具输入
func (c *BkRepoClient) pullToolInput() (*object.ToolInput, error) {
	url := c.Args.Url + analystTemporaryPrefix + "/scan/subtask/input?executionCluster=" + c.Args.ExecutionCluster +
		"&token=" + c.Args.Token

	var toolInput *object.ToolInput = nil
	var err error = nil
	var pullRetry = c.Args.PullRetry
	for (toolInput == nil || toolInput.TaskId == "") && pullRetry != 0 {
		if err != nil {
			return nil, err
		}
		util.Info("try to pull subtask...")
		toolInput, err = c.doFetchToolInput(url)
		pullRetry--
		if toolInput == nil || toolInput.TaskId == "" {
			time.Sleep(5 * time.Second)
		}
	}

	return toolInput, err
}

func (c *BkRepoClient) doFetchToolInput(url string) (*object.ToolInput, error) {
	response, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(response.Body)
		errMsg := fmt.Sprintf(
			"get tool input failed, status: %d, error body: %s", http.StatusOK, string(errBody),
		)
		return nil, errors.New(errMsg)
	}

	res := new(Response[object.ToolInput])
	if err := json.NewDecoder(response.Body).Decode(res); err != nil {
		return nil, err
	}
	return &res.Data, nil
}
