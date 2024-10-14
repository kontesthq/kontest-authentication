# Use the official Go image as a build stage
FROM golang:1.23.2-alpine AS builder

# Install dependencies required for building with CGO and musl
RUN apk --no-cache add git gcc libc-dev

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project directory into the working directory
COPY . ./

# Build the Go application with CGO enabled and musl tag for librdkafka
ARG GIT_TAG_NAME
RUN CGO_ENABLED=1 GOOS=linux go build -tags musl -ldflags "-s -w -X main.Version=$GIT_TAG_NAME" -o /docker-gs-ping ./main.go

# Start a new scratch image for a smaller final image
FROM alpine:latest

# Copy the binary from the builder image
COPY --from=builder /docker-gs-ping /docker-gs-ping

# Set the command to run the binary
CMD ["/docker-gs-ping"]
