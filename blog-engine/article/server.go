package article

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HTTPServer struct {
	port uint16
	host string

	uc *ArticleUseCase

	router *chi.Mux
}

func (s HTTPServer) Host() string {
	return s.host
}

func (s HTTPServer) Port() uint16 {
	return s.port
}

func (s *HTTPServer) setupRoute() {
	r := s.router

	r.Post("/articles", s.NewArticleHandler)
	r.Put("/articles/{articleID}", s.EditArticleHandler)
}

func NewHTTPServer(options ...func(*HTTPServer) error) (*HTTPServer, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	mem := CreateMemStore()

	uc, err := NewArticleUseCase(mem)

	if err != nil {
		return nil, err
	}

	httpServer := &HTTPServer{
		host:   "127.0.0.1",
		port:   8000,
		router: r,
		uc:     uc,
	}

	httpServer.setupRoute()

	if len(options) == 0 {
		return httpServer, nil
	}

	for _, opt := range options {
		if err := opt(httpServer); err != nil {
			return nil, err
		}
	}

	return httpServer, nil
}

func (s *HTTPServer) Start() {
	listen := fmt.Sprintf("%s:%d", s.host, s.port)

	http.ListenAndServe(listen, s.router)
}

func wrapError(err error) []byte {
	wrapper := struct {
		Message string `json:"message"`
	}{Message: err.Error()}

	j, _ := json.Marshal(wrapper)

	return j
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Add("content-type", "application-json")
	w.WriteHeader(status)
	w.Write(wrapError(err))
}

func writeJSON(w http.ResponseWriter, status int, result interface{}) {
	w.Header().Add("content-type", "application-json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(result)
}

var (
	ErrInvalidRequestPayload = errors.New("invalid request payload")
)

func (s *HTTPServer) NewArticleHandler(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	ctx := r.Context()

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidRequestPayload)
		return
	}

	article, err := s.uc.CreateArticle(ctx, payload.Title, payload.Content)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}

	result := struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
	}{article.ID.String(), article.CreatedAt.Format(time.RFC3339)}

	writeJSON(w, http.StatusCreated, result)

}

func (s *HTTPServer) EditArticleHandler(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, ErrNotImplemented)
}
