GIT_HASH:=$(shell git rev-parse --short HEAD)
DIRTY:=$(shell test -z "`git status --porcelain`" || echo "-dirty")
VERSION:=$(GIT_HASH)$(DIRTY)
TIME:=$(shell date -u -Iseconds)

BIN:=big-switch
PACKAGE:=./cmd/big-switch

.PHONY: dev pi deps update-deps
dev: deps
	go build -ldflags "-X main.buildVersion=$(VERSION) -X main.buildTime=$(TIME)" -o $(BIN) $(PACKAGE)

update-deps:
	go get -u ./...
	go mod tidy

deps:
	go mod download

pi: deps
	GOOS=linux GOARCH=arm GOARM=5 go build -o $(BIN) -tags=pi,arm -ldflags "-X main.buildVersion=$(VERSION) -X main.buildTime=$(TIME)" $(PACKAGE)

run:
	go run $(PACKAGE) start