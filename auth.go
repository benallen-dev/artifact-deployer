package main

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"os"
)

func checkHandshake(handshake string, headSha string) bool {
	secret := os.Getenv("DEPLOY_SECRET")

	// Concatenate the secret with the head SHA
	concatted := secret + headSha
	hashThis := []byte(concatted)

	// Hash the concatenated string
	hasher := sha1.New()
	hasher.Write(hashThis)
	ourHash := hex.EncodeToString(hasher.Sum(nil))

	// Compare the hash with the handshake
	log.Println("Our hash:  ", ourHash)
	log.Println("Their hash:", handshake)

	return ourHash == handshake
}
