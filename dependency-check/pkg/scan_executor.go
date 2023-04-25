package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"os"
	"path/filepath"
	"strings"
)

const PackageTypeNpm = "NPM"

// DependencyCheckExecutor DependencyCheck分析器
type DependencyCheckExecutor struct{}

// Execute 执行分析
func (e DependencyCheckExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	offline, err := config.GetBoolArg(ConfigOffline)
	if err != nil {
		return nil, err
	}

	inputFile := file.Name()
	if config.GetStringArg(util.ArgKeyPkgType) == PackageTypeNpm {
		if err := npmPrepare(file); err != nil {
			return nil, err
		}
		inputFile = filepath.Join(filepath.Dir(inputFile), "package-lock.json")
	}

	// 下载漏洞库
	downloader := &util.DefaultDownloader{}
	dbUrl := config.GetStringArg(ConfigDbUrl)
	if len(dbUrl) > 0 {
		if err := util.ExtractTarUrl(dbUrl, DirDependencyCheckData, 0770, downloader); err != nil {
			return nil, err
		}
	}

	// 执行扫描
	reportFile, err := doExecute(inputFile, offline)
	if err != nil {
		return nil, err
	}
	return transform(reportFile)
}

func npmPrepare(file *os.File) error {
	fileAbsPath := file.Name()
	fileBaseName := filepath.Base(fileAbsPath)
	workDir := filepath.Dir(fileAbsPath)

	// npm install
	if err := util.ExecAndLog("npm", []string{"install", file.Name()}, workDir); err != nil {
		return err
	}

	// 获取 pkgName 和 pkgVersion
	indexOfHyphens := strings.Index(fileBaseName, "-")
	if indexOfHyphens == -1 {
		return errors.New("'-' not found in file name " + fileBaseName)
	}
	indexOfLastDot := strings.LastIndex(fileBaseName, ".")
	if indexOfLastDot == -1 {
		return errors.New("'.' not found in file name " + fileBaseName)
	}
	pkgName := fileBaseName[:indexOfHyphens]
	pkgVersion := fileBaseName[indexOfHyphens+1 : indexOfLastDot]
	util.Info("npm package %s, version %s", pkgName, pkgVersion)

	// 替换 package-lock.json中的file:xxx 为实际版本号
	sedExp := fmt.Sprintf(
		"s/\\\"%s\\\": \\\"file:%s\\\"/\\\"%s\\\": \\\"%s\\\"/",
		pkgName, fileBaseName, pkgName, pkgVersion,
	)
	if err := sed(sedExp, filepath.Join(workDir, "package-lock.json")); err != nil {
		return err
	}
	if err := sed(sedExp, filepath.Join(workDir, "package.json")); err != nil {
		return err
	}

	sedExp = fmt.Sprintf(
		"s/\\\"version\\\": \\\"file:%s\\\"/\\\"version\\\": \\\"%s\\\"/",
		fileBaseName, pkgVersion,
	)
	if err := sed(sedExp, filepath.Join(workDir, "package-lock.json")); err != nil {
		return err
	}
	return nil
}

func sed(exp string, fileAbsPath string) error {
	args := []string{"-i", exp, fileAbsPath}
	if err := util.ExecAndLog("sed", args, ""); err != nil {
		return err
	}
	return nil
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

	if err := util.ExecAndLog(CMDDependencyCheck, args, ""); err != nil {
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
