package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/carlos/mcp-repo-monitor/internal/domain/entity"
	"github.com/carlos/mcp-repo-monitor/internal/infrastructure/logging"
	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

type Client struct {
	gh          *github.Client
	rateLimiter *RateLimiter
	retryer     *Retryer
	logger      *logging.Logger
}

// rateLimitTransport wraps http.RoundTripper to capture rate limit headers.
type rateLimitTransport struct {
	base        http.RoundTripper
	rateLimiter *RateLimiter
}

func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if resp != nil {
		t.rateLimiter.UpdateFromResponse(resp)
	}
	return resp, err
}

func NewClient(token string, logger *logging.Logger) *Client {
	clientLogger := logger.WithComponent("github")
	rateLimiter := NewRateLimiter(10, logger)
	retryer := NewRetryer(DefaultRetryConfig(), logger)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	baseClient := oauth2.NewClient(ctx, ts)

	// Wrap transport to capture rate limit headers
	baseClient.Transport = &rateLimitTransport{
		base:        baseClient.Transport,
		rateLimiter: rateLimiter,
	}

	clientLogger.Info("github client initialized")

	return &Client{
		gh:          github.NewClient(baseClient),
		rateLimiter: rateLimiter,
		retryer:     retryer,
		logger:      clientLogger,
	}
}

func (c *Client) ListRepositories(ctx context.Context, filter string, includeArchived bool) ([]entity.Repository, error) {
	c.rateLimiter.Wait()

	opts := &github.RepositoryListOptions{
		Affiliation: "owner,organization_member",
		Sort:        "updated",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []entity.Repository
	for {
		var repos []*github.Repository
		var resp *github.Response

		err := c.retryer.Do(ctx, "ListRepositories", func() error {
			var err error
			repos, resp, err = c.gh.Repositories.List(ctx, "", opts)
			return err
		})
		if err != nil {
			return nil, err
		}

		for _, r := range repos {
			if !includeArchived && r.GetArchived() {
				continue
			}
			if filter != "" && !strings.Contains(strings.ToLower(r.GetFullName()), strings.ToLower(filter)) {
				continue
			}

			allRepos = append(allRepos, toRepository(r))
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	c.logger.Debug("listed repositories", "count", len(allRepos))
	return allRepos, nil
}

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*entity.Repository, error) {
	c.rateLimiter.Wait()

	var r *github.Repository
	err := c.retryer.Do(ctx, "GetRepository", func() error {
		var err error
		r, _, err = c.gh.Repositories.Get(ctx, owner, repo)
		return err
	})
	if err != nil {
		return nil, err
	}
	result := toRepository(r)
	return &result, nil
}

func (c *Client) ListPullRequests(ctx context.Context, filter entity.PRFilter) ([]entity.PullRequest, error) {
	state := "open"
	if filter.State != "" {
		state = filter.State
	}

	limit := 30
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	var allPRs []entity.PullRequest

	if filter.Repository != "" {
		parts := strings.Split(filter.Repository, "/")
		if len(parts) == 2 {
			prs, err := c.listRepoPRs(ctx, parts[0], parts[1], state, limit)
			if err != nil {
				return nil, err
			}
			allPRs = append(allPRs, prs...)
		}
	} else {
		repos, err := c.ListRepositories(ctx, "", false)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			parts := strings.Split(repo.FullName, "/")
			if len(parts) != 2 {
				continue
			}

			prs, err := c.listRepoPRs(ctx, parts[0], parts[1], state, limit)
			if err != nil {
				continue
			}
			allPRs = append(allPRs, prs...)

			if len(allPRs) >= limit {
				break
			}
		}
	}

	if len(allPRs) > limit {
		allPRs = allPRs[:limit]
	}

	return allPRs, nil
}

func (c *Client) listRepoPRs(ctx context.Context, owner, repo, state string, limit int) ([]entity.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State:       state,
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: limit},
	}

	prs, _, err := c.gh.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	var result []entity.PullRequest
	for _, pr := range prs {
		result = append(result, toPullRequest(pr, owner+"/"+repo))
	}

	return result, nil
}

func (c *Client) GetPullRequest(ctx context.Context, owner, repo string, number int) (*entity.PullRequest, error) {
	pr, _, err := c.gh.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, err
	}
	result := toPullRequest(pr, owner+"/"+repo)
	return &result, nil
}

func (c *Client) CreatePullRequest(ctx context.Context, owner, repo, title, body, head, base string, draft bool) (*entity.PullRequest, error) {
	c.rateLimiter.Wait()

	newPR := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &head,
		Base:  &base,
		Draft: &draft,
	}

	var pr *github.PullRequest
	err := c.retryer.Do(ctx, "CreatePullRequest", func() error {
		var err error
		pr, _, err = c.gh.PullRequests.Create(ctx, owner, repo, newPR)
		return err
	})
	if err != nil {
		return nil, err
	}

	c.logger.Info("created pull request",
		"repo", owner+"/"+repo,
		"number", pr.GetNumber(),
		"head", head,
		"base", base,
	)

	result := toPullRequest(pr, owner+"/"+repo)
	return &result, nil
}

