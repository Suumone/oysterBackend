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
	"oysterProject/utils"
	"strings"
	"time"
)

const (
	oauthGoogleUrlAPI     = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
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

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			http.Error(w, "Missing or malformed JWT", http.StatusForbidden)
			return
		}

		tokenStr := authHeader[1]
		claims := &jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		if !token.Valid {
			http.Error(w, "Expired token", http.StatusForbidden)
			return
		}

		collection := database.GetCollection("blacklistedTokens")
		filter := bson.M{"token": token.Raw}
		var blackListedToken model.Token
		collection.FindOne(context.Background(), filter).Decode(&blackListedToken)
		if !utils.IsEmptyStruct(blackListedToken) {
			http.Error(w, "Invalid token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HandleEmailPassAuth(w http.ResponseWriter, r *http.Request) {
	var user model.User
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
	user.IsMentor = false
	user.IsNewUser = true

	user.Id, err = database.SaveMentor(user)
	if err != nil {
		http.Error(w, "Error inserting user into database", http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSONResponse(w, http.StatusCreated, tokenString)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var user model.User
	json.NewDecoder(r.Body).Decode(&user)

	foundUser, err := database.GetUserByEmail(user)
	if err != nil {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password)); err != nil {
		WriteJSONResponse(w, http.StatusUnauthorized, "Wrong password")
		return
	}

	tokenString, err := generateToken(foundUser)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSONResponse(w, http.StatusOK, tokenString)
}

func generateToken(user model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.Id,
		"email": user.Email,
		"exp":   time.Now().Add(jwtExpiration).Unix(),
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

func getUserDataFromGoogle(code string) (model.User, error) {
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		return model.User{}, err
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return model.User{}, err
	}
	defer response.Body.Close()

	var oauth2User Oauth2User
	if err := json.NewDecoder(response.Body).Decode(&oauth2User); err != nil {
		return model.User{}, err
	}
	return model.User{Email: oauth2User.Email, Username: oauth2User.Name}, nil
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

	userInfo, err = database.GetUserByEmail(userInfo)
	if errors.Is(err, mongo.ErrNoDocuments) {
		userInfo.IsMentor = false
		userInfo.IsNewUser = true
		userInfo.Id, err = database.SaveMentor(userInfo)
		if err != nil {
			http.Error(w, "Database insert error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Database search error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := generateToken(userInfo)
	if err != nil {
		http.Error(w, "Failed to generate JWT: "+err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJSONResponse(w, http.StatusOK, tokenString)
}

func HandleLogOut(w http.ResponseWriter, r *http.Request) {
	tokenStr := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	expiry, _ := claims["exp"].(float64)
	expiresAt := time.Unix(int64(expiry), 0)

	collection := database.GetCollection("blacklistedTokens")
	_, err = collection.InsertOne(context.TODO(), model.BlacklistedToken{
		Token:     tokenStr,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	WriteJSONResponse(w, http.StatusOK, "Successfully logged out")
}
