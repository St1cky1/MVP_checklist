FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git if needed for dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Check structure (useful for logs if build fails)
RUN ls -R

# Build the application
RUN go build -o main ./cmd/server/main.go

# Build the seeder
RUN go build -o seed ./cmd/seed/main.go

# Build the migrator
RUN go build -o migrate ./cmd/migrate/main.go

# Final stage
FROM alpine:latest

# Install CA certificates for S3/AWS
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy the binaries
COPY --from=builder /app/main .
COPY --from=builder /app/seed .
COPY --from=builder /app/migrate .
# Copy templates and migrations
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/migrations ./migrations

# Create uploads directory
RUN mkdir -p uploads

# Copy start script
COPY scripts/start.sh ./start.sh
RUN chmod +x ./start.sh

# Expose port
EXPOSE 8080

# Command to run
CMD ["./start.sh"]
