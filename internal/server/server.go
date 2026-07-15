package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gophernment/gophern/internal/parser"
	"github.com/gophernment/gophern/web"
)

// Server keeps track of the presentation markdown file path,
// the current slide index, and the SSE Broker.
type Server struct {
	markdownFile string
	broker       *Broker

	mu           sync.RWMutex
	currentIndex int
}

// NewServer creates a new Server instance.
func NewServer(markdownFile string) *Server {
	var s *Server
	s = &Server{
		markdownFile: markdownFile,
	}
	s.broker = NewBroker(func() int {
		s.mu.RLock()
		idx := s.currentIndex
		s.mu.RUnlock()

		// Clamp to the current slide count in case the markdown file was
		// edited (e.g. slides removed) since the index was last set, so a
		// reconnecting client (page refresh, hot-reload) lands on the last
		// slide that still exists instead of silently resetting to slide 1.
		pres, err := parser.ParseMarkdownFile(s.markdownFile)
		if err != nil {
			return idx
		}
		if idx >= len(pres.Slides) {
			idx = len(pres.Slides) - 1
		}
		if idx < 0 {
			idx = 0
		}
		return idx
	})
	return s
}

// Router returns the http.Handler configured with all server routes.
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handlePresentation)
	mux.HandleFunc("GET /presenter", s.handlePresenter)
	mux.HandleFunc("GET /events", s.broker.ServeHTTP)
	mux.HandleFunc("POST /api/slide", s.handleUpdateSlide)
	mux.HandleFunc("POST /api/slide/next", s.handleNextSlide)
	mux.HandleFunc("POST /api/slide/prev", s.handlePrevSlide)
	mux.Handle("GET /static/", http.FileServer(http.FS(web.Assets)))
	assetDir := filepath.Join(filepath.Dir(s.markdownFile), "asset")
	mux.Handle("GET /asset/", http.StripPrefix("/asset/", http.FileServer(http.Dir(assetDir))))
	return mux
}

// Start launches the HTTP server on the specified port.
func Start(markdownFile, port string, stdout io.Writer) error {
	s := NewServer(markdownFile)
	go s.watchFile()
	fmt.Fprintf(stdout, "Starting server on port %s...\n", port)
	return http.ListenAndServe(":"+port, s.Router())
}

func (s *Server) watchFile() {
	info, err := os.Stat(s.markdownFile)
	if err != nil {
		return
	}
	lastModTime := info.ModTime()

	for {
		time.Sleep(500 * time.Millisecond)
		currentInfo, err := os.Stat(s.markdownFile)
		if err != nil {
			continue
		}
		if currentInfo.ModTime().After(lastModTime) {
			lastModTime = currentInfo.ModTime()
			s.broker.Broadcast("data: {\"reload\":true}\n\n")
		}
	}
}

func (s *Server) handlePresentation(w http.ResponseWriter, r *http.Request) {
	pres, err := parser.ParseMarkdownFile(s.markdownFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing markdown: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("presentation.html").Funcs(template.FuncMap{
		"safe": func(content string) template.HTML { return template.HTML(content) },
	}).ParseFS(web.Assets, "templates/presentation.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, pres); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
	}
}

func (s *Server) handlePresenter(w http.ResponseWriter, r *http.Request) {
	pres, err := parser.ParseMarkdownFile(s.markdownFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing markdown: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("presenter.html").Funcs(template.FuncMap{
		"safe": func(content string) template.HTML { return template.HTML(content) },
		"json": func(v interface{}) (template.JS, error) {
			b, err := json.Marshal(v)
			return template.JS(b), err
		},
	}).ParseFS(web.Assets, "templates/presenter.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading template: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, pres); err != nil {
		http.Error(w, fmt.Sprintf("Error executing template: %v", err), http.StatusInternalServerError)
	}
}

type SlidePayload struct {
	Index int `json:"index"`
}

func (s *Server) handleUpdateSlide(w http.ResponseWriter, r *http.Request) {
	var payload SlidePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.currentIndex = payload.Index
	s.mu.Unlock()

	s.broker.Broadcast(fmt.Sprintf("data: {\"slide\":%d}\n\n", payload.Index))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handleNextSlide(w http.ResponseWriter, r *http.Request) {
	pres, err := parser.ParseMarkdownFile(s.markdownFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing markdown: %v", err), http.StatusInternalServerError)
		return
	}
	totalSlides := len(pres.Slides)

	s.mu.RLock()
	curr := s.currentIndex
	s.mu.RUnlock()

	if curr >= totalSlides-1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	s.mu.Lock()
	if s.currentIndex < totalSlides-1 {
		s.currentIndex++
	}
	newIdx := s.currentIndex
	s.mu.Unlock()

	s.broker.Broadcast(fmt.Sprintf("data: {\"slide\":%d}\n\n", newIdx))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) handlePrevSlide(w http.ResponseWriter, r *http.Request) {
	_, err := parser.ParseMarkdownFile(s.markdownFile)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing markdown: %v", err), http.StatusInternalServerError)
		return
	}

	s.mu.RLock()
	curr := s.currentIndex
	s.mu.RUnlock()

	if curr <= 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}

	s.mu.Lock()
	if s.currentIndex > 0 {
		s.currentIndex--
	}
	newIdx := s.currentIndex
	s.mu.Unlock()

	s.broker.Broadcast(fmt.Sprintf("data: {\"slide\":%d}\n\n", newIdx))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
