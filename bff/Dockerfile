FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY bff/go.mod bff/go.sum ./
RUN go mod download
COPY bff/ .
RUN go build -o bff ./cmd/bff

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bff .
COPY frontend ./frontend
EXPOSE 8080
CMD ["./bff"]
