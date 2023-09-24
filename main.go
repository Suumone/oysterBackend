package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"os"
	"oysterProject/model"
)

var mongoClient *mongo.Client

func main() {
	log.Println("Application started")
	mongoClient = connectToMongoDB()
	defer closeMongoDBConnection()

	r := mux.NewRouter()
	r.HandleFunc("/createMentor", handleCreateMentor).Methods(http.MethodPost)
	r.HandleFunc("/getMentorList", handleGetMentors).Methods(http.MethodGet)
	r.HandleFunc("/getMentor/{id}", handleGetMentorByID).Methods(http.MethodGet)
	r.HandleFunc("/updateMentor/{id}", handleUpdateMentor).Methods(http.MethodPost)

	http.Handle("/", r)
	err := http.ListenAndServe(os.Getenv("port"), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handleCreateMentor(w http.ResponseWriter, r *http.Request) {
	var payload model.Users
	if err := parseJSONRequest(w, r, &payload); err != nil {
		return
	}
	normalizeSocialLinks(&payload)

	insertedID, err := saveMentorInDB(payload)
	if err != nil {
		writeMessageResponse(w, http.StatusInternalServerError, "Error saving user to MongoDB")
		return
	}
	writeJSONResponse(w, http.StatusCreated, insertedID)
}

func parseJSONRequest(w http.ResponseWriter, r *http.Request, payload interface{}) error {
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		writeMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		log.Println(err)
	}
	return err
}

func saveMentorInDB(user model.Users) (string, error) {
	collection := mongoClient.Database("Oyster").Collection("users")
	doc, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Println(err)
		return "", err
	}
	log.Printf("User(name: %s, insertedID: %s) inserted successfully\n", user.Username, doc.InsertedID)
	return doc.InsertedID.(primitive.ObjectID).Hex(), nil
}

func handleGetMentors(w http.ResponseWriter, r *http.Request) {
	users := getMentorsFromDB()
	writeJSONResponse(w, http.StatusOK, users)
}

func getMentorsFromDB() []model.Users {
	collection := mongoClient.Database("Oyster").Collection("users")
	filter := bson.M{"mentor": true}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to find documents: %v\n", err)
		return nil
	}
	defer cursor.Close(context.Background())

	var users []model.Users
	for cursor.Next(context.Background()) {
		var user model.Users
		if err := cursor.Decode(&user); err != nil {
			log.Printf("Failed to decode document: %v", err)
		}
		users = append(users, user)
	}
	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error: %v", err)
	}
	return users
}

func handleGetMentorByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	mentor := getMentorByIDFromDB(id)
	if isEmptyStruct(mentor) {
		writeMessageResponse(w, http.StatusNotFound, "Mentor not found")
		return
	}
	writeJSONResponse(w, http.StatusOK, mentor)
}

func getMentorByIDFromDB(id string) model.Users {
	collection := mongoClient.Database("Oyster").Collection("users")
	idToFind, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": idToFind}
	var user model.Users
	err := collection.FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Println("Document not found")
		} else {
			log.Printf("Failed to find document: %v\n", err)
		}
		return model.Users{}
	}
	return user
}

func handleUpdateMentor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var payload model.Users
	if err := parseJSONRequest(w, r, &payload); err != nil {
		return
	}

	if err := updateMentorInDB(payload, id); err != nil {
		writeMessageResponse(w, http.StatusInternalServerError, "Error updating user to MongoDB")
		return
	}
	writeJSONResponse(w, http.StatusOK, id)
}

func updateMentorInDB(user model.Users, id string) error {
	idToFind, _ := primitive.ObjectIDFromHex(id)
	collection := mongoClient.Database("Oyster").Collection("users")
	filter := bson.M{"_id": idToFind}
	updateOp := bson.M{"$set": user}
	_, err := collection.UpdateOne(context.Background(), filter, updateOp)
	if err != nil {
		return err
	}

	log.Printf("User(id: %s) updated successfully!\n", id)
	return nil
}

func writeMessageResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	writeResponse(w, message)
}

func writeJSONResponse(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}

func writeResponse(w http.ResponseWriter, message string) {
	_, err := w.Write([]byte(message))
	if err != nil {
		log.Println(err)
	}
}
