// Command compilecards generates fully executable CardDef source files for the
// strictly supported subset of a Scryfall Oracle Cards bulk-data file.
package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle"
)

type config struct {
	inputPath              string
	outputRoot             string
	reportPath             string
	format                 string
	supportedPath          string
	unsupportedPath        string
	unsupportedReasonsPath string
	readmePath             string
	workers                int
}

type job struct {
	index int
	card  cardgen.ScryfallCard
}

type result struct {
	index       int
	card        cardgen.ScryfallCard
	relative    string
	superseded  string
	source      string
	exclusion   cardgen.CorpusExclusionReason
	diagnostics []oracle.Diagnostic
	err         error
}

type report struct {
	CardCount        int           `json:"card_count"`
	EligibleCount    int           `json:"eligible_count"`
	GeneratedCount   int           `json:"generated_count"`
	UnsupportedCount int           `json:"unsupported_count"`
	ExcludedCount    int           `json:"excluded_count"`
	Unsupported      []unsupported `json:"unsupported"`
	Excluded         []excluded    `json:"excluded"`
}

type unsupported struct {
	ID          string             `json:"id,omitempty"`
	OracleID    string             `json:"oracle_id,omitempty"`
	Name        string             `json:"name"`
	Layout      string             `json:"layout,omitempty"`
	Diagnostics []reportDiagnostic `json:"diagnostics"`
}

type excluded struct {
	ID       string                        `json:"id,omitempty"`
	OracleID string                        `json:"oracle_id,omitempty"`
	Name     string                        `json:"name"`
	Layout   string                        `json:"layout,omitempty"`
	Reason   cardgen.CorpusExclusionReason `json:"reason"`
}

