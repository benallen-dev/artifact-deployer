package main

import (
	"archive/zip"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getBaseDir() (string, error) {
	
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homedir + "/www/" + os.Getenv("SITE_DIR"), nil
}

func getDeployDestination(headSha string) (string, error) {

	baseDir, err := getBaseDir()
	if err != nil {
		return "", err
	}

	// If it already exists, rename the existing directory
	dst := baseDir + "-" + headSha[0:12]

	if _, err := os.Stat(dst); err == nil {
		// Shit son it already exists
		return "", errors.New("Commit is already deployed")
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

func updateSymlink(dst string) error {

	baseDir, err := getBaseDir()
	if err != nil {
		return err
	}

	// Remove the existing symlink
	err = os.Remove(baseDir)
	if err != nil && !os.IsNotExist(err) {
		// If the error is not "file does not exist", it's an actual error
		return err
	}

	// Create a new symlink
	err = os.Symlink(dst, baseDir)
	if err != nil {
		return err
	}

	log.Println("Updated symlink")

	// No errors, no problem
	return nil
}
