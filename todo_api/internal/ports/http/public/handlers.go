package public

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"github.com/100bench/infr_training/Internal/entities"
	"context"
)

type Server struct {
	service TodoService
	router  *chi.Mux
}


func NewServer(service TodoService) (*Server, error) {
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
		r.Post("/subscriptions", s.handleCreateTask)
		r.Get("/subscriptions/{id}", s.handleGetTask)
		r.Delete("/subscriptions/{id}", s.handleDeleteTask)
	})
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	