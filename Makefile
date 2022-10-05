check: error
	golangci-lint run

check-clean-cache:
	golangci-lint cache clean

error:
	go run github.com/layer5io/meshkit/cmd/errorutil -d . analyze -i ./helpers -o ./helpers