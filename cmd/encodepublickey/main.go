package main

import (
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr/nip19"
)

func main() {
	pk := os.Args[1]
	fmt.Printf("Given hex public key: %s\n", pk)

	npub, err := nip19.EncodePublicKey(pk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("npub: %s\n", npub)
}
