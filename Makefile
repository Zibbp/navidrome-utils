include .env
export $(shell sed 's/=.*//' .env)

run:
	go run cmd/main.go