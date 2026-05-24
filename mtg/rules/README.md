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

`GameResult` is the structured output from a completed game. It records the winner, elimination order, loss reasons, turn count, and per-turn draw/loss/action/resolve logs. The `report` package will consume `[]GameResult` to produce deck analytics.

## Current implementation status

Implemented now:

- `Engine` skeleton and deterministic RNG configuration.
- `Engine.NewGame` for deterministic game setup using the engine RNG.
- `PlayerAgent`, `PlayerObservation`, and result/log data types.
- Opening hand setup and card drawing.
- Phase helpers for beginning, main, combat placeholder, ending, cleanup, and advancing to the next turn.
- Extra turn handling in LIFO order, skipping eliminated players.
- Priority loop with multiplayer pass-around-table behavior and stack-aware all-pass handling.
- State-based actions for player elimination from 0 life, lethal poison, lethal commander damage, and failed draws.
- Legal action generation for passing and playing lands.
- Action application for passing and playing lands.
- Basic mana cost payment helpers that can auto-tap untapped basic lands for colored and generic costs.
- Simple stack resolution for creature spells entering the battlefield and instant/sorcery spells moving to graveyard.
- Effect primitive execution for drawing cards, gaining life, losing life, and player damage.
- Player-targeted spell action generation using `TargetSpec` and runtime `game.Target` values.

Not implemented yet:

- Explicit mana ability actions, permanent targeting, and combat resolution.
- Mulligans and maximum hand-size discard.

## Legal actions

The current engine generates these actions:

- `action.PlayLand(cardID)` for lands in the active player's hand during a main phase when the stack is empty and the land drop is available.
- `action.CastSpell(cardID, targets, xValue, modes)` for supported creature, instant, and sorcery spells. Current cast support covers non-X mana costs, simple player targets, and untargeted spells.
- `action.Pass()` for every player with priority.

Legal actions are ordered as play land, cast spell, then pass so simple agents develop mana before spending it and choose productive actions before passing.

The priority loop treats agent output as untrusted: if an agent returns an action not present in the legal action list, the engine substitutes `Pass`.

When all active players pass in succession, the loop ends the current phase or step only if the stack is empty. If the stack has an object, the engine resolves the top object, resets the pass count, returns priority to the active player, and continues.

## Mana payment

The first mana-payment layer supports normal colored and generic costs. `canPayCost` and `payCost` use current mana pools first, then greedily tap untapped basic lands controlled by the player. Basic land mana is inferred from the land's name or subtype: Plains for white, Island for blue, Swamp for black, Mountain for red, and Forest for green.

Mana pools empty at phase and step boundaries before later priority windows can use stale mana.

## State-based actions

`applyStateBasedActions` loops until stable and panics if state-based actions do not converge. Current checks eliminate players for:

- Life total 0 or less.
- 10 or more poison counters.
- 21 or more commander damage from one commander.
- A failed draw from an empty library (`game.Game.FailedDraws`).

## Package boundaries

`rules` may import `mtg/game` and `mtg/game/action`. It should keep engine internals unexported unless another package genuinely needs them.

The `game` package must remain pure data and should not import `rules`.
