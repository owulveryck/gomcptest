.PHONY: all clean tools servers install run

# Build everything
all: tools servers

# Define the bin directory
BIN_DIR := bin

# Tools to build
TOOLS := LS GrepTool Edit GlobTool Replace View duckdbserver dispatch_agent Bash imagen imagen_edit plantuml plantuml_check sleep

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

# Build tools using recursive make
$(BIN_DIR)/LS:
	$(MAKE) -C tools/LS build

$(BIN_DIR)/GrepTool:
	$(MAKE) -C tools/GrepTool build

$(BIN_DIR)/Edit:
	$(MAKE) -C tools/Edit build

$(BIN_DIR)/GlobTool:
	$(MAKE) -C tools/GlobTool build

$(BIN_DIR)/Replace:
	$(MAKE) -C tools/Replace build

$(BIN_DIR)/View:
	$(MAKE) -C tools/View build

$(BIN_DIR)/dispatch_agent:
	$(MAKE) -C tools/dispatch_agent build

$(BIN_DIR)/Bash:
	$(MAKE) -C tools/Bash build

$(BIN_DIR)/imagen:
	$(MAKE) -C tools/imagen build

$(BIN_DIR)/imagen_edit:
	$(MAKE) -C tools/imagen_edit build

$(BIN_DIR)/plantuml:
	$(MAKE) -C tools/plantuml build

$(BIN_DIR)/plantuml_check:
	$(MAKE) -C tools/plantuml_check build

$(BIN_DIR)/duckdbserver:
	$(MAKE) -C tools/duckdbserver build

$(BIN_DIR)/sleep:
	$(MAKE) -C tools/sleep build

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

# Build servers using recursive make
$(BIN_DIR)/cliGCP:
	$(MAKE) -C host/cliGCP build

$(BIN_DIR)/openaiserver:
	$(MAKE) -C host/openaiserver build

# Clean the bin directory
clean:
	rm -rf $(BIN_DIR)
