package config

import (
	"context"
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go"
)

var Cld *cloudinary.Cloudinary
var Ctx context.Context

func InitCloudinary() {
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	var err error
	Cld, err = cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	Ctx = context.Background()
}
