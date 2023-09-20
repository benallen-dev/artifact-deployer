package main

import (
	"log"
	"net/http"
)

func http500Error(w http.ResponseWriter, err error, msg string) {
	log.Println(msg, err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Internal Server Error"))
}
