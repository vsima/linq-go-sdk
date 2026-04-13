.PHONY: all test cover lint fmt vet tidy bench ci

all: fmt vet test

test:
	go test -race -timeout 60s ./...

cover:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -func=coverage.txt | tail -n 1

cover-html: cover
	go tool cover -html=coverage.txt -o coverage.html

lint: fmt vet

fmt:
	@diff=$$(gofmt -l .); if [ -n "$$diff" ]; then echo "gofmt needed:"; echo "$$diff"; exit 1; fi

vet:
	go vet ./...

tidy:
	go mod tidy

bench:
	go test -bench=. -benchmem ./...

ci: tidy fmt vet test
