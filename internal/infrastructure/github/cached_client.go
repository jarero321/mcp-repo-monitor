package github

import (
	"context"
	"fmt"

	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/cache"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/logging"
)

// CachedClient wraps a Client with caching capabilities.
type CachedClient struct {
	*Client
	cache  *cache.Cache
	logger *logging.Logger
}

// NewCachedClient creates a new cached client wrapper.
func NewCachedClient(client *Client, cache *cache.Cache, logger *logging.Logger) *CachedClient {
	return &CachedClient{
		Client: client,
		cache:  cache,
		logger: logger.WithComponent("cached-client"),
	}
}

// ListRepositories returns cached repositories or fetches from API.
func (c *CachedClient) ListRepositories(ctx context.Context, filter string, includeArchived bool) ([]entity.Repository, error) {
	key := fmt.Sprintf("repos:%s:%v", filter, includeArchived)

	if cached, ok := c.cache.Get(key); ok {
		c.logger.Debug("cache hit", "key", key)
		return cached.([]entity.Repository), nil
	}

	c.logger.Debug("cache miss", "key", key)
	repos, err := c.Client.ListRepositories(ctx, filter, includeArchived)
	if err != nil {
		return nil, err
	}

	c.cache.Set(key, repos)
	return repos, nil
}

// ListPullRequests returns cached PRs or fetches from API.
func (c *CachedClient) ListPullRequests(ctx context.Context, filter entity.PRFilter) ([]entity.PullRequest, error) {
	key := fmt.Sprintf("prs:%s:%s:%d", filter.Repository, filter.State, filter.Limit)

	if cached, ok := c.cache.Get(key); ok {
		c.logger.Debug("cache hit", "key", key)
		return cached.([]entity.PullRequest), nil
	}

	c.logger.Debug("cache miss", "key", key)
	prs, err := c.Client.ListPullRequests(ctx, filter)
	if err != nil {
		return nil, err
	}

	c.cache.Set(key, prs)
	return prs, nil
}

// Note: CompareBranches is NOT cached because it needs real-time data for drift detection.
// GetRepository, CreatePullRequest, and write operations are also not cached.
