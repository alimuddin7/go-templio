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

### 1. Initialize a New Project
Create a new project by cloning the boilerplate and renaming the module:
```bash
# Initialize with project name (module name defaults to project name)
templio init my-awesome-project

# Initialize with custom module name
templio init my-app --module github.com/username/my-app
```

### 2. Scaffold a Resource
```bash
# Define your resource name
templio generate-resource --name=Post
```

---

## 🗺️ Navigation & Sub-menus

You can easily organize your sidebar with sub-menus by editing `navigation.yaml`.

### Sub-menu Sample
```yaml
- label: CMS Content
  icon: folder
  href: "#"
  order: 3
  children:
    - label: Posts
      href: /posts
    - label: Categories
      href: /categories
```
*Note: Set `href: "#"` for parent items that only serve as dropdown toggles.*

---

## 🛠️ Modifying Existing Resources

If you need to add or change fields in an existing module (e.g., adding a `content` field to `Post`):

1. **Update Domain Struct**: Open `internal/domain/post/entity.go` and add the field.
2. **Create Migration**: 
   ```bash
   make migrate-create name=add_content_to_posts
   ```
   In the new `.up.sql` file, add: `ALTER TABLE posts ADD COLUMN content TEXT;`.
3. **Update Repository**: Update the SQL queries in `internal/repository/post/repository.go`.
4. **Update Views**: Add the new field's input in `views/post/create.templ` and `update.templ`.
5. **Re-generate**: Run `make generate build-css`.

---

## 🔗 Handling Table Relations

**go-templio** intelligently detects relationship patterns. If your struct contains a field ending in `ID` (e.g., `CategoryID`), it will automatically generate a **SelectBox** component.

### Relation Sample (Category -> Post)

1. **Scaffold Category first**:
   ```bash
   templio generate-resource --name=Category
   ```
2. **Define Post with CategoryID**:
   ```go
   type Post struct {
       ID         int64
       Title      string
       CategoryID int64 `templ:"type:select"` // Generator will pick this up
       CreatedAt  time.Time
   }
   ```
3. **Manual Wire-up in Handler**:
   In `internal/transport/http/handler/post/handler.go`, fetch the categories and pass them to the view:
   ```go
   // Inside createForm handler
   categories, _ := h.categorySvc.List(c.Context(), "", 1, 100)
   return engine.Render(c, postviews.Create(h.nav.Items(), categories))
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
