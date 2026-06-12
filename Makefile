VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DIR := build
LDFLAGS := -ldflags="-s -w"

.PHONY: build dist clean

build:
	@mkdir -p $(BUILD_DIR)/moon-$(VERSION)
	go build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/moon-$(VERSION)/moon ./cmd/moon/
	go build -trimpath $(LDFLAGS) -o $(BUILD_DIR)/moon-$(VERSION)/moon-installer ./cmd/installer/
	strip $(BUILD_DIR)/moon-$(VERSION)/moon $(BUILD_DIR)/moon-$(VERSION)/moon-installer 2>/dev/null || true
	cp config.example.yaml $(BUILD_DIR)/moon-$(VERSION)/
	cp -r static $(BUILD_DIR)/moon-$(VERSION)/
	@echo "binaries:"
	@ls -lh $(BUILD_DIR)/moon-$(VERSION)/moon $(BUILD_DIR)/moon-$(VERSION)/moon-installer

dist: build
	@mkdir -p $(BUILD_DIR)/release
	tar czf $(BUILD_DIR)/release/moon-$(VERSION)-linux-amd64.tar.gz \
		-C $(BUILD_DIR) moon-$(VERSION)
	@echo "release:"
	@ls -lh $(BUILD_DIR)/release/

clean:
	rm -rf $(BUILD_DIR)
