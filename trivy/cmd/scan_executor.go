package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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
)

// TrivyExecutor Trivy分析工具执行器实现
type TrivyExecutor struct{}

// Execute 执行分析
func (e TrivyExecutor) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	if err := downloadAllDB(config); err != nil {
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

func downloadAllDB(config *object.ToolConfig) error {
	url := config.GetStringArg("dbDownloadUrl")
	if err := downloadDB(url, constant.DbDir); err != nil {
		return err
	}
	javaDbUrl := config.GetStringArg("javaDbDownloadUrl")
	if err := downloadDB(javaDbUrl, constant.JavaDbDir); err != nil {
		return err
	}
	return nil
}

func downloadDB(url string, dbDir string) error {
	response, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New(fmt.Sprintf("download db to %s failed, status %s", dbDir, response.Status))
	}
	p := filepath.Join(constant.DbCacheDir, dbDir, "temp.tar")
	if err := os.MkdirAll(filepath.Dir(p), 0666); err != nil {
		return err
	}
	file, err := os.Create(p)
	if err != nil {
		return err
	}
	written, err := io.Copy(file, response.Body)
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := extract(p); err != nil {
		return err
	}
	if err := os.Remove(p); err != nil {
		return err
	}
	util.Info("download db to %s success, size is %d", p, written)
	return nil
}

func extract(tarPath string) error {
	gzipStream, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer gzipStream.Close()
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()
	tarReader := tar.NewReader(uncompressedStream)

	dir := filepath.Dir(tarPath)
	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		p := filepath.Join(dir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(p, 0444); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(p)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		default:
			return errors.New(fmt.Sprintf("unknown tar type %d in %s", header.Typeflag, header.Name))
		}
	}

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
		constant.FlagSkipJavaDbUpdate,
		constant.FlagOfflineScan,
	}

	if !scanSensitive {
		args = append(args, constant.FlagSecurityChecks, constant.CheckVuln)
	}

	if _, err := os.Stat(constant.SecretRuleFilePath); !errors.Is(err, os.ErrNotExist) {
		args = append(args, constant.FlagSecretConfig, constant.SecretRuleFilePath)
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
