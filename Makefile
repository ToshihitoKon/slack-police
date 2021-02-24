help:
	@cat Makefile | grep '^\w'

run:
	go run ./...

build:
	go build -o bin/slack-police ./...
