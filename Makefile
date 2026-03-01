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
	docker build --platform linux/amd64 -t us-central1-docker.pkg.dev/burn-budgeter/burn-budgeter/api:latest .

docker-push:
	docker push us-central1-docker.pkg.dev/burn-budgeter/burn-budgeter/api:latest

docker-up:
	docker compose up -d
