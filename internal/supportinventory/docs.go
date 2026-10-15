package supportinventory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteDocs refreshes inventory consumers without recompiling the legacy corpus.
// Compiler diagnostic details and the maintenance backlog remain preserved.
func (inv Inventory) WriteDocs(root string) error {
	var supported, unsupported, inventory strings.Builder
	authored, loadable := 0, 0
	for _, support := range inv.Production {
		if support.Loadable {
			loadable++
		}
		if support.Implementation == "authored" && support.BehaviorVerified {
			authored++
		}
	}
	summary := fmt.Sprintf("Production registry: **%d entries, %d loadable, %d behavior-verified authored identities**. Frozen compiler: **%d/%d eligible Oracle records accepted**. These are separate measurements; name-only legacy entries are not certified exact Oracle/face identities or behavior-verified games.", len(inv.Production), loadable, authored, inv.Accepted, inv.Eligible)
	binding := fmt.Sprintf("\n\nPinned corpus SHA256 `%s`; compiler evidence SHA256 `%s`; Comprehensive Rules 2026-09-25. Reproduce with `mage supportInventory` (no corpus compilation).\n\n", inv.CorpusSHA256, inv.CompilerSHA256)
	_, _ = fmt.Fprintf(&supported, "# Supported Cards\n\n%s%s## Installed production registrations\n\nLegacy entries are name-only and loadable, not individually behavior-certified. Exact authored evidence appears in [support-inventory.md](support-inventory.md).\n\n", summary, binding)
	for _, entry := range inv.Production {
		label := "legacy name-only"
		if entry.Implementation == "authored" {
			label = "authored exact Oracle/front; behavior not verified"
			if entry.BehaviorVerified {
				label = "authored exact Oracle/front; behavior verified"
			}
		}
		if !entry.Loadable {
			label = "not loadable: " + entry.Failure
		}
		_, _ = fmt.Fprintf(&supported, "- %s — %s\n", escape(entry.Name), label)
	}
	_, _ = fmt.Fprintf(&unsupported, "# Cards without exact authored production support\n\n%s%sCompiler rejection is maintenance evidence, not a gameplay verdict. The list below excludes working exact authored identities; legacy candidates (if any) retain only name-based evidence. Accepted compiler records without exact production identity are also not certified gameplay.\n\n", summary, binding)
	_, _ = fmt.Fprintf(&inventory, "# Production support inventory\n\n%s%s## Exact authored Oracle/front admissions\n\nSource revisions bind authored Go (all handlers and dependency declarations), not just CardDefs. Shared-rules revisions bind non-test Go in `mtg/{rules,game,cardprogram}` and `opt`. `go generate ./mtg/cards/authored` updates them; a native test rejects stale bindings. Digests detect changes; they do not prove behavior.\n\n| Card | Oracle identity | Compiler accepted | Loadable | Behavior evidence | Program versions |\n| --- | --- | --- | --- | --- | --- |\n", summary, binding)
	for _, record := range inv.Corpus {
		if record.Production != nil {
			s := record.Production
			_, _ = fmt.Fprintf(&inventory, "| %s | `%s/front` | %t | %t | %s | `%v` |\n", escape(s.Name), s.OracleID, record.CompilerAccepted, s.Loadable, s.Verification, s.Programs)
		}
		if record.Blocker != "" {
			_, _ = fmt.Fprintf(&unsupported, "- **%s** (`%s`) — generated source checked in; excluded from playable registry pending [runtime blocker](%s)\n",
				escape(record.Name), record.OracleID, record.Blocker)
			continue
		}
		if record.CompilerAccepted || record.Production != nil && record.Production.Loadable {
			continue
		}
		_, _ = fmt.Fprintf(&unsupported, "- **%s** (`%s`) — compiler rejected; legacy candidate entries: %d; exact production identity not established\n", escape(record.Name), record.OracleID, len(record.LegacyCandidates))
	}
	_, _ = fmt.Fprint(&supported, "\n## Compiler-accepted Oracle records (not gameplay certification)\n\nAll 18,700 accepted source definitions are checked in: 17,946 cards and 754 separate tokens. See [installation coverage](testdata/card-preservation/installed-compiler.json). Authored overrides remain preferred; proven runtime blocks stay excluded from deck lookup. Source presence and name-based loading do not establish behavior verification.\n\n")
	for _, record := range inv.Corpus {
		if record.CompilerAccepted {
			label := "compiler accepted; checked-in definition"
			if record.Blocker != "" {
				label += "; playable admission blocked: " + record.Blocker
			}
			_, _ = fmt.Fprintf(&supported, "- %s (`%s`) — %s\n", escape(record.Name), record.OracleID, label)
		}
	}
	for _, support := range inv.Production {
		if support.Implementation == "authored" {
			_, _ = fmt.Fprintf(&inventory, "\n- %s: authored `%s`; shared rules `%s`.\n", escape(support.Name), support.SourceRevision, support.RulesRevision)
		}
	}
	_, _ = fmt.Fprint(&inventory, "\n## Fidelity and consumers\n\n`cards.Registry.Inventory/SupportFor` is authoritative for installed entries; deck loading returns the same evidence. JSON inventory output joins corpus IDs without guessing legacy identity. Engine, Simulator, CLI and WASM share the installed catalog. A deliberately absent handler fails checked loading or produces `unsupported-execution`, never a rules draw.\n\nRules wins/draws, turn budgets, work budgets, unsupported and invalid execution are distinct. Action/choice fallbacks and invalid answers are counted separately. `GameResult.Exact()` rejects declared restrictions and policy fallbacks. Limits include representative combat declarations ([#3126](https://github.com/natefinch/council4/issues/3126)) and incomplete evaluation knowledge. Exact X/multikicker announcements are validated independently of bounded numeric policy samples ([#3284](https://github.com/natefinch/council4/issues/3284)). Search action width, determinization count, lookahead and deep width are policy budgets, not rules. Unknown authored/effect valuation is explicit; search executes real programs, not metadata.\n\nSee [mixed production validation](docs/perf/mixed-production.md) for executed scenarios and measurements. CI runs code/scenario checks and registry preservation for authored edits, not frozen-corpus regeneration. Legacy compiler diagnostics remain in `unsupported-reasons.md` and `card-backlog.md`; they are not the automatic authoring roadmap.\n")
	for name, data := range map[string]string{"supported.md": supported.String(), "unsupported.md": unsupported.String(), "support-inventory.md": inventory.String()} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0o600); err != nil {
			return err
		}
	}
	for _, name := range []string{"card-backlog.md", "unsupported-reasons.md"} {
		path := filepath.Join(root, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		const start = "<!-- production-inventory:start -->"
		const end = "<!-- production-inventory:end -->"
		note := start + "\n**Legacy compiler maintenance only, not the authored-card roadmap.** " + summary +
			"\nWorking authored identities are excluded from production unsupported status. See [support-inventory.md](support-inventory.md); diagnostic counts and historical routing below remain frozen compiler evidence, not loadability or behavioral claims.\n" + end
		if at := strings.Index(text, start); at >= 0 {
			until := strings.Index(text[at:], end)
			if until < 0 {
				return fmt.Errorf("unclosed inventory marker in %s", path)
			}
			text = text[:at] + note + text[at+until+len(end):]
		} else if at := strings.Index(text, "\n\n"); at >= 0 {
			text = text[:at+2] + note + "\n\n" + text[at+2:]
		}
		//nolint:gosec // Caller-selected repository root; every output filename is fixed, never derived from card data.
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			return err
		}
	}
	readmePath := filepath.Join(root, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	text := string(data)
	start, end := strings.Index(text, "<!-- card-support:start -->"), strings.Index(text, "<!-- card-support:end -->")
	if start < 0 || end < start {
		return errors.New("README support markers missing")
	}
	text = text[:start] + "<!-- card-support:start -->\n" + summary + "\n\nSee [production identity/status inventory](support-inventory.md), [installed and compiler lists](supported.md), and [compiler maintenance diagnostics](unsupported-reasons.md).\n" + text[end:]
	//nolint:gosec // Caller-selected repository root and fixed README filename, not a path derived from card data.
	return os.WriteFile(readmePath, []byte(text), 0o600)
}

func escape(name string) string {
	return strings.NewReplacer("\\", "\\\\", "_", "\\_", "*", "\\*", "|", "\\|", "<", "&lt;", ">", "&gt;").Replace(name)
}
