package main

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
)

func getLatestArtifact() {
	// TODO: Refactor all these steps into functions
}

func http500Error(w http.ResponseWriter, err error, msg string) {
	log.Println(msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Internal Server Error"))
}

func deploy(w http.ResponseWriter, r *http.Request) {

	httpMethod := r.Method

	log.Printf("Received %s request", httpMethod)

	ctx := context.Background()
	token := "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"
	client := github.NewClient(nil).WithAuthToken(token)

	_, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		http500Error(w, err, "Error getting user: ")
		return
	}

	// If a Token Expiration has been set, it will be displayed.
	if !resp.TokenExpiration.IsZero() {
		log.Printf("Token Expiration: %v\n", resp.TokenExpiration)
	}

	// List artifacts for the website repo
	artifacts, resp, err := client.Actions.ListArtifacts(ctx, "benallen-dev", "benallen-dot-dev", nil)
	if err != nil {
		http500Error(w, err, "Error getting artifacts: ")
		return
	}

	// Get the newest artifact
	sort.Slice(artifacts.Artifacts, func(i, j int) bool {
		return artifacts.Artifacts[i].CreatedAt.After(*artifacts.Artifacts[j].CreatedAt.GetTime())
	})

	newest := artifacts.Artifacts[0]

	headSha := newest.GetWorkflowRun().HeadSHA

	log.Println("Head SHA: ", github.Stringify(headSha))

	if httpMethod == "GET" {
		fmt.Fprintf(w, "Newest artifact SHA: %v\n\n", github.Stringify(headSha))
		fmt.Fprintf(w, "SHA requested for deployment: %v\n\n", r.URL.Query().Get("sha"))
		return
	}

	if httpMethod == "POST" {
		// meh
	}

	if httpMethod == "PUT" {
		// here we gooooo
		// Download the artifact
		url, _, err := client.Actions.DownloadArtifact(ctx, "benallen-dev", "benallen-dot-dev", newest.GetID(), true)
		if err != nil {
			http500Error(w, err, "Error downloading artifact: ")
			return
		}

		log.Println("Artifact download URL: ", url)

		fileContent, err := client.Client().Get(url.String())
		if err != nil {
			http500Error(w, err, "Error downloading artifact: ")
			return
		}
		defer fileContent.Body.Close()

		// Create the artifact file
		artifactFilename := "artifact.zip"

		file, err := os.Create(artifactFilename)
		if err != nil {
			http500Error(w, err, "Error creating artifact file: ")
			return
		}

		size, err := io.Copy(file, fileContent.Body)
		if err != nil {
			http500Error(w, err, "Error writing artifact file: ")
			return
		}

		defer file.Close()

		log.Println("Artifact file size: ", size)

		// If it already exists, rename the existing directory
		homedir, err := os.UserHomeDir()
		if err != nil {
			http500Error(w, err, "Error getting user home directory: ")
			return
		}
		dst := homedir + "/www/benallen.dev"

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
		archive, err := zip.OpenReader(artifactFilename)
		if err != nil {
			http500Error(w, err, "Error opening artifact file: ")
			panic(err)
		}
		defer archive.Close()

		for _, f := range archive.File {
			filePath := filepath.Join(dst, f.Name)
			fmt.Println("unzipping file ", filePath)

			if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
				http500Error(w, err, "Error extracting file: ")
				return
			}
			if f.FileInfo().IsDir() {
				fmt.Println("creating directory...")
				os.MkdirAll(filePath, os.ModePerm)
				continue
			}

			if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				http500Error(w, err, "Error creating directory: ")
				panic(err)
			}

			dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				http500Error(w, err, "Error opening file on disk: ")
				panic(err)
			}

			fileInArchive, err := f.Open()
			if err != nil {
				http500Error(w, err, "Error opening file in archive: ")
				panic(err)
			}

			if _, err := io.Copy(dstFile, fileInArchive); err != nil {
				http500Error(w, err, "Error copying file: ")
				panic(err)
			}

			dstFile.Close()
			fileInArchive.Close()
		}

		log.Println("Unzipped artifact")

		// Delete the archive
		err = os.Remove(artifactFilename)
		if err != nil {
			http500Error(w, err, "Error deleting artifact file: ")
			return
		}

		// Guess we're done here!
		// Oh, maybe respond to the HTTP request
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Sucessfully deployed"))
	}
}

func hello(w http.ResponseWriter, r *http.Request) {

	log.Println("Received request for path: ", r.URL.Path)
	fmt.Fprintf(w, "Hello World!")

}

func main() {

	log.SetPrefix("[AD] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmsgprefix)

	// Create http client
	// Register routes
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/deploy", deploy)

	// Start the server
	log.Println("Starting server on port 8080")
	err := http.ListenAndServe(":8080", nil)
	log.Fatal(err)

}
