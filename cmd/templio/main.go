// Command templio is the code-generation CLI for the go-templio CMS engine.
// Usage: templio generate-resource --name=Post
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/alimuddin7/go-templio/cmd/templio/commands"
)

func main() {
	root := &cobra.Command{
		Use:   "templio",
		Short: "go-templio CLI — scaffold CMS modules",
		Long: `templio is the code-generation tool for the go-templio CMS engine.

It reads Go struct definitions and scaffolds full CRUD modules:
entity, repository, service, HTTP handler, and Templ views.`,
		SilenceUsage: true,
	}

	root.AddCommand(commands.GenerateResourceCmd())
	root.AddCommand(commands.InitCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