func (c *Client) ListCommits(ctx context.Context, filter entity.CommitFilter) ([]entity.Commit, error) {
	limit := 30
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	var allCommits []entity.Commit

	if filter.Repository != "" {
		parts := strings.Split(filter.Repository, "/")
		if len(parts) == 2 {
			commits, err := c.listRepoCommits(ctx, parts[0], parts[1], filter.Branch, filter.Since, limit)
			if err != nil {
				return nil, err
			}
			allCommits = append(allCommits, commits...)
		}
	} else {
		repos, err := c.ListRepositories(ctx, "", false)
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			parts := strings.Split(repo.FullName, "/")
			if len(parts) != 2 {
				continue
			}

			commits, err := c.listRepoCommits(ctx, parts[0], parts[1], filter.Branch, filter.Since, limit)
			if err != nil {
				continue
			}
			allCommits = append(allCommits, commits...)

			if len(allCommits) >= limit {
				break
			}
		}
	}

	if len(allCommits) > limit {
		allCommits = allCommits[:limit]
	}

	return allCommits, nil
}

func (c *Client) listRepoCommits(ctx context.Context, owner, repo, branch string, since *time.Time, limit int) ([]entity.Commit, error) {
	opts := &github.CommitsListOptions{
		SHA:         branch,
		ListOptions: github.ListOptions{PerPage: limit},
	}
	if since != nil {
		opts.Since = *since
	}

	commits, _, err := c.gh.Repositories.ListCommits(ctx, owner, repo, opts)
	if err != nil {
		return nil, err
	}

	var result []entity.Commit
	for _, commit := range commits {
		result = append(result, toCommit(commit, owner+"/"+repo, branch))
	}

	return result, nil
}

func (c *Client) ListWorkflowRuns(ctx context.Context, filter entity.CIFilter) ([]entity.WorkflowRun, error) {
	limit := 10
	if filter.Limit > 0 {
		limit = filter.Limit
	}

	parts := strings.Split(filter.Repository, "/")
	if len(parts) != 2 {
		return nil, nil
	}

	opts := &github.ListWorkflowRunsOptions{
		Branch:      filter.Branch,
		ListOptions: github.ListOptions{PerPage: limit},
	}

	var runs *github.WorkflowRuns
	var err error

	if filter.Workflow != "" {
		runs, _, err = c.gh.Actions.ListWorkflowRunsByFileName(ctx, parts[0], parts[1], filter.Workflow, opts)
	} else {
		runs, _, err = c.gh.Actions.ListRepositoryWorkflowRuns(ctx, parts[0], parts[1], opts)
	}

	if err != nil {
		return nil, err
	}

	var result []entity.WorkflowRun
	for _, run := range runs.WorkflowRuns {
		result = append(result, toWorkflowRun(run, filter.Repository))
	}

	return result, nil
}

func (c *Client) RerunWorkflow(ctx context.Context, owner, repo string, runID int64) error {
	_, err := c.gh.Actions.RerunWorkflowByID(ctx, owner, repo, runID)
	return err
}

func (c *Client) TriggerWorkflow(ctx context.Context, owner, repo, workflowID, ref string) error {
	event := github.CreateWorkflowDispatchEventRequest{
		Ref: ref,
	}
	_, err := c.gh.Actions.CreateWorkflowDispatchEventByFileName(ctx, owner, repo, workflowID, event)
	return err
}

func (c *Client) CompareBranches(ctx context.Context, owner, repo, base, head string) (*entity.BranchComparison, error) {
	c.rateLimiter.Wait()

	var comparison *github.CommitsComparison
	err := c.retryer.Do(ctx, "CompareBranches", func() error {
		var err error
		comparison, _, err = c.gh.Repositories.CompareCommits(ctx, owner, repo, base, head, nil)
		return err
	})
	if err != nil {
		return nil, err
	}

	var commits []entity.Commit
	for _, commit := range comparison.Commits {
		commits = append(commits, toCommit(commit, owner+"/"+repo, head))
	}

	var files []entity.ChangedFile
	for _, file := range comparison.Files {
		files = append(files, entity.ChangedFile{
			Filename:  file.GetFilename(),
			Status:    file.GetStatus(),
			Additions: file.GetAdditions(),
			Deletions: file.GetDeletions(),
			Changes:   file.GetChanges(),
			Patch:     file.GetPatch(),
		})
	}

	return &entity.BranchComparison{
		Repository:   owner + "/" + repo,
		ProdBranch:   base,
		DevBranch:    head,
		AheadBy:      comparison.GetAheadBy(),
		BehindBy:     comparison.GetBehindBy(),
		TotalCommits: comparison.GetTotalCommits(),
		GitHubStatus: comparison.GetStatus(),
		Commits:      commits,
		Files:        files,
	}, nil
}

