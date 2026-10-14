# Pinned card preservation

`corpus.json` identifies exact compressed and converted Scryfall bytes, their
publication date, the rules edition and the existing code baseline. Its dated
download URL is used directly: no current bulk-metadata lookup or daily cache key.
If the artifact disappears, the check fails explicitly; an archived copy with the
same verified bytes may be supplied through `COUNCIL4_ORACLE_CARDS`.

`baseline.json` records two independent inventories on that input:

- **Compiler:** 33,016 eligible records, 18,700 accepted. Stable Oracle identities,
  layout/face names, Oracle-text hashes, generated-definition hashes and exact
  refusals are preserved.
- **Registry:** 2,779 deck entries and 591 separate token definitions, all
  constructor/structurally loadable at capture. This is not gameplay certification.
  No registered definition requires a custom `ImplementationID` at this baseline.
  Legacy registration stores names rather than Oracle IDs: `oracle_ids` are
  explicitly candidate matches, not guessed identity proof. Entry keys preserve
  registration occurrences. The authored catalog will supply exact identities.

Compiler fingerprints hash generated Go; registry fingerprints hash every
structural CardDef value, including private sealed payloads, without pointer
addresses or renderer assumptions. They do **not** hash Engine behavior or Go
handler implementations. Existing rules/clone/choice scenarios remain the
behavioral regression net; no tests or generated definitions were replaced.

Existing focused scenarios retained by the runtime gates include
`mtg/rules/lki_test.go`, `replacement_property_test.go`,
`continuous_property_test.go`, and `mtg/game/clone_test.go`. Run relevant
packages directly during development; shared changes select the full suite in
CI. The preservation hashes are not substitutes for those scenarios.

## Validation matrix

| Change | CI |
| --- | --- |
| Docs/instructions only | Stable successful job names; no game/lint/corpus work |
| Isolated card definition | Short tests, lint, build/vet, short race tests, registry inventory |
| Shared Engine/model/Agent/consumer | Full tests including slow games, lint/build/vet/race, registry inventory |
| Legacy compiler/tooling or CI | Full tests, lint/build/vet/race, pinned compiler and registry delta, generated-package build/vet |
| Explicit corpus-only upgrade | Verify new pin/baseline; distinguish dataset changes from code comparisons |

`go run ./cmd/ciplan -base <base-sha> -head HEAD` emits the actual path-based plan.
Empty/zero base handles an initial push conservatively. Unknown nondocumentation
paths still select code checks. CI-plan failures propagate into the required
jobs rather than producing successful-looking skipped checks.

## Commands

```bash
# Read-only, no corpus download:
go run github.com/magefile/mage@v1.15.0 preserveRegistry

# Read-only pinned compiler comparison, plus generated build/vet:
go run github.com/magefile/mage@v1.15.0 preserveCards

# Explicit candidate baseline replacement; never commits or pushes:
go run github.com/magefile/mage@v1.15.0 refreshPreservation
```

Scratch reports and generated sources live under `.cardwork/preservation/`.
CI compares with the **base commit's** inventory, so removing a baseline entry
in the same code change cannot hide a loss. Initial adoption bootstraps the
committed inventory. A corpus-only upgrade changes both pin and baseline without
code/workflow changes; otherwise mismatched inputs fail and require a separate,
explicit dataset update. Baseline-only rewrites do not receive that exception.

`exceptions.json` acknowledges an intentional correctness-related loss using
the exact inventory/key, pinned corpus digest, previous definition fingerprint,
reason and repository issue URL. Obtain approval for the correction, document
its rules scenario and inspect the complete delta. Do not insert a waiver merely
to make a failing gate green. Admissions, changed definitions and diagnostic
changes are reported; unacknowledged accepted/loadable losses fail.

The manual **Explicit pinned corpus refresh** workflow reproduces the pin and
exports artifacts. It has read-only permissions and no credential persistence,
commit or push step. The retained `cardSupport`, `supportDocs`, `cardBacklog`,
`compileCheck` and `parserCoverage` commands also use the verified pin, but remain
explicit maintenance operations. Parser-subset checking uses the legacy compiler
inventory, not the mixed authored-card support list.
