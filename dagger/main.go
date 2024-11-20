// A generated module for FastapiDagger functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"dagger/fastapi-dagger/internal/dagger"
	"net/http"
)

type FastapiDagger struct{}

// ContainerEcho returns a container that echoes the provided string argument.
func (f *FastapiDagger) ContainerEcho(ctx context.Context, stringArg string) (*dagger.Container, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(io.Discard))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	container := client.Container().
		From("alpine:latest").
		WithExec([]string{"echo", stringArg})
	return container, nil
}

// GrepDir returns lines that match a pattern in the files of the provided directory.
func (f *FastapiDagger) GrepDir(ctx context.Context, directory *dagger.Directory, pattern string) (string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(io.Discard))
	if err != nil {
		return "", err
	}
	defer client.Close()

	container := client.Container().
		From("alpine:latest").
		WithMountedDirectory("/mnt", directory).
		WithWorkdir("/mnt").
		WithExec([]string{"grep", "-R", pattern, "."})

	stdout, err := container.Stdout(ctx)
	if err != nil {
		return "", err
	}

	return stdout, nil
}

// BuildAndPush builds a Docker image and pushes it to a container registry.
func (f *FastapiDagger) BuildAndPush(ctx context.Context, registry, imageName string, source *dagger.Directory) (string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(io.Discard))
	if err != nil {
		return "", err
	}
	defer client.Close()

	image := client.Container().
		Build(source).
		WithLabel("version", "1.0.0")

	imageURL := fmt.Sprintf("%s/%s:latest", registry, imageName)
	err = image.Publish(ctx, imageURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Image successfully pushed to %s", imageURL), nil
}

// ScanAndPR scans the application for issues and posts the results as a comment on a GitHub pull request.
func (f *FastapiDagger) ScanAndPR(ctx context.Context, pullRequestNumber, githubRepo, githubToken string, source *dagger.Directory) (string, error) {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(io.Discard))
	if err != nil {
		return "", err
	}
	defer client.Close()

	container := client.Container().
		From("python:3.10").
		WithMountedDirectory("/src", source).
		WithWorkdir("/src").
		WithExec([]string{"pip", "install", "-r", "requirements.txt"}).
		WithExec([]string{"sh", "-c", "flake8 app || true"})

	scanResults, err := container.Stdout(ctx)
	if err != nil {
		return "", err
	}

	commentBody := fmt.Sprintf("## Scan Results\n\n```\n%s\n```", scanResults)
	commentURL := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s/comments", githubRepo, pullRequestNumber)
	reqBody := fmt.Sprintf(`{"body": %q}`, commentBody)

	req, err := http.NewRequest("POST", commentURL, strings.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", githubToken))
	req.Header.Set("Content-Type", "application/json")

	clientHTTP := &http.Client{}
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return "Comment posted successfully!", nil
	}

	body, _ := io.ReadAll(resp.Body)
	return "", fmt.Errorf("failed to post comment: %s", string(body))
}
