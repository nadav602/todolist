.PHONY: build build-push help

# Default target
help:
	@echo "Todolist API service Targets"
	@echo "=========================================="
	@echo "Build target (requires VERSION, PREFIX optional):"
	@echo "  make build       - Build API service container image"
	@echo ""
	@echo "Build and Push target (requires VERSION, PREFIX optional):"
	@echo "  make build-push - Build and push API service image"
	@echo ""
	@echo "Environment Variables:"
	@echo "  CONTAINER_REGISTRY   - Default registry prefix (used if PREFIX not provided)"
	@echo "                        If set, will be offered as default in prompt"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=1.0"
	@echo "  make build-push VERSION=1.0 PREFIX=ghcr.io/bennyro-mta"
	@echo "  CONTAINER_REGISTRY=ghcr.io/bennyro-mta make build-push VERSION=1.0"
	@echo "  export CONTAINER_REGISTRY=ghcr.io/bennyro-mta"
	@echo "  make build-push VERSION=1.0  # Will use CONTAINER_REGISTRY as default"

# Build API
build:
	@VERSION=$(VERSION); \
	PREFIX=$(PREFIX); \
	if [ -z "$$VERSION" ]; then \
		read -p "Enter version for app image: " VERSION; \
	else \
		VERSION=$$VERSION; \
	fi; \
	if [ -z "$$PREFIX" ]; then \
		PREFIX="$$CONTAINER_REGISTRY"; \
		if [ -z "$$PREFIX" ]; then \
			read -p "Enter registry prefix (e.g., ghcr.io/bennyro-mta): " PREFIX; \
		else \
			read -p "Enter registry prefix [$$PREFIX]: " PREFIX_INPUT; \
			PREFIX=$${PREFIX_INPUT:-$$PREFIX}; \
		fi; \
	else \
		PREFIX=$$PREFIX; \
	fi; \
	echo "PREFIX=$$PREFIX" > .build.env; \
	echo "VERSION=$$VERSION" >> .build.env; \
	docker build -t $$PREFIX/todos-api:$$VERSION -f Dockerfile . && \
	echo "✓ Built $$PREFIX/todos-api:$$VERSION"

# Build and push API
build-push: build
	@. ./.build.env; \
	docker push $$PREFIX/todos-api:$$VERSION && \
	echo "✓ Built and pushed $$PREFIX/todos-api:$$VERSION"
