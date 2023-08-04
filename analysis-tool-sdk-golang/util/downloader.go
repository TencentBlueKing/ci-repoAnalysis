package util

import (
	"errors"
	"io"
	"net/http"
)

// Downloader 下载器接口
type Downloader interface {
	// Download 从指定url获取输入流
	Download(url string) (io.ReadCloser, error)
}

// DefaultDownloader 默认下载器实现
type DefaultDownloader struct {
	client *http.Client
}

// NewDownloader 创建默认下载器
func NewDownloader(client *http.Client) Downloader {
	var c = client
	if client == nil {
		c = http.DefaultClient
	}

	return &DefaultDownloader{
		client: c,
	}
}

// Download 从指定url获取输入流
func (d *DefaultDownloader) Download(url string) (io.ReadCloser, error) {
	Info("downloading %s", url)
	response, err := d.client.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("download failed, status: " + response.Status)
	}

	return response.Body, nil
}
