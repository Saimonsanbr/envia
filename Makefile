.PHONY: run build test fmt clean

BIN=bin/envia

run:
	go run ./cmd/envia

build:
	mkdir -p bin
	go build -trimpath -ldflags "-s -w" -o $(BIN) ./cmd/envia
	@echo "built $(BIN)"

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin/ dist/
	go clean

