package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"os"
	"oysterProject/httpHandlers"
	"strings"
)

func ConfigureRoutes(r *chi.Mux) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/", httpHandlers.HandleEmailPassAuth)
		r.Get("/google", httpHandlers.HandleGoogleAuth)
		r.Get("/google/callback", httpHandlers.HandleAuthCallback)
	})
	r.Post("/signIn", httpHandlers.SignIn)
	r.With(httpHandlers.AuthMiddleware).Post("/signOut", httpHandlers.SignOut)
	r.With(httpHandlers.AuthMiddleware).Post("/refreshAuthSession", httpHandlers.RefreshAuthSession)

	r.With(httpHandlers.AuthMiddleware).Get("/getMentorList", httpHandlers.GetMentorsList)
	r.With(httpHandlers.AuthMiddleware).Post("/calculateBestMentors", httpHandlers.CalculateBestMentors)

	r.Get("/getMentorListFilters", httpHandlers.GetMentorListFilters)
	r.Get("/getMentor", httpHandlers.GetMentor)
	r.Get("/getTopMentors", httpHandlers.GetTopMentors)
	r.Get("/getReviews", httpHandlers.GetMentorReviews)
	r.Get("/getUserImage", httpHandlers.GetUserImage)
	r.Get("/getImageConfigurations", httpHandlers.GetImageConfigurations)
	r.Get("/getListValues", httpHandlers.GetListValues)

	r.Get("/getUserAvailableWeekdays", httpHandlers.GetUserAvailableWeekdays)
	r.Get("/getUserAvailableSlots", httpHandlers.GetUserAvailableSlots)

	r.With(httpHandlers.AuthMiddleware).Route("/myProfile", func(r chi.Router) {
		r.Get("/", httpHandlers.GetProfileByToken)
		r.Post("/update", httpHandlers.UpdateUserProfile)
		r.Post("/visibility", httpHandlers.UpdateVisibility)
		r.Post("/updatePassword", httpHandlers.ChangePassword)
		r.Get("/getCurrentState", httpHandlers.GetCurrentState)
		r.Post("/updateCurrentState", httpHandlers.UpdateCurrentState)
		r.Post("/uploadProfilePicture", httpHandlers.UploadUserImage)
	})

	r.With(httpHandlers.AuthMiddleware).Route("/session", func(r chi.Router) {
		r.Get("/", httpHandlers.GetSession)
		r.Get("/getUserSessions", httpHandlers.GetUserSessions)
		r.Post("/create", httpHandlers.CreateSession)
		r.Post("/rescheduleRequest", httpHandlers.RescheduleRequest)
		r.Post("/confirmRescheduleRequest", httpHandlers.ConfirmSessionRequest)
		r.Post("/cancelRescheduleRequest", httpHandlers.CancelRescheduleRequest)
		r.Post("/{sessionId}/createSessionReview", httpHandlers.CreateSessionReview)
	})

	r.With(httpHandlers.AuthMiddleware).Post("/createPublicReview", httpHandlers.CreatePublicReview)
}

func ConfigureCors(r *chi.Mux) {
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(os.Getenv("ALLOWED_ORIGINS"), ";"),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "X-CSRF-Token", "Set-Cookie", httpHandlers.SessionHeaderName},
		ExposedHeaders:   []string{httpHandlers.SessionHeaderName},
		AllowCredentials: true,
	})
	r.Use(corsConfig.Handler)
}
