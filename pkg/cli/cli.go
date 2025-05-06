package cli

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/alecthomas/kong"
	"github.com/kjkondratuk/gh-workflow-monitor/pkg/github"
)

// CLI represents the command-line interface
type CLI struct {
	List struct {
		Days int `help:"Number of days to look back" default:"7"`
	} `cmd:"" help:"List all failed workflow runs"`

	Check struct {
		PR   string `help:"PR number to check" required:""`
		Repo string `help:"Repository name" required:""`
	} `cmd:"" help:"Check workflow failures for a specific PR"`
}

// HandleList handles the list command
func HandleList(client github.Client, days int) error {
	failures, err := client.ListAllFailedWorkflows(context.Background(), days)
	if err != nil {
		return fmt.Errorf("failed to list workflow failures: %w", err)
	}

	if len(failures) == 0 {
		fmt.Printf("No failed workflow runs found in the last %d days\n", days)
		return nil
	}

	fmt.Printf("Found %d failed workflow runs in the last %d days:\n\n", len(failures), days)
	for prURL, prFailures := range failures {
		fmt.Printf("PR: %s\n", prURL)
		for _, failure := range prFailures {
			fmt.Printf("  - Repository: %s\n", failure.Repo)
			fmt.Printf("    Workflow: %s\n", failure.Workflow)
			fmt.Printf("    Started: %s\n", failure.StartedAt.Format(time.RFC3339))
			fmt.Printf("    URL: %s\n\n", failure.URL)
		}
	}

	return nil
}

// HandleCheck handles the check command
func HandleCheck(client github.Client, prNumber string, repo string) error {
	if repo == "" {
		return fmt.Errorf("repository name is required")
	}

	err := client.GetFailedWorkflows(context.Background(), prNumber, repo)
	if err != nil {
		return fmt.Errorf("failed to check workflow failures: %w", err)
	}

	return nil
}

// Run executes the CLI application
func Run(client github.Client) error {
	var cli CLI
	ctx := kong.Parse(&cli)

	switch ctx.Command() {
	case "list":
		return HandleList(client, cli.List.Days)
	case "check":
		return HandleCheck(client, cli.Check.PR, cli.Check.Repo)
	default:
		return fmt.Errorf("unknown command: %s", ctx.Command())
	}
}

func handleList(client github.Client, days int) error {
	ctx := context.Background()
	fmt.Printf("Fetching repositories...\n")

	failures, err := client.ListAllFailedWorkflows(ctx, days)
	if err != nil {
		return fmt.Errorf("error listing failed workflows: %v", err)
	}

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

func handleCheck(client github.Client, repo, pr string) error {
	ctx := context.Background()
	return client.GetFailedWorkflows(ctx, pr, repo)
}
