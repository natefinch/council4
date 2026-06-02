// Command cardbatch manages resumable card-generation batches.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/natefinch/council4/cardgen"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "parse":
		err = runParse(os.Args[2:])
	case "fetch":
		err = runFetch(os.Args[2:])
	case "missing":
		err = runMissing(os.Args[2:])
	case "worklist":
		err = runWorklist(os.Args[2:])
	case "validate":
		err = runValidate(os.Args[2:])
	case "report":
		err = runReport(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func runParse(args []string) error {
	flags := flag.NewFlagSet("parse", flag.ExitOnError)
	inPath := flags.String("in", "", "card list input file")
	outPath := flags.String("out", ".cardwork/cards.json", "manifest output path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *inPath == "" {
		return errors.New("parse requires -in")
	}
	file, err := os.Open(*inPath)
	if err != nil {
		return err
	}
	defer file.Close()
	items, err := cardgen.ParseCardList(file)
	if err != nil {
		return err
	}
	manifest := cardgen.NewManifestFromItems(items)
	if err := cardgen.SaveManifest(*outPath, manifest); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stderr, "wrote %s with %d unique cards\n", *outPath, len(manifest.Cards))
	return nil
}

func runFetch(args []string) error {
	flags := flag.NewFlagSet("fetch", flag.ExitOnError)
	manifestPath := flags.String("manifest", ".cardwork/cards.json", "manifest path")
	outPath := flags.String("out", "", "manifest output path; defaults to -manifest")
	cacheDir := flags.String("cache", ".cardwork/cache/scryfall", "Scryfall JSON cache directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outPath == "" {
		*outPath = *manifestPath
	}
	manifest, err := cardgen.LoadManifest(*manifestPath)
	if err != nil {
		return err
	}
	cardgen.FetchManifest(&manifest, *cacheDir)
	if err := cardgen.SaveManifest(*outPath, manifest); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stderr, "updated %s\n", *outPath)
	return nil
}

func runMissing(args []string) error {
	flags := flag.NewFlagSet("missing", flag.ExitOnError)
	manifestPath := flags.String("manifest", ".cardwork/cards.json", "manifest path")
	outPath := flags.String("out", "", "manifest output path; defaults to -manifest")
	repoRoot := flags.String("repo", ".", "repository root")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outPath == "" {
		*outPath = *manifestPath
	}
	manifest, err := cardgen.LoadManifest(*manifestPath)
	if err != nil {
		return err
	}
	cardgen.MarkExistingFiles(&manifest, *repoRoot)
	for i := range manifest.Cards {
		card := &manifest.Cards[i]
		if card.FileStatus == cardgen.BatchFileStatusMissing {
			name := card.CanonicalName
			if name == "" {
				name = card.InputName
			}
			fmt.Println(name)
		}
	}
	return cardgen.SaveManifest(*outPath, manifest)
}

func runWorklist(args []string) error {
	flags := flag.NewFlagSet("worklist", flag.ExitOnError)
	manifestPath := flags.String("manifest", ".cardwork/cards.json", "manifest path")
	repoRoot := flags.String("repo", ".", "repository root")
	limit := flags.Int("limit", 0, "maximum number of cards to print; 0 means all")
	format := flags.String("format", "names", "output format: names or commands")
	if err := flags.Parse(args); err != nil {
		return err
	}
	manifest, err := cardgen.LoadManifest(*manifestPath)
	if err != nil {
		return err
	}
	cardgen.MarkExistingFiles(&manifest, *repoRoot)
	for _, name := range cardgen.MissingWorklist(manifest, *limit) {
		switch *format {
		case "names":
			fmt.Println(name)
		case "commands":
			fmt.Printf("go run .agents/skills/card-impl/main.go %s\n", shellQuote(name))
		default:
			return fmt.Errorf("unknown worklist format %q", *format)
		}
	}
	return nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func runValidate(args []string) error {
	flags := flag.NewFlagSet("validate", flag.ExitOnError)
	manifestPath := flags.String("manifest", ".cardwork/cards.json", "manifest path")
	outPath := flags.String("out", "", "manifest output path; defaults to -manifest")
	repoRoot := flags.String("repo", ".", "repository root")
	generate := flags.Bool("generate", false, "run go generate ./mtg/cards/... before validation")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outPath == "" {
		*outPath = *manifestPath
	}
	if *generate {
		if err := cardgen.RunGoGenerateCards(*repoRoot); err != nil {
			return err
		}
	}
	manifest, err := cardgen.LoadManifest(*manifestPath)
	if err != nil {
		return err
	}
	cardgen.MarkExistingFiles(&manifest, *repoRoot)
	if err := cardgen.ValidateManifestGeneratedCards(&manifest, *repoRoot); err != nil {
		return err
	}
	return cardgen.SaveManifest(*outPath, manifest)
}

func runReport(args []string) error {
	flags := flag.NewFlagSet("report", flag.ExitOnError)
	manifestPath := flags.String("manifest", ".cardwork/cards.json", "manifest path")
	repoRoot := flags.String("repo", ".", "repository root")
	mdPath := flags.String("md", ".cardwork/unsupported.md", "Markdown report path; use - for stdout")
	jsonPath := flags.String("json", ".cardwork/unsupported.json", "JSON report path; use - for stdout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	manifest, err := cardgen.LoadManifest(*manifestPath)
	if err != nil {
		return err
	}
	cardgen.MarkExistingFiles(&manifest, *repoRoot)
	report := cardgen.BuildUnsupportedReportWithSource(manifest, *repoRoot)
	if err := writeReport(*mdPath, func(w *os.File) error {
		return cardgen.WriteUnsupportedReportMarkdown(w, report)
	}); err != nil {
		return err
	}
	return writeReport(*jsonPath, func(w *os.File) error {
		return cardgen.WriteUnsupportedReportJSON(w, report)
	})
}

func writeReport(path string, write func(*os.File) error) error {
	if path == "" {
		return nil
	}
	if path == "-" {
		return write(os.Stdout)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := write(file); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, "usage: cardbatch <parse|fetch|missing|worklist|validate|report> [flags]")
}
