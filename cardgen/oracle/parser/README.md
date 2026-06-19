# Oracle parser

Package `parser` owns Oracle syntax and grammatical recognition.
`Parse(source, Context)` lexes source and returns a lossless `Document` plus
localized diagnostics. `Context` contains only card-face facts needed to
classify syntax: `InstantOrSorcery`, `Planeswalker`, `Saga`, and the card's own
`CardName` (used to recognize self references). `CardName` is threaded onto the
returned `Document`.

The package owns syntax ability kinds, source-spanned phrases and sentences,
literal Oracle vocabulary, typed trigger clauses, activation restrictions,
static-rule syntax, resolving-effect syntax, and selection syntax. Unrecognized
or ambiguous grammar preserves source metadata and fails closed rather than
inventing typed syntax.

## Serializable representation

A `Document` is trivially serializable to human-readable indented JSON. The
parser's shared enum types (ability, effect, condition, selection, trigger,
keyword, reference, cost, and atom kinds, plus `shared.Kind`/`shared.Severity`)
are string-backed (`type Foo string` with constant-name values), so they print
and serialize as readable names rather than integers; each enum's zero value is
the empty string, preserving its old `*Unknown == 0` fail-closed semantics. The
recognitions a body owns are plain exported `Ability`/`Mode` fields (no
value-returning accessors); `Parse` materializes the eagerly-computed semantic
fields during its emit passes. `json` tags exclude representation noise: every
`shared.Span`/`shared.Position` position field, every `[]shared.Token` slice, and
the reminder/quoted `Delimited` token carriers are tagged `json:"-"`, so the JSON
shows only the typed semantic structure. Remaining fields use `omitempty`/
`omitzero` for readability. Because `Ability`/`Mode` carry their inlined fields by
value they exceed the `hugeParam` size threshold; the parser and `cardgen` pass
them by `*Ability`/`*Mode` pointer. The `mage parse "<card name>"` target loads
the cached Scryfall corpus, parses a card's Oracle text with this library
in-process, and prints the resulting `Document` as indented JSON.

Triggered abilities use mutually exclusive typed clause paths for phase/step,
player-event, and all other supported event families. `TriggerEventClause`
composes a source-spanned event kind with typed subjects, actors, selections,
actions, zone movement, recipients, causes, counters, stack objects, and
qualifiers. Its grammar covers spell cast and ability activation; permanent
entry, death, and other zone changes; attack, block, became-blocked, and damage;
counter placement, tap, untap, face-up, sacrifice, mutate, and became-target
events. Zone-change and combat productions compose their verb, subject,
selection, relation, zone, recipient, and qualifier grammar rather than matching
whole event phrases. Exactly one event family must recognize the entire clause;
unknown, ambiguous, partial, and inexact forms keep their lossless `Phrase`
metadata but receive no typed event node. Trigger-event syntax is emitted after
semantic atoms so card-name and explicit self references are recognized here.
Supported event-history conditions use the same typed event clauses plus a
source-spanned current-turn or previous-turn window and explicit negation. The
parser composes their actor, subject, event, and window; unsupported event/window
combinations receive no typed event-history node.

`condition.go` emits typed `ConditionClause` syntax for the remaining supported
conditions. Each clause carries its source span, introducer kind, a closed
predicate (controller life/hand/opponent-count resources, player-life-at-most,
graveyard card and card-type counts, creature power diversity, controls,
event-subject history, counter placement, controlled-damage source, token
creation, source-death, and object match/exists), and any composable parameters:
a control scope and numeric comparison, a literal threshold, a counter kind, an
object binding, a subject span, and a source-independent `ConditionSelection`
(required types, supertypes, canonical subtype identities, colors, colorless,
exclude-source, tapped state, and power filter). Selections are composed from
type, supertype, subtype, color, tapped, and power productions rather than
whole-phrase aliases; a bare subtype noun emits only its subtype identity, while
required types come from explicit card-type words. Exactly one predicate
recognizer must accept the whole clause body; unknown, ambiguous, near-miss, and
partial wordings receive no typed clause so the compiler fails the condition
closed.

