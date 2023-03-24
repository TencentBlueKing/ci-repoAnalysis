package pkg

import (
	"encoding/json"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
)

// DependencyCheckExecutor DependencyCheck分析器
type DependencyCheckExecutor struct{}

// Execute 执行分析
func (e DependencyCheckExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	offline, err := config.GetBoolArg(ConfigOffline)
	if err != nil {
		return nil, err
	}

	// 下载漏洞库
	dbUrl := config.GetStringArg(ConfigDbUrl)
	if len(dbUrl) > 0 {
		if err := util.ExtractTarUrl(dbUrl, DirDependencyCheckData, 0770); err != nil {
			return nil, err
		}
	}

	// 执行扫描
	reportFile, err := doExecute(file.Name(), offline)
	if err != nil {
		return nil, err
	}
	return transform(reportFile)
}

// doExecute 执行扫描，扫描成功后返回报告路径
func doExecute(inputFile string, offline bool) (string, error) {
	// dependency-check.sh --scan /src --format JSON --out /report

	const reportFile = "/report"
	args := []string{
		"--scan", inputFile,
		"--format", "JSON",
		"--out", reportFile,
	}

	if offline {
		args = append(
			args, "--noupdate",
			"--disableYarnAudit", "--disablePnpmAudit", "--disableNodeAudit", "--disableOssIndex", "--disableCentral")
	}

	if err := util.ExecAndLog(CMDDependencyCheck, args); err != nil {
		return "", err
	}

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
