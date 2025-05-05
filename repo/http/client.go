package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ginproject/middleware/log"
)

const (
	defaultTimeout = 10 * time.Second
)

// HTTPClient 是HTTP客户端的接口
type HTTPClient interface {
	Get(ctx context.Context, url string) ([]byte, error)
}

// Client 是HTTP客户端的实现
type Client struct {
	client  *http.Client
	baseURL string
}

// NewClient 创建一个新的HTTP客户端
func NewClient(baseURL string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		baseURL: baseURL,
	}
}

// Get 发送HTTP GET请求
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.ErrorWithContext(ctx, "创建HTTP请求失败", "错误", err, "URL", url)
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	log.InfoWithContext(ctx, "发送HTTP请求", "方法", http.MethodGet, "URL", url)
	resp, err := c.client.Do(req)
	if err != nil {
		log.ErrorWithContext(ctx, "HTTP请求失败", "错误", err, "URL", url)
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.ErrorWithContext(ctx, "HTTP响应状态码非200", "状态码", resp.StatusCode, "URL", url)
		return nil, fmt.Errorf("HTTP响应状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.ErrorWithContext(ctx, "读取HTTP响应体失败", "错误", err, "URL", url)
		return nil, fmt.Errorf("读取HTTP响应体失败: %w", err)
	}

	log.InfoWithContext(ctx, "HTTP请求成功", "URL", url, "状态码", resp.StatusCode)
	return body, nil
}

// ParseResponse 使用上下文解析HTTP响应体到指定的结构体
func ParseResponse(ctx context.Context, data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		log.ErrorWithContext(ctx, "解析HTTP响应失败", "错误", err, "数据", string(data))
		return fmt.Errorf("解析HTTP响应失败: %w", err)
	}
	return nil
}
