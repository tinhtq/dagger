package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/dagger/dagger/client"
	"github.com/go-resty/resty/v2"
)

type FastapiDagger struct{}

func (f *FastapiDagger) scanAndPR(
	ctx context.Context, 
	pullRequestNumber string, 
	githubRepo string, 
	githubToken string, 
	source *client.Directory,
) (string, error) {
	// Create Dagger client

	cl, err := client.New(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create dagger client: %v", err)
	}
	defer cl.Close()

	// Create container and run scan
	container := cl.Container().
		From("python:3.10").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"pip", "install", "-r", "requirements.txt"}).
		WithExec([]string{"sh", "-c", "flake8 app || true"})

	// Capture stdout
	stdout, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get stdout: %v", err)
	}

	// Prepare GitHub comment
	commentBody := map[string]string{
		"body": fmt.Sprintf("## Scan Results\n\n```\n%s\n```", stdout),
	}

	// Create HTTP client and post comment
	httpClient := resty.New()
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", githubRepo, pullRequestNumber)
	
	resp, err := httpClient.R().
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", githubToken)).
		SetBody(commentBody).
		Post(commentURL)

	if err != nil {
		return "", fmt.Errorf("failed to post comment: %v", err)
	}

	if resp.StatusCode() == 201 {
		return "Comment posted successfully!", nil
	}

	return fmt.Sprintf("Failed to post comment: %s", resp.String()), nil
}