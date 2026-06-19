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
   prior-instruction results. `activation.go` composes the generic activated
   shell from typed cost components, timing, zone of function, activation
   condition, bound references, and shared Ability Content. Mana and non-mana
   activated abilities use that same shell preparation while retaining distinct
   runtime types. Known shell failures report activation cost, timing, zone,
   condition, reference, mode, or structure diagnostics instead of a generic
   activated-ability failure. `static_declaration.go` is the single mechanical
   adapter from semantic Static Declarations to `game.StaticAbility`,
   `game.ContinuousEffect`, `game.RuleEffect`, and `game.CostModifier` values.
   Mixed static paragraphs lower through that adapter as multiple declarations
   sharing one runtime static ability. Recognized semantics
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
   continuous effects. Until-end-of-turn keyword-grant spells over a controlled,
   attacking, blocking, or all-creatures group (`Creatures you control gain
   trample until end of turn.`, including multiple keywords joined by `and`)
   lower through `lowerGroupTemporaryKeywordSpell` into a `game.LayerAbility`
   `AddKeywords` continuous effect over the group for
   `game.DurationUntilEndOfTurn`; color-filtered groups and quoted granted
   abilities remain fail-closed. Static power/toughness and keyword group anthems also
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
   predicate. Keyword-filter ("Creatures with flying"), excluded-supertype
   ("Nonlegendary"), color-exclusion ("Nonblack"), granted quoted-ability, group
   rule, and dynamic group anthems remain fail-closed. The static source-tied control grant on control Auras
   ("You control enchanted creature/permanent") lowers to a layer-2 control
   continuous effect over the attached object whose new controller is the Aura's
   controller. Exact
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
   OR (not AND) semantics. The `MoveCard`/`PutOnBattlefield` graveyard-return
   primitives are rebasable inside an ordered effect sequence (`Return target
   creature card from your graveyard to your hand, then create a token.`): the
   sequence target-rebaser rewrites their `CardReference.TargetIndex` by the count
   of preceding card-kind target specs (`cardTargetSpecsBefore`), which differs
   from the global target-spec offset used for object/player references because
   `CardReference.TargetIndex` is numbered among card targets only. A
   `PutOnBattlefield` carrying entry counters or continuous effects, and the mixed
   inherited-plus-owned-target remap path for these primitives, stay fail-closed.
   Mass return-to-hand spells (`Return all <group> to their owners' hands.`,
   including the `you control` self-control variant) lower to a single
   `game.Bounce` over a `BattlefieldGroup` Selection built by the shared
   `massGroupSelection`, mirroring mass destroy/exile; the only tolerated
   reference is the destination's possessive pronoun, and choice-based color
   filters, `except for` riders, `all but one`, and the single-choose
   `Return a permanent you control` form stay fail-closed.
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
   Exact fixed, `X`, and supported dynamic placement of recognized named
   counters lowers from supported spell, activated, loyalty, triggered,
   ordered-effect, and Saga chapter bodies into typed `game.AddCounter`
   permanent instructions or `game.AddPlayerCounter` instructions for poison,
   energy, and experience. The placement object may be a single target, every
   permanent in a filtered battlefield group (`Put a +1/+1 counter on each
   creature you control.`), each of several targets for the multi-target
   `each of up to N target <permanent>s` form (lowered to one `game.AddCounter`
   per target slot, mirroring multi-target graveyard return, with optional
   `other` self-exclusion and controller clause), the
   source permanent itself for fixed self-placement bodies
   (`Put a +1/+1 counter on this creature.`, lowered to
   `game.SourcePermanentReference()`), or a prior clause's target referenced by
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
   triggers. The same path lowers exact fixed until-end-of-turn self-pump
   bodies (`This creature gets +X/+Y until end of turn.`) when the sole subject
   reference is `ReferenceBindingSource`, and inherited-target pump bodies
   (`… It gets +X/+Y until end of turn.`) when the sole reference binds to a
   prior clause's target; the object lowers to
   `game.SourcePermanentReference()` or a target reference accordingly.
   Dynamic until-end-of-turn pumps whose `where X is …` count machinery is
   already supported lower each side independently, so asymmetric and mixed-sign
   forms (`Target creature gets +X/+0 …`, `… +X/-X …`, `… -X/-X …`) lower
   alongside the symmetric `+X/+X` form. Exact fixed until-end-of-turn pumps on a
   single target slot also lower through `lowerFixedModifyPTTargets`, which reuses
   the shared `permanentTargetSpecWithCardinality` and emits one `ModifyPT` per
   target slot: plural (`Two target creatures each get -1/-1 until end of turn.`),
   optional (`Up to one/two target creatures … gets/each get …`), and creature-
   subtype (`Target Human you control gets +2/+2 …`) targets are supported, with
   declined "up to" slots no-opping on their unresolved target index. Non-creature
   pump targets, dynamic multi-target amounts, and riders stay fail-closed.
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
   variant restricting the player half to opponents. Fixed-amount group damage
   recipients (`each creature with flying`) accept a runtime-modelable keyword
   filter on the recipient Selection; keywords the runtime cannot model as a
   selector predicate stay fail-closed.
   A damage recipient that is the controller or owner of the prior removal target
   in an ordered sequence (`Destroy target land. <name> deals 2 damage to that
   land's controller.`, `… deals N damage to its owner.`) lowers through
   `lowerReferencedPlayerDamageSpell` to a `game.PlayerDamageRecipient` wrapping
   `game.ObjectControllerReference`/`game.ObjectOwnerReference` of the inherited
   target (a `TargetPermanentReference` for a permanent antecedent, a
   `TargetStackObjectReference` for a countered spell); the damage carries
   `game.SourcePermanentReference()` only when the source is a permanent
   ("This creature deals …"), and only fixed or `X` amounts are accepted.
   A destroy spell carrying a parser-folded regeneration rider ("It/They can't
   be regenerated.") lowers to a `game.Destroy` with `PreventRegeneration: true`,
   for the single-target, multi-target, and mass forms alike; the renderer emits
   the flag explicitly so the generated card bypasses regeneration shields.
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
   shared event-player draw/discard/mill paths using exact "they" pronoun
   forms, resolving the player via `game.EventPlayerReference()`. Exact
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
   Sacrifice costs recognize a subtype, an explicit count, the source itself
   ("Sacrifice this <subtype>"), "another" (an exclude-source sacrifice), and a
   two-type union of permanent types ("Sacrifice an artifact or creature");
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
   costs accept a fixed count, a typed card ("exile a creature card"), or an
   `X`-bound count ("exile X cards from your graveyard") that resolves against the
   spell's announced X.
   Exact trailing activation restrictions lower to typed sorcery, combat,
   upkeep, and once-per-turn timing checks.
   Common enters-tapped life, opponent-count, land-count, and
   basic-land-subtype conditions lower into typed replacement predicates.
   Plain self enters-tapped replacements lower from the parser-owned
   `EntersTappedSelf` flag, which recognizes the tapped entry qualifier (for any
   subject noun or card-name phrasing) rather than matching whole Oracle
   sentences. Exact optional pay-2-life and reveal-a-land-or-creature-subtype
   entry wordings lower into typed resolution payments for enters-tapped
   replacements from their typed effect structure. Entry-time choice replacements
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
   (card/land/creature/artifact/enchantment), the `basic` supertype, or a union of
   basic land subtypes ("Forest or Island", "basic Forest, Plains, or Island"),
   moved to hand or the battlefield (optionally tapped) and optionally revealed
   first. The runtime treats the count as a maximum and lets the searching player
   legally fail to find. An optional tutor ("You may search your library for …")
   lowers through the same exact round-trip — the parser strips the leading "you
   may" before reconstructing the canonical search shape — and marks the single
   resulting `game.Search` instruction `Optional` so the runtime offers the player
   the choice to decline. Graveyard-also searches, other players' libraries,
   "with different names", mana-value/power/color filters, variable `X` counts,
   and unsupported destinations remain fail-closed.
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
rather than by re-parsing the reminder text. Token-creation effects synthesize a
token `*game.CardDef` from the typed token spec (subtype, types, colors, fixed
power/toughness, and an optional single granted keyword) and emit a
`game.CreateToken` instruction; the recipient is the controller by default, or
the controller of a referenced object (`game.ObjectControllerReference`) for the
"Its controller creates …" follow-on form in an ordered sequence (the Beast
Within pattern). A "tapped" entry modifier ("Create a tapped … token.") sets the
instruction's `EntryTapped` flag so each created token enters the battlefield
tapped; the modifier applies to both synthesized creature tokens and predefined
named artifact tokens. A leading "for each <X>," iteration prefix on a controller
create types the iterator as a dynamic count (the for-each subject) with the
single created token as its multiplier, so the instruction's `Amount` is a
`game.Dynamic` count of the iterated objects and one token is created per object
instead of exactly one. The renderer collects
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