type reportDiagnostic struct {
	Severity string      `json:"severity"`
	Summary  string      `json:"summary"`
	Detail   string      `json:"detail,omitempty"`
	Span     oracle.Span `json:"span"`
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	if err := run(cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func parseFlags(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("compilecards", flag.ContinueOnError)
	flags.StringVar(&cfg.inputPath, "in", "", "Scryfall Oracle Cards bulk-data JSON file")
	flags.StringVar(&cfg.outputRoot, "out", filepath.Join("mtg", "cards"), "output cards package root")
	flags.StringVar(&cfg.reportPath, "report", "-", "unsupported report path, or - for stdout")
	flags.StringVar(&cfg.format, "format", "json", "report format: json or text")
	flags.StringVar(&cfg.supportedPath, "supported", "", "supported-card Markdown path")
	flags.StringVar(&cfg.unsupportedPath, "unsupported", "", "unsupported-card Markdown path")
	flags.StringVar(&cfg.unsupportedReasonsPath, "unsupported-reasons", "", "card-support planning Markdown path")
	flags.StringVar(&cfg.readmePath, "readme", "", "README path whose card-support block should be updated")
	flags.IntVar(&cfg.workers, "workers", runtime.NumCPU(), "number of compiler workers")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func run(cfg config) error {
	if cfg.inputPath == "" {
		return errors.New("-in is required")
	}
	if cfg.workers < 1 {
		return errors.New("-workers must be at least 1")
	}
	if cfg.format != "json" && cfg.format != "text" {
		return fmt.Errorf("unsupported -format %q", cfg.format)
	}
	input, err := os.Open(cfg.inputPath)
	if err != nil {
		return fmt.Errorf("opening input: %w", err)
	}
	defer input.Close()

	results, err := compileCorpus(input, cfg.workers)
	if err != nil {
		return err
	}
	report := buildReport(results)
	if err := writeSupported(cfg.outputRoot, results); err != nil {
		return err
	}
	if err := writeSupportDocumentation(cfg, report, results); err != nil {
		return err
	}
	return writeReport(cfg.reportPath, cfg.format, report)
}

const (
	readmeSupportStart = "<!-- card-support:start -->"
	readmeSupportEnd   = "<!-- card-support:end -->"
)

func writeSupportDocumentation(cfg config, output report, results []result) error {
	if cfg.supportedPath != "" {
		if err := writeSupportedMarkdown(cfg.supportedPath, output, results); err != nil {
			return err
		}
	}
	if cfg.unsupportedPath != "" {
		if err := writeUnsupportedMarkdown(cfg.unsupportedPath, output); err != nil {
			return err
		}
	}
	if cfg.unsupportedReasonsPath != "" {
		if err := writeUnsupportedReasonsMarkdown(cfg.unsupportedReasonsPath, output); err != nil {
			return err
		}
	}
	if cfg.readmePath != "" {
		if err := updateReadmeSupport(cfg.readmePath, output); err != nil {
			return err
		}
	}
	return nil
}

func writeSupportedMarkdown(path string, output report, results []result) error {
	names := make([]string, 0, output.GeneratedCount)
	for _, result := range results {
		if result.exclusion == "" && result.err == nil && len(result.diagnostics) == 0 {
			names = append(names, result.card.Name)
		}
	}
	slices.SortFunc(names, func(a, b string) int {
		if compared := cmp.Compare(strings.ToLower(a), strings.ToLower(b)); compared != 0 {
			return compared
		}
		return cmp.Compare(a, b)
	})

	var builder strings.Builder
	_, _ = builder.WriteString("# Supported Cards\n\n")
	_, _ = builder.WriteString(supportSummary(output))
	_, _ = builder.WriteString("\n\n")
	for _, name := range names {
		_, _ = fmt.Fprintf(&builder, "- %s\n", markdownInline(name))
	}
	return writeDocumentationFile(path, builder.String())
}

func writeUnsupportedMarkdown(path string, output report) error {
	cards := append([]unsupported(nil), output.Unsupported...)
	slices.SortFunc(cards, func(a, b unsupported) int {
		if compared := cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)); compared != 0 {
			return compared
		}
		if compared := cmp.Compare(a.Name, b.Name); compared != 0 {
			return compared
		}
		return cmp.Compare(a.OracleID, b.OracleID)
	})

	var builder strings.Builder
	_, _ = builder.WriteString("# Unsupported Cards\n\n")
	_, _ = builder.WriteString(supportSummary(output))
	_, _ = builder.WriteString("\n\n")
	_, _ = builder.WriteString(
		"These cards are eligible for paper support but cardgen cannot yet generate them. " +
			"Cards excluded by the corpus policy are not listed.\n\n",
	)
	for _, card := range cards {
		reasons := make([]string, 0, len(card.Diagnostics))
		for _, diagnostic := range card.Diagnostics {
			reason := markdownInline(diagnostic.Summary)
			if diagnostic.Detail != "" {
				reason += ": " + markdownInline(diagnostic.Detail)
			}
			reasons = append(reasons, reason)
		}
		_, _ = fmt.Fprintf(
			&builder,
			"- **%s** — %s\n",
			markdownInline(card.Name),
			strings.Join(reasons, "; "),
		)
	}
	return writeDocumentationFile(path, builder.String())
}

func writeUnsupportedReasonsMarkdown(path string, output report) error {
	analysis := analyzeSupport(output)
	var builder strings.Builder
	_, _ = builder.WriteString("# Card-Support Planning Report\n\n")
	_, _ = builder.WriteString(
		"Capability-aware blockers for eligible paper cards that cannot yet be generated. " +
			"Each distinct diagnostic summary and capability is counted at most once per card.\n\n",
	)
	_, _ = builder.WriteString("## Diagnostic reasons\n\n")
	_, _ = builder.WriteString(
		"A sole blocker is the card's only distinct diagnostic summary. " +
			"The most common co-blocker excludes the reason in its own row.\n\n",
	)
	_, _ = builder.WriteString(
		"| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |\n",
	)
	_, _ = builder.WriteString("| ---: | --- | ---: | ---: | ---: | --- |\n")
	for index, reason := range analysis.reasons {
		coBlocker := "-"
		if reason.mostCommonCoBlocker != "" {
			coBlocker = markdownTableCell(reason.mostCommonCoBlocker)
		}
		_, _ = fmt.Fprintf(
			&builder,
			"| %d | %s | %s | %s | %.1f%% | %s |\n",
			index+1,
			markdownTableCell(reason.summary),
			formatCount(reason.affectedCards),
			formatCount(reason.soleBlockerCards),
			reason.soleBlockerPercentage(),
			coBlocker,
		)
	}
	_, _ = builder.WriteString("\n## Capability clusters\n\n")
	_, _ = builder.WriteString(
		"A fully unlockable card has every distinct diagnostic summary in one capability cluster. " +
			"Constituent summaries list the diagnostics currently observed in that cluster.\n\n",
	)
	_, _ = builder.WriteString(
		"| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |\n",
	)
	_, _ = builder.WriteString("| --- | ---: | ---: | --- |\n")
	for _, capability := range analysis.capabilities {
		summaries := make([]string, len(capability.summaries))
		for index, summary := range capability.summaries {
			summaries[index] = markdownTableCell(summary)
		}
		_, _ = fmt.Fprintf(
			&builder,
			"| %s | %s | %s | %s |\n",
			capability.id,
			formatCount(capability.affectedCards),
			formatCount(capability.fullyUnlockableCards),
			strings.Join(summaries, "; "),
		)
	}
	return writeDocumentationFile(path, builder.String())
}

func updateReadmeSupport(path string, output report) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading README support summary: %w", err)
	}
	content := string(data)
	if strings.Count(content, readmeSupportStart) != 1 || strings.Count(content, readmeSupportEnd) != 1 {
		return fmt.Errorf("README %s must contain one %s / %s marker pair", path, readmeSupportStart, readmeSupportEnd)
	}
	start := strings.Index(content, readmeSupportStart)
	end := strings.Index(content, readmeSupportEnd)
	if end < start {
		return fmt.Errorf("README %s must contain an ordered %s / %s marker pair", path, readmeSupportStart, readmeSupportEnd)
	}
	start += len(readmeSupportStart)
	replacement := "\n" + supportSummary(output) +
		" See [`supported.md`](./supported.md) and [`unsupported.md`](./unsupported.md) for the complete lists, " +
		"and [`unsupported-reasons.md`](./unsupported-reasons.md) for capability-aware blocker planning.\n"
	content = content[:start] + replacement + content[end:]
	return writeDocumentationFile(path, content)
}

