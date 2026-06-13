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

`effect_syntax.go` composes resolving instructions from parser-owned productions.
Each `Sentence` carries ordered, source-spanned `EffectSyntax` and `TargetSyntax`
nodes. Effects carry their typed verb and contextual variant, fixed or dynamic
amount, power/toughness deltas, duration and delayed timing, local Selection,
origin and destination zones, counter kind, exact add-mana output, replacement
modifier, static subject, references, and embedded resolution payment. Each
effect also owns its exact clause, targets, references, and grammatical-subject
targets/references; coordinated follow-ons carry an explicit prior-subject
context instead of inferring it from verb spelling. Targets carry typed cardinality
and a Selection containing object kind, controller relation, flags, types,
supertypes, subtypes, colors, keyword, zone, and numeric filters. Retained text
and tokens are lossless metadata, not the source of downstream meaning.
Target selections require every token in the noun phrase to belong to a typed
atom or a narrow composition production; unknown qualifiers and unknown
cardinalities invalidate the target rather than weakening it.

Effect grammar excludes activation costs, trigger introductions, reminder text,
quoted text, and typed trailing activation restrictions. Coordinated instructions
remain ordered clauses, while malformed dynamic formulas, payments, contextual
verbs, and target forms fail closed at the narrow production that could not be
recognized. Specialized replacement modifiers are attached only to the replacing
effect and reject selection modifiers that the runtime replacement cannot
represent.

It also owns the reusable, composable semantic atoms that downstream stages
consume without re-inspecting source spelling. `atoms.go` recognizes colors,
card types, supertypes, subtypes, object nouns, zones, counter kinds, cardinal
and ordinal number words, selection modifiers, and plural→singular noun
normalization, returning typed values. `keyword.go` owns the complete supported
keyword vocabulary and emits source-spanned `Keyword` syntax with composable
typed parameter shapes: mana costs, integers, Enchant targets, and Protection
predicates over colors, card types, and creature/land subtypes. It also emits
typed `with`/`without` keyword-selector syntax. Mana-symbol parsing, canonical
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
