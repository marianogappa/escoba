.PHONY: test build run release lint

test:
	go test -v ./...

build:
	go build -o escoba-game .

run:
	./escoba-game

release:
	rm -rf dist && goreleaser

lint:
	golangci-lint run
