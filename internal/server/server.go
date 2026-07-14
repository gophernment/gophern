package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sync"

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
		defer s.mu.RUnlock()
		return s.currentIndex
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
	mux.Handle("GET /static/", http.FileServer(http.FS(web.Assets)))
	return mux
}

// Start launches the HTTP server on the specified port.
func Start(markdownFile, port string, stdout io.Writer) error {
	s := NewServer(markdownFile)
	fmt.Fprintf(stdout, "Starting server on port %s...\n", port)
	return http.ListenAndServe(":"+port, s.Router())
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
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			return string(b), err
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

	s.broker.Broadcast(payload.Index)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
