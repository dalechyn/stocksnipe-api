package tokens

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/square/go-jose"
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
	accessTokenExpiry = 3 * time.Minute
	accessTokenType   = "access"

	refreshTokenExpiry = 3 * time.Hour
	refreshTokenType   = "refresh"
)

type accessTokenClaims struct {
	userID string
	tokenType string
	jwt.StandardClaims
}

type refreshTokenClaims struct {
	Id                 uuid.UUID
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
		FindOne(context.TODO(), bson.M{"tokenID": refreshTokenUID}).
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

func getSecretKey() (string, error) {
	secretKey, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		err := errors.New("secret key not found")
		return "", err
	}
	return secretKey, nil
}

func ValidateAccessToken(accessToken string) error {
	// Pulling out secret key to validate and decode access JWT
	secretKey, err := getSecretKey()
	if err != nil {
		log.Error(err)
		return err
	}
	token, err := jwt.ParseWithClaims(accessToken, &accessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Checking encryption algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil
	})

	return nil
}

// Validates refresh token by type, algorithm and expiry
func validateRefreshToken(refreshToken string) (*uuid.UUID, error) {
	// Pulling out secret key to validate and decode refresh JWT
	secretKey, err := getSecretKey()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	token, err := jwt.ParseWithClaims(refreshToken, &refreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Pulling out userID stored in JWT
		claims, ok := token.Claims.(*refreshTokenClaims)
		if !ok {
			err := errors.New("refreshTokenClaims destruction failed")
			log.Error(err)
			return nil, err
		}

		// RefreshToken type must be a refresh token
		if claims.TokenType != refreshTokenType {
			err := errors.New("bad token type")
			log.Error(err)
			return nil, err
		}

		// Checking encryption algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secretKey), nil
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Checking if token is valid
	if !ok || !token.Valid || claims.Valid() {
		log.Error(err)
		return nil, err
	}



	// Checking token expiry
	if !claims.VerifyExpiresAt(time.Now().Unix(), false) {
		err := errors.New("refreshToken expired")
		log.Error(err)
		return nil, err
	}

	return &claims.Id, nil
}

// Generates Access token which has the userID inside
func generateAccessToken(userID string) (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate an encrypter using RSA-OAEP with AES128-GCM. An error would
	// indicate that the selected algorithm(s) are not currently supported.
	publicKey := &privateKey.PublicKey
	encrypter, err := jose.NewEncrypter(jose.A128GCM, jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKey}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Preparing payload
	payload, err := json.Marshal(struct {
		usedID string
		tokenType string
	}{userID, accessTokenType})
	object, err := encrypter.Encrypt(payload)
	if err != nil {
		log.Fatal(err)
	}





	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims{
		userID,
		accessTokenType,
		jwt.StandardClaims{ExpiresAt: time.Now().Add(accessTokenExpiry).Unix()}})

	secretKey, err := getSecretKey()
	if err != nil {
		log.Error(err)
		return "", err
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims{
		uid,
		refreshTokenType,
		jwt.StandardClaims{ExpiresAt: time.Now().Add(refreshTokenExpiry).Unix()}})

	secretKey, err := getSecretKey()
	if err != nil {
		log.Error(err)
		return nil, "", err
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

