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

func deploy(w http.ResponseWriter, r *http.Request) {

	if r.Method == "PUT" {

		log.Println("Received deployment request from " + r.RemoteAddr)

		// Parse the request body
		params, err := getDeployParameters(r)
		if err != nil {
			http400Error(w, err, "Error parsing request body: ")
			return
		}

		// Create a github client
		ctx := context.Background()
		token := os.Getenv("GITHUB_PAT")
		artifactFilename := "/tmp/" + os.Getenv("TEMP_FILENAME")

		client := github.NewClient(nil).WithAuthToken(token)

		// Fetch the latest artifact
		newest, err := getNewestArtifact(ctx, client)
		if err != nil {
			http500Error(w, err, "")
			return
		}

		headSha := newest.GetWorkflowRun().GetHeadSHA()
		log.Println("Requested commit SHA:", params.HeadSha)
		log.Println("Head commit SHA:     ", headSha)

		// Check if the SHA requested matches the newest artifact's commit SHA
		if headSha != params.HeadSha {
			err = errors.New("Requested SHA does not match the latest artifact.")
			http409Error(w, err, "")
			return
		}

		// Check if the handshake matches
		ok := checkHandshake(params.Handshake, headSha)
		if !ok {
			http403Error(w, err, "Incorrect handshake.")
			return
		}

		// Prepare the deployment directory
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
		log.Println("Received "+r.Method+" for path:", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "400 - Bad Request")
	}
}
