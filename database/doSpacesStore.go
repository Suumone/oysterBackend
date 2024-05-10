package database

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

const (
	BucketName         = "oyster"
	ProfilePicturePath = "profilePictures"
)

func UploadProfilePictureToDigitalOceanSpaces(destFilePath string, fileBytes *[]byte) error {
	s3Input := s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(destFilePath),
		Body:   bytes.NewReader(*fileBytes),
	}
	_, err := S3Client.PutObject(&s3Input)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
