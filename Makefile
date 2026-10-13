MIHOMO_VERSION := $(shell git -C core/mihomo describe --tags 2>/dev/null)
VERSION        ?= $(shell git describe --tags --always --dirty 2>/dev/null)
LDFLAGS := -s -w \
	-X github.com/ZeroMarker/clashon/cli/cmd.Version=$(VERSION) \
	-X github.com/metacubex/mihomo/constant.Version=$(MIHOMO_VERSION)
TAGS := with_gvisor

.PHONY: build test vet
build:
	go build -tags $(TAGS) -trimpath -ldflags '$(LDFLAGS)' -o bin/ ./cli/cmd/clashon ./tui/cmd/clashon-tui

test:
	go test ./...

vet:
	go vet ./...
