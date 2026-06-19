# Engine performance log

This directory tracks the ongoing effort to make full-game simulation fast
enough for large playtests, and records the game-time / allocation impact of
each change.

## Baseline decks

`cmd/council4/testdata/perf/deck1.txt` … `deck4.txt` are four legal four-player
Commander decks built entirely from currently supported cards (the default
registry). They are generated to be legal (singleton nonbasics within the
commander's color identity, basics to fill out 99 cards) and validated with
`rules.ValidateCommanderConfigs`.

Commanders:

| Deck  | Commander                     | Identity |
|-------|-------------------------------|----------|
| deck1 | Halana and Alena, Partners    | R/G      |
| deck2 | Neyith of the Dire Hunt       | R/G      |
| deck3 | Bugenhagen, Wise Elder        | G        |
| deck4 | Hazoret, Godseeker            | R        |

Caveat: the supported-card pool is small (~59 cards), so singleton legality
forces ~53 basic lands per deck. Combined with the pool's several ramp spells,
these decks build very large boards and play long games — an intentionally heavy
allocation/effective-value stress test, heavier than a tuned real deck would be.

## How to measure

A repeatable benchmark loads the committed decks and plays one full game:

```
go test ./cmd/council4/ -run '^$' -bench BenchmarkPerfDeckGame -benchtime=1x -benchmem -timeout 600s
```

`BenchmarkPerfDeckGameFirstLegal` is deterministic (seed 20260619, FirstLegal
agent) and is the headline number tracked below. Because absolute wall time
varies with machine load, **allocations/op is the stable metric** to compare
across changes. (`BenchmarkPerfDeckGameGeneric` runs the rule-based agent; it is
even heavier and used for spot checks.)

## Results

### `BodyAt` boxing (the change in #786)

`CardFace.BodyAt` returned each ability as an `Ability` interface; with value
receivers that boxed a heap copy of the (large) ability struct on every call.
`mtg/game` `BenchmarkBodyAt` reads all abilities of an ability-dense face:

| Code | ns/op | B/op | allocs/op |
|------|-------|------|-----------|
| Baseline (merge-base) | (noisy) | 24,402 | 5 |
| After #786 (pointer receivers) | ~77 | **0** | **0** |

The per-call boxing — five large ability structs, ~24 KB — is eliminated
entirely. This benefits every effective-value computation, and most of all
ability-dense boards (a commander or anthem that is a static source queried
repeatedly), where ability boxing dominated allocation in earlier profiles.

### Full baseline game

`BenchmarkPerfDeckGameFirstLegal`, one full game (seed 20260619). These decks
are deliberately land-heavy (singleton legality over a ~59-card pool forces ~53
basics), so most permanents carry a single ability and the board is not
ability-dense; the per-call `BodyAt` win therefore moves the whole-game total
only modestly. Allocations/op is the comparison metric — wall time on the shared
build box is too noisy to compare run-to-run.

| Change | allocs/op | B/op | notes |
|--------|-----------|------|-------|
| Baseline (before #786) | ~2.37 billion | ~881 GB | merge-base |
| After #786 | ~2.33 billion | ~647 GB | ~25% fewer bytes; land-heavy board limits the alloc-count delta |

The byte allocation dropped ~25% (the boxed ability structs are large), while the
allocation *count* moved less because each basic land boxes only one small mana
ability. On ability-dense boards the count delta is far larger (see the
per-call benchmark above and the Bugenhagen profile in #717's discussion).

Earlier work for context (#717): the static-source frame cache made
effective-value computation ~14× faster and realistic micro full games ~2×
faster, and it already limits how often `BodyAt` is called (once per permanent
per read frame). See PR #724.
