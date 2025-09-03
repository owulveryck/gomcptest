.PHONY: all clean tools servers install run

# Build everything
all: tools servers

# Define the bin directory
BIN_DIR := bin

# Tools to build
TOOLS := LS GrepTool Edit GlobTool Replace View duckdbserver dispatch_agent Bash imagen imagen_edit plantuml_check sleep

# Servers to build
SERVERS := cliGCP openaiserver

# Default install directory (can be overridden via command line or environment)
INSTALL_DIR ?= ~/openaiserver

# Ensure the bin directory exists
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# Build all tools
tools: $(BIN_DIR) $(addprefix $(BIN_DIR)/, $(TOOLS))

# Build all servers
servers: $(BIN_DIR) $(addprefix $(BIN_DIR)/, $(SERVERS)) 

# Install binaries to target directory
# Usage: make install INSTALL_DIR=/path/to/install
install: all
	mkdir -p $(INSTALL_DIR)/bin
	cp $(BIN_DIR)/cliGCP $(BIN_DIR)/openaiserver $(INSTALL_DIR)
	cp $(addprefix $(BIN_DIR)/, $(TOOLS)) $(INSTALL_DIR)/bin

# Special case for tools with main.go in the root directory

$(BIN_DIR)/duckdbserver: tools/duckdbserver/main.go
	go build -o $(BIN_DIR)/duckdbserver ./tools/duckdbserver

# Rule for tools with main.go in cmd/ subdirectory
$(BIN_DIR)/LS: tools/LS/cmd/main.go
	go build -o $(BIN_DIR)/LS ./tools/LS/cmd

$(BIN_DIR)/GrepTool: tools/GrepTool/cmd/main.go
	go build -o $(BIN_DIR)/GrepTool ./tools/GrepTool/cmd

$(BIN_DIR)/Edit: tools/Edit/cmd/main.go
	go build -o $(BIN_DIR)/Edit ./tools/Edit/cmd

$(BIN_DIR)/GlobTool: tools/GlobTool/cmd/main.go
	go build -o $(BIN_DIR)/GlobTool ./tools/GlobTool/cmd

$(BIN_DIR)/Replace: tools/Replace/cmd/main.go
	go build -o $(BIN_DIR)/Replace ./tools/Replace/cmd

$(BIN_DIR)/View: tools/View/cmd/main.go
	go build -o $(BIN_DIR)/View ./tools/View/cmd

$(BIN_DIR)/dispatch_agent: tools/dispatch_agent/cmd/main.go
	go build -o $(BIN_DIR)/dispatch_agent ./tools/dispatch_agent/cmd

$(BIN_DIR)/Bash: tools/Bash/cmd/main.go
	go build -o $(BIN_DIR)/Bash ./tools/Bash/cmd

$(BIN_DIR)/imagen: tools/imagen/cmd/main.go
	go build -o $(BIN_DIR)/imagen ./tools/imagen/cmd

$(BIN_DIR)/imagen_edit: tools/imagen_edit/cmd/main.go
	go build -o $(BIN_DIR)/imagen_edit ./tools/imagen_edit/cmd

$(BIN_DIR)/plantuml_check: tools/plantuml_check/cmd/main.go
	cd tools/plantuml_check && go build -o ../../$(BIN_DIR)/plantuml_check ./cmd

$(BIN_DIR)/sleep: tools/sleep/main.go
	go build -o $(BIN_DIR)/sleep ./tools/sleep

# Build and run a specific tool
# Usage: make run TOOL=Bash
run:
	@if [ -z "$(TOOL)" ]; then \
		echo "Please specify a tool with TOOL=<tool_name>"; \
		echo "Available tools: $(TOOLS)"; \
		exit 1; \
	fi
	@if [ ! -d "tools/$(TOOL)" ]; then \
		echo "Tool $(TOOL) not found"; \
		echo "Available tools: $(TOOLS)"; \
		exit 1; \
	fi
	@echo "Running $(TOOL)..."
	@cd tools/$(TOOL)/cmd && go run main.go

# Server binaries
$(BIN_DIR)/cliGCP: host/cliGCP/cmd/main.go
	go build -o $(BIN_DIR)/cliGCP ./host/cliGCP/cmd

$(BIN_DIR)/openaiserver: host/openaiserver/main.go
	go build -o $(BIN_DIR)/openaiserver ./host/openaiserver

# Clean the bin directory
clean:
	rm -rf $(BIN_DIR)
