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
	database.MongoDBClient = database.ConnectToMongoDB()
	database.MongoDBOyster = database.MongoDBClient.Database("Oyster")
	defer database.CloseMongoDBConnection(database.MongoDBClient)

	r := chi.NewRouter()
	routes.ConfigureRoutes(r)
	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
