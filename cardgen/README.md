# cardgen

Package `cardgen` fetches card data from the Scryfall API and generates partial `game.CardDef` Go source files for the council4 card registry.

## What it does

Given a Magic: The Gathering card name, the library:

1. Fetches the card's data from Scryfall's `/cards/named` API endpoint.
2. Parses the mechanical fields: name, mana cost, mana value, colors, color identity, types, subtypes, supertypes, power/toughness, loyalty, defense, simple ETB tapped text, and double-faced card faces.
3. Generates a Go source file with a `game.CardDef` literal, leaving the `Abilities` slice empty for LLM completion.

## Usage

The library is typically used via the `card-impl` skill's `main.go`:

```bash
go run .agents/skills/card-impl/main.go "Lightning Bolt"
```

This creates `mtg/cards/l/lightning_bolt.go` with the mechanical fields populated.

## Double-faced layouts

`cardgen` accepts Scryfall `transform`, `modal_dfc`, `double_faced_token`, `meld`, and `reversible_card` layouts. Transform, modal DFC, and double-faced token cards emit `Layout` plus per-face `[]game.CardFace` data. Meld cards generate the front card with `LayoutMeld`; full melded-permanent behavior is rules/card-implementation work. Reversible cards generate separate `CardDef` variables for each playable side rather than a face-selectable card.

## Key functions

- `FetchCard(name string)` — fetches a card from Scryfall by exact name.
- `GenerateCardSource(card, pkgName)` — generates Go source for a `CardDef`.
- `ParseManaCostLiteral(cost)` — converts Scryfall mana cost strings to Go code.
- `ManaValueFromCost(cost)` — computes a face's mana value from a mana-cost string.
- `ParseTypeLine(typeLine)` — splits a type line into supertypes, types, subtypes.
- `CardNameToVarName(name)` — converts card names to Go exported variable names.
- `CardNameToFileName(name)` — converts card names to snake_case file names.
- `CardNameToPackageLetter(name)` — returns the first letter for sub-package routing.
