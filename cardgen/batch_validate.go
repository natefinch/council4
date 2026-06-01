package cardgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type manifestValidationRunResult struct {
	Found  []string          `json:"found"`
	Issues []ValidationIssue `json:"issues"`
}

// MissingWorklist returns generated-card names that should be attempted with
// the card-impl workflow. It includes missing files and invalid generated cards,
// skips fetch errors, and respects limit when limit is positive.
func MissingWorklist(manifest Manifest, limit int) []string {
	var names []string
	for _, card := range manifest.Cards {
		if card.Status == BatchStatusFetchError {
			continue
		}
		if card.FileStatus != BatchFileStatusMissing && card.Validation != BatchValidationStatusInvalid {
			continue
		}
		name := manifestCardName(card)
		if name == "" {
			continue
		}
		names = append(names, name)
		if limit > 0 && len(names) == limit {
			return names
		}
	}
	return names
}

// ValidateManifestGeneratedCards validates manifest rows that have existing
// generated source files by importing their card packages through a temporary Go
// program. It updates each row's Issues and Validation fields.
func ValidateManifestGeneratedCards(manifest *Manifest, repoRoot string) error {
	wanted := map[string]bool{}
	for i := range manifest.Cards {
		card := &manifest.Cards[i]
		if card.FileStatus != BatchFileStatusExisting {
			continue
		}
		card.Issues = nil
		card.Validation = BatchValidationStatusUnvalidated
		name := manifestCardName(*card)
		if name != "" {
			wanted[name] = true
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	result, err := runGeneratedCardValidation(repoRoot, wanted)
	if err != nil {
		return err
	}
	found := map[string]bool{}
	for _, name := range result.Found {
		found[name] = true
	}
	issuesByCard := map[string][]ValidationIssue{}
	for _, issue := range result.Issues {
		issuesByCard[issue.CardName] = append(issuesByCard[issue.CardName], issue)
	}
	for i := range manifest.Cards {
		card := &manifest.Cards[i]
		if card.FileStatus != BatchFileStatusExisting {
			continue
		}
		name := manifestCardName(*card)
		card.Issues = append(card.Issues, issuesByCard[name]...)
		if !found[name] {
			card.Issues = append(card.Issues, ValidationIssue{
				CardName: name,
				Code:     IssueGeneratedCardNotFound,
				Message:  "expected generated card file exists, but no matching CardDef was found in the package Cards slice",
			})
		}
		if len(card.Issues) == 0 {
			card.Validation = BatchValidationStatusValid
		} else {
			card.Validation = BatchValidationStatusInvalid
		}
	}
	return nil
}

// RunGoGenerateCards runs the card registry go:generate directives.
func RunGoGenerateCards(repoRoot string) error {
	cmd := exec.Command("go", "generate", "./mtg/cards/...")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go generate ./mtg/cards/... failed: %w\n%s", err, string(output))
	}
	return nil
}

func runGeneratedCardValidation(repoRoot string, wanted map[string]bool) (manifestValidationRunResult, error) {
	result, err := runValidationProgram(repoRoot, validationProgram(wanted))
	if err == nil {
		return result, nil
	}
	var combined manifestValidationRunResult
	for _, letter := range validationLetters(wanted) {
		letterWanted := map[string]bool{}
		for name := range wanted {
			if CardNameToPackageLetter(name) == letter {
				letterWanted[name] = true
			}
		}
		letterResult, letterErr := runValidationProgram(repoRoot, validationProgram(letterWanted))
		if letterErr != nil {
			for _, name := range sortedWantedNames(letterWanted) {
				combined.Issues = append(combined.Issues, ValidationIssue{
					CardName: name,
					Code:     IssueValidationRunFailed,
					Message:  letterErr.Error(),
				})
			}
			continue
		}
		combined.Found = append(combined.Found, letterResult.Found...)
		combined.Issues = append(combined.Issues, letterResult.Issues...)
	}
	return combined, nil
}

func runValidationProgram(repoRoot string, program string) (manifestValidationRunResult, error) {
	tmpDir := filepath.Join(repoRoot, ".cardwork", "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return manifestValidationRunResult{}, err
	}
	file, err := os.CreateTemp(tmpDir, "cardbatch-validate-*.go")
	if err != nil {
		return manifestValidationRunResult{}, err
	}
	defer os.Remove(file.Name())
	if _, err := file.WriteString(program); err != nil {
		file.Close()
		return manifestValidationRunResult{}, err
	}
	if err := file.Close(); err != nil {
		return manifestValidationRunResult{}, err
	}
	cmd := exec.Command("go", "run", file.Name())
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return manifestValidationRunResult{}, fmt.Errorf("generated card validation failed: %w\n%s", err, string(output))
	}
	var result manifestValidationRunResult
	decoder := json.NewDecoder(bytes.NewReader(output))
	if err := decoder.Decode(&result); err != nil {
		return manifestValidationRunResult{}, fmt.Errorf("decoding generated card validation output: %w\n%s", err, string(output))
	}
	return result, nil
}

func validationProgram(wanted map[string]bool) string {
	letters := validationLetters(wanted)
	var b strings.Builder
	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"os\"\n\n")
	b.WriteString("\t\"github.com/natefinch/council4/cardgen\"\n")
	b.WriteString("\t\"github.com/natefinch/council4/mtg/game\"\n")
	for i, letter := range letters {
		b.WriteString(fmt.Sprintf("\tp%d \"github.com/natefinch/council4/mtg/cards/%s\"\n", i, letter))
	}
	b.WriteString(")\n\n")
	b.WriteString("type result struct { Found []string `json:\"found\"`; Issues []cardgen.ValidationIssue `json:\"issues\"` }\n\n")
	b.WriteString("func main() {\n")
	b.WriteString("\twanted := map[string]bool{\n")
	for _, name := range sortedWantedNames(wanted) {
		encoded, _ := json.Marshal(name)
		b.WriteString(fmt.Sprintf("\t\t%s: true,\n", encoded))
	}
	b.WriteString("\t}\n")
	b.WriteString("\tvar cards []*game.CardDef\n")
	for i := range letters {
		b.WriteString(fmt.Sprintf("\tfor _, card := range p%d.Cards { if wanted[card.Name] { cards = append(cards, card) } }\n", i))
	}
	b.WriteString("\tres := result{Issues: cardgen.ValidateCards(cards, cardgen.ValidationOptions{ReportImplementationIDs: true})}\n")
	b.WriteString("\tfor _, card := range cards { res.Found = append(res.Found, card.Name) }\n")
	b.WriteString("\tif err := json.NewEncoder(os.Stdout).Encode(res); err != nil { panic(err) }\n")
	b.WriteString("}\n")
	return b.String()
}

func validationLetters(wanted map[string]bool) []string {
	seen := map[string]bool{}
	for name := range wanted {
		letter := CardNameToPackageLetter(name)
		if letter != "" {
			seen[letter] = true
		}
	}
	letters := make([]string, 0, len(seen))
	for letter := range seen {
		letters = append(letters, letter)
	}
	sort.Strings(letters)
	return letters
}

func sortedWantedNames(wanted map[string]bool) []string {
	names := make([]string, 0, len(wanted))
	for name := range wanted {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func manifestCardName(card ManifestCard) string {
	if card.CanonicalName != "" {
		return card.CanonicalName
	}
	return card.InputName
}
