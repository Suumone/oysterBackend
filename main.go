package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"oysterProject/database"
	"oysterProject/routes"
)

func main() {
	log.Println("Application started")
	client, err := database.ConnectToMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	database.MongoDBClient = client
	database.MongoDBOyster = database.MongoDBClient.Database("Oyster")
	defer database.CloseMongoDBConnection()

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
