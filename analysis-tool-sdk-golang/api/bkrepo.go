package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"github.com/hashicorp/go-retryablehttp"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
func (c *BkRepoClient) Start(ctx context.Context, cancel context.CancelFunc) (*object.ToolInput, error) {
	if c.ToolInput == nil {
		if err := c.initToolInput(); err != nil {
			return nil, err
		}
		util.Info("init tool input success: %s", c.ToolInput.TaskId)

		// 是在线任务时，更新任务状态为执行中
		if c.Args.Online() {
			if err := c.updateSubtaskStatus(); err != nil {
				return nil, err
			}
			if c.Args.Heartbeat > 0 {
				go c.heartbeat(ctx, cancel)
			}
			util.Info("update subtask status success")
		}
	}
	return c.ToolInput, nil
}

// Finish 分析结束
func (c *BkRepoClient) Finish(cancel context.CancelFunc, toolOutput *object.ToolOutput) {
	toolOutput.TaskId = c.ToolInput.TaskId
	cancel()
	if c.Args.Offline() {
		if err := util.WriteToFile(c.Args.OutputFilePath, toolOutput); err != nil {
			panic("Finish analyze failed: " + err.Error())
		}
	} else {
		reqUrl := c.Args.Url + analystTemporaryPrefix + "/scan/report"
		result := StandardScanExecutorResult{"standard", toolOutput.Status, toolOutput}
		reqBody, err := json.Marshal(
			ReportResultRequest{
				SubTaskId:          c.ToolInput.TaskId,
				ScanStatus:         toolOutput.Status,
				ScanExecutorResult: &result,
				Token:              c.Args.Token,
			},
		)
		if err != nil {
			panic("Finish analyze failed: " + err.Error())
		}
		req, err := retryablehttp.NewRequest("POST", reqUrl, bytes.NewReader(reqBody))
		if err != nil {
			panic("Finish analyze failed: " + err.Error())
		}
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
		res, err := util.DefaultClient.Do(req)
		if err != nil {
			panic("Report analysis result failed, taskId: " + toolOutput.TaskId + ", err: " + err.Error())
		}
		defer util.DrainBody(res.Body)
		if res.StatusCode != 200 {
			panic("Report analysis result failed, taskId: " + toolOutput.TaskId)
		}
	}
	c.ToolInput = nil
}

func (c *BkRepoClient) Failed(cancel context.CancelFunc, err error) {
	util.Error("analyze failed %s", err)
	output := object.NewFailedOutput(err)
	c.Finish(cancel, output)
}

// GenerateInputFile 生成待分析文件
func (c *BkRepoClient) GenerateInputFile() (*os.File, error) {
	downloader, err := c.createDownloader()
	if err != nil {
		return nil, err
	}
	return util.GenerateInputFile(c.ToolInput, downloader)
}

func (c *BkRepoClient) createDownloader() (util.Downloader, error) {
	var downloader util.Downloader
	workerCount, _ := c.ToolInput.ToolConfig.GetIntArg(util.ArgKeyDownloaderWorkerCount)
	if workerCount > 0 {
		// 解析header
		downloaderHeadersStr := c.ToolInput.ToolConfig.GetStringArg(util.ArgKeyDownloaderWorkerHeaders)
		headers := make(map[string]string)
		if len(downloaderHeadersStr) > 0 {
			downloaderHeaders := strings.Split(downloaderHeadersStr, ",")
			for i := range downloaderHeaders {
				h := strings.Split(downloaderHeaders[i], ":")
				if len(h) != 2 {
					return nil, errors.New("headers error: " + downloaderHeaders[i])
				}
				headers[strings.TrimSpace(h[0])] = strings.TrimSpace(h[1])
			}
		}
		// 创建下载器并生成待分析文件
		downloader = util.NewChunkDownloader(int(workerCount), util.WorkDir, headers)
	} else {
		downloader = util.NewDownloader()
	}
	return downloader, nil
}

