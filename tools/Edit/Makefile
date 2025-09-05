BINARY_NAME := Edit
OUTPUT_DIR := ../../bin

.PHONY: build clean force-build

build: $(OUTPUT_DIR)/$(BINARY_NAME)

$(OUTPUT_DIR)/$(BINARY_NAME):
	mkdir -p $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd

clean:
	rm -f $(OUTPUT_DIR)/$(BINARY_NAME)

force-build:
	$(MAKE) clean
	$(MAKE) build