// Package render produces architecture views from the IR:
// ASCII (TOON terminal), D2 declarative diagrams, Mermaid flowcharts,
// and structural diffs between snapshots.
package render

import (
	"fmt"

	"github.com/DanDo385/map-axi/internal/ir"
)

// Format identifies the output format for rendering.
type Format string

const (
	FormatASCII   Format = "ascii"
	FormatD2      Format = "d2"
	FormatMermaid Format = "mermaid"
)

// Render dispatches to the appropriate renderer based on format.
func Render(arch *ir.Architecture, format Format) (string, error) {
	switch format {
	case FormatASCII:
		return RenderASCII(arch), nil
	case FormatD2:
		return RenderD2(arch), nil
	case FormatMermaid:
		return RenderMermaid(arch), nil
	default:
		return "", fmt.Errorf("unknown format: %s (use ascii, d2, or mermaid)", format)
	}
}

// RenderDiffFile loads two architecture files and renders the diff.
func RenderDiffFile(basePath, headPath string) (string, error) {
	base, err := ir.Load(basePath)
	if err != nil {
		return "", fmt.Errorf("load base: %w", err)
	}
	head, err := ir.Load(headPath)
	if err != nil {
		return "", fmt.Errorf("load head: %w", err)
	}
	d := Diff(base, head)
	return RenderDiff(d), nil
}