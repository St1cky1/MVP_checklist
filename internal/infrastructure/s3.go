package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client *s3.Client
}

func NewS3Storage(client *s3.Client) *S3Storage {
	return &S3Storage{client: client}
}

func (s *S3Storage) Upload(ctx context.Context, bucket, key string, data []byte) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to s3: %w", err)
	}
	return key, nil
}

func (s *S3Storage) GetURL(ctx context.Context, bucket, key string) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	presignedUrl, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(time.Hour*24))
	
	if err != nil {
		return "", fmt.Errorf("failed to presign s3 url: %w", err)
	}
	return presignedUrl.URL, nil
}
