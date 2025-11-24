package proxy

import (
	"time"

	"vibeway/pkg/logger"

	"github.com/valyala/fasthttp"
)

type ProxyClient struct {
	client *fasthttp.Client
}

func NewProxyClient(readTimeout, writeTimeout time.Duration) *ProxyClient {
	return &ProxyClient{
		client: &fasthttp.Client{
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
	}
}

func (p *ProxyClient) Do(req *fasthttp.Request, resp *fasthttp.Response, upstreamURL string) error {
	// Prepare request
	req.SetRequestURI(upstreamURL)
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")

	// Execute request
	start := time.Now()
	err := p.client.Do(req, resp)
	duration := time.Since(start)

	// Log result
	fields := map[string]interface{}{
		"upstream": upstreamURL,
		"latency":  duration.String(),
		"status":   resp.StatusCode(),
	}

	if err != nil {
		logger.Error("Proxy request failed", err, fields)
		return err
	}

	logger.Info("Proxy request success", fields)
	return nil
}
