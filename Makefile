.PHONY:

build:
	docker build --no-cache -t url-shortener .

up: build
	docker compose up --force-recreate --no-deps -d

down:
	docker compose down
