# rules package

`mtg/rules` contains the Magic rules engine. It owns turn progression, priority, state-based actions, effect execution, and the game loop that asks agents to choose actions.

The package operates on the core data types from `mtg/game`. The `game` package stores state; `rules` changes that state according to Magic rules.

## Main types

### Engine

`Engine` is the entry point for rules execution:

```go
rng := rand.New(rand.NewPCG(1, 2))
engine := rules.NewEngine(rng)
gameState := engine.NewGame(configs)
result := engine.RunGame(gameState, agents)
```

The engine receives a `*rand.Rand` so simulations and tests can be deterministic. Passing `nil` uses a deterministic default seed.

Use `Engine.NewGame` when you want the engine's RNG to control both initial library shuffles and later in-game randomness.

`RunGame` currently supports the minimal game loop: opening hands, turn progression, drawing, passing priority, playing lands, state-based player elimination, and game termination.

### PlayerAgent

`PlayerAgent` is the interface the engine consumes when it needs a player decision:

```go
type PlayerAgent interface {
	ChooseAction(obs PlayerObservation, legal []action.Action) action.Action
}
```

The interface lives here because `rules.Engine` consumes it. Concrete agents live in `mtg/agent` later.

### PlayerObservation

`PlayerObservation` is the fog-of-war-safe view passed to an agent. It starts minimal and should grow only as agents need more information.

Do not pass `*game.Game` directly to agents; agents should not see hidden information such as opponents' hands or library order.

### GameResult

`GameResult` is the structured output from a completed game. It records the winner, elimination order, loss reasons, turn count, and per-turn draw/loss/action logs. The `report` package will consume `[]GameResult` to produce deck analytics.

## Current implementation status

Implemented now:

- `Engine` skeleton and deterministic RNG configuration.
- `Engine.NewGame` for deterministic game setup using the engine RNG.
- `PlayerAgent`, `PlayerObservation`, and result/log data types.
- Opening hand setup and card drawing.
- Phase helpers for beginning, main, combat placeholder, ending, cleanup, and advancing to the next turn.
- Extra turn handling in LIFO order, skipping eliminated players.
- Priority loop with multiplayer pass-around-table behavior.
- State-based actions for player elimination from 0 life, lethal poison, lethal commander damage, and failed draws.
- Legal action generation for passing and playing lands.
- Action application for passing and playing lands.

Not implemented yet:

- Spells, stack, mana abilities, and combat resolution.
- Mulligans and maximum hand-size discard.

## Minimal legal actions

The current engine only generates these actions:

- `action.PlayLand(cardID)` for lands in the active player's hand during a main phase when the stack is empty and the land drop is available.
- `action.Pass()` for every player with priority.

Play-land actions are returned before pass so simple agents that choose the first legal action will make progress before passing.

The priority loop treats agent output as untrusted: if an agent returns an action not present in the legal action list, the engine substitutes `Pass`.

## State-based actions

`applyStateBasedActions` loops until stable and panics if state-based actions do not converge. Current checks eliminate players for:

- Life total 0 or less.
- 10 or more poison counters.
- 21 or more commander damage from one commander.
- A failed draw from an empty library (`game.Game.FailedDraws`).

## Package boundaries

`rules` may import `mtg/game` and `mtg/game/action`. It should keep engine internals unexported unless another package genuinely needs them.

The `game` package must remain pure data and should not import `rules`.
