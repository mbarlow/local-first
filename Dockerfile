# Multi-stage build for Go WASM and Node development
FROM golang:1.21-alpine AS go-base

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Node stage for frontend development
FROM node:20-alpine AS node-base

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci || npm install

# Development stage combining both
FROM golang:1.21-alpine AS development

# Install Node.js and npm
RUN apk add --no-cache nodejs npm git make

WORKDIR /app

# Copy everything
COPY . .

# Install Node dependencies
RUN npm install

# Build WASM
RUN GOOS=js GOARCH=wasm go build -o web/main.wasm cmd/wasm/main.go

# Copy wasm_exec.js from Go installation
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/

# Expose ports
EXPOSE 5173 8080

# Default command for development
CMD ["npm", "run", "dev"]

# Production build stage
FROM golang:1.21-alpine AS wasm-builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build optimized WASM
RUN GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/main.wasm cmd/wasm/main.go
RUN cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" web/

# Build Go server
RUN go build -o server cmd/server/main.go

# Final production stage
FROM alpine:latest AS production

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy built artifacts
COPY --from=wasm-builder /app/server ./
COPY --from=wasm-builder /app/web ./web

EXPOSE 8080

CMD ["./server"]