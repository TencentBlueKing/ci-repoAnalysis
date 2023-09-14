package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

// ChunkDownloader 分片下载器
type ChunkDownloader struct {
	WorkerCount int
	TmpDir      string
	Headers     map[string]string
}

// NewChunkDownloader 创建分片下载器
func NewChunkDownloader(workerCount int, tmpDir string, headers map[string]string) *ChunkDownloader {
	return &ChunkDownloader{
		WorkerCount: workerCount,
		TmpDir:      tmpDir,
		Headers:     headers,
	}
}

// Download 分片下载
func (d *ChunkDownloader) Download(url string) (io.ReadCloser, error) {
	defer timer("chunk download finished,")()
	Info("downloading %s", url)
	file, err := os.CreateTemp(d.TmpDir, "*-download.tmp")
	if err != nil {
		return nil, err
	}

	if err = d.chunkDownload(url, file); err != nil {
		return nil, err
	}

	return file, nil
}

func (d *ChunkDownloader) chunkDownload(url string, outputFile *os.File) error {
	fileSize, err := d.getFileSize(url)
	if err != nil {
		return err
	}

	var chunkSize int
	workCount := runtime.NumCPU()
	if d.WorkerCount > 0 {
		workCount = d.WorkerCount
	}
	chunkSize = fileSize / workCount
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(workCount)

	for i := 0; i < workCount; i++ {
		start := i * chunkSize
		end := start + chunkSize - 1
		if i == workCount-1 {
			end = fileSize - 1
		}

		g.Go(
			func() error {
				return d.doDownload(ctx, url, outputFile, start, end)
			},
		)
	}

	return g.Wait()
}

func (d *ChunkDownloader) doDownload(ctx context.Context, url string, file *os.File, start int, end int) error {
	defer timer(fmt.Sprintf("download chunk %d-%d success,", start, end))()
	Info("start download chunk %d-%d", start, end)
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	d.setHeaders(req)
	rangeHeader := "bytes=" + strconv.Itoa(start) + "-" + strconv.Itoa(end)
	req.Header.Set("Range", rangeHeader)

	res, err := DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer DrainBody(res.Body)

	if res.StatusCode != http.StatusPartialContent {
		return errors.New("download chunk failed: " + res.Status)
	}

	buf := make([]byte, 4*1024)
	off := start
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := res.Body.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				return nil
			}

			_, err = file.WriteAt(buf[:n], int64(off))
			if err != nil {
				return err
			}
			off += n
		}

	}
}

func (d *ChunkDownloader) getFileSize(url string) (int, error) {
	req, _ := retryablehttp.NewRequest("HEAD", url, nil)
	d.setHeaders(req)
	res, err := DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer DrainBody(res.Body)

	if res.StatusCode != http.StatusOK {
		return 0, errors.New("get file size failed, status: " + res.Status)
	}

	sizeHeader := res.Header.Get("Content-Length")
	size, err := strconv.Atoi(sizeHeader)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (d *ChunkDownloader) setHeaders(req *retryablehttp.Request) {
	for k, v := range d.Headers {
		req.Header.Set(k, v)
	}
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
