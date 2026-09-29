package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DanDo385/map-axi/internal/ir"
)

// RenderASCII produces a TOON (text+icon) terminal view of the architecture.
// The top-level is collapsed; drill-down via node IDs.
func RenderASCII(arch *ir.Architecture) string {
	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("\n  \033[1;36m%s\033[0m\n", arch.Name))
	b.WriteString(fmt.Sprintf("  %s\n\n", strings.Repeat("\u2500", 40)))

	// Group nodes by parent
	tree := buildTree(arch.Nodes)

	// Render top-level (no parent nodes)
	renderTree(&b, tree, "", true)

	// Metrics summary
	b.WriteString(fmt.Sprintf("\n  \033[1;33m%s\033[0m\n", "Metrics"))
	b.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("\u2500", 40)))
	totalLOC := 0
	totalComplexity := 0
	for _, n := range arch.Nodes {
		totalLOC += n.Metrics.LOC
		totalComplexity += n.Metrics.Complexity
	}
	b.WriteString(fmt.Sprintf("  %s nodes   %s edges   %s LOC   %s complexity\n",
		metricVal(len(arch.Nodes)),
		metricVal(len(arch.Edges)),
		metricVal(totalLOC),
		metricVal(totalComplexity),
	))

	// Entrypoints
	if len(arch.Entrypoints) > 0 {
		b.WriteString(fmt.Sprintf("\n  \033[1;33m%s\033[0m\n", "Entrypoints"))
		b.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("\u2500", 40)))
		for _, ep := range arch.Entrypoints {
			b.WriteString(fmt.Sprintf("  \U0001F3AF %s\n", ep))
		}
	}

	// External systems
	if len(arch.ExternalSystems) > 0 {
		b.WriteString(fmt.Sprintf("\n  \033[1;33m%s\033[0m\n", "External Systems"))
		b.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("\u2500", 40)))
		for _, es := range arch.ExternalSystems {
			icon := sysIcon(es.Type)
			b.WriteString(fmt.Sprintf("  %s %s (%s) \u2192 %d connections\n",
				icon, es.Name, es.Type, len(es.Connections)))
		}
	}

	// Edge summary (collapsed)
	if len(arch.Edges) > 0 {
		b.WriteString(fmt.Sprintf("\n  \033[1;33m%s\033[0m\n", "Top Dependencies"))
		b.WriteString(fmt.Sprintf("  %s\n", strings.Repeat("\u2500", 40)))
		topEdges := topEdges(arch.Edges, 10)
		for _, e := range topEdges {
			b.WriteString(fmt.Sprintf("  %s \u2192 %s  (%s, x%d)\n",
				truncate(e.Source, 25),
				truncate(e.Target, 25),
				e.Type,
				e.Weight,
			))
		}
	}

	return b.String()
}

type treeNode struct {
	node     ir.Node
	children []*treeNode
}

func buildTree(nodes []ir.Node) map[string]*treeNode {
	index := make(map[string]*treeNode)
	for i := range nodes {
		index[nodes[i].ID] = &treeNode{node: nodes[i]}
	}
	for _, n := range nodes {
		if n.Parent != "" {
			if parent, ok := index[n.Parent]; ok {
				parent.children = append(parent.children, index[n.ID])
			}
		}
	}
	return index
}

func renderTree(b *strings.Builder, index map[string]*treeNode, prefix string, isRoot bool) {
	// Collect top-level nodes (no parent)
	var roots []*treeNode
	for _, tn := range index {
		if tn.node.Parent == "" {
			roots = append(roots, tn)
		}
	}

	// Also include nodes whose parent is not in the index (treated as top-level)
	for _, tn := range index {
		if tn.node.Parent != "" {
			if _, ok := index[tn.node.Parent]; !ok {
				roots = append(roots, tn)
			}
		}
	}

	sort.Slice(roots, func(i, j int) bool {
		if roots[i].node.Type != roots[j].node.Type {
			return roots[i].node.Type < roots[j].node.Type
		}
		return roots[i].node.Label < roots[j].node.Label
	})

	for i, tn := range roots {
		isLast := i == len(roots)-1
		connector := "\u251C\u2500\u2500 " // ├──
		if isLast {
			connector = "\u2514\u2500\u2500 " // └──
		}

		b.WriteString(fmt.Sprintf("  %s%s\n", connector, tn.node.Label))
		renderChildren(b, tn.children, isLast)
	}
}

func renderChildren(b *strings.Builder, children []*treeNode, parentLast bool) {
	sort.Slice(children, func(i, j int) bool {
		return children[i].node.Label < children[j].node.Label
	})

	indent := "  "
	if !parentLast {
		indent = "\u2502   " // │
	} else {
		indent = "    "
	}

	for i, child := range children {
		isLast := i == len(children)-1
		connector := "\u251C\u2500\u2500 "
		if isLast {
			connector = "\u2514\u2500\u2500 "
		}

		b.WriteString(fmt.Sprintf("  %s%s%s\n", indent, connector, child.node.Label))
	}
}

func topEdges(edges []ir.Edge, n int) []ir.Edge {
	if len(edges) <= n {
		return edges
	}
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].Weight > edges[j].Weight
	})
	return edges[:n]
}

func metricVal(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "\u2026"
}

func sysIcon(t string) string {
	switch t {
	case "database":
		return "\U0001F5C4" // 🗄
	case "cache":
		return "\U0001F4BF" // 💿
	case "queue":
		return "\U0001F4E8" // 📨
	case "storage":
		return "\U0001F4E6" // 📦
	case "api":
		return "\U0001F310" // 🌐
	default:
		return "\U0001F517" // 🔗
	}
}