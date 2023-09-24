package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"os"
	"oysterProject/database"
	"oysterProject/routes"
)

func main() {
	log.Println("Application started")
	database.MongoDBClient = database.ConnectToMongoDB()
	defer database.CloseMongoDBConnection(database.MongoDBClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	routes.ConfigureRoutes(r)
	server := &http.Server{
		Addr:    os.Getenv("port"),
		Handler: r,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
