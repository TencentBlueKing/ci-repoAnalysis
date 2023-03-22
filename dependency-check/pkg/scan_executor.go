package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
	"os/exec"
)

// DependencyCheckExecutor DependencyCheck分析器
type DependencyCheckExecutor struct{}

// Execute 执行分析
func (e DependencyCheckExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	reportFile, err := doExecute(file.Name())
	if err != nil {
		return nil, err
	}
	return transform(reportFile)
}

// doExecute 执行扫描，扫描成功后返回报告路径
func doExecute(inputFile string) (string, error) {
	// dependency-check.sh --scan /src --format JSON --out /report

	const reportFile = "/report"
	args := []string{
		"--scan", inputFile,
		"--format", "JSON",
		"--out", reportFile,
	}

	cmd := exec.Command(CMDDependencyCheck, args...)
	util.Info(cmd.String())

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", errors.New("error: " + err.Error() + "\n" + stderr.String())
	}
	util.Info(out.String())
	util.Info(stderr.String())
	return reportFile + "/dependency-check-report.json", nil
}

// transform 转换DependencyCheck输出的报告为制品库扫描结果格式
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
