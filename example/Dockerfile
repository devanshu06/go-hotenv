# Dockerfile
# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download


RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
# COPY --from=builder /app/secrets/.env ./secrets/.env

ENV PORT=8080
# ENV CONFIG_FILE=/app/secrets/.env

EXPOSE 8080

CMD ["./main"]⏎