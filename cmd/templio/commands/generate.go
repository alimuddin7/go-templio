// Package commands provides the Cobra command definitions for the templio CLI.
package commands

import (
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"

	"gopkg.in/yaml.v3"
	"github.com/spf13/cobra"
)

//go:embed templates
var templateFS embed.FS

// FieldDef describes a single struct field parsed from Go source.
type FieldDef struct {
	GoName      string // e.g. "FirstName"
	GoType      string // e.g. "string"
	ColumnName  string // e.g. "first_name"
	Label       string // e.g. "First Name"
	FormName    string // e.g. "first_name"
	Placeholder string // e.g. "Enter first name"
	InputType   string // html input type: "text", "email", "number", etc.
	Component   string // UI component: "Input", "Textarea", "DatePicker", etc.
	Options     []string // For Select, Radio, etc.
}

// ResourceDef is the data fed into all code generation templates.
type ResourceDef struct {
	ModulePath  string      // go.mod module path
	Name        string      // e.g. "Post"
	PackageName string      // e.g. "post"
	PluralName  string      // e.g. "Posts"
	TableName   string      // e.g. "posts"
	URLPrefix   string      // e.g. "posts"
	Force       bool        // Overwrite existing files
	StructFile  string      // Path to original struct file (to prevent overwriting itself)
	Fields      []FieldDef
}

