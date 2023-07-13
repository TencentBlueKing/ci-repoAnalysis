package framework

import (
	"errors"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/api"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
)

// Executor 分析执行器
type Executor interface {
	// Execute 框架会调用该函数执行扫描，传入的参数config为工具相关配置，file为待分析的制品
	// 扫描成功时返回toolOutput，出错时返回error，工具框架会自动上报或输出结果给制品分析服务
	Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error)
}

// Analyze 执行分析
func Analyze(executor Executor) {
	args := object.GetArgs()
	for {
		doAnalyze(executor, args)
		if args.ShouldKeepRunning() {
			if err := util.CleanWorkDir(); err != nil {
				panic("clean work dir failed: " + err.Error())
			}
		} else {
			break
		}
	}
}

func doAnalyze(executor Executor, arguments *object.Arguments) {
	client := api.GetClient(arguments)
	input, err := client.Start()
	if err != nil {
		panic("Start analyze failed: " + err.Error())
	}
	if input == nil || input.TaskId == "" {
		util.Info("no subtask found, exit")
		os.Exit(0)
	}
	file, err := client.GenerateInputFile()
	defer file.Close()
	if err != nil {
		client.Failed(errors.New("Generate input file failed: " + err.Error()))
		os.Exit(1)
	}
	util.Info("generate input file success")
	output, err := executor.Execute(&input.ToolConfig, file)
	if err != nil {
		client.Failed(errors.New("Execute analysis failed: " + err.Error()))
	} else {
		client.Finish(output)
	}
	util.Info("analyze finished")
}
