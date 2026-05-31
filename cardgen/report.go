package cardgen

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// UnsupportedReport summarizes cards that are not currently ready for runtime
// use after a batch attempt.
type UnsupportedReport struct {
	Summary UnsupportedReportSummary `json:"summary"`
	Cards   []UnsupportedReportCard  `json:"cards"`
}

// UnsupportedReportSummary counts cards by batch outcome.
type UnsupportedReportSummary struct {
	ManifestTotal        int `json:"manifest_total"`
	UnsupportedTotal     int `json:"unsupported_total"`
	FetchError           int `json:"fetch_error"`
	MissingGeneratedFile int `json:"missing_generated_file"`
	Invalid              int `json:"invalid"`
	ValidationPending    int `json:"validation_pending"`
}

// UnsupportedReportCard is one unsupported-card row in the report.
type UnsupportedReportCard struct {
	Name       string            `json:"name"`
	InputName  string            `json:"input_name,omitempty"`
	Section    string            `json:"section,omitempty"`
	Quantity   int               `json:"quantity"`
	FirstLine  int               `json:"first_line"`
	Status     string            `json:"status,omitempty"`
	FileStatus string            `json:"file_status,omitempty"`
	Validation string            `json:"validation,omitempty"`
	FilePath   string            `json:"file_path,omitempty"`
	FetchError string            `json:"fetch_error,omitempty"`
	OracleText string            `json:"oracle_text,omitempty"`
	Faces      []ManifestFace    `json:"faces,omitempty"`
	Issues     []ValidationIssue `json:"issues,omitempty"`
	NextWork   []string          `json:"next_work,omitempty"`
}

// BuildUnsupportedReport builds a report from manifest rows that still need
// work: fetch errors, missing generated files, or invalid validation.
func BuildUnsupportedReport(manifest Manifest) UnsupportedReport {
	report := UnsupportedReport{}
	for _, card := range manifest.Cards {
		report.Summary.ManifestTotal++
		row, ok := unsupportedReportCard(card)
		if !ok {
			continue
		}
		report.Cards = append(report.Cards, row)
		report.Summary.UnsupportedTotal++
		switch {
		case card.Status == BatchStatusFetchError:
			report.Summary.FetchError++
		case card.Validation == BatchValidationStatusInvalid:
			report.Summary.Invalid++
		case card.FileStatus == BatchFileStatusMissing:
			report.Summary.MissingGeneratedFile++
		case card.FileStatus == BatchFileStatusExisting && card.Validation == BatchValidationStatusUnvalidated:
			report.Summary.ValidationPending++
		}
	}
	sort.SliceStable(report.Cards, func(i, j int) bool {
		return report.Cards[i].Name < report.Cards[j].Name
	})
	return report
}

// WriteUnsupportedReportJSON writes report as indented JSON.
func WriteUnsupportedReportJSON(w io.Writer, report UnsupportedReport) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// WriteUnsupportedReportMarkdown writes a human-readable unsupported-card report.
func WriteUnsupportedReportMarkdown(w io.Writer, report UnsupportedReport) error {
	if _, err := fmt.Fprintln(w, "# Unsupported card report"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Manifest cards: %d; unsupported or pending: %d; fetch errors: %d; missing generated files: %d; invalid generated cards: %d; validation pending: %d.\n", report.Summary.ManifestTotal, report.Summary.UnsupportedTotal, report.Summary.FetchError, report.Summary.MissingGeneratedFile, report.Summary.Invalid, report.Summary.ValidationPending); err != nil {
		return err
	}
	for _, card := range report.Cards {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "## %s\n\n", card.Name); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "- Quantity: %d\n- Section: %s\n- Source line: %d\n", card.Quantity, emptyDash(card.Section), card.FirstLine); err != nil {
			return err
		}
		if card.FilePath != "" {
			if _, err := fmt.Fprintf(w, "- Expected file: %s\n", inlineCode(card.FilePath)); err != nil {
				return err
			}
		}
		if card.FetchError != "" {
			if _, err := fmt.Fprintf(w, "- Fetch error: %s\n", inlineCode(card.FetchError)); err != nil {
				return err
			}
		}
		for _, issue := range card.Issues {
			if _, err := fmt.Fprintf(w, "- %s: %s", inlineCode(string(issue.Code)), issue.Message); err != nil {
				return err
			}
			if issue.Path != "" {
				if _, err := fmt.Fprintf(w, " (%s)", inlineCode(issue.Path)); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if len(card.NextWork) > 0 {
			if _, err := fmt.Fprintln(w, "- Suggested next work:"); err != nil {
				return err
			}
			for _, next := range card.NextWork {
				if _, err := fmt.Fprintf(w, "  - %s\n", next); err != nil {
					return err
				}
			}
		}
		if card.OracleText != "" {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w, "Oracle text:"); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "> %s\n", quoteOracle(card.OracleText)); err != nil {
				return err
			}
		}
		for _, face := range card.Faces {
			if face.OracleText == "" {
				continue
			}
			if _, err := fmt.Fprintf(w, "\n%s oracle text:\n\n> %s\n", face.Name, quoteOracle(face.OracleText)); err != nil {
				return err
			}
		}
	}
	return nil
}

