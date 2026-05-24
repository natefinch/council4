# action package

`mtg/game/action` defines the data shape for player decisions. Actions are produced by the rules engine, evaluated by agents, and applied back to the game by the rules engine.

This package intentionally contains data types only. It does not decide whether an action is legal and does not mutate game state.

## When to use it

Use this package anywhere code needs to represent a player choice:

- `rules.Engine` returns legal actions to an agent.
- An agent chooses one action from the legal set.
- Game logs record which action a player took.
- Future replay/debug tooling serializes action sequences.

## Design

`Action` is a tagged struct:

```go
act := action.PlayLand(cardID)
switch act.Kind {
case action.ActionPlayLand:
	// use act.PlayLand.CardID
case action.ActionPass:
	// pass priority
}
```

The top-level `Kind` field says which payload is meaningful. Payloads are grouped by kind (`PlayLandAction`, `CastSpellAction`, etc.) so action data stays explicit without using an interface hierarchy.

## Current action kinds

- `ActionPass` - pass priority or decline to take an available action.
- `ActionPlayLand` - play a land card from hand.
- `ActionCastSpell` - cast a spell with chosen runtime `game.Target` values.
- `ActionActivateAbility` - activate an ability from a permanent or other source with chosen runtime `game.Target` values.
- `ActionDeclareAttackers` - declare attackers during combat.
- `ActionDeclareBlockers` - declare blockers during combat.

## Package boundaries

`action` imports `mtg/game` for shared domain types such as `AttackDeclaration`, `BlockDeclaration`, and runtime `Target` values. The dependency intentionally points from action data to core game data; `mtg/game` must not import `mtg/game/action`.

The rules engine validates action legality. Agents should normally return one of the legal actions they were given, but the engine still treats returned actions as untrusted input.

Cast-spell actions identify the card to cast and carry chosen targets, modes, and X value. Early spell support only generates untargeted, non-X casts; targeting and modal choices are layered in later.
