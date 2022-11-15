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

.PHONY: build-all
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-386 build-linux-amd64 build-linux-arm64

build-darwin-amd64:
	GOOS="darwin" GOARCH="amd64" go build -ldflags="-X 'main.Version=$(VERSION)'" -o out/darwin_amd64/bin/gitzombie

build-darwin-arm64:
	GOOS="darwin" GOARCH="arm64" go build -ldflags="-X 'main.Version=$(VERSION)'" -o out/darwin_arm64/bin/gitzombie

build-linux-amd64:
	GOOS="linux" GOARCH="amd64" go build -ldflags="-X 'main.Version=$(VERSION)'" -o out/linux_amd64/bin/gitzombie

build-linux-arm64:
	GOOS="linux" GOARCH="arm64" go build -ldflags="-X 'main.Version=$(VERSION)'" -o out/linux_arm64/bin/gitzombie

build-linux-386:
	GOOS="linux" GOARCH="386" go build -ldflags="-X 'main.Version=$(VERSION)'" -o out/linux_386/bin/gitzombie
