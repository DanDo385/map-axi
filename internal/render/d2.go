package render

import (
	"fmt"
	"strings"

	"github.com/DanDo385/map-axi/internal/ir"
)

// RenderD2 produces a D2 diagram of the architecture.
// D2 (d2lang.com) is a declarative diagramming language.
func RenderD2(arch *ir.Architecture) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s architecture\n", arch.Name))
	b.WriteString("direction: right\n\n")

	// Nodes as D2 shapes
	for _, n := range arch.Nodes {
		shape := d2Shape(n.Type)
		label := sanitizeD2Label(n.Label)
		b.WriteString(fmt.Sprintf("%s: {\n", d2ID(n.ID)))
		b.WriteString(fmt.Sprintf("  label: %q\n", label))
		b.WriteString(fmt.Sprintf("  shape: %s\n", shape))
		if n.Metrics.LOC > 0 {
			b.WriteString(fmt.Sprintf("  width: %d\n", d2Width(n.Metrics.LOC)))
			b.WriteString(fmt.Sprintf("  height: %d\n", d2Height(n.Metrics.LOC)))
		}
		b.WriteString("}\n\n")
	}

	// Edges as connections
	for _, e := range arch.Edges {
		label := d2EdgeLabel(e)
		b.WriteString(fmt.Sprintf("%s -> %s: %q\n",
			d2ID(e.Source), d2ID(e.Target), label))
	}

	// External systems
	for _, es := range arch.ExternalSystems {
		esID := d2ID("ext:" + es.Name)
		b.WriteString(fmt.Sprintf("%s: {\n", esID))
		b.WriteString(fmt.Sprintf("  label: %q\n", es.Name))
		b.WriteString(fmt.Sprintf("  shape: %s\n", d2SysShape(es.Type)))
		b.WriteString(fmt.Sprintf("  style.fill: %q\n", d2SysColor(es.Type)))
		b.WriteString("}\n\n")
		for _, conn := range es.Connections {
			b.WriteString(fmt.Sprintf("%s -> %s\n", d2ID(conn), esID))
		}
	}

	return b.String()
}

func d2ID(raw string) string {
	s := strings.NewReplacer(
		"/", "_",
		".", "_",
		"-", "_",
		":", "_",
		" ", "_",
	).Replace(raw)
	return strings.TrimPrefix(s, "_")
}

func d2Shape(kind string) string {
	switch kind {
	case "package":
		return "package"
	case "class":
		return "class"
	case "function":
		return "rectangle"
	case "module":
		return "folder"
	default:
		return "rectangle"
	}
}

func d2SysShape(t string) string {
	switch t {
	case "database":
		return "cylinder"
	case "queue":
		return "queue"
	default:
		return "hexagon"
	}
}

func d2SysColor(t string) string {
	switch t {
	case "database":
		return "#FFE082"
	case "cache":
		return "#EF9A9A"
	case "queue":
		return "#A5D6A7"
	case "api":
		return "#90CAF9"
	default:
		return "#E0E0E0"
	}
}

func d2Width(loc int) int {
	w := 200 + loc*2
	if w > 800 {
		w = 800
	}
	return w
}

func d2Height(loc int) int {
	h := 80 + loc
	if h > 400 {
		h = 400
	}
	return h
}

func d2EdgeLabel(e ir.Edge) string {
	if e.Weight > 1 {
		return fmt.Sprintf("%s (x%d)", e.Type, e.Weight)
	}
	return e.Type
}

func sanitizeD2Label(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}