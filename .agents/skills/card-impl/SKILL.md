---
name: card-impl
description: >
  Generate a council4 CardDef from a Magic card name. Fetches card data from
  Scryfall, generates the mechanical Go source, then parses the oracle text
  into categorized CardFace ability fields. Use when the user says "implement card X",
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

2. **Read the generated file.** It has the mechanical fields filled in; the categorized ability fields are left empty for you to complete. Also read `mtg/cards/k/karplusan_forest.go` — this is the canonical source-formatting reference for all new card definitions.

3. **Read the CARD-IMPLEMENTATION-GUIDE.md** in this skill directory. It contains:
   - The full Go type definitions for `AbilityDef`, `Effect`, `Keyword`, etc.
   - Mapping rules from oracle text patterns to struct fields.
   - Worked examples of real cards.
   - Current generated-card conventions: `mtg/game/types` for
     supertypes/card types/subtypes, `mtg/game/compare` for integer
     predicates, and optional `CardDef.Back` for double-faced back faces.

4. **Parse the oracle text** (shown in the comment block at the top of the generated file) and fill in the categorized ability fields:
   - `ManaAbilities`, `ActivatedAbilities`, `TriggeredAbilities`, `StaticAbilities`, `LoyaltyAbilities`, `ReplacementAbilities`, or `SpellAbility` as appropriate — see CARD-IMPLEMENTATION-GUIDE.md
   - `EntersTapped` — if the oracle text says "enters tapped"
   - `EntersWithCounters` — if the oracle text says "enters with N counters"
   - Any other fields derivable from oracle text
   Do **not** populate the legacy `Abilities []AbilityDef` slice.

5. **Present the completed CardDef** for human review. Explain your reasoning for each ability you parsed.

6. **Run the finishing steps:**
   - Run `gofmt` on the file: `gofmt -w <file>`
   - Run `go generate ./mtg/cards/...` to update the card list
   - Run `go build ./mtg/cards/...` to verify compilation
   - Run `mage lint` and fix every reported issue before considering the code complete
   - If the card is in a new letter directory, verify the `doc.go` was created automatically

## Important rules

- Use only existing typed `game.Primitive` variants, `Keyword` values, and other enums. Do not author `game.Effect` or invent new primitives in a card file.
- Use `types.Creature`/`types.Forest`/etc. from `mtg/game/types`; do not use old `game.Type*` or `game.*Subtype*` names.
- `mtg/game/types` includes named constants for every Comprehensive Rules 205.3 subtype. Prefer those constants for new card definitions instead of `types.Sub("...")`; fall back to `types.Sub` only if the subtype truly is not present.
- For multiple plain non-parameterized keywords, append one reusable `StaticAbilityBody` template per keyword to `StaticAbilities` (for example `game.DeathtouchStaticBody, game.IndestructibleStaticBody`). Do not use old `game.DeathtouchAbility`-style `AbilityDef` constructors in new card source.
- Write card source in the canonical expanded style shown in `mtg/cards/k/karplusan_forest.go`: vertically-expanded `CardDef` and `CardFace` literals; `ColorIdentity` before `CardFace` in the struct; `OracleText` and every ability `Text` field using indented raw multiline string literals (opening backtick on its own line, one oracle paragraph per source line, closing backtick indented on its own line); categorized ability slices and bodies expanded one-brace-level-per-line with no compact `{{` forms. The generator produces this layout — preserve it rather than compacting.
- For double-faced cards, edit front-face data on `CardDef` and back-face data on `Back: opt.Val(game.CardFace{...})`; do not add a `Faces` slice.
- If a card has effects that cannot be expressed with the existing effect primitives, set `ImplementationID` to a descriptive name and leave a comment explaining what hand-written code would need to do.
- Keep the oracle text comment block at the top of the file — it's useful for human review.
- Run `gofmt` on the file after editing.
- Do not call implementation work complete until `mage lint` passes.
- Use the CARD-IMPLEMENTATION-GUIDE.md as your primary reference, not general MTG knowledge.

</what-to-do>
