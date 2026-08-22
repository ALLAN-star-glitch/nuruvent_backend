# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

# Install tools
RUN apk add --no-cache make git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Swagger docs before building
RUN swag init -g cmd/api/main.go -o docs

# Build the API
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o api ./cmd/api

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/internal/shared/database/migrations ./migrations
COPY --from=builder /app/docs ./docs  
COPY --from=builder /go/bin/goose /usr/local/bin/goose

EXPOSE 8080

CMD ["./api"]