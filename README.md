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

For authorized issue work, the default is implement, test, inspect and
commit directly to main, with one writer and no mandatory PR. Push permission
remains separate. The [repository instructions](./.github/copilot-instructions.md)
and [impl-loop](./.agents/skills/impl-loop/SKILL.md) describe the execution rules.

## Card support

<!-- card-support:start -->
Production registry: **17973 entries, 17973 loadable, 50 behavior-verified authored identities**. Frozen compiler: **18700/33016 eligible Oracle records accepted**. These are separate measurements; name-only legacy entries are not certified exact Oracle/face identities or behavior-verified games.

See [production identity/status inventory](support-inventory.md), [installed and compiler lists](supported.md), and [compiler maintenance diagnostics](unsupported-reasons.md).
<!-- card-support:end -->

The inventory separates installed code, behavior evidence and frozen compiler
coverage. Neither compiler acceptance nor loadability certifies a complete,
exact game. Compiler diagnostic rankings are maintenance evidence, not the
default new-card work queue.

`supportInventory` is an explicit reporting-maintenance command. It reads
existing compiler evidence without compilation; do not run it before or after
ordinary implementation work. Corpus recompilation, bulk imports and fingerprint
refresh/baseline workflows are retired. Existing cards and behavioral tests stay.

## Parser coverage

[`parser-coverage.md`](./parser-coverage.md) preserves historical Oracle parser
coverage. It is not a production queue or an authored-card acceptance requirement.

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

## Development checks

Follow the [validation policy](./.github/copilot-instructions.md#verify-proportionally)
for local checks, CI and review. Do not duplicate its gates in individual guides.

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
