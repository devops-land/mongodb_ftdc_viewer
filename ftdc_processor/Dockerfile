# step 1: build ftdc_decoder

FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ftdc_processor .

# end of building ftdc_processor

FROM alpine:3.18

# Set the working directory
WORKDIR /app

# Copy the necessary files into the container
COPY --from=builder /app/ftdc_processor /app/


# Make sure the scripts and binaries have the correct permissions
RUN chmod +x /app/ftdc_processor
