run:
	go run internal/main.go

generate-api:
	go run bin/gonzo-api.go generate --input test --output server