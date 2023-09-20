package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/joho/godotenv"
)


func deploy(w http.ResponseWriter, r *http.Request) {

	httpMethod := r.Method
	ctx := context.Background()
	token := os.Getenv("GITHUB_PAT")

	log.Printf("Received %s request", httpMethod)

	client := github.NewClient(nil).WithAuthToken(token)

	newest, err := getNewestArtifact(ctx, client)
	if err != nil {
		http500Error(w, err, "")
		return
	}

	headSha := newest.GetWorkflowRun().HeadSHA
	log.Println("Head SHA: ", github.Stringify(headSha))

	if httpMethod == "GET" {
		fmt.Fprintf(w, "Newest artifact SHA: %v\n\n", github.Stringify(headSha))
		fmt.Fprintf(w, "SHA requested for deployment: %v\n\n", r.URL.Query().Get("sha"))
		return
	}

	if httpMethod == "PUT" {
		artifactFilename := "artifact.zip"

		// Download the artifact
		err = downloadArtifact(ctx, client, newest, artifactFilename)
		if err != nil {
			http500Error(w, err, "Error downloading artifact: ")
			return
		}

		// If it already exists, rename the existing directory
		homedir, err := os.UserHomeDir()
		if err != nil {
			http500Error(w, err, "Error getting user home directory: ")
			return
		}
		
		dst := homedir + "/www/" + os.Getenv("SITE_DIR")

		if _, err := os.Stat(dst); err == nil {
			log.Println("Destination directory already exists, renaming...")
			// get current datetime
			// rename existing directory to include datetime
			datetime := time.Now().Format("2006-01-02--15-04-05")

			err = os.Rename(dst, dst+"--"+datetime)
			if err != nil {
				http500Error(w, err, "Error renaming existing directory: ")
				return
			}
		}

		// Unzip the artifact
		err = extractArtifact(artifactFilename, dst)
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
		w.Write([]byte("Sucessfully deployed"))
	}
}

func hello(w http.ResponseWriter, r *http.Request) {

	log.Println("Received request for path: ", r.URL.Path)
	fmt.Fprintf(w, "Hello World!")

}

func main() {

	// Set up logging
	log.SetPrefix("[AD] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)

	// Load environment variables
	err := godotenv.Load()

	welcomeMsg := os.Getenv("WELCOME_MSG")
	log.Println(welcomeMsg)

	// Register routes
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/deploy", deploy)

	// Start the server
	log.Println("Starting server on port 8080")
	err = http.ListenAndServe(":8080", nil)
	log.Fatal(err)

}
