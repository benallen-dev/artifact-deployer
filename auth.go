package main

import (
	"crypto/sha1"
	"log"
	"os"
)

func checkHandshake(handshake string, headSha string) bool {
	secret := os.Getenv("DEPLOY_SECRET")

	// Concatenate the secret with the head SHA
	hashThis := secret + headSha

	// Hash the concatenated string
	hasher := sha1.New()
	hasher.Write([]byte(hashThis))
	ourHash := hasher.Sum(nil)

	// Compare the hash with the handshake
	log.Println("Our hash: ", ourHash)
	log.Println("Their hash: ", []byte(handshake))


	return false
}
