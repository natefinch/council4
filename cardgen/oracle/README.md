# Oracle text

Package `oracle` is the deterministic front end for turning Scryfall
`oracle_text` into council4's typed `game.CardFace` ability data. It is kept
inside `cardgen` because parsing card text is generation-time tooling, not
runtime game behavior.

**Cards supported: 4,790 / 31,838**

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
modal choose headers with bullet options, including modal headers after an
activated-ability cost. Triggered abilities carry a source-spanned trigger
clause and typed introduction. Recognized phase and step clauses additionally
carry composable quantifier, player/controller relation, attached-subject, and
literal phase/step-name nodes. Their raw clause text and tokens remain only as
lossless source metadata; unsupported or ambiguous grammar keeps that metadata
without receiving a typed phase/step node. Simple source-rule sentences carry a
composable typed declaration: source creature or spell subject, prohibition or
requirement, attack/block/counter operation with active or passive voice, and
source-spanned `each combat` and `if able` qualifiers. The parser owns the
accepted literal grammar, including `can't`/`cannot` prohibition forms and
implicit or explicit must-attack forms. Unknown combinations remain ordinary
source-spanned sentences. Mode spans exclude the bullet marker. Delimiters
inside quotes or reminder text remain owned by that enclosing construct rather
than creating overlapping sibling nodes. The parser classifies spell,
activated, loyalty, triggered, chapter, replacement, static, and reminder
paragraphs. This classification is syntactic; lowering typed syntax and other
English phrases into executable game primitives is a separate compiler stage.
Activated abilities also carry ordered, source-spanned typed nodes for trailing
`Activate only` restrictions. The grammar composes sorcery-timing forms,
once-per-turn count and period forms, and the existing phase/step quantifier,
player relation, and name nodes. Consecutive restriction sentences remain
separate nodes so supported combinations compose without sentence aliases.
Unknown or ambiguous `Activate only` grammar becomes an explicit unsupported
restriction, while `Activate only if` remains activation-condition syntax.

Malformed delimiters and lexical errors produce localized diagnostics. Parsing
continues at paragraph boundaries, so callers receive a partial tree rather
than losing the remainder of the card.

## Semantic compiler

`Compile(source, context)` runs the lexer and parser, then lowers the syntax
tree into a source-spanned semantic intermediate representation.
`CompileDocument` accepts an existing syntax tree when callers need to inspect
or transform it first.

The intermediate representation mirrors the information needed by categorized
`game.CardFace` abilities without constructing runtime game values yet. The
reusable body content — targets, conditions, effects, keywords, references, and
nested modes — is grouped into a single `oracle.AbilityContent` value. Each
`oracle.CompiledAbility` carries its shell semantics (cost, trigger clause,
activation timing and typed zone of function, loyalty change, chapter numbers,
text, span, optional flag) plus one
`oracle.AbilityContent`; each `oracle.CompiledMode` likewise carries its mode
text and span plus one `oracle.AbilityContent`. The content group is the unit
passed to `lowerAbilityContent` in `cardgen`. It records:

- ordered activated and loyalty cost components, including `{T}`, `{Q}`, exile,
  and counter-removal costs;
- trigger clauses and conditions. Conditions use a closed predicate, kind
  (`if`, `unless`, `only if`, or `as long as`), negation, threshold, counter,
  bound source/event object reference, and Selection vocabulary (including
  tapped state). Exact wording recognition lives in the condition adapter;
  unsupported wording remains an explicit predicate. Event-history
  conditions (`ConditionPredicateEventHistory`) carry a `TriggerPattern` and a
  window (current-turn or previous-turn) so the lowering layer can delegate
  directly to `lowerTriggerPattern` and runtime evaluation reuses
  `triggerMatchesEvent`;
- source-spanned semantic trigger patterns. Typed phase/step syntax lowers
  directly without consulting event wording. A small registry of exact
  event-family templates recognizes the remaining permanent zone-change,
  spell/ability, combat, permanent-state, and player events. Wording variants
  share those templates and may bind only closed trigger kind, event,
  self/attached-source and controller relation, Selection, affected-player,
  zone, combat-qualifier, batching, and intervening-condition slots. Unknown,
  ambiguous, or unsupported syntax fails closed. Raw event-clause text is
  retained only for diagnostics and exact source consumption;
- modes and inclusive target cardinalities;
- conservative selectors and controller constraints;
- keyword abilities and parameters;
- instruction verbs, fixed, exact `X`, and typed dynamic amounts, mana symbols,
  negation, and common durations;
- card-name, `this`-object, `that`-object, and pronoun references. A conservative
  binding pass records whether each occurrence denotes the source, a specific
  target occurrence, the triggering event permanent, the triggering event
  player, or a prior instruction result. Player-event trigger bodies
  (`ReferenceBindingEventPlayer`) bind "they/their/them" when the trigger
  event's subject role is authoritatively a player (draw, discard, cycle,
  scry, surveil, life events); permanent-event trigger bodies continue to
  bind "they/their/them" via `ReferenceBindingEventPermanent` when the
  trigger event's subject role is authoritatively a permanent (attack,
  die, tap, untap, and related events). Ambiguous and unsupported occurrences
  remain explicit and fail closed; the compiler never guesses an antecedent;
