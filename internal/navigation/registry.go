// Package navigation provides a centralized, zero-global-state registry
// for sidebar menu items. Modules self-register during wire-up.
package navigation

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Item represents a single navigation entry in the sidebar.
type Item struct {
	Label    string `yaml:"label"`
	Icon     string `yaml:"icon"`     // lucide icon name e.g. "users"
	Href     string `yaml:"href"`
	Children []Item `yaml:"children"` // optional sub-menu
	Order    int    `yaml:"order"`    // lower = higher position
}

// Registry holds all menu items and is passed as a value through the DI chain.
// It is safe for concurrent use (modules register during startup, reads during request).
type Registry struct {
	mu    sync.RWMutex
	items []Item
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry { return &Registry{} }

// Register adds one or more items. Concurrent-safe.
func (r *Registry) Register(items ...Item) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items = append(r.items, items...)
}

// Items returns a sorted copy of all registered items.
func (r *Registry) Items() []Item {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Item, len(r.items))
	copy(out, r.items)
	sortItems(out)
	return out
}

// LoadYAML reads a YAML file and registers all items found there.
// This is called ONCE at startup; module code-registration supplements it.
func (r *Registry) LoadYAML(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var items []Item
	if err := yaml.Unmarshal(data, &items); err != nil {
		return err
	}
	r.Register(items...)
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func sortItems(items []Item) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && items[j].Order < items[j-1].Order; j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}
