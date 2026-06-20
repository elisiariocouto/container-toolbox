BINARY := container-toolbox

.PHONY: build run test fmt vet tidy clean install

build:
	go build -o bin/$(BINARY) .

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin

install:
	go install .
