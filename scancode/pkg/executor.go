package pkg

import (
	"encoding/json"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
	"path/filepath"
)

// Scancode 扫描器
type Scancode struct{}

// Execute 执行扫描
func (e Scancode) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	//extractcode /path/to/inputFile
	if err := util.ExecAndLog("extractcode", []string{file.Name()}, ""); err != nil {
		return nil, err
	}
	util.Info("success extract file %s", file.Name())

	//scancode --license-score 100 --license --max-depth 4 -n 4 --only-findings --json resultFile inputFile-extract
	resultFile := "/bkrepo/workspace/result.json"
	args := []string{
		"--license-score", "100",
		"--license",
		"--max-depth", "4",
		"-n", "4",
		"--only-findings",
		"--json", resultFile,
		filepath.Base(file.Name()) + "-extract",
	}

	if err := util.ExecAndLog("scancode", args, ""); err != nil {
		return nil, err
	}

	return transform(resultFile)
}

// transform 转换输出报告为标准格式
func transform(reportFile string) (*object.ToolOutput, error) {
	reportContent, err := os.ReadFile(reportFile)
	if err != nil {
		return nil, err
	}

	report := new(Report)
	if err := json.Unmarshal(reportContent, report); err != nil {
		return nil, err
	}

	return object.NewOutput(object.StatusSuccess, Convert(report)), nil
}
