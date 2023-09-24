package main

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/url"
	"os"
	"oysterProject/model"
	"reflect"
	"strings"
	"time"
)

func isEmptyStruct(input interface{}) bool {
	zeroValue := reflect.New(reflect.TypeOf(input)).Elem().Interface()
	return reflect.DeepEqual(input, zeroValue)
}

func normalizeSocialLinks(user *model.Users) {
	user.LinkedInLink = makeURL(user.LinkedInLink, "linkedin.com/")
	user.InstagramLink = makeURL(user.InstagramLink, "instagram.com/")
	user.FacebookLink = makeURL(user.FacebookLink, "facebook.com/")
	user.CalendlyLink = makeURL(user.CalendlyLink, "calendly.com/")
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
