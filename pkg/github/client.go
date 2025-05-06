package github

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// WorkflowFailure represents a failed workflow run
type WorkflowFailure struct {
	Repo      string
	PRNumber  int
	Workflow  string
	StartedAt time.Time
	URL       string
	PRURL     string
}

// Client defines the interface for GitHub operations
type Client interface {
	GetFailedWorkflows(ctx context.Context, prNumber string, repo string) error
	ListAllFailedWorkflows(ctx context.Context, days int) (map[string][]WorkflowFailure, error)
}

// GitHubClient implements the Client interface
type GitHubClient struct {
	client *github.Client
	owner  string
}

// NewClient creates a new GitHub client
func NewClient(token, owner string) Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GitHubClient{
		client: github.NewClient(tc),
		owner:  owner,
	}
}

// GetFailedWorkflows retrieves failed workflows for a specific PR
func (g *GitHubClient) GetFailedWorkflows(ctx context.Context, prNumber string, repo string) error {
	if repo == "" {
		return fmt.Errorf("repository name is required")
	}

	// Convert PR number to int
	prNum, err := strconv.Atoi(prNumber)
	if err != nil {
		return fmt.Errorf("invalid PR number: %v", err)
	}

	// Get PR details
	pr, _, err := g.client.PullRequests.Get(ctx, g.owner, repo, prNum)
	if err != nil {
		return fmt.Errorf("error getting PR: %v", err)
	}

	// Get workflow runs for the PR
	opts := &github.ListWorkflowRunsOptions{
		Branch: *pr.Head.Ref,
		Status: "failure",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	runs, _, err := g.client.Actions.ListRepositoryWorkflowRuns(ctx, g.owner, repo, opts)
	if err != nil {
		return fmt.Errorf("error getting workflow runs: %v", err)
	}

	if len(runs.WorkflowRuns) == 0 {
		return nil
	}

	return nil
}

// ListAllFailedWorkflows retrieves all failed workflows across repositories
func (g *GitHubClient) ListAllFailedWorkflows(ctx context.Context, days int) (map[string][]WorkflowFailure, error) {
	// List all repositories for the organization with pagination
	var allRepos []*github.Repository
	listOpts := &github.RepositoryListByOrgOptions{
		Type:      "all",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    1,
		},
	}

	page := 1
	for {
		listOpts.Page = page

		repos, resp, err := g.client.Repositories.ListByOrg(ctx, g.owner, listOpts)
		if err != nil {
			return nil, fmt.Errorf("error listing repositories: %v", err)
		}

		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	if len(allRepos) == 0 {
		return nil, fmt.Errorf("no repositories found for organization %s", g.owner)
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	failures := make(map[string][]WorkflowFailure)

	for _, repo := range allRepos {
		opts := &github.ListWorkflowRunsOptions{
			Status: "failure",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		runs, _, err := g.client.Actions.ListRepositoryWorkflowRuns(ctx, g.owner, *repo.Name, opts)
		if err != nil {
			continue
		}

		for _, run := range runs.WorkflowRuns {
			if run.CreatedAt.Before(cutoffTime) {
				continue
			}

			if run.PullRequests != nil && len(run.PullRequests) > 0 {
				pr := run.PullRequests[0]
				prURL := fmt.Sprintf("https://github.com/%s/%s/pull/%d", g.owner, *repo.Name, *pr.Number)

				failure := WorkflowFailure{
					Repo:      *repo.Name,
					PRNumber:  *pr.Number,
					Workflow:  *run.Name,
					StartedAt: run.CreatedAt.Time,
					URL:       *run.HTMLURL,
					PRURL:     prURL,
				}

				failures[prURL] = append(failures[prURL], failure)
			}
		}
	}

	return failures, nil
}
