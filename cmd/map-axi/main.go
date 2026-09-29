// map-axi: architecture visualization tool.
//
// Principles (AXI):
//   - TOON output: text+icon terminal views.
//   - Content-first: default mode shows the architecture summary.
//   - help[]: every view surfaces available next steps.
//
// Usage:
//
//	map-axi                          # home view (reads architecture.json in CWD)
//	map-axi scan [path]              # extract architecture from source
//	map-axi view                     # ASCII terminal tree (collapsed top-level)
//	map-axi view --format d2         # D2 declarative diagram
//	map-axi view --format mermaid    # Mermaid flowchart
//	map-axi diff <base.json>         # structural diff from base snapshot
//	map-axi help                     # full help with next steps
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DanDo385/map-axi/internal/extract"
	"github.com/DanDo385/map-axi/internal/ir"
	"github.com/DanDo385/map-axi/internal/render"
)

const defaultArchFile = "architecture.json"

func main() {
	if len(os.Args) < 2 {
		homeView()
		return
	}

	switch os.Args[1] {
	case "scan":
		scanCmd(os.Args[2:])
	case "view":
		viewCmd(os.Args[2:])
	case "diff":
		diffCmd(os.Args[2:])
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "map-axi: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// homeView is the content-first default: reads architecture.json and shows summary.
func homeView() {
	arch, err := ir.Load(defaultArchFile)
	if err != nil {
		fmt.Printf("\n  \033[1;36mmap-axi\033[0m  \u2014  architecture visualization\n\n")
		fmt.Printf("  No %s found in this directory.\n\n", defaultArchFile)
		fmt.Printf("  \033[1mGet started:\033[0m\n")
		fmt.Printf("    map-axi scan ./src     extract architecture from source\n")
		fmt.Printf("    map-axi help           see all commands\n\n")
		fmt.Printf("  \033[2mWatchexec for live re-render:\033[0m\n")
		fmt.Printf("    watchexec -w src/ -e go,ts,js,py -- map-axi scan && map-axi view\n\n")
		return
	}

	fmt.Print(render.RenderASCII(arch))
	printHelpFooter()
}

func scanCmd(args []string) {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	fmt.Fprintf(os.Stderr, "map-axi: scanning %s ...\n", root)

	arch, err := extract.Run(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map-axi: %v\n", err)
		os.Exit(1)
	}

	if err := arch.Save(defaultArchFile); err != nil {
		fmt.Fprintf(os.Stderr, "map-axi: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  \u2705 %d nodes, %d edges written to %s\n",
		len(arch.Nodes), len(arch.Edges), defaultArchFile)
	printHelpFooter()
}

func viewCmd(args []string) {
	fs := flag.NewFlagSet("view", flag.ExitOnError)
	format := fs.String("format", "ascii", "output format: ascii, d2, mermaid")
	archFile := fs.String("file", defaultArchFile, "architecture JSON file")
	fs.Parse(args)

	arch, err := ir.Load(*archFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map-axi: %v\n", err)
		os.Exit(1)
	}

	out, err := render.Render(arch, render.Format(*format))
	if err != nil {
		fmt.Fprintf(os.Stderr, "map-axi: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out)

	if *format != "ascii" {
		printHelpFooter()
	}
}

func diffCmd(args []string) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "map-axi: diff requires a base architecture.json path\n")
		os.Exit(1)
	}

	headFile := defaultArchFile
	baseFile := args[0]
	if len(args) > 1 {
		headFile = args[1]
	}

	// Resolve relative paths
	if !filepath.IsAbs(baseFile) {
		baseFile = filepath.Join(".", baseFile)
	}

	out, err := render.RenderDiffFile(baseFile, headFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "map-axi: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(out)
	printHelpFooter()
}

func printHelp() {
	fmt.Printf(`
  \033[1;36mmap-axi\033[0m  \u2014  architecture visualization

  \033[1mAXI principles:\033[0m
    TOON output     text+icon terminal views
    Content-first   default view shows architecture immediately
    help[]          every surface exposes next steps

  \033[1mCommands:\033[0m
    map-axi                          home view (reads architecture.json)
    map-axi scan [<root>]            extract architecture from source
    map-axi view                     ASCII terminal tree
    map-axi view --format d2         D2 declarative diagram
    map-axi view --format mermaid    Mermaid flowchart
    map-axi diff <base.json> [head]  structural diff vs base snapshot
    map-axi help                     this help

  \033[1mIR schema:\033[0m
    architecture.json  nodes, edges, entrypoints, routes, externalSystems
                       per-node metrics: loc, complexity, churn, hotspot, coverage

  \033[1mExtractors:\033[0m
    ast-grep (sg)     symbols (functions, classes, modules)
                       dependencies (imports, call graphs)
    languages          Go, TypeScript, JavaScript, Python

  \033[1mRenderers:\033[0m
    ASCII              collapsed top-level tree + metrics summary
    D2                 d2lang.com shapes and connections
    Mermaid            flowchart with subgraphs and edge styles
    Diff               added/removed/changed nodes and edges

  \033[1mLive re-render:\033[0m
    watchexec -w src/ -e go,ts,js,py -- map-axi scan && map-axi view

  \033[1mNext steps:\033[0m
    [] map-axi scan ./src          extract your architecture
    [] map-axi view                see the ASCII tree
    [] map-axi view --format d2    export for d2lang.com
    [] map-axi view --format mermaid  embed in markdown
    [] map-axi diff v1.json        review changes

`)
	printHelpFooter()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: map-axi [scan|view|diff|help]\n")
	fmt.Fprintf(os.Stderr, "Try:   map-axi help\n")
}

func printHelpFooter() {
	fmt.Printf("\n  \033[2mhelp[]\033[0m map-axi help\n\n")
}