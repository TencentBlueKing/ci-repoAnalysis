package util

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const ArgKeyPkgType = "packageType"
const PackageTypeDocker = "DOCKER"
const WorkDir = "/bkrepo/workspace"
const manifestPath = "manifest.json"

// WriteToFile 工具输出写入文件
func WriteToFile(outputFilePath string, toolOutput *object.ToolOutput) error {
	toolOutputContent, err := json.Marshal(toolOutput)
	if err != nil {
		return err
	}

	output, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = output.Write(toolOutputContent)
	return err
}

// GenerateInputFile 生成输入文件
func GenerateInputFile(toolInput *object.ToolInput) (*os.File, error) {
	if toolInput.FilePath != "" {
		return os.Open(toolInput.FilePath)
	}
	if err := os.MkdirAll(WorkDir, 0766); err != nil {
		return nil, err
	}

	if toolInput.ToolConfig.GetStringArg(ArgKeyPkgType) == PackageTypeDocker {
		return generateImageTar(toolInput)
	} else {
		fileUrl := toolInput.FileUrls[0]
		file, err := os.Create(filepath.Join(WorkDir, fileUrl.Name))
		if err != nil {
			return nil, err
		}
		reader, err := Download(fileUrl.Url)
		defer reader.Close()
		if err != nil {
			return nil, err
		}
		if _, err := writeAndCheckSha256(reader, file, fileUrl.Sha256); err != nil {
			return nil, err
		}

		return file, nil
	}
}

// ExtractTarUrl 从指定url解压到指定路径
func ExtractTarUrl(url string, dstDir string, perm fs.FileMode) error {
	Info("extracting url %s to %s", url, dstDir)
	reader, err := Download(url)
	defer reader.Close()
	if err != nil {
		return err
	}
	return Extract(reader, dstDir, perm)
}

// ExtractTarFile 解压文件到指定路径
func ExtractTarFile(tarPath string, dstDir string, perm fs.FileMode) error {
	Info("extracting file %s to %s", tarPath, dstDir)
	fileReader, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer fileReader.Close()
	return Extract(fileReader, dstDir, perm)
}

// Extract 解压tar.gz到指定路径
func Extract(reader io.Reader, dstDir string, perm fs.FileMode) error {
	if _, err := os.Stat(dstDir); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(dstDir, perm); err != nil {
			return err
		}
	}

	uncompressedStream, err := gzip.NewReader(reader)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()
	tarReader := tar.NewReader(uncompressedStream)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		p := filepath.Join(dstDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(p, perm); err != nil {
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

	Info("extract to %s success", dstDir)
	return nil
}

func generateImageTar(toolInput *object.ToolInput) (*os.File, error) {
	// 获取manifest
	manifestUrl := toolInput.FileUrls[0]
	manifestResponse, err := Download(manifestUrl.Url)
	if err != nil {
		return nil, err
	}
	manifest := new(object.ManifestV2)
	err = json.NewDecoder(manifestResponse).Decode(manifest)
	if err != nil {
		return nil, err
	}
	if err := manifestResponse.Close(); err != nil {
		return nil, err
	}
	Info("get image manifest success")

	// 构建镜像tar包
	fileUrlMap := make(map[string]object.FileUrl, len(toolInput.FileUrls))
	for _, url := range toolInput.FileUrls {
		fileUrlMap[url.Sha256] = url
	}
	imageFile, err := os.Create(filepath.Join(WorkDir, "image.tar"))
	if err != nil {
		return nil, err
	}
	tarWriter := tar.NewWriter(imageFile)
	defer tarWriter.Close()

	// load config layer
	configFilePath := manifest.Config.Sha256() + ".json"
	configFileUrl := fileUrlMap[manifest.Config.Sha256()]
	if err := loadLayerToTar(configFilePath, &configFileUrl, tarWriter); err != nil {
		return nil, err
	}
	layers := make([]string, 0, len(manifest.Layers))

	// load layer
	for _, layer := range manifest.Layers {
		url := fileUrlMap[layer.Sha256()]
		if err := putArchiveEntry(url.Sha256+"/", 0, nil, tarWriter); err != nil {
			return nil, err
		}
		layerPath := url.Sha256 + "/layer.tar"
		layers = append(layers, layerPath)
		if err := loadLayerToTar(layerPath, &url, tarWriter); err != nil {
			return nil, err
		}
	}

	// 打包manifest
	manifestV1 := []object.ManifestV1{
		{
			Config:   configFilePath,
			RepoTags: []string{},
			Layers:   layers,
		},
	}
	manifestV1Json, err := json.Marshal(manifestV1)
	if err != nil {
		return nil, err
	}
	err = putArchiveEntry(manifestPath, int64(len(manifestV1Json)), bytes.NewReader(manifestV1Json), tarWriter)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return imageFile, nil
}

func loadLayerToTar(name string, fileUrl *object.FileUrl, tarWriter *tar.Writer) error {
	layerRes, err := Download(fileUrl.Url)
	if err != nil {
		return err
	}
	defer layerRes.Close()

	header := &tar.Header{
		Name: name,
		Size: fileUrl.Size,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = writeAndCheckSha256(layerRes, tarWriter, fileUrl.Sha256)
	if err != nil {
		return err
	}

	return nil
}

func putArchiveEntry(name string, size int64, reader io.Reader, tarWriter *tar.Writer) error {
	header := &tar.Header{
		Name: name,
		Size: size,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if reader != nil {
		if _, err := io.Copy(tarWriter, reader); err != nil {
			return err
		}
	}
	return nil
}

// Download 从指定url获取输入流
func Download(url string) (io.ReadCloser, error) {
	Info("downloading %s", url)
	response, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("download failed, status: " + response.Status)
	}

	return response.Body, nil
}

func writeAndCheckSha256(reader io.Reader, writer io.Writer, realSha256 string) (string, error) {
	sha := sha256.New()
	written, err := io.Copy(io.MultiWriter(sha, writer), reader)
	if err != nil {
		return "", err
	}
	downloadedFileSha256 := strings.ToUpper(hex.EncodeToString(sha.Sum(nil)))
	if downloadedFileSha256 != strings.ToUpper(realSha256) {
		return "", errors.New("download failed, file broken " + downloadedFileSha256)
	}
	Info("download file success, size is %d", written)
	return downloadedFileSha256, nil
}
