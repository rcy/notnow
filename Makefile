include env.mk

start:
	air

build:
	go build -o ./tmp/main .

up:
	docker compose up -d

stop:
	docker compose stop

down:
	docker compose down
