COMMIT=$(shell git rev-parse HEAD)
TAG=$(shell git describe --tags --abbrev=0 2>/dev/null)

DATE=$(shell date '+%FT%TZ')

VERSION=$(if $(TAG),$(TAG),commit-$(COMMIT))

.PHONY: fmt
fmt:
	@find . -name \*.go -exec goimports -w {} \;

.PHONY: install
install:
	@go install -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.BuildDate=$(DATE)'"

.PHONY: build
build:
	@go build -ldflags="-X 'main.Version=$(VERSION)' -X 'main.Commit=$(COMMIT)' -X 'main.BuildDate=$(DATE)'" -o bin/gitzombie


.PHONY: install-check
install-check:
	@go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@go install github.com/client9/misspell/cmd/misspell@latest
	@go install github.com/gordonklaus/ineffassign@latest
	@go install golang.org/x/tools/cmd/goimports@latest

.PHONY: check
check:
	@echo "==> check ineffassign"
	@ineffassign ./...
	@echo "==> check spell"
	@find . -type f -name '*.go' | xargs misspell -error
	@echo "==> check gocyclo"
	@gocyclo -over 15 .
	@echo "==> go vet"
	@go vet ./...
	@echo "==> check goimports"
	@bash hack/check-fmt.sh
