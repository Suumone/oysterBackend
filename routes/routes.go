package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"os"
	"oysterProject/httpHandlers"
	"strings"
)

func ConfigureRoutes(r *chi.Mux) {
	r.Post("/createMentor", httpHandlers.HandleCreateMentor)
	r.Get("/getMentorList", httpHandlers.HandleGetMentors)
	r.Get("/getMentorListFilters", httpHandlers.HandleGetMentorListFilters)
	r.Route("/getMentor/{id}", func(r chi.Router) {
		r.Get("/", httpHandlers.HandleGetMentorByID)
		r.Get("/getReviews", httpHandlers.HandleGetMentorReviews)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/", httpHandlers.HandleEmailPassAuth)
		r.Get("/google", httpHandlers.HandleGoogleAuth)
		r.Get("/google/callback", httpHandlers.HandleAuthCallback)
	})
	r.Post("/login", httpHandlers.HandleLogin)
	r.With(httpHandlers.JWTMiddleware).Post("/logout", httpHandlers.HandleLogOut)

	r.With(httpHandlers.JWTMiddleware).Route("/myProfile", func(r chi.Router) {
		r.Get("/", httpHandlers.HandleGetProfileByToken)
		r.Post("/update", httpHandlers.HandleUpdateProfileByToken)
	})
}

func ConfigureCors(r *chi.Mux) {
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(os.Getenv("ALLOWED_ORIGINS"), ";"),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})
	r.Use(corsConfig.Handler)
}
