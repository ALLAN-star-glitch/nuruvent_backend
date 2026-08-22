# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

# Install make and goose for migrations
RUN apk add --no-cache make git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o main ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary
COPY --from=builder /app/main .

# Copy configs
COPY --from=builder /app/configs ./configs

# Copy migrations
COPY --from=builder /app/internal/shared/database/migrations ./migrations

# Copy Makefile (for migrations)
COPY --from=builder /app/Makefile ./Makefile

# Copy goose binary
COPY --from=builder /go/bin/goose /usr/local/bin/goose

EXPOSE 8080

CMD ["./main"]