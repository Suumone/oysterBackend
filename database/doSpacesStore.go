package database

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"os"
)

const (
	ProfilePicturePath   = "profilePictures"
	ACLForProfilePicture = "public-read"
)

var (
	ProfilePicturePathPrefix = os.Getenv("DO_CDN_ENDPOINT")
	BucketName               = os.Getenv("DO_BUCKET_NAME")
)

func UploadProfilePictureToDigitalOceanSpaces(destFilePath string, fileBytes []byte) error {
	s3Input := &s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(destFilePath),
		Body:   bytes.NewReader(fileBytes),
	}
	if _, err := S3Client.PutObject(s3Input); err != nil {
		log.Println("UploadProfilePictureToDigitalOceanSpaces: failed to upload to s3:", err)
		return err
	}

	s3InputACL := s3.PutObjectAclInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(destFilePath),
		ACL:    aws.String(ACLForProfilePicture),
	}
	if _, err := S3Client.PutObjectAcl(&s3InputACL); err != nil {
		log.Println("UploadProfilePictureToDigitalOceanSpaces: failed to upload acl:", err)
		return err
	}

	return nil
}
