package render

import (
	"fmt"
	"strings"

	"github.com/DanDo385/map-axi/internal/ir"
)

// DiffResult holds the comparison between two architecture snapshots.
type DiffResult struct {
	AddedNodes      []ir.Node
	RemovedNodes    []ir.Node
	ChangedNodes    []NodeChange
	AddedEdges      []ir.Edge
	RemovedEdges    []ir.Edge
	AddedEntrypoints []string
	RemovedEntrypoints []string
}

// NodeChange represents a node whose metrics changed between snapshots.
type NodeChange struct {
	ID   string
	Old  ir.Metrics
	New  ir.Metrics
}

// Diff compares two architecture snapshots and returns the delta.
func Diff(base, head *ir.Architecture) *DiffResult {
	d := &DiffResult{}

	baseNodes := make(map[string]ir.Node)
	for _, n := range base.Nodes {
		baseNodes[n.ID] = n
	}
	headNodes := make(map[string]ir.Node)
	for _, n := range head.Nodes {
		headNodes[n.ID] = n
	}

	// Node diffs
	for id, hn := range headNodes {
		if bn, ok := baseNodes[id]; ok {
			if bn.Metrics != hn.Metrics {
				d.ChangedNodes = append(d.ChangedNodes, NodeChange{
					ID:  id,
					Old: bn.Metrics,
					New: hn.Metrics,
				})
			}
		} else {
			d.AddedNodes = append(d.AddedNodes, hn)
		}
	}
	for id, bn := range baseNodes {
		if _, ok := headNodes[id]; !ok {
			d.RemovedNodes = append(d.RemovedNodes, bn)
		}
	}

	// Edge diffs
	baseEdges := make(map[string]ir.Edge)
	for _, e := range base.Edges {
		baseEdges[edgeKey(e)] = e
	}
	headEdges := make(map[string]ir.Edge)
	for _, e := range head.Edges {
		headEdges[edgeKey(e)] = e
	}
	for k, e := range headEdges {
		if _, ok := baseEdges[k]; !ok {
			d.AddedEdges = append(d.AddedEdges, e)
		}
	}
	for k, e := range baseEdges {
		if _, ok := headEdges[k]; !ok {
			d.RemovedEdges = append(d.RemovedEdges, e)
		}
	}

	// Entrypoint diffs
	d.AddedEntrypoints = setDiff(head.Entrypoints, base.Entrypoints)
	d.RemovedEntrypoints = setDiff(base.Entrypoints, head.Entrypoints)

	return d
}

// RenderDiff formats the diff as a terminal-friendly string.
func RenderDiff(d *DiffResult) string {
	var b strings.Builder

	b.WriteString("\n  \033[1;35mArchitecture Diff\033[0m\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", strings.Repeat("\u2500", 40)))

	if len(d.AddedNodes) > 0 {
		b.WriteString(fmt.Sprintf("  \033[32m+%d nodes added\033[0m\n", len(d.AddedNodes)))
		for _, n := range d.AddedNodes {
			b.WriteString(fmt.Sprintf("    + %s\n", n.Label))
		}
		b.WriteString("\n")
	}

	if len(d.RemovedNodes) > 0 {
		b.WriteString(fmt.Sprintf("  \033[31m-%d nodes removed\033[0m\n", len(d.RemovedNodes)))
		for _, n := range d.RemovedNodes {
			b.WriteString(fmt.Sprintf("    - %s\n", n.Label))
		}
		b.WriteString("\n")
	}

	if len(d.ChangedNodes) > 0 {
		b.WriteString(fmt.Sprintf("  \033[33m~%d nodes changed\033[0m\n", len(d.ChangedNodes)))
		for _, c := range d.ChangedNodes {
			b.WriteString(fmt.Sprintf("    ~ %s: LOC %d\u2192%d, complexity %d\u2192%d\n",
				c.ID,
				c.Old.LOC, c.New.LOC,
				c.Old.Complexity, c.New.Complexity,
			))
		}
		b.WriteString("\n")
	}

	if len(d.AddedEdges) > 0 {
		b.WriteString(fmt.Sprintf("  \033[32m+%d edges added\033[0m\n", len(d.AddedEdges)))
	}
	if len(d.RemovedEdges) > 0 {
		b.WriteString(fmt.Sprintf("  \033[31m-%d edges removed\033[0m\n", len(d.RemovedEdges)))
	}

	b.WriteString(fmt.Sprintf("\n  Summary: +%d/-%d nodes, +%d/-%d edges, ~%d changed\n",
		len(d.AddedNodes), len(d.RemovedNodes),
		len(d.AddedEdges), len(d.RemovedEdges),
		len(d.ChangedNodes),
	))

	return b.String()
}

func edgeKey(e ir.Edge) string {
	return fmt.Sprintf("%s|%s|%s", e.Source, e.Target, e.Type)
}

func setDiff(a, b []string) []string {
	bSet := make(map[string]bool)
	for _, s := range b {
		bSet[s] = true
	}
	var diff []string
	for _, s := range a {
		if !bSet[s] {
			diff = append(diff, s)
		}
	}
	return diff
}