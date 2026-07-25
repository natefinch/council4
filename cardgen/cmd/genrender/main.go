// Command genrender generates the parts of the cardgen renderer that are fully
// determined by the shape of the mtg/game value types.
//
// The renderer emits Go source for generated card definitions, so every named
// constant a game value can hold needs a value-to-constant-name mapping. Writing
// those mappings by hand meant adding a constant to mtg/game silently left the
// renderer unable to emit it: half of game.Step, zone.Stack, and
// game.ResolutionChoiceCardName all had no rendering, so any lowering that
// produced one would have failed to render and dropped the card.
//
// Usage (typically via go generate from the repository root):
//
//	go generate ./cardgen/...
//
// The command must run with its working directory inside this module: it
// type-checks mtg/game from source using the standard library's source importer,
// which resolves imports through go/build.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// generatedFile is the generated file's path relative to the repository root.
const (
	generatedFile       = "cardgen/render_literals_generated.go"
	generatedStructFile = "cardgen/render_structs_generated.go"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "genrender: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("genrender", flag.ContinueOnError)
	root := flags.String("root", ".", "repository root containing mtg/game")
	list := flags.Bool("list", false, "print the reachable type graph instead of writing the generated file")
	fields := flags.Bool("fields", false, "print every reachable struct's fields and their resolved render kinds")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if *list || *fields {
		graph, err := load(*root)
		if err != nil {
			return err
		}
		if *fields {
			writeFieldReport(stdout, graph)
			return nil
		}
		writeSummary(stdout, graph)
		return nil
	}

	graph, err := load(*root)
	if err != nil {
		return err
	}
	for _, out := range generatedOutputs() {
		src, err := out.gen(graph)
		if err != nil {
			return err
		}
		path := filepath.Join(*root, filepath.FromSlash(out.path))
		if err := os.WriteFile(path, src, 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	return nil
}

// generatedOutput is one file this command writes.
type generatedOutput struct {
	path string
	gen  func(*reachable) ([]byte, error)
}

// generatedOutputs lists every file the generator owns. The staleness test walks
// the same list, so a new generated file cannot be added without also being
// covered by the drift guard.
func generatedOutputs() []generatedOutput {
	return []generatedOutput{
		{generatedFile, func(g *reachable) ([]byte, error) { return renderLiterals(g.Enums) }},
		{generatedStructFile, renderStructs},
	}
}

// load type-checks the game packages and returns the type graph the renderer can
// encounter.
func load(root string) (*reachable, error) {
	pkgs, err := loadPackages(root)
	if err != nil {
		return nil, err
	}
	enums := collectEnums(pkgs)
	if len(enums) == 0 {
		return nil, fmt.Errorf("no exported constants found under %s", root)
	}
	graph, err := walk(pkgs, enums)
	if err != nil {
		return nil, err
	}
	if err := checkBitmaskTypes(graph.Enums); err != nil {
		return nil, err
	}
	if err := checkOpenStringTypes(graph, pkgs); err != nil {
		return nil, err
	}
	if err := checkPreRenderValidators(graph); err != nil {
		return nil, err
	}
	if err := checkAlwaysEmit(graph, pkgs); err != nil {
		return nil, err
	}
	return graph, nil
}

// generate returns the formatted contents of the generated literal file.
func generate(root string) ([]byte, error) {
	graph, err := load(root)
	if err != nil {
		return nil, err
	}
	return renderLiterals(graph.Enums)
}

// writeSummary prints what the walk found, for inspecting the generator's view
// of the type graph without writing a file.
func writeSummary(w io.Writer, graph *reachable) {
	_, _ = fmt.Fprintf(w, "reachable: %d enums, %d structs, %d interfaces, %d opaque\n\n",
		len(graph.Enums), len(graph.Structs), len(graph.Interfaces), len(graph.Opaque))
	for _, e := range graph.Enums {
		_, _ = fmt.Fprintln(w, "enum   ", e.describe())
	}
	for _, iface := range graph.Interfaces {
		_, _ = fmt.Fprintf(w, "iface   %-30s %d implementations\n", qualified(iface), len(graph.Implementations[qualified(iface)]))
	}
	for _, named := range graph.Opaque {
		_, _ = fmt.Fprintln(w, "opaque ", qualified(named))
	}
}
