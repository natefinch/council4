# parsercoverage

`parsercoverage` measures how completely the Oracle parser
(`cardgen/oracle/parser`) represents the eligible Scryfall corpus as typed
syntax, **without** running the compiler or lowering. It is a parser-stage
progress metric: the score rises as parser grammar improves, and the report
ranks the grammar the parser cannot yet represent so that work can be
prioritized.

It stream-decodes a Scryfall Oracle Cards bulk-data array, applies the same
corpus-eligibility policy as [`compilecards`](../compilecards/README.md), parses
each eligible card's executable faces (`cardgen.ParseCardFaces`, which stops
after `parser.Parse`), and classifies coverage with
`parser.DocumentCoverage`.

## The metrics

Two distinct metrics are reported; they must not be conflated.

**Parser-complete (typed coverage).** An ability is **parser-complete** when
every must-cover token span (`parser.Ability.CoverageSpans`) is accounted for by
a recognized span built **only** from the parser's typed output — a typed effect
clause, a recognized trigger clause, a recognized cost, recognized condition
segments, recognized static declarations, keywords, semantic references,
reminders, the ability-word clause, chapter headings, and the additional-cost
declaration — and every condition introducer it owns resolves to a recognized
clause. For modal abilities the choice header and every mode must be recognized.
A card is parser-complete when every ability of every executable face is
parser-complete. Typed coverage only requires a *kind-recognized* element
(`Kind != EffectUnknown`); it does **not** require byte-exact reconstruction, so
it is an upper bound on what the lowerer could consume.

**Exact round-trip.** Strictly stronger: the parser reconstructs the original
text byte-for-byte. An effect is exact when `Exact` (or `Mana.LegacyBodyExact`
for add-mana bodies) is set and its whole sentence is represented. A card is
exact round-trip when it is parser-complete **and** every resolving effect is
exact.

This mirrors the spans the lowering coverage consumers assert against, but is
reconstructed purely from parser data. Because lowering is downstream of
parsing, every card the lowerer can fully generate **must** be parser-complete;
`parsercoverage` checks that invariant directly (see `-generated`).

## Reported numbers

- **Card-level parser-complete % (typed coverage)** — parser-complete cards /
  eligible cards. The upper bound on lowerable cards.
- **Card-level exact round-trip %** — cards that are parser-complete and whose
  every resolving effect is exact / eligible cards.
- **Effect-level exact round-trip %** — exact effects / all resolving effects.
- **Uncovered grammar work queue** — the unrepresented grammar, clustered by the
  owning component's normalized text with example card names, ranked by
  frequency. Each cluster is attributed to a blocker family (effect, trigger,
  cost, condition, modal).

## Flags

- `-in` — Scryfall Oracle Cards bulk-data JSON file (required).
- `-report` — JSON report path, or `-` for stdout (default `-`).
- `-out` — parser-coverage Markdown path (default `parser-coverage.md`).
- `-generated` — optional supported-card Markdown (`supported.md`). When set,
  the tool asserts that every generated card name is parser-complete and reports
  any violations. A violation usually means the recognized-span union is too
  strict and is a bug to fix by adding the missing recognized span, not a metric
  to loosen. A small residue can legitimately remain when the parser recognizes
  a construct semantically but exposes no source span covering all its tokens
  (coordinated trigger/condition lists, "for each" iteration prefixes,
  reflexive/delayed trigger preambles); those cards are listed by name rather
  than hidden by over-crediting.
- `-workers` — number of parser workers (0 selects `NumCPU`).

## Usage

```bash
go run ./cardgen/oracle/cmd/parsercoverage \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out parser-coverage.md \
  -report .cardwork/parser-coverage-report.json \
  -generated supported.md
```

Or regenerate the committed report with `mage parserCoverage`.
