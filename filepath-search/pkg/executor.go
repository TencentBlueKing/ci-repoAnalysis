package pkg

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

// FilepathSearch 文件路径匹配工具
type FilepathSearch struct{}

// Execute 执行扫描，在镜像tar包中搜索匹配指定正则表达式的路径
func (e FilepathSearch) Execute(config *object.ToolConfig, file *os.File) (*object.ToolOutput, error) {
	regStr := config.GetStringArg("regex")
	if len(regStr) == 0 {
		return nil, errors.New("regex config not found")
	}

	securityResults, err := scan(file.Name(), regStr)
	if err != nil {
		return nil, err
	}

	return object.NewOutput(
		object.StatusSuccess,
		&object.Result{
			SecurityResults: securityResults,
		},
	), nil
}

func scan(filepath string, regex string) ([]object.SecurityResult, error) {
	reg, _ := regexp.Compile(regex)
	img, err := tarball.Image(fileOpener(filepath), nil)
	if err != nil {
		return nil, err
	}

	layers, err := img.Layers()
	if err != nil {
		return nil, err
	}

	searchedLayers := make(map[string]struct{})
	securityResults := make([]object.SecurityResult, 0)

	for i := range layers {
		l := layers[i]
		diffId, err := l.DiffID()
		if err != nil {
			return nil, err
		}

		// 已经搜索过的layer直接跳过
		if _, ok := searchedLayers[diffId.Hex]; ok {
			continue
		}

		if err := walk(l, func(filePath string, info os.FileInfo, reader io.Reader) error {
			if reg.MatchString(filePath) {
				securityResults = append(securityResults, object.SecurityResult{
					VulId:    "filepath-match",
					VulName:  "filepath-match",
					Path:     filePath,
					PkgName:  filePath,
					Des:      fmt.Sprintf("File path [%s] matches the regex [%s]", filePath, reg),
					Severity: "CRITICAL",
				})
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return securityResults, nil
}

type WalkFunc func(filePath string, info os.FileInfo, reader io.Reader) error

func walk(layer v1.Layer, processFunc WalkFunc) error {
	rc, err := layer.Uncompressed()
	if err != nil {
		return err
	}
	defer rc.Close()
	tr := tar.NewReader(rc)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		filePath := strings.TrimLeft(path.Clean(hdr.Name), "/")

		if err := processFunc(filePath, hdr.FileInfo(), tr); err != nil {
			return err
		}
	}

	return nil
}

func fileOpener(file string) func() (io.ReadCloser, error) {
	return func() (io.ReadCloser, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		br := bufio.NewReader(f)
		var r io.Reader = br
		if isGzip(br) {
			var err error
			if r, err = gzip.NewReader(br); err != nil {
				return nil, err
			}
		}
		return io.NopCloser(r), nil
	}
}

func isGzip(f *bufio.Reader) bool {
	buf, err := f.Peek(3)
	if err != nil {
		return false
	}
	return buf[0] == 0x1F && buf[1] == 0x8B && buf[2] == 0x8
}
