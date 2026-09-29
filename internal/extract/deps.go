package extract

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/DanDo385/map-axi/internal/ir"
)

// ExtractDeps discovers dependency edges by extracting imports from ast-grep outline.
// Uses `ast-grep outline --items imports --json` for import declarations.
func ExtractDeps(root string) ([]ir.Edge, error) {
	binary, err := findBinary()
	if err != nil {
		return nil, err
	}

	// Get imports from outline
	cmd := exec.Command(binary, "outline", root,
		"--items", "imports",
		"--json",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ast-grep outline imports: %w", err)
	}

	var files []outlineFile
	if err := json.Unmarshal(out, &files); err != nil {
		return nil, fmt.Errorf("parse imports: %w", err)
	}

	var edges []ir.Edge
	seen := make(map[string]bool)
	for _, f := range files {
		source := parentFromFile(f.Path)
		for _, item := range f.Items {
			if item.Name == "" {
				continue
			}
			module := strings.Trim(item.Name, `"`)
			if module == "" || strings.HasPrefix(module, ".") {
				continue // skip local/relative imports
			}

			key := fmt.Sprintf("%s->%s", source, module)
			if seen[key] {
				// Increment weight for duplicate imports
				for i := range edges {
					if edges[i].Source == source && edges[i].Target == module {
						edges[i].Weight++
						break
					}
				}
				continue
			}
			seen[key] = true

			edges = append(edges, ir.Edge{
				Source: source,
				Target: module,
				Type:   "import",
				Weight: 1,
			})
		}
	}

	return edges, nil
}