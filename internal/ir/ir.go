// Package ir defines the Architecture IR -- the canonical JSON schema
// for map-axi's codebase model: nodes, edges, entrypoints, routes,
// external systems, and per-node metrics.
package ir

import (
	"encoding/json"
	"fmt"
	"os"
)

// Node represents a structural element: package, module, class, function, or file.
type Node struct {
	ID      string  `json:"id"`
	Label   string  `json:"label"`
	Type    string  `json:"type"` // package, module, class, function, file
	Parent  string  `json:"parent,omitempty"`
	Metrics Metrics `json:"metrics"`
}

// Edge represents a directed relationship between two nodes.
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // import, calls, implements, extends
	Weight int    `json:"weight,omitempty"`
}

// Metrics holds quantitative measures for a single node.
type Metrics struct {
	LOC        int     `json:"loc"`
	Complexity int     `json:"complexity"`
	Churn      int     `json:"churn"`
	Hotspot    float64 `json:"hotspot"`
	Coverage   float64 `json:"coverage"`
}

// Route represents an HTTP or RPC route discovered in the codebase.
type Route struct {
	Path    string `json:"path"`
	Handler string `json:"handler"`
}

// ExternalSystem represents an external dependency (database, queue, API).
type ExternalSystem struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // database, queue, cache, api, storage
	Connections []string `json:"connections"` // node IDs
}

// Architecture is the top-level IR document.
type Architecture struct {
	Name            string           `json:"name"`
	Nodes           []Node           `json:"nodes"`
	Edges           []Edge           `json:"edges"`
	Entrypoints     []string         `json:"entrypoints"`
	Routes          []Route          `json:"routes"`
	ExternalSystems []ExternalSystem `json:"externalSystems"`
}

// Load reads an architecture.json file from disk.
func Load(path string) (*Architecture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var a Architecture
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &a, nil
}

// Save writes the architecture to disk as JSON.
func (a *Architecture) Save(path string) error {
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}