OS=linux
BIN=bin
ARCH=amd64
TARGET=gateway
RELEASE_BASE=release
RELEASE_DIR=$(RELEASE_BASE)/$(TARGET)
INSTALL_TIME=$(shell date +%Y-%m-%d5%H:%M:%S)
RELEASE_VERSION=$(TARGET)_$(INSTALL_TIME)

$(TARGET): prebuilt
	#@GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BIN)/$@ src/gateway.go
	go build -o $(BIN)/$@ gateway.go

prebuilt:
	@mkdir -p $(BIN)

.PHONY: install
install:
	@mkdir -p $(RELEASE_DIR)
	@touch $(RELEASE_DIR)/$(RELEASE_VERSION)
	@cp $(BIN)/$(TARGET) $(RELEASE_DIR)/
	@cp start.sh $(RELEASE_DIR)/
	@cp configure/configure.json $(RELEASE_DIR)/

.PHONY: clean
clean:
	@rm -rf $(BIN)
	@rm -rf $(RELEASE_BASE)
