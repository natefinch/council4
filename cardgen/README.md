# cardgen

Package `cardgen` is the isolated home for **Card Generation** tooling. It turns
Scryfall bulk data and Oracle text into executable `game.CardDef` Go source for
the Card Registry. Runtime game, rules, registry, and simulation behavior live
outside this directory.

There is one generation path:

```text
Scryfall JSON
  -> Oracle recognition
  -> typed game values
  -> CardDef validation
  -> deterministic Go source
```

The compiler is fail-closed. It emits a card only when every face, ability,
semantic element, and meaningful source token is supported. Trigger wording is
recognized here into a source-spanned `oracle.TriggerPattern` with closed
semantic event, relation, Selection, zone, step, combat, batching, and
intervening-condition vocabulary. Exact condition wording is recognized once
into a closed, source-spanned semantic predicate and exact object wording is
bound conservatively to its source, target occurrence, triggering event
subject, or prior instruction result. Ambiguous and unsupported references
remain explicit semantic values. The retained raw text is used only for
diagnostics and exact source consumption. Unsupported cards
receive source-spanned diagnostics; `cardgen` never emits TODOs, partial ability
data, or guessed behavior.

Trigger recognition belongs to the Oracle parser. Its composable grammar emits
source-spanned typed syntax for permanent zone-change, spell/ability, combat,
damage, phase/step, permanent-state, counter, sacrifice, mutate, targeting, and
player events. The semantic compiler and cardgen lowering mechanically map only
those closed values; ambiguous, partial, or unsupported event grammar fails
closed and retained event text is diagnostic metadata only.

Before compilation, `CorpusPolicy` limits the working corpus to cards that are
legal, restricted, or banned in Standard, Pioneer, Modern, Legacy, Pauper,
Vintage, or Commander. Playable paper token definitions are retained as a
special exception. Alchemy, digital-only identities, memorabilia, illegal
Un-set cards, minigames, art-series records, emblems, planes, schemes, and
Vanguard cards are excluded with explicit report reasons.

## Compiler stages

1. **Recognition (`cardgen/oracle`).** The lexer and parser preserve exact source
   spans. The parser recognizes resolving effects, targets, selections, amounts,
   durations, zones, embedded effect payments, keywords, references, and every
   supported trigger-event family; the semantic compiler mechanically maps that
   syntax and recognizes remaining shell and declaration families
   conservatively. Reusable
   body content (targets, conditions, effects, keywords, references, nested modes)
   is grouped into `oracle.AbilityContent`; each `oracle.CompiledAbility` and
   `oracle.CompiledMode` carries one `oracle.AbilityContent` value alongside its
   shell-specific fields (cost, trigger clause, loyalty change, chapter numbers,
   text, span, optional flag). Static wording is recognized separately into one
   or more source-spanned `oracle.StaticDeclaration` values because declarations
   never resolve and are not Instructions. A declaration carries a closed group
   domain plus Selection, optional shared condition, and a typed continuous
   layer operation, rule domain and operation, cost modifier, or non-battlefield
   card-ability grant. Unsupported groups, conditions, durations, operations,
   and shells remain explicit capability blockers.
