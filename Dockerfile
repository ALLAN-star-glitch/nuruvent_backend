# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

# Install make and git
RUN apk add --no-cache make git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build both API and Worker binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o api ./cmd/api

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o worker ./cmd/worker

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binaries
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

# Copy configs
COPY --from=builder /app/configs ./configs

# Copy migrations
COPY --from=builder /app/internal/shared/database/migrations ./migrations

# Copy Makefile (for migrations)
COPY --from=builder /app/Makefile ./Makefile

# Copy goose binary
COPY --from=builder /go/bin/goose /usr/local/bin/goose

EXPOSE 8080

CMD ["./api"]