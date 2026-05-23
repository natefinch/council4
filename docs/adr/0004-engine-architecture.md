# Engine Architecture

The rules engine lives in `mtg/rules/` as an `Engine` struct that owns the full game loop. The `game/` package remains pure data types; behavior lives in `rules/`. This separation keeps the type definitions stable while allowing the rules engine to evolve freely.

## Package Layout

```
council4/                          ← module root (go.mod)
├── mtg/                           ← grouping directory, no Go code
│   ├── game/                      ← core data types (Game, Player, Permanent, etc.)
│   │   ├── id/                    ← unique object identifiers
│   │   ├── mana/                  ← mana colors, costs, pools
│   │   ├── counter/               ← counter types
│   │   └── action/                ← Action type + payload structs (imports game/)
│   ├── rules/                     ← rules engine (Engine struct, game loop, SBAs, effect execution)
│   ├── agent/                     ← agent implementations (imports rules/ for PlayerAgent interface)
│   ├── sim/                       ← game runner, parallel tournament
│   ├── cards/                     ← card registry, pure data (no behavior)
│   └── deck/                      ← decklist parser
├── report/                        ← analytics + output
├── cmd/council4/                  ← CLI
```

Dependency direction: `agent/` and `sim/` depend on `rules/` and `game/`; `cmd/council4/` depends on everything. `game/` depends on nothing outside its leaf sub-packages.

## Key Decisions

**Action type** — a tagged struct with a `Kind` discriminator and kind-specific nested payload structs (e.g., `PlayLandAction`, `CastSpellAction`). Lives in `game/action/`, which imports `game/` for shared types like `PlayerID` and `AttackDeclaration`. Chosen over a flat struct (too many invalid states) and an interface (adds ceremony without eliminating switch logic).

**Engine struct** — `rules.Engine` holds configuration (seeded `*rand.Rand`, future variant rules). Exposes `RunGame(g *game.Game, agents [4]PlayerAgent) *GameResult` which runs a complete game, calling agents via the `PlayerAgent` interface when decisions are needed. The sim layer just calls `RunGame` in parallel.

**PlayerAgent interface** — defined in `rules/`, implemented by `agent/`. Follows Go convention of defining interfaces where they're consumed.

**PlayerObservation** — a purpose-built struct in `rules/` providing a fog-of-war filtered view, not a copy of `*game.Game`. Agents never see the full game state.

**GameResult** — a structured result with per-turn logs, living in `rules/`. Contains enough data for `report/` to compute all metrics (win rate, per-card performance, mana analysis, tempo).

**State-based actions** — unexported functions inside `rules/`, checked before granting priority, looping until stable.

**Effects** — rich declarative `Effect` struct in `game/` carries parameters (amount, targets, duration). The rules engine in `rules/` switches on `EffectType` to execute them. Cards too complex for declarative effects use hand-written resolver functions registered in `rules/`.

**Card data** — `cards.Registry` maps card names to `*game.CardDef` values. It is pure data with no behavior. The `deck/` package uses the registry to resolve decklists. Custom card resolvers for the escape hatch live in `rules/`.

**Randomness** — a seeded `*rand.Rand` is injected into the `Engine`. Each game in a simulation gets its own `*rand.Rand` derived from a master seed for reproducibility and goroutine safety.

## Implementation Order

1. Module restructure (move `go.mod` to repo root, `game/` → `mtg/game/`)
2. Minimal game loop (untap → draw → play land → pass turn)
3. Spells + stack (casting, priority passing, resolution)
4. Combat (attackers, blockers, damage)
