package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/util"
	"github.com/TencentBlueKing/ci-repoAnalysis/trivy/constant"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type TrivyExecutor struct{}

func (e TrivyExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	if err := downloadDB(config); err != nil {
		return nil, err
	}
	maxTime, err := config.GetIntArg("maxTime")
	if err != nil {
		return nil, err
	}
	scanSensitive, _ := config.GetBoolArg("scanSensitive")

	if err := execTrivy(file.Name(), maxTime, scanSensitive); err != nil {
		return nil, err
	}
	return transformOutputJson()
}

func downloadDB(config *object.ToolConfig) error {
	url := config.GetStringArg("dbDownloadUrl")
	response, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("download failed")
	}
	file, err := os.Create(filepath.Join(constant.DbCacheDir, constant.DbFilename))
	if err != nil {
		return err
	}
	defer file.Close()
	written, err := io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	util.Info("download db, size is " + strconv.FormatInt(written, 10))
	return nil
}

func execTrivy(fileName string, maxTime int64, scanSensitive bool) error {
	// trivy --cache-dir /root/.cache/trivy image --input filePath -f json
	//       -o /bkrepo/workspace/trivy-output.json --skip-db-update --offline-scan

	args := []string{
		constant.FlagCacheDir, constant.CacheDir,
		constant.SubCmdImage,
		constant.FlagInput, fileName,
		constant.FlagFormat, constant.FormatJson,
		constant.FlagOutput, constant.OutputPath,
		constant.FlagTimeout, fmt.Sprintf("%dms", maxTime),
		constant.FlagSkipDbUpdate,
		constant.FlagOfflineScan,
	}

	if !scanSensitive {
		args = append(args, constant.FlagSecurityChecks, constant.CheckVuln)
	}

	cmd := exec.Command(constant.CmdTrivy, args...)
	util.Info(cmd.String())

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.New("error: " + err.Error() + "\n" + stderr.String())
	}
	util.Info(out.String())
	util.Info(stderr.String())
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
