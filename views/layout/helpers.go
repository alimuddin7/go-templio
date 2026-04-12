package layout

import (
	"strings"
	"unicode"
)

// Breadcrumb represents a single segment in the navigation breadcrumb trail.
type Breadcrumb struct {
	Label  string
	Href   string
	IsLast bool
}

// UserInitials extracts the first letter of the first and last name (if available).
func UserInitials(name string) string {
	if name == "" {
		return "U"
	}
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "U"
	}
	if len(parts) == 1 {
		return strings.ToUpper(parts[0][:1])
	}
	return strings.ToUpper(parts[0][:1] + parts[len(parts)-1][:1])
}

// IsActive checks if the current URL path matches the navigation item's path.
func IsActive(currentPath, itemPath string) bool {
	if itemPath == "/" {
		return currentPath == "/"
	}
	return strings.HasPrefix(currentPath, itemPath)
}

// BuildBreadcrumbs decomposes a URL path into a slice of breadcrumb segments.
func BuildBreadcrumbs(path string) []Breadcrumb {
	if path == "/" || path == "" {
		return nil
	}

	segments := strings.Split(strings.Trim(path, "/"), "/")
	breadcrumbs := make([]Breadcrumb, len(segments))
	
	currentHref := ""
	for i, seg := range segments {
		currentHref += "/" + seg
		
		// titleize segment
		label := ""
		if len(seg) > 0 {
			runes := []rune(seg)
			runes[0] = unicode.ToUpper(runes[0])
			label = string(runes)
		}

		breadcrumbs[i] = Breadcrumb{
			Label:  label,
			Href:   currentHref,
			IsLast: i == len(segments)-1,
		}
	}

	return breadcrumbs
}
