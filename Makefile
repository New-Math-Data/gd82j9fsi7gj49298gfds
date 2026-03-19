.PHONY: build test lint fmt vet clean

build:
	go build -o opendss-assessment .

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

clean:
	rm -f opendss-assessment opendss-assessment.exe
