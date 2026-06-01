package cardgen

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// UnsupportedReport summarizes cards that are not currently ready for runtime
// use after a batch attempt.
type UnsupportedReport struct {
	Summary       UnsupportedReportSummary   `json:"summary"`
	Cards         []UnsupportedReportCard    `json:"cards"`
	Functionality []UnsupportedFunctionality `json:"functionality,omitempty"`
}

// UnsupportedReportSummary counts cards by batch outcome.
type UnsupportedReportSummary struct {
	ManifestTotal        int `json:"manifest_total"`
	UnsupportedTotal     int `json:"unsupported_total"`
	FetchError           int `json:"fetch_error"`
	MissingGeneratedFile int `json:"missing_generated_file"`
	Invalid              int `json:"invalid"`
	ValidationPending    int `json:"validation_pending"`
	FunctionalityBlocked int `json:"functionality_blocked"`
}

// UnsupportedReportCard is one unsupported-card row in the report.
type UnsupportedReportCard struct {
	Name          string            `json:"name"`
	InputName     string            `json:"input_name,omitempty"`
	Section       string            `json:"section,omitempty"`
	Quantity      int               `json:"quantity"`
	FirstLine     int               `json:"first_line"`
	Status        string            `json:"status,omitempty"`
	FileStatus    string            `json:"file_status,omitempty"`
	Validation    string            `json:"validation,omitempty"`
	FilePath      string            `json:"file_path,omitempty"`
	FetchError    string            `json:"fetch_error,omitempty"`
	OracleText    string            `json:"oracle_text,omitempty"`
	Faces         []ManifestFace    `json:"faces,omitempty"`
	Issues        []ValidationIssue `json:"issues,omitempty"`
	Functionality []string          `json:"functionality,omitempty"`
	NextWork      []string          `json:"next_work,omitempty"`
}

// UnsupportedFunctionality groups a reusable missing rules/parser capability
// with the cards that would use it.
type UnsupportedFunctionality struct {
	Capability string   `json:"capability"`
	Cards      []string `json:"cards"`
	Details    []string `json:"details,omitempty"`
}

// BuildUnsupportedReport builds a report from manifest rows that still need
// work: fetch errors, missing generated files, or invalid validation.
func BuildUnsupportedReport(manifest Manifest) UnsupportedReport {
	return BuildUnsupportedReportWithSource(manifest, "")
}

