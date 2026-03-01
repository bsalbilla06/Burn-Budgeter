.PHONY: build run test clean docker-build docker-up

build:
	go build -o bin/api cmd/api/main.go

run:
	go run cmd/api/main.go

test:
	go test ./...

clean:
	rm -rf bin/

docker-build:
	docker compose build

docker-up:
	docker compose up -d
