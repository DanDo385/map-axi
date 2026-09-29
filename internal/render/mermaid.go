package render

import (
	"fmt"
	"strings"

	"github.com/DanDo385/map-axi/internal/ir"
)

// RenderMermaid produces a Mermaid flowchart diagram.
func RenderMermaid(arch *ir.Architecture) string {
	var b strings.Builder

	b.WriteString("```mermaid\n")
	b.WriteString("flowchart LR\n")

	// Subgraph for each parent group
	groups := groupByParent(arch.Nodes)
	for parent, nodes := range groups {
		if parent == "" {
			parent = arch.Name
		}
		groupID := mermaidID("group_" + parent)
		b.WriteString(fmt.Sprintf("  subgraph %s[%s]\n", groupID, parent))
		for _, n := range nodes {
			nID := mermaidID(n.ID)
			shape := mermaidShape(n.Type)
			label := sanitizeMermaidLabel(n.Label)
			b.WriteString(fmt.Sprintf("    %s%s[%s]\n", nID, shape, label))
		}
		b.WriteString("  end\n\n")
	}

	// Edges
	for _, e := range arch.Edges {
		src := mermaidID(e.Source)
		tgt := mermaidID(e.Target)
		style := mermaidEdgeStyle(e.Type)
		b.WriteString(fmt.Sprintf("  %s %s %s\n", src, style, tgt))
	}

	// External systems
	for _, es := range arch.ExternalSystems {
		esID := mermaidID("ext_" + es.Name)
		shape := mermaidSysShape(es.Type)
		b.WriteString(fmt.Sprintf("  %s%s{{%s}}\n", esID, shape, es.Name))
		for _, conn := range es.Connections {
			b.WriteString(fmt.Sprintf("  %s -.-> %s\n", mermaidID(conn), esID))
		}
	}

	b.WriteString("```\n")
	return b.String()
}

func mermaidID(raw string) string {
	s := strings.NewReplacer(
		"/", "_",
		".", "_",
		"-", "_",
		":", "_",
		" ", "_",
	).Replace(raw)
	return strings.TrimPrefix(s, "_")
}

func mermaidShape(kind string) string {
	switch kind {
	case "function":
		return "" // standard rectangle
	case "class":
		return "{{" // hexagon-ish: needs closing }}
	default:
		return ""
	}
}

func mermaidSysShape(t string) string {
	switch t {
	case "database":
		return "[( " // cylinder: needs closing )]
	default:
		return "{{" // hexagon
	}
}

func mermaidEdgeStyle(edgeType string) string {
	switch edgeType {
	case "import":
		return "-.->"
	case "calls":
		return "-->"
	case "implements":
		return "-.->|implements|"
	case "extends":
		return "==>"
	default:
		return "-->"
	}
}

func sanitizeMermaidLabel(s string) string {
	s = strings.ReplaceAll(s, `"`, "#quot;")
	s = strings.ReplaceAll(s, "[", "#91;")
	s = strings.ReplaceAll(s, "]", "#93;")
	return s
}

func groupByParent(nodes []ir.Node) map[string][]ir.Node {
	groups := make(map[string][]ir.Node)
	for _, n := range nodes {
		groups[n.Parent] = append(groups[n.Parent], n)
	}
	return groups
}