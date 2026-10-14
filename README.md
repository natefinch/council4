# Council4

Council4 is a Go playtesting engine for Magic: The Gathering Commander decks. The goal is to run many automated 4-player games with AI-controlled agents and produce analytics about deck performance.

## Development direction

New card support uses **authored Go over shared Engine operations**, not expansion
of the Oracle compiler. Preserve existing generated cards and behavioral tests;
keep costs, targets and declarations inspectable, and rules mutation engine-owned.
See [ADR 0012](./docs/adr/0012-authored-go-card-implementations.md) and the ordered
[migration epic](https://github.com/natefinch/council4/issues/3269).

The [program interface](./mtg/cardprogram/README.md) supplies versioned synchronous
Go bodies through the existing Engine. Later migration slices deepen batch
outcomes, linked state and exceptional rules. Keep a card unsupported until its
whole behavior is implemented; do not expand parser/lowering or invent a parallel
runtime to avoid a missing shared capability.

For authorized issue work, the default is implement, test/lint, inspect and
commit directly to main, with one writer and no mandatory PR. Push permission
remains separate. The [repository instructions](./.github/copilot-instructions.md)
and [impl-loop](./.agents/skills/impl-loop/SKILL.md) describe the execution rules.

## Card support

<!-- card-support:start -->
Production registry: **2790 entries, 2790 loadable, 11 behavior-verified authored identities**. Frozen compiler: **18700/33016 eligible Oracle records accepted**. These are separate measurements; name-only legacy entries are not certified exact Oracle/face identities or behavior-verified games.

See [production identity/status inventory](support-inventory.md), [installed and compiler lists](supported.md), and [compiler maintenance diagnostics](unsupported-reasons.md).
<!-- card-support:end -->

The inventory separates installed code, behavior evidence and frozen compiler
coverage. Neither compiler acceptance nor loadability certifies a complete,
exact game. Compiler diagnostic rankings are maintenance evidence, not the
default new-card work queue.

For an explicitly requested legacy corpus refresh, run
`go run github.com/magefile/mage@v1.15.0 cardSupport`. It uses ignored `.cardwork`
space and updates compiler support documentation. `COUNCIL4_ORACLE_CARDS` may
provide the exact bytes identified by the committed pin; mismatches fail.
`cardTree /path/to/cards` emits a compiled tree for another consumer, and
`supportDocs` explicitly refreshes compiler evidence and overlays production
status. `supportInventory` and `cardBacklog` reuse existing same-input compiler
evidence without compilation. None is an every-issue gate.

CI is read-only and proportional on main pushes and optional PRs. It does not
commit support documentation. Shared runtime/card changes check the actual
registry without downloading the corpus; relevant compiler changes also compare
the pinned inventory and build generated packages. See
[pinned preservation](./testdata/card-preservation/README.md) for the validation
matrix, loss acknowledgments and explicit manual refresh.

## Parser coverage

[`parser-coverage.md`](./parser-coverage.md) records legacy Oracle parser coverage.
Regenerate it with `go run github.com/magefile/mage@v1.15.0 parserCoverage` only
for relevant maintenance. Parser completeness is not an acceptance requirement
for an authored Go card.

## Run the CLI

Run the current land-only test game:

```bash
go run ./cmd/council4
```

Useful flags:

```bash
go run ./cmd/council4 -seed 1 -deck-size 8 -verbose -nopass
go run ./cmd/council4 -mode spells -seed 1 -verbose -nopass
go run ./cmd/council4 -mode combat -seed 1 -verbose -nopass
```

- `-mode` chooses `land`, `spells`, or `combat`.
- `-seed` controls deterministic shuffling.
- `-deck-size` controls how many basic Forests each land-mode test deck contains.
- `-verbose` prints the per-turn draw, action, resolve, combat damage, creature damage, death, and loss log.
- `-nopass` omits pass actions from verbose log output.

Example output:

```text
Council4 test game
Seed: 1
Mode: land
Deck size: 8 Forests per player
Turns: 7
Winner: Player 4
Battlefield permanents: 4
```

## Build and test

```bash
go test ./...
go run github.com/magefile/mage@v1.15.0 lint
```

Run focused package tests while editing code, then the suite/linter once before
committing. Add relevant build, race or platform checks when needed. Docs and
comment-only changes use link/instruction checks and `git diff --check`.

The module root is the repository root:

```text
github.com/natefinch/council4
```

## Repository layout

```text
.
├── cmd/
│   └── council4/          # CLI entry point
├── mtg/
│   ├── game/              # Core game state data types
│   │   ├── action/        # Player action data
│   │   ├── counter/       # Counter types and counter sets
│   │   ├── id/            # Unique object identifiers
│   │   └── mana/          # Mana colors, symbols, costs, pools
│   ├── rules/             # Rules engine and minimal game loop
│   └── agent/             # AI agent implementations
├── docs/
│   ├── adr/               # Architecture decision records
│   └── research/          # MTG and AI reference documents
└── CONTEXT.md             # Shared project vocabulary
```

Every Go package directory has its own `README.md` describing that package and how to use it.

## Engine coverage

Use [`docs/playtest-rule-coverage.md`](./docs/playtest-rule-coverage.md) and the
owning package documentation for current behavior and limitations. The
[`ROADMAP.md`](./ROADMAP.md) preserves earlier implementation milestones; its
historical compiler-expansion checklists are not the migration execution queue.

## Documentation

- [`CONTEXT.md`](./CONTEXT.md) defines project vocabulary.
- [`ADR 0012`](./docs/adr/0012-authored-go-card-implementations.md) defines the authored-Go direction and direct-to-main workflow.
- [`docs/playtest-rule-coverage.md`](./docs/playtest-rule-coverage.md) summarizes supported rule coverage and known limitations for the playtest path.
- [`docs/adr/`](./docs/adr/) records architectural decisions.
- Package-level READMEs document implementation details and usage for each package.
