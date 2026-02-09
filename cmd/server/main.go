package main

import (
	"MVP_checklist/internal/domain"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"MVP_checklist/internal/delivery"
	"MVP_checklist/internal/infrastructure"
	"MVP_checklist/internal/repository"
	"MVP_checklist/internal/usecase"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Set Moscow timezone
	time.Local = time.FixedZone("MSK", 3*60*60)

	ctx := context.Background()

	// 1. Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/checklist_db"
	}
	dbPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// 2. S3 client
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3Region := os.Getenv("AWS_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}
	s3Key := os.Getenv("AWS_ACCESS_KEY_ID")
	s3Secret := os.Getenv("AWS_SECRET_ACCESS_KEY")

	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(s3Region))

	// Если есть эндпоинт (LocalStack)
	if s3Endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if service == s3.ServiceID {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           s3Endpoint,
					SigningRegion: s3Region,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})
		opts = append(opts, config.WithEndpointResolverWithOptions(customResolver))
	}

	// Если ключи не заданы и мы не в LocalStack, используем "пустые" ключи для предотвращения поиска IMDS роли
	// Либо, если ключи заданы, используем их
	if s3Key != "" && s3Secret != "" {
		opts = append(opts, config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     s3Key,
				SecretAccessKey: s3Secret,
			}, nil
		})))
	} else if s3Endpoint != "" {
		// Для LocalStack часто подходят любые заглушки
		opts = append(opts, config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}, nil
		})))
	}

	s3Config, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v\n", err)
	}

	s3Client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		if s3Endpoint != "" {
			o.UsePathStyle = true
		}
	})

	// 3. Infrastructure
	var storage domain.FileStorage
	if s3Key == "" && s3Endpoint == "" {
		fmt.Println("Using FileSystem storage (uploads/ folder)")
		storage = infrastructure.NewFileSystemStorage("uploads")
	} else {
		bucketName := os.Getenv("S3_BUCKET_NAME")
		if bucketName == "" {
			bucketName = "checklist-photos" // default
		}
		fmt.Printf("Using S3 storage (bucket: %s)\n", bucketName)
		storage = infrastructure.NewS3Storage(s3Client, bucketName)
	}

	// 4. Repositories
	repo := repository.NewPostgresRepository(dbPool)

	// 5. UseCases
	templateUC := usecase.NewTemplateUseCase(repo)
	inspectionUC := usecase.NewInspectionUseCase(repo, storage)
	analyticsUC := usecase.NewAnalyticsUseCase(repo, storage)
	ocrUC := usecase.NewOCRUseCase()

	// 6. Delivery
	adminHandler := delivery.NewAdminHandler(templateUC, analyticsUC)
	publicHandler := delivery.NewPublicHandler(inspectionUC, ocrUC)

	// 7. Routing
	mux := http.NewServeMux()

	// Admin routes
	mux.Handle("/admin/", adminHandler)

	// Public routes (for inspectors)
	mux.Handle("/", publicHandler)

	// Static uploads route (for FileSystem storage)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v\n", err)
	}
}