// BuildUnsupportedReportWithSource builds a report and, when repoRoot is set,
// reads generated card source comments to include missing functionality notes.
func BuildUnsupportedReportWithSource(manifest Manifest, repoRoot string) UnsupportedReport {
	report := UnsupportedReport{}
	for _, card := range manifest.Cards {
		report.Summary.ManifestTotal++
		functionality := missingFunctionalityForCard(repoRoot, card.FilePath)
		row, ok := unsupportedReportCard(card, functionality)
		if !ok {
			continue
		}
		report.Cards = append(report.Cards, row)
		report.Summary.UnsupportedTotal++
		if len(row.Functionality) > 0 {
			report.Summary.FunctionalityBlocked++
		}
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
	report.Functionality = rollupFunctionality(report.Cards)
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
	if _, err := fmt.Fprintf(w, "Manifest cards: %d; unsupported or pending: %d; fetch errors: %d; missing generated files: %d; invalid generated cards: %d; validation pending: %d; functionality-blocked cards: %d.\n", report.Summary.ManifestTotal, report.Summary.UnsupportedTotal, report.Summary.FetchError, report.Summary.MissingGeneratedFile, report.Summary.Invalid, report.Summary.ValidationPending, report.Summary.FunctionalityBlocked); err != nil {
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
		if len(card.Functionality) > 0 {
			if _, err := fmt.Fprintln(w, "- Missing functionality:"); err != nil {
				return err
			}
			for _, functionality := range card.Functionality {
				if _, err := fmt.Fprintf(w, "  - %s\n", functionality); err != nil {
					return err
				}
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
	if len(report.Functionality) > 0 {
		if _, err := fmt.Fprintln(w, "\n## Missing functionality rollup"); err != nil {
			return err
		}
		for _, functionality := range report.Functionality {
			if _, err := fmt.Fprintf(w, "\n### %s\n\n", functionality.Capability); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "- Cards: %s\n", strings.Join(functionality.Cards, ", ")); err != nil {
				return err
			}
			if len(functionality.Details) > 0 {
				if _, err := fmt.Fprintln(w, "- Details:"); err != nil {
					return err
				}
				for _, detail := range functionality.Details {
					if _, err := fmt.Fprintf(w, "  - %s\n", detail); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func unsupportedReportCard(card ManifestCard, functionality []string) (UnsupportedReportCard, bool) {
	if card.Status != BatchStatusFetchError && card.FileStatus != BatchFileStatusMissing && card.Validation != BatchValidationStatusInvalid && !(card.FileStatus == BatchFileStatusExisting && card.Validation == BatchValidationStatusUnvalidated) && len(functionality) == 0 {
		return UnsupportedReportCard{}, false
	}
	name := manifestCardName(card)
	row := UnsupportedReportCard{
		Name:          name,
		InputName:     card.InputName,
		Section:       card.Section,
		Quantity:      card.Quantity,
		FirstLine:     card.FirstLine,
		Status:        card.Status,
		FileStatus:    card.FileStatus,
		Validation:    card.Validation,
		FilePath:      card.FilePath,
		FetchError:    card.FetchError,
		OracleText:    card.OracleText,
		Faces:         append([]ManifestFace(nil), card.Faces...),
		Issues:        append([]ValidationIssue(nil), card.Issues...),
		Functionality: append([]string(nil), functionality...),
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
	if len(card.Functionality) > 0 {
		add("Implement or model the missing rules/parser functionality listed for this card.")
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
	case IssueImplementationRequired:
		return "Replace the ImplementationID escape hatch with declarative parser/rules support, or register hand-written runtime support."
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

func missingFunctionalityForCard(repoRoot string, filePath string) []string {
	if repoRoot == "" || filePath == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(repoRoot, filePath))
	if err != nil {
		return nil
	}
	return parseMissingFunctionalityComments(string(data))
}

func parseMissingFunctionalityComments(source string) []string {
	lines := strings.Split(source, "\n")
	var items []string
	for i := 0; i < len(lines); i++ {
		text, ok := lineCommentText(lines[i])
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(text)
		lower := strings.ToLower(trimmed)
		switch {
		case strings.HasPrefix(lower, "missing primitive"):
			collected, next := collectBulletCommentBlock(lines, i+1)
			items = append(items, collected...)
			i = next
		case strings.HasPrefix(lower, "note:"):
			note, next := collectPlainCommentBlock(lines, i)
			if noteMentionsMissingFunctionality(note) {
				items = append(items, strings.TrimSpace(strings.TrimPrefix(note, "Note:")))
			}
			i = next
		}
	}
	return dedupeStrings(items)
}

func collectBulletCommentBlock(lines []string, start int) ([]string, int) {
	var items []string
	current := ""
	i := start
	for ; i < len(lines); i++ {
		text, ok := lineCommentText(lines[i])
		if !ok {
			break
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			break
		}
		if strings.HasPrefix(trimmed, "- ") {
			if current != "" {
				items = append(items, current)
			}
			current = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			continue
		}
		if current == "" {
			continue
		}
		current += " " + trimmed
	}
	if current != "" {
		items = append(items, current)
	}
	return items, i
}

func collectPlainCommentBlock(lines []string, start int) (string, int) {
	var parts []string
	i := start
	for ; i < len(lines); i++ {
		text, ok := lineCommentText(lines[i])
		if !ok {
			break
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			break
		}
		parts = append(parts, trimmed)
	}
	return strings.Join(parts, " "), i
}

func lineCommentText(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "//") {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, "//")), true
}

func noteMentionsMissingFunctionality(note string) bool {
	lower := strings.ToLower(note)
	return strings.Contains(lower, "primitive") ||
		strings.Contains(lower, "implementationid") ||
		strings.Contains(lower, "full rules accuracy") ||
		strings.Contains(lower, "cannot be expressed")
}

func rollupFunctionality(cards []UnsupportedReportCard) []UnsupportedFunctionality {
	type entry struct {
		cards   map[string]bool
		details map[string]bool
	}
	grouped := map[string]*entry{}
	for _, card := range cards {
		for _, detail := range card.Functionality {
			capability := functionalityCapability(detail)
			if grouped[capability] == nil {
				grouped[capability] = &entry{cards: map[string]bool{}, details: map[string]bool{}}
			}
			grouped[capability].cards[card.Name] = true
			grouped[capability].details[detail] = true
		}
	}
	capabilities := make([]string, 0, len(grouped))
	for capability := range grouped {
		capabilities = append(capabilities, capability)
	}
	sort.Strings(capabilities)
	rollup := make([]UnsupportedFunctionality, 0, len(capabilities))
	for _, capability := range capabilities {
		entry := grouped[capability]
		rollup = append(rollup, UnsupportedFunctionality{
			Capability: capability,
			Cards:      sortedKeys(entry.cards),
			Details:    sortedKeys(entry.details),
		})
	}
	return rollup
}

var functionalityIdentifierRE = regexp.MustCompile(`\b[A-Z][A-Za-z0-9]+(?:\.[A-Za-z0-9]+)?\b`)

func functionalityCapability(detail string) string {
	lower := strings.ToLower(detail)
	switch {
	case strings.Contains(lower, "auto-derive subtype mana"):
		return "subtype mana ability derivation"
	case strings.Contains(lower, "conditional etb-tapped"):
		return "conditional ETB-tapped replacement"
	case strings.Contains(lower, "shuffle a permanent into its owner's library"):
		return "shuffle permanent into owner's library"
	case strings.Contains(lower, "creature-sourced damage"):
		return "creature-sourced damage"
	case strings.Contains(lower, "chooser"):
		return "target chooser"
	case strings.Contains(lower, "pay life to suppress enters-tapped"):
		return "pay-life ETB replacement"
	}
	if strings.HasPrefix(strings.TrimSpace(detail), "\"") {
		if quoted, ok := firstQuotedPhrase(detail); ok {
			return quoted
		}
	}
	for _, identifier := range functionalityIdentifierRE.FindAllString(detail, -1) {
		if ignoredFunctionalityIdentifier(identifier) {
			continue
		}
		return identifier
	}
	if quoted, ok := firstQuotedPhrase(detail); ok {
		return quoted
	}
	if before, _, ok := strings.Cut(detail, ";"); ok {
		return strings.TrimSpace(before)
	}
	if before, _, ok := strings.Cut(detail, "."); ok {
		return strings.TrimSpace(before)
	}
	return strings.TrimSpace(detail)
}

func ignoredFunctionalityIdentifier(identifier string) bool {
	switch identifier {
	case "A", "An", "The", "There", "No", "ImplementationID", "CardDef":
		return true
	default:
		return false
	}
}

func firstQuotedPhrase(detail string) (string, bool) {
	start := strings.Index(detail, "\"")
	if start < 0 {
		return "", false
	}
	rest := detail[start+1:]
	end := strings.Index(rest, "\"")
	if end < 0 {
		return "", false
	}
	return rest[:end], true
}

func sortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}

func dedupeStrings(values []string) []string {
	seen := map[string]bool{}
	var result []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
