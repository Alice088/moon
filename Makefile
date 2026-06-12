VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DIR := build

.PHONY: build dist clean

build:
	@mkdir -p $(BUILD_DIR)/moon
	go build -trimpath -ldflags="-w -X main.version=$(VERSION)" -o $(BUILD_DIR)/moon/moon ./cmd/moon/
	go build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/moon/moon-installer ./cmd/installer/
	strip $(BUILD_DIR)/moon/moon-installer 2>/dev/null || true
	cp config.example.yaml $(BUILD_DIR)/moon/
	cp -r static $(BUILD_DIR)/moon/
	@echo "binaries:"
	@ls -lh $(BUILD_DIR)/moon/moon $(BUILD_DIR)/moon/moon-installer

dist: build
	@mkdir -p $(BUILD_DIR)/release
	tar czf $(BUILD_DIR)/release/moon-$(VERSION)-linux-amd64.tar.gz \
		-C $(BUILD_DIR) moon
	@echo "release:"
	@ls -lh $(BUILD_DIR)/release/

clean:
	rm -rf $(BUILD_DIR)
