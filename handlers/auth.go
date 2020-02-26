package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

const ExpiryTime = time.Hour * 24

var validate *validator.Validate

func passwordValidation(fl validator.FieldLevel) bool {
	pass := fl.Field().String()

	hasLowerCase := false
	hasUpperCase := false
	hasDigit := false
	for _, c := range pass {
		if c >= 'a' && c <= 'z' {
			hasLowerCase = true
		} else if c >= 'A' && c <= 'Z' {
			hasUpperCase = true
		} else if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}

	return hasUpperCase && hasLowerCase && hasDigit && len(pass) < 8 && len(pass) > 32
}


func init() {
	validate = validator.New()
	if err := validate.RegisterValidation("password", passwordValidation); err != nil {
		panic(err)
	}
}

func Register(w http.ResponseWriter, r *http.Request) {
	var registerRequestBody struct {
		Email string `json:"email" validate:"required"`
		Login string `json:"login" validate:"required"`
		Password string `json:"password" validate:"min=8,max=32"`
	}

	err := json.NewDecoder(r.Body).Decode(&registerRequestBody)
	if err != nil {
		log.Errorln(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Println(registerRequestBody)

	if errs := validate.Struct(registerRequestBody); errs != nil {
		log.Errorln(errs)
		http.Error(w, errs.Error(), http.StatusUnprocessableEntity)
		return
	}

	fmt.Println(registerRequestBody)

}

// Login request structure
func Login(w http.ResponseWriter, r *http.Request) {
	var loginRequestBody struct {
		Login string `json:"login"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginRequestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// JWT creation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(ExpiryTime).Unix(),
	})

	// Pulling the secret key
	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		log.Error("Secret key not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, _ := token.SignedString([]byte(secretKey))

	w.Header().Set("Content-Type", "application/json")

	res := struct {
		Login string `json:"Login"`
		JWT   string `json:"JWT"`
	}{loginRequestBody.Login, tokenString}

	resBytes, err := json.Marshal(res)
	if err != nil {
		log.WithField("err", err).Error("JSON Marshal() error")
	}

	if _, err := w.Write(resBytes); err != nil {
		log.Error(err)
	}
}
