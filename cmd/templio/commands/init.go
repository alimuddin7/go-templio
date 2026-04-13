package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// InitCmd returns the init cobra command.
func InitCmd() *cobra.Command {
	var modulePath string

	cmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new go-templio project",
		Long: `Initializes a new project by cloning the go-templio boilerplate,
removing the .git history, and renaming the Go module.

Example:
  templio init my-new-app --module github.com/myuser/my-app`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]
			if modulePath == "" {
				modulePath = projectName
			}

			// 1. Clone the boilerplate
			fmt.Printf("💠 Initializing project %q...\n", projectName)
			fmt.Println("🚀 Cloning boilerplate from GitHub...")
			cloneCmd := exec.Command("git", "clone", "https://github.com/alimuddin7/go-templio.git", projectName)
			cloneCmd.Stdout = os.Stdout
			cloneCmd.Stderr = os.Stderr
			if err := cloneCmd.Run(); err != nil {
				return fmt.Errorf("failed to clone repository: %w", err)
			}

			// 2. Remove .git folder
			fmt.Println("🧹 Cleaning up git history...")
			if err := os.RemoveAll(filepath.Join(projectName, ".git")); err != nil {
				return fmt.Errorf("failed to remove .git folder: %w", err)
			}

			// 3. Rename module
			fmt.Printf("📝 Renaming module to %q...\n", modulePath)
			if err := renameModule(projectName, "github.com/alimuddin7/go-templio", modulePath); err != nil {
				return fmt.Errorf("failed to rename module: %w", err)
			}

			// 4. Go mod tidy
			fmt.Println("📦 Running go mod tidy...")
			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = projectName
			tidyCmd.Stdout = os.Stdout
			tidyCmd.Stderr = os.Stderr
			if err := tidyCmd.Run(); err != nil {
				fmt.Printf("⚠️  Warning: go mod tidy failed: %v\n", err)
			}

			fmt.Printf("\n✅ Project %q initialized successfully!\n", projectName)
			fmt.Printf("👉 Next steps:\n   cd %s\n   cp .env.example .env\n   make migrate\n   make dev\n", projectName)

			return nil
		},
	}

	cmd.Flags().StringVarP(&modulePath, "module", "m", "", "Go module path (defaults to project name)")

	return cmd
}

func renameModule(root, oldModule, newModule string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == "node_modules" || info.Name() == ".git" || info.Name() == "static" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".go" || ext == ".templ" || info.Name() == "go.mod" || info.Name() == "Makefile" || info.Name() == "README.md" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			content := string(data)
			if strings.Contains(content, oldModule) {
				newContent := strings.ReplaceAll(content, oldModule, newModule)
				if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
					return err
				}
				fmt.Printf("   ✓ updated %s\n", path)
			}
		}

		return nil
	})
}
