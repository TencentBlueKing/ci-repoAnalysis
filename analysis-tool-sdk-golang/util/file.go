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
	"os"
	"path/filepath"
	"strings"
)

const ArgKeyPkgType = "packageType"
const PackageTypeDocker = "DOCKER"
const WorkDir = "/bkrepo/workspace"
const manifestPath = "manifest.json"

// CleanWorkDir 清理工作空间
func CleanWorkDir() error {
	return os.RemoveAll(WorkDir)
}

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
func GenerateInputFile(toolInput *object.ToolInput, downloader Downloader) (*os.File, error) {
	if toolInput.FilePath != "" {
		return os.Open(toolInput.FilePath)
	}
	if err := os.MkdirAll(WorkDir, 0766); err != nil {
		return nil, err
	}

	if toolInput.ToolConfig.GetStringArg(ArgKeyPkgType) == PackageTypeDocker {
		return generateImageTar(toolInput, downloader)
	} else {
		fileUrl := toolInput.FileUrls[0]
		file, err := os.Create(filepath.Join(WorkDir, fileUrl.Name))
		if err != nil {
			return nil, err
		}
		reader, err := downloader.Download(fileUrl.Url)
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		if _, err := writeAndCheckSha256(reader, file, fileUrl.Sha256); err != nil {
			return nil, err
		}

		return file, nil
	}
}

// ExtractTarUrl 从指定url解压到指定路径
func ExtractTarUrl(url string, dstDir string, perm fs.FileMode, downloader Downloader) error {
	Info("extracting url %s to %s", url, dstDir)
	reader, err := downloader.Download(url)
	if err != nil {
		return err
	}
	defer reader.Close()
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

func generateImageTar(toolInput *object.ToolInput, downloader Downloader) (*os.File, error) {
	// 获取manifest
	manifest, err := loadManifest(&toolInput.FileUrls[0], downloader)
	if err != nil {
		return nil, err
	}

	// 构建镜像tar包
	imageFile, err := os.Create(filepath.Join(WorkDir, "image.tar"))
	if err != nil {
		return nil, err
	}
	tarWriter := tar.NewWriter(imageFile)
	defer tarWriter.Close()

	// 将config写入tar中
	fileUrlMap := toolInput.FileUrlMap()
	configFileUrl := fileUrlMap[manifest.Config.Sha256()]
	configFilePath := manifest.Config.Sha256() + ".json"

	err = loadFromUrlToTar(configFilePath, &configFileUrl, tarWriter, false, "", downloader)
	if err != nil {
		return nil, err
	}

	// 将layer写入tar中
	layers, err := loadLayersToTar(manifest, fileUrlMap, tarWriter, downloader)
	if err != nil {
		return nil, err
	}

	// 写入manifest到tar
	if err := writeManifestToTar(configFilePath, layers, tarWriter); err != nil {
		return nil, err
	}

	return imageFile, nil
}

func loadManifest(manifestUrl *object.FileUrl, downloader Downloader) (*object.ManifestV2, error) {
	manifestResponse, err := downloader.Download(manifestUrl.Url)
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
	return manifest, nil
}

func loadLayersToTar(
	manifest *object.ManifestV2,
	fileUrlMap map[string]object.FileUrl,
	tarWriter *tar.Writer,
	downloader Downloader,
) ([]string, error) {
	cacheDir := filepath.Join(WorkDir, "layer-cache")
	if err := os.MkdirAll(cacheDir, 0766); err != nil {
		return nil, err
	}
	layerCount := manifest.LayerCount()
	layers := make([]string, 0, len(manifest.Layers))
	for _, layer := range manifest.Layers {
		s := layer.Sha256()
		url := fileUrlMap[s]
		if err := writeTarHeader(url.Sha256+"/", 0, tarWriter); err != nil {
			return nil, err
		}

		layerPath := url.Sha256 + "/layer.tar"
		layers = append(layers, layerPath)
		dup := layerCount[s] > 1
		if err := loadFromUrlToTar(layerPath, &url, tarWriter, dup, cacheDir, downloader); err != nil {
			return nil, err
		}
	}
	if err := os.RemoveAll(cacheDir); err != nil {
		return nil, err
	}
	return layers, nil
}

// loadFromUrlToTar 从url中加载数据并写入tar中
// 如果dup为true表示为重复使用的url，会先尝试从本地缓存加载数据，加载不到会从url下载并写入缓存
func loadFromUrlToTar(
	name string,
	fileUrl *object.FileUrl,
	tarWriter *tar.Writer,
	dup bool,
	cacheDir string,
	downloader Downloader,
) error {
	cached := false
	cacheFile := filepath.Join(cacheDir, fileUrl.Sha256)
	if dup {
		_, err := os.Stat(cacheFile)
		cached = err == nil
	}

	// 获取数据来源
	var src io.ReadCloser
	if cached {
		f, err := os.Open(cacheFile)
		if err != nil {
			return err
		}
		defer f.Close()
		src = f
	} else {
		layerRes, err := downloader.Download(fileUrl.Url)
		if err != nil {
			return err
		}
		defer layerRes.Close()
		src = layerRes
	}

	// 判断是否需要同时写缓存
	var dst io.Writer
	if !dup || dup && cached {
		dst = tarWriter
	} else {
		f, err := os.Create(cacheFile)
		if err != nil {
			return err
		}
		defer f.Close()
		dst = io.MultiWriter(tarWriter, f)
	}

	// 数据写入tar
	header := &tar.Header{
		Name: name,
		Size: fileUrl.Size,
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err := writeAndCheckSha256(src, dst, fileUrl.Sha256)
	if err != nil {
		return err
	}

	return nil
}

func writeManifestToTar(configFilePath string, layers []string, tarWriter *tar.Writer) error {
	manifestV1 := []object.ManifestV1{
		{
			Config:   configFilePath,
			RepoTags: []string{},
			Layers:   layers,
		},
	}
	manifestV1Json, err := json.Marshal(manifestV1)
	if err != nil {
		return err
	}

	if err := writeTarHeader(manifestPath, int64(len(manifestV1Json)), tarWriter); err != nil {
		return err
	}
	if _, err := io.Copy(tarWriter, bytes.NewReader(manifestV1Json)); err != nil {
		return err
	}

	return nil
}

func writeTarHeader(name string, size int64, tarWriter *tar.Writer) error {
	header := &tar.Header{
		Name: name,
		Size: size,
	}
	return tarWriter.WriteHeader(header)
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
