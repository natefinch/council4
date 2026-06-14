package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func buildReport(results []result) report {
	output := report{CardCount: len(results)}
	for _, result := range results {
		if result.exclusion != "" {
			output.Excluded = append(output.Excluded, excluded{
				ID:       result.card.ID,
				OracleID: result.card.OracleID,
				Name:     result.card.Name,
				Layout:   result.card.Layout,
				Reason:   result.exclusion,
			})
			continue
		}
		output.EligibleCount++
		if result.err == nil && len(result.diagnostics) == 0 {
			output.GeneratedCount++
			continue
		}
		diagnostics := result.diagnostics
		if result.err != nil {
			diagnostics = []shared.Diagnostic{{
				Severity: shared.SeverityError,
				Summary:  "source generation failed",
				Detail:   result.err.Error(),
			}}
		}
		output.Unsupported = append(output.Unsupported, unsupported{
			ID:          result.card.ID,
			OracleID:    result.card.OracleID,
			Name:        result.card.Name,
			Layout:      result.card.Layout,
			Diagnostics: reportDiagnostics(diagnostics),
		})
	}
	output.UnsupportedCount = len(output.Unsupported)
	output.ExcludedCount = len(output.Excluded)
	return output
}

func reportDiagnostics(diagnostics []shared.Diagnostic) []reportDiagnostic {
	output := make([]reportDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		output = append(output, reportDiagnostic{
			Severity: diagnosticSeverityName(diagnostic.Severity),
			Summary:  diagnostic.Summary,
			Detail:   diagnostic.Detail,
			Span:     diagnostic.Span,
		})
	}
	return output
}

func diagnosticSeverityName(severity shared.Severity) string {
	switch severity {
	case shared.SeverityError:
		return "error"
	case shared.SeverityWarning:
		return "warning"
	default:
		return "unknown"
	}
}

func writeReport(path, reportFormat string, output report) error {
	writer := io.Writer(os.Stdout)
	var file *os.File
	if path != "-" {
		var err error
		file, err = os.Create(path)
		if err != nil {
			return fmt.Errorf("creating report: %w", err)
		}
		defer file.Close()
		writer = file
	}
	switch reportFormat {
	case "json":
		encoder := json.NewEncoder(writer)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return fmt.Errorf("writing JSON report: %w", err)
		}
	case "text":
		if _, err := fmt.Fprintf(
			writer,
			"cards: %d\neligible: %d\ngenerated: %d\nunsupported: %d\nexcluded: %d\n",
			output.CardCount,
			output.EligibleCount,
			output.GeneratedCount,
			output.UnsupportedCount,
			output.ExcludedCount,
		); err != nil {
			return fmt.Errorf("writing text report summary: %w", err)
		}
		for _, card := range output.Unsupported {
			for _, diagnostic := range card.Diagnostics {
				if _, err := fmt.Fprintf(
					writer,
					"%s\t%s\t%s\n",
					card.Name,
					diagnostic.Summary,
					diagnostic.Detail,
				); err != nil {
					return fmt.Errorf("writing text report: %w", err)
				}
			}
		}
		for _, card := range output.Excluded {
			if _, err := fmt.Fprintf(writer, "%s\texcluded\t%s\n", card.Name, card.Reason); err != nil {
				return fmt.Errorf("writing text report exclusion: %w", err)
			}
		}
	default:
		return fmt.Errorf("unsupported report format %q", reportFormat)
	}
	return nil
}