`static_declaration.go` and `static_declaration_operations.go` emit typed
`StaticDeclarationSyntax` for every supported static-declaration family. A static
ability composes a source-spanned subject—source creature/spell, the card's own
name, a typed `EffectStaticSubject` group, or the controller's hand—with one or
more ordered operations: power/toughness changes, keyword grants, and the typed
`StaticRuleSyntax` of `static_rule_syntax.go`. A rule operation in a compound
declaration accepts only a single subject—the source or its attached object
(Aura/Equipment)—while battlefield group rules still receive no typed declaration
so the compiler fails closed. Recognized `EffectStaticSubject`
group subjects include battlefield-wide creatures ("All/Other creatures"),
combat-state creatures ("Attacking/Blocking creatures" and "Attacking creatures
you control"), and battlefield creature-subtype groups ("All/Other <Subtype>
creatures"); battlefield color and keyword-filter groups stay unrecognized so the
compiler fails closed. Operations are joined by an
explicit comma/"and" connector, keyword grants compose a lookahead-delimited
keyword list, and a single supported condition clause may scope the whole
declaration. Cost-modifier and card-ability-grant declarations (cycling cost
reductions and replacements, and "Each <land/creature/historic> card in your
hand has cycling {N}") are recognized as their own typed families. The static
source-tied control grant printed on control Auras ("You control enchanted
creature/permanent") is recognized as its own family whose affected group is the
attached object. A power/toughness change is marked dynamic only when a recognized "for each"/"equal
to" tail scales it. Exactly one family must consume the entire body; unknown
verbs, dangling connectors, unsupported keyword slots, and group rules receive no
typed declaration so the compiler fails closed.

`effect_syntax.go` composes resolving instructions from parser-owned productions.
Each `Sentence` carries ordered, source-spanned `EffectSyntax` and `TargetSyntax`
nodes. Effects carry their typed verb and contextual variant, fixed or dynamic
amount, power/toughness deltas (each side independently a fixed integer, zero, or
a variable `X`, so asymmetric and mixed-sign pumps round-trip; a plural or "up to
N" target distributes the pump with the byte-exact `<subject> each get <p>/<t>
until end of turn.` wording), duration and
delayed timing, local Selection,
origin and destination zones, counter kind, exact add-mana output, replacement
modifier, static subject, references, and embedded resolution payment. Exact
add-mana output (`EffectManaSyntax`) carries the recognized symbol strings and,
when every symbol is a basic color token (`{W}{U}{B}{R}{G}{C}`), the typed
`Colors []mana.Color` and `ColorsKnown` flag, so a consumer builds add-mana
content from typed colors instead of re-parsing the rendered symbol strings. Entry
effects distinguish their modification through typed flags—`EntersTappedSelf`
for a plain tapped entry (any subject noun or card-name phrasing),
`EntersWithCounters` for counter entry, `EntersColorChoice` (with
`EntersColorChoiceExclude` naming a forbidden color) for "As ~ enters, choose a
color[ other than <color>]", and `EntersTypeChoice` for "As ~ enters, choose a
creature type."—so downstream stages never re-read the
entry sentence. Each
effect also owns its exact clause, targets, references, and grammatical-subject
targets/references; coordinated follow-ons carry an explicit prior-subject
context instead of inferring it from verb spelling. A prior-subject life change
whose subject is elided (inherited from the prior effect in a compound sentence,
as in "Target player draws two cards and loses 2 life") reconstructs from its
bare third-person verb, but only when its amount is self-contained—a fixed value
or the spell's cost `X`. A trailing "where X is …" amount defines a single `X`
shared by every effect yet binds to only one of them, so that form stays
inexact and the drain sequence fails closed. Targets carry typed cardinality
and a Selection containing object kind, controller relation, flags, types,
supertypes, subtypes, colors, keyword, zone, and numeric filters. Retained text
and tokens are lossless metadata, not the source of downstream meaning.
Target selections require every token in the noun phrase to belong to a typed
atom or a narrow composition production; unknown qualifiers and unknown
cardinalities invalidate the target rather than weakening it. Permanent target
reconstruction byte-exactly rebuilds an optional `with <keyword>` or `without
<keyword>` qualifier and a
`" or "`-joined multi-color filter, and `parseSelection` records a combined
`target player or planeswalker` / `target opponent or planeswalker` recipient via
a `PlayerOrPlaneswalker` flag; fixed-amount group damage recipients likewise
rebuild a `with <keyword>` or `without <keyword>` qualifier after the group noun. A damage recipient
that is the controller or owner of a referenced object—"deals N damage to its
controller", "... to that <object>'s controller", "... to its owner", or "... to
that <object>'s owner"—is recorded on a `DamageRecipientReference` field and gated
on a byte-exact reconstruction of the verb clause (fixed or `X` amount only), so
lowering can aim the damage at the prior removal target's controller or owner
while every unrelated possessive (such as the convoke reminder "that creature's
color") stays untouched. Multi-target and
optional permanent targets (`up to N target <noun>s`, `N target <noun>s`,
`up to one target <noun>`) reconstruct a plain permanent noun with an optional
plural `other` self-exclusion and controller clause, pluralizing the noun and
failing closed for every other qualifier. Keywords whose Oracle
word the parser cannot render stay fail-closed.
Graveyard-card
return/put targets ("Return target <noun> from <owner> graveyard ...") gate on a
byte-exact canonical reconstruction of the noun phrase from the Selection's typed
fields: a single card type, a `" or "`-joined union of card types, a permanent
card, a single color, a colorless or multicolored card, a single subtype, or the
plain "card" noun, with an optional "with mana value N or less" qualifier, an
optional self-exclusion, and an optional multi-target or "up to N" count whose
nouns pluralize. Single instant/sorcery types and any other unrenderable
qualifier (supertype, excluded type, color+type combination, "and/or" union) fail
closed so the card keeps failing rather than lowering to a wrong predicate.
Library-search effects ("Search your library for … , then shuffle.") gate on a
byte-exact canonical reconstruction of the whole clause from the typed Selection
and count: a singular ("a"/"an") or bounded "up to N" search of your own library
for a plain card, a single card type, a `basic` supertype, or a `" or "`/`", "`-
joined union of basic land subtypes, moved to hand or the battlefield (optionally
tapped) and optionally revealed first. A resolving optional tutor ("You may
search your library for …") carries its choice as the effect's `Optional` flag;
the canonical reconstruction strips the leading "you may" so it round-trips
against the same shape as a mandatory tutor. Any rider the runtime `SearchSpec`
cannot express—extra source zones, "with different names",
mana-value/power/color
filters, variable `X` counts, non-basic-land subtype unions, or other
destinations—fails closed.
The same controller-scoped stripping generalizes to other resolving "you may"
bodies: a direct `You may gain N life` or `You may create … token` reconstructs
its canonical verb clause byte-exactly (the leading "you may" is dropped), so the
life and token recognizers mark it `Exact` and the lowerer routes the mandatory
body through normal lowering while flagging the instruction `Optional`. Only a
direct controller "you may" is stripped; non-controller wordings ("each opponent
may", "target player may") keep the "may" in their clause, never round-trip, and
stay fail-closed because a single controller-asked optional instruction cannot
model another player's choice.
Mass return-to-hand effects ("Return all <group> to their owners' hands.", with
the singular "to their owner's hand." used for the `you control` variant) reuse
the shared mass-group phrase recognizer between the "Return all " prefix and the
destination suffix, so the same group filters that mass destroy/exile accept also
recognize a board-wide bounce. The "each", "a permanent you control", "all but
one", and "except for" wordings stay fail-closed.

Effect grammar excludes activation costs, trigger introductions, reminder text,
quoted text, typed trailing activation restrictions, and the typed trailing
trigger-frequency qualifier ("This ability triggers only once each turn.").
Coordinated instructions
remain ordered clauses, while malformed dynamic formulas, payments, contextual
verbs, and target forms fail closed at the narrow production that could not be
recognized. Specialized replacement modifiers are attached only to the replacing
effect and reject selection modifiers that the runtime replacement cannot
represent.

A destroy effect immediately followed by a regeneration rider — a separate
zero-effect sentence "It can't be regenerated." or "They can't be regenerated." —
folds onto the destroy as a `PreventRegeneration` flag with a coverage span over
the rider sentence, and the rider's pronoun is dropped from the ability's semantic
references. Crediting is restricted to the "it"/"they" pronoun forms and applies
only when the ability has exactly one effect, that effect is an exact destroy, and
no other sentence is unrecognized; subject-phrase forms ("That creature …", "A
creature destroyed this way …") and any other shape stay fail-closed.

It also owns the reusable, composable semantic atoms that downstream stages
consume without re-inspecting source spelling. `atoms.go` recognizes colors,
card types, supertypes, subtypes, object nouns, zones, counter kinds, cardinal
and ordinal number words, selection modifiers, and plural→singular noun
normalization, returning typed values. `keyword.go` owns the complete supported
keyword vocabulary and emits source-spanned `Keyword` syntax with composable
typed parameter shapes: mana costs, integers, Enchant targets, and Protection
predicates over colors, card types, and creature/land subtypes. It also emits
typed `with`/`without` keyword-selector syntax. A keyword whose name span is
covered by a `with`/`without` keyword-selector qualifier (e.g. the "flying" in
"creature with flying") is excluded from the ability's semantic keywords, so it
stays a target/group filter and never doubles as a content keyword ability.
Mana-symbol parsing, canonical
keyword names, Protection list grammar, and Enchant target normalization live
only in the parser; malformed or ambiguous parameter grammar leaves the keyword
unparameterized and therefore fails closed downstream. `references.go`
recognizes explicit self/source references (the card's own name, `this`/`that`
objects, and exact pronouns) as typed `Reference` values. `Parse` emits these atoms
as source-spanned typed values attached to each `Ability` and modal `Mode` node
(the `Atoms` field), so the compiler consumes them by span rather than calling
recognizers on raw tokens. Recognizers fail closed on unknown or ambiguous
spelling.

`parser` imports `lexer` and `shared`, never `compiler`. `ParseSentences` is the
lossless sentence splitter used internally and remains available to syntax
clients; semantic compilation consumes the typed nodes emitted by `Parse`.

Per-ability and per-mode semantic scoping is parser-owned so the compiler never
re-scans token slices. `semantic_scan.go` materializes `SemanticReferences` and
`SemanticKeywords` fields on `Ability` and `Mode`, holding the references and
keywords already scoped to that body (with any ability-word span excluded), plus
a `ContentSpan` field for the body's content span. Each `Reference` carries its
rendered `Text` so the compiler copies the display string rather than rejoining
tokens.
`condition_segment.go` emits `ConditionSegment` values via
`Ability.ConditionSegments`, `Ability.TriggerConditionSegments`, and
`Mode.ConditionSegments`: each segment carries its kind, source span, rendered
text, intervening flag, and any "Activate" keyword span, reproducing the clause
segmentation the compiler once derived by scanning `Period`/`Comma` punctuation.
The ability word is the typed `AbilityWordClause` (label plus span), the trigger
event is a rendered `Event` string plus `EventSpan`, and the source cost phrase is
parser-internal. The compiler-facing AST therefore exposes no `parser.Phrase` and
the compiler ingests no raw `shared.Token` stream to recognize keywords or
references, segment conditions, or reconstruct rendered text.

The parser also owns all positional reasoning, emitting it as typed
relationships so the compiler never inspects source-span byte offsets. Node
identity is a stable `NodeID`: `collectReferences` numbers every reference in an
ability's (or mode's) authoritative reference set, `condition_boundary.go`
numbers each `ConditionBoundary`, the matching `ConditionSegment` inherits that
`NodeID`, and `recognizeSourceDeathCondition` records the source-subject
reference's `NodeID` as `ConditionClause.SubjectRefID`. Source order and
containment are dense per-ability ranks: `source_order.go`'s `emitSourceOrder`
runs last in the pipeline, ranks the union of every participating node's span
boundaries (references, effects and their verbs, targets, the trigger, cost and
its components, condition segments, payments, and static-rule spans, across the
ability and its modes), and stamps each node's `Order`/`VerbOrder`
(`shared.SourceOrder`) field. Because a dense rank is strictly monotonic in the
underlying offset, every order comparison and span-containment test the compiler
once computed from offsets is reproduced exactly in rank space while absolute
positions are discarded.

Ability-level recognitions that downstream stages once derived from Oracle
wording are emitted as typed `Ability`/`Modal` fields. Modal headers carry typed
minimum/maximum mode counts (`Modal.MinModes`/`MaxModes`/`ChoiceKnown`),
recognizing `Choose one or both` and fixed cardinal choices and failing closed on
non-numeric forms. Saga lore-counter reminders (`Ability.SagaReminder`), Read
Ahead recognition and its sacrifice chapter (`Ability.ReadAheadRecognized`/
`ReadAheadSacrificeChapter`, recognized through the roman-numeral chapter
grammar), and Devoid recognition (`Ability.DevoidRecognized`) are parser-owned
typed flags; their fixed reminder vocabulary is recognized here so lowering never
inspects the reminder text. A fully-parenthesized reminder ability ("({T}: Add
{G}.)") has its inner text parsed once into a typed inner document, exposed via
`Ability.ReminderInner()`; a consumer lowering a reminder mana ability compiles
and lowers that typed inner document instead of re-parsing the reminder wording
itself.

Condition introduction, optionality, and activation costs are likewise emitted as
typed syntax. `condition_boundary.go` emits an ordered `[]ConditionBoundary` per
ability and mode: each boundary carries the introducer kind (`if`/`unless`/`only
if`/`as long as`), the triggered intervening-if position, a duration-skip flag for
"for as long as ..."/"as long as this [type] remains/is on the battlefield"
source durations, and the span of an "Activate" keyword preceding an "only if"
restriction. The same pass drops the trailing "if able" of "attacks each combat if
able" so it never becomes a standalone condition. `emitOptional` sets
`Ability.Optional`/`OptionalSpan` for a triggered "you may" body. `cost.go` emits
the typed `Cost`/`CostComponent` grammar, including mana-symbol components and the
"from your graveyard" source zone. Sacrifice cost objects recognize a subtype
("Sacrifice a Goblin"), an explicit count ("Sacrifice three Treasures"), the
source itself ("Sacrifice this Aura"/"Sacrifice this Equipment" via `SourceSelf`),
and "another" via the `ExcludeSource` flag ("Sacrifice another creature");
"Exile this card from your graveyard" sets `SourceSelf` with a graveyard source
zone. Graveyard-exile card objects recognize a fixed count, a typed card noun
("exile a creature card"), an explicit count ("exile two creature cards"), and an
`X`-bound count ("exile X cards from your graveyard") via `AmountFromX`.
Unrecognized sacrifice or exile wordings reset to no typed object so the
compiler fails the cost closed. The compiler and lowering consume all of these
by source position or typed value; they never inspect introducer, "you may",
mana-symbol, or "Activate" spelling. This boundary is enforced by the
`TestCompilerIsTextBlind` and `TestLoweringTextInterpretationIsAllowlisted` AST
analyzers in package `cardgen`.

Structural body boundaries are emitted as typed spans so consumers slice an
ability's token stream at parser-recognized boundaries instead of scanning for
separator token kinds. `Ability.BodySpan` is the source span of the resolving
body (after the activated/loyalty cost colon, the triggered event comma, and any
ability-word or Saga chapter prefix); `Ability.BodySeparatorSpan` is the span of
the single separator token that introduces it (the colon, comma, or chapter em
dash); and `AbilityWordClause.SeparatorSpan` is the em dash that follows an
ability-word label. The exported `TokensInSpan(stream, span)` and
`TokensFrom(stream, offset)` helpers return the contiguous token sub-slice a
consumer needs to build a body sub-ability, keyed off these typed boundaries
rather than off colon/em-dash/comma token kinds.

`Ability.CoverageSpans()` (and `Mode.CoverageSpans()`) emit the ability's
"must-cover" source spans: every token except the structural sentence
punctuation the parser owns (the commas, colons, and periods that separate
clauses and costs). A consumer enforcing a fail-closed source-coverage gate
asserts each must-cover span is covered by a span it recognized, instead of
walking the raw token stream and classifying token kinds itself; reminder,
quoted, and separator tokens stay in the set so an ability with un-recognized
reminder or separator text still fails closed.

`DocumentCoverage(doc)` and `AbilityCoverage(a)` build on `CoverageSpans()` to
measure parser-only coverage: how completely the parser represents an ability as
typed syntax, with no compiler or lowering involvement. Two distinct metrics
come out of this. An ability is **parser-complete (typed coverage)** when every
must-cover span is accounted for by a span reconstructed from recognized typed
output (typed effect clauses, a recognized trigger or cost, recognized condition
segments and static declarations, keywords, references, reminders, the
ability-word clause, chapter headings, the additional-cost declaration, and
recognized-construct spans for a coordinated card-type/subtype list, a "for each"
iteration prefix, and a reflexive/delayed trigger preamble) and
every condition introducer resolves to a recognized clause; a modal ability also
requires a known choice header and recognized modes. Typed coverage only needs a
kind-recognized element (`Kind != EffectUnknown`), so it is an upper bound on
what the lowerer could consume and is **not** the same as byte-exact
reconstruction. The strictly stronger **exact round-trip** metric counts an
effect only when the parser reproduced its text byte-for-byte (`Exact`) and its
whole sentence is represented, so a sentence with any unrepresented clause
contributes no exact effects even if one clause round-tripped. A recognized
effect credits only the tokens it actually represents: its clause is clipped
backward across every top-level boundary (comma, semicolon, "then", or "and")
before the effect's subject and, when the effect is not exact, forward to the
next top-level boundary, so an adjacent unrepresented clause — leading or
trailing, joined by any connector ("Goad target creature**, then** draw a card")
— stays uncovered instead of being absorbed. Legitimate leading material is kept
covered by its own recognized spans (the trigger event clause, the linked
condition clause, a leading sequencing "then"), not by over-crediting the
effect. The reports carry the resolving/exact effect tallies, the uncovered
token runs, and — clustered by the owning grammatical component — the
unrepresented grammar (`UncoveredComponent`) with a blocker family, which the
`cmd/parsercoverage` tool ranks into a parser work queue. Because lowering is
downstream of parsing, every card the lowerer can generate is parser-complete;
`cmd/parsercoverage` asserts that invariant against `supported.md`. Constructs the
parser recognizes semantically but whose typed output stops short of all their
tokens — coordinated trigger/condition lists, "for each" iteration prefixes, and
reflexive/delayed trigger preambles — emit a span tightly bounded to that
recognized grammar (`appendConstructRecognizedSpans`), so the invariant holds
with zero violations rather than carrying a residue or loosening the metric.
