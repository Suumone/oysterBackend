package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"oysterProject/database"
	"oysterProject/emailNotifications"
	"oysterProject/routes"
	"oysterProject/schedulerJobs"
)

func main() {
	log.Println("Application started")
	err := database.ConnectToMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	defer database.CloseMongoDBConnection()
	database.ConnectToS3()
	schedulerJobs.StartJobs()
	emailNotifications.InitMailClient()

	r := chi.NewRouter()
	routes.ConfigureCors(r)
	routes.ConfigureRoutes(r)
	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
