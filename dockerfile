# Stage 1: Build the Go binary
FROM golang:1.23.6-alpine AS builder
ENV CGO_ENABLED=0
WORKDIR /app
COPY src/go.mod src/go.sum ./
RUN go mod download
COPY src/ .
RUN go build -o url-shortener .

# Stage 2: Create a lightweight final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/url-shortener .
COPY --from=builder /app/static ./static
EXPOSE 8080 9090

CMD ["./url-shortener"]
