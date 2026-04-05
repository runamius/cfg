.PHONY: up down build logs seed

up:
	docker-compose up --build

down:
	docker-compose down

build:
	docker-compose build

logs:
	docker-compose logs -f

seed:
	docker-compose exec db psql -U postgres -d booking -f /schema/seed.sql
test:
	cd backend && go test ./test/unit/... -v

test-coverage:
	cd backend && go test ./test/unit/... -v -coverpkg=./... -coverprofile=coverage.out
	cd backend && go tool cover -func=coverage.out | grep total
