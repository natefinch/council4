# cardbacklog

`cardbacklog` turns the parser-coverage data into an actionable, **routed** task
backlog by joining it with the lowering/compile signal. For every eligible
Scryfall corpus card it computes two independent signals and routes the card to
the layer that actually blocks it, emitting two ranked task queues plus a
headline that partitions the corpus.

It is the join of [`parsercoverage`](../parsercoverage/README.md) (the
parser-only signal) and [`compilecards`](../compilecards/README.md) (the
full-compile lowering signal): the former says whether the grammar is
recognized, the latter says whether the recognized grammar can be lowered.

## The two signals

For each eligible card (the same `cardgen.CorpusPolicy` exclusion the other tools
apply is applied first, so excluded cards are dropped before routing):

- **Parser signal** (parser-only). `cardgen.ParseCardFaces` parses every
  executable face; `parser.DocumentCoverage` classifies each face. The card is
  **parser-complete** when every face is parser-complete. The owning component
  and normalized cluster of each uncovered span are collected for the parser
  queue. This is exactly the signal `parsercoverage` reports.
- **Lowering signal** (full compile). Generated membership is read from
  **compilecards' canonical JSON report** (`-compile-report`), not decided by
  cardbacklog alone. compilecards is the single source of truth because it runs
  corpus-wide collision and parse-rejection passes (`rejectPathCollisions`,
  `rejectIdentifierCollisions`, `disambiguateCollisions`) that demote some cards a
  per-card view considers clean — a card can compile in isolation yet be
  unsupported once the whole corpus is generated together. A card is **generated**
  iff its Scryfall `id` is absent from the report's `unsupported` and `excluded`
  sets; an unsupported card carries that report entry's distinct blocking
  diagnostic `Summary` strings. Routing on per-record `id` (not name) avoids the
  ambiguity that `supported.md` has: dozens of names recur across multiple
  `oracle_id`s (tokens and reprints), so name-membership would mis-route cards
  that share a name with a generated card.

### Reconciliation guard

cardbacklog keeps an **independent** per-card recompile
(`cardgen.GeneratedIdentity` + `ExecutableGenerator.GenerateCardSource`) purely as
a cross-check. That recompile deliberately omits the corpus-wide collision passes,
so it can only ever consider *more* cards generated than the authoritative report.
After routing, the tool reconciles the two: it asserts the per-card generated
count equals the report's count and lists, by name and `id`, every card where they
disagree. Any divergence — for example a collision-demoted card the recompile
still thinks is clean — is routed **out** of supported into the lowering queue
under the synthetic reason `generated collision or parse rejection`, and fails the
run with a non-zero exit. The check is non-tautological: unlike the
supported+lowering+parser=eligible partition (which the routing switch makes
trivially true), this guard compares two genuinely independent pipelines and will
fire loudly if they ever drift.

## Routing and the partition

Every eligible (non-excluded) card lands in exactly one bucket:

- **Supported** — the card is generated. (Lowering already works.)
- **Lowering backlog** — the card is parser-complete but not generated. Parsing
  is already done, so this is the lowest-risk backlog: lower the recognized
  grammar.
- **Parser backlog** — the card is not parser-complete and not generated. The
  grammar must be recognized before lowering can even be attempted.

These three buckets are a strict partition of the eligible cards:
`supported + lowering-backlog + parser-backlog = eligible`. The tool asserts this
and prints it to stderr and the report.

A small residue of cards are **generated but not parser-complete**: the lowerer
fully generates them, but the parser-coverage harness does not expose a source
span covering all their must-cover tokens (the residue documented in
`parser-coverage.md`). These stay in the **supported** bucket — they are not
blocked work — and are reported separately by name rather than double-counted
into a queue.

## The two queues

1. **Lowering queue.** The lowering-backlog cards (parser-complete, not
   generated), bucketed by each distinct lowering diagnostic summary and ranked
   by affected-card count, with the sole-blocker count (cards whose only distinct
   blocker is that summary) and example cards. This is `unsupported-reasons.md`
   restricted to the parser-complete subset — the same reasons, with smaller
   counts, because parser-recognition reasons (e.g. `unsupported Oracle
   construct`) cannot appear for parser-complete cards.
2. **Parser queue.** The parser-backlog cards (not parser-complete, not
   generated), bucketed by (owning component family × normalized uncovered-span
   cluster) and ranked by occurrence, with example cards. It mirrors the
   `parser-coverage.md` work queue. Cluster normalization is shared with
   `parsercoverage` via
   [`cmd/internal/cluster`](../internal/cluster).

Each queue row lists a few example card names so it maps directly onto a
capability-family child issue with objective before/after metrics.

## Flags

- `-in` — Scryfall Oracle Cards bulk-data JSON file (required).
- `-compile-report` — compilecards JSON report providing the authoritative
  generated/unsupported set (**required**; run `compilecards` first). The tool
  refuses to run without it rather than fall back to an unsound self-contained
  mode.
- `-out` — card-backlog Markdown path (default `card-backlog.md`).
- `-report` — JSON report path, or `-` for stdout (default `-`).
- `-workers` — number of workers (0 selects `NumCPU`).

## Usage

```bash
# 1. Produce the authoritative compile report.
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/card-backlog-generated \
  -report .cardwork/card-backlog-compile-report.json

# 2. Join it with the parser signal and route the backlog.
go run ./cardgen/oracle/cmd/cardbacklog \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out card-backlog.md \
  -report .cardwork/card-backlog-report.json \
  -compile-report .cardwork/card-backlog-compile-report.json
```

`mage cardBacklog` runs both steps in order and regenerates the committed report.
