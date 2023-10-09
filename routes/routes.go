package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"os"
	"oysterProject/httpHandlers"
	"strings"
)

func ConfigureRoutes(r *chi.Mux) {
	r.Get("/getMentorList", httpHandlers.GetMentors)
	r.Get("/getMentorListFilters", httpHandlers.GetMentorListFilters)
	r.Get("/getMentor", httpHandlers.GetMentor)
	r.Get("/getTopMentors", httpHandlers.GetTopMentors)
	r.Get("/getReviews", httpHandlers.GetMentorReviews)

	r.Route("/auth", func(r chi.Router) {
		r.Post("/", httpHandlers.HandleEmailPassAuth)
		r.Get("/google", httpHandlers.HandleGoogleAuth)
		r.Get("/google/callback", httpHandlers.HandleAuthCallback)
	})
	r.Post("/signIn", httpHandlers.HandleLogin)
	r.With(httpHandlers.JWTMiddleware).Post("/signOut", httpHandlers.HandleLogOut)

	r.With(httpHandlers.JWTMiddleware).Route("/myProfile", func(r chi.Router) {
		r.Get("/", httpHandlers.GetProfileByToken)
		r.Post("/update", httpHandlers.UpdateProfileByToken)
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
