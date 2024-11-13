package pkg

import (
	"context"
	"encoding/json"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/api"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
)

type CmdBuilder interface {
	// Build 生成扫描执行器启动命令, 返回（命令名，参数列表）
	Build(input *object.ToolInput) (string, []string, error)
}

// StandardAdapterExecutor 标准扫描器适配器
type StandardAdapterExecutor struct {
	CmdBuilder CmdBuilder
	WorkDir    string
}

// Execute 执行扫描
func (e StandardAdapterExecutor) Execute(
	ctx context.Context,
	_ *object.ToolConfig,
	file *os.File,
) (*object.ToolOutput, error) {
	// 将toolInput写入/bkrepo/workspace/input.json
	toolInput := api.GetClient(object.GetArgs()).ToolInput
	newToolInput := &object.ToolInput{
		TaskId:     toolInput.TaskId,
		ToolConfig: toolInput.ToolConfig,
		FilePath:   file.Name(),
		Sha256:     "",
		FileUrls:   nil,
	}
	toolInputFile := util.WorkDir + "/input.json"
	if err := e.writeToolInput(newToolInput, toolInputFile); err != nil {
		return nil, err
	}

	// 执行扫描
	toolOutputFile := util.WorkDir + "/output.json"
	cmd, args, err := e.CmdBuilder.Build(newToolInput)
	if err != nil {
		return nil, err
	}
	args = append(args, "--input", toolInputFile, "--output", toolOutputFile)
	err = util.ExecAndLog(ctx, cmd, args, e.WorkDir)
	if err != nil {
		return nil, err
	}

	// 从/bkrepo/workspace/output.json读取扫描结果
	return e.readToolOutput(toolOutputFile)
}

func (e StandardAdapterExecutor) writeToolInput(input *object.ToolInput, f string) error {
	file, err := os.Create(f)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(input)
	if err != nil {
		return err
	}
	util.Info("write input.json success")
	return nil
}

func (e StandardAdapterExecutor) readToolOutput(f string) (*object.ToolOutput, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	output := new(object.ToolOutput)
	err = json.NewDecoder(file).Decode(output)
	if err != nil {
		return nil, err
	}

	for i := range output.Result.SecurityResults {
		r := &output.Result.SecurityResults[i]
		if r.PkgVersions == nil {
			r.PkgVersions = []string{}
		}
		if r.References == nil {
			r.References = []string{}
		}
	}
	return output, nil
}
