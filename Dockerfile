FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/taas-api ./cmd/taas-api

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

COPY --from=builder /app/taas-api /app/taas-api
COPY --from=builder /app/migrations /app/migrations
COPY .env.example /app/.env.example

EXPOSE 8080

CMD ["/app/taas-api"]
