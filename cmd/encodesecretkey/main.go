package main

import (
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

func main() {
	sk := os.Args[1]
	fmt.Printf("Given hex secret key: %s\n", sk)

	nsec, err := nip19.EncodePrivateKey(sk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("nsec: %s\n", nsec)

	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("pk: %s\n", pk)

	npub, err := nip19.EncodePublicKey(pk)
	if err != nil {
		panic(err)
	}
	fmt.Printf("npub: %s\n", npub)
}
