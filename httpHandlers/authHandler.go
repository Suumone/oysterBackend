package httpHandlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"log"
	"net/http"
	"net/mail"
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
			WriteJSONResponse(w, http.StatusForbidden, "Missing or malformed JWT")
			return
		}

		tokenStr := authHeader[1]
		claims := &jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			WriteJSONResponse(w, http.StatusForbidden, "Invalid token")
			return
		}

		if !token.Valid {
			WriteJSONResponse(w, http.StatusForbidden, "Expired token")
			return
		}

		collection := database.GetCollection("blacklistedTokens")
		filter := bson.M{"token": token.Raw}
		var blackListedToken model.Token
		collection.FindOne(context.Background(), filter).Decode(&blackListedToken)
		if !utils.IsEmptyStruct(blackListedToken) {
			WriteJSONResponse(w, http.StatusForbidden, "Invalid token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func HandleEmailPassAuth(w http.ResponseWriter, r *http.Request) {
	var authData model.Auth
	err := ParseJSONRequest(r, &authData)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	_, err = mail.ParseAddress(authData.Email)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Email is not valid")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(authData.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error hashing password")
		return
	}

	user := model.User{
		Email:     authData.Email,
		Password:  string(hashedPassword),
		IsMentor:  false,
		IsNewUser: true,
	}

	user.Id, err = database.CreateMentor(user)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Error inserting user into database")
		return
	}

	tokenString, err := generateToken(user)
	if err != nil {
		WriteMessageResponse(w, http.StatusInternalServerError, "Failed to generate JWT: "+err.Error())
		return
	}
	WriteJSONResponse(w, http.StatusCreated, tokenString)
}

func HandleSignIn(w http.ResponseWriter, r *http.Request) {
	var signInData model.Auth
	err := ParseJSONRequest(r, &signInData)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}
	if signInData.Password == "" {
		WriteMessageResponse(w, http.StatusNotFound, "Empty password")
		return
	}

	foundUser, err := database.GetUserByEmail(signInData.Email)
	if err != nil {
		WriteMessageResponse(w, http.StatusNotFound, "User not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(signInData.Password)); err != nil {
		WriteJSONResponse(w, http.StatusUnauthorized, "Wrong password")
		return
	}

	tokenString, err := generateToken(foundUser)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, "Failed to generate JWT: "+err.Error())
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

	userInfo, err = database.GetUserByEmail(userInfo.Email)
	if errors.Is(err, mongo.ErrNoDocuments) {
		userInfo.IsMentor = false
		userInfo.IsNewUser = true
		userInfo.Id, err = database.CreateMentor(userInfo)
		if err != nil {
			WriteJSONResponse(w, http.StatusInternalServerError, "Database insert error: "+err.Error())
			return
		}
	} else if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, "Database search error: "+err.Error())
		return
	}

	tokenString, err := generateToken(userInfo)
	if err != nil {
		WriteJSONResponse(w, http.StatusInternalServerError, "Failed to generate JWT: "+err.Error())
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
		WriteJSONResponse(w, http.StatusBadRequest, "Invalid token")
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
		WriteJSONResponse(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	WriteJSONResponse(w, http.StatusOK, "Successfully logged out")
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, err := getTokenClaimsFromRequest(r)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Invalid token")
		return
	}
	userId, _ := claims["id"].(string)
	var passwordPayload model.PasswordChange
	err = ParseJSONRequest(r, &passwordPayload)
	if err != nil {
		WriteMessageResponse(w, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	err = database.ChangePassword(userId, passwordPayload)
	if err != nil {
		log.Printf("Error updating password: %v\n", err)
		WriteMessageResponse(w, http.StatusInternalServerError, "Error updating password")
		return
	}
	WriteJSONResponse(w, http.StatusOK, "Password successfully updated")
}
