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
	})
	r.Route("/getMentor/{id}/getReviews", func(r chi.Router) {
		r.Get("/", httpHandlers.HandleGetMentorReviews)
	})
	r.Route("/updateMentor/{id}", func(r chi.Router) {
		r.Post("/", httpHandlers.HandleUpdateMentor)
	})
}
