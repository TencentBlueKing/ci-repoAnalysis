package object

import (
	"flag"
	"fmt"
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
	Heartbeat        int
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
	flag.IntVar(&args.Heartbeat, "heartbeat", 0, "任务心跳上报间隔，0表示不上报")
	flag.Parse()

	fmt.Printf(
		"url: %s, token: %s, taskId: %s, executionCluster: %s, pull-retry: %d, keep-running: %t, heartbeat: %d"+
			"inputFilePath: %s, outputFilePath: %s\n",
		args.Url,
		args.Token,
		args.TaskId,
		args.ExecutionCluster,
		args.PullRetry,
		args.KeepRunning,
		args.Heartbeat,
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