- source-spanned Static Declarations attached directly to a static ability,
  separate from resolving `AbilityContent`. Their closed semantic vocabulary
  records affected group domain plus Selection, source exclusion, optional
  condition, continuous-effect layer and operation, rule domain and operation,
  zone, cost modifier, or non-battlefield card-ability grant. Typed simple
  source-rule declarations lower into this vocabulary solely from their syntax
  subject, constraint, operation, voice, and qualifiers; retained sentence text
  and tokens are source metadata and are not inspected on that path. Static
  Declarations never resolve and do not reference runtime `game` values;
- activation timing restrictions lower only from typed parser nodes. The
  compiler maps typed sorcery timing, once-per-turn frequency, combat, and
  controller-relative upkeep nodes, composes sorcery timing with once per turn,
  and derives exclusion spans from the ordered nodes. It does not inspect
  retained sentence wording or tokens on this path.

Recognition is deliberately conservative. Reminder and quoted text do not leak
into the containing ability's semantics. Trigger conditions and activation
costs are excluded from resolving effects. Any non-reminder construct that has
neither a recognized action nor keyword receives a warning diagnostic covering
its exact source span. Unknown costs receive their own warning. The compiler
never substitutes guessed executable behavior for unsupported wording.

The strict executable backend currently lowers plain non-parameterized
keywords, exact `Devoid (This card has no color.)`, positive-integer Toxic, and
mana-cost Kicker, Madness, Morph, Disguise, Ward, Cycling, Ninjutsu, and Equip. It also
lowers base-type Enchant, fixed color, card-type, subtype, multicolored,
monocolored, each-color, and everything Protection, supported fixed and choice
mana outputs with exact typed activation costs, ordinary and modal activated
abilities with exact typed costs and supported effect bodies, and exact trailing
typed activation timing restrictions,
unconditional enters-tapped replacements and common land-count or basic-land-subtype
conditions, fixed, `X`, or supported typed dynamic single-target damage,
destruction, exile, return-to-hand, and power/toughness changes with common
controller, tapped-state, and combat-state target restrictions, narrow mass
destruction, fixed, `X`, or supported typed dynamic draw and life changes,
dynamic target-creature P/T changes for exact `for each` and `where X` formulas,
fixed controller scry and surveil,
exact investigate and proliferate, fixed controller or target-player discard
and mill, fixed, `X`, and supported dynamic recognized named-counter placement
on one valid target permanent or player, excluding Stun and Finality counters
until their mandatory runtime mechanics are implemented (#222 and #223),
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
It also lowers the typed simple source-rule grammar for a source creature that
cannot block, cannot be blocked, or must attack each combat if able, and for a
source spell that cannot be countered, to source-scoped rule effects in their
appropriate zones. `can't` and `cannot` prohibition forms compose to the same
rule, as do implicit `attacks each combat if able` and explicit
`must attack each combat if able` requirements.
All supported static power/toughness changes, keyword grants, these rule
declarations, Cycling cost modifiers, and hand-card Cycling grants first
recognize into semantic Static Declarations and then use one shared mechanical
lowering adapter. Fully understood mixed paragraphs may produce multiple
declarations; for example, Dragon's Rage Channeler's Delirium paragraph
produces separate conditional power/toughness, Flying-grant, and must-attack
declarations. Adjacent unsupported groups, conditions, durations, operations,
and shells fail closed with capability-specific diagnostics rather than a
wording-family fallback.
Adventure, split, and exact enters-prepared layouts are supported when each
printed face is otherwise
exactly representable; these layouts keep the front face in the root
`game.CardDef`, emit the second spell face as `Alternate`, and derive per-face
colors from mana costs when Scryfall omits face colors. An exact
`This creature enters prepared.` ability lowers to `CardFace.EntersPrepared`;
other effects that prepare or unprepare permanents remain deferred.
Supported sentence-sized effects may be lowered in Oracle order with independent
targets for each supported clause. It lowers exact supported permanent
zone-change triggers with ordered supported spell-like effects. Self-enter
triggers may use exact
intervening conditions for `if it was kicked`, cast entry, or controlling a
permanent of a named permanent card type. One shared recognizer covers exact
self, attached, single-subject (`a`/`an`/`another`), and `one or more` permanent
enter, die, leave, exile, return-to-hand, and battlefield-to-graveyard clauses.
It binds exact controller/owner relations, origin/destination zones, self
exclusion, face-down state, and Selection predicates for type unions,
supertypes, subtypes (including Outlaw), colors, token/tapped/combat state,
keywords, mana value, power, and toughness. `Leaves ... without dying` is an
exact excluded-destination pattern.
The parser grammatically composes supported phase and step triggered abilities
from `At`, an exact boundary introduction, a quantifier, a player/controller
relation, and a literal upkeep, draw, end, combat, combat-step, or main-phase
name. It explicitly parses irregular first/second-main-phase, end-of-combat,
turn-qualified combat, and enchanted-permanent-controller forms. The semantic
compiler maps those typed nodes to shared relation-and-step slots without
inspecting Oracle wording. Combat templates bind
named/self/attached and semantic Selection subjects, the other blocking
combatant, attacked player or permanent recipients, damage-source and
damage-recipient Selections, combat/noncombat qualifiers, and exact player
relations. Player-level attack wording and `one or more` attack, block, and
combat-damage wording bind explicit batch semantics, including per-attack-target
coalescing. Unsupported phrase variants, compound events, temporal qualifiers,
and unavailable runtime relations remain fail-closed.
The same phase/step vocabulary is reused for trailing activation restrictions.
Supported activation grammar includes `as a sorcery`, `at sorcery speed`, and
`any time you could cast a sorcery`; `once` or `one time` with `each`, `every`,
or `per` turn; combat; and controller-relative upkeep. This admits equivalent
wording variants without expanding runtime timing kinds. Unsupported phase,
player, frequency, and mixed-restriction combinations remain typed and fail
closed.
Exact draw-card ordinals and first-time-this-turn life gain/loss, scry, and
surveil wording bind a shared player-event ordinal slot rather than a
phrase-specific runtime matcher.
Self-dies triggers support exact
absence checks for +1/+1 or -1/-1 counters. Exact fixed-damage self-dies
triggers using `it` preserve the departed permanent as the damage source.
Exact singular permanent-zone-change event-card references support
returning the card from its owner's graveyard to hand, and self-dies references
support granting its Adventure face graveyard-cast permission through the end
of the controller's next turn. Bound event permanents also lower through the
shared reference adapter for supported trigger-body effects such as damage,
power/toughness modification, and explore, using runtime LKI after the permanent
leaves the battlefield. Remaining non-self-dies cards require effect or dynamic
amount vocabulary rather than another reference path. Spell-cast triggered
abilities with `Whenever ... casts ...` lower for three exact player prefixes
(`you cast`, `a player casts`, `an opponent casts`) and seventeen exact spell
phrases:
filters: `a spell` (wildcard), `a noncreature spell`, `a creature spell`,
`an instant or sorcery spell`, `an instant spell`/`an instant`, `a sorcery
spell`, `an artifact spell`, `an enchantment spell`, `a land spell`, `a
planeswalker spell`, `a noncreature, nonland spell`, and single-color forms `a
white/blue/black/red/green spell`. Exact Threshold, Delirium, Domain, Metalcraft,
Hellbent, Ferocious, and Coven conditions lower into typed live-state
predicates and dynamic amounts. Event-history intervening conditions recognize
eight exact phrases covering current-turn attacked/died/gained-life/lost-life
and previous-turn lost-life/no-spells-cast; each carries a `TriggerPattern` and
an `EventHistoryWindow` so the lowering path reuses `lowerTriggerPattern` and
runtime evaluation reuses `triggerMatchesEvent`. `Negated` semantics are
preserved for "no spells were cast last turn" and the predicate is only allowed
in the intervening-trigger context. Self-cast (`when you cast this spell`),
`TriggerWhen`, unsupported intervening-if conditions, unknown or non-exact
ability-word forms, modes, and all other player or spell-phrase forms are
fail-closed. An exact
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

This compiler IR is the recognition stage. Trigger phrase tables and Static
Declaration adapters live here; simple source rules arrive from typed parser
syntax, and `cardgen/lower.go` never interprets retained raw trigger-event or
static-declaration text.
Permanent action templates recognize exact tapped, untapped, and turned-face-up
events while binding self, attached, controller-relative, and Selection-filtered
subject relations.
Became-target patterns preserve the targeted subject and the targeting
spell-or-ability controller as independent semantic roles.
Player-event templates preserve controller-relative and any-player Cycling
relations through the same closed player-relation slot.
Sacrifice templates bind the sacrificing player independently from the
sacrificed permanent's shared Selection subject.
Scry and surveil remain distinct player-action event templates; compound event
wordings remain fail-closed.
Activated-ability templates bind the activating player and source-permanent
Selection only for exact non-mana-ability wording. Unrestricted forms remain
unknown until payment-time mana activations join the authoritative event stream.
The strict backend in `cardgen`
consumes it and lowers each recognized ability into a second, **typed**
intermediate representation made of `game.*` values (`game.ActivatedAbility`,
`game.ManaAbility`, `game.TriggeredAbility`, and so on), assembles a
`game.CardDef`, validates it with `game.ValidateCardDef`, and only then renders
Go source. The single entry point for lowering ability body content is
`lowerAbilityContent` in `cardgen/lower.go`; it accepts an `oracle.AbilityContent`
value and an `oracle.Ability` syntax node and is the path used by all supported
shells (spell, activated, triggered, loyalty, chapter, modal option). This
compiler package stays purely about Oracle-text recognition; it
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
comparison, scratch support-list generation, generated-package validation, and
review-manifest generation. See
[`cmd/corpusdelta/README.md`](cmd/corpusdelta/README.md).
