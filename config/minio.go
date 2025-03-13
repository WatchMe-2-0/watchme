package config

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

var MinioClient *minio.Client

func InitMinio() {
	endpoint := viper.GetString("MINIO_ENDPOINT")
	accessKeyID := viper.GetString("MINIO_ACCESSKEYID")
	secretAccessKey := viper.GetString("MINIO_SECRETACCESSKEY")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false, // Set to true if using HTTPS
	})

	if err != nil {
		fmt.Printf("❌ Failed to connect to MinIO: %v\n", err)
		log.Fatal(err)
	}

	MinioClient = client
	fmt.Println("✅ MinIO Connected Successfully")

	// Ensure required buckets exist and are public
	createBucket("movies")
	createBucket("posters")
}

// createBucket checks if a bucket exists; if not, it creates it and makes it public
func createBucket(bucketName string) {
	exists, err := MinioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Fatalf("❌ Failed to check bucket %s: %v", bucketName, err)
	}

	if !exists {
		err = MinioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			log.Fatalf("❌ Failed to create bucket %s: %v", bucketName, err)
		}
		fmt.Printf("✅ Bucket '%s' created successfully\n", bucketName)

		// Set public access policy for the bucket
		setPublicPolicy(bucketName)
	} else {
		fmt.Printf("✅ Bucket '%s' already exists\n", bucketName)
	}
}

// setPublicPolicy makes the bucket publicly accessible
func setPublicPolicy(bucketName string) {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": "s3:GetObject",
				"Resource": "arn:aws:s3:::%s/*"
			}
		]
	}`, bucketName)

	err := MinioClient.SetBucketPolicy(context.Background(), bucketName, policy)
	if err != nil {
		log.Fatalf("❌ Failed to set public policy for bucket %s: %v", bucketName, err)
	} else {
		fmt.Printf("🌍 Bucket '%s' is now publicly accessible\n", bucketName)
	}
}