2. **Typed lowering (`lower.go`, `activation.go`, `static_declaration.go`, `condition.go`,
   `reference.go`, `trigger_pattern.go`, and `executable.go`).**
   `lowerTriggerPattern` is the single mechanical adapter from
   `oracle.TriggerPattern` to `game.TriggerPattern`; trigger shell lowerers never
   interpret raw event-clause text. `lowerAbilityContent`
   is the single entry point that lowers an `oracle.AbilityContent` value into
   `game.AbilityContent`. All supported shells — spell, activated body, triggered
   body, loyalty body, chapter body, modal option, and ordered-effect clauses —
   call `lowerAbilityContent` directly; no shell lowerer constructs a fake spell
   ability to reach body lowering. `condition.go` is the single
   `oracle.CompiledCondition` to `game.Condition` adapter and requires an
   explicit static, activation, replacement, or intervening-trigger context.
   `reference.go` is the single adapter from bound semantic references to typed
   runtime object and card references, including event-permanent LKI and linked
   prior-instruction results. Ordered lowering also supports the exact linked
   shuffle/reveal/permanent-hit sequence: shuffle one targeted permanent into its
   owner's library, reveal that owner's top card, then put the same linked card
   onto the battlefield under that owner's control only when it is a permanent
   card. The parser supplies the actor, card source, and card-type condition;
   lowering never re-reads Oracle text, and optional, multi-card, different-actor,
   or different-filter variants fail closed. `activation.go` composes the generic activated
   shell from typed cost components, timing, zone of function, activation
   condition, bound references, and shared Ability Content. Sentence-leading
   `Then if` conditions are parser-classified as resolving conditions rather
   than activation restrictions. The bounded linked-search rider lowers one
   land searched from library onto the battlefield tapped, publishes that
   permanent, then conditionally untaps only that permanent when the controller
   has the typed at-least-N land count after the search and shuffle. Other
   timings, counts, zones, types, targets, or qualified selectors fail closed.
   Mana and non-mana
   activated abilities use that same shell preparation while retaining distinct
   runtime types. Known shell failures report activation cost, timing, zone,
   condition, reference, mode, or structure diagnostics instead of a generic
   activated-ability failure. `static_declaration.go` is the single mechanical
   adapter from semantic Static Declarations to `game.StaticAbility`,
   `game.ContinuousEffect`, `game.RuleEffect`, and `game.CostModifier` values.
   Mixed static paragraphs lower through that adapter as multiple declarations
   sharing one runtime static ability; the polymorph lose-abilities-become family
   lowers to one `game.ContinuousEffect` per layer (a `RemoveAllAbilities`
   ability-layer effect plus set-color, set-type/subtype, and base power/toughness
   effects on the attached object). A source-spell cost-reduction special case
   lowers the exact ability "This spell costs {N} less to cast for each
   &lt;countable battlefield object&gt;." (e.g. Blasphemous Act) ahead of the
   generic spell/static lowering: it emits a source-scoped
   `game.StaticAbility` whose `RuleEffect` is an `AffectedSource`
   `game.CostModifier` carrying `PerObjectReduction` and a `CountSelection`
   built from the same typed count machinery used elsewhere, so the rest of the
   card (such as the 13-damage body) lowers normally. Card-zone counts,
   variable `{X}` reductions, and other unmodeled shapes stay fail-closed.
   Exact counter-then-next-turn-upkeep draw sequences lower from typed effects
   into an immediate counter plus independent one-shot delayed triggers.
   Target-controller draws retain the targeted stack object's controller
   reference, while `up to N` draws use a bounded numeric resolution choice;
   lowering does not inspect Oracle text or gate those triggers on whether the
   counter instruction succeeded.
   Recognized semantics
   become typed `game.*` ability values, including chapter-numbered
   `game.ChapterAbility` values and the `game.ReadAheadStaticBody` Saga keyword
   template. `assembleCardDefs` combines
   those values with printed Scryfall fields and calls
   [`game.ValidateCardDef`](../mtg/game/README.md#carddef-structural-validation).
   Keyword identity, keyword-selector identity, and keyword parameters arrive
   from parser-owned typed syntax. Lowering maps typed keyword kinds to runtime
   templates and consumes already-parsed mana costs, integers, Enchant targets,
   and Protection predicates; it never parses keyword names or parameter text.
   Multi-keyword lines whose keywords are separated by semicolons (e.g. older
   `First strike; reach` wording) lower like their comma-separated equivalents:
   the keyword-only coverage gate credits the semicolon separator token the same
   way the parser already drops list commas, while every keyword word remains
   must-cover so a line mixing supported and unmodeled keywords still fails closed.
   Parameterized Kicker, Madness, Morph, Disguise, Mutate, and Toxic lines lower
   into their corresponding sealed `game.KeywordAbility` values; unsupported
   parameter forms remain fail-closed. Exact "Whenever this creature mutates"
   triggers lower to `game.EventPermanentMutated`. Exact static power/toughness bonuses may
   also grant supported keywords through separate layer-6 and layer-7
   continuous effects. Standalone keyword grants to supported controlled,
   creature-subtype-filtered, and attached permanent groups lower to layer-6
   continuous effects. Until-end-of-turn keyword-grant spells over controlled
   permanents or a controlled, attacking, blocking, or all-creatures group
   (`Permanents you control gain hexproof and indestructible until end of turn.`,
   including multiple keywords joined by `and`)
   lower through `lowerGroupTemporaryKeywordSpell` into a `game.LayerAbility`
   `AddKeywords` continuous effect over the group for
   `game.DurationUntilEndOfTurn`. Resolution snapshots the matching permanent
   object IDs, so later entrants do not gain the keywords; color-filtered groups,
   opponent-permanent groups, parameterized keywords, and quoted granted
   abilities remain fail-closed. Static power/toughness and keyword group anthems
   also
   cover battlefield-wide creature groups ("All/Other creatures"), combat-state
   groups ("Attacking/Blocking creatures" and "Attacking creatures you control"),
   and battlefield creature-subtype groups ("All/Other <Subtype> creatures"),
   each lowering to a `BattlefieldGroup`/`ObjectControlledGroup` Selection that
   carries the matching combat state, subtype, or source exclusion. They further
   cover battlefield color creature groups ("White creatures get ...", "Other
   black creatures get ..."), creature-token groups ("Creature tokens [you
   control] get ..."), controlled legendary groups ("Legendary creatures you
   control get ..."), and controlled tapped/untapped groups ("Untapped creatures
   you control get ...", "Other tapped creatures you control have ..."), lowering
   to Selections that carry the matching color, token-only, supertype, or tapped
   predicate. They also cover keyword-filter groups ("Creatures with flying get
   ...", "Creatures you control with flying have ...", "Creatures with flying
   your opponents control get ...") and the keyword-exclusion form ("Creatures
   without flying get ..."), controlled artifact-creature groups ("[Other]
   artifact creatures you control get ..."), and controlled nontoken groups
   ("[Other] nontoken creatures you control get ..."), lowering to Selections that
   carry the matching `Keyword`/`ExcludedKeyword`, conjunctive multi-type, or
   `NonToken` predicate. Excluded-supertype ("Nonlegendary"), color-exclusion
   ("Nonblack"), parametrized-keyword ("Creatures with a flying ability"), granted
   quoted-ability, group rule, and dynamic group anthems remain fail-closed. The static source-tied control grant on control Auras
   ("You control enchanted creature/permanent") lowers to a layer-2 control
   continuous effect over the attached object whose new controller is the Aura's
   controller. A composed single-subject rule operation on the source or its
   attached object lowers to one or more `game.RuleEffect`s, including the
   defender-restricted can't-attack ("can't attack you or planeswalkers you
   control", carrying `DefendingPlayer: game.PlayerYou`) and the single-blocker
   can't-be-blocked (`game.RuleEffectCantBeBlockedByMoreThanOne`) and the
   bounded blocker-restriction can't-be-blocked
   (`game.RuleEffectCantBeBlockedByCreaturesWith`, carrying a
   `game.BlockerRestriction` for "creatures with flying", "... power N or less",
   "... power N or greater", "<color> creatures", and "artifact creatures").
   Source creatures and source permanents also support the exact
   "doesn't untap during your untap step" prohibition as an affected-source
   `game.RuleEffectDoesntUntap`. The fixed
   player-rule static "You have no maximum hand size." lowers to the shared
   `game.NoMaximumHandSizeStaticBody`, carrying a controller-scoped
   `game.RuleEffectNoMaximumHandSize` that suppresses cleanup-step discard. The
   resolving temporary evasion effect "Target creature can't be blocked this
   turn." lowers to a `game.ApplyRule` instruction (`lower_cant_be_blocked.go`)
   that grants the single targeted creature a `game.RuleEffectCantBeBlocked`
   restriction for `game.DurationThisTurn`; the lowerer accepts only the exact
   single-creature-target shape and fails closed on a non-target context, a group
   recipient, any condition, mode, keyword, or reference. Rendering emits the
   `game.PrimitiveApplyRule` instruction (`renderApplyRulePrimitive`) and the
   `game.DurationThisTurn` enum. Exact
   Resolving-effect identity, target cardinality and Selection, amount, duration,
   zones, counters, add-mana output, replacement modifiers, references, and embedded payments arrive from parser-owned
   typed syntax. Target lowering builds runtime predicates from typed selectors
   rather than target wording; retained text is display metadata and diagnostic
   context. Replacement and add-mana lowering likewise consume typed fields rather
   than effect wording. Single-object lowerers require exact one-target
   cardinality, and replacement lowerers reject typed qualifiers they cannot
   represent. Source-relative keyword grants gated by controlling supported permanent
   types, subtypes, colors, colorless, multicolored, or token permanents use
   condition-gated layer-6
   effects. Exact `X` quantities, supported count/life/opponent/source-power
   formulas, parser-owned reusable Oracle atom values, and common target
   restrictions lower into runtime quantities and structured target predicates.
   Targeted graveyard-return/put spells lower a card-zone target through
   `cardSelectionForSelector`, building a `game.Selection` from the typed selector
   for single card types, card-type unions, permanent cards, single colors,
   colorless/multicolored cards, subtypes, and a mana-value bound, then emitting
   one `MoveCard`/`PutOnBattlefield` per target slot so multi-target and "up to N"
   forms lower automatically. A card-type union carries its members as a
   disjunctive `RequiredTypesAny`; the selector Kind's conjunctive single-type
   `RequiredTypes` is dropped whenever a union is present so the predicate keeps
   OR (not AND) semantics. Targeted graveyard-card exile (`Exile target card
   from a graveyard.`, `Exile target creature card from a graveyard.`, `Exile up
   to one target card from your graveyard.`) reuses the same card-zone target
   spec and lowers through `lowerTargetedGraveyardExile` to one
   `MoveCard{FromZone: Graveyard, Destination: Exile}` per target slot; it gates
   on a graveyard `FromZone` and the exact graveyard-card target wording, so the
   shared-graveyard "from a single graveyard" constraint and exile-then-return
   riders stay fail-closed. The whole-graveyard form (`Exile target player's
   graveyard.`, `Exile target opponent's graveyard.`) instead lowers through
   `lowerPlayerGraveyardExile` to the player-zone group form of `MoveCard`
   (`MoveCard{Player: TargetPlayerReference(0), FromZone: Graveyard, Destination:
   Exile}`), which moves every card in the chosen player's graveyard to exile at
   once rather than targeting individual cards; it pairs that primitive with a
   target-player `TargetSpec` (the opponent wording adds a `PlayerOpponent`
   predicate) and the parser's exact graveyard-zone-exile reconstruction, so
   "that/each player's graveyard", "all graveyards", "up to N cards", chosen
   cards, multiple graveyards, and exile-until-return riders all stay
   fail-closed. The `MoveCard`/`PutOnBattlefield` graveyard-return
   primitives are rebasable inside an ordered effect sequence (`Return target
   creature card from your graveyard to your hand, then create a token.`): the
   sequence target-rebaser rewrites their `CardReference.TargetIndex` by the count
   of preceding card-kind target specs (`cardTargetSpecsBefore`), which differs
   from the global target-spec offset used for object/player references because
   `CardReference.TargetIndex` is numbered among card targets only. A
   `PutOnBattlefield` carrying entry counters or continuous effects, and the mixed
   inherited-plus-owned-target remap path for these primitives, stay fail-closed.
   The single-target immediate blink sequence (`Exile target <permanent>, then
   return it/that card to the battlefield [tapped] [under its owner's|your
   control] [with a +1/+1 counter on it].`) lowers through
   `lowerImmediateBlinkReturn`: the leading `Exile` clause gains an
   `ExileLinkedKey`, and the `then`-linked return clause emits a
   `game.PutOnBattlefield` from a `LinkedBattlefieldSource`, carrying
   `EntryTapped`, the `your control` recipient, and any named entry counter. It
   gates strictly on the `then` connection (so leading- or trailing-position
   "at the beginning of the next end step" delayed-return wording, which the
   parser leaves untimed, stays fail-closed) and on a single-target exile;
   plural/group blink and color/type-choice returns remain fail-closed.
   The single-target tap-down (stun) sequence (`Tap target <permanent>. <It /
   That permanent> doesn't untap during its controller's next untap step.`,
   Frost Lynx, Take into Custody) lowers through `lowerTapDownSequence` to a
   two-instruction `Mode.Sequence`: a `game.Tap` of the single target followed
   by a `game.SkipNextUntap` of that same `TargetPermanentReference(0)`. The
   `SkipNextUntap` primitive sets the permanent's `Exerted` flag, which the
   untap step consumes by skipping the permanent's next untap. It gates on the
   parser-exact inherited-subject "next untap step" clause, a single-target tap,
   and references that all bind to the tapped target, so the multi-step "next
   two untap steps" window, the open-ended "for as long as you control" and
   "your next untap step" durations all stay fail-closed.
   The multi-target tap-stun sequence (`Tap up to two target creatures. Those
   creatures don't untap during their controller's next untap step.`, Frost
   Breath, Decision Paralysis) lowers through `lowerTapStunSequence`, which
   generalizes the tap-down to the plural "those creatures" prior-subject form
   the parser leaves as an `EffectContextUnknown` stun clause with
   ambiguously-bound anaphora. It emits one `game.Tap` per target slot followed
   by one `game.SkipNextUntap` per slot over a single multi-target permanent
   spec carrying the chosen `0..N` cardinality; the runtime handlers no-op on an
   unfilled "up to" slot. It gates on the parser-exact stun clause, a single
   exact tap whose only target carries the multi-target cardinality, and
   references that all fall within the stun clause and resolve to the tapped
   target, so added clauses, the multi-step "next two untap steps" window (which
   the parser splits into three effects), and the mass "all creatures target
   player controls" form all stay fail-closed.
   The characteristic life-rider sequence (`Exile target creature. Its controller
   gains life equal to its power.`, Swords to Plowshares; `Destroy target
   attacking creature. You gain life equal to its power.`, Chastise; `Exile target
   attacking creature. Its controller gains life equal to its toughness.`, Avenger
   en-Dal; `Destroy target creature or enchantment. You lose life equal to its
   mana value.`, Feed the Swarm; `Put target creature card from a graveyard onto
   the battlefield under your control. You lose life equal to that card's mana
   value.`, Reanimate) lowers through `lowerCharacteristicLifeRider`, the
   per-clause hook in
   `lowerDelayedSequenceClause`. The trailing clause is a life gain or loss whose
   amount is the power, toughness, or mana value of the permanent an earlier clause
   acted on; it
   emits a `game.GainLife`/`game.LoseLife` whose amount is a
   `game.DynamicAmountObjectPower`, `game.DynamicAmountObjectToughness`, or
   `game.DynamicAmountObjectManaValue` over that permanent, read from last-known
   information when the permanent has left the battlefield or from the fresh
   permanent created by a linked graveyard return. Two recipients are modeled:
   the spell's controller (`You gain …`,
   `game.ControllerReference()`) and the acted-on permanent's controller (`Its
   controller gains …`, `game.ObjectControllerReference(TargetPermanentReference)`).
   The amount referent binds either directly to the inherited target (`its power`
   when the recipient took no target binding) or to the prior instruction's result
   (`Its controller gains … its power`, where the recipient already consumed the
   target binding); in the latter case the preceding `game.Exile` is rewritten to
   publish the exiled object under a linked key so the amount reads the exiled
   creature's last-known characteristic through a `LinkedObjectReference`. It gates
   on a single exact, non-negated, non-optional life clause with an `equal to`
   amount of multiplier one and no conditions/keywords/modes. The mana-value amount
   is additionally gated on either a single-target `game.Destroy` of that same
   permanent or an exact single-creature, any-graveyard return to the battlefield
   under the controller's control whose `that card` reference binds to the prior
   result. The return publishes the fresh permanent plus an instruction result;
   the life rider requires success, so an illegal or destination-replaced move
   does not apply it. Gain and lose variants and the parser's exact `Put`/`Return`
   verbs share this path. Auras/noncreatures, multiple cards, hand destinations,
   owner control, fixed or power/toughness riders, and optional or conditional
   variants do not enter this linked category.
   Mass return-to-hand spells (`Return all <group> to their owners' hands.`,
   including the `you control` self-control variant) lower to a single
   `game.Bounce` over a `BattlefieldGroup` Selection built by the shared
   `massGroupSelection`, mirroring mass destroy/exile; the only tolerated
   reference is the destination's possessive pronoun, and choice-based color
   filters, `except for` riders, and `all but one` stay fail-closed.
   The single-choose `Return a/an/another <permanent> you control to its owner's
   hand.` form (no target — the parser records the choosable group on the
   effect's selector) lowers through `lowerControlledBounceSpell` to a
   `game.Bounce{ControlledChoice: true, Amount: game.Fixed(1)}` whose `Group`
   is the `you control` candidate pool (with `ExcludeSource` for `another`); the
   resolving controller chooses one matching permanent at resolution. It accepts
   the destination possessive pronoun under any binding (so triggered "When this
   creature enters" bodies work) and stays fail-closed without the `you control`
   relation, for `each`, and for excluded-type predicates.
   The self form `Return <subject> to its owner's hand.`, where the subject is the
   source permanent itself named as `this <object>` or by the card's own name
   (`Return Selenia to its owner's hand.`), lowers through `lowerFixedBounceSpell`
   to a `game.Bounce{Object: game.SourcePermanentReference()}`; both naming forms
   bind to the source, so the runtime returns the permanent that activated the
   ability.
   Targeted battlefield bounce reuses the shared multi-target permanent
   machinery: the single-target `Return target <permanent> to its owner's hand.`
   form lowers one `game.Bounce` per slot through `lowerFixedBounceSpell`, while
   plural (`Return two target creatures to their owners' hands.`), optional-plural
   (`Return up to two target nonland permanents to their owners' hands.`), and
   optional-singular (`Return up to one (other) target <permanent> [you control]
   to its owner's hand.`) forms lower through `lowerMultiTargetBounceSpell`. The
   target predicate carries the cardinality range plus any excluded card type
   (`nonland permanent`), `other` self-exclusion, or controller clause; the
   tolerated reference is the destination's possessive pronoun (`their` for the
   plural form, `its` for the optional-singular form), and declined "up to" slots
   no-op on their unresolved target index.
   Ordered effect clauses retain parser-owned independent target, reference,
   grammatical-subject, and clause ownership; lowering clips diagnostic syntax
   to those spans rather than rediscovering ownership from Oracle wording.
   A clause may lower to more than one runtime instruction — "up to two target
   creatures each get +1/+2" expands to one `game.ModifyPT` per target slot and
   "Add {R}{R}" expands to one `game.AddMana` per pip — so the sequence lowerer
   only requires each clause to contribute at least one instruction (an empty
   lowering fails closed as a silent drop) and proves completeness through the
   exact consumed-target/reference/keyword/condition counts rather than a
   one-instruction-per-effect tally (`Tandem Tactics`, `Calamitous Tide`,
   `Seismic Spike`).
   Exact fixed, `X`, and supported dynamic placement of recognized named
   counters lowers from supported spell, activated, loyalty, triggered,
   ordered-effect, and Saga chapter bodies into typed `game.AddCounter`
   permanent instructions or `game.AddPlayerCounter` instructions for poison,
   energy, and experience. The placement object may be a single target, every
   permanent in a filtered battlefield group (`Put a +1/+1 counter on each
   creature you control.`, including keyword-granting counters such as
   `Put a deathtouch counter on each creature you control.` whose lowering
   tolerates the parser's benign keyword artifact for the counter name), each of several targets for the multi-target
   `each of up to N target <permanent>s` form (lowered to one `game.AddCounter`
   per target slot, mirroring multi-target graveyard return, with optional
   `other` self-exclusion and controller clause), an optional single target for
   the `up to one target <permanent>` form (lowered to one optional `game.AddCounter`
   slot that no-ops when the target is declined), the
   source permanent itself for fixed self-placement bodies
   (`Put a +1/+1 counter on this creature.`, lowered to
   `game.SourcePermanentReference()`), the permanent an Aura source is attached
   to for fixed `enchanted creature` placement bodies
   (`At the beginning of your upkeep, put a +1/+1 counter on enchanted creature.`,
   lowered to `game.SourceAttachedPermanentReference()`), or a prior clause's target referenced by
   "it" in an ordered sequence (`… Put a +1/+1 counter on it.`). Counter kinds and target domains are checked
   strictly. Distribution (`among`) and dynamic per-target amounts on multi-target
   placements remain fail-closed. Stun and finality placement remain fail-closed until their
   mandatory runtime mechanics are
   implemented ([#222](https://github.com/natefinch/council4/issues/222),
   [#223](https://github.com/natefinch/council4/issues/223)). Self-enter triggers support exact intervening
   conditions for kicked or cast entry and controlling one
   permanent of a named permanent card type. Permanent zone-change triggers
   share one lowering path for self, attached, single-subject, and `one or more`
   enter, die, leave, exile, return-to-hand, and battlefield-to-graveyard
   clauses. Exact patterns may bind controller and owner relations, origin and
   destination zones, self exclusion, face-down state, and event-subject
   Selection predicates for type unions, supertypes, subtypes (including
   Outlaw), colors, token state, tapped state, combat state, keywords, mana
   value, power, and toughness. `Leaves ... without dying` excludes the
   graveyard destination. A cosmetic ability-word label (e.g. `Chainsword —`)
   no longer blocks lowering of a die, leave, or other non-enter zone-change
   trigger body; ability words carry no rules meaning (CR 207.2c) and are
   excluded from the lowered body span, matching enter-trigger behavior. Exact fixed until-end-of-turn power/toughness
   changes to the triggering permanent (`It gets +X/+Y until end of turn.`)
   lower through the shared `lowerFixedModifyPTSpell` path when the sole
   non-target subject reference is `ReferenceBindingEventPermanent`; the
   object lowers via `lowerObjectReference` to `game.EventPermanentReference()`
   and is available in every saturated trigger shell, not only zone-change
   triggers. The same path lowers exact until-end-of-turn self-pump
   bodies (`This creature gets +X/+Y until end of turn.`) when the sole subject
   reference is `ReferenceBindingSource`, and inherited-target pump bodies
   (`… It gets +X/+Y until end of turn.`) when the sole reference binds to a
   prior clause's target; the object lowers to
   `game.SourcePermanentReference()` or a target reference accordingly. These
   source and event-permanent subjects also carry dynamic count amounts whose
   `where X is the number of …` / `for each …` machinery is already supported,
   so self and triggering-permanent pumps that scale with a battlefield count
   (`This creature gets +X/+X until end of turn, where X is the number of
   artifacts you control.`, `… it gets +1/+1 … for each enchantment you
   control.`) lower through `referencedModifyPTQuantities`, computing each side's
   fixed or dynamic delta independently. Pumps scaled by a permanent's own power
   (`where X is its power` / `… this creature's power` / `… <name>'s power`)
   lower through `lowerSourcePowerModifyPTSpell`, which binds the power referent
   (the reference whose span matches the amount's referent span) to the
   permanent whose power supplies `X` and emits a `game.DynamicAmountObjectPower`
   delta the runtime snapshots at resolution. It covers a single creature target
   (`Target creature gets +X/+0 … where X is its power.`, reading the target's
   power), an activated/triggered pump of that target scaled by the source
   (`… where X is this creature's power.`/`… <name>'s power.`), the source itself
   (`This creature gets +X/+X … where X is its power.`, `EffectContextSource`),
   and the triggering permanent or a prior clause's target addressed by "it"
   (`EffectContextReferencedObject`). Riders, keyword grants, conditions, modes,
   plural or non-creature targets, and any reference set that is not exactly the
   power referent plus the single subject stay fail-closed.
   Dynamic until-end-of-turn pumps whose `where X is …` count machinery is
   already supported lower each side independently, so asymmetric and mixed-sign
   forms (`Target creature gets +X/+0 …`, `… +X/-X …`, `… -X/-X …`) lower
   alongside the symmetric `+X/+X` form. Exact until-end-of-turn pumps on a
   single target slot also lower through `lowerFixedModifyPTTargets`, which reuses
   the shared `permanentTargetSpecWithCardinality` and emits one `ModifyPT` per
   target slot: plural (`Two target creatures each get -1/-1 until end of turn.`),
   optional (`Up to one/two target creatures … gets/each get …`), and creature-
   subtype (`Target Human you control gets +2/+2 …`) targets are supported, with
   declined "up to" slots no-opping on their unresolved target index. Each power/
   toughness side may be a fixed signed amount or the spell's variable `X`
   (`Target creature gets +X/+0 until end of turn.`, `… -X/-X …`, `… -X/+X …`)
   when `X` comes from an `{X}` mana cost or an `AmountFromX` additional/activation
   cost; the variable side lowers to the runtime `DynamicAmountX` (negated for a
   `-X` side) and snapshots the chosen X at resolution. Non-creature pump targets,
   rules-derived dynamic multi-target amounts, and riders stay fail-closed.
   Exact until-end-of-turn combined buffs that pump and grant keywords across one
   or more target slots (`Up to two target creatures each get +1/+1 and gain
   trample until end of turn.`, `Two target creatures each get +2/+2 and gain
   first strike until end of turn.`) lower through `lowerTemporaryPTKeywordSpell`,
   which reuses `permanentTargetSpecWithCardinality` and emits one
   `game.ApplyContinuous` per target slot carrying both the layer-7 power/toughness
   delta and the layer-6 `AddKeywords` grant; single-target output stays
   byte-identical to the prior single-slot form. Color-filtered or quoted-ability
   grants and dynamic amounts remain fail-closed. Exact fixed and dynamic damage bodies whose damage source
   reference is `ReferenceBindingEventPermanent` also lower through shared
   `lowerFixedDamageSpell` and `lowerGroupDamageSpell` paths; the `It deals`
   pronoun form is accepted alongside the card-name form when the source
   binding is `ReferenceBindingEventPermanent`, and `DamageSource` is
   preserved as `game.EventPermanentReference()` for LKI. The self form
   (`This creature deals N damage ...`, `ReferenceThisObject` bound to
   `ReferenceBindingSource`) is also accepted; its `DamageSource` is left
   default, which the runtime resolves to the ability's source permanent.
   Single-target damage recipients additionally accept exact keyword-qualified
   (`target creature with flying`), multi-color (`target white or blue
   creature`), and combined `target player or planeswalker` / `target opponent
   or planeswalker` wordings; the player-or-planeswalker forms lower to a
   target spec allowing a player or a planeswalker permanent, with the opponent
   variant restricting the player half to opponents. Group damage recipients
   (`each creature with flying`) accept a runtime-modelable keyword
   filter on the recipient Selection; keywords the runtime cannot model as a
   selector predicate stay fail-closed. Group damage amounts may be an exact
   fixed value or the spell's `X` (`Earthquake deals X damage to each creature
   without flying and each player.`), each dealt to every member of every
   recipient group; the dynamic `equal to …` and dual-recipient `where X is …`
   forms stay fail-closed, but a single-recipient `where X is the number of …`
   count amount is supported (`Gates Ablaze deals X damage to each creature,
   where X is the number of Gates you control.`, `Chain Reaction deals X damage
   to each creature, where X is the number of creatures on the battlefield.`):
   the recipient phrase is scoped to the tokens before the trailing count clause
   so the count subject's filters bind to the count selector rather than the
   recipient, and the battlefield count is resolved once through the shared
   dynamic-amount lowerer and dealt to every recipient. Count kinds with no
   battlefield selector (e.g. basic land types) and two-recipient count damage
   stay fail-closed.
   A damage recipient that is the controller or owner of the prior removal target
   in an ordered sequence (`Destroy target land. <name> deals 2 damage to that
   land's controller.`, `… deals N damage to its owner.`) lowers through
   `lowerReferencedPlayerDamageSpell` to a `game.PlayerDamageRecipient` wrapping
   `game.ObjectControllerReference`/`game.ObjectOwnerReference` of the inherited
   target (a `TargetPermanentReference` for a permanent antecedent, a
   `TargetStackObjectReference` for a countered spell); the damage carries
   `game.SourcePermanentReference()` only when the source is a permanent
   ("This creature deals …"), and only fixed or `X` amounts are accepted.
   A damage recipient that is the controller of the resolving spell or ability—a
   lone "you" (`<name> deals N damage to you.`)—lowers through
   `lowerControllerDamageSpell` to a `game.Damage` whose recipient is
   `game.PlayerDamageRecipient(game.ControllerReference())`, accepting only fixed
   or `X` amounts with no targets, conditions, keywords, or modes. A self-damage
   rider (`<name> deals A damage to <target> and B damage to you.`, Char, Psionic
   Blast) appends a second fixed `game.Damage` instruction aimed at the same
   controller reference after the primary single-target damage, so the chosen
   target takes A and the controller takes B; variable primary amounts and any
   non-"you" second recipient stay fail-closed.
   A target-controller rider (`<name> deals A damage to <target> and B damage to
   that creature's/permanent's controller/owner.` or `... and B damage to its
   controller.`, Chandra's Outrage, First Volley, Unleash Shell) appends a second
   fixed `game.Damage` aimed at the primary target's
   `game.ObjectControllerReference`/`game.ObjectOwnerReference`. A two-target
   rider (`<name> deals A damage to <target0> and B damage to <target1>.`, Hungry
   Flames, Lunge, Punish the Enemy, Reckless Rage) lowers through
   `lowerTwoTargetDamageSpell` to one fixed `game.Damage` per target, addressing
   `game.AnyTargetDamageRecipient(0)` and `game.AnyTargetDamageRecipient(1)` in
   Oracle order. Both rider forms require Known (fixed, `>= 1`) amounts and stay
   fail-closed for variable amounts, dynamic counts, or any condition, keyword, or
   mode; a second clause not introduced by "target" (such as "any target") leaves
   the primary target non-exact and fails closed.
   A source-power damage body in which a target creature deals damage equal to
   its own power (`Target creature deals damage to itself equal to its power.`,
   `Target creature you control deals damage equal to its power to target
   creature you don't control.`) lowers through `lowerSourcePowerDamageSpell` to
   a `game.Damage` whose amount is `game.DynamicAmountObjectPower` of the dealing
   target and whose `DamageSource` is that same `TargetPermanentReference`, so
   the dealing creature's keywords (deathtouch, lifelink) apply. The dealing
   creature is identified by the single `its power` reference occurrence; the
   self form aims the recipient at that same target, the two-target form aims it
   at the other target. The recipient half accepts the `any target`, `another
   target creature`, and `creature or planeswalker you don't control` wordings.
   The mass `Each creature deals damage to itself …` form stays fail-closed.
   An "each of N targets" body (`<name> deals N damage to each of two targets.`,
   `… to each of two target creatures.`, `… to each of up to two target
   creatures.`) lowers through `lowerEachOfTargetsDamageSpell` to one fixed
   `game.Damage` instruction per flat target slot, each addressing its own
   `game.AnyTargetDamageRecipient` index; declined `up to N` slots simply no-op
   at the unresolved index. This deals the full amount to every chosen target
   (distinct from divided damage, which splits one total). Dynamic amounts,
   divided wording, and riders stay fail-closed.
   A destroy spell carrying a parser-folded regeneration rider ("It/They can't
   be regenerated.") lowers to a `game.Destroy` with `PreventRegeneration: true`,
   for the single-target, multi-target, and mass forms alike; the renderer emits
   the flag explicitly so the generated card bypasses regeneration shields. The
   rider now also folds when a recognized non-destroy sibling effect accompanies
   the lone destroy, so the ordered-effect-sequence lowerer accepts shapes such as
   "Destroy target creature. It can't be regenerated. Its controller creates a 3/3
   green Ape creature token." (Pongify, Rapid Hybridization, Afterlife) and the
   controller-life riders of Crumble and Sever Soul: the destroy emits its
   `PreventRegeneration` instruction and the sibling clause lowers as its own
   sequenced instruction.
   Mass destroy/exile `massGroupSelection` now also carries a bare or card-type-
   qualified subtype filter (`Destroy all Islands.`, `Destroy all Dragon
   creatures.`) as `Selection.SubtypesAny`, allowing a `SelectorUnknown` group
   kind when the subtype set is non-empty so a bare-subtype wipe selects any
   permanent of that subtype; an untapped group (`Destroy all untapped
   creatures.`) sets `Selection.Tapped = TriFalse`; and a non-creature numeric
   mass (`Destroy all nonland permanents with mana value N or less.`) carries the
   excluded card type plus the mana-value bound. A single permanent target can
   carry the same mana-value bound on a card-type union (`Destroy target creature
   or planeswalker with mana value N or less.`), which lowers through the shared
   `permanentTargetSpecWithCardinality` to a `TargetPredicate` whose disjunctive
   `PermanentTypes` and `ManaValue` are honored by the runtime.
   Mass tap and untap reuse the same `massGroupSelection` machinery: `Tap all
   <group>.` and `Untap all <group>.` (for example `Tap all creatures your
   opponents control.`, `Untap all creatures you control.`) lower to a
   `game.Tap{Group}` or `game.Untap{Group}` rather than a single-object tap.
   The exact Frantic Search clause `Untap up to three lands.` lowers to
   `game.Untap{Group, ChooseUpTo: true, Amount: game.Fixed(3)}`. The rules engine
   makes that distinct zero-to-three land choice during resolution, after earlier
   draw and discard instructions; other groups, controller qualifiers, random
   selection, and counts remain fail-closed.
   The `game.Tap` primitive carries an optional `Group` alongside its `Object`
   (exactly one is set), mirroring `game.Untap`; the rules engine taps or untaps
   every permanent the group matches, honoring controller, subtype, color, and
   `ExcludeSource` ("all other creatures") filters just as mass destroy does.
   Exact destroy,
   exile, tap, untap, bounce-to-owner's-hand, and sacrifice bodies whose
   sole subject reference is `ReferenceBindingEventPermanent` (the triggering
   permanent) or `ReferenceBindingTarget` (a prior clause's target referenced by
   "it" in an ordered sequence) lower through
   the shared `lowerReferencedPronounEffect` path using exact "it"
   pronoun forms and the `ReferenceThatObject` demonstrative ("that creature"/
   "that permanent") that binds the same prior-clause target; both gate on
   no-target, no-negation, and exact wording. This covers ordered buff
   sequences whose trailing clause refers back to the buffed target by
   demonstrative (`Target creature gets +3/+3 until end of turn. Untap that
   creature.`, `… Tap that creature.`). Exact fixed-count draw, discard, and mill bodies whose
   sole subject reference is `ReferenceBindingEventPlayer` lower through the
   shared event-player draw/discard/mill paths using exact "they" pronoun forms
   or a spell-cast body's exact `that player`, resolving the player via
   `game.EventPlayerReference()`. The same
   draw/discard/mill paths additionally lower group recipients: an `Each player`
   (`EffectContextEachPlayer`) or `Each opponent` (`EffectContextEachOpponent`)
   subject with no targets or references lowers to a `PlayerGroup`
   (`game.AllPlayersReference()`/`game.OpponentsReference()`) on the
   `game.Draw`/`game.Discard`/`game.Mill` primitive, mirroring the group
   life-change recipients; a `Target opponent` recipient lowers like
   `Target player` through `playerTargetSpec`. Exact
   source-bound `Sacrifice it.` with `ReferenceBindingSource` or
   `ReferenceBindingEventPermanent` and no targets lowers to a
   `game.Sacrifice` primitive using `lowerObjectReference` in the
   `lowerSacrificeSpell` path. Phase and step triggered abilities
   using `At the beginning of …` lower for
   exact supported controller-relative upkeep, draw, end, combat, combat-step,
   and main-phase variants, including steps belonging to the controller of an
   enchanted permanent. Typed combat-event syntax binds named/self/attached and semantic
   Selection subjects, the other blocking combatant, attacked player or
   permanent recipients, damage-source and damage-recipient Selections,
   combat/noncombat qualifiers, and exact player relations. Player-level attack
   wording and `one or more` attack, block, and combat-damage wording lower only
   through declaration/damage batch IDs, with per-attack-target batching where
   Oracle semantics require it. Compound events, temporal qualifiers, and
   unavailable Selection predicates remain fail-closed with missing-event or
   missing-runtime-capability diagnostics. Exact
   permanent-tapped, permanent-untapped, and turned-face-up action triggers
   share the semantic Trigger Pattern path; face-up triggers may bind self,
   attached, controller-relative, and Selection-filtered subjects. Became-target
   patterns bind the targeted subject's controller independently from the
   targeting spell or ability's controller. Typed player-action syntax includes
   controller-relative and any-player Cycling events. Sacrifice triggers bind
   the sacrificing player independently from the sacrificed permanent's shared
   Selection subject. Discard triggers may additionally filter the discarded
   card by type, lowering `Whenever you discard a creature card`, `... a land
   card`, `... a nonland card`, and `... a noncreature, nonland card` forms (and
   their `one or more` and opponent variants) into a card-type `CardSelection`
   matched against the discarded card's types. Scry and surveil use distinct player-action Trigger
   Pattern events. Activated-ability patterns bind the activating player and
   source-permanent Selection, but lower only when they explicitly exclude mana
   abilities; unrestricted forms fail closed until payment-time mana
   activations join the authoritative event stream.
   Supported draw, life-gain/loss, scry, and surveil patterns may also bind the
   affected player's exact event ordinal during the current turn.
   self-dies triggers support exact `if it had no +1/+1 counters` and
   `if it had no -1/-1 counters` conditions using the departed permanent's
   last-known information. Fixed-damage bodies preserve that permanent as the
   damage source through an event reference. Exact event-card references can
   return the departed card from its owner's graveyard to hand or grant its
   Adventure face a graveyard-cast permission through the end of its
   controller's next turn. Spell-cast triggered abilities using `Whenever ...
   casts ...` lower for three exact player prefixes (`you cast`,
   `a player casts`, `an opponent casts`) and seventeen exact spell phrases:
   `a spell` (wildcard), `a noncreature spell`, `a creature spell`,
   `an instant or sorcery spell`, `an instant spell`/`an instant`,
   `a sorcery spell`, `an artifact spell`, `an enchantment spell`,
   `a land spell`, `a planeswalker spell`, `a noncreature, nonland spell`, and single-color forms
   `a white/blue/black/red/green spell`. Self-cast (`when you cast this spell`),
   `TriggerWhen`, unknown or non-exact ability-word forms, modes, and all other
   spell-phrase forms are fail-closed. Draw, discard, cycling, life-gain/loss,
   damage, spell-cast, and generic-pattern triggers all support recognized
   `lowerCondition`-compatible intervening-if conditions (life threshold,
   controls-permanent selection (including tapped, subtype, power, aggregate
   total-power threshold, and source-exclusion predicates), referenced source/event-permanent existence or
   Selection matching, any-player-life-at-most, opponent-count, graveyard-card
   counts, hand empty, creature-power diversity, and event-history). Referenced
   objects lower through the shared reference adapter; event permanents retain
   current/LKI matching. Parser-typed event-history intervening conditions carry
   a lowered `game.TriggerPattern` plus an `EventHistoryWindow`; the shared
   `lowerTriggerPattern` path ensures consistent filter semantics and runtime
   evaluation reuses `triggerMatchesEvent`. A trailing
   `This ability triggers only once each turn.` (or `twice`) qualifier lowers
   from the parser-owned typed `TriggerFrequency` restriction into
   `game.TriggeredAbility.MaxTriggersPerTurn` without inspecting Oracle wording.
   Recognized phrases: `if you attacked
   this turn`, `if a creature died this turn`, `if you gained life this turn`,
   `if an opponent lost life this turn`, `if you lost life this turn`, `if an
   opponent lost life last turn`, `if you lost life last turn`, and `if no spells
   were cast last turn` (negated). Conditions not in that shared set fail closed
   with a condition diagnostic.
   Exact Threshold, Delirium, Domain, Metalcraft, Hellbent, Ferocious, and Coven
   conditions lower into typed live-state predicates and dynamic amounts.
   Purely cosmetic ability-word labels that carry no rules meaning (for example
   Morbid, Survival, Raid, Revolt, Celebration, Corrupted, Formidable, Lieutenant,
   Enrage, Inspired, Flurry, Opus, Parley) are stripped so the body lowers
   normally; this is safe because such words always restate their game condition
   explicitly in the card's own text (e.g. "Enrage — Whenever this creature is
   dealt damage"). On the non-trigger paths (activated, keyword, and static
   abilities) only the narrow `rulesFreeAbilityWordLabel` whitelist is dropped,
   because there a label printed before an em-dash may instead be a keyword
   ability that carries rules meaning (Boast, Exhaust, Cohort, Renew, ...). On the
   trigger paths every label is dropped via `triggerContentUnsupported`, without
   consulting the whitelist: an ability-word label on a triggered ability always
   precedes a When/Whenever/At trigger word, never an activation cost, so it is
   always genuine rules-free flavor regardless of whether the word is whitelisted.
   A trigger body shaped as an optional resolving sequence ("you may X. If you do,
   Y") lowers through the shared ordered-effect-sequence path: the optional first
   instruction publishes its result and the following instruction gates on it,
   while the rendered `game.TriggeredAbility.Optional` flag stays false because the
   trigger fires unconditionally. A single "if you do" may govern several
   and-joined trailing effects ("you may X. If you do, Y and Z"); each compiles to
   its own effect that structurally contains the gate condition, so every effect
   in the contiguous gated tail is gated on the optional having succeeded. An
   independent later sentence ("… If you do, Y. Z.") does not contain the gate
   condition and would resolve unconditionally, so the whole body fails closed
   rather than gating only part of the tail. A non-optional trigger body that
   carries a resolution condition ("Whenever X, EFFECT. If STATE, EFFECT2." or
   "Whenever X, if STATE, EFFECT.") keeps that condition in the body and routes it
   through the shared content lowering exactly as the same condition lowers on a
   spell; the body span widens to cover the condition clause whether it precedes
   or follows the effects, and the shared lowering fails closed if the condition
   itself is unrepresentable. Any other trigger body whose conditions are not the
   intervening-if condition, this optional-flow gate, or a resolution condition
   fails closed.
   A targetless trigger body shaped as `you may <supported controller benefit>
   unless that player pays <fixed mana>` lowers to a `game.Pay` instruction
   whose payer is `game.EventPlayerReference()`, followed by the optional
   benefit gated on payment failure. This preserves the required choice order:
   the event actor decides whether to pay first, then the trigger controller
   decides whether to take the benefit only when payment was declined or
   impossible. The Oracle-distinct `that player may pay <fixed mana>. If the
   player doesn't, <supported controller benefit>` form uses the same typed
   payment-result envelope but keeps the failure consequence mandatory
   (Smothering Tithe). The reusable envelope accepts only a single exact
   controller benefit that already lowers to one targetless instruction.
   The controller-owned success form `you may pay <fixed mana>. If you do,
   <effect>` likewise lowers to `game.Pay`, followed by one source-relative,
   targetless instruction gated on successful payment (Mana Vault's paid upkeep
   untap). The trigger itself remains mandatory, and the resolving stack object's
   captured controller makes the payment choice. Variable, nonmana,
   targeted/group-recipient, multi-effect,
   replacement, frequency-qualified, non-trigger, static-tax, and
   cumulative-upkeep forms remain fail-closed.
   A controller optional whose body is the causative "you may have <subject>
   <action>" ("you may have this creature deal 1 damage to each creature", "you
   may have it deal 1 damage to any target") lowers through
   `lowerOptionalHaveEffect`: the parser models "have"/"has" as a leading
   structural `EffectGrantKeyword` carrying the resolving optionality while the
   real action (deal damage, …) compiles as a second effect sharing the same
   sentence span, so the lowerer drops the empty "have" effect, lowers the action
   through the normal single-effect path, and marks that one instruction
   `Optional`. It fails closed unless the causative "have" belongs to the ability
   controller (`EffectContextController`), keeping the non-controller "<player>
   may have <subject> <action>" shape ("that creature's controller may have it
   deal …") unsupported, and any action the single-effect path cannot lower (for
   example "have each opponent discard a card") likewise stays unsupported.
   Ordinary battlefield activations
   lower exact mana, tap, untap, sacrifice, discard, pay-life, source-exile,
   graveyard-exile, and source-counter-removal costs into typed payment data.
   Sacrifice costs recognize a subtype, a subtype with its permanent-type noun
   ("Sacrifice a Goblin creature", "Sacrifice two Blood tokens"), an explicit
   count, the source itself ("Sacrifice this <subtype>"), "another" (an
   exclude-source sacrifice), a counted "other" that also excludes the source
   ("Sacrifice two other creatures"), a two-type union of permanent types joined
   by "or" or "and/or", with an optional article before the second type
   ("Sacrifice an artifact or creature", "Sacrifice another creature or an
   enchantment"), and a two-subtype union ("Sacrifice a Forest or Plains",
   "Sacrifice another Orc or Goblin") lowered into `SubtypesAny`.
   Tap-permanents costs ("Tap two untapped artifacts and/or creatures you
   control") lower a count plus an object that is a permanent type, a subtype from
   any permanent family (including land subtypes such as "Gate" or "Desert"), or a
   two-type union, all requiring untapped, you-control permanents.
   "Exile this card from your graveyard" lowers to a graveyard source-exile.
   A spell's leading additional cost ("As an additional cost to cast this spell,
   <cost>.") lowers through the same shared cost machinery: the parser recognizes
   the fixed prefix as an `AbilitySpellAdditionalCost` paragraph whose cost phrase
   is parsed by `parseCost`, and cardgen emits the recognized components as
   `game.CardFace.AdditionalCosts` while the remaining spell body lowers normally;
   it fails closed for any cost component the shared cost lowering does not yet
   recognize. The prefix is recognized on permanent spells (creatures, artifacts)
   as well as instants and sorceries, so a vanilla creature whose only Oracle text
   is its additional cost (e.g. Makeshift Mauler) still generates. Graveyard-exile
   costs accept any explicit count ("exile a creature card", "exile three cards"),
   a card subtype ("exile an Elf card from your graveyard") lowered into
   `SubtypesAny`, or an `X`-bound count ("exile X cards from your graveyard") that
   resolves against the spell's announced X.
   Exact trailing activation restrictions lower to typed sorcery, combat,
   upkeep, during-your-turn, and once-per-turn timing checks. The
   during-your-turn check (`Activate only during your turn.`) permits activation
   at any time the source's controller is the active player; restrictions tied
   to another player's turn (`Activate only during an opponent's turn.`) fail
   closed. An `Activate only if <event> this
   turn` (or `last turn`) restriction lowers, like the intervening-trigger
   path, into a `game.Condition` event-history predicate that the runtime
   evaluates at activation time against the source's controller; a graveyard
   ability has no battlefield source, so its controller-relative event-history
   restriction fails closed rather than emitting a permanently dead ability.
   Source-state activation restrictions (`Activate only if this creature is
   attacking`/`blocking`/`attacking or blocking` and `Activate only if this
   creature's power is N or greater`) lower to a source-bound `ObjectMatches`
   condition reusing the `game.Selection` combat-state and power filters, and
   `Activate only if an opponent has N or more poison counters`, `Activate only
   if you have exactly N cards in hand`, and `Activate only if you control a
   creature with <keyword>` lower to the matching controller-state and
   controls-with-keyword predicates. Unmodelable variants (e.g. a "blocked"
   combat state or an unrecognized keyword) fail closed.
   Common enters-tapped life, opponent-count, land-count,
   basic-land-subtype, and legendary-creature conditions lower into typed
   replacement predicates.
   Plain self enters-tapped replacements lower from the parser-owned
   `EntersTappedSelf` flag, which recognizes the tapped entry qualifier (for any
   subject noun or card-name phrasing) rather than matching whole Oracle
   sentences. Exact optional pay-2-life and reveal-a-land-or-creature-subtype
   entry wordings lower into typed resolution payments for enters-tapped
   replacements from their typed effect structure. Enters-with-counters
   replacements lower from the typed counter kind and fixed amount: a plain
   `enters with N <kind> counters on it`, a combined `enters tapped with N
   <kind> counters on it` (Vivid land cycle), and a conditional
   `enters with N <kind> counters on it if <condition>` whose predicate is a
   modeled enters-time condition (current-turn event history for Raid, Morbid,
   and opponent-lost-life, or a controlled-permanent count for Ferocious). The
   conditional form threads the entering permanent as the condition source so the
   runtime resolves its event-history predicate at entry time. Dynamic amounts
   (`for each`/X), unknown counter kinds, and unmodeled predicates (e.g. Revolt)
   fail closed. Entry-time choice replacements
   lower from typed parser flags: `EntersColorChoice`/`EntersColorChoiceExclude`
   produce "choose a color[ other than <color>]" replacements (the Gate/Thriving
   land cycle, paired with a fixed-or-chosen composite mana ability), and
   `EntersTypeChoice` produces a "choose a creature type" replacement; both
   record the choice on the permanent for later abilities and fail closed on any
   other entry-choice shape. Modal headers lower from typed
   minimum/maximum mode counts (`Modal.MinModes`/`MaxModes`/`ChoiceKnown`),
   including `Choose one or both`, and loyalty costs lower from the typed signed
   amount (`CostComponent.AmountValue`/`AmountKnown`/`AmountFromX`); neither
   re-reads Oracle wording. Saga lore-counter reminders, Read Ahead recognition
   and its sacrifice chapter, and Devoid recognition are parser-owned typed
   `Ability` flags consumed by lowering.
   Library-search bodies lower to a single `game.Search` primitive from the
   parser-owned exact "Search your library for … , then shuffle." round-trip. The
   supported envelope is a search of your own library for a singular card or an
   "up to N" bounded count, filtered by a plain card type
   (card/land/creature/artifact/enchantment/planeswalker), a `permanent` card
   (optionally with a subtype, e.g. "Rebel permanent"), the `basic` or `legendary`
   supertype, a subtype union with no separate type noun (basic land subtypes like
   "Forest or Island", or other subtypes like "Sliver" and "Aura or Equipment"), or
   a subtype paired with a card type or "permanent" ("Myr creature", "Dragon
   creature", "Rebel permanent"), optionally narrowed by a `with mana value N or
   less` rider (`SearchSpec.MaxManaValue`), moved to hand or the battlefield
   (optionally tapped) and optionally revealed first. A singular card-type union
   such as "artifact or enchantment" lowers through `SearchSpec.CardTypesAny`.
   The exact "then shuffle and put that card on top" family lowers to a
   library destination with `SearchPositionTop`, preserving optional reveal and
   a following fixed controller life-loss rider (Vampiric Tutor, Enlightened
   Tutor). The runtime removes the found card, shuffles the remainder, then
   replaces it on top. Qualified searches may legally fail to find and still
   shuffle; an unrestricted exact "a card" search carries
   `SearchMustFindIfAvailable`, so it must select one when the library is nonempty.
   A split-destination "up to two" tutor ("put one onto the battlefield tapped and the other into your
   hand") lowers to one `game.Search` whose `SearchSpec.SplitDestination` carries
   the secondary single-card slot; the parser records both typed slots on the
   put clause so lowering distributes the found cards across the battlefield and
   hand slots without re-reading text (Cultivate, Kodama's Reach). A correlated
   "up to two" tutor whose found cards "share a land type" lowers to one
   `game.Search` whose `SearchSpec.SharedSubtype` records the constraint; the
   parser owns the "that share a land type" rider (only the two-card basic-land
   shape) so lowering stays text-blind and the runtime forces every found card to
   share a land subtype (Myriad Landscape). The runtime
   treats an "up to" count as a maximum and lets qualified searches legally fail
   to find. An optional tutor ("You may search your library for …") lowers through the
   same exact round-trip — the parser strips the leading "you may" before
   reconstructing the canonical search shape — and marks the single resulting
   `game.Search` instruction `Optional` so the runtime offers the player the choice
   to decline; once an unrestricted exact-card search is accepted, it must find a
   card when the library is nonempty. Graveyard-also searches, other players' libraries, "with different
   names", power/color filters, mana-value bounds other than a fixed "or less"
   (including variable `X` bounds), variable `X` counts, conjunctive or unsupported card-type unions,
   instant/sorcery filters, a split destination on any count other than "up to
   two", a "that share a land type" constraint on any shape other than the
   two-card basic-land tutor, a non-land or non-subtype correlation ("share a
   color", "share a card type"), and
   unsupported destinations remain fail-closed.
   A targeted removal spell that compensates the affected permanent's controller
   with an optional basic-land fetch — the Path to Exile / Assassin's Trophy
   rider ("Exile target creature. Its controller may search their library for a
   basic land card, put it onto the battlefield tapped, then shuffle.") — lowers
   to the removal instruction (single-target `Exile`/`Destroy`) followed by an
   `Optional` `game.Search` whose `Player` and `OptionalActor` both name the
   removal target's controller via
   `ObjectControllerReference(TargetPermanentReference(0))`. The affected player,
   not the spell's controller, decides whether to search (declining skips the
   whole search-and-shuffle) and searches their own library; the searcher
   reference reads the target's controller from last-known information after the
   permanent has left the battlefield. The parser reconstructs the "Its
   controller may search their library for …" clause byte-for-byte against the
   same canonical search envelope as a self-tutor. Any other searcher subject, a
   non-removal leading effect, a missing "may", or a search the envelope cannot
   model remains fail-closed.
   Impulse "dig" bodies lower to a single `game.Dig` primitive from the
   parser-owned two-sentence shape "Look at the top N cards of your library. Put M
   of them into your hand and the rest/the other into your graveyard." Each
   sentence reconstructs byte-exactly: the look clause classifies as `EffectDig`
   and the put clause as a dig-flavored `EffectPut`. The supported envelope is a
   fixed look count of at least two, a fixed take count of one, two, or three that
   is strictly smaller than the look count, and a graveyard remainder; the runtime
   lets the player choose which seen cards go to hand and puts the rest into the
   graveyard. Library-bottom remainders ("on the bottom of your library in any
   order/in a random order") carry unmodeled ordering riders and remain
   fail-closed, as do variable counts and look counts that do not exceed the take
   count.
   Draw-then-put-back bodies such as Brainstorm lower from the parser-owned
   `HandLibraryPut` marker into an ordered `Draw` followed by the selected
   player-zone form of `MoveCard`. The supported envelope is exactly "Draw N
   cards, then put M cards from your hand on top of your library in any order."
   with fixed positive counts. The choice is made after drawing, may include the
   newly drawn cards, requires distinct cards, and uses the returned selection
   order as top-to-bottom library order. Bottom/random/same-order wording,
   opponent hands, variable counts, reveals, and other destinations remain
   fail-closed.
   Draw-then-discard bodies such as Faithless Looting use the parser-owned
   `HandDiscard` marker to lower an exact fixed controller draw followed by an
   exact fixed controller `Discard`. The discard choice sees the post-draw hand,
   requires distinct cards, and discards every available card when fewer than
   requested remain. Targeted/opponent, random, typed-card, and variable-count
   discard forms do not receive this marker. Exact fixed-mana `Flashback` lowers
   to `game.FlashbackKeyword`; variable and non-mana/compound flashback costs
   remain fail-closed.
3. **Rendering (`render.go`).** `Renderer.RenderCardSource` walks only validated
   typed values, derives imports from those values, and emits byte-deterministic,
   gofmt-stable Go source.

The bulk compiler detects distinct Oracle cards that map to the same filename or
Go identifier and appends a stable Scryfall-derived suffix to both generated
identities. Playable tokens always use `mtg/cards/tokens/<letter>` and include
their complete normalized Oracle UUID in both filename and Go identifier.
Printed `CardDef.Name` values remain unchanged.

`mtg/game` owns typed Card Definition data and structural validity;
`mtg/rules` owns behavior; `cardgen` owns recognition, lowering, and rendering.
See [ADR 0008](../docs/adr/0008-typed-ir-lowering.md).

Lowering is text-blind: it consumes the compiler's typed semantics and never
interprets Oracle source text or tokens to derive meaning. Add-mana output is
lowered from the parser's typed `mana.Color` values rather than by re-parsing the
rendered mana-symbol strings, and a fully-parenthesized reminder mana ability is
lowered from the parser's typed inner document (`parser.Ability.ReminderInner`)
rather than by re-parsing the reminder text. The filter-land cycle (Mystic Gate,
Sunken Ruins, and the rest) lowers its `{X/Y}, {T}: Add {X}{X}, {X}{Y}, or
{Y}{Y}.` ability from the typed `FilterPair` flag to `game.TwoColorFilterManaAbility`,
which adds two mana, each independently chosen from the pair's two colors (the
three printed combinations are exactly the unordered two-mana multisets over the
pair). The dynamic lands-produce sources (Exotic Orchard, Reflecting Pool,
Fellwar Stone) lower their `{T}: Add one mana of any color/type that a land you
control / an opponent controls could produce.` ability from the typed
`LandsProduce` flag, scope, and `LandsProduceAnyType` flag to
`game.TapManaLandsProduceAbility`, whose colors (and colorless, for the "any
type" wording) are recomputed from the matching battlefield lands' production at
resolution. A mana ability may carry a single
self-damage rider (`<name> deals N damage to you.`, the painlands, the painland
Talismans, Ancient Tomb, and Tarnished Citadel): the add-mana effect is followed
by one fixed-amount `game.Damage` instruction dealt by the source permanent to
its own controller, recognized from the typed `EffectDealDamage` "you" recipient.
Any other trailing effect, a targeted or grouped recipient, a divided or variable
amount, or a negated rider fails closed. A commander-identity add-mana ability
may instead carry a mana-spend rider (Path of Ancestry): the typed
`EffectManaSpendRider` following the add-mana effect lowers to
`game.TapManaCommanderIdentityWithSpendRiderAbility`, attaching a
`game.ManaSpendRider` (the commander-creature-type spend condition plus a `scry N`
`game.Mode`) to the produced mana via the add-mana instruction's `SpendRider`
option. Lowering accepts the rider only when the typed condition, effect, and a
positive scry amount are all present, so every wording the parser left untyped
fails closed.

The mana-ability lowering branch credits the activated ability's trailing
reminder-text spans toward source coverage (mirroring the general activated-
ability branch). This unblocks mana abilities whose only remaining uncovered
tokens were parenthetical reminder text — for example Path of Ancestry's scry
reminder, and basic-utility mana sources whose reminder explains `{C}`, milling,
or the untap symbol `{Q}`. Token-creation effects synthesize a
token `*game.CardDef` from the typed token spec (subtype, types, colors, fixed
power/toughness, and an optional single granted keyword) and emit a
`game.CreateToken` instruction; the recipient is the controller by default, the
controller of a referenced object (`game.ObjectControllerReference`) for the
"Its controller creates …" follow-on form in an ordered sequence (the Beast
Within pattern), or a single targeted player (`game.TargetPlayerReference(0)`)
for the "Target opponent/player creates …" form, which also emits the matching
player `TargetSpec` on the mode. A targeted-player recipient is accepted only for
fixed counts and only when the target is exactly one player or opponent; player-
group recipients ("Each opponent creates …", "Each player creates …") have no
single player reference and stay fail-closed. A "tapped" entry modifier ("Create a tapped … token.") sets the
instruction's `EntryTapped` flag so each created token enters the battlefield
tapped; the modifier applies to both synthesized creature tokens and predefined
named artifact tokens. A trailing attacking-entry clause ("Create a … creature
token that's tapped and attacking.") sets the instruction's `EntryAttacking` flag
so each created creature token is put onto the battlefield already attacking (CR
508.4); its "tapped" word lives in that clause and continues to set `EntryTapped`. The token count may also be the spell's variable `X`
(lowered to the runtime `game.DynamicAmountX`) or a rules-derived dynamic count.
A "for each <X>" iterator (in either the leading "For each <X>, create …"
position or the trailing "Create … for each <X>" position), a "number of …
equal to <X>" count, and a "where X is <X>" count all lower their iterated or
counted subject through the shared dynamic-amount lowerer, so the instruction's
`Amount` is a `game.Dynamic` count instead of exactly one. A dynamic count whose
subject the dynamic-amount lowerer cannot represent (source power, devotion,
fractions of a count, ...) fails closed. The renderer collects
each synthesized token def and writes it as a card-scoped package-level `var`
alongside the card that creates it (`renderCtx.tokenDefVar`). The whole-card Oracle
text is emitted once as each generated card's top-level `OracleText`; the
renderer no longer reproduces the source text of each sub-portion (ability,
mode, condition, etc.). Retained source text survives into rendered cards only
where the runtime reads it — the additional-cost `Text` (the "discard this card"
cost check in `mtg/rules`) and replacement-ability descriptions — plus
unsupported-card diagnostic messages and exact source-span consumption
accounting. Lowering's fail-closed source-coverage gate (which rejects any card
whose source is not fully accounted for by recognized semantics) consumes the
parser's `CoverageSpans()` must-cover assertion and checks each span against the
spans it recognized, rather than walking the raw token stream and classifying
comma/colon/period/em-dash token kinds itself. This boundary is enforced
automatically by
`TestLoweringTextInterpretationIsAllowlisted` in
`text_blindness_enforcement_test.go`, an AST analyzer that fails if any `cardgen`
lowering code inspects Oracle-text-valued data (`strings`/`regexp`/word
normalization over token `.Text`/`.Event` values, or string-literal comparisons
of that text) outside a small, individually justified allowlist of diagnostic and
rendering uses. The companion `TestCompilerIsTextBlind` proves the
`oracle/compiler` package performs no such interpretation at all (empty
allowlist), and `TestEnforcementDetectsViolations` checks the analyzer against
synthetic violating and clean sources.

## Usage

Compile the Scryfall Oracle Cards corpus into a temporary Card Registry tree:

```bash
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/generated-cards \
  -report .cardwork/oracle-compile-report.json
```

After inspecting and validating the temporary tree, use `-out mtg/cards` only
when intentionally updating repository card definitions. The command
regenerates affected letter-package `cards.go` files.

Cards outside compiler coverage remain unsupported by Card Generation. Truly
exceptional mechanics may still use a hand-written **Card Implementation** with
an `ImplementationID`; that escape hatch is independent of this compiler.

## Tooling layout

- `cardgen/oracle`: lexer, parser, semantic compiler, corpus checks, and bulk
  compilation command.
- `cardgen`: typed lowering, Card Definition assembly, deterministic rendering,
  Scryfall data shapes, and source naming helpers.
- `cardgen/cmd/gencardlist`: `go generate` helper that writes each
  `mtg/cards/<letter>/cards.go` Card Registry list.

## Supported layouts

The source generator can represent Scryfall `normal`, `token`, `leveler`,
`saga`, `class`, `case`, `prototype`, `host`, `augment`, `emblem`, `mutate`,
`planar`, `scheme`, `vanguard`, `transform`, `modal_dfc`, `meld`,
`double_faced_token`, and `reversible_card` layouts. Corpus policy is narrower:
it excludes nonstandard game objects such as emblems, planes, schemes, and
Vanguard cards before source generation.

Transform, modal DFC, and double-faced token cards emit front-face fields on
`CardDef` and an optional `Back` face. Meld cards emit their front card with
`LayoutMeld`; complete meld behavior remains rules work. Reversible cards emit
one Card Definition per playable side.

## Key interfaces

- `GenerateExecutableCardSource(card, pkgName)` recognizes, lowers, validates,
  and renders a card, or returns diagnostics without source.
- `ExecutableGenerator` configures source identity disambiguation for bulk
  generation.
- `Renderer.RenderCardSource(card, defs, hints, pkgName)` renders validated typed
  Card Definitions deterministically.
- `ParseTypeLine(typeLine)` splits a type line into supertypes, types, and
  subtypes.
- `GeneratedIdentity` selects a generated card's category, package, filename,
  variable name, and migration path. `CardNameToVarName`,
  `CardNameToFileName`, and `CardNameToPackageLetter` provide its component
  naming rules.

Prepare layouts use `CardFace.EntersPrepared` on the creature face and
`CardDef.Alternate` for the spell face. The generator accepts them only when
both faces and the exact enters-prepared ability are fully lowerable.

Current executable mechanic coverage and the corpus support count live in
[`oracle/README.md`](oracle/README.md). The numbered expansion checklist lives
in [`../docs/oracle-compiler-expansion.md`](../docs/oracle-compiler-expansion.md).
