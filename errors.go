package main

import (
	"log"
	"net/http"
)

// Used when the request body can't be parsed into DeployParameters
func http400Error(w http.ResponseWriter, err error, msg string) {
	log.Println("[HTTP 400]", msg, err)
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("400 - Bad Request"))
}

// Used when the handshake fails
func http401Error(w http.ResponseWriter, err error, msg string) {
	log.Println("[HTTP 401]", msg, err)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("401 - Unauthorized"))
}

// Used when attmepting to to deploy a commit SHA that doesn't match the latest artifact
func http409Error(w http.ResponseWriter, err error, msg string) {
	log.Println("[HTTP 409]", msg, err)
	w.WriteHeader(http.StatusConflict)
	w.Write([]byte("409 - Conflict"))
}

// Used for any error where we want to be opaque about what went wrong
func http500Error(w http.ResponseWriter, err error, msg string) {
	log.Println("[HTTP 500]", msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Internal Server Error"))
}

