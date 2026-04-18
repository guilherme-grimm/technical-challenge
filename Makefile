BINARY := bin/api
CMD    := ./cmd/api

.PHONY: generate build test run clean

generate:
	go generate ./...

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

run: build
	$(BINARY)

clean:
	rm -rf bin
