// A Dagger module to say hello to the world
package main

import (
	"context"
	"fmt"
	"hello/internal/dagger"
	// "strings"

	"github.com/go-resty/resty/v2"
)

// var defaultFigletContainer = dag.
// 	Container().
// 	From("alpine:latest").
// 	WithExec([]string{
// 		"apk", "add", "figlet",
// 	})

// A Dagger module to say hello to the world!
type Fastapi struct{}

// Say hello to the world!
// func (fastapi *Fastapi) Scan(ctx context.Context,
// 	greeting string,
// 	name string,
// 	giant bool,
// 	shout bool,
// 	source *dagger.Directory
// ) (string, error) {
// 	container := source.
// 	WithExec([]string{"pip", "install", "-r", "requirements.txt"}).
// 	WithExec([]string{"sh", "-c", "flake8 app || true"})

// 	stdout, err := container.Stdout(ctx)
// 	if err != nil {
// 		return "", fmt.Errorf("scan failed: %v", err)
// 	}

// 	return stdout, nil
// }

func (m *Fastapi) Scan(
	ctx context.Context,
	pullRequestNumber string,
	githubRepo string,
	githubToken string,
	// prn string,
	// githubrepo string,
	// githubtoken string,
	source *dagger.Directory,
) (string, error) {
	// Create Dagger client
	cl := dagger.Connect()
	// if err != nil {
	// 	return "", fmt.Errorf("failed to create dagger client: %v", err)
	// }
	// defer cl.Close()

	// Create container and run scan
	container := cl.Container().
		From("python:3.10").
		WithDirectory("/src", source).
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