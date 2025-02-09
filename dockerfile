# Stage 1: Build the Go binary
FROM golang:1.23.6-alpine AS builder
ENV CGO_ENABLED=0 GOOS=linux GOARCH=arm64
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/ .
RUN go build -o url-shortener .

# Stage 2: Create a lightweight final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/url-shortener .
EXPOSE 8080

CMD ["./url-shortener"]
