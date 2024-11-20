package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"net/http"

	"dagger.io/dagger"
)

type FastapiDagger struct{}

// ContainerEcho returns a container that echoes the provided string argument.
func (f *FastapiDagger) ContainerEcho(ctx context.Context, stringArg string) (*dagger.Container, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Dagger client: %w", err)
	}
	defer client.Close()

	container := client.Container().
		From("alpine:latest").
		WithExec([]string{"echo", stringArg})

	return container, nil
}

// GrepDir returns lines that match a pattern in the files of the provided directory.
func (f *FastapiDagger) GrepDir(ctx context.Context, directory *dagger.Directory, pattern string) (string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Dagger client: %w", err)
	}
	defer client.Close()

	container := client.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directory).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."})

	stdout, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get stdout: %w", err)
	}

	return stdout, nil
}

// BuildAndPush builds a Docker image and pushes it to a container registry.
func (f *FastapiDagger) BuildAndPush(ctx context.Context, registry, imageName string, source *dagger.Directory) (string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Dagger client: %w", err)
	}
	defer client.Close()

	image := client.Container().
		Build(source).
		WithLabel("version", "1.0.0")

	imageURL := fmt.Sprintf("%s/%s:latest", registry, imageName)
	err = image.Publish(ctx, imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to publish image: %w", err)
	}

	return fmt.Sprintf("Image successfully pushed to %s", imageURL), nil
}

// ScanAndPR scans the application for issues and posts the results as a comment on a GitHub pull request.
func (f *FastapiDagger) ScanAndPR(ctx context.Context, pullRequestNumber, githubRepo, githubToken string, source *dagger.Directory) (string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return "", fmt.Errorf("failed to connect to Dagger client: %w", err)
	}
	defer client.Close()

	container := client.Container().
		From("python:3.10").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"pip", "install", "-r", "requirements.txt"}).
		WithExec([]string{"sh", "-c", "flake8 app || true"})

	// Fetch the results of the scan
	scanResults, err := container.Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get scan results: %w", err)
	}

	// Post scan results to GitHub PR as a comment
	commentBody := fmt.Sprintf("## Scan Results\n\n```\n%s\n```", scanResults)
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", githubRepo, pullRequestNumber)
	reqBody := fmt.Sprintf(`{"body": %q}`, commentBody)

	req, err := http.NewRequest("POST", commentURL, strings.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	req.Header.Set("Content-Type", "application/json")

	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to post comment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return "Comment posted successfully!", nil
	}

	body, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("failed to post comment: %s", string(body))
}

func main() {
	ctx := context.Background()
	f := &FastapiDagger{}

	// Example usage (replace with actual values)
	registry := "ghcr.io"
	imageName := "example-image"
	source := dagger.Host().Directory(".") // Get the current working directory

	githubRepo := os.Getenv("GITHUB_REPO")
	githubToken := os.Getenv("GITHUB_TOKEN")
	prNumber := os.Getenv("PR_NUMBER")

	if githubRepo == "" || githubToken == "" || prNumber == "" {
		fmt.Println("Please set the environment variables: GITHUB_REPO, GITHUB_TOKEN, PR_NUMBER")
		return
	}

	// Build and Push Example
	fmt.Println("Starting Build and Push...")
	result, err := f.BuildAndPush(ctx, registry, imageName, source)
	if err != nil {
		fmt.Printf("Build and Push Error: %s\n", err)
	} else {
		fmt.Println(result)
	}

	// Scan and PR Example
	fmt.Println("Starting Scan and PR...")
	scanResult, err := f.ScanAndPR(ctx, prNumber, githubRepo, githubToken, source)
	if err != nil {
		fmt.Printf("Scan and PR Error: %s\n", err)
	} else {
		fmt.Println(scanResult)
	}
}
