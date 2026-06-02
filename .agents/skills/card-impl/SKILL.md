---
name: card-impl
description: >
  Generate a council4 CardDef from a Magic card name. Fetches card data from
  Scryfall, generates the mechanical Go source, then parses the oracle text
  to fill in Abilities. Use when the user says "implement card X",
  "add card X", "generate card definition for X", or similar.
---

<what-to-do>

## Workflow

Given one or more Magic: The Gathering card names:

1. **Run the cardgen script** for each card to produce the mechanical CardDef scaffold:
   ```bash
   go run .agents/skills/card-impl/main.go "Card Name"
   ```
   This fetches from Scryfall and writes a `.go` file under `mtg/cards/<letter>/`.

2. **Read the generated file.** It has the mechanical fields filled in and `Abilities` left empty.

3. **Read the CARD-IMPLEMENTATION-GUIDE.md** in this skill directory. It contains:
   - The full Go type definitions for `AbilityDef`, `Effect`, `Keyword`, etc.
   - Mapping rules from oracle text patterns to struct fields.
   - Worked examples of real cards.
   - Current generated-card conventions: `mtg/game/types` for
     supertypes/card types/subtypes, `mtg/game/compare` for integer
     predicates, and optional `CardDef.Back` for double-faced back faces.

4. **Parse the oracle text** (shown in the comment block at the top of the generated file) and fill in:
   - `Abilities` — the `[]game.AbilityDef` slice
   - `EntersTapped` — if the oracle text says "enters tapped"
   - `EntersWithCounters` — if the oracle text says "enters with N counters"
   - Any other fields derivable from oracle text

5. **Present the completed CardDef** for human review. Explain your reasoning for each ability you parsed.

6. **Run the finishing steps:**
   - Run `gofmt` on the file: `gofmt -w <file>`
   - Run `go generate ./mtg/cards/...` to update the card list
   - Run `go build ./mtg/cards/...` to verify compilation
   - Run `mage lint` and fix every reported issue before considering the code complete
   - If the card is in a new letter directory, verify the `doc.go` was created automatically

## Important rules

- Use only `EffectType`, `Keyword`, and other enum values that exist in the codebase. Do not invent new ones.
- Use `types.Creature`/`types.Forest`/etc. from `mtg/game/types`; do not use old `game.Type*` or `game.*Subtype*` names.
- `mtg/game/types` includes named constants for every Comprehensive Rules 205.3 subtype. Prefer those constants for new card definitions instead of `types.Sub("...")`; fall back to `types.Sub` only if the subtype truly is not present.
- For multiple plain non-parameterized keywords in one oracle line, add one reusable helper ability per keyword (for example `game.DeathtouchAbility, game.IndestructibleAbility`) instead of combining them into one `AbilityDef`.
- For double-faced cards, edit front-face data on `CardDef` and back-face data on `Back: opt.Val(game.CardFace{...})`; do not add a `Faces` slice.
- If a card has effects that cannot be expressed with the existing effect primitives, set `ImplementationID` to a descriptive name and leave a comment explaining what hand-written code would need to do.
- Keep the oracle text comment block at the top of the file — it's useful for human review.
- Run `gofmt` on the file after editing.
- Do not call implementation work complete until `mage lint` passes.
- Use the CARD-IMPLEMENTATION-GUIDE.md as your primary reference, not general MTG knowledge.

</what-to-do>
