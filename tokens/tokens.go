package tokens

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/h0tw4t3r/stocksnipe_api/db"
	"github.com/h0tw4t3r/stocksnipe_api/models"
	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/uuid"
)

const (
	accessTokenExpiry = 2 * time.Minute
	accessTokenType   = "access"

	refreshTokenExpiry = 3 * time.Minute
	refreshTokenType   = "refresh"
)

type refreshTokenClaims struct {
	UID                uuid.UUID
	TokenType          string
	jwt.StandardClaims
}

var tokenCollection *mongo.Collection

func init() {
	tokenCollection = db.MongoInit().Database("StockSnipeAPI").Collection("tokens")
}

// Looks up for user ID from database
func GetUserID(refreshToken string) (string, error) {
	refreshTokenUID, err := validateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Looking up for userID in database by decoded refresh token UID
	accessToken := &models.TokenSchema{}
	if err := tokenCollection.
		FindOne(context.TODO(), bson.E{"tokenID", refreshTokenUID}).
		Decode(accessToken); err != nil {
		log.Error(err)
		return "", err
	}

	return accessToken.UserId, nil
}

// Generates new pair of tokens, access-first
func UpdateTokens(userID string) (string, string, error) {
	accessToken, err := generateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	id, refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	err = replaceDBRefreshToken(id, userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Validates refresh token by type, algorithm and expiry
func validateRefreshToken(refreshToken string) (*uuid.UUID, error) {
	// Pulling out secret key to validate and decode refresh JWT
	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		err := errors.New("secret key not found")
		log.Error(err)
		return nil, err
	}

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		// Checking encryption algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secretKey, nil
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Pulling out userID stored in JWT
	claims, ok := token.Claims.(refreshTokenClaims)

	// Checking if token is valid
	if !ok || !token.Valid {
		log.Error(err)
		return nil, err
	}

	// RefreshToken type must be a refresh token
	if claims.TokenType != refreshTokenType {
		err := errors.New("bad token type")
		log.Error(err)
		return nil, err
	}

	// Checking token expiry
	if !claims.VerifyExpiresAt(time.Now().Unix(), false) {
		err := errors.New("refreshToken expired")
		log.Error(err)
		return nil, err
	}

	return &claims.UID, nil
}

// Generates Access token which has the userID inside
func generateAccessToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, struct {
		userID string
		tokenType string
		jwt.StandardClaims
	}{
		userID,
		accessTokenType,
		jwt.StandardClaims{ExpiresAt: time.Now().Add(accessTokenExpiry).Unix()}})

	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		log.Error("Secret key not found")
		return "", errors.New("secret key not found")
	}

	return token.SignedString([]byte(secretKey))
}

// Generates Refresh token
func generateRefreshToken() (*uuid.UUID, string, error) {
	uid, err := uuid.New()
	if err != nil {
		log.Error(err)
		return nil, "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, struct {
		id uuid.UUID
		tokenType string
		jwt.StandardClaims
	}{
		uid,
		refreshTokenType,
		jwt.StandardClaims{ExpiresAt: time.Now().Add(refreshTokenExpiry).Unix()}})

	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		log.Error("Secret key not found")
		return nil, "", errors.New("secret key not found")
	}

	res, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Error(err)
		return nil, "", err
	}

	return &uid, res, nil
}

// Replaces old Refresh token ID with a new one given
func replaceDBRefreshToken(refreshTokenID *uuid.UUID, userID string) error {
	replaced := &models.TokenSchema{}
	if err := tokenCollection.
		FindOneAndReplace(context.TODO(), bson.E{"_id",userID},
			models.TokenSchema{TokenId: *refreshTokenID, UserId: userID},
			options.FindOneAndReplace().SetUpsert(true)).
		Decode(&replaced); err != nil && err != mongo.ErrNoDocuments {
		log.Error(err)
		return err
	}
	log.Debug(log.WithField("replaced", replaced), "Replaced Refresh token")

	return nil
}