// updateSubtaskStatus 更新任务状态为执行中
func (c *BkRepoClient) updateSubtaskStatus() error {
	reqUrl := c.Args.Url + analystTemporaryPrefix + "/scan/subtask/" + c.ToolInput.TaskId + "/status?token=" + c.Args.Token + "&status=EXECUTING"
	request, err := retryablehttp.NewRequest("PUT", reqUrl, nil)
	if err != nil {
		return err
	}
	response, err := util.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer util.DrainBody(response.Body)
	if response.StatusCode != 200 {
		return errors.New("更新扫描任务[" + c.ToolInput.TaskId + "]状态失败, status: " + response.Status)
	}

	res := new(Response[bool])
	if err := json.NewDecoder(response.Body).Decode(res); err != nil {
		return err
	}

	if !res.Data {
		return errors.New("更新扫描任务[" + c.ToolInput.TaskId + "]状态失败, msg: " +
			res.Message + "code: " + strconv.Itoa(res.Code))
	}
	return nil
}

func (c *BkRepoClient) heartbeat(ctx context.Context, cancel context.CancelFunc) {
	ticker := time.NewTicker(time.Duration(c.Args.Heartbeat) * time.Second)
	taskId := c.ToolInput.TaskId
	reqUrl := c.Args.Url + analystTemporaryPrefix + "/scan/subtask/" + taskId + "/heartbeat"
	data := url.Values{}
	data.Set("token", c.Args.Token)
	body := data.Encode()
	for {
		select {
		case <-ctx.Done():
			util.Info("stop heartbeat of task: " + taskId)
			ticker.Stop()
			return
		case <-ticker.C:
			request, err := retryablehttp.NewRequest(http.MethodPost, reqUrl, strings.NewReader(body))
			if err != nil {
				util.Error("heartbeat failed: " + err.Error())
			}
			request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			response, err := util.DefaultClient.Do(request)
			if err != nil {
				util.Error("heartbeat failed: " + err.Error())
				return
			}
			if response.StatusCode != http.StatusOK {
				cancel()
				b, _ := io.ReadAll(response.Body)
				util.Error("heartbeat failed: " + response.Status + ", message: " + string(b))
			}
			util.DrainBody(response.Body)
		}
	}
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
		if c.ToolInput, err = c.fetchToolInput(c.Args.TaskId); err != nil {
			return err
		}
	} else {
		var err error
		if c.ToolInput, err = c.pullToolInput(); err != nil {
			return err
		}
	}
	return nil
}

// fetchToolInput 从制品分析服务拉取工具输入
func (c *BkRepoClient) fetchToolInput(taskId string) (*object.ToolInput, error) {
	reqUrl := c.Args.Url + analystTemporaryPrefix + "/scan/subtask/" + taskId + "/input?token=" + c.Args.Token
	return c.doFetchToolInput(reqUrl)
}

// pullTooInput 从制品分析服务拉取工具输入
func (c *BkRepoClient) pullToolInput() (*object.ToolInput, error) {
	reqUrl := c.Args.Url + analystTemporaryPrefix + "/scan/subtask/input?executionCluster=" + c.Args.ExecutionCluster +
		"&token=" + c.Args.Token

	var toolInput *object.ToolInput = nil
	var err error = nil
	var pullRetry = c.Args.PullRetry
	for (toolInput == nil || toolInput.TaskId == "") && pullRetry != 0 {
		if err != nil {
			return nil, err
		}
		util.Info("try to pull subtask...")
		toolInput, err = c.doFetchToolInput(reqUrl)
		pullRetry--
		if toolInput == nil || toolInput.TaskId == "" {
			time.Sleep(5 * time.Second)
		}
	}

	return toolInput, err
}

func (c *BkRepoClient) doFetchToolInput(url string) (*object.ToolInput, error) {
	response, err := util.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer util.DrainBody(response.Body)
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
