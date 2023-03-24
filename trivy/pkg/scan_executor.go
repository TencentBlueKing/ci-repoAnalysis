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
)

// TrivyExecutor Trivy分析工具执行器实现
type TrivyExecutor struct{}

// Execute 执行分析
func (e TrivyExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	offline := len(config.GetStringArg(constant.ArgDbDownloadUrl)) > 0
	if offline {
		if err := downloadAllDB(config); err != nil {
			return nil, err
		}
	}
	maxTime, err := config.GetIntArg("maxTime")
	if err != nil {
		return nil, err
	}
	scanSensitive, _ := config.GetBoolArg("scanSensitive")

	if err := execTrivy(file.Name(), maxTime, scanSensitive, offline); err != nil {
		return nil, err
	}
	return transformOutputJson()
}

func downloadAllDB(config *object.ToolConfig) error {
	// download db
	url := config.GetStringArg(constant.ArgDbDownloadUrl)
	dbDir := filepath.Join(constant.DbCacheDir, constant.DbDir)
	if err := util.ExtractTarUrl(url, dbDir, 0770); err != nil {
		return err
	}

	// download java db
	javaDbUrl := config.GetStringArg(constant.ArgJavaDbDownloadUrl)
	javaDbDir := filepath.Join(constant.DbCacheDir, constant.JavaDbDir)
	if err := util.ExtractTarUrl(javaDbUrl, javaDbDir, 0770); err != nil {
		return err
	}
	return nil
}

func execTrivy(fileName string, maxTime int64, scanSensitive bool, offline bool) error {
	// trivy --cache-dir /root/.cache/trivy image --input filePath -f json
	//       -o /bkrepo/workspace/trivy-output.json --skip-db-update --offline-scan

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

	if !scanSensitive {
		args = append(args, constant.FlagSecurityChecks, constant.CheckVuln)
	}

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
