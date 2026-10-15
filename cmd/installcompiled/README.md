# Install a compiled corpus

This is an explicit, additive legacy-corpus import, not the new-card authoring
workflow. It copies the pinned compiler output into the checked-in card tree,
preserving existing implementations and alternate filenames with the same Go
constructor. It never writes into `authored/` or deletes existing card sources.

From the repository root:

```sh
go run ./cardgen/oracle/cmd/compilecards \
  -in .cardwork/corpus/oracle-cards.json \
  -out .cardwork/compiled-import \
  -report .cardwork/compiled-import-report.json \
  -inventory .cardwork/compiled-import-inventory.json
go run ./cmd/installcompiled
go generate ./mtg/cards/...
go run ./cmd/cardqueue -operation reconcile-compiled
COUNCIL4_COMPILER_REPORT=.cardwork/compiled-import-report.json \
  go run github.com/magefile/mage@v1.15.0 supportInventory
```

Use a fresh compiler output directory. The compiler's `-inventory` verifies the
input bytes against `testdata/card-preservation/corpus.json`; the installer
requires the same pin, accepted source digests, and complete source coverage.
The resulting `installed-compiler.json` records every accepted Oracle identity,
its actual installed source path, constructor and source digest, including
preserved implementations. Repeating the installation is idempotent.

The default registry deliberately prefers the explicit authored replacements.
`cards.CompilerExclusions()` keeps proven runtime blockers out of deck lookup
without deleting their generated sources. New overlaps or exclusions require
explicit review; the installer does not mark a card behavior-verified.

Run the normal test/lint gate, source coverage tests, registry preservation and
affected platform checks. This operation does not commit or push. Compilation
and structural loading do not certify gameplay fidelity.
