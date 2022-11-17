COMMIT=$(shell git log -1 --pretty=format:"%h")
TAG=$(shell git describe --tags --abbrev=0 2>/dev/null)

VERSION=$(if $(TAG),$(TAG),commit-$(COMMIT))

.PHONY: fmt
fmt:
	@find . -name \*.go -exec goimports -w {} \;

.PHONY: install
install:
	@go install -ldflags="-X 'main.Version=$(VERSION)'"

.PHONY: build
build:
	@go build -ldflags="-X 'main.Version=$(VERSION)'"
