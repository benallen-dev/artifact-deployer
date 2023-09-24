package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v55/github"
)

type DeployParameters struct {
	HeadSha   string
	Handshake string
}

func getDeployParameters(r *http.Request) (*DeployParameters, error) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var deployParams DeployParameters
	err = json.Unmarshal(body, &deployParams)
	if err != nil {
		return nil, err
	}

	return &deployParams, nil

}

func doDeploy(paramSha string) {
	
		waitTime := 10 * time.Second
		log.Println("Waiting", waitTime, "for artifact to become available...")
		time.Sleep(waitTime);

		// Create a github client
		ctx := context.Background()
		token := os.Getenv("GITHUB_PAT")
		artifactFilename := "/tmp/" + os.Getenv("TEMP_FILENAME")

		client := github.NewClient(nil).WithAuthToken(token)

		// Fetch the latest artifact
		newest, err := getNewestArtifact(ctx, client)
		if err != nil {
			log.Println("Error fetching newest artifact:", err)
			return
		}

		headSha := newest.GetWorkflowRun().GetHeadSHA()
		log.Println("Requested commit SHA:", paramSha)
		log.Println("Head commit SHA:     ", headSha)

		// Check if the SHA requested matches the newest artifact's commit SHA
		if headSha != paramSha {
			err = errors.New("Requested SHA does not match the latest artifact.")
			log.Println("Error:", err)
			return
		}

		// Prepare the deployment directory
		destination, err := getDeployDestination(headSha)
		if err != nil {
			log.Println("Error preparing deployment directory:", err)
			return
		}

		log.Println("Deploying to:", destination)

		// Download the artifact
		// What would be super cool is to keep the file in memory instead of writing to disk
		err = downloadArtifact(ctx, client, newest, artifactFilename)
		if err != nil {
			log.Println("Error downloading artifact:", err)
			return
		}

		// Unzip the artifact
		err = extractArtifact(artifactFilename, destination)
		if err != nil {
			log.Println("Error extracting artifact:", err)
			return
		}

		// Delete the archive
		err = os.Remove(artifactFilename)
		if err != nil {
			log.Println("Error deleting artifact:", err)
			return
		}

		// Update the symlink
		err = updateSymlink(destination)
		if err != nil {
			log.Println("Error updating symlink:", err)
			return
		}

		// Guess we're done here!
		log.Println("Sucessfully deployed")
}

func deploy(w http.ResponseWriter, r *http.Request) {

	if r.Method == "PUT" {

		log.Println("Received deployment request from " + r.RemoteAddr)

		// Parse the request body
		params, err := getDeployParameters(r)
		if err != nil {
			http400Error(w, err, "Error parsing request body: ")
			return
		}

		// Check if the handshake matches
		ok := checkHandshake(params.Handshake, params.HeadSha)
		if !ok {
			http403Error(w, err, "Incorrect handshake.")
			return
		}

		// At this point we know the request is valid, but the artifact isn't available on the
		// REST API yet. We need to wait for the workflow to finish before we can fetch it.
		// 
		// Tidy solution: use the GraphQL API to subscribe to workflow_run events and continue when 
		// the workflow is finished.
		// The downside to this is I built all of this lot around the REST API, and I want a PoC 
		// because I have about 15 minutes to implement this today.
		//
		// The hack-it-together solution I'm actually gonna do right now is "just spawn a new thread,
		// return HTTP 202 while we wait X seconds and hope the artifact is available by then."

		// Don't judge me ok I'm not proud of it either
		go doDeploy(params.HeadSha)

		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "202 - Accepted")
	} else {
		log.Println("Received "+r.Method+" for path:", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 - Bad Request")
	}
}
