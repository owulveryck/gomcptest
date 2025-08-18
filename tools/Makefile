.PHONY: all build clean

DIRS := Bash Edit GlobTool GrepTool LS Replace View dispatch_agent imagen
BINDIR := bin

all: build

build:
	@mkdir -p $(BINDIR)
	@for dir in $(DIRS); do \
		echo "Building $$dir..."; \
		(cd $$dir/cmd && go build -o ../../$(BINDIR)/$$dir); \
	done
	@echo "All binaries built in $(BINDIR)/"

clean:
	@echo "Cleaning..."
	@rm -rf $(BINDIR)
	@echo "Done"

# Build and run a specific function server
# Usage: make run FUNC=Bash
run:
	@if [ -z "$(FUNC)" ]; then \
		echo "Please specify a function with FUNC=<function_name>"; \
		echo "Available functions: $(DIRS)"; \
		exit 1; \
	fi
	@if [ ! -d "$(FUNC)" ]; then \
		echo "Function $(FUNC) not found"; \
		echo "Available functions: $(DIRS)"; \
		exit 1; \
	fi
	@echo "Running $(FUNC)..."
	@cd $(FUNC)/cmd && go run main.go
