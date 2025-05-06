package main

import (
	"log"
	"os"

	"github.com/kjkondratuk/gh-workflow-monitor/internal/config"
	"github.com/kjkondratuk/gh-workflow-monitor/pkg/cli"
	"github.com/kjkondratuk/gh-workflow-monitor/pkg/github"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	client := github.NewClient(cfg.GitHubToken, cfg.GitHubOwner)

	if err := cli.Run(client); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
