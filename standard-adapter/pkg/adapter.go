package pkg

import (
	"encoding/json"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/api"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
	"strings"
)

// StandardAdapterExecutor 标准扫描器适配器
type StandardAdapterExecutor struct {
	Cmd     string
	WorkDir string
}

// Execute 执行扫描
func (e StandardAdapterExecutor) Execute(_ *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	// 将toolInput写入/bkrepo/workspace/input.json
	toolInput, err := api.GetClient(object.GetArgs()).Start()
	if err != nil {
		return nil, err
	}
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
	var args []string
	cmd := e.Cmd
	splitCmd := strings.Split(e.Cmd, " ")
	if len(splitCmd) > 1 {
		cmd = splitCmd[0]
		args = append(args, splitCmd[1:]...)
	}
	args = append(args, "--input", toolInputFile, "--output", toolOutputFile)
	err = util.ExecAndLog(cmd, args, e.WorkDir)
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
	return output, nil
}
