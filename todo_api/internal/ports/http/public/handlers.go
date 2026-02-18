package public

import (
	"encoding/json"
	"net/http"

	er "github.com/100bench/infr_training/pkg/errors"
	mw "github.com/100bench/infr_training/todo_api/internal/ports/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	service ServicePort
	router  *chi.Mux
    readyChecker ReadyChecker
}


func NewServer(service ServicePort, readyChecker ReadyChecker) (*Server, error) {
	if service == nil {
		return nil, errors.Wrap(er.ErrNilDependency, "public server service")
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
    r.Use(mw.MetricsMiddleware)
	s := &Server{
		service: service,
		router:  r,
		readyChecker: readyChecker,
	}
	s.setupRoutes()
	return s, nil
}

func (s *Server) GetRouter() *chi.Mux {
	return s.router
}

func (s *Server) setupRoutes() {  
    s.router.Handle("/metrics", promhttp.Handler())
    s.router.Get("/live", s.handleLive)
    s.router.Get("/ready", s.handleReady)
    s.router.Route("/api/v1", func(r chi.Router) {
		r.Post("/task", s.handleCreateTask)
		r.Get("/task/{id}", s.handleGetTask)
		r.Delete("/task/{id}", s.handleDeleteTask)
	})
}


func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
    ctx:= r.Context()
    if err:= s.readyChecker.Ready(ctx); err != nil{
        http.Error(w, "service not ready",  http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLive(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    var req struct {
        Title       string `json:"title"`
        Description string `json:"description"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if req.Title == "" {
        http.Error(w, "title required", http.StatusBadRequest)
        return
    }
    task, err := s.service.CreateTask(ctx, req.Title, req.Description)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
    // if err := json.NewEncoder(w).Encode(task); err != nil {
    //   logger.Error("failed to encode response", zap.Error(err))
    // }
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    id := chi.URLParam(r, "id")
    if _, err := uuid.Parse(id); err != nil {
      http.Error(w, "invalid uuid", http.StatusBadRequest)
      return
    }
    task, err := s.service.GetTask(ctx, id)
    if errors.Is(err, er.ErrEntityNotFound){
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := chi.URLParam(r, "id")
    err := s.service.DeleteTask(ctx, id)
    if errors.Is(err, er.ErrEntityNotFound){
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
