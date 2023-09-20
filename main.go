package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

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
