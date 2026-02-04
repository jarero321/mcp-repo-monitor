package github

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/logging"
)

// RateLimiter tracks GitHub API rate limits and blocks requests when near the limit.
type RateLimiter struct {
	mu           sync.Mutex
	remaining    int
	resetTime    time.Time
	threshold    int // Minimum remaining before waiting
	logger       *logging.Logger
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(threshold int, logger *logging.Logger) *RateLimiter {
	return &RateLimiter{
		remaining: 5000, // GitHub default
		threshold: threshold,
		logger:    logger.WithComponent("ratelimit"),
	}
}

// UpdateFromResponse updates rate limit info from HTTP response headers.
func (r *RateLimiter) UpdateFromResponse(resp *http.Response) {
	if resp == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			r.remaining = val
		}
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			r.resetTime = time.Unix(val, 0)
		}
	}

	if r.remaining <= r.threshold {
		r.logger.Warn("rate limit low",
			"remaining", r.remaining,
			"reset_at", r.resetTime.Format(time.RFC3339),
		)
	}
}

// Wait blocks if rate limit is near the threshold.
// Returns true if it had to wait.
func (r *RateLimiter) Wait() bool {
	r.mu.Lock()
	remaining := r.remaining
	resetTime := r.resetTime
	r.mu.Unlock()

	if remaining > r.threshold {
		return false
	}

	waitDuration := time.Until(resetTime)
	if waitDuration <= 0 {
		return false
	}

	// Cap wait time at 60 seconds
	if waitDuration > 60*time.Second {
		waitDuration = 60 * time.Second
	}

	r.logger.Info("waiting for rate limit reset",
		"remaining", remaining,
		"wait_duration", waitDuration.String(),
	)

	time.Sleep(waitDuration)
	return true
}

// Remaining returns the current remaining rate limit.
func (r *RateLimiter) Remaining() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.remaining
}

// ResetTime returns when the rate limit resets.
func (r *RateLimiter) ResetTime() time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.resetTime
}
