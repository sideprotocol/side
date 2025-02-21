package base

import (
	"time"
)

const DefaultRequestAttempts = 10
const DefaultRequestInterval = 500 * time.Millisecond

// RequestOptions defines the options for the HTTP request
type RequestOptions struct {
	Headers map[string]string

	Body   []byte
	IsJSON bool

	Attempts int
	Interval time.Duration
}

// NewRequestOptions creates a new RequestOptions instance
func NewRequestOptions(headers map[string]string, body []byte, isJSON bool, attempts int, interval time.Duration) *RequestOptions {
	req := new(RequestOptions)

	req.Headers = headers

	req.Body = body
	req.IsJSON = isJSON

	req.Attempts = attempts
	req.Interval = interval

	return req
}

// GetOptions returns the options with default value if not provided
func GetOptions(opts *RequestOptions) (headers map[string]string, body []byte, isJSON bool, attempts int, interval time.Duration) {
	attempts = DefaultRequestAttempts
	interval = DefaultRequestInterval

	if opts != nil {
		headers = opts.Headers

		body = opts.Body
		isJSON = opts.IsJSON

		if opts.Attempts > 0 {
			attempts = opts.Attempts
		}

		if opts.Interval > 0 {
			interval = opts.Interval
		}
	}

	return
}
