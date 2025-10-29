package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/gin-gonic/gin"
	config "github.com/sachinggsingh/notify/internal/config"
	database "github.com/sachinggsingh/notify/internal/db"
	"github.com/sachinggsingh/notify/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var uploadImageCollection *mongo.Collection = database.OpenCollection(database.Client, "image")

func UploadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
			return
		}

		// Open the file
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image file"})
			return
		}
		defer f.Close()

		publicID := c.DefaultPostForm("public_id", "")

		// Upload to Cloudinary
		uploadResult, err := config.Cld.Upload.Upload(ctx, f, uploader.UploadParams{
			PublicID: publicID,
			Folder:   "notify",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cloudinary upload error: " + err.Error()})
			return
		}

		var image models.Image
		image.ID = primitive.NewObjectID()
		image.Image_id = image.ID.Hex()

		imageDoc := models.Image{
			ID:       image.ID,
			Uploaded: time.Now(),
			URL:      uploadResult.SecureURL,
			PublicID: uploadResult.PublicID,
			Image_id: image.Image_id,
		}

		result, err := uploadImageCollection.InsertOne(ctx, imageDoc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save upload details"})
			return
		}

		go func() {
			genCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			caption, err := config.GenerateCaption(genCtx, imageDoc.URL)

			if err != nil {
				caption = "A new file has been uploaded"
			}
			msg := models.Message{
				UserID:    c.PostForm("user_id"), // if user_id is submitted in the form
				Content:   caption,
				Timestamp: time.Now(),
			}
			// store message in DB and publish to Redis so subscribers receive it
			_, _ = messageCollection.InsertOne(genCtx, msg) // messageCollection is in package controllers (pub_sub_controller.go)
			data, _ := json.Marshal(msg)
			_ = config.RDB.Publish(genCtx, "message", data).Err()
		}()

		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
			"data": gin.H{
				"id":        result.InsertedID,
				"url":       imageDoc.URL,
				"public_id": imageDoc.PublicID,
				"image_id":  imageDoc.Image_id,
			},
		})
	}
}

// GetImage retrieves an image record by public_id (image_id) from MongoDB
func GetImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		imageID := c.Param("image_id")
		if imageID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_id param is required"})
			return
		}

		var image models.Image
		err := uploadImageCollection.FindOne(ctx, bson.M{"image_id": imageID}).Decode(&image)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"image": image})
	}
}

// image_id mein error aa rha hai toh uslo kal dekhna
