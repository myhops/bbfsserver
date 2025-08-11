# Stage 1: Build stage
FROM cir-cn-devops.chp.belastingdienst.nl/obp-pnr/golang:1.24-alpine AS build

COPY certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Set the working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o bbfsserver ./cmd/bbfsserver

# Stage 2: Final stage
FROM cir-cn-cpet.chp.belastingdienst.nl/external/docker.io/alpine:3.22.1

# Set the working directory
WORKDIR /app

# Copy the certs
COPY certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the binary from the build stage
COPY --from=build /app/bbfsserver .

# Set the timezone and install CA certificates
RUN apk --no-cache add tzdata

# Set the entrypoint command
ENTRYPOINT ["/app/bbfsserver"]