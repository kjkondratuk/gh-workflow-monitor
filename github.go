package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/joho/godotenv"
)

type GitHubClient struct {
	client *github.Client
	owner  string
}

type WorkflowFailure struct {
	Repo      string
	PRNumber  int
	Workflow  string
	StartedAt time.Time
	URL       string
	PRURL     string
}

func NewGitHubClient() (*GitHubClient, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return nil, fmt.Errorf("GITHUB_OWNER environment variable is required")
	}

	ts := github.BasicAuthTransport{
		Username: strings.TrimSpace(token),
	}

	client := github.NewClient(ts.Client())

	return &GitHubClient{
		client: client,
		owner:  owner,
	}, nil
}

func (g *GitHubClient) GetFailedWorkflows(prNumber string, repoInput string) error {
	ctx := context.Background()

	if repoInput == "" {
		return fmt.Errorf("repository name is required. Use -r flag to specify the repository")
	}

	// Parse repository name
	var repo string
	parts := strings.Split(repoInput, "/")
	if len(parts) == 2 {
		// If owner/repo format is provided, use that owner
		g.owner = parts[0]
		repo = parts[1]
	} else {
		repo = repoInput
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
		fmt.Println("No failed workflow runs found for this PR.")
		return nil
	}

	fmt.Printf("\nFailed workflow runs for PR #%s in %s/%s:\n", prNumber, g.owner, repo)
	fmt.Println("----------------------------------------")

	for _, run := range runs.WorkflowRuns {
		fmt.Printf("Workflow: %s\n", *run.Name)
		fmt.Printf("Status: %s\n", *run.Conclusion)
		fmt.Printf("Started at: %s\n", run.CreatedAt.Format(time.RFC3339))
		fmt.Printf("URL: %s\n", *run.HTMLURL)
		fmt.Println("----------------------------------------")
	}

	return nil
}

func (g *GitHubClient) ListAllFailedWorkflows(days int) error {
	ctx := context.Background()

	fmt.Printf("Fetching repositories for organization %s...\n", g.owner)

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
		fmt.Printf("\rFetching repositories page %d...", page)
		listOpts.Page = page

		repos, resp, err := g.client.Repositories.ListByOrg(ctx, g.owner, listOpts)
		if err != nil {
			return fmt.Errorf("error listing repositories: %v", err)
		}

		fmt.Printf("\nPage %d: Found %d repositories\n", page, len(repos))
		allRepos = append(allRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	if len(allRepos) == 0 {
		return fmt.Errorf("no repositories found for organization %s. Please check your GITHUB_TOKEN permissions", g.owner)
	}

	fmt.Printf("\nFound %d repositories in total. Checking for failed workflows...\n", len(allRepos))
	fmt.Println("----------------------------------------")

	cutoffTime := time.Now().AddDate(0, 0, -days)
	checkedRepos := 0
	failures := make(map[string][]WorkflowFailure) // Map of PR URL to failures

	for _, repo := range allRepos {
		checkedRepos++
		fmt.Printf("\rChecking repository (%d/%d): %s", checkedRepos, len(allRepos), *repo.Name)

		// Get workflow runs for the repository
		opts := &github.ListWorkflowRunsOptions{
			Status: "failure",
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		}

		runs, _, err := g.client.Actions.ListRepositoryWorkflowRuns(ctx, g.owner, *repo.Name, opts)
		if err != nil {
			fmt.Printf("\nWarning: Error getting workflow runs for %s: %v\n", *repo.Name, err)
			continue
		}

		// Filter runs by date and check if they're associated with a PR
		for _, run := range runs.WorkflowRuns {
			if run.CreatedAt.Before(cutoffTime) {
				continue
			}

			// Get PR information if available
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

	// Clear the progress line
	fmt.Printf("\r")
	fmt.Printf("\nChecked %d repositories in total.\n", len(allRepos))

	if len(failures) == 0 {
		fmt.Printf("No failed workflow runs found in the last %d days.\n", days)
		return nil
	}

	// Print summary
	fmt.Printf("\nFound %d PRs with failed workflows in the last %d days:\n", len(failures), days)
	fmt.Println("----------------------------------------")

	for prURL, prFailures := range failures {
		// Sort failures by time (most recent first)
		sort.Slice(prFailures, func(i, j int) bool {
			return prFailures[i].StartedAt.After(prFailures[j].StartedAt)
		})

		latestFailure := prFailures[0]
		fmt.Printf("\nPR: %s\n", prURL)
		fmt.Printf("Repository: %s\n", latestFailure.Repo)
		fmt.Printf("Latest failure: %s (%s)\n", latestFailure.Workflow, latestFailure.StartedAt.Format(time.RFC3339))
		fmt.Printf("Total failures: %d\n", len(prFailures))
		fmt.Println("----------------------------------------")
	}

	return nil
}
