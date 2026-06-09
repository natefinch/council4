# Oracle text

Package `oracle` is the deterministic front end for turning Scryfall
`oracle_text` into council4's typed `game.CardFace` ability data. It is kept
inside `cardgen` because parsing card text is generation-time tooling, not
runtime game behavior.

**Cards supported: 2,010 / 37,628**

The pipeline is:

```text
Oracle text -> lexer -> syntax tree -> semantic compiler -> CardFace data
```

## Lexer

`NewLexer(source)` constructs a synchronous pull scanner. Repeated calls to
`Next` return tokens until `EOF`.

The lexer recognizes structural Oracle-text syntax:

- words and integers;
- braced symbols such as `{T}`, `{2/W}`, and `{R/G}`;
- significant newlines;
- punctuation, parentheses, and quoted granted abilities;
- modal bullets (`•`) and ability-word em dashes (`—`);
- loyalty and power/toughness signs;
- standalone possessive apostrophes, brackets, ampersands, and other printable
  glyphs used by unusual card names or supplemental products.

English vocabulary is intentionally not encoded as token kinds. For example,
`Whenever`, `target`, and `destroy` are all `Word` tokens. Their meaning depends
on surrounding syntax and card-face context, so the parser and compiler own
that interpretation.

Horizontal whitespace is skipped. Every emitted token stores its exact source
slice and a half-open byte span. Positions also include one-based rune line and
column coordinates for diagnostics. Byte offsets are authoritative for slicing.
CRLF is emitted as one `Newline` token.

A UTF-8 BOM is accepted only at byte zero. Valid but unclassified Unicode is
preserved as a `Glyph` token so the parser can diagnose it in context. Invalid
UTF-8, NUL, later BOMs, and unclosed braced symbols produce `Invalid` tokens.
Invalid input always consumes bytes, allowing callers to diagnose an error and
continue without stalling.

## Example

```go
lexer := oracle.NewLexer("{T}: Add {G}.")
for {
	token := lexer.Next()
	if token.Kind == oracle.EOF {
		break
	}
	// Inspect token.Kind, token.Text, and token.Span.
}
```

## Syntax parser

`Parse(source, context)` returns a lossless `Document` plus diagnostics. Card
context identifies instant and sorcery faces because otherwise identical text
can be a spell ability or a static ability. It also identifies planeswalker
faces so loyalty costs are not confused with ordinary activated abilities.

The syntax tree preserves ordered abilities and exact source spans. It
represents ability-word prefixes, top-level activation costs, sentences,
parenthesized reminder text, quoted granted abilities, and modal choose headers
with bullet options. Mode spans exclude the bullet marker. Delimiters inside
quotes or reminder text remain owned by that enclosing construct rather than
creating overlapping sibling nodes. The parser classifies spell, activated, loyalty, triggered,
replacement, static, and reminder paragraphs. This classification is syntactic;
lowering English phrases into executable game primitives is a separate compiler
stage.

Malformed delimiters and lexical errors produce localized diagnostics. Parsing
continues at paragraph boundaries, so callers receive a partial tree rather
than losing the remainder of the card.

## Semantic compiler

`Compile(source, context)` runs the lexer and parser, then lowers the syntax
tree into a source-spanned semantic intermediate representation.
`CompileDocument` accepts an existing syntax tree when callers need to inspect
or transform it first.

The intermediate representation mirrors the information needed by categorized
`game.CardFace` abilities without constructing runtime game values yet. It
records:

- ordered activated and loyalty cost components;
- trigger clauses and intervening-if conditions;
- modes and inclusive target cardinalities;
- conservative selectors and controller constraints;
- keyword abilities and parameters;
- instruction verbs, fixed amounts, mana symbols, negation, and common
  durations;
- card-name, `this`-object, `that`-object, and pronoun references.

Recognition is deliberately conservative. Reminder and quoted text do not leak
into the containing ability's semantics. Trigger conditions and activation
costs are excluded from resolving effects. Any non-reminder construct that has
neither a recognized action nor keyword receives a warning diagnostic covering
its exact source span. Unknown costs receive their own warning. The compiler
never substitutes guessed executable behavior for unsupported wording.

The strict executable backend currently lowers plain non-parameterized
keywords, mana-cost Ward, Cycling, and Equip, base-type Enchant, color-based
Protection, supported tap mana choices, ordinary activated abilities with exact
mana-only, tap-only, or mana-then-tap costs and supported effect bodies,
unconditional enters-tapped replacements, fixed single-target damage,
destruction, exile, return-to-hand, and power/toughness changes, narrow mass
destruction, fixed draw and life changes, fixed controller scry and surveil,
exact investigate and proliferate, fixed controller or target-player discard
and mill, one-target tap, untap, and regeneration, and exact fights between two
target creatures. Supported sentence-sized effects may be lowered in Oracle
order when at most one clause targets. It also lowers exact supported self-enter
and self-dies triggers with ordered supported spell-like effects. An exact
leading `you may` on a single-effect trigger maps to trigger-level optionality;
partially optional sequences remain unsupported. Every semantic element and
meaningful source token must be consumed; otherwise the whole card is rejected.

This compiler IR is the recognition stage. The strict backend in `cardgen`
consumes it and lowers each recognized ability into a second, **typed**
intermediate representation made of `game.*` values (`game.ActivatedAbility`,
`game.ManaAbility`, `game.TriggeredAbility`, and so on), assembles a
`game.CardDef`, validates it with `game.ValidateCardDef`, and only then renders
Go source. This compiler package stays purely about Oracle-text recognition; it
never constructs runtime `game` values itself. See
[`cardgen/README.md`](../README.md#compiler-stages)
and [ADR 0008](../../docs/adr/0008-typed-ir-lowering.md).

## Testing

Unit tests cover representative activated, loyalty, modal, keyword, reminder,
Class, and quoted-ability text. A fuzz test enforces termination and span
invariants. When the ignored local Scryfall cache is available at
`.cardwork/deck/cache/scryfall`, the package tests every root and face
`oracle_text` entry and rejects any invalid token. Compiler corpus tests also
require every non-reminder ability to produce semantic content or an explicit
unsupported diagnostic.

## Full-corpus lexer check

`cmd/checklexer` streams a Scryfall card bulk-data array and checks every root
and card-face Oracle text with a bounded parallel worker pool. It emits
deterministic JSON or text reports listing unsupported cards, exact invalid
tokens, reasons, and source spans. See
[`cmd/checklexer/README.md`](cmd/checklexer/README.md) for usage.

`cmd/checkparser` performs the corresponding full-corpus lexer-plus-parser
check, including card-face type context for spell and loyalty classification.
See [`cmd/checkparser/README.md`](cmd/checkparser/README.md).

`cmd/compilecards` performs strict semantic compilation and bulk source
generation. It emits only fully executable cards and reports every unsupported
card without creating a partial definition. See
[`cmd/compilecards/README.md`](cmd/compilecards/README.md).
