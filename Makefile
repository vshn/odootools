.PHONY: test
test:
	go test -v -race ./...

lint:
	golangci-lint run ./...

format:
	go fmt ./...