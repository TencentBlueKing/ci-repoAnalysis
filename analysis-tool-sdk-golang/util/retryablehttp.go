package util

import (
	"context"
	"github.com/hashicorp/go-retryablehttp"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

var (
	// DefaultClient 默认HTTP客户端
	DefaultClient = CreateHttpClient(CreateTransport(nil))

	// We need to consume response bodies to maintain http connections, but
	// limit the size we consume to respReadLimit.
	respReadLimit = int64(4096)

	// Default retry configuration
	defaultRetryWaitMin = 1 * time.Second
	defaultRetryWaitMax = 30 * time.Second
	defaultRetryMax     = 4
)

// SetDefault 设置默认HTTP客户端
func SetDefault(client *retryablehttp.Client) {
	DefaultClient = client
}

// CreateHttpClient 创建http客户端
func CreateHttpClient(transport *http.Transport) *retryablehttp.Client {
	return &retryablehttp.Client{
		HTTPClient:   &http.Client{Transport: transport},
		Logger:       slog.Default(),
		RetryWaitMin: defaultRetryWaitMin,
		RetryWaitMax: defaultRetryWaitMax,
		RetryMax:     defaultRetryMax,
		CheckRetry:   retryablehttp.DefaultRetryPolicy,
		Backoff:      retryablehttp.DefaultBackoff,
	}
}

// CreateTransport 创建Transport
func CreateTransport(
	dialContext func(ctx context.Context, network, addr string) (net.Conn, error),
) *http.Transport {
	d := dialContext
	if d == nil {
		d = (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext
	}

	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           d,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}

// DrainBody 读取[body]并丢弃数据方便复用连接
func DrainBody(body io.ReadCloser) {
	defer body.Close()
	_, err := io.Copy(io.Discard, io.LimitReader(body, respReadLimit))
	if err != nil {
		Error("[ERR] error reading response body: %v", err)
	}
}
