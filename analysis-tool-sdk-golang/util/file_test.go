package util

import (
	"bytes"
	"errors"
	"github.com/TencentBlueKing/ci-repoAnalysis/analysis-tool-sdk-golang/object"
	"io"
	"os"
	"strings"
	"testing"
)

func TestGenerateInputFile(t *testing.T) {
	input := &object.ToolInput{
		ToolConfig: object.ToolConfig{
			Args: []object.Argument{
				{
					Type:  "STRING",
					Key:   ArgKeyPkgType,
					Value: PackageTypeDocker,
				},
			},
		},
		FileUrls: []object.FileUrl{
			{
				Url: "manifest.json",
			},
			{
				Url:    "config",
				Sha256: "b79606fb3afea5bd1609ed40b622142f1c98125abcfe89a76a661b0e8e343910",
				Size:   int64(len("config")),
			},
			{
				Url:    "layer1",
				Sha256: "77ea7eee3d80b1a38f83906dd3048e2689457eb90e18a7d12f839c5ae37106a2",
				Size:   int64(len("layer1")),
			},
			{
				Url:    "layer2",
				Sha256: "95cf1a2e1698fe3ca1fcc3f653119146b271d0b62e487ec264441e886a11bd06",
				Size:   int64(len("layer2")),
			},
		},
	}
	downloader := &MockDownloader{usedUrl: make(map[string]struct{})}
	file, err := GenerateInputFile(input, downloader)
	if err != nil {
		t.Fatalf("Generate file failed: %s", err.Error())
	}
	defer file.Close()
	if _, err := os.Stat(file.Name()); err != nil {
		t.Fatalf("Generated file not exists: %s", file.Name())
	}
	os.Remove(file.Name())
}

type MockDownloader struct {
	usedUrl map[string]struct{}
}

func (d *MockDownloader) Download(url string) (io.ReadCloser, error) {
	if _, ok := d.usedUrl[url]; ok {
		return nil, errors.New("url already been used")
	} else {
		d.usedUrl[url] = struct{}{}
	}

	if strings.HasSuffix(url, "manifest.json") {
		return os.Open("testdata/test-manifest.json")
	} else {
		return io.NopCloser(bytes.NewReader([]byte(url))), nil
	}
}
