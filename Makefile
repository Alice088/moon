VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_DIR := build

.PHONY: build dist clean

build:
	@mkdir -p $(BUILD_DIR)/moon
	go build -trimpath -ldflags="-w -X main.version=$(VERSION)" -o $(BUILD_DIR)/moon/moon ./cmd/moon/
	cp config.example.yaml $(BUILD_DIR)/moon/
	@echo "binary:"
	@ls -lh $(BUILD_DIR)/moon/moon

dist: build
	@mkdir -p $(BUILD_DIR)/release
	tar czf $(BUILD_DIR)/release/moon-$(VERSION)-linux-amd64.tar.gz \
		-C $(BUILD_DIR) moon
	@echo "release:"
	@ls -lh $(BUILD_DIR)/release/

clean:
	rm -rf $(BUILD_DIR)
