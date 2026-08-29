PROJECT_NAME := "jupiter"
PKG := "github.com/weijinwen/jupiter"

.DEFAULT_GOAL := help
.PHONY: help unit-test covsh-unit golint

# Run unit tests
unit-test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# View unit test coverage
covsh-unit:
	gocovsh --profile coverage.txt

# Lint go files (requires golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golint:
	golangci-lint run -v

# Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
