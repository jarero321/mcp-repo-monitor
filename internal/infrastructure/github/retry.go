package github

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/logging"
	"github.com/google/go-github/v60/github"
)

// RetryConfig configures retry behavior.
type RetryConfig struct {
	MaxRetries     int           // Maximum number of retry attempts
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	Multiplier     float64       // Backoff multiplier
}

// DefaultRetryConfig returns sensible retry defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Second,
		MaxBackoff:     30 * time.Second,
		Multiplier:     2.0,
	}
}

// Retryer handles retry logic with exponential backoff.
type Retryer struct {
	config RetryConfig
	logger *logging.Logger
}

// NewRetryer creates a new retryer.
func NewRetryer(config RetryConfig, logger *logging.Logger) *Retryer {
	return &Retryer{
		config: config,
		logger: logger.WithComponent("retry"),
	}
}

// isRetryableStatusCode checks if the status code should trigger a retry.
func isRetryableStatusCode(code int) bool {
	switch code {
	case http.StatusTooManyRequests,      // 429
		http.StatusInternalServerError,   // 500
		http.StatusBadGateway,            // 502
		http.StatusServiceUnavailable,    // 503
		http.StatusGatewayTimeout:        // 504
		return true
	default:
		return false
	}
}

// isRetryableError checks if an error should trigger a retry.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for GitHub error response
	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) {
		return isRetryableStatusCode(ghErr.Response.StatusCode)
	}

	// Check for rate limit error
	var rateLimitErr *github.RateLimitError
	if errors.As(err, &rateLimitErr) {
		return true
	}

	// Check for abuse rate limit error
	var abuseErr *github.AbuseRateLimitError
	if errors.As(err, &abuseErr) {
		return true
	}

	return false
}

// Do executes a function with retry logic.
func (r *Retryer) Do(ctx context.Context, operation string, fn func() error) error {
	var lastErr error
	backoff := r.config.InitialBackoff

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			r.logger.Info("retrying operation",
				"operation", operation,
				"attempt", attempt,
				"max_retries", r.config.MaxRetries,
				"backoff", backoff.String(),
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			// Calculate next backoff
			backoff = time.Duration(float64(backoff) * r.config.Multiplier)
			if backoff > r.config.MaxBackoff {
				backoff = r.config.MaxBackoff
			}
		}

		err := fn()
		if err == nil {
			if attempt > 0 {
				r.logger.Info("operation succeeded after retry",
					"operation", operation,
					"attempts", attempt+1,
				)
			}
			return nil
		}

		lastErr = err

		if !isRetryableError(err) {
			r.logger.Debug("error is not retryable",
				"operation", operation,
				"error", err.Error(),
			)
			return err
		}

		r.logger.Warn("operation failed with retryable error",
			"operation", operation,
			"attempt", attempt,
			"error", err.Error(),
		)
	}

	r.logger.Error("operation failed after all retries",
		"operation", operation,
		"attempts", r.config.MaxRetries+1,
		"error", lastErr.Error(),
	)

	return lastErr
}
