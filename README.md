# 💠 go-templio

**go-templio** is a modular Content Management System (CMS) engine built for speed, aesthetics, and developer happiness. It combines the performance of **Go** with the modern UI flexibility of **Tailwind CSS v4** and the type-safety of **Templ**.

---

## ✨ Features

- **🚀 Performance-First**: Built with [Fiber v3](https://gofiber.io/) and [Bun ORM](https://bun.uptrace.dev/).
- **🎨 Premium UI**: Beautiful [TemplUI](https://templui.io/) components with native **Tailwind CSS v4** styling.
- **🏗️ Scaffolder (The Compiler)**: A built-in CLI to generate full CRUD modules (Entity, Service, Handler, Views, & Migrations) in seconds.
- **🌓 Dark Mode**: Built-in Zinc/Indigo theme support with manual toggle.
- **🛠️ Type-Safe Views**: Component-based UI using [Templ](https://templ.guide/).

---

## 🛠️ Tech Stack

- **Backend**: Go 1.25+, Fiber v3
- **ORM**: Bun (Supports PostgreSQL, SQLite)
- **Frontend**: Templ, Tailwind CSS v4, Lucide Icons
- **CLI**: Cobra

---

## 🚦 Getting Started

### 1. Prerequisites
Ensure you have the following installed:
- [Go 1.25+](https://go.dev/dl/)
- [Make](https://www.gnu.org/software/make/)

### 2. Install the CLI (The Compiler)
To install the scaffolding tool globally via GitHub:
```bash
go install github.com/alimuddin7/go-templio/cmd/templio@latest
```
Alternatively, install it locally from the project root:
```bash
go install ./cmd/templio
```

### 3. Setup Environment
Copy the example environment file and update your database credentials:
```bash
cp .env.example .env
```

### 4. Database Migration
Initialize and run the initial schema:
```bash
make migrate
```

### 5. Run Development Server
Start the server with hot-reloading (requires `templ` installed):
```bash
make dev
```
The app will be available at `http://localhost:3000`.

---

## 📂 Usage (Scaffolding a Resource)

**go-templio** is designed to grow with your needs. You can generate a full feature set (CRUD) for any resource instantly.

### Example: Generate a "Post" Module
```bash
# Define your resource name
templio generate-resource --name=Post
```

This will automatically create:
- `internal/domain/post/` (Entity & Ports)
- `internal/repository/post/` (Bun Repository)
- `internal/service/post/` (Business Logic)
- `internal/transport/http/handler/post/` (Fiber Handler)
- `views/post/` (Templ List, Create, & Update pages)
- `internal/database/migrations/` (SQL Up/Down migrations)

Then, simply run:
```bash
make generate build-css
```

---

## 📝 Project Commands

| Command | Description |
| :--- | :--- |
| `make dev` | Start dev server with Templ watch & Go hot-reload |
| `make build` | Build CSS and compile binaries |
| `make resource name=X` | Shortcut to scaffold a new resource named X |
| `make migrate` | Run all pending migrations |
| `make clean` | Remove all generated templ files and binaries |

---

## 🌓 Naming Conventions

The generator intelligently converts your Go struct names to database columns:
- `URL` → `url`
- `HTTPServer` → `http_server`
- `isActive` → `is_active`

Enjoy building your next masterpiece with **go-templio**! 🚀
