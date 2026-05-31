// Command cardbatch manages resumable card-generation batches.
package main

import (
	"flag"
	"fmt"
	"os"

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
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
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
		return fmt.Errorf("parse requires -in")
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
	fmt.Fprintf(os.Stderr, "wrote %s with %d unique cards\n", *outPath, len(manifest.Cards))
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
	fmt.Fprintf(os.Stderr, "updated %s\n", *outPath)
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
	for _, card := range manifest.Cards {
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

func usage() {
	fmt.Fprintln(os.Stderr, "usage: cardbatch <parse|fetch|missing> [flags]")
}