// GenerateResourceCmd returns the generate-resource cobra command.
func GenerateResourceCmd() *cobra.Command {
	var (
		name       string
		structFile string
		modulePath string
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "generate-resource",
		Short: "Scaffold a full CRUD module from a Go struct",
		Long: `Reads a Go struct definition (or uses --name for a blank scaffold)
and generates: entity, repository, service, HTTP handler, and Templ views.

Examples:
  templio generate-resource --name=Post
  templio generate-resource --name=Post --struct=./internal/domain/post/entity.go`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			// Auto-detect module path if not provided.
			if modulePath == "" {
				var err error
				modulePath, err = readModulePath("go.mod")
				if err != nil {
					return fmt.Errorf("could not read go.mod: %w\n(use --module to specify manually)", err)
				}
			}

			def := ResourceDef{
				ModulePath:  modulePath,
				Name:        pascal(name),
				PackageName: strings.ToLower(name),
				PluralName:  plural(pascal(name)),
				TableName:   toSnake(plural(name)),
				URLPrefix:   toSnake(plural(name)),
				Force:       force,
				StructFile:  structFile,
			}

			// Parse struct fields if a file is provided.
			if structFile != "" {
				fields, err := parseStructFields(structFile, def.Name)
				if err != nil {
					return fmt.Errorf("parse struct: %w", err)
				}
				def.Fields = fields
			} else {
				// Default scaffold with a title field.
				def.Fields = []FieldDef{
					{
						GoName: "Title", GoType: "string",
						ColumnName: "title", Label: "Title",
						FormName: "title", Placeholder: "Enter title",
						InputType: "text",
						Component: "Input",
					},
				}
			}

			return runGenerate(def)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Resource name in PascalCase (e.g. Post)")
	cmd.Flags().StringVarP(&structFile, "struct", "s", "", "Path to Go file containing the struct definition")
	cmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (auto-detected from go.mod if omitted)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files (except entity.go if it is being parsed)")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

// runGenerate executes all template rendering steps for a ResourceDef.
func runGenerate(def ResourceDef) error {
	steps := []struct {
		tmplPath string
		outPath  string
	}{
		{
			"templates/entity.tmpl",
			fmt.Sprintf("internal/domain/%s/entity.go", def.PackageName),
		},
		{
			"templates/repository.tmpl",
			fmt.Sprintf("internal/repository/%s/repository.go", def.PackageName),
		},
		{
			"templates/service.tmpl",
			fmt.Sprintf("internal/service/%s/service.go", def.PackageName),
		},
		{
			"templates/ports.tmpl",
			fmt.Sprintf("internal/domain/%s/ports.go", def.PackageName),
		},
		{
			"templates/handler.tmpl",
			fmt.Sprintf("internal/transport/http/handler/%s/handler.go", def.PackageName),
		},
		{
			"templates/views/list.tmpl",
			fmt.Sprintf("views/%s/list.templ", def.PackageName),
		},
		{
			"templates/views/create.tmpl",
			fmt.Sprintf("views/%s/create.templ", def.PackageName),
		},
		{
			"templates/views/update.tmpl",
			fmt.Sprintf("views/%s/update.templ", def.PackageName),
		},
		{
			"templates/migration_up.tmpl",
			fmt.Sprintf("internal/database/migrations/%s_create_%s.up.sql", time.Now().Format("20060102150405"), def.TableName),
		},
		{
			"templates/migration_down.tmpl",
			fmt.Sprintf("internal/database/migrations/%s_create_%s.down.sql", time.Now().Format("20060102150405"), def.TableName),
		},
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"len": func(s []FieldDef) int { return len(s) },
		"lower": strings.ToLower,
	}

	for _, step := range steps {
		// Read template from embedded FS.
		tmplBytes, err := templateFS.ReadFile(step.tmplPath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", step.tmplPath, err)
		}

		tmpl, err := template.New(filepath.Base(step.tmplPath)).Funcs(funcMap).Parse(string(tmplBytes))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", step.tmplPath, err)
		}

		// Create output directory.
		if err := os.MkdirAll(filepath.Dir(step.outPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(step.outPath), err)
		}

		// Guard: skip if file already exists, unless force is true.
		// Never overwrite the struct file if it's the one we are parsing from.
		if def.Force && !(step.tmplPath == "templates/entity.tmpl" && def.StructFile != "") {
			// allow overwrite
		} else if _, err := os.Stat(step.outPath); err == nil {
			fmt.Printf("  ⚠ skip  %s (already exists)\n", step.outPath)
			continue
		}

		f, err := os.Create(step.outPath)
		if err != nil {
			return fmt.Errorf("create file %s: %w", step.outPath, err)
		}

		if err := tmpl.Execute(f, def); err != nil {
			f.Close()
			return fmt.Errorf("execute template %s: %w", step.tmplPath, err)
		}
		f.Close()
		fmt.Printf("  ✓ wrote %s\n", step.outPath)
	}

	// Automate main.go registration
	if err := updateMainGo(def); err != nil {
		fmt.Printf("  ⚠ could not update main.go: %v\n", err)
	}

	// Automate navigation.yaml update
	if err := updateNavigationYAML(def); err != nil {
		fmt.Printf("  ⚠ could not update navigation.yaml: %v\n", err)
	}

	fmt.Printf("\n✅ Resource %q scaffolded successfully!\n", def.Name)
	return nil
}

func updateNavigationYAML(def ResourceDef) error {
	navPath := "navigation.yaml"
	data, err := os.ReadFile(navPath)
	if err != nil {
		return err
	}

	var navItems []any
	if err := yaml.Unmarshal(data, &navItems); err != nil {
		return err
	}

	// Check if already exists
	resourceHref := "/" + def.URLPrefix
	for _, item := range navItems {
		if m, ok := item.(map[string]any); ok {
			if m["href"] == resourceHref {
				return nil // Already exists
			}
		}
	}

	// Append new item
	newItem := map[string]any{
		"label": def.PluralName,
		"icon":  "file-text", // Default icon
		"href":  resourceHref,
		"order": 10,
	}
	navItems = append(navItems, newItem)

	// Marshal back
	out, err := yaml.Marshal(navItems)
	if err != nil {
		return err
	}

	return os.WriteFile(navPath, out, 0644)
}

// updateMainGo injects the new resource into cmd/server/main.go using anchor comments.
func updateMainGo(def ResourceDef) error {
	mainPath := "cmd/server/main.go"
	data, err := os.ReadFile(mainPath)
	if err != nil {
		return err
	}
	content := string(data)

	// Build snippets
	repoAlias := def.PackageName + "repo"
	svcAlias := def.PackageName + "svc"
	handlerAlias := def.PackageName + "handler"

	importSnippet := fmt.Sprintf("\t%s \"%s/internal/repository/%s\"\n\t%s \"%s/internal/service/%s\"\n\t%s \"%s/internal/transport/http/handler/%s\"\n\t// [GEN-IMPORT]",
		repoAlias, def.ModulePath, def.PackageName,
		svcAlias, def.ModulePath, def.PackageName,
		handlerAlias, def.ModulePath, def.PackageName,
	)

	repoSnippet := fmt.Sprintf("\t%sRepo := %s.New(db.DB)\n\t// [GEN-REPO]", def.PackageName, repoAlias)

	migrateSnippet := fmt.Sprintf("\tif err := %sRepo.Migrate(ctx); err != nil { log.Fatalf(\"migrate %s: %%v\", err) }\n\t// [GEN-MIGRATE]",
		def.PackageName, def.PluralName)

	svcSnippet := fmt.Sprintf("\t%sService := %s.New(%sRepo)\n\t// [GEN-SERVICE]", def.PackageName, svcAlias, def.PackageName)

	registerSnippet := fmt.Sprintf("\tapp.Register(%s.New(%sService, nav).AsModule())\n\t// [GEN-REGISTER]",
		handlerAlias, def.PackageName)

	// Replace anchors
	content = strings.Replace(content, "// [GEN-IMPORT]", importSnippet, 1)
	content = strings.Replace(content, "// [GEN-REPO]", repoSnippet, 1)
	content = strings.Replace(content, "// [GEN-MIGRATE]", migrateSnippet, 1)
	content = strings.Replace(content, "// [GEN-SERVICE]", svcSnippet, 1)
	content = strings.Replace(content, "// [GEN-REGISTER]", registerSnippet, 1)

	return os.WriteFile(mainPath, []byte(content), 0644)
}

// ── AST struct parser ─────────────────────────────────────────────────────────

// parseStructFields opens a Go source file and extracts fields from the named struct.
func parseStructFields(filePath, structName string) ([]FieldDef, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, err
	}

	var fields []FieldDef
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != structName {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue // embedded
				}
				goName := field.Names[0].Name
				// Skip framework-internal fields.
				if goName == "ID" || goName == "CreatedAt" || goName == "UpdatedAt" {
					continue
				}
				goType := exprToString(field.Type)
				col := toSnake(goName)
				inputType, component, options := goTypeToInputAndComponent(goType, field.Tag)
				fields = append(fields, FieldDef{
					GoName:      goName,
					GoType:      goType,
					ColumnName:  col,
					Label:       toLabel(goName),
					FormName:    col,
					Placeholder: "Enter " + strings.ToLower(toLabel(goName)),
					InputType:   inputType,
					Component:   component,
					Options:     options,
				})
			}
		}
	}
	return fields, nil
}

func exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + exprToString(e.X)
	case *ast.SelectorExpr:
		return exprToString(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(e.Elt)
	default:
		return "interface{}"
	}
}

func parseTemplTag(tagVal string) (string, []string) {
	var component string
	var options []string

	parts := strings.Split(tagVal, ",")
	for _, p := range parts {
		kv := strings.SplitN(p, ":", 2)
		switch len(kv) {
		case 2:
			switch kv[0] {
			case "type":
				switch kv[1] {
				case "select":
					component = "SelectBox"
				case "radio":
					component = "Radio"
				case "switch":
					component = "Switch"
				case "rating":
					component = "Rating"
				case "timepicker":
					component = "TimePicker"
				case "datepicker":
					component = "DatePicker"
				case "checkbox":
					component = "Checkbox"
				case "textarea":
					component = "Textarea"
				}
			case "options":
				options = strings.Split(kv[1], "|")
			}
		case 1:
			// Fallback for simple tags
			switch kv[0] {
			case "textarea":
				component = "Textarea"
			case "datepicker":
				component = "DatePicker"
			}
		}
	}
	return component, options
}

func goTypeToInputAndComponent(goType string, tag *ast.BasicLit) (string, string, []string) {
	var component string
	var options []string

	// 1. Check for custom templ tag
	if tag != nil {
		tagStr := tag.Value
		if idx := strings.Index(tagStr, `templ:"`); idx != -1 {
			endIdx := strings.Index(tagStr[idx+7:], `"`)
			if endIdx != -1 {
				templVal := tagStr[idx+7 : idx+7+endIdx]
				component, options = parseTemplTag(templVal)
			}
		}
	}

	if component != "" {
		// Provide reasonable input types for text-like inputs
		inputType := "text"
		if component == "DatePicker" || component == "TimePicker" {
			inputType = "text" // typically custom components use hidden or custom
		}
		return inputType, component, options
	}

	// 2. Default inference
	switch goType {
	case "int", "int64", "int32", "float64":
		return "number", "Input", nil
	case "bool":
		return "checkbox", "Checkbox", nil
	case "time.Time":
		return "date", "DatePicker", nil
	default:
		if strings.Contains(strings.ToLower(goType), "email") {
			return "email", "Input", nil
		}
		return "text", "Input", nil
	}
}

