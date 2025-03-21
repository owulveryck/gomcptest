.PHONY: all clean tools servers 

# Build everything
all: tools servers

# Define the bin directory
BIN_DIR := bin

# Tools to build
TOOLS := LS GrepTool Edit GlobTool Replace View duckdbserver dispatch_agent Bash

# Servers to build
SERVERS := cliGCP openaiserver

# Ensure the bin directory exists
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# Build all tools
tools: $(BIN_DIR) $(addprefix $(BIN_DIR)/, $(TOOLS))

# Build all servers
servers: $(BIN_DIR) $(addprefix $(BIN_DIR)/, $(SERVERS)) 

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

# Server binaries
$(BIN_DIR)/cliGCP: host/cliGCP/cmd/main.go
	go build -o $(BIN_DIR)/cliGCP ./host/cliGCP/cmd

$(BIN_DIR)/openaiserver: host/openaiserver/main.go
	go build -o $(BIN_DIR)/openaiserver ./host/openaiserver

# Clean the bin directory
clean:
	rm -rf $(BIN_DIR)
