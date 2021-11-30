package ratelimiting

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/sirupsen/logrus"
)

const (
	ctxId = ctxIdType("id")
)

// ctxIdType is used to avoid collisions between packages using context
type ctxIdType string

// rateLimitTransport implements GitHub's best practices around respecting abuse rate limits and secondary rate limits
// This implementation allows git-xargs to operate as a good citizen and avoid being rate limited when making lots of concurrent requests
// Inspired by: https://github.com/integrations/terraform-provider-github/blob/696168b940e7b2a116fdc757a7200314454bb9bd/github/transport.go
type RateLimitTransport struct {
	transport        http.RoundTripper
	delayNextRequest bool
	writeDelay       time.Duration

	m sync.Mutex
}

func (rlt *RateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	logger := logging.GetLogger("git-xargs")

	// Make requests for a single user or client ID serially
	// This is also necessary for safely saving and restoring bodies between retries below
	rlt.lock(req)

	// If you're making a large number of POST, PATCH, PUT or DELETE requests
	// for a single user or client ID, wait at least one second between each request
	if rlt.delayNextRequest {
		logger.WithFields(logrus.Fields{
			"Write delay": rlt.writeDelay,
		}).Debug("git-xargs HTTP RoundTrip sleeping between write operations")
		time.Sleep(rlt.writeDelay)
	}

	rlt.delayNextRequest = isWriteMethod(req.Method)

	resp, err := rlt.transport.RoundTrip(req)
	if err != nil {
		rlt.unlock(req)
		return resp, err
	}

	// Make response body accessible for retries & debugging
	// (work around bug in GitHub SDK)
	// See https://github.com/google/go-github/pull/986
	r1, r2, err := drainBody(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = r1
	ghErr := github.CheckResponse(resp)
	resp.Body = r2

	// When you have been limited, use the Retry-After response header to slow down
	if arlErr, ok := ghErr.(*github.AbuseRateLimitError); ok {
		logger := logging.GetLogger("git-xargs")

		rlt.delayNextRequest = false
		retryAfter := arlErr.GetRetryAfter()
		logger.WithFields(logrus.Fields{
			"retryAfter": retryAfter,
		}).Debug(fmt.Sprintf("Abuse detection mechanism triggered, sleeping for %s before retrying", retryAfter))
		time.Sleep(retryAfter)
		rlt.unlock(req)
		return rlt.RoundTrip(req)
	}

	if rlErr, ok := ghErr.(*github.RateLimitError); ok {
		logger := logging.GetLogger("git-xargs")

		rlt.delayNextRequest = false
		retryAfter := time.Until(rlErr.Rate.Reset.Time)
		logger.WithFields(logrus.Fields{
			"retryAfter": retryAfter,
			"Rate Limit": rlErr.Rate.Limit,
		}).Debug(fmt.Sprintf("Rate limit %d reached, sleeping for %s (until %s) before retrying", rlErr.Rate.Limit, retryAfter, time.Now().Add(retryAfter)))
		time.Sleep(retryAfter)
		rlt.unlock(req)
		return rlt.RoundTrip(req)
	}

	rlt.unlock(req)

	return resp, nil

}

func (rlt *RateLimitTransport) lock(req *http.Request) {
	logger := logging.GetLogger("git-xargs")
	ctx := req.Context()
	logger.Trace(fmt.Sprintf("Acquiring lock for Github API request (%q)", ctx.Value(ctxId)))
	rlt.m.Lock()
}

func (rlt *RateLimitTransport) unlock(req *http.Request) {
	logger := logging.GetLogger("git-xargs")
	ctx := req.Context()
	logger.Trace(fmt.Sprintf("Releasing lock for Github API request (%q)", ctx.Value(ctxId)))
	rlt.m.Unlock()
}

type RateLimitTransportOption func(*RateLimitTransport)

// NewRateLimitTransport takes in an http.RoundTripper and a variadic list of
// optional functions that modify the RateLimitTransport struct itself. This
// may be used to alter the write delay in between requests, for example
func NewRateLimitTransport(rt http.RoundTripper, options ...RateLimitTransportOption) *RateLimitTransport {
	// Default to 1 second of delay if none is provided
	rlt := &RateLimitTransport{transport: rt, writeDelay: 1 * time.Second}

	for _, opt := range options {
		opt(rlt)
	}

	return rlt
}

// WithWriteDelay is used to set the write delay between requests
func WithWriteDelay(d time.Duration) RateLimitTransportOption {
	return func(rlt *RateLimitTransport) {
		rlt.writeDelay = d
	}
}

// drainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes
func drainBody(b io.ReadCloser) (r1, r2 io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed
		return http.NoBody, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return ioutil.NopCloser(&buf), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func isWriteMethod(method string) bool {
	switch method {
	case "POST", "PATCH", "PUT", "DELETE":
		return true
	}
	return false
}
