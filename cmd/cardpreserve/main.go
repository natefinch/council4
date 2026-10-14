// Command cardpreserve checks pinned compiler/registry preservation without changing source.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/natefinch/council4/internal/preservation"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("cardpreserve", flag.ContinueOnError)
	baselinePath := flags.String("baseline", "testdata/card-preservation/baseline.json", "preserved baseline")
	candidatePath := flags.String("candidate", "", "candidate inventory")
	exceptionsPath := flags.String("exceptions", "testdata/card-preservation/exceptions.json", "explicit loss acknowledgments")
	reportPath := flags.String("report", "", "complete delta report")
	registryOnly := flags.Bool("registry-only", false, "capture/check the registry without compiling or downloading the corpus")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *candidatePath == "" || *reportPath == "" {
		return errors.New("-candidate and -report are required")
	}
	baseline, err := preservation.Read[preservation.Manifest](*baselinePath)
	if err != nil {
		return err
	}
	candidate := baseline
	if *registryOnly {
		candidate.Registry, err = preservation.RegistryInventory(baseline.Compiler)
		if err != nil {
			return err
		}
		if err := preservation.Write(*candidatePath, candidate); err != nil {
			return err
		}
	} else {
		candidate, err = preservation.Read[preservation.Manifest](*candidatePath)
		if err != nil {
			return err
		}
	}
	exceptions, err := preservation.Read[[]preservation.Exception](*exceptionsPath)
	if err != nil {
		return err
	}
	changes, comparisonError := preservation.Compare(baseline, candidate, exceptions, *registryOnly)
	if err := preservation.Write(*reportPath, struct {
		RegistryOnly bool                  `json:"registry_only"`
		Changes      []preservation.Change `json:"changes"`
		Error        string                `json:"error,omitempty"`
	}{RegistryOnly: *registryOnly, Changes: changes, Error: errorText(comparisonError)}); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(os.Stdout, "%d preservation differences; registry-only=%v; report=%s\n", len(changes), *registryOnly, *reportPath)
	return comparisonError
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
