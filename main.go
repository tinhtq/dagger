// package main

// import (
// 	"context"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"

// 	"dagger.io/dagger"
// )

// type FastapiDagger struct{}

// func main() {
// 	ctx := context.Background()
// 	f := &FastapiDagger{}

// 	// Initialize Dagger client
// 	client, err := dagger.Connect(ctx)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to Dagger client: %v", err)
// 	}
// 	defer client.Close()

// 	source := client.Host().Directory(".") // Current working directory

// 	// Environment variables
// 	registry := "ghcr.io"
// 	imageName := "example-image"
// 	githubRepo := os.Getenv("GITHUB_REPOSITORY")
// 	githubToken := os.Getenv("GITHUB_TOKEN")
// 	prNumber := os.Getenv("PR_NUMBER")

// 	if githubRepo == "" || githubToken == "" || prNumber == "" {
// 		log.Println("Please set the environment variables: GITHUB_REPOSITORY, GITHUB_TOKEN, PR_NUMBER")
// 		return
// 	}

// 	fmt.Println("Starting Build and Push...")
// 	result, err := f.BuildAndPush(ctx, registry, imageName, source)
// 	if err != nil {
// 		log.Printf("Build and Push Error: %s\n", err)
// 	} else {
// 		fmt.Println(result)
// 	}

// 	fmt.Println("Starting Scan and PR...")
// 	scanResult, err := f.ScanAndPR(ctx, prNumber, githubRepo, githubToken, source)
// 	if err != nil {
// 		log.Printf("Scan and PR Error: %s\n", err)
// 	} else {
// 		fmt.Println(scanResult)
// 	}
// }

// File: ci/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"github.com/yourusername/yourrepo/daggerclient" // Update this import path
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// Initialize the Dagger client
	client, err := dagger.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Dagger engine: %w", err)
	}
	defer client.Close()

	// Initialize our Dagger client wrapper
	fd := daggerclient.NewFastapiDagger(client)

	// Get environment variables
	githubToken := os.Getenv("GITHUB_TOKEN")
	registry := os.Getenv("DOCKER_REGISTRY")
	imageName := os.Getenv("DOCKER_IMAGE_NAME")
	
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create a Directory object from the current directory
	source := client.Host().Directory(wd)

	// If this is a PR, get the PR number from GITHUB_REF
	prNumber := ""
	githubRef := os.Getenv("GITHUB_REF")
	if strings.HasPrefix(githubRef, "refs/pull/") {
		parts := strings.Split(githubRef, "/")
		if len(parts) >= 4 {
			prNumber = parts[2]
		}
	}

	// Run code scanning if this is a PR
	if prNumber != "" {
		log.Println("Running code scan...")
		result, err := fd.ScanAndPR(
			ctx,
			prNumber,
			imageName, // This will be in the format owner/repo
			githubToken,
			source,
		)
		if err != nil {
			return fmt.Errorf("failed to run scan: %w", err)
		}
		log.Println(result)
	}

	// Build and push the container image
	log.Println("Building and pushing container...")
	result, err := fd.BuildAndPush(
		ctx,
		registry,
		imageName,
		source,
	)
	if err != nil {
		return fmt.Errorf("failed to build and push: %w", err)
	}
	log.Println(result)

	return nil
}