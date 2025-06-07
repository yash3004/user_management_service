.PHONY: build run clean test

build:
	go build -o server ./cmd/server

clean:
	rm -f server

test:
	go test ./...

dev: build
	./server