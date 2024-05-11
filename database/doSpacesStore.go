package database

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"os"
)

const (
	BucketName         = "oyster"
	ProfilePicturePath = "profilePictures"
)

var (
	ProfilePicturePathPrefix = "https://oyster." + os.Getenv("DO_ENDPOINT")
)

func UploadProfilePictureToDigitalOceanSpaces(destFilePath string, fileBytes []byte) error {
	s3Input := &s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(destFilePath),
		Body:   bytes.NewReader(fileBytes),
	}
	if _, err := S3Client.PutObject(s3Input); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
