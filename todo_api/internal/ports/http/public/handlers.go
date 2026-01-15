package public

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/pkg/errors"
    "github.com/100bench/infr_training/internal/entities"
    "encoding/json"
    "net/http"
)

type Server struct {
	service ServicePort
	router  *chi.Mux
}


func NewServer(service ServicePort) (*Server, error) {
	if service == nil {
		return nil, errors.Wrap(entities.ErrNilDependency, "public server service")
	}
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	s := &Server{
		service: service,
		router:  r,
	}
	s.setupRoutes()
	return s, nil
}

func (s *Server) GetRouter() *chi.Mux {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Post("/task", s.handleCreateTask)
		r.Get("/task/{id}", s.handleGetTask)
		r.Delete("/task/{id}", s.handleDeleteTask)
	})
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
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    id := chi.URLParam(r, "id")
    task, err := s.service.GetTasks(ctx, id)
    if errors.Is(err, entities.ErrTaskNotFound){
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
    if errors.Is(err, entities.ErrTaskNotFound){
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}
