FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies for CGO and tesseract
RUN apk add --no-cache git gcc g++ musl-dev tesseract-ocr-dev leptonica-dev

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the application with CGO enabled for gosseract
RUN CGO_ENABLED=1 go build -o main ./cmd/server/main.go

# Build the seeder
RUN CGO_ENABLED=1 go build -o seed ./cmd/seed/main.go

# Build the migrator
RUN CGO_ENABLED=1 go build -o migrate ./cmd/migrate/main.go

# Final stage
FROM alpine:latest

# Install CA certificates and tesseract runtime dependencies
RUN apk add --no-cache ca-certificates tesseract-ocr leptonica

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
