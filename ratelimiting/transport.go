package ratelimiting

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/gruntwork-io/go-commons/logging"
	"github.com/sirupsen/logrus"
)

const (
	ctxId               = ctxIdType("id")
	defaultSleepSeconds = 2
	maxSleepSeconds     = 10
	minSleepSeconds     = 3
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

	// Determine if we have tripped the secondary rate limits, which are separate from both:
	// 1. Abuse rate limits
	// 2. Regular rate limits
	// Secondary rate limits are tripped when creating resources that generate notifications (such as Pull Requests)
	// in rapid succession. We key off the URL provided in the DocumentationURL of the error response provided by Github
	// to know when we've tripped the secondary rate limits (which is the same method used by the go-github library's CheckResponse method)
	// Note that this is an unfortunate and quite brittle workaround, but we're forced to use this because Github's error responses are
	// inconsistent. For more information on Github's secondary rate limits, see this document:
	// https://docs.github.com/en/rest/guides/best-practices-for-integrators#dealing-with-secondary-rate-limits
	var errMessage string
	if ghErr != nil {
		errMessage = ghErr.(*github.ErrorResponse).DocumentationURL
	}

	// When you have been limited, use the Retry-After response header to slow down
	if arlErr, ok := ghErr.(*github.AbuseRateLimitError); ok {
		logger := logging.GetLogger("git-xargs")

		logger.Debug("git-xargs received Abuse RateLimitError from Github")
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

		logger.Debug("git-xargs received Regular RateLimitError from Github")
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

	if strings.Contains(errMessage, "secondary-rate-limits") {
		logger := logging.GetLogger("git-xargs")

		logger.Debug("git-xargs received Secondary rate limit error from Github!")
		// Secondary Rate limits can be tripped when creating lots of resources that trigger notifications, such as PRs
		// Unfortunately, in these cases the `Retry-After` header may intentionally be omitted from the response
		// Therefore, when this occurs we have to add a small random sleep
		// See: https://github.community/t/retry-after-header-missing/179442/2
		if v := resp.Header["Retry-After"]; len(v) > 0 {
			// It's almost certain this header will be unavailable to us, so we use the logic below this
			// if statement to pick a small random duration to sleep between write (PUT, POST, PATCH, etc) requests
			retryAfterSeconds, _ := strconv.ParseInt(v[0], 10, 64)
			retryAfter := time.Duration(retryAfterSeconds) * time.Second
			time.Sleep(retryAfter)
			rlt.unlock(req)
			return rlt.RoundTrip(req)
		}

		rand.Seed(time.Now().UnixNano())
		duration := rand.Intn(maxSleepSeconds-minSleepSeconds) + minSleepSeconds
		logger.Debug(fmt.Sprintf("git-xargs http transport sleeping %d seconds before retrying because Github's secondary rate limits have been tripped", duration))
		time.Sleep(time.Duration(duration) * time.Second)
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
	// Default to 2 second of delay if none is provided
	rlt := &RateLimitTransport{transport: rt, writeDelay: defaultSleepSeconds * time.Second}

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
