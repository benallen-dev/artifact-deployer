package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	
	"github.com/google/go-github/v55/github"
)

func deploy(w http.ResponseWriter, r *http.Request) {

	if r.Method == "PUT" {

		ctx := context.Background()
		token := os.Getenv("GITHUB_PAT")
		artifactFilename := os.Getenv("TEMP_FILENAME")

		client := github.NewClient(nil).WithAuthToken(token)

		newest, err := getNewestArtifact(ctx, client)
		if err != nil {
			http500Error(w, err, "")
			return
		}

		headSha := newest.GetWorkflowRun().HeadSHA
		log.Println("Head SHA: ", github.Stringify(headSha))

		destination, err := getDeployDestination()
		if err != nil {
			http500Error(w, err, "Error preparing deployment: ")
			return
		}

		// Download the artifact
		// What would be super cool is to keep the file in memory instead of writing to disk
		err = downloadArtifact(ctx, client, newest, artifactFilename)
		if err != nil {
			http500Error(w, err, "Error downloading artifact: ")
			return
		}

		// Unzip the artifact
		err = extractArtifact(artifactFilename, destination)
		if err != nil {
			http500Error(w, err, "Error extracting artifact: ")
			return
		}

		// Delete the archive
		err = os.Remove(artifactFilename)
		if err != nil {
			http500Error(w, err, "Error deleting artifact file: ")
			return
		}

		// Guess we're done here!
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Sucessfully deployed")
		log.Println("Sucessfully deployed")
	} else {
		log.Println("Received " + r.Method + " for path:", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 - Bad Request")
	}
}