func (c *Client) MergePullRequest(ctx context.Context, owner, repo string, number int, method, commitTitle string) (*entity.MergeResult, error) {
	c.rateLimiter.Wait()

	opts := &github.PullRequestOptions{
		MergeMethod: method,
	}

	var result *github.PullRequestMergeResult
	err := c.retryer.Do(ctx, "MergePullRequest", func() error {
		var err error
		result, _, err = c.gh.PullRequests.Merge(ctx, owner, repo, number, commitTitle, opts)
		return err
	})
	if err != nil {
		return nil, err
	}

	c.logger.Info("merged pull request",
		"repo", owner+"/"+repo,
		"number", number,
		"method", method,
	)

	return &entity.MergeResult{
		Success:     result.GetMerged(),
		SHA:         result.GetSHA(),
		Message:     result.GetMessage(),
		PRURL:       fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repo, number),
		PRNumber:    number,
		MergeMethod: entity.MergeMethod(method),
	}, nil
}

func (c *Client) DeleteBranch(ctx context.Context, owner, repo, branch string) error {
	c.rateLimiter.Wait()

	ref := "refs/heads/" + branch
	err := c.retryer.Do(ctx, "DeleteBranch", func() error {
		_, err := c.gh.Git.DeleteRef(ctx, owner, repo, ref)
		return err
	})
	if err != nil {
		return err
	}

	c.logger.Info("deleted branch",
		"repo", owner+"/"+repo,
		"branch", branch,
	)

	return nil
}

func (c *Client) GetCurrentUser(ctx context.Context) (string, error) {
	user, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}

func toRepository(r *github.Repository) entity.Repository {
	return entity.Repository{
		ID:            r.GetID(),
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		Description:   r.GetDescription(),
		Private:       r.GetPrivate(),
		Archived:      r.GetArchived(),
		Fork:          r.GetFork(),
		DefaultBranch: r.GetDefaultBranch(),
		Language:      r.GetLanguage(),
		Stars:         r.GetStargazersCount(),
		Forks:         r.GetForksCount(),
		OpenIssues:    r.GetOpenIssuesCount(),
		HTMLURL:       r.GetHTMLURL(),
		CloneURL:      r.GetCloneURL(),
		UpdatedAt:     r.GetUpdatedAt().Time,
		PushedAt:      r.GetPushedAt().Time,
	}
}

func toPullRequest(pr *github.PullRequest, repo string) entity.PullRequest {
	var labels []string
	for _, l := range pr.Labels {
		labels = append(labels, l.GetName())
	}

	var reviewers []string
	for _, r := range pr.RequestedReviewers {
		reviewers = append(reviewers, r.GetLogin())
	}

	result := entity.PullRequest{
		ID:           pr.GetID(),
		Number:       pr.GetNumber(),
		Title:        pr.GetTitle(),
		Body:         pr.GetBody(),
		State:        pr.GetState(),
		Draft:        pr.GetDraft(),
		HTMLURL:      pr.GetHTMLURL(),
		User:         pr.GetUser().GetLogin(),
		HeadBranch:   pr.GetHead().GetRef(),
		BaseBranch:   pr.GetBase().GetRef(),
		Mergeable:    pr.Mergeable,
		Additions:    pr.GetAdditions(),
		Deletions:    pr.GetDeletions(),
		ChangedFiles: pr.GetChangedFiles(),
		CreatedAt:    pr.GetCreatedAt().Time,
		UpdatedAt:    pr.GetUpdatedAt().Time,
		Labels:       labels,
		Reviewers:    reviewers,
		Repository:   repo,
	}

	if pr.MergedAt != nil {
		t := pr.MergedAt.Time
		result.MergedAt = &t
	}
	if pr.ClosedAt != nil {
		t := pr.ClosedAt.Time
		result.ClosedAt = &t
	}

	return result
}

func toCommit(c *github.RepositoryCommit, repo, branch string) entity.Commit {
	commit := c.GetCommit()
	author := commit.GetAuthor()

	return entity.Commit{
		SHA:         c.GetSHA(),
		Message:     commit.GetMessage(),
		Author:      author.GetName(),
		AuthorEmail: author.GetEmail(),
		Date:        author.GetDate().Time,
		HTMLURL:     c.GetHTMLURL(),
		Additions:   c.GetStats().GetAdditions(),
		Deletions:   c.GetStats().GetDeletions(),
		Repository:  repo,
		Branch:      branch,
	}
}

func toWorkflowRun(run *github.WorkflowRun, repo string) entity.WorkflowRun {
	return entity.WorkflowRun{
		ID:           run.GetID(),
		Name:         run.GetName(),
		WorkflowID:   run.GetWorkflowID(),
		WorkflowName: run.GetName(),
		HeadBranch:   run.GetHeadBranch(),
		HeadSHA:      run.GetHeadSHA(),
		Status:       run.GetStatus(),
		Conclusion:   run.GetConclusion(),
		HTMLURL:      run.GetHTMLURL(),
		RunNumber:    run.GetRunNumber(),
		RunAttempt:   run.GetRunAttempt(),
		CreatedAt:    run.GetCreatedAt().Time,
		UpdatedAt:    run.GetUpdatedAt().Time,
		Repository:   repo,
		Actor:        run.GetActor().GetLogin(),
		Event:        run.GetEvent(),
	}
}
