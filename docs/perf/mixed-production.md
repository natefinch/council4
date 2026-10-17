# Mixed production acceptance (#3278)

**Historical acceptance and performance evidence.** Measurements and passed
checks below describe that checkpoint, not current validation requirements.
Follow the [repository policy](../../.github/copilot-instructions.md#verify-proportionally)
for new work and [replay documentation](../../mtg/sim/README.md#replay) for the
current ID/version compatibility contract; source-hash bindings are retired.

## What was exercised

All eleven proving identities are installed in the default registry: Terastodon,
Hive Mind, Gifts Ungiven, Gilded Drake, Worldgorger Dragon, Nim Deathmantle,
Abundance, Blood Moon, Muldrotha, Skullbriar and Mindslaver. The frozen compiler
is not a prerequisite to their admission.

`mtg/rules/mixed_production_test.go` casts every proving card through production
legal actions, the live Engine and a determinized search Simulator. The three
interaction fixtures below execute both legacy instructions and authored Go.
The retained named-card tests additionally execute Abundance's draw replacement,
Blood Moon's type/ability layers, Muldrotha's graveyard permissions, Skullbriar's
combat/counter retention, Mindslaver's activated turn control, and Nim
Deathmantle's paid return/attachment. These are rules/consumer tests, not merely
authored-constructor assertions.

`mtg/deck/production_load_test.go` reloads serialized mixed decklists into a fresh
production registry and preserves exact authored evidence, name-only legacy
evidence and explicit Commander legality diagnostics. Small scenario decks are
fixtures, not claims of legal 100-card Commander construction. The committed
performance decks below pass Commander deck validation.

`mtg/sim/production_test.go` runs mixed decks with one and two workers, records
and JSON-reloads replay scripts, and compares complete results. At this checkpoint
replay bound authored executable source and shared rules, including static-only
Blood Moon decks and transitive Go-only scheduled handlers. Changing a
source/rules binding failed explicitly. Program declarations remain data, not
serialized functions.
The existing missing-handler tests reject loading/preflight and retain
`unsupported-execution`, with no completed win or draw.

The same mixed Engine, Simulator, replay, inventory and evaluation scenarios
execute under Node using Go's actual `js/wasm` runner, not just a cross-build.
WASM requires `GOMAXPROCS=1`; native benchmark concurrency is recorded separately.
The initial attempt with native `GOMAXPROCS=4` failed before tests with
`newosproc: not implemented`; the single-thread runtime run passed.

## Outcomes and fidelity

Rules wins/draws (and single-player rules losses) are distinct from turn-budget,
decision-work-budget, unsupported and invalid execution. All remaining players
losing simultaneously draw; previously eliminated players do not inherit that
draw. Matchup work exhaustion stays a budget stop even when accompanied by a
failure record. Incomplete slots are excluded from completed-result rates and
length statistics, never subtracted from a misleading "draw" total.

Invalid actions, invalid choices, deterministic fallbacks and hypothetical-search
budget stops are inspectable counters. Search work starts independently of live
decisions. Neither these counters nor the invariant observation limitations
reveal hidden card identities. Authored effect intent is descriptive; unknown
heuristic effects are marked unknown rather than certified zero. Search still
executes actual handlers.

The default live decision cap is 100,000; hypothetical branches use 10,000.
Turn caps and Agent candidate width/lookahead are separate policy budgets.
Declared legal-action restrictions remain: X/multikicker enumeration/acceptance
through 20 ([#3284](https://github.com/natefinch/council4/issues/3284)) and
representative combat declarations/order
([#3126](https://github.com/natefinch/council4/issues/3126)). Evaluation knowledge
is incomplete. These conservative declarations prevent `GameResult.Exact()`
from claiming exact gameplay, even when a rules winner is produced. This slice
does not silently drop restrictions or claim to solve those blockers.

## Fixed-input measurements

Apple M3 Max, macOS arm64, Go 1.26.4, `GOMAXPROCS=4`, **one simulation worker**.
Numbers are medians of three samples, not confidence intervals. Full games use
`-benchtime=1x`; mixed interactions use `-benchtime=10x`. Allocation bytes are
cumulative, not peak memory.

The step-02 planning baseline is
`69156a33a219a9ad9fb9edce779d2ff0fe5a9964`. The preserved copied tree at
`.cardwork/step12-baseline02` was checked against that exact Git tree: all
non-test production Go, module files and perf inputs match; only the three
benchmark files add decision/termination reporting. It is **not** a different
step-02 implementation or a claimed comparable snapshot inferred from its name.
Step-11 baseline is `e5c532169d7152749222aaed51f4b84aa40b611b`, with identical
production source and benchmark-only instrumentation. Its mixed fixture adds
the same setup, seed and policy to a revision where those authored cards exist.
Step 02 predates them, so no invented step-02 mixed-card number is reported.

The unchanged committed inputs are `cmd/council4/testdata/perf/deck{1,2,3,4}.txt`:

| File | SHA256 |
| --- | --- |
| deck1 | `9d28c18c78a2b254b822e4dfffa7b03538682a578f1530be64edb8032e5637c8` |
| deck2 | `d82dde2db20560629f66bf36bd433c7813e67d30d160a783d7c7086626577cc2` |
| deck3 | `5b55e918653293fc43c28dc941f82a88b97f99054ae20695b46ab4e7cf78f9ed` |
| deck4 | `9a4b4675f5122e6c62a9a30fe4a3a44aa9881798de6890cc84ee6bb50644b4f7` |

Full perf decks use seed `20260619` and FirstLegal or Generic as labeled.
Forest/flying-haste fixtures use `NewEngine(nil)`'s fixed PCG `(1,2)` with
FirstLegal; the smoke fixture uses seed `20240101` and its existing FirstLegal
factory. No turn/decision/search budget was shortened to improve these numbers.

| Workload | Step-02 ns/op | Current ns/op | Step-02 B/op | Current B/op | Step-02 allocs/op | Current allocs/op | Logged decisions, old → current | Turns / termination |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |
| Forest fixture | 241,818,416 | 292,385,291 | 164,799,640 | 430,822,376 | 1,812,541 | 1,927,483 | 1,636 → 1,811 | 48 / rules win |
| Flying/haste fixture | 892,854,250 | 1,114,451,583 | 443,028,480 | 1,340,766,640 | 5,035,989 | 5,282,355 | 2,850 → 3,120 | 88 / rules win |
| Simulation smoke | 41,063,125 | 48,424,208 | 24,842,552 | 44,496,904 | 222,876 | 243,895 | 639 → 724 | 23 / rules win |
| Committed decks / FirstLegal | 6,863,966,667 | 7,886,957,875 | 4,201,967,816 | 5,136,226,432 | 25,620,809 | 26,657,921 | 2,358 → 2,572 | 54 / rules win |
| Committed decks / Generic | 4,402,101,042 | 4,825,728,625 | 4,237,874,280 | 5,026,448,376 | 16,222,253 | 16,731,242 | 2,122 → 2,304 | 47 / rules win |

The original planning single samples (~237.5 ms Forest, 872.2 ms flying/haste,
42.1 ms smoke, ~3.99 s Generic) remain historical evidence, not replacements
for the same-machine three-sample comparison above. Turns/winners remain stable,
but the migration adds real engine-mediated choices; the action/choice traces
are not falsely described as identical across all fourteen steps.

Mixed seed `20261014`, PCG `(seed, seed^7)`, pass-priority policy with deterministic
engine-valid choices. Stable priority is **scenario completion**, not a rules
draw or a completed multiplayer game.

| Interaction | Step-11 ns/op | Current ns/op | Step-11 B/op | Current B/op | allocs/op (both) | Decisions | Termination |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| Hive Mind + Lightning Bolt | 554,912 | 553,654 | 1,110,947 | 1,110,952 | 3,666 | 27 | stable priority |
| Terastodon + Basilisk Collar / Forest | 1,152,017 | 819,804 | 1,153,276 | 1,153,299 | 5,080 | 14 | stable priority |
| Worldgorger Dragon + Beast Within / Collar | 1,397,067 | 973,500 | 3,138,646 | 3,138,635 | 5,786 | 24 | stable priority |

## Regression investigation

Before this slice's narrow fix, step-11 committed decks took 25.027 s FirstLegal
and 14.102 s Generic: respectively ~3.65×/~3.20× the step-02 remeasurement.
A fresh current Generic CPU profile reproduced 14.163 s and attributed
**52.3% flat sampled CPU** to `hasBasicLandTypeSet`, which copied large static
bodies/effect structs during repeated negative scans. Iterating those immutable
bodies/effects by pointer, without changing applicability, ordering, framing
or rules, removed that copying. An immediate same-input profile fell to 5.151 s.
Final unprofiled medians above are 7.887/4.826 s, still **14.9%/9.6% slower than
step 02**, but **68.5%/65.8% faster than step 11**. The Blood Moon interaction,
dependency, clone and frame tests remain the correctness gate; this is not a
new search algorithm or a relaxed rules workload.

Allocation regressions are not hidden: current Forest/flying/smoke allocate
2.61×/3.03×/1.79× baseline bytes; full decks allocate 1.22×/1.19×. A separate
step-11 synthetic remeasurement gave 430,822,216 / 1,340,766,304 / 44,496,584 B/op,
essentially the same as current: these large costs predate the new diagnostics.
The current Forest allocation profile attributes 51.5% of sampled bytes to
`CardFace.clone`, with `ToCardDef` and effective-value construction also prominent.
The Generic profile shows continuous/static-rule effect materialization and
face/ability cloning dominating cumulative allocations. Bounded measured
diagnostics did not explain or eliminate that inherited representation cost.
Further clone/materialization optimization is separate work; no arbitrary-game
throughput target, zero regression claim or complete-fidelity claim is made.

## Reproduction and preservation

Raw local evidence (ignored, retained): `.cardwork/step12-baseline02-bench.log`,
`step12-baseline11-bench.log`, `step12-baseline11-synthetic.log`,
`step12-final-bench.log`, `step12-final-mixed-bench.log`, source-provenance JSON
and the current/after CPU and allocation profiles. The mixed table uses the last
standalone mixed run; earlier noisy runs are retained, not overwritten to create
a better baseline. Re-run:

```sh
GOMAXPROCS=4 go test ./cmd/council4 -run '^$' -bench '^BenchmarkPerfDeckGame(FirstLegal|Generic)$' -benchtime=1x -count=3 -benchmem -timeout=600s
GOMAXPROCS=4 go test ./mtg/rules -run '^$' -bench '^Benchmark(RealisticGame|RealForestGame)$' -benchtime=1x -count=3 -benchmem
GOMAXPROCS=4 go test ./mtg/rules -run '^$' -bench '^BenchmarkMixedProduction$' -benchtime=10x -count=3 -benchmem
GOMAXPROCS=4 go test ./mtg/sim -run '^$' -bench '^BenchmarkRunOneGame$' -benchtime=1x -count=3 -benchmem
```

Support evidence uses the pinned Oracle JSON SHA256
`6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4`,
compiler-report SHA256
`f8f1f1447e8b56484b40ab91f86bd5d3e427b71991a03b3926e64972c1b78601`,
and Comprehensive Rules **2026-09-25**. `supportInventory` reuses this existing
compiler report; it does not rerun or expand the legacy compiler. Compiler
acceptance (18,700/33,016), loadable production entries (2,790) and named authored
behavior evidence (11 identities) remain separate, with no inferred exact Oracle
identities for name-only legacy entries.

At this checkpoint `TestValidationMatrix` selected registry preservation and
code checks for authored edits and compiler-maintenance checks for support
tooling. That matrix and its corpus/fingerprint gates are retired.
This slice's recorded local final validation used full tests followed by lint,
focused race tests, actual WASM scenarios/builds, registry preservation and diff
inspection. For current validation policy, follow the
[repository instructions](../../.github/copilot-instructions.md#verify-proportionally).
