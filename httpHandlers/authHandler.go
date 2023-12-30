package httpHandlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	"time"
)

const (
	oauthGoogleUrlAPI     = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	expirationTime        = 30 * 24 * time.Hour
	oauthCookieExpiration = 365 * 24 * time.Hour
	sessionCookieName     = "sessionId"
	oauthStateCookieName  = "oauthState"
)

var (
	conf = &oauth2.Config{
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

type session struct {
	SessionId primitive.ObjectID `bson:"_id,omitempty"`
	UserId    primitive.ObjectID `bson:"userId"`
	Expiry    int64              `bson:"expiry"`
}

func (s session) isExpired() bool {
	return s.Expiry < time.Now().Unix()
}

func saveSession(s session) (string, error) {
	collection := database.GetCollection(database.AuthSessionCollectionName)
	result, err := collection.InsertOne(context.Background(), s)
	if err != nil {
		log.Printf("Error saving session in db: %v\n", err)
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), err
}

func getUserSessionFromRequest(r *http.Request) *session {
	userSession, ok := r.Context().Value("userSession").(*session)
	if !ok {
		log.Printf("getUserSessionFromRequest: no session in context")
		return nil
	}

	return userSession
}

func findSession(sessionId string) (*session, bool) {
	collection := database.GetCollection(database.AuthSessionCollectionName)
	sessionIdObj, err := primitive.ObjectIDFromHex(sessionId)
	if err != nil {
		log.Printf("Failed to convert string identifier to object(%s): %v\n", sessionId, err)
		return nil, false
	}
	filter := bson.M{"_id": sessionIdObj}
	result := collection.FindOne(context.Background(), filter)
	err = result.Err()
	if err != nil {
		log.Printf("Session was not found(%s): %v\n", sessionId, err)
		return nil, false
	}
	var s session
	err = result.Decode(&s)
	if err != nil {
		log.Printf("Error decoding session(%s): %v\n", sessionId, err)
		return nil, false
	}
	if time.Now().Unix() > s.Expiry {
		log.Printf("Session(%s) expired\n", sessionId)
		return nil, false
	}
	return &s, true
}

func deleteSession(s *session) error {
	collection := database.GetCollection(database.AuthSessionCollectionName)
	filter := bson.M{"_id": s.SessionId}
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil || result.DeletedCount == 0 {
		log.Printf("Error deleting session(%s): %v\n", s.SessionId, err)
		return err
	}
	return nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				WriteMessageResponse(w, r, http.StatusUnauthorized, "Missed session id")
				return
			}
			WriteMessageResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}
		sessionId := c.Value
		userSession, ok := findSession(sessionId)
		if !ok {
			WriteMessageResponse(w, r, http.StatusUnauthorized, "User unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), "userSession", userSession)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	var credentials model.Auth

	if err := render.DecodeJSON(r.Body, &credentials); err != nil {
		WriteMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	if _, err := mail.ParseAddress(credentials.Email); err != nil {
		WriteMessageResponse(w, r, http.StatusBadRequest, "Email is not valid")
		return
	}

	user, err := database.GetUserByEmail(credentials.Email)
	if err != nil || user == nil {
		WriteMessageResponse(w, r, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		WriteMessageResponse(w, r, http.StatusUnauthorized, "Wrong password")
		return
	}

	expiresAt := time.Now().Add(expirationTime)
	sessionId, err := saveSession(session{
		UserId: user.Id,
		Expiry: expiresAt.Unix(),
	})
	if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Database saving session error")
		return
	}

	writeSessionCookie(w, sessionCookieName, sessionId, expiresAt)
	WriteMessageResponse(w, r, http.StatusOK, "Sign in successful")
}

func SignOut(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		WriteMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	err := deleteSession(userSession)
	if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Error deleting session")
		return
	}

	deleteCookie(w, sessionCookieName)
	WriteMessageResponse(w, r, http.StatusOK, "Sign out successful")
}

func HandleEmailPassAuth(w http.ResponseWriter, r *http.Request) {
	var authData model.Auth
	err := ParseJSONRequest(r, &authData)
	if err != nil {
		WriteMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	_, err = mail.ParseAddress(authData.Email)
	if err != nil {
		WriteMessageResponse(w, r, http.StatusBadRequest, "Email is not valid")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(authData.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Error hashing password")
		return
	}

	user := model.User{
		Email:     authData.Email,
		Password:  string(hashedPassword),
		IsNewUser: true,
		AsMentor:  authData.AsMentor,
	}

	user.Id, err = database.CreateMentor(&user)
	if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Error inserting user into database")
		return
	}

	expiresAt := time.Now().Add(expirationTime)
	sessionId, err := saveSession(session{
		UserId: user.Id,
		Expiry: expiresAt.Unix(),
	})
	if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Database saving session error")
		return
	}

	writeSessionCookie(w, sessionCookieName, sessionId, expiresAt)
	WriteMessageResponse(w, r, http.StatusOK, "Sign up successful")
}

func generateStateOauthCookie(w http.ResponseWriter) (string, error) {
	expiration := time.Now().Add(oauthCookieExpiration)

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Printf("Failed to generate random state: %v", err)
		return "", err
	}

	state := base64.URLEncoding.EncodeToString(b)
	writeSessionCookie(w, oauthStateCookieName, state, expiration)
	return state, nil
}

func HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	oauthState, err := generateStateOauthCookie(w)
	if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Failed to generate state")
		return
	}
	url := conf.AuthCodeURL(oauthState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func getUserDataFromGoogle(code string) (*model.User, error) {
	token, err := conf.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Failed to exchange token: %v\n", err)
		return nil, err
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		log.Printf("Failed to get user info from google: %v\n", err)
		return nil, err
	}
	defer response.Body.Close()

	var oauth2User Oauth2User
	if err = json.NewDecoder(response.Body).Decode(&oauth2User); err != nil {
		log.Printf("Failed to decode response from google: %v\n", err)
		return nil, err
	}
	return &model.User{Email: oauth2User.Email, Username: oauth2User.Name}, nil
}

func HandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie(oauthStateCookieName)
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

	user, err := database.GetUserByEmail(userInfo.Email)
	if errors.Is(err, mongo.ErrNoDocuments) {
		user.IsNewUser = true
		user.Id, err = database.CreateMentor(userInfo)
		if err != nil {
			WriteMessageResponse(w, r, http.StatusInternalServerError, "Database insert error: "+err.Error())
			return
		}
	} else if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Database search error: "+err.Error())
		return
	}

	expiresAt := time.Now().Add(expirationTime)
	sessionId, err := saveSession(session{
		UserId: user.Id,
		Expiry: expiresAt.Unix(),
	})
	if err != nil {
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Database saving session error")
		return
	}

	writeSessionCookie(w, sessionCookieName, sessionId, expiresAt)

	WriteMessageResponse(w, r, http.StatusOK, "Sign up successful")
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	userSession := getUserSessionFromRequest(r)
	if userSession == nil {
		WriteMessageResponse(w, r, http.StatusBadRequest, "No user session info was found")
		return
	}
	var passwordPayload model.PasswordChange
	err := ParseJSONRequest(r, &passwordPayload)
	if err != nil {
		WriteMessageResponse(w, r, http.StatusBadRequest, "Error parsing JSON from request")
		return
	}

	err = database.ChangePassword(userSession.UserId, passwordPayload)
	if err != nil {
		log.Printf("Error updating password: %v\n", err)
		WriteMessageResponse(w, r, http.StatusInternalServerError, "Error updating password")
		return
	}
	WriteMessageResponse(w, r, http.StatusOK, "Password successfully updated")
}
