.PHONY:

unit_tests:
	cd src && go test -v ./...

build:
	docker build --no-cache -t url-shortener .

up: build
	docker compose up --force-recreate --no-deps -d

integration_tests: up
	sleep 5
	cd src && go test -tags=integration ./integrationtests

down:
	docker compose down
