package api

import (
	"github.com/Insulince/triio-api/pkg/models"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"time"
	"log"
	"github.com/Insulince/triio-api/pkg/mongo"
	"encoding/base64"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"encoding/json"
	"github.com/Insulince/triio-api/pkg/configuration"
)

func Home(w http.ResponseWriter, r *http.Request) () {
	_, aw := models.NewApiCommunication(r, w)
	aw.Respond(struct{ Message string `json:"message"` }{Message: "Welcome!"}, http.StatusOK)
}

func HealthCheck(w http.ResponseWriter, r *http.Request) () {
	_, aw := models.NewApiCommunication(r, w)
	aw.Respond(struct{ Message string `json:"message"` }{Message: "OK"}, http.StatusOK)
}

func NotFound(w http.ResponseWriter, r *http.Request) () {
	_, aw := models.NewApiCommunication(r, w)
	aw.Respond(struct{ Message string `json:"message"` }{Message: "Unsupported URL provided."}, http.StatusNotFound)
}

func Register(config *configuration.Config) (func(w http.ResponseWriter, r *http.Request) ()) {
	return func(w http.ResponseWriter, r *http.Request) () {
		ar, aw := models.NewApiCommunication(r, w)

		type PostBody struct {
			Email    string `json:"email"`
			Username string `json:"username"`
			Password string `json:"password"`
		}

		type Response struct {
			Message string `json:"message"`
			Result  bool   `json:"result"`
		}

		rawPostBody, err := ar.GetRequestBody()
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Could not read request body.", Result: false}, http.StatusBadRequest)
			return
		}

		var postBody PostBody
		err = json.Unmarshal(rawPostBody, &postBody)
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Could not parse request body.", Result: false}, http.StatusBadRequest)
			return
		}

		if postBody.Email == "" || postBody.Username == "" || postBody.Password == "" {
			log.Println("Malformed request body.")
			aw.Respond(&Response{Message: "Malformed request body.", Result: false}, http.StatusBadRequest)
			return
		}

		user, err := mongo.FindUserByEmail(postBody.Email)
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Could not lookup provided email.", Result: false}, http.StatusInternalServerError)
			return
		}
		if user != nil {
			log.Println("Email already registered.")
			aw.Respond(&Response{Message: "Email already registered.", Result: false}, http.StatusBadRequest)
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(postBody.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Could not generate password hash.", Result: false}, http.StatusInternalServerError)
			return
		}

		err = mongo.InsertUser(models.User{Email: postBody.Email, PasswordHash: passwordHash, CreationTimestamp: time.Now().Unix()})
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Could not insert user.", Result: false}, http.StatusInternalServerError)
			return
		}

		aw.Respond(&Response{Message: "User registered successfully.", Result: true}, http.StatusOK)
	}
}

func Login(config *configuration.Config) (func(w http.ResponseWriter, r *http.Request) ()) {
	return func(w http.ResponseWriter, r *http.Request) () {
		ar, aw := models.NewApiCommunication(r, w)

		type Response struct {
			Message string `json:"message"`
			Result  bool   `json:"result"`
			Token   string `json:"token"`
		}

		authorizationHeader := ar.GetHeader("Authorization")
		if authorizationHeader == "" {
			log.Println("Empty/absent \"Authorization\" header value.")
			aw.Respond(&Response{Message: "Empty/absent \"Authorization\" header value.", Result: false, Token: ""}, http.StatusBadRequest)
			return
		}
		if strings.Index(authorizationHeader, "Basic ") != 0 {
			log.Println("Missing \"Basic \" in Authorization header.")
			aw.Respond(&Response{Message: "Malformed Authorization header.", Result: false, Token: ""}, http.StatusBadRequest)
			return
		}
		authorizationHeader = authorizationHeader[6:] // Remove "Basic " from header.

		type Authorization struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		var authorization Authorization
		decodedAuthorizationBytes, err := base64.StdEncoding.DecodeString(authorizationHeader)
		decodedAuthorizationString := string(decodedAuthorizationBytes)
		authorization.Email = decodedAuthorizationString[:strings.Index(decodedAuthorizationString, ":")]
		authorization.Password = decodedAuthorizationString[strings.Index(decodedAuthorizationString, ":")+1:]

		if authorization.Email == "" || authorization.Password == "" {
			log.Println("Malformed authorization header.")
			aw.Respond(&Response{Message: "Malformed Authorization header.", Result: false, Token: ""}, http.StatusBadRequest)
			return
		}

		user, err := mongo.FindUserByEmail(authorization.Email)
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Failed to validate credentials.", Result: false, Token: ""}, http.StatusInternalServerError)
			return
		}
		if user == nil {
			log.Printf("No user found for provided email \"%v\"\n", authorization.Email)
			aw.Respond(&Response{Message: "Invalid email or password.", Result: false, Token: ""}, http.StatusBadRequest)
			return
		}

		err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(authorization.Password))
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Invalid email or password.", Result: false, Token: ""}, http.StatusBadRequest)
			return
		}

		token := jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			jwt.MapClaims{
				"email":        user.Email,
				"passwordHash": user.PasswordHash,
			},
		)

		tokenString, err := token.SignedString([]byte(config.JwtSecret))
		if err != nil {
			log.Println(err)
			aw.Respond(&Response{Message: "Unable to generate token.", Result: false, Token: ""}, http.StatusInternalServerError)
			return
		}

		aw.Respond(Response{Message: "Success.", Result: true, Token: tokenString}, http.StatusOK)
	}
}
