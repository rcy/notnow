include ../env.mk

status:
	~/go/bin/tern status -m ./migrations

create:
	psql ${DATABASE_ROOT_URL}/template1 -c "create database ${PGDATABASE};"

drop:
	psql ${DATABASE_ROOT_URL}/template1 -c "drop database if exists ${PGDATABASE};"

up:
	~/go/bin/tern migrate -m ./migrations
	${MAKE} schema

down:
	~/go/bin/tern migrate -m ./migrations -d -1

reset: drop create up

new:
	~/go/bin/tern new -m ./migrations unnamed

sql:
	psql ${DATABASE_URL}

dump:
	docker exec -it yikes_postgres_1 pg_dump --dbname=${PGDATABASE} --user=postgres --table=public.* --schema-only

schema:
	${MAKE} -s dump > schema.sql

generate:
	sqlc generate