func unsupportedReportCard(card ManifestCard) (UnsupportedReportCard, bool) {
	if card.Status != BatchStatusFetchError && card.FileStatus != BatchFileStatusMissing && card.Validation != BatchValidationStatusInvalid && !(card.FileStatus == BatchFileStatusExisting && card.Validation == BatchValidationStatusUnvalidated) {
		return UnsupportedReportCard{}, false
	}
	name := manifestCardName(card)
	row := UnsupportedReportCard{
		Name:       name,
		InputName:  card.InputName,
		Section:    card.Section,
		Quantity:   card.Quantity,
		FirstLine:  card.FirstLine,
		Status:     card.Status,
		FileStatus: card.FileStatus,
		Validation: card.Validation,
		FilePath:   card.FilePath,
		FetchError: card.FetchError,
		OracleText: card.OracleText,
		Faces:      append([]ManifestFace(nil), card.Faces...),
		Issues:     append([]ValidationIssue(nil), card.Issues...),
	}
	row.NextWork = nextWorkForCard(row)
	return row, true
}

func nextWorkForCard(card UnsupportedReportCard) []string {
	seen := map[string]bool{}
	add := func(text string) {
		if text != "" && !seen[text] {
			seen[text] = true
			card.NextWork = append(card.NextWork, text)
		}
	}
	if card.FetchError != "" {
		add("Fix the card-list entry, Scryfall fetch/layout support, or oracle cache before implementation.")
	}
	if card.FetchError == "" && card.FileStatus == BatchFileStatusMissing {
		add("Run the card through card-impl, regenerate mtg/cards package lists, then validate again.")
	}
	if card.FileStatus == BatchFileStatusExisting && card.Validation == BatchValidationStatusUnvalidated {
		add("Run cardbatch validate so this existing generated card is checked before use.")
	}
	for _, issue := range card.Issues {
		add(nextWorkForIssue(issue.Code))
	}
	return card.NextWork
}

func nextWorkForIssue(code ValidationCode) string {
	switch code {
	case IssueOracleWithoutAbilities:
		return "Fill generated AbilityDef data for this oracle text or add a hand-written ImplementationID."
	case IssueUnexecutedEffect:
		return "Implement rules execution for the generated effect type or map the card to a hand-written implementation."
	case IssueMissingSearchSpec, IssueUnsupportedSearchSpec:
		return "Expand SearchSpec modeling or search-effect rules for this tutor/reveal pattern."
	case IssueTargetIndexOutOfRange, IssueInvalidTargetSpec:
		return "Fix target parsing/generation so effects and target specs line up."
	case IssueUnregisteredImplementation:
		return "Register or remove the hand-written ImplementationID."
	case IssueGeneratedCardNotFound:
		return "Regenerate mtg/cards package lists and ensure the CardDef name matches the Scryfall canonical name."
	case IssueValidationRunFailed:
		return "Fix generated package compile/import errors, then validate again."
	default:
		return "Review the generated CardDef and add missing parser or rules support."
	}
}

func emptyDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func quoteOracle(text string) string {
	text = strings.TrimRight(text, "\n")
	lines := strings.Split(text, "\n")
	return strings.Join(lines, "\n> ")
}

func inlineCode(text string) string {
	if strings.Contains(text, "`") {
		return "`` " + text + " ``"
	}
	return "`" + text + "`"
}
