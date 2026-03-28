BINARY := goclip
GO := go

.PHONY: build run test clean install

build:
	mkdir -p bin
	$(GO) build -o bin/$(BINARY) .

run:
	$(GO) run .

test:
	$(GO) test ./...

clean:
	rm -rf bin

install:
	$(GO) install .
