package routes

import (
	"github.com/go-chi/chi/v5"
	"oysterProject/httpHandlers"
)

func ConfigureRoutes(r *chi.Mux) {
	r.Post("/createMentor", httpHandlers.HandleCreateMentor)
	r.Get("/getMentorList", httpHandlers.HandleGetMentors)
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
