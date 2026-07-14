package server

import (
	"fmt"
	"net/http"
	"sync"
)

// Broker handles multiple SSE clients.
type Broker struct {
	mu           sync.Mutex
	clients      map[chan int]bool
	initialIndex func() int
}

// NewBroker creates a new Broker instance.
func NewBroker(initialIndex func() int) *Broker {
	return &Broker{
		clients:      make(map[chan int]bool),
		initialIndex: initialIndex,
	}
}

// ServeHTTP implements the http.Handler interface for the SSE endpoint.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for this client
	ch := make(chan int, 10)

	// Register client
	b.mu.Lock()
	b.clients[ch] = true
	b.mu.Unlock()

	// Remove client when connection is closed
	defer func() {
		b.mu.Lock()
		delete(b.clients, ch)
		b.mu.Unlock()
		close(ch)
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial handshake and current slide index
	fmt.Fprintf(w, ": ok\n\n")
	if b.initialIndex != nil {
		fmt.Fprintf(w, "data: {\"slide\":%d}\n\n", b.initialIndex())
	}
	flusher.Flush()

	// Listen for client context done
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case idx, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: {\"slide\":%d}\n\n", idx)
			flusher.Flush()
		}
	}
}

// Broadcast sends the slide index to all registered clients.
func (b *Broker) Broadcast(index int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.clients {
		select {
		case ch <- index:
		default:
			// Discard if client is blocked to prevent hanging
		}
	}
}
