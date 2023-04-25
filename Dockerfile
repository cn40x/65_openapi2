# Use the official golang image as the base image
# Use the official golang image as the base image
FROM golang:1.17-alpine

# Install Python and OpenJDK
RUN apk add --no-cache python3 openjdk11

# Set the working directory to /app
WORKDIR /app

# Copy the go.mod and go.sum files to the container
COPY go.mod .
COPY go.sum .

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application code to the container
COPY . .

# Build the Go binary
RUN go build -o app

# Expose port 8081 for the Gin server to listen on
# EXPOSE 8081

# Start the Gin server when the container starts
CMD ["./app"]

# RUN go mod download
# FROM golang:1.19-alpine3.16 AS builder

# RUN mkdir /app/
# WORKDIR /app

# COPY go.mod go.sum ./
# RUN go mod download
# COPY . /app

# RUN go build -o main .

# RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping
#Build small image

# FROM alpine:3.16
# WORKDIR /app
# COPY --from=builder /app/main .

# EXPOSE 8081
# ENV HOST=0.0.0.0
# CMD ["/app/main"]