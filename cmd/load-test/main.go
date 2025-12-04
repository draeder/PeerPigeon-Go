package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type LoadTestMetrics struct {
	ConnectedPeers   int64
	FailedConnects   int64
	PeersDiscovered  int64
	MessagesReceived int64
	StartTime        time.Time
}

func generatePeerID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", b)
}

func testPeer(hubUrl string, metrics *LoadTestMetrics, wg *sync.WaitGroup, testDuration time.Duration) {
	defer wg.Done()

	peerId := generatePeerID()
	u, err := url.Parse(hubUrl)
	if err != nil {
		atomic.AddInt64(&metrics.FailedConnects, 1)
		return
	}

	q := u.Query()
	q.Set("peerId", peerId)
	u.RawQuery = q.Encode()

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		atomic.AddInt64(&metrics.FailedConnects, 1)
		return
	}
	defer ws.Close()

	atomic.AddInt64(&metrics.ConnectedPeers, 1)

	// Announce self
	announceMsg := map[string]interface{}{
		"type": "announce",
		"data": map[string]interface{}{
			"peerId": peerId,
		},
	}
	ws.WriteJSON(announceMsg)

	// Listen for messages in background
	done := make(chan struct{})
	go func() {
		for {
			var msg map[string]interface{}
			if err := ws.ReadJSON(&msg); err != nil {
				close(done)
				return
			}
			if msgType, ok := msg["type"].(string); ok && msgType == "peer-discovered" {
				atomic.AddInt64(&metrics.PeersDiscovered, 1)
			}
			atomic.AddInt64(&metrics.MessagesReceived, 1)
		}
	}()

	// Keep connection open for test duration
	select {
	case <-time.After(testDuration):
		ws.Close()
	case <-done:
	}
}

func main() {
	hubUrl := flag.String("hub", "ws://localhost:8080", "hub URL")
	numPeers := flag.Int("peers", 100, "number of peers to simulate")
	testDurationSeconds := flag.Int("duration", 30, "test duration in seconds")
	printInterval := flag.Int("interval", 5, "metrics print interval in seconds")
	flag.Parse()

	fmt.Printf("🚀 Load Testing PeerPigeon Hub\n")
	fmt.Printf("================================\n")
	fmt.Printf("Hub URL: %s\n", *hubUrl)
	fmt.Printf("Peers: %d\n", *numPeers)
	fmt.Printf("Duration: %d seconds\n", *testDurationSeconds)
	fmt.Printf("\n")

	metrics := &LoadTestMetrics{
		StartTime: time.Now(),
	}

	testDuration := time.Duration(*testDurationSeconds) * time.Second
	var wg sync.WaitGroup

	// Start all peers
	startTime := time.Now()
	for i := 0; i < *numPeers; i++ {
		wg.Add(1)
		go testPeer(*hubUrl, metrics, &wg, testDuration)

		// Stagger peer connections
		time.Sleep(time.Duration(*testDurationSeconds) * time.Millisecond / time.Duration(*numPeers))
	}

	// Print metrics periodically
	ticker := time.NewTicker(time.Duration(*printInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime)
			fmt.Printf("\n📊 Metrics (elapsed: %.1fs)\n", elapsed.Seconds())
			fmt.Printf("  Connected peers: %d\n", atomic.LoadInt64(&metrics.ConnectedPeers))
			fmt.Printf("  Failed connects: %d\n", atomic.LoadInt64(&metrics.FailedConnects))
			fmt.Printf("  Peers discovered: %d\n", atomic.LoadInt64(&metrics.PeersDiscovered))
			fmt.Printf("  Messages received: %d\n", atomic.LoadInt64(&metrics.MessagesReceived))

			if elapsed > testDuration*2 {
				goto done
			}
		}
	}

done:
	wg.Wait()

	elapsed := time.Since(startTime)
	fmt.Printf("\n✅ Load Test Complete\n")
	fmt.Printf("================================\n")
	fmt.Printf("Total time: %.2fs\n", elapsed.Seconds())
	fmt.Printf("Connected: %d/%d peers\n", atomic.LoadInt64(&metrics.ConnectedPeers), *numPeers)
	fmt.Printf("Failed: %d peers\n", atomic.LoadInt64(&metrics.FailedConnects))
	fmt.Printf("Success rate: %.1f%%\n", float64(atomic.LoadInt64(&metrics.ConnectedPeers))/float64(*numPeers)*100)
	fmt.Printf("Peers discovered: %d\n", atomic.LoadInt64(&metrics.PeersDiscovered))
	fmt.Printf("Messages received: %d\n", atomic.LoadInt64(&metrics.MessagesReceived))
	fmt.Printf("Msg/sec: %.0f\n", float64(atomic.LoadInt64(&metrics.MessagesReceived))/elapsed.Seconds())
}
