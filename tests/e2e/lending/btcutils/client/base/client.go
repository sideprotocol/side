package base

import (
	"fmt"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

// Client defines the base client
type Client struct {
	HTTPClient *fasthttp.Client

	Attempts int
	Interval time.Duration
}

// NewClient creates a base client instance
func NewClient(attempts int, interval time.Duration) *Client {
	return &Client{
		HTTPClient: &fasthttp.Client{},
		Attempts:   attempts,
		Interval:   interval,
	}
}

// GetBaseOptions builds the base request options
func (c *Client) GetBaseOptions() *RequestOptions {
	return &RequestOptions{
		Headers:  make(map[string]string, 0),
		Attempts: c.Attempts,
		Interval: c.Interval,
	}
}

// Request initiates an HTTP request with the given options
func (c *Client) Request(method, url string, opts *RequestOptions) (int, []byte, error) {
	headers, body, isJSON, attempts, interval := GetOptions(opts)

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)

	req.Header.SetMethod(method)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if body != nil {
		if isJSON {
			req.Header.Set("Content-Type", "application/json")
		}

		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))

		req.SetBody(body)
	}

	resp := fasthttp.AcquireResponse()

	var err error

	for i := 0; i < attempts; i++ {
		err = c.HTTPClient.Do(req, resp)
		if err == nil && resp.StatusCode() == http.StatusOK {
			break
		}

		if i+1 < attempts {
			time.Sleep(interval)
		}
	}

	if err != nil {
		err = fmt.Errorf("failed to request, err: %v", err)
	}

	return resp.StatusCode(), resp.Body(), err
}
