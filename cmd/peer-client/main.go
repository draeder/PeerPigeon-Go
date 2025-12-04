package main

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type AnnounceData struct {
	PeerID string `json:"peerId"`
	Info   string `json:"info,omitempty"`
}

type PeerDiscoveredData struct {
	PeerID string `json:"peerId"`
	Info   string `json:"info,omitempty"`
}

func generatePeerID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", b)
}

func main() {
	hubURL := flag.String("hub", "ws://localhost:3000", "hub URL (ws://pigeonhub-b.fly.dev or wss://pigeonhub-b.fly.dev)")
	name := flag.String("name", "peer-client", "peer name for logging")
	listenTime := flag.Duration("listen", 5*time.Second, "how long to listen for peer discoveries")
	flag.Parse()

	peerId := generatePeerID()
	fmt.Printf("[%s] Generated peer ID: %s\n", *name, peerId)

	// Build connection URL with peerId query parameter
	u, err := url.Parse(*hubURL)
	if err != nil {
		log.Fatalf("Invalid hub URL: %v", err)
	}
	q := u.Query()
	q.Set("peerId", peerId)
	u.RawQuery = q.Encode()

	fmt.Printf("[%s] Connecting to hub: %s\n", *name, u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("[%s] Connection failed: %v", *name, err)
	}
	defer ws.Close()

	fmt.Printf("[%s] ✅ Connected to hub\n", *name)

	// Announce ourselves
	announceMsg := Message{
		Type: "announce",
		Data: func() json.RawMessage {
			data := AnnounceData{
				PeerID: peerId,
				Info:   *name,
			}
			b, _ := json.Marshal(data)
			return b
		}(),
	}

	if err := ws.WriteJSON(announceMsg); err != nil {
		log.Fatalf("[%s] Announce failed: %v", *name, err)
	}
	fmt.Printf("[%s] 📢 Announced self\n", *name)

	// Listen for peer discoveries
	done := make(chan struct{})
	discoveredPeers := make(map[string]bool)

	go func() {
		defer close(done)
		for {
			var msg Message
			if err := ws.ReadJSON(&msg); err != nil {
				fmt.Printf("[%s] Read error: %v\n", *name, err)
				return
			}

			switch msg.Type {
			case "peer-discovered":
				var data PeerDiscoveredData
				if err := json.Unmarshal(msg.Data, &data); err != nil {
					fmt.Printf("[%s] Unmarshal error: %v\n", *name, err)
					continue
				}
				if data.PeerID != peerId && !discoveredPeers[data.PeerID] {
					discoveredPeers[data.PeerID] = true
					fmt.Printf("[%s] 🔍 Discovered peer: %s (info: %s)\n", *name, data.PeerID[:8], data.Info)
				}
			case "hub-stats":
				fmt.Printf("[%s] Hub stats received\n", *name)
			default:
				// Silently ignore other message types
			}
		}
	}()

	// Wait for discoveries or timeout
	select {
	case <-time.After(*listenTime):
		fmt.Printf("[%s] Timeout - stopping listen\n", *name)
	case <-done:
	}

	if len(discoveredPeers) > 0 {
		fmt.Printf("[%s] 📊 Total peers discovered: %d\n", *name, len(discoveredPeers))
	} else {
		fmt.Printf("[%s] ⚠️  No peers discovered (this is OK if you're the first peer)\n", *name)
	}

	fmt.Printf("[%s] Disconnecting...\n", *name)
}
