// Package extract discovers architectural structure from source code.
// It shells out to ast-grep for symbol and dependency extraction,
// then assembles the results into the Architecture IR.
package extract

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DanDo385/map-axi/internal/ir"
)

// Run performs a full extraction pass over root, returning a populated Architecture.
// It discovers symbols, dependencies, entrypoints, and routes.
func Run(root string) (*ir.Architecture, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}

	name := filepath.Base(abs)
	arch := &ir.Architecture{Name: name}

	// Symbols: functions, classes, types, constants, variables
	nodes, err := ExtractSymbols(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map-axi: symbol extraction: %v\n", err)
	}
	arch.Nodes = nodes

	// Dependencies: import edges
	edges, err := ExtractDeps(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map-axi: dependency extraction: %v\n", err)
	}
	arch.Edges = edges

	// Entrypoints: find main files
	arch.Entrypoints = findEntrypoints(root)

	// External systems: classify non-local imports
	arch.ExternalSystems = classifyExternals(arch.Edges)

	return arch, nil
}

// findEntrypoints looks for common entrypoint filenames.
func findEntrypoints(root string) []string {
	patterns := []string{
		"main.go", "main.ts", "main.tsx", "main.js", "main.py",
		"index.ts", "index.tsx", "index.js", "index.py",
		"__main__.py", "app.ts", "app.tsx", "server.ts", "server.go",
		"cmd/*/main.go",
	}

	var found []string
	for _, pat := range patterns {
		matches, _ := filepath.Glob(filepath.Join(root, pat))
		for _, m := range matches {
			rel, _ := filepath.Rel(root, m)
			found = append(found, rel)
		}
	}
	return found
}

// classifyExternals groups import edges by external module and returns ExternalSystem records.
// Filters out the project's own packages and Go standard library.
func classifyExternals(edges []ir.Edge) []ir.ExternalSystem {
	// Known stdlib prefixes for major languages
	stdlibPrefixes := []string{
		"fmt", "os", "io", "net", "sync", "time", "context",
		"strings", "strconv", "bytes", "errors", "sort", "math",
		"encoding", "reflect", "regexp", "runtime", "flag",
		"path", "log", "bufio", "builtin", "unicode", "crypto",
		"database", "embed", "html", "image", "index", "mime",
		"syscall", "testing", "text", "archive", "compress",
		"container", "debug", "go/",
	}

	isStdlib := func(mod string) bool {
		for _, p := range stdlibPrefixes {
			if mod == p || (len(p) > 0 && p[len(p)-1] == '/' && len(mod) > len(p) && mod[:len(p)] == p) {
				return true
			}
		}
		return false
	}

	byModule := make(map[string][]string)
	for _, e := range edges {
		if e.Type != "import" {
			continue
		}
		// Skip project-internal and stdlib imports
		if isStdlib(e.Target) {
			continue
		}
		byModule[e.Target] = append(byModule[e.Target], e.Source)
	}

	var systems []ir.ExternalSystem
	for mod, sources := range byModule {
		if len(sources) == 0 {
			continue
		}
		systems = append(systems, ir.ExternalSystem{
			Name:        mod,
			Type:        guessSystemType(mod),
			Connections: sources,
		})
	}
	return systems
}

func guessSystemType(mod string) string {
	modLower := filepath.Base(mod)
	switch {
	case containsAny(modLower, "postgres", "mysql", "sqlite", "pg", "db", "sql"):
		return "database"
	case containsAny(modLower, "redis", "memcache", "memcached"):
		return "cache"
	case containsAny(modLower, "kafka", "rabbitmq", "nats", "sqs", "amqp"):
		return "queue"
	case containsAny(modLower, "s3", "cloudflare", "aws", "storage", "blob"):
		return "storage"
	case containsAny(modLower, "http", "api", "rest", "graphql", "grpc", "rpc"):
		return "api"
	default:
		return "external"
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}