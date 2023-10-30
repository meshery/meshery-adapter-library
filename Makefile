.PHONY: check
check: lint
check: tidy
check: verify


lint:
	golangci-lint run -c .golangci.yml -v ./...

tidy:
	go mod tidy

verify:
	go mod verify

test:
	go test --short ./... -race -coverprofile=coverage.txt -covermode=atomic
