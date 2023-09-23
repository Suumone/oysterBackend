package main

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"oysterProject/model"
	"strings"
	"time"
)

var mongoClient *mongo.Client

func main() {
	log.Println("Application started")
	mongoClient = connectToMongoDB()
	defer closeMongoDBConnection()

	http.HandleFunc("/createMentor", createMentor)
	err := http.ListenAndServe(":3333", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func createMentor(w http.ResponseWriter, r *http.Request) {
	log.Println("Received createMentor request")
	switch r.Method {
	case http.MethodPost:
		handleCreateRequest(w, r)
	case http.MethodOptions:
		handleOptionsRequest(w)
	default:
		writeMessageResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleCreateRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeMessageResponse(w, http.StatusBadRequest, "Error reading request")
		return
	}
	defer r.Body.Close()

	var payload model.Users
	if err := json.Unmarshal(body, &payload); err != nil {
		writeMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from createRequest")
		log.Println(err)
		return
	}
	payload.LinkedInLink = makeURL(payload.LinkedInLink, "linkedin.com/")
	payload.InstagramLink = makeURL(payload.InstagramLink, "instagram.com/")
	payload.FacebookLink = makeURL(payload.FacebookLink, "facebook.com/")
	payload.CalendlyLink = makeURL(payload.CalendlyLink, "calendly.com/")
	saveMentorInDB(payload)
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Println(err)
	}
}

func handleOptionsRequest(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PATCH, DELETE, PUT")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	w.WriteHeader(http.StatusOK)
}

func writeMessageResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	writeResponse(w, message)
}

func writeResponse(w http.ResponseWriter, message string) {
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Println(err)
	}
}

func connectToMongoDB() *mongo.Client {
	uri := os.Getenv("dbAddress")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	return client
}

func closeMongoDBConnection() {
	err := mongoClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}

func saveMentorInDB(user model.Users) {
	oysterDB := mongoClient.Database("Oyster")
	collection := oysterDB.Collection("users")
	doc, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("User(name: %s, insertedID: %s) inserted successfully!\n", user.Username, doc.InsertedID)
}

func makeURL(text string, urlPrefix string) string {
	if _, err := url.ParseRequestURI(text); err == nil {
		return text
	}
	if strings.HasPrefix(text, urlPrefix) {
		return "https://www." + text
	}

	return "https://www." + urlPrefix + strings.ReplaceAll(text, " ", "_")
}
