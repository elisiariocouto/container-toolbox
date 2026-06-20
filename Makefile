BINARY := container-toolbox

.PHONY: build build-linux-amd64 run test fmt vet tidy clean install

build:
	go build -o bin/$(BINARY) .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY)-linux-amd64 .

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
