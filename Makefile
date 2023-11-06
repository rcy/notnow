include env.mk

start:
	air

up:
	docker-compose up -d

stop:
	docker-compose stop

down:
	docker-compose down
