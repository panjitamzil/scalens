.PHONY: build test vet compose-up compose-down integration cluster-up preflight cpu database smoke

build:
	mkdir -p bin
	go build -o bin/dbys ./cmd/dbys
	go build -o bin/order-service ./cmd/order-service
	go build -o bin/payment-service ./cmd/payment-service

test:
	go test ./...

vet:
	go vet ./...

compose-up:
	docker compose -f deploy/docker/compose.yaml up -d --build

compose-down:
	docker compose -f deploy/docker/compose.yaml down

integration:
	TEST_DATABASE_URL='postgres://scalens:scalens@127.0.0.1:15432/scalens?sslmode=disable' go test ./tests/integration -count=1 -v

preflight:
	bash scripts/preflight.sh

cluster-up:
	bash scripts/cluster-up.sh

cpu: build
	bash scripts/run-suite.sh cpu

database: build
	bash scripts/run-suite.sh database

smoke: build
	bash scripts/run-suite.sh cpu results/smoke/cpu smoke
	bash scripts/run-suite.sh database results/smoke/database smoke
