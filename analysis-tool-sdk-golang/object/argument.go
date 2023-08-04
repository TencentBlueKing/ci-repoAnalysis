package object

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Arguments 输入参数
type Arguments struct {
	Url              string
	Token            string
	TaskId           string
	ExecutionCluster string
	PullRetry        int
	InputFilePath    string
	OutputFilePath   string
	KeepRunning      bool
	downloaderClient *http.Client
}

var args *Arguments

func newArguments() *Arguments {
	args = new(Arguments)
	flag.StringVar(&args.Url, "url", "", "制品库地址")
	flag.StringVar(&args.Token, "token", "", "制品库临时令牌")
	flag.StringVar(&args.TaskId, "task-id", "", "扫描任务Id")
	flag.StringVar(&args.ExecutionCluster, "execution-cluster", "", "所在扫描执行集群名")
	flag.IntVar(&args.PullRetry, "pull-retry", -1, "拉取模式下拉取任务的次数，-1表示一直拉取直到拉取到任务")
	flag.BoolVar(&args.KeepRunning, "keep-running", true, "是否一直运行，仅在拉取任务模式下生效")
	flag.StringVar(&args.InputFilePath, "input", "", "输入文件路径")
	flag.StringVar(&args.OutputFilePath, "output", "", "输出文件路径")
	flag.Parse()

	fmt.Printf(
		"url: %s, token: %s, taskId: %s, executionCluster: %s, pull-retry: %d, keep-running: %t, "+
			"inputFilePath: %s, outputFilePath: %s\n",
		args.Url,
		args.Token,
		args.TaskId,
		args.ExecutionCluster,
		args.PullRetry,
		args.KeepRunning,
		args.InputFilePath,
		args.OutputFilePath,
	)
	if (args.Offline() || args.Online()) == false {
		panic("缺少必要输入参数")
	}

	return args
}

// GetArgs 获取输入参数
func GetArgs() *Arguments {
	if args == nil {
		return newArguments()
	} else {
		return args
	}
}

// Offline 离线扫描
func (arg *Arguments) Offline() bool {
	return arg.InputFilePath != "" && arg.OutputFilePath != ""
}

// Online 在线扫描
func (arg *Arguments) Online() bool {
	return arg.Url != "" && arg.Token != "" && (arg.TaskId != "" || arg.ExecutionCluster != "")
}

// ShouldKeepRunning 是否无限循环拉取任务执行
func (arg *Arguments) ShouldKeepRunning() bool {
	return arg.Online() && arg.KeepRunning && arg.TaskId == ""
}

// CustomDownloaderHttpClientDialContext 自定义下载器使用的http客户端DNS解析
func (arg *Arguments) CustomDownloaderHttpClientDialContext(
	dialContext func(ctx context.Context, network, addr string) (net.Conn, error),
) *http.Client {
	if dialContext == nil {
		arg.downloaderClient = http.DefaultClient
	} else {
		transport := &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}

		arg.downloaderClient = &http.Client{
			Transport: transport,
		}
	}
	return arg.downloaderClient
}

// GetDownloaderClient 获取下载器HTTP客户端
func (arg *Arguments) GetDownloaderClient() *http.Client {
	if arg.downloaderClient == nil {
		return http.DefaultClient
	} else {
		return arg.downloaderClient
	}
}
