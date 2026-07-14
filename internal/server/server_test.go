package server_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gophernment/gophern/internal/server"
)

func TestServerRouter(t *testing.T) {
	// Create a temporary markdown file to serve as presentation source
	tmpFile, err := os.CreateTemp("", "test_presentation_*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	markdownContent := `---
title: Test Slide Deck
author: Gopher
theme: default
---
# Slide 1
Welcome to the test.
<!-- Notes for slide 1 -->
---
# Slide 2
Here is details.
<!-- Notes for slide 2 -->
`
	if _, err := tmpFile.WriteString(markdownContent); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	srv := server.NewServer(tmpFile.Name())
	router := srv.Router()

	// 1. Test GET / (Presentation View)
	t.Run("GET /", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "<title>Test Slide Deck</title>") {
			t.Errorf("expected title in html, got: %s", body)
		}
		if !strings.Contains(body, "Welcome to the test.") {
			t.Errorf("expected slide content in html, got: %s", body)
		}
	})

	// 2. Test GET /presenter (Presenter View)
	t.Run("GET /presenter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/presenter", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Presenter Console - Test Slide Deck") {
			t.Errorf("expected presenter title, got: %s", body)
		}
		if !strings.Contains(body, "Notes for slide 1") {
			t.Errorf("expected speaker notes payload, got: %s", body)
		}
	})

	// 3. Test POST /api/slide
	t.Run("POST /api/slide", func(t *testing.T) {
		payload := map[string]int{"index": 1}
		jsonBytes, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/slide", bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if rec.Body.String() != "OK" {
			t.Errorf("expected OK, got: %s", rec.Body.String())
		}
	})

	// 3a. Test POST /api/slide/prev
	t.Run("POST /api/slide/prev", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/slide/prev", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if rec.Body.String() != "OK" {
			t.Errorf("expected OK, got: %s", rec.Body.String())
		}
	})

	// 3b. Test POST /api/slide/next
	t.Run("POST /api/slide/next", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/slide/next", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if rec.Body.String() != "OK" {
			t.Errorf("expected OK, got: %s", rec.Body.String())
		}
	})

	// 4. Test GET /static/...
	t.Run("GET /static/css/styles.css", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/static/css/styles.css", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "--slide-width") {
			t.Errorf("expected css contents, got: %s", rec.Body.String())
		}
	})
}

func TestSSELiveStream(t *testing.T) {
	// Create a temporary markdown file to serve as presentation source
	tmpFile, err := os.CreateTemp("", "test_presentation_*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	srv := server.NewServer(tmpFile.Name())
	testServer := httptest.NewServer(srv.Router())
	defer testServer.Close()

	// Connect to SSE endpoint /events
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", testServer.URL+"/events", nil)
	if err != nil {
		t.Fatalf("failed to create SSE request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to connect to SSE stream: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)

	// Helper to read SSE data lines
	readNextSSEData := func() (string, error) {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return "", err
			}
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "data:") {
				return strings.TrimSpace(strings.TrimPrefix(line, "data:")), nil
			}
		}
	}

	// 1. Initial event is sent containing the initial slide index (0)
	initialData, err := readNextSSEData()
	if err != nil {
		t.Fatalf("failed to read initial SSE data: %v", err)
	}
	if initialData != `{"slide":0}` {
		t.Errorf("expected initial slide index JSON, got: %s", initialData)
	}

	// 2. Trigger a slide update via API in a goroutine
	go func() {
		time.Sleep(100 * time.Millisecond)
		payload := map[string]int{"index": 2}
		jsonBytes, _ := json.Marshal(payload)
		resp, err := http.Post(testServer.URL+"/api/slide", "application/json", bytes.NewReader(jsonBytes))
		if err != nil {
			t.Errorf("failed to trigger slide update: %v", err)
			return
		}
		resp.Body.Close()
	}()

	// 3. Read the next event containing the updated slide index (2)
	updatedData, err := readNextSSEData()
	if err != nil {
		t.Fatalf("failed to read updated SSE data: %v", err)
	}
	if updatedData != `{"slide":2}` {
		t.Errorf("expected updated slide index JSON, got: %s", updatedData)
	}
}
