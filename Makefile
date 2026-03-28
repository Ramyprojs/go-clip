BINARY := goclip
GO := go

.PHONY: build cross-build run test clean install

build:
	mkdir -p bin
	$(GO) build -o bin/$(BINARY) .

cross-build:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GO) build -o dist/$(BINARY)_linux_amd64 .
	GOOS=linux GOARCH=arm64 $(GO) build -o dist/$(BINARY)_linux_arm64 .
	GOOS=darwin GOARCH=amd64 $(GO) build -o dist/$(BINARY)_darwin_amd64 .
	GOOS=darwin GOARCH=arm64 $(GO) build -o dist/$(BINARY)_darwin_arm64 .
	GOOS=windows GOARCH=amd64 $(GO) build -o dist/$(BINARY)_windows_amd64.exe .
	GOOS=windows GOARCH=arm64 $(GO) build -o dist/$(BINARY)_windows_arm64.exe .

run:
	$(GO) run .

test:
	$(GO) test ./...

clean:
	rm -rf bin dist

install:
	$(GO) install .