// ── string helpers ─────────────────────────────────────────────────────────────

func pascal(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func plural(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") {
		return s + "es"
	}
	return s + "s"
}

func toSnake(s string) string {
	commonInitialisms := map[string]bool{
		"ACL": true, "API": true, "ASCII": true, "CPU": true, "CSS": true, "DNS": true,
		"EOF": true, "GUID": true, "HTML": true, "HTTP": true, "HTTPS": true, "ID": true,
		"IP": true, "JSON": true, "LHS": true, "QPS": true, "RAM": true, "RHS": true,
		"RPC": true, "SLA": true, "SMTP": true, "SQL": true, "SSH": true, "TCP": true,
		"TLS": true, "TTL": true, "UDP": true, "UI": true, "UID": true, "UUID": true,
		"URI": true, "URL": true, "UTF8": true, "VM": true, "XML": true, "XMPP": true,
		"XSRF": true, "XSS": true,
	}

	var res []string
	runes := []rune(s)
	length := len(runes)

	for i := 0; i < length; {
		// Identify the next word
		start := i
		// If upper, could be an initialism or StartOfWord
		if unicode.IsUpper(runes[i]) {
			// Find how many consecutive uppers
			j := i + 1
			for j < length && unicode.IsUpper(runes[j]) {
				j++
			}

			// If we have multiple uppers, check for initialism
			if j-i > 1 {
				// Potential initialism
				// If followed by lower, the last upper belongs to the next word (e.g., JSONData -> JSON + Data)
				if j < length && unicode.IsLower(runes[j]) {
					word := string(runes[i : j-1])
					if commonInitialisms[word] {
						res = append(res, strings.ToLower(word))
						i = j - 1
						continue
					}
				} else {
					word := string(runes[i:j])
					if commonInitialisms[word] {
						res = append(res, strings.ToLower(word))
						i = j
						continue
					}
				}
			}
			// Just a single upper, or not a recognized initialism
			i++
			for i < length && !unicode.IsUpper(runes[i]) {
				i++
			}
			res = append(res, strings.ToLower(string(runes[start:i])))
		} else {
			// Lower case start
			for i < length && !unicode.IsUpper(runes[i]) {
				i++
			}
			res = append(res, strings.ToLower(string(runes[start:i])))
		}
	}
	return strings.Join(res, "_")
}

func toLabel(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func readModulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}
	return "", fmt.Errorf("module declaration not found in %s", goModPath)
}


