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
   - If the card is in a new letter directory, verify the `doc.go` was created automatically

## Important rules

- Use only `EffectType`, `Keyword`, and other enum values that exist in the codebase. Do not invent new ones.
- If a card has effects that cannot be expressed with the existing effect primitives, set `ImplementationID` to a descriptive name and leave a comment explaining what hand-written code would need to do.
- Keep the oracle text comment block at the top of the file — it's useful for human review.
- Run `gofmt` on the file after editing.
- Use the CARD-IMPLEMENTATION-GUIDE.md as your primary reference, not general MTG knowledge.

</what-to-do>
