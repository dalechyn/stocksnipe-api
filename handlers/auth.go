package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"time"
)

const EXPIRY_TIME = time.Hour * 24

// Login request structure
type LoginJSON struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var p LoginJSON

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// JWT creation
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(EXPIRY_TIME).Unix(),
	})

	// Pulling the secret key
	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		fmt.Println("SECRET KEY NOT FOUND")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokenString, _ := token.SignedString([]byte(secretKey))

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(tokenString)); err != nil {
		fmt.Println(err)
	}
}
