postgres15:
	docker run --name postgres15.1 -p 5432:5432 -e POSTGRES_USER=root -e  POSTGRES_PASSWORD=secret -d postgres:15.1-alpine
createdb:
	docker exec -it postgres15.1 createdb --username=root --owner=root simple_bank
dropdb:
	docker exec -it postgres15.1 dropdb simple_bank
startContainer:
	docker start postgres15.1
stopContainer:
	docker stop postgres15.1
migrateup:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
sqlc:
	sqlc generate
test:
	go test -v -cover ./...

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test


# migrate create -ext sql -dir db/migration -seq init_schema