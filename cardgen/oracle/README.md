# Oracle text

Package `oracle` is the deterministic front end for turning Scryfall
`oracle_text` into council4's typed `game.CardFace` ability data. It is kept
inside `cardgen` because parsing card text is generation-time tooling, not
runtime game behavior.

**Cards supported: 3,402 / 31,835**

The pipeline is:

```text
Oracle text -> lexer -> syntax tree -> semantic compiler -> CardFace data
```

Playable `token` and `double_faced_token` records are generated under
`mtg/cards/tokens/<letter>` with their complete normalized Oracle UUID in both
the filename and Go identifier. This keeps token identities distinct from
sanctioned cards and from same-name tokens.

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
faces so loyalty costs are not confused with ordinary activated abilities, and
Saga faces so Roman-numeral chapter headings are not confused with ability words.

The syntax tree preserves ordered abilities and exact source spans. It
represents ability-word prefixes, top-level activation costs, sentences,
parenthesized reminder text, quoted granted abilities, Saga chapter numbers, and
modal choose headers with bullet options. Mode spans exclude the bullet marker. Delimiters inside
quotes or reminder text remain owned by that enclosing construct rather than
creating overlapping sibling nodes. The parser classifies spell, activated, loyalty, triggered,
chapter, replacement, static, and reminder paragraphs. This classification is syntactic;
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

- ordered activated and loyalty cost components, including `{T}`, `{Q}`, exile,
  and counter-removal costs;
- trigger clauses and intervening-if conditions;
- modes and inclusive target cardinalities;
- conservative selectors and controller constraints;
- keyword abilities and parameters;
- instruction verbs, fixed, exact `X`, and typed dynamic amounts, mana symbols,
  negation, and common durations;
- card-name, `this`-object, `that`-object, and pronoun references.

Recognition is deliberately conservative. Reminder and quoted text do not leak
into the containing ability's semantics. Trigger conditions and activation
costs are excluded from resolving effects. Any non-reminder construct that has
neither a recognized action nor keyword receives a warning diagnostic covering
its exact source span. Unknown costs receive their own warning. The compiler
never substitutes guessed executable behavior for unsupported wording.

The strict executable backend currently lowers plain non-parameterized
keywords, exact `Devoid (This card has no color.)`, positive-integer Toxic, and
mana-cost Kicker, Madness, Morph, Disguise, Ward, Cycling, Ninjutsu, and Equip. It also
lowers base-type Enchant, color-based Protection, supported tap and untap mana
choices, ordinary activated abilities with exact typed costs and supported
effect bodies, and exact trailing activation timing restrictions,
unconditional enters-tapped replacements and common land-count or basic-land-subtype
conditions, fixed, `X`, or supported typed dynamic single-target damage,
destruction, exile, return-to-hand, and power/toughness changes with common
controller, tapped-state, and combat-state target restrictions, narrow mass
destruction, fixed, `X`, or supported typed dynamic draw and life changes,
dynamic target-creature P/T changes for exact `for each` and `where X` formulas,
fixed controller scry and surveil,
exact investigate and proliferate, fixed controller or target-player discard
and mill, fixed +1/+1 and -1/-1 counter placement on one target permanent,
one-target tap, untap, and regeneration, exact fights between two
target creatures, and fixed power/toughness buffs on enchanted creature, equipped creature,
creatures you control, other creatures you control, Walls, artifacts, tokens,
and creatures your opponents control. These exact static buffs may also grant one or more supported
non-parameterized keywords. Exact standalone grants lower for the same
controlled-creature and attached-creature subjects, known controlled creature
subtypes, and controlled artifacts, Walls, and tokens. Source-relative grants also lower for exact
`as long as` conditions that require controlling supported permanent types,
subtypes, colors, or colorless permanents.
Exact `Choose N` and `Choose one or both` modal headers lower to runtime-enforced
minimum and maximum mode counts when every mode is otherwise supported.
It also lowers exact `This creature can't block.`,
`This creature can't be blocked.`, `This creature attacks each combat if
able.`, and `This spell can't be countered.` static declarations to
source-scoped rule effects in their appropriate zones.
Adventure, split, and exact enters-prepared layouts are supported when each
printed face is otherwise
exactly representable; these layouts keep the front face in the root
`game.CardDef`, emit the second spell face as `Alternate`, and derive per-face
colors from mana costs when Scryfall omits face colors. An exact
`This creature enters prepared.` ability lowers to `CardFace.EntersPrepared`;
other effects that prepare or unprepare permanents remain deferred.
Supported sentence-sized effects may be lowered in Oracle order with independent
targets for each supported clause. It also lowers exact supported self-enter and self-dies triggers with
ordered supported spell-like effects. Self-enter triggers may use exact
intervening conditions for `if it was kicked`, cast entry, or controlling a
permanent of a named permanent card type. Exact non-self permanent
enters-the-battlefield triggers lower for single-subject (`a`/`an`/`another`,
optional `nontoken` qualifier) and `one or more` subject forms, with optional
permanent type filter (creature, artifact, enchantment, land, planeswalker, or
unfiltered) and optional you-control or opponent-controls controller constraints.
Phase and step triggered abilities with `At the beginning of …` lower for ten
exact step-trigger phrases: your upkeep, each upkeep, each player's upkeep, each
opponent's upkeep, your end step, each end step, each player's end step, combat
on your turn, each combat, and your draw step. Intervening-if conditions on step
triggers are always rejected fail-closed.
Self-dies triggers support exact
absence checks for +1/+1 or -1/-1 counters. Exact fixed-damage self-dies
triggers using `it` preserve the departed permanent as the damage source.
Exact self-dies event-card references support returning the card from its
owner's graveyard to hand and granting its Adventure face graveyard-cast
permission through the end of the controller's next turn. An exact
leading `you may` on a single-effect trigger maps to trigger-level optionality;
partially optional sequences remain unsupported. Exact ordinary battlefield
activations may combine mana, tap, and untap costs with typed sacrifice,
discard, pay-life, source-exile, graveyard-exile, and source-counter-removal
costs. Every semantic element and meaningful source token must be consumed;
otherwise the whole card is rejected.

Supported dynamic effect amounts are deliberately closed: exact creature,
artifact, enchantment, land, or permanent counts (controlled by you, controlled
by opponents, or on the battlefield), opponent count, controller life, and an
exactly referenced source object's power. Count and opponent formulas may use
their printed integer multiplier or “twice.” Arithmetic offsets, mixed groups,
zone counts, and ambiguous pronouns remain unsupported.

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

`cmd/corpusdelta` orchestrates expansion-corpus compilation, stable-ID report
comparison, `docs/supported.md` regeneration, generated-package validation, and
review-manifest generation. See
[`cmd/corpusdelta/README.md`](cmd/corpusdelta/README.md).
