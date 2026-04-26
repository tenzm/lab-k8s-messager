run:
	go run ./cmd/message-service/...

test:
	go test ./...

migrate:
	goose -dir migrations postgres "$(DB_DSN)" up

.PHONY: run test migrate
