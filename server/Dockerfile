# Build stage
FROM golang:1.22-alpine AS build

WORKDIR /app

# Enable Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the backend source
COPY . .

# Build the API binary
RUN go build -o chainsight-api ./cmd/chainsight-api

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Minimal runtime deps
RUN apk add --no-cache ca-certificates

# Copy built binary
COPY --from=build /app/chainsight-api ./chainsight-api

ENV PORT=8080
EXPOSE 8080

CMD ["./chainsight-api"]
