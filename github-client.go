package main

import (
	"context"
	"github.com/google/go-github/v55/github"
	"io"
	"log"
	"os"
	"sort"
)

func getNewestArtifact(ctx context.Context, client *github.Client) (*github.Artifact, error) {

	// List artifacts for the website repo
	artifacts, _, err := client.Actions.ListArtifacts(ctx, "benallen-dev", "benallen-dot-dev", nil)
	if err != nil {
		return nil, err
	}

	// Get the newest artifact
	sort.Slice(artifacts.Artifacts, func(i, j int) bool {
		return artifacts.Artifacts[i].CreatedAt.After(*artifacts.Artifacts[j].CreatedAt.GetTime())
	})

	newest := artifacts.Artifacts[0]

	headSha := newest.GetWorkflowRun().HeadSHA

	log.Println("Head SHA: ", github.Stringify(headSha))

	return newest, nil
}

func downloadArtifact(ctx context.Context, client *github.Client, artifact *github.Artifact, artifactFilename string) error {

	githubUser := os.Getenv("GITHUB_USER")
	githubRepo := os.Getenv("GITHUB_REPO")

	url, _, err := client.Actions.DownloadArtifact(ctx, githubUser, githubRepo, artifact.GetID(), true)
	if err != nil {
		return err
	}

	log.Println("Artifact download URL: ", url)

	// client.Client is the underlying http.Client used by the github client
	fileContent, err := client.Client().Get(url.String())
	if err != nil {
		return err
	}
	// Maybe we can just do this manually at the end of the function?
	defer fileContent.Body.Close()

	// Create the artifact file
	file, err := os.Create(artifactFilename)
	if err != nil {
		return err
	}

	size, err := io.Copy(file, fileContent.Body)
	if err != nil {
		return err
	}

	defer file.Close()

	log.Println(artifactFilename, ": ", size)

	return nil
}
