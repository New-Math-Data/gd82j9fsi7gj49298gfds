.PHONY: build test lint fmt vet clean run

build:
	-mkdir -p bin
	go build -o bin/opendss-assessment .

run: build
	bash scripts/run.sh

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	-rm -rf bin
