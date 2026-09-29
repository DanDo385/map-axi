package extract

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/DanDo385/map-axi/internal/ir"
)

// outlineItem is a single symbol or import entry from `ast-grep outline --json`.
type outlineItem struct {
	Role       string `json:"role"`
	SymbolType string `json:"symbolType"`
	Name       string `json:"name"`
	Signature  string `json:"signature"`
	IsImport   bool   `json:"isImport"`
	IsExported bool   `json:"isExported"`
}

// outlineFile contains outline entries for one source file.
type outlineFile struct {
	Path     string         `json:"path"`
	Language string         `json:"language"`
	Items    []outlineItem  `json:"items"`
}

// ExtractSymbols runs ast-grep outline to discover symbols in the codebase.
// Uses `ast-grep outline --items structure --json` for structure items.
func ExtractSymbols(root string) ([]ir.Node, error) {
	binary, err := findBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(binary, "outline", root,
		"--items", "structure",
		"--json",
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ast-grep outline: %w", err)
	}

	var files []outlineFile
	if err := json.Unmarshal(out, &files); err != nil {
		return nil, fmt.Errorf("parse outline: %w", err)
	}

	var nodes []ir.Node
	seen := make(map[string]bool)
	for _, f := range files {
		for _, item := range f.Items {
			if item.Name == "" {
				continue
			}
			id := nodeID(f.Path, item.Name)
			if seen[id] {
				continue
			}
			seen[id] = true

			kind := mapSymbolKind(item.SymbolType)
			label := fmt.Sprintf("%s %s", kindEmoji(kind), item.Name)
			if item.Signature != "" {
				label = fmt.Sprintf("%s %s%s", kindEmoji(kind), item.Name, item.Signature)
			}

			nodes = append(nodes, ir.Node{
				ID:     id,
				Label:  label,
				Type:   kind,
				Parent: parentFromFile(f.Path),
			})
		}
	}
	return nodes, nil
}

// mapSymbolKind maps ast-grep's symbolType to our node type.
func mapSymbolKind(st string) string {
	switch st {
	case "function", "method":
		return "function"
	case "class", "struct", "interface", "enum", "typeParameter":
		return "class"
	case "module":
		return "module"
	case "variable", "constant":
		return "module" // treat as module-level
	default:
		return "file"
	}
}

// findBinary looks for ast-grep or sg in PATH.
func findBinary() (string, error) {
	for _, name := range []string{"ast-grep", "sg"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("ast-grep not found in PATH; install it: mise use ast-grep")
}

func nodeID(file, name string) string {
	file = strings.TrimPrefix(file, "./")
	file = strings.TrimSuffix(file, "/")
	return fmt.Sprintf("%s:%s", file, name)
}

func parentFromFile(file string) string {
	file = strings.TrimPrefix(file, "./")
	parts := strings.Split(file, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

func kindEmoji(kind string) string {
	switch kind {
	case "function":
		return "\U0001F527" // 🔧
	case "class":
		return "\u2699\uFE0F" // ⚙️
	case "module":
		return "\U0001F537" // 🔷
	case "package":
		return "\U0001F4E6" // 📦
	default:
		return "\U0001F4C4" // 📄
	}
}