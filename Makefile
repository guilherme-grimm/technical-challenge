BINARY  := bin/api
CMD     := ./cmd/api
GEN_DIR := internal/api/openapi

.PHONY: generate build test run clean

generate:
	rm -rf $(GEN_DIR)
	go generate ./...

build: generate
	go build -o $(BINARY) $(CMD)

test: generate
	go test ./...

run: build
	$(BINARY)

clean:
	rm -rf bin $(GEN_DIR)
