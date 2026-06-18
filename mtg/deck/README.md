# deck package

`mtg/deck` parses Magic: The Gathering Commander decklists in the standard
Moxfield/MTGO text export format into structured data the engine can load.

A **Decklist** is the file specification — card names with quantities, plus the
commander. It is distinct from the in-game **Library** zone: a Decklist
describes what a player registered, while the Library is where those cards live
during a game.

## Main types

### Entry

`Entry` is one decklist line: a positive `Quantity` and a card `Name` with any
leading quantity and trailing set/collector or foil annotations removed.

### Decklist

`Decklist` holds the parsed `Commander` entries (usually one card, two for
partner/background pairings) and the main-deck `Cards`. `Count` returns the
total number of cards across both.

## Parsing

```go
d, err := deck.ParseFile("atraxa.txt")
// or
d, err := deck.Parse(reader)
```

`Parse` recognizes:

- `N Card Name` and `Nx Card Name` entries.
- A commander section introduced by a `// Commander` header or a `COMMANDER:`
  line (an inline `COMMANDER: Name` designates one commander without changing
  the section for later lines).
- A `// Deck` / `// Mainboard` header to return to the main deck.
- A `// Deck` / `// Mainboard` header to return to the main deck. A blank line
  also ends the commander section, so a `// Commander` layout without an
  explicit `// Deck` header still routes the deck correctly.
- Blank lines, `//` and `#` comments (including category comments such as
  `// Creatures (30)`), and sideboard/companion lines (`SB:` prefix or a
  `// Sideboard` / `// Companion` header), which are ignored.

Trailing set/collector annotations (`(2X2) 117`) and foil markers (`*F*`) are
trimmed from names; real parenthetical card names are preserved.

### Errors

`Parse` always returns a best-effort `Decklist`. When one or more lines cannot
be parsed it skips them and returns an error joining every `*ParseError`; each
carries the 1-based `Line`, the offending `Text`, and a `Reason`. Use
`errors.As` to inspect individual failures.

## Package boundaries

`deck` resolves card *names*; it does not look cards up. The four-deck input
model (a separate package) uses `mtg/cards` to turn an `Entry` name into a
validated `game.CardDef`.
