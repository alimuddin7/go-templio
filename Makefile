# ── Project Configuration ─────────────────────────────────────────────────────
BINARY_NAME=templio-server
CLI_NAME=templio
TEMPL_VER=latest

# ── Build Targets ─────────────────────────────────────────────────────────────

.PHONY: all
all: generate build

.PHONY: generate
generate:
	@echo "🎨 Generating Templ views..."
	go run github.com/a-h/templ/cmd/templ@v0.3.1001 generate views

.PHONY: build-css
build-css:
	@echo "💅 Building Tailwind CSS..."
	./tailwindcss-linux-x64 -i ./static/css/tailwind-input.css -o ./static/css/tailwind.css

.PHONY: build
build: build-css generate
	@echo "🏗️ Building binaries..."
	go build -o $(BINARY_NAME) ./cmd/server/main.go
	go build -o $(CLI_NAME) ./cmd/templio/main.go

.PHONY: dev
dev:
	@echo "🚀 Starting development server..."
	go run github.com/a-h/templ/cmd/templ@v0.3.1001 generate --watch --proxy="http://localhost:3001" --cmd="go run ./cmd/server/main.go" views

.PHONY: migrate-init
migrate-init:
	@echo "🗄️ Initializing migrations..."
	go run ./cmd/migrate/main.go init

.PHONY: migrate
migrate:
	@echo "🗄️ Running migrations up..."
	@go run ./cmd/migrate/main.go init 2>/dev/null || true
	go run ./cmd/migrate/main.go up

.PHONY: migrate-down
migrate-down:
	@echo "🗄️ Running migrations down..."
	go run ./cmd/migrate/main.go down

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(name)" ]; then echo "Error: name is required. Use: make migrate-create name=my_migration"; exit 1; fi
	@echo "🗄️ Creating empty migration $(name)..."
	go run ./cmd/migrate/main.go create $(name)

.PHONY: resource
resource:
	@if [ -z "$(name)" ]; then echo "Error: name is required. Use: make resource name=EntityName"; exit 1; fi
	@echo "🚀 Scaffolding resource: $(name)..."
	go run ./cmd/templio/main.go generate-resource --name=$(name)
	@make generate build-css
	@echo "\n✅ Resource $(name) created. Restart 'make dev' to see changes."

.PHONY: page
page: resource

.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	rm -f $(BINARY_NAME) $(CLI_NAME)
	rm -f static/css/tailwind.css
	rm -f static/css/tailwind.css
	find . -name "*_templ.go" -delete
	find . -name "*_templ.txt" -delete

# ── Help ──────────────────────────────────────────────────────────────────────

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make generate  - Generate Go code from Templ templates"
	@echo "  make build     - Build CSS and project binaries"
	@echo "  make dev       - Start dev server with hot-reload (Templ + Go)"
	@echo "  make migrate   - Run database migrations"
	@echo "  make resource  - Scaffold a new CRUD module (usage: make resource name=Post)"
	@echo "  make page      - Alias for make resource"
	@echo "  make clean     - Remove generated files and binaries"
