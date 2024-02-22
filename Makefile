run_local:
	go run cmd/sso/main.go --config=./config/local.yaml

migrate:
	go run ./cmd/migrator --storage-dsn=postgres:postgres@localhost/elif_grpc --migrations-path=./migrations/postgres 
