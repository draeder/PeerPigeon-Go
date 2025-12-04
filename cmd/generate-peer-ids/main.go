package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
)

func generatePeerID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", b)
}

func main() {
	count := flag.Int("n", 1, "number of peer IDs to generate")
	flag.Parse()

	fmt.Printf("🎲 Generating %d random peer ID(s):\n\n", *count)

	urls := []string{
		"ws://pigeonhub-b.fly.dev/ws",
		"wss://pigeonhub-b.fly.dev/ws",
		"ws://pigeonhub-c.fly.dev/ws",
		"wss://pigeonhub-c.fly.dev/ws",
	}

	for i := 0; i < *count; i++ {
		peerId := generatePeerID()
		fmt.Printf("%d. %s\n", i+1, peerId)
		for _, url := range urls {
			fmt.Printf("   %s?peerId=%s\n", url, peerId)
		}
		fmt.Println()
	}

	fmt.Printf("✅ %d peer ID(s) generated.\n\n", *count)
	fmt.Println("Usage:")
	fmt.Println("  go run cmd/generate-peer-ids/main.go -n 5")
	fmt.Println("  ./generate-peer-ids -n 10")
}
