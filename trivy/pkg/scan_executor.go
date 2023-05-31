package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"github.com/TencentBlueKing/ci-repoAnalysis/trivy/pkg/constant"
	"os"
	"path/filepath"
	"strings"
)

// TrivyExecutor Trivy分析工具执行器实现
type TrivyExecutor struct{}

// Execute 执行分析
func (e TrivyExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	offline, err := config.GetBoolArg(constant.ConfigOffline)
	if err != nil {
		offline = len(config.GetStringArg(constant.ArgDbDownloadUrl)) > 0
	}
	if offline {
		if err := downloadAllDB(config); err != nil {
			return nil, err
		}
	}

	if err := execTrivy(file.Name(), offline, config); err != nil {
		return nil, err
	}
	return transformOutputJson()
}

func downloadAllDB(config *object.ToolConfig) error {
	downloader := &util.DefaultDownloader{}
	// download db
	url := config.GetStringArg(constant.ArgDbDownloadUrl)
	if len(url) > 0 {
		dbDir := filepath.Join(constant.DbCacheDir, constant.DbDir)
		if err := util.ExtractTarUrl(url, dbDir, 0770, downloader); err != nil {
			return err
		}
	}

	// download java db
	javaDbUrl := config.GetStringArg(constant.ArgJavaDbDownloadUrl)
	if len(javaDbUrl) > 0 {
		javaDbDir := filepath.Join(constant.DbCacheDir, constant.JavaDbDir)
		if err := util.ExtractTarUrl(javaDbUrl, javaDbDir, 0770, downloader); err != nil {
			return err
		}
	}
	return nil
}

func execTrivy(fileName string, offline bool, config *object.ToolConfig) error {
	// trivy --cache-dir /root/.cache/trivy image --input filePath -f json
	//       -o /bkrepo/workspace/trivy-output.json --skip-db-update --offline-scan

	maxTime, err := config.GetIntArg("maxTime")
	if err != nil {
		return err
	}
	scanSensitive, _ := config.GetBoolArg("scanSensitive")
	scanLicense, _ := config.GetBoolArg("scanLicense")
	licenseFull, _ := config.GetBoolArg("licenseFull")

	args := []string{
		constant.FlagCacheDir, constant.CacheDir,
		constant.SubCmdImage,
		constant.FlagInput, fileName,
		constant.FlagFormat, constant.FormatJson,
		constant.FlagOutput, constant.OutputPath,
		constant.FlagTimeout, fmt.Sprintf("%dms", maxTime),
	}

	if offline {
		args = append(args, constant.FlagSkipDbUpdate, constant.FlagSkipJavaDbUpdate, constant.FlagOfflineScan)
	}

	// 设置要使用的扫描器
	scanners := []string{constant.ScannerVuln}
	if scanSensitive {
		scanners = append(scanners, constant.ScannerSecret)
	}
	if scanLicense {
		scanners = append(scanners, constant.ScannerLicense)
	}
	args = append(args, constant.FlagScanners, strings.Join(scanners, ","))

	if scanLicense && licenseFull {
		args = append(args, constant.FlagLicenseFull)
	}

	// 指定敏感信息规则
	if _, err := os.Stat(constant.SecretRuleFilePath); !errors.Is(err, os.ErrNotExist) {
		args = append(args, constant.FlagSecretConfig, constant.SecretRuleFilePath)
	}

	if err := util.ExecAndLog(constant.CmdTrivy, args); err != nil {
		return err
	}
	return nil
}

func transformOutputJson() (*object.ToolOutput, error) {
	trivyOutputContent, err := os.ReadFile(constant.OutputPath)
	if err != nil {
		return nil, err
	}

	trivyOutput := new(Output)
	if err := json.Unmarshal(trivyOutputContent, trivyOutput); err != nil {
		return nil, err
	}

	return object.NewOutput(object.StatusSuccess, ConvertToToolResults(trivyOutput)), nil
}
