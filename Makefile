#=================================#
#===   ENVIRONMENT VARIABLES   ===#
#=================================#

UID := $(shell id -u)
GID := $(shell id -g)

#=================================#
#===           BUILD           ===#
#=================================#

up:
	docker compose up -d --build
watch:
	docker compose up --build
down:
	docker compose down

#=================================#
#===    DATABASE MIGRATIONS    ===#
#=================================#

migrate-up:
	docker compose --profile tools run --rm migrate up

migrate-down:
	docker compose --profile tools run --rm migrate down

migrate-create:
	docker compose --profile tools run --rm --user $(UID):$(GID) migrate create -ext sql -dir ./migrations -seq $(name)

#=================================#
#===    CONNECT TO DATABASE    ===#
#=================================#

db-shell:
	docker compose exec database psql -U $(user) -d $(db)

db-exec:
	docker compose exec database psql -U $(user) -d $(db) -c "$(sql)"
