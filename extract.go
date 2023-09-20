package main

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func getDeployDestination() (string, error) {

	// If it already exists, rename the existing directory
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dst := homedir + "/www/" + os.Getenv("SITE_DIR")

	if _, err := os.Stat(dst); err == nil {
		log.Println("Destination directory already exists, renaming existing dir with current time...")
		// get current datetime
		// rename existing directory to include datetime
		datetime := time.Now().Format("2006-01-02--15-04-05")

		err = os.Rename(dst, dst+"--"+datetime)
		if err != nil {
			return "", err
		}
	}

	return dst, nil
}

func extractArtifact(artifactFilename string, dst string) error {

	archive, err := zip.OpenReader(artifactFilename)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(dst, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			return errors.New("illegal file path: " + filePath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return err
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	log.Println("Unzipped artifact")

	// No errors, no problem
	return nil
}
