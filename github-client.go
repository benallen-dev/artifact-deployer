package main

import (
	"context"
	"io"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/google/go-github/v55/github"
)

func getNewestArtifact(ctx context.Context, client *github.Client) (*github.Artifact, error) {

	githubUser := os.Getenv("GITHUB_USER")
	githubRepo := os.Getenv("GITHUB_REPO")

	// List artifacts for the website repo
	artifacts, _, err := client.Actions.ListArtifacts(ctx, githubUser, githubRepo, nil)
	if err != nil {
		return nil, err
	}

	// Get the newest artifact
	sort.Slice(artifacts.Artifacts, func(i, j int) bool {
		return artifacts.Artifacts[i].CreatedAt.After(*artifacts.Artifacts[j].CreatedAt.GetTime())
	})

	newest := artifacts.Artifacts[0]

	return newest, nil
}

func downloadArtifact(ctx context.Context, client *github.Client, artifact *github.Artifact, artifactFilename string) error {

	githubUser := os.Getenv("GITHUB_USER")
	githubRepo := os.Getenv("GITHUB_REPO")

	url, _, err := client.Actions.DownloadArtifact(ctx, githubUser, githubRepo, artifact.GetID(), true)
	if err != nil {
		return err
	}

	// client.Client is the underlying http.Client used by the github client
	fileContent, err := client.Client().Get(url.String())
	if err != nil {
		return err
	}

	// Create the artifact file
	file, err := os.Create(artifactFilename)
	if err != nil {
		return err
	}

	size, err := io.Copy(file, fileContent.Body)
	if err != nil {
		return err
	}

	// Log the size of the artifact
	if size < 1024 {
		log.Println(artifactFilename + ": " + strconv.FormatInt(size, 10) + " B")
	} else if size < 1024*1024 {
		filesizeKb := strconv.FormatFloat(float64(size)/1024.0, 'f', 2, 64)
		log.Println(artifactFilename + ": " + filesizeKb + " kB")
	} else {
		filesizeMb := strconv.FormatFloat(float64(size)/1024.0/1024.0, 'f', 2, 64)
		log.Println(artifactFilename + ": " + filesizeMb + " MB")
	}

	// Close the HTTP response
	err = fileContent.Body.Close()
	if err != nil {
		return err
	}

	// Close the file we just wrote
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
