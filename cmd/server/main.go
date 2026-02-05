package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	s3Endpoint := os.Getenv("S3_ENDPOINT") // Для локальной разработки "http://localhost:4566"
	s3Region := os.Getenv("AWS_REGION")
	if s3Region == "" {
		s3Region = "us-east-1"
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if s3Endpoint != "" && service == s3.ServiceID {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           s3Endpoint,
				SigningRegion: s3Region,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	s3Config, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s3Region),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v\n", err)
	}
	s3Client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		if s3Endpoint != "" {
			o.UsePathStyle = true // Важно для LocalStack
		}
	})

	// 3. Infrastructure
	storage := infrastructure.NewS3Storage(s3Client)

	// 4. Repositories
	repo := repository.NewPostgresRepository(dbPool)

	// 5. UseCases
	templateUC := usecase.NewTemplateUseCase(repo)
	inspectionUC := usecase.NewInspectionUseCase(repo, storage)
	analyticsUC := usecase.NewAnalyticsUseCase(repo, storage)

	// 6. Delivery
	adminHandler := delivery.NewAdminHandler(templateUC, analyticsUC)
	publicHandler := delivery.NewPublicHandler(inspectionUC)

	// 7. Routing
	mux := http.NewServeMux()
	
	// Admin routes
	mux.Handle("/admin/", adminHandler)
	
	// Public routes (for inspectors)
	mux.Handle("/", publicHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v\n", err)
	}
}
