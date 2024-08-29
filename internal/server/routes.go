package server

import (
	"ama/internal/middlewares"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("application/json", "text/xml"))
	r.Use(middleware.CleanPath)
	r.Use(middleware.Logger)

	r.Get("/", s.HelloWorldHandler)
	r.Get("/health", s.healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/sign-up", s.SignUp)
		r.Post("/sign-in", s.SignIn)
		r.Put("/verify", s.VerifyUser)
		r.Post("/send-message", s.SendMessage)

		r.Group(func(r chi.Router) {
			r.Use(middlewares.Auth)
			r.Post("/sign-out", s.SignOut)
			r.Put("/accept-messages", s.AcceptMessages)
		})
	})

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, _ := json.Marshal(s.db.Health())
	_, _ = w.Write(jsonResp)
}
