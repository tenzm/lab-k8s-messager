FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o message-service ./cmd/message-service

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/message-service .
COPY migrations ./migrations
RUN mkdir -p uploads
EXPOSE 8082
CMD ["./message-service"]
