package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDownload(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "bkrepo-analysis-download")
	errorToPanic(func() error { return os.RemoveAll(tmpDir) })
	errorToPanic(func() error { return os.Mkdir(tmpDir, 0766) })

	downloader := NewChunkDownloader(
		8,
		tmpDir,
		map[string]string{
			"X-BKREPO-DOWNLOAD-REDIRECT-TO": "INNERCOS",
			"Authorization":                 os.Getenv("AUTHORIZATION"),
		},
		nil,
	)
	file, err := downloader.Download(os.Getenv("URL"))
	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		fmt.Println("error calculating hash:", err)
	} else {
		fmt.Printf("SHA256 hash of file: %x\n", h.Sum(nil))
	}

	if err != nil {
		fmt.Printf("download failed: %s\n", err.Error())
	}
	if err = file.Close(); err != nil {
		fmt.Printf("close file failed: %s\n", err.Error())
	}

	errorToPanic(func() error { return os.RemoveAll(tmpDir) })
}

func errorToPanic(f func() error) {
	if err := f(); err != nil {
		panic(err.Error())
	}
}
