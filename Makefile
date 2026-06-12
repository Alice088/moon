run:
	go run ./cmd/api

tidy:
	go fmt ./... && echo "OK FMT" ; go vet ./... && echo "OK VET" 

test:
	go test ./...

lint:
	golangci-lint run

swagger:
	swag init -g cmd/api/main.go

migrate-up:
	migrate -path migrations -database $(DB_URL) up

migrate-down:
	migrate -path migrations -database $(DB_URL) down