func supportSummary(output report) string {
	percentage := 0.0
	if output.EligibleCount > 0 {
		percentage = 100 * float64(output.GeneratedCount) / float64(output.EligibleCount)
	}
	return fmt.Sprintf(
		"Council4 currently supports **%s of %s cards eligible for paper support (%.1f%%)**. "+
			"The Scryfall Oracle Cards corpus contains %s additional digital, special-format, memorabilia, "+
			"or non-sanctioned-paper records that are excluded from that total.",
		formatCount(output.GeneratedCount),
		formatCount(output.EligibleCount),
		percentage,
		formatCount(output.ExcludedCount),
	)
}

func formatCount(value int) string {
	text := strconv.Itoa(value)
	for i := len(text) - 3; i > 0; i -= 3 {
		text = text[:i] + "," + text[i:]
	}
	return text
}

func markdownInline(text string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"\n", " ",
		"\r", " ",
		"*", "\\*",
		"_", "\\_",
		"[", "\\[",
		"]", "\\]",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(text)
}

func markdownTableCell(text string) string {
	return strings.ReplaceAll(markdownInline(text), "|", "\\|")
}

func writeDocumentationFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating documentation directory for %s: %w", path, err)
	}
	if existing, err := os.ReadFile(path); err == nil && string(existing) == content {
		return nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("reading existing documentation %s: %w", path, err)
	}
	//nolint:gosec // The caller explicitly selects documentation output paths through CLI flags.
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing documentation %s: %w", path, err)
	}
	return nil
}

func compileCorpus(input io.Reader, workers int) ([]result, error) {
	decoder := json.NewDecoder(input)
	tkn, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("reading bulk-data array: %w", err)
	}
	if delimiter, ok := tkn.(json.Delim); !ok || delimiter != '[' {
		return nil, errors.New("bulk data from Scryfall must be a top-level JSON array")
	}

	jobs := make(chan job)
	results := make(chan result)
	var workersDone sync.WaitGroup
	workersDone.Add(workers)
	for range workers {
		go func() {
			defer workersDone.Done()
			for item := range jobs {
				results <- compileCard(item)
			}
		}()
	}
	go func() {
		workersDone.Wait()
		close(results)
	}()

	decodeError := make(chan error, 1)
	go func() {
		defer close(jobs)
		sent := 0
		for decoder.More() {
			var card cardgen.ScryfallCard
			if err := decoder.Decode(&card); err != nil {
				decodeError <- fmt.Errorf("decoding card %d: %w", sent, err)
				return
			}
			jobs <- job{index: sent, card: card}
			sent++
		}
		if _, err := decoder.Token(); err != nil {
			decodeError <- fmt.Errorf("closing bulk-data array: %w", err)
			return
		}
		decodeError <- nil
	}()

	var all []result
	for compiled := range results {
		all = append(all, compiled)
	}
	if err := <-decodeError; err != nil {
		return nil, err
	}
	disambiguateCollisions(all)
	rejectPathCollisions(all)
	rejectIdentifierCollisions(all)
	slices.SortFunc(all, func(a, b result) int {
		return cmp.Compare(a.index, b.index)
	})
	return all, nil
}

func compileCard(item job) result {
	card := item.card
	compiled := result{index: item.index, card: card}
	if reason, excluded := (cardgen.CorpusPolicy{}).Exclusion(card); excluded {
		compiled.exclusion = reason
		return compiled
	}
	identity, err := cardgen.GeneratedIdentity(&card, false)
	if err != nil {
		compiled.diagnostics = []oracle.Diagnostic{{
			Severity: oracle.SeverityWarning,
			Summary:  "invalid generated identity",
			Detail:   err.Error(),
		}}
		return compiled
	}
	letter := identity.PackageName
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		compiled.diagnostics = []oracle.Diagnostic{{
			Severity: oracle.SeverityWarning,
			Summary:  "unsupported package letter",
			Detail:   fmt.Sprintf("card name %q does not map to an ASCII a-z package", card.Name),
		}}
		return compiled
	}
	compiled.relative = identity.RelativePath
	compiled.superseded = identity.SupersededPath
	compiled.source, compiled.diagnostics, compiled.err =
		(cardgen.ExecutableGenerator{IdentifierSuffix: identity.IdentifierSuffix}).
			GenerateCardSource(&card, identity.PackageName)
	return compiled
}

func disambiguateCollisions(results []result) {
	parent := make([]int, len(results))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(index int) int {
		if parent[index] != index {
			parent[index] = find(parent[index])
		}
		return parent[index]
	}
	union := func(indexes []int) {
		if len(indexes) < 2 {
			return
		}
		root := find(indexes[0])
		for _, index := range indexes[1:] {
			parent[find(index)] = root
		}
	}
	byPath := make(map[string][]int)
	byIdentifier := make(map[string][]int)
	for i := range results {
		if results[i].err != nil || len(results[i].diagnostics) > 0 {
			continue
		}
		byPath[results[i].relative] = append(byPath[results[i].relative], i)
		file, err := parser.ParseFile(token.NewFileSet(), results[i].relative, results[i].source, 0)
		if err != nil {
			results[i].err = fmt.Errorf("parsing generated source: %w", err)
			continue
		}
		for _, name := range cardDefNames(file) {
			key := filepath.Dir(results[i].relative) + "\x00" + name
			byIdentifier[key] = append(byIdentifier[key], i)
		}
	}
	for _, indexes := range byPath {
		union(indexes)
	}
	for _, indexes := range byIdentifier {
		union(indexes)
	}
	components := make(map[int][]int)
	for i := range results {
		if results[i].err == nil && len(results[i].diagnostics) == 0 {
			components[find(i)] = append(components[find(i)], i)
		}
	}
	colliding := make(map[int]bool)
	for _, indexes := range components {
		if len(indexes) < 2 {
			continue
		}
		slices.SortFunc(indexes, func(a, b int) int {
			aKey := cardIdentityKey(&results[a].card)
			bKey := cardIdentityKey(&results[b].card)
			if byIdentity := strings.Compare(aKey, bKey); byIdentity != 0 {
				return byIdentity
			}
			return cmp.Compare(results[a].index, results[b].index)
		})
		for _, index := range indexes[1:] {
			colliding[index] = true
		}
	}
	for index := range colliding {
		card := &results[index].card
		identity, err := cardgen.GeneratedIdentity(card, true)
		if err != nil {
			results[index].source = ""
			results[index].diagnostics = []oracle.Diagnostic{{
				Severity: oracle.SeverityWarning,
				Summary:  "generated identity collision",
				Detail:   err.Error(),
			}}
			continue
		}
		results[index].relative = identity.RelativePath
		results[index].superseded = identity.SupersededPath
		results[index].source, results[index].diagnostics, results[index].err =
			(cardgen.ExecutableGenerator{IdentifierSuffix: identity.IdentifierSuffix}).GenerateCardSource(
				card,
				identity.PackageName,
			)
	}
	finalPaths := make(map[string]bool)
	for i := range results {
		if results[i].err == nil && len(results[i].diagnostics) == 0 {
			finalPaths[results[i].relative] = true
		}
	}
	for i := range results {
		if finalPaths[results[i].superseded] {
			results[i].superseded = ""
		}
	}
}

func cardIdentityKey(card *cardgen.ScryfallCard) string {
	if card.OracleID != "" {
		return card.OracleID
	}
	return card.ID
}

func rejectPathCollisions(results []result) {
	byPath := make(map[string][]int)
	for i := range results {
		if results[i].exclusion == "" && results[i].err == nil && len(results[i].diagnostics) == 0 {
			byPath[results[i].relative] = append(byPath[results[i].relative], i)
		}
	}
	for path, indexes := range byPath {
		if len(indexes) < 2 {
			continue
		}
		for _, index := range indexes {
			results[index].source = ""
			results[index].diagnostics = []oracle.Diagnostic{{
				Severity: oracle.SeverityWarning,
				Summary:  "generated path collision",
				Detail:   fmt.Sprintf("%d Oracle cards map to %s", len(indexes), path),
			}}
		}
	}
}

func rejectIdentifierCollisions(results []result) {
	byName := make(map[string][]int)
	for i := range results {
		if results[i].exclusion != "" || results[i].err != nil || len(results[i].diagnostics) > 0 {
			continue
		}
		file, err := parser.ParseFile(
			token.NewFileSet(),
			results[i].relative,
			results[i].source,
			0,
		)
		if err != nil {
			results[i].err = fmt.Errorf("parsing generated source: %w", err)
			continue
		}
		seen := make(map[string]bool)
		for _, name := range cardDefNames(file) {
			if seen[name] {
				results[i].source = ""
				results[i].diagnostics = []oracle.Diagnostic{{
					Severity: oracle.SeverityWarning,
					Summary:  "duplicate generated identifier",
					Detail:   fmt.Sprintf("generated source declares %s more than once", name),
				}}
				continue
			}
			seen[name] = true
			byName[name] = append(byName[name], i)
		}
	}
	for name, indexes := range byName {
		if len(indexes) < 2 {
			continue
		}
		for _, index := range indexes {
			results[index].source = ""
			results[index].diagnostics = []oracle.Diagnostic{{
				Severity: oracle.SeverityWarning,
				Summary:  "generated identifier collision",
				Detail: fmt.Sprintf(
					"%d Oracle cards declare %s in package %s",
					len(indexes),
					name,
					filepath.Dir(results[index].relative),
				),
			}}
		}
	}
}

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
			diagnostics = []oracle.Diagnostic{{
				Severity: oracle.SeverityError,
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

func reportDiagnostics(diagnostics []oracle.Diagnostic) []reportDiagnostic {
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

func diagnosticSeverityName(severity oracle.Severity) string {
	switch severity {
	case oracle.SeverityError:
		return "error"
	case oracle.SeverityWarning:
		return "warning"
	default:
		return "unknown"
	}
}

func writeSupported(root string, results []result) error {
	affected := make(map[string]bool)
	finalPaths := make(map[string]bool)
	generatedPrefixes := make(map[string]bool)
	tokenPrefixes := make(map[string]bool)
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) > 0 {
			continue
		}
		finalPaths[result.relative] = true
		directory := filepath.Dir(result.relative)
		base := cardgen.CardNameToSafeFileName(result.card.Name)
		if result.card.Layout == "token" || result.card.Layout == "double_faced_token" {
			tokenPrefixes[filepath.Join(directory, base+"_")] = true
		} else {
			generatedPrefixes[filepath.Join(directory, base+"_scryfall")] = true
		}
	}
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) > 0 || result.superseded == "" {
			continue
		}
		path := filepath.Join(root, result.superseded)
		err := os.Remove(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing superseded source for %s: %w", result.card.Name, err)
		}
		if err == nil {
			affected[filepath.Dir(path)] = true
		}
	}
	for prefix := range generatedPrefixes {
		matches, err := filepath.Glob(filepath.Join(root, prefix+"*.go"))
		if err != nil {
			return fmt.Errorf("matching generated identity paths for %s: %w", prefix, err)
		}
		for _, path := range matches {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return fmt.Errorf("resolving generated identity path %s: %w", path, err)
			}
			if finalPaths[relative] {
				continue
			}
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("removing obsolete generated identity path %s: %w", path, err)
			}
			affected[filepath.Dir(path)] = true
		}
	}
	for prefix := range tokenPrefixes {
		matches, err := filepath.Glob(filepath.Join(root, prefix+"*.go"))
		if err != nil {
			return fmt.Errorf("matching generated token identity paths for %s: %w", prefix, err)
		}
		for _, path := range matches {
			if !isTokenIdentityPath(path, filepath.Join(root, prefix)) {
				continue
			}
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return fmt.Errorf("resolving generated token identity path %s: %w", path, err)
			}
			if finalPaths[relative] {
				continue
			}
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("removing obsolete generated token identity path %s: %w", path, err)
			}
			affected[filepath.Dir(path)] = true
		}
	}
	for _, result := range results {
		if result.exclusion != "" || result.err != nil || len(result.diagnostics) > 0 {
			continue
		}
		path := filepath.Join(root, result.relative)
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return fmt.Errorf("creating package directory for %s: %w", result.card.Name, err)
		}
		if err := os.WriteFile(path, []byte(result.source), 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
		affected[filepath.Dir(path)] = true
	}
	directories := make([]string, 0, len(affected))
	for directory := range affected {
		directories = append(directories, directory)
	}
	slices.Sort(directories)
	for _, directory := range directories {
		if err := writeCardList(directory); err != nil {
			return err
		}
	}
	return writeTokenPackages(root, results)
}

func isTokenIdentityPath(path, prefix string) bool {
	suffix := strings.TrimSuffix(strings.TrimPrefix(path, prefix), ".go")
	if len(suffix) != 32 {
		return false
	}
	for _, r := range suffix {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return false
		}
	}
	return true
}

func writeTokenPackages(root string, results []result) error {
	letters := make(map[string]bool)
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) > 0 ||
			(result.card.Layout != "token" && result.card.Layout != "double_faced_token") {
			continue
		}
		letters[filepath.Base(filepath.Dir(result.relative))] = true
	}
	if len(letters) == 0 {
		return nil
	}
	tokenRoot := filepath.Join(root, "tokens")
	if err := os.MkdirAll(tokenRoot, 0o750); err != nil {
		return fmt.Errorf("creating token package: %w", err)
	}
	rootReadme := "# Tokens\n\n" +
		"Package `tokens` collects generated definitions for playable paper tokens. " +
		"Token definitions live in letter subpackages and use their complete Oracle ID " +
		"in filenames and Go identifiers so same-name tokens remain distinct.\n\n" +
		"Tokens are not included in `cards.Registry`. In the repository tree, use " +
		"`tokens.Cards` when all token definitions are needed.\n"
	if err := os.WriteFile(filepath.Join(tokenRoot, "README.md"), []byte(rootReadme), 0o600); err != nil {
		return fmt.Errorf("writing token package README: %w", err)
	}
	ordered := make([]string, 0, len(letters))
	for letter := range letters {
		ordered = append(ordered, letter)
	}
	slices.Sort(ordered)
	for _, letter := range ordered {
		doc := fmt.Sprintf(
			"// Package %s contains generated playable token definitions.\npackage %s\n\n"+
				"//go:generate go run github.com/natefinch/council4/cardgen/cmd/gencardlist\n",
			letter,
			letter,
		)
		docPath := filepath.Join(tokenRoot, letter, "doc.go")
		if err := os.WriteFile(docPath, []byte(doc), 0o600); err != nil {
			return fmt.Errorf("writing token letter package documentation: %w", err)
		}
		readme := fmt.Sprintf(
			"# %s tokens\n\nPackage `%s` contains generated playable token definitions whose names begin with %s. "+
				"Use `Cards` to iterate over every token definition in this package.\n",
			strings.ToUpper(letter), letter, strings.ToUpper(letter),
		)
		path := filepath.Join(tokenRoot, letter, "README.md")
		if err := os.WriteFile(path, []byte(readme), 0o600); err != nil {
			return fmt.Errorf("writing token letter package README: %w", err)
		}
	}
	if !isRepositoryCardsRoot(root) {
		return nil
	}

	var builder strings.Builder
	_, _ = builder.WriteString("// Code generated by compilecards; DO NOT EDIT.\n\n")
	_, _ = builder.WriteString("// Package tokens provides playable token definitions.\n")
	_, _ = builder.WriteString("package tokens\n\n")
	_, _ = builder.WriteString("import (\n\t\"slices\"\n\n")
	for _, letter := range ordered {
		_, _ = fmt.Fprintf(
			&builder,
			"\t\"github.com/natefinch/council4/mtg/cards/tokens/%s\"\n",
			letter,
		)
	}
	_, _ = builder.WriteString(")\n\n")
	_, _ = builder.WriteString("// Cards lists all generated playable token definitions.\n")
	_, _ = builder.WriteString("var Cards = slices.Concat(\n")
	for _, letter := range ordered {
		_, _ = fmt.Fprintf(&builder, "\t%s.Cards,\n", letter)
	}
	_, _ = builder.WriteString(")\n")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("formatting token package: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tokenRoot, "cards.go"), formatted, 0o600); err != nil {
		return fmt.Errorf("writing token package: %w", err)
	}
	return nil
}

func isRepositoryCardsRoot(root string) bool {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	repositoryRoot, err := filepath.Abs(filepath.Join("mtg", "cards"))
	return err == nil && absoluteRoot == repositoryRoot
}

func writeCardList(directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("reading %s: %w", directory, err)
	}
	varNames := make([]string, 0, len(entries))
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "cards.go" || strings.HasSuffix(name, "_test.go") ||
			!strings.HasSuffix(name, ".go") {
			continue
		}
		file, err := parser.ParseFile(files, filepath.Join(directory, name), nil, 0)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", name, err)
		}
		varNames = append(varNames, cardDefNames(file)...)
	}
	slices.Sort(varNames)

	var builder strings.Builder
	_, _ = builder.WriteString("// Code generated by compilecards; DO NOT EDIT.\n\n")
	_, _ = fmt.Fprintf(&builder, "package %s\n\n", filepath.Base(directory))
	_, _ = builder.WriteString("import \"github.com/natefinch/council4/mtg/game\"\n\n")
	_, _ = builder.WriteString("// Cards lists all card definitions in this package.\n")
	_, _ = builder.WriteString("var Cards = []*game.CardDef{\n")
	for _, name := range varNames {
		_, _ = fmt.Fprintf(&builder, "\t%s,\n", name)
	}
	_, _ = builder.WriteString("}\n")
	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("formatting %s/cards.go: %w", directory, err)
	}
	if err := os.WriteFile(filepath.Join(directory, "cards.go"), formatted, 0o600); err != nil {
		return fmt.Errorf("writing %s/cards.go: %w", directory, err)
	}
	return nil
}

func cardDefNames(file *ast.File) []string {
	var names []string
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.VAR {
			continue
		}
		for _, specification := range general.Specs {
			values, ok := specification.(*ast.ValueSpec)
			if !ok || !isCardDef(values) {
				continue
			}
			for _, name := range values.Names {
				if name.Name != "" && unicode.IsUpper(rune(name.Name[0])) {
					names = append(names, name.Name)
				}
			}
		}
	}
	return names
}

func isCardDef(values *ast.ValueSpec) bool {
	for _, value := range values.Values {
		unary, ok := value.(*ast.UnaryExpr)
		if !ok {
			continue
		}
		composite, ok := unary.X.(*ast.CompositeLit)
		if !ok {
			continue
		}
		selector, ok := composite.Type.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		packageName, ok := selector.X.(*ast.Ident)
		if ok && packageName.Name == "game" && selector.Sel.Name == "CardDef" {
			return true
		}
	}
	return false
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
