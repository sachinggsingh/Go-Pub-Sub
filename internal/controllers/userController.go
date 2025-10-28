package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	database "github.com/sachinggsingh/notify/internal/db"
	helper "github.com/sachinggsingh/notify/internal/helpers"
	"github.com/sachinggsingh/notify/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func CreateUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if user.Email == nil || user.Password == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
			return
		}

		validate := validator.New()
		if err := validate.Struct(user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while checking existing email"})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}

		hashedPassword, err := helper.HashPassword(*user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = &hashedPassword

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		tokenResponse := helper.GenerateToken(*user.Email, user.User_id)
		if tokenResponse.Err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": tokenResponse.Err.Error()})
			return
		}

		user.Token = &tokenResponse.Token
		user.Refresh_Token = &tokenResponse.RefreshToken

		result, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User created successfully", "data": result})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var loggedDetails models.User
		var foundUser models.User

		if err := c.BindJSON(&loggedDetails); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// finding user
		err := userCollection.FindOne(ctx, bson.M{"email": loggedDetails.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
			return
		}

		passwordIsValid, msg := helper.VerifyPassword(*foundUser.Password, *loggedDetails.Password)
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		tokenResponse := helper.GenerateToken(*foundUser.Email, foundUser.User_id)
		if tokenResponse.Err != nil {
			c.JSON(http.StatusInternalServerError, bson.M{"error": "Internal server error"})
			return
		}
		if tokenResponse := helper.UpdateToken(tokenResponse.Token, tokenResponse.RefreshToken, foundUser.User_id); tokenResponse.Err != nil {
			c.JSON(http.StatusInternalServerError, bson.M{"error": "Internal server error"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, tokenResponse)
	}
}
