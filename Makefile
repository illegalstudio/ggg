.PHONY: build install test fmt vet clean

BINARY := ggg
PKG    := ./cmd/ggg

build:
	go build -o $(BINARY) $(PKG)

install:
	go install $(PKG)

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
