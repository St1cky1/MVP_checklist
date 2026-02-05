FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/server/main.go

# Build the seeder
RUN go build -o seed ./cmd/seed/main.go

# Build the migrator
RUN go build -o migrate ./cmd/migrate/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binaries
COPY --from=builder /app/main .
COPY --from=builder /app/seed .
COPY --from=builder /app/migrate .
# Copy templates and migrations
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/migrations ./migrations

# Copy start script
COPY scripts/start.sh ./start.sh
RUN chmod +x ./start.sh

# Expose port
EXPOSE 8080

# Command to run
CMD ["./start.sh"]
