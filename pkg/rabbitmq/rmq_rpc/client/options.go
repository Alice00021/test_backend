package client

import "time"

// Option -.
type Option func(*Client)

// Timeout - apply timeout.
func Timeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// ConnWaitTime - apply connection wait time.
func ConnWaitTime(timeout time.Duration) Option {
	return func(c *Client) {
		c.conn.WaitTime = timeout
	}
}

// ConnAttempts - apply attempts.
func ConnAttempts(attempts int) Option {
	return func(c *Client) {
		c.conn.Attempts = attempts
	}
}

// Headers - apply headers.
func Headers(headers map[string]interface{}) Option {
	return func(c *Client) {
		c.headers = headers
	}
}
