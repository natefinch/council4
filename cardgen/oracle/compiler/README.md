# Oracle compiler

Package `compiler` owns semantic compilation and the semantic intermediate
representation consumed by card generation. `Compile(parser.Document, Context)`
lowers an already parsed document; it never accepts raw source or invokes the
top-level `parser.Parse` or `parser.ParseSentences` entry points. Compiler
`Context` is currently empty; card-name, source-reference, and resolving-effect
recognition belong to parser syntax.

The package owns compiled abilities and content, semantic trigger patterns,
conditions, references, selectors, static declarations, costs, effects, and
fail-closed semantic diagnostics. Typed parser paths—resolving effects, targets and selections, amounts, durations,
zones, counters, add-mana output, replacement modifiers, references, embedded effect payments, phase/step and
player-event triggers, activation restrictions, spell alternative costs,
source-scoped activation cost reductions, and static-rule syntax—compile
from typed nodes without consulting retained literal text.

The resolving `EffectCantBeBlocked` effect ("Target creature can't be blocked
this turn.") maps from the parser's typed effect kind, carrying the
`DurationThisTurn` duration and the single creature target onto the engine
without inspecting source text. It shares the `EffectCantBeBlocked` enum with the
static can't-be-blocked prohibition, which the parser distinguishes by the "this
turn" turn duration, so the resolving evasion grant and the continuous static
restriction never collide. Compilation stays text-blind and fails closed on every
non-exact wording the parser already rejected.

The `EffectManaSpendRider` effect (Path of Ancestry's spend-linked scry rider)
maps from the parser's typed kind through `compileManaSpendRider`, which copies
the closed `CompiledManaSpendRider{Condition, Effect, ScryAmount}` fields and
never reads source text; nil maps to nil so ordinary effects carry no rider. The
preceding commander-identity add-mana effect keeps its `CommanderIdentity` flag,
so lowering sees the add-mana effect and its rider as a typed pair.

Every supported trigger family reaches `TriggerPattern` through a mechanical
typed adapter. Phase/step, player-event, zone-change, spell/ability, combat,
damage, permanent-state, counter, sacrifice, mutate, and targeting meaning is
already present in parser syntax. `trigger_pattern.go` maps only closed enums,
typed selections, relations, zones, recipients, causes, and qualifiers; it
contains no Oracle trigger wording recognizers or subject-text parsing. Invalid
or partial constructed syntax fails closed, and retained event text and tokens
cannot change compilation. Event-history conditions likewise arrive as typed
parser event syntax and a typed turn window; condition compilation reuses the
same mechanical trigger adapters.

`condition.go` compiles the remaining conditions from typed parser
`ConditionClause` nodes matched to each condition by the parser-resolved
`ClauseIndex`/`EventHistoryIndex` rather than by comparing source spans. It maps the
parser's closed predicate, control scope, comparison, threshold, counter, object
binding, subject span, and `ConditionSelection` onto the engine condition
vocabulary, and contains no Oracle condition wording recognition: no
prefix/suffix/contains text matching and no token-spelling interpretation. A
numeric "at most N" comparison becomes a negated "at least N+1" minimum, and the
introducer kind supplies the base negation. Any clause whose typed selection,
counter, scope, or predicate falls outside the closed semantic vocabulary leaves
the predicate unsupported. Condition boundaries themselves are now parser-owned
typed syntax: the parser emits a `ConditionBoundary` for each introducer,
carrying its introducer kind, intervening-if position, duration-skip
classification, resolving `Then if` classification, and any preceding
"Activate" keyword span, plus a stable
`NodeID`. `compiler.go` matches each boundary to its content condition by that
typed `NodeID` and consumes it mechanically; it no
longer inspects "if"/"unless"/"only if"/"as long as" spelling, deletes an
"if able" restriction by text, or recognizes the "Activate only if" keyword by
spelling.

`static_declaration.go` compiles static declarations from the typed
`StaticDeclarationSyntax` nodes the parser emits, matched to the ability by
declaration family and consumed mechanically. It dispatches on the parser's
ordered declaration kinds—power/toughness change, keyword grant, control grant,
rule, player rule, cost modifier, card-ability grant, continuous characteristic
set/addition ("is [a] <color(s)>"/"is <card type>", including "is all colors"
which sets all five colors), and the polymorph
lose-abilities-become shape—and reads the affected group, deltas,
granted keywords, rule meaning, cost shape, and card filter from already-compiled
content and typed parser payloads. A "this creature" or card's-own-name subject
on a continuous declaration resolves to the source group. It contains no Oracle static-declaration
wording recognition: no `matches*`/token-sequence recognizers, no
prefix/suffix/contains text matching, and no token-spelling interpretation.
Source/group asymmetries (a source keyword grant requires a condition; a group
grant forbids one), the dynamic-amount agreement check, and the supported-rule
table are enforced over typed nodes and compiled effects alone. A compound
power/toughness change paired with a single creature rule (for example
"gets +2/+2 and can't block") derives that rule from the typed parser node rather
than the resolving content, because the resolving compiler drops the rule effect
in compounds; it is recognized only for the source or its attached object, where
the runtime rule-effect model already enforces a single subject. The same
typed-node mapping recognizes the bounded-exception prohibitions
`StaticRuleCantAttackYou` ("can't attack you or planeswalkers you control") and
`StaticRuleCantBeBlockedByMoreThanOne` ("can't be blocked by more than one
creature"), and `StaticRuleCantBeBlockedByCreaturesWith` (the bounded
blocker-restriction prohibitions "can't be blocked by creatures with flying",
"... with power N or less", "... with power N or greater", "... by <color>
creatures", and "... by artifact creatures"), and a keyword
grant may stand in for the power/toughness change
("has hexproof and can't be blocked by more than one creature"). Supported anthem
group subjects map to a typed `StaticSelection` carrying battlefield versus
controller domain, combat state, creature subtype, color, token-only, the
Legendary supertype, tapped state, a single keyword filter (present or excluded),
a conjunctive multi-type requirement (artifact-creature), a nontoken
requirement, and source exclusion; subjects outside that
closed set record a group blocker. The polymorph lose-abilities-become family
lowers to layer-faithful continuous declarations: a remove-all-abilities
ability-layer declaration plus set-color, set-type/subtype, and base
power/toughness declarations that replace the affected object's printed
characteristics. Any ability whose
typed declarations or compiled content fall outside the closed vocabulary records
a structural blocker instead of a declaration.

Reusable semantic atoms—colors including excluded/non-color forms, card types
including excluded/non-type forms, supertypes, subtypes, object nouns, zones,
counters, cardinal and ordinal numbers, and explicit self/source references—are
recognized by the parser and mapped here to engine types (`color.Color`,
`types.Card`, `types.Sub`, `zone.Type`, `counter.Kind`, and reference kinds).
Keywords and keyword selectors are likewise parser-owned typed syntax.
`CompiledKeyword` carries the parser-recognized keyword kind plus typed mana,
integer, Enchant-target, or Protection parameter data; canonical name and
parameter text remain metadata only. The compiler maps typed Protection atoms
to engine predicates but performs no keyword-name or parameter recognition.
`effect_syntax.go` is a mechanical adapter from parser resolving syntax to
compiler IR. It maps enums and typed values; it contains no Oracle vocabulary or
effect recognizers. This includes the parser-owned controlled-permanent static
subject used by temporary group keyword grants; lowering receives
`StaticSubjectControlledPermanents` without re-reading its wording. The core
effect, keyword, target, reference, amount, zone,
counter, trigger, and condition compilation consumes parser syntax or atoms by
span rather than deriving these meanings from token spelling. Genuine identity
values, such as subtypes, remain typed engine values. Compiled effects preserve
parser-owned clause, target, reference, and grammatical-subject ownership so
ordered-effect lowering does not rediscover clause boundaries from tokens. Entry
modifications carry the parser's typed `EntersTappedSelf` and `EntersWithCounters`
flags. A destroy effect's parser-folded regeneration rider arrives as the typed
`PreventRegeneration` flag plus `RegenerationRiderSpan`, both copied verbatim onto
the compiled effect for lowering to read. Loyalty cost components compile the signed activation amount into typed
`AmountValue`/`AmountKnown` (or `AmountFromX` for variable `X`) so lowering reads
the loyalty change as data rather than re-parsing the `+`/`−`/`-` sign text. The
adapter also copies the parser's typed source-spell cost-reduction fields
(`SourceSpellCostReduction` / `SourceSpellCostReductionAmount`) onto
`CompiledEffect` verbatim, so cardgen lowering builds the source-scoped cost
modifier from typed data without re-reading "This spell costs {N} less …" text.
The parser's whole-graveyard exile recognition arrives the same way: the typed
`GraveyardZoneExile` kind (`TargetPlayer`/`TargetOpponent` for "Exile target
player's/opponent's graveyard.") is copied verbatim onto `CompiledEffect`, so
lowering emits the player-zone group `MoveCard` from data instead of re-reading
the "target player's graveyard" object phrase.
The parser's `HandLibraryPut` marker is copied onto `CompiledEffect` the same
way. Combined-sequence lowering pairs it with a preceding typed draw and never
re-reads the retained "from your hand ... in any order" text.
The parser's `HandDiscard` marker follows the same text-blind path for exact
fixed-cardinality controller discards, allowing draw-then-discard lowering
without downstream Oracle-text inspection.
`EffectReorderLibraryTop` and the exact optional controller `EffectShuffle`
likewise cross the compiler as typed kinds, amounts, optionality, references,
and source spans. The compiler mechanically copies those fields; Ponder
sequence recognition belongs to lowering and never re-reads retained wording.

Optional "you may" abilities, mana-symbol cost components, and the remaining
reference/selection forms likewise arrive as typed parser syntax. Optionality is
the parser's `Ability.Optional` flag with its source span; the cost grammar is
the parser's typed `Cost`/`CostComponent` list (including mana-symbol components,
the "from your graveyard" source zone, the `SourceSelf` self-reference, and the
`ExcludeSource` "another" flag); the compiler reads them as data and
never inspects `{T}`/`{Q}`/`{E}` spelling or "you may" tokens.
Embedded effect payments are copied mechanically with their typed payer and
mana cost. Both `unless that player pays` and the parser-distinguished `that
player may pay. If the player doesn't` form use the closed
event-player-does-not-pay predicate, while retaining enough typed form and
condition identity to preserve optional versus mandatory consequences. The
accompanying `that player` reference binds to the triggering event actor for
authoritative player events such as spell casts and card draws.

The compiler performs no semantic interpretation of Oracle source text or
tokens. It consumes parser syntax and reusable source-spanned atoms mechanically;
retained `.Text` survives only as rendering/diagnostic metadata and for exact
source-span accounting. The compiler also no longer ingests raw `[]shared.Token`
streams: keyword and reference recognition arrives as the parser's
`SemanticKeywords`/`SemanticReferences` accessors, condition segmentation as the
parser's `ConditionSegments`/`TriggerConditionSegments` (replacing punctuation
scanning), the body content span as `ContentSpan`, and rendered reference and
condition strings as parser-emitted `Text` (replacing token rejoining). The
compiler-facing AST exposes no `parser.Phrase`: the ability word is a typed
`AbilityWordClause`, the trigger event a rendered string plus span, and cost
presence the typed `CostSyntax()`. `shared.Token` no longer appears in compiler
semantics or rendering. This boundary is enforced automatically: the
`TestCompilerIsTextBlind` AST analyzer in package `cardgen`
(`text_blindness_enforcement_test.go`) fails if any non-test file in this package
applies a string-inspection operation (a `strings` predicate/search/split call, a
comparison against a string literal, a switch on Oracle wording, `regexp`, or
`shared.NormalizedWords`) to a value that flows from a `.Text`/`.Event` field. The
compiler allowlist is empty.

The compiler also performs no positional reasoning over source-span byte
offsets: it never derives node identity, containment, or ordering from raw
positions. The parser emits those as typed relationships that the compiler
consumes mechanically. Node identity arrives as stable parser-assigned `NodeID`
values (references, condition boundaries, and a condition's source-subject
reference), so the compiler matches "the same node" by identity instead of
comparing spans. Ordering and containment arrive as dense per-ability
source-order ranks (`shared.SourceOrder`, carried by the `Order`/`VerbOrder`
fields the parser stamps and the compiler copies onto its IR): reference binding
compares these ranks to order references against effects, targets, the trigger,
and one another, and `SourceOrder.Contains` replaces span-offset containment for
the cost/component, effect, and condition membership tests. Spans survive into
the compiler only as pass-through values for diagnostics
(`unsupportedDiagnostic`) and for lowering's retained-text rendering and
source-consumption accounting; the compiler never reads a span's `.Offset` and
never compares spans for identity. This boundary is enforced automatically: the
`TestCompilerIsPositionBlind` AST analyzer in package `cardgen`
(`position_blindness_enforcement_test.go`) fails if any non-test file in this
package reads a span boundary's byte `Offset` or compares a `*Span` field for
equality. The compiler allowlist is empty.

`compiler` imports `parser` and `shared`. Cardgen lowering consumes compiler IR;
retained source metadata remains available for diagnostics and strict source
consumption checks.
