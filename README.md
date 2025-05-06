# GitHub Actions Checker

A CLI tool to monitor and summarize failed GitHub Actions workflows across your organization's repositories.

## Features

- List all failed workflow runs across all repositories in your organization
- Check failed workflows for a specific pull request
- Filter results by time period
- Group and summarize failures by PR
- Direct links to failed workflows and PRs

## Prerequisites

- Go 1.16 or higher
- GitHub Personal Access Token with the following permissions:
  - `repo` (Full control of private repositories)
  - `workflow` (Update GitHub Action workflows)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd <repository-name>
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the executable:
```bash
go build -o gh-actions-checker
```

4. Create a `.env` file in the project root with your GitHub credentials:
```env
GITHUB_TOKEN=your_github_token_here
GITHUB_OWNER=your_organization_name
```

## Usage

### List All Failed Workflows

To list all failed workflows across all repositories in your organization:

```bash
./gh-actions-checker list
```

By default, this will show failures from the last 7 days. You can specify a different number of days using the `-d` flag:

```bash
./gh-actions-checker list -d 30  # Show failures from last 30 days
```

### Check Specific PR

To check failed workflows for a specific pull request:

```bash
./gh-actions-checker check -r repository-name -p pr-number
```

For example:
```bash
./gh-actions-checker check -r my-repo -p 123
```

You can also specify the repository in the format `owner/repo`:
```bash
./gh-actions-checker check -r owner/my-repo -p 123
```

## Output Format

The tool provides a summary of failed workflows, including:

- PR URL (clickable link)
- Repository name
- Latest failure workflow name and timestamp
- Total number of failures for that PR

## Security Notes

- Never commit your `.env` file containing the GitHub token
- The `.gitignore` file is configured to exclude the `.env` file
- Keep your GitHub token secure and rotate it periodically

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 