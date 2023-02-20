run:
	go run internal/main.go

generate-api:
	go run bin/gonzo-api.go generate --input test --output server

db-migration:
	go run db/db-migrations.go

new-db-migration:
	migrate create -ext sql -dir db/migrations -seq 