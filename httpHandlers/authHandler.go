package httpHandlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"log"
	"net/http"
	"os"
	"oysterProject/database"
	"oysterProject/model"
	"time"
)

const (
	oauthGoogleUrlAPI     = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	dbTimeout             = 5 * time.Second
	jwtExpiration         = 7 * 24 * time.Hour
	oauthCookieExpiration = 365 * 24 * time.Hour
)

var (
	jwtKey = []byte(os.Getenv("JWT_SECRET"))
	conf   = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("ENV_URL") + "/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     endpoints.Google,
	}
)

type Oauth2User struct {
	Name  string `json:"name" bson:"name"`
	Email string `json:"email" bson:"email"`
}

func getCollection(collectionName string) *mongo.Collection {
	return database.MongoDBOyster.Collection(collectionName)
}

func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, d)
}

func HandleEmailPassAuth(w http.ResponseWriter, r *http.Request) {
	var user model.Users
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	ctx, cancel := withTimeout(context.Background(), dbTimeout)
	defer cancel()
	if _, err := getCollection("users").InsertOne(ctx, user); err != nil {
		http.Error(w, "Error inserting user into database", http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSONResponse(w, http.StatusCreated, []byte(tokenString))
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user model.Users
	json.NewDecoder(r.Body).Decode(&user)

	collection := database.MongoDBOyster.Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	var foundUser model.Users
	filter := bson.M{"email": user.Email}
	collection.FindOne(ctx, filter).Decode(&foundUser)

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password)); err != nil {
		WriteJSONResponse(w, http.StatusUnauthorized, "Wrong password")
		return
	}

	tokenString, err := generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSONResponse(w, http.StatusOK, []byte(tokenString))
}

func generateToken(user model.Users) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user.Email,
		"exp":  time.Now().Add(jwtExpiration).Unix(),
	})

	tokenString, _ := token.SignedString(jwtKey)
	return tokenString, nil
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(oauthCookieExpiration)

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("Failed to generate random state: %v", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	oauthState := generateStateOauthCookie(w)
	url := conf.AuthCodeURL(oauthState, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func getUserDataFromGoogle(code string) (model.Users, error) {
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return model.Users{}, err
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return model.Users{}, err
	}
	defer response.Body.Close()

	var oauth2User Oauth2User
	if err := json.NewDecoder(response.Body).Decode(&oauth2User); err != nil {
		return model.Users{}, err
	}
	return model.Users{Email: oauth2User.Email, Username: oauth2User.Name}, nil
}

func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie("oauthstate")
	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	userInfo, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	filter := bson.M{"email": userInfo.Email}
	var result model.Users
	collection := database.MongoDBOyster.Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	err = collection.FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		_, err = collection.InsertOne(ctx, userInfo)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(userInfo)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSONResponse(w, http.StatusOK, []byte(tokenString))
}
