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

	// Ensure required buckets exist
	createBucket("movies")  // For storing movie files
	createBucket("posters") // For storing movie posters
}

// createBucket checks if a bucket exists; if not, it creates it
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
	} else {
		fmt.Printf("✅ Bucket '%s' already exists\n", bucketName)
	}
}
