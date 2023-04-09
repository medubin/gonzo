run:
	go run internal/main.go

generate-api:
	go run api/bin/gonzo-api.go generate --input server --output server

db-migration:
	go run db/db-migrations.go

db-migration-down:
	migrate -database postgres://postgres:postgres@localhost:5432/gonzo?sslmode=disable -path db/migrations down

new-db-migration:
	migrate create -ext sql -dir db/migrations -seq ${name}

generate-sqlc:
	cd db && sqlc generate