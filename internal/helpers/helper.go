package helpers

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	database "github.com/sachinggsingh/notify/internal/db"
	"github.com/sachinggsingh/notify/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	brcypt "golang.org/x/crypto/bcrypt"
)

// user collection is not defined
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY = os.Getenv("SECRET")

func GenerateToken(email string, user_id string) models.TokenResponse {
	claims := &models.SignedDetails{
		Email:  email,
		UserID: user_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	refreshClaims := &models.SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token, tokenErr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if tokenErr != nil {
		return models.TokenResponse{
			Token:        "",
			RefreshToken: "",
			Err:          tokenErr,
		}
	}

	refreshToken, refreshTokenErr := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if refreshTokenErr != nil {
		return models.TokenResponse{
			Token:        "",
			RefreshToken: "",
			Err:          refreshTokenErr,
		}
	}

	return models.TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
		Err:          tokenErr,
	}
}

func UpdateToken(signedToken string, signedRefreshtoken string, user_id string) models.TokenResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"user_id": user_id}
	update := bson.M{
		"$set": bson.M{
			"token":         signedToken,
			"refresh_token": signedRefreshtoken,
		},
	}
	upsert := true

	opts := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(ctx, filter, update, &opts)
	if err != nil {
		return models.TokenResponse{
			Token:        "",
			RefreshToken: "",
			Err:          err,
		}
	}

	return models.TokenResponse{
		Token:        signedToken,
		RefreshToken: signedRefreshtoken,
		Err:          nil,
	}
}

func ValidateToken(signedToken string) (*models.SignedDetails, error) {
	token, err := jwt.ParseWithClaims(signedToken, &models.SignedDetails{}, func(token *jwt.Token) (any, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.SignedDetails)
	if !ok {
		return nil, errors.New("could not parse claims")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := brcypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	check := true
	msg := ""
	if err != nil {
		check = false
		msg = "login or password is incorrect"
	}
	return check, msg
}

func HashPassword(password string) (string, error) {
	bytes, err := brcypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes), nil
}
