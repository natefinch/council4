# Oracle Compiler Expansion

This document is both the rollout checklist and the execution guide for
expanding executable Oracle-text compilation. An agent should be able to resume
from this file without relying on conversation history.

**Current corpus support: 2,234 / 38,101 cards**

The expansion plan was established in `7e65c8e` (`compiler expansion plan`).
Expansion steps 1–10 are complete.

## Goal

Card Generation turns Scryfall bulk data and Oracle text into executable
`game.CardDef` Go source:

```text
Scryfall JSON
  -> cardgen/oracle recognition
  -> cardgen typed lowering
  -> game.ValidateCardDef
  -> cardgen deterministic rendering
```

There is no partial or agent-completed generation path. A card is generated only
when every face and every ability is represented exactly. Unsupported cards
receive source-spanned diagnostics and no source file.

## Architecture and ownership

- `cardgen/oracle` owns lexical, syntactic, and semantic recognition. Its values
  describe what the Oracle text says and preserve exact source spans. It must
  not construct runtime `game` values.
- `cardgen/lower.go` owns semantic-to-typed lowering. It turns recognized
  abilities into `game.*` values and rejects wording that the typed model cannot
  represent exactly.
- `cardgen/executable.go` assembles typed Card Definitions and calls
  `game.ValidateCardDef` before rendering.
- `cardgen/render.go` owns deterministic Go spelling. Rendering changes must not
  change recognition, and recognition changes must not contain source templates.
- `mtg/game` owns pure typed Card Definition data and structural validation.
- `mtg/rules` owns runtime behavior, resolution, targeting, costs, events, and
  choices.
- `cardgen/oracle/cmd/compilecards` is the only bulk generation command. It emits
  complete supported cards and a deterministic unsupported report.

Read these before changing the compiler:

- [`../CONTEXT.md`](../CONTEXT.md), especially Card Definition, Card
  Implementation, Effect Primitive, Instruction, Selection, and Card Generation.
- [`adr/0008-typed-ir-lowering.md`](adr/0008-typed-ir-lowering.md).
- [`../cardgen/README.md`](../cardgen/README.md).
- [`../cardgen/oracle/README.md`](../cardgen/oracle/README.md).

## Non-negotiable invariants

1. **Fail closed.** Never infer behavior from a nearby phrase, silently discard
   unsupported text, or emit a partially lowered ability.
2. **Exhaustive semantic consumption.** `abilityLowering.complete` must account
   for every cost, trigger, mode, target, condition, effect, keyword, and
   reference in the semantic intermediate representation.
3. **Exhaustive source consumption.** Every meaningful syntax token must be
   covered by an accepted source span. Punctuation and explicitly parsed
   reminder text are the only routine exceptions.
4. **Exact typed fidelity.** Do not lower a phrase unless `game` data and `rules`
   behavior represent its targets, controller relationships, timing, choices,
   costs, optionality, and ordering.
5. **Validate before render.** Generated source must come only from assembled
   typed values that pass `game.ValidateCardDef`.
6. **Deterministic output.** Identical input must produce byte-identical,
   gofmt-stable output. Sort externally visible collections and never depend on
   map iteration.
7. **One behavior path.** Reuse shared effect, target, cost, reference, and
   renderer modules. Do not add a second string-first generator or special
   source injection path.
8. **Preserve Oracle order.** Instruction sequences and mode contents resolve in
   printed order.
9. **Prefer existing model depth.** Reuse Effect Primitives, mechanic templates,
   Selection, and typed references. Add `game` or `rules` capability only when
   the wording cannot be represented faithfully without it.
10. **Keep renderer checks strict.** Canonical mechanic-template calls may be
    emitted only when the typed value exactly equals the canonical template.

## Working discipline

### Before a numbered step

1. Confirm the worktree is clean and inspect the latest corpus count.
2. Read the relevant semantic types, lowering code, renderer path, Card
   Definition validation, and runtime handlers before designing new data.
3. Query the unsupported report for the intended diagnostic families and sample
   real card text. Counts in this document are planning signals, not promised
   generated-card gains.
4. Establish a saved baseline report and generated tree for comparison.
5. Mark only the current numbered step in progress. Leave later steps untouched.

### Test-driven vertical slices

Use test-driven development: one observable behavior at a time, each taken from
recognition through generated source.

For each wording family:

1. Add one public semantic compiler test and confirm it fails.
2. Implement the smallest recognition change and make that test pass.
3. Add one typed or generated-source test through
   `GenerateExecutableCardSource` and confirm it fails.
4. Add the smallest lowering and renderer change and make it pass.
5. Add a near-miss rejection test that proves the support boundary.
6. Add or identify a runtime rules test when new typed behavior or a new
   combination is introduced.
7. Refactor only after the slice is green.

Tests should assert behavior through public compiler interfaces, not private
helper structure. A refactor should not invalidate them unless generated
behavior changes.

### Do not broaden accidentally

Every accepted family needs explicit rejected neighbors, such as:

- fixed versus variable quantities;
- one occurrence versus repeated instructions;
- exact target nouns versus qualified or linked targets;
- controller versus opponent versus "you don't control" constraints;
- mandatory versus optional instructions;
- one mode versus unsupported mode cardinality;
- represented costs versus unrepresented cost components;
- exact self events versus other-object or group events.

If a near-miss can pass while losing semantics, stop and deepen the typed model
or narrow recognition.

## Corpus workflow

Use the checked-in Scryfall Oracle Cards corpus through `corpusdelta`:

```bash
go run ./cardgen/oracle/cmd/corpusdelta \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -baseline .cardwork/step-PREV-report.json \
  -out .cardwork/step-N-generated \
  -report .cardwork/step-N-report.json \
  -manifest .cardwork/step-N-delta.json
```

Never target `mtg/cards` during development. Overwrite repository cards only
when the user explicitly requests it after the temporary tree is accepted.

The command compares reports by stable card ID, verifies counts and newly
generated source paths, updates `docs/supported.md`, emits the inspection
manifest, and tests and vets every generated package.

Record:

- total cards;
- generated cards;
- unsupported cards;
- the previous count and exact delta;
- every newly generated card;
- every diagnostic family that disappeared or materially shrank.

Inspect every newly generated source listed in `.cardwork/step-N-delta.json`.
Classify each by wording family and
confirm that every ability is represented, targets and costs are exact,
instruction order is correct, and no reminder or qualifier was lost. A positive
count is not success if even one new card is a false positive.

`corpusdelta` validates generated packages from a temporary directory inside the
Go module and removes that directory afterward.

Do not run `mage lint` while tests are creating or deleting temporary
`cardgen/roundtrippkg*` packages; lint type-checking can observe a transient
package.

## Review, validation, and commit gate

After each numbered step:

1. Run the full corpus and inspect every newly generated card.
2. Send the complete uncommitted diff, intended support boundary, rejection
   cases, corpus before/after counts, and inspected families to an independent
   Opus 4.8 review.
3. Ask the reviewer to focus on false positives, source-span safety, unconsumed
   semantics, target and cost fidelity, renderer completeness, runtime behavior,
   panic risks, and missing negative tests.
4. Fix every material finding and request a re-review of the fix.
5. Run:

   ```bash
   gofmt -w <changed Go files>
   go test ./... -count=1
   go vet ./...
   go build ./...
   mage lint
   git diff --check
   ```

6. Update the support count near the top of this file and
   `cardgen/oracle/README.md`.
7. Check off the numbered step only after all gates pass.
8. Commit the step separately, including the required Copilot co-author trailer.
9. Begin the next numbered step only after that commit exists.

If validation is blocked by infrastructure, report the blocker and do not check
off or commit the step as complete.

## Completed work

- [x] **1. Lower parameterized Equip using `EquipActivatedAbility`**
  - Planning signal: approximately 430 blockers.
  - Exact `Equip MANA_COST` lowering with near-miss rejection.
  - Completed in `2992ba8` (`Lower parameterized Equip abilities`).
- [x] **2. Add complete Enchant/Aura and Protection templates and lowering**
  - Planning signal: 1,083 blockers.
  - Base Enchant target types and color-only Protection.
  - Completed in `a41f015` (`Lower Enchant and Protection keywords`).
- [x] **3. Build composable multi-effect sequence lowering**
  - Planning signal: foundation for thousands of cards.
  - Reuses exact single-effect lowering in Oracle order. Multiple targeted
    clauses remain rejected until target-index remapping is implemented.
  - Completed in `8df2720` (`Lower ordered effect sequences`).
- [x] **4. Expand supported spells using sequence lowering**
  - Planning signal: 4,625 blockers.
  - Added Surveil, Investigate, Proliferate, Regenerate, Fight, and
    reminder-aware exact lowering.
  - Corpus moved from 1,809 to 1,838 generated cards.
  - Completed in `dd1695b` (`Expand supported Oracle spells`).

## Remaining steps

### 5. Ordinary activated abilities with mana and tap costs

- [x] Complete and commit step 5.

**Planning signal:** 5,947 activated-ability blockers plus 690 cost issues.

Build an ordinary non-mana `game.ActivatedAbility` path rather than expanding
the existing mana-ability special case.

Initial scope:

- exact mana-only costs;
- exact tap-only costs;
- exact mana-plus-tap costs in printed order;
- an effect body accepted by the existing single-effect or ordered-effect
  lowering;
- battlefield zone of function unless Oracle semantics prove another supported
  zone;
- targets and instruction sequences identical to the equivalent spell wording.

Implementation guidance:

- Lower `oracle.CompiledCost.Components` into `ManaCost` and
  `[]cost.Additional`; reuse typed mana parsing and `cost.Tap`.
- Split the syntax at the top-level colon and pass only the post-colon body to
  shared effect lowering.
- Consume the complete cost span and all body semantic/source spans.
- Preserve the distinction between `game.ManaAbility` and
  `game.ActivatedAbility`; do not classify an ability by generated primitive
  shape alone.
- Add renderer coverage for ordinary activated abilities only through typed
  fields already present on `game.ActivatedAbility`.

Reject initially:

- untap, sacrifice, discard, life, exile, counter-removal, or variable costs
  unless each component is separately implemented and runtime payment is proven;
- activation restrictions, timing clauses, "activate only," and non-battlefield
  zones;
- X-dependent effects, choices, conditional costs, and costs referring to other
  objects;
- any body not already accepted by shared effect lowering.

Completed with exact mana-only, tap-only, and mana-then-tap costs. The corpus
moved from 1,838 to 1,992 generated cards; all 154 newly supported cards were
inspected with no false positives.

### 6. Broader enter and dies triggers

- [x] Complete and commit step 6.

**Planning signal:** 9,599 trigger blockers plus 632 related issues.

Generalize the current self-enter/self-dies path without weakening event
semantics.

Initial scope:

- multiple supported sentence-sized effects through the shared ordered-effect
  path;
- exact optional "you may" triggers using `game.TriggeredAbility.Optional`;
- additional self subjects only when they map exactly to the existing self event
  pattern;
- enter and dies bodies with the same target restrictions as supported spells.

Implementation guidance:

- Isolate the trigger clause at the first top-level comma and lower the remaining
  body through the same shared content module as spells and activated abilities.
- Keep `TriggerSourceSelf` and the exact `EventKind`; do not turn other-object or
  group triggers into self triggers.
- Preserve dies-trigger last-known behavior supplied by the rules event model.
- Treat optionality as trigger-level or instruction-level according to the
  printed wording; do not erase the distinction.
- Include reminder spans and all body references in exhaustive consumption.

Reject initially:

- intervening-if conditions;
- "whenever one or more," other-object, leaves-the-battlefield, cast, attack, or
  damage triggers;
- delayed, reflexive, linked-object, and once-per-turn triggers;
- optional wording whose choice point is not represented exactly.

Completed with ordered supported effect bodies and exact leading `you may` for
single-effect trigger-level optionality. Partially optional sequences remain
rejected. The corpus moved from 1,992 to 2,010 generated cards; all 18 newly
supported cards were inspected with no false positives. Typed `Host` supertype
support was added so the newly generated host-layout cards preserve their full
type lines.

### 7. Conditional enters-tapped replacements

- [x] Complete and commit step 7.

**Planning signal:** 1,247 blockers.

Start from exact conditions representable by `game.Condition` and existing
replacement constructors such as `EntersTappedIfReplacement` and
`EntersTappedUnlessPaidReplacement`.

Implementation guidance:

- Recognize and lower one condition family at a time.
- Prove the condition uses the correct controller, source, zone, and evaluation
  time for an entering permanent.
- Keep conditional replacement recognition separate from unconditional
  `EntersTappedReplacement`.
- Add runtime tests showing both condition outcomes before accepting a family.

Reject any condition that requires missing information, linked choices,
unrepresented land-type logic, or a payment form the rules cannot offer and
apply exactly.

Completed with two unlocked families. The corpus moved from 2,010 to 2,080
generated cards (+70); all newly supported cards were inspected with no false
positives.

**Family A — conditional enters-tapped (10 cards):** Canopy Vista, Cinder
Glade, Eclipsed Steppe, Prairie Stream, Radiant Summit, Scorched Geyser,
Smoldering Marsh, Sodden Verdure, Sunken Hollow, Vernal Fen. Wording: "This
land enters tapped unless you control two or more basic lands." Lowered to
`game.EntersTappedIfReplacement` with `game.Condition{Negate: true,
ControllerControls: game.PermanentFilter{Types: []types.Card{types.Land},
Supertypes: []types.Super{types.Basic}, MinCount: 2}}`.

**Family B — parenthesized mana ability reminder (60 cards):** Basic lands
(Forest, Island, Mountain, Plains, Swamp, snow-covered variants), old dual
lands (Badlands, Bayou, Plateau, Savannah, Scrubland, Taiga, Tropical Island,
Tundra, Underground Sea, Volcanic Island), bicycle lands (Canyon Slough, Fetid
Pools, etc.), and triomes (Indatha Triome, Jetmir's Garden, Ketria Triome,
etc.). These cards had only parenthesized reminder mana ability text; the
compiler now recognises `AbilityReminder` abilities whose inner content is an
activated mana ability and lowers them through the existing `lowerTapManaAbility`
path. Hybrid-mana cost reminders (e.g. `({R/W} can be paid with…)`) remain
unsupported and are correctly rejected.

### 8. Common static effects

- [x] Complete and commit step 8.

**Planning signal:** 3,609 blockers.

Add high-frequency static declarations through existing `game.StaticAbility`,
`game.ContinuousEffect`, `game.RuleEffect`, Selection, and cost-modifier data.

Prioritize:

- fixed power/toughness buffs with exact affected groups;
- attack, block, cast, activate, or targeting restrictions already represented
  by Rule Effects;
- exact cost increases or reductions already handled by the payment rules;
- paragraphs mixing supported keywords with one supported static declaration.

Implementation guidance:

- Static abilities declare data; they do not emit resolving Instructions.
- Match affected groups with Selection and Group Reference instead of adding
  one-off predicates.
- Mixed keyword text must consume every keyword and every declaration in the
  paragraph.
- Add runtime effective-value or legality tests for each declaration family.

Reject characteristic-defining abilities, dependency-sensitive layers,
unbounded durations, choice-dependent text, and restrictions that existing
rules do not enforce.

Completed with four P/T buff families. The corpus moved from 2,080 to 2,158
generated cards (+78); all 78 newly supported cards were inspected with no
false positives.

A new `StaticSubjectKind` enum in `cardgen/oracle` captures the pre-verb
subject for static declarations: `None`, `AttachedObject`,
`ControlledCreatures`, and `OtherControlledCreatures`. The lowering recognises
the exact token sequence for each family and maps it to the appropriate
`GroupReference` for the runtime `ContinuousEffect`.

**Family A — Enchanted creature gets +X/+Y (~38 cards):** Auras with a simple
P/T buff on the enchanted creature. Wording: "Enchanted creature gets +N/+N."
Lowered to a `game.StaticAbility` with a `ContinuousEffect` using
`AttachedObjectGroup(SourcePermanentReference())` as the group reference and
`LayerPowerToughnessModify` as the layer.

**Family B — Equipped creature gets +X/+Y (~25 cards):** Equipment with a
simple P/T buff on the equipped creature. Wording: "Equipped creature gets
+N/+N." Same group-reference pattern as enchanted-creature.

**Family C — Creatures you control get +X/+Y (~9 cards):** Anthem-style
permanents. Wording: "Creatures you control get +N/+N." Lowered with
`ObjectControlledGroup(SourcePermanentReference(),
Selection{RequiredTypes:[Creature]})`.

**Family D — Other creatures you control get +X/+Y (~5 cards):** Lord-style
permanents that do not buff themselves. Wording: "Other creatures you control
get +N/+N." Lowered with
`ObjectControlledGroupExcluding(SourcePermanentReference(),
Selection{RequiredTypes:[Creature]}, SourcePermanentReference())`.

### 9. Loyalty and modal abilities

- [x] Complete and commit step 9.

**Planning signal:** 695 loyalty blockers plus 335 modal blockers.

Reuse shared effect content rather than building loyalty- or mode-specific
primitive lowerers.

Loyalty scope:

- exact signed loyalty costs represented by `game.LoyaltyAbility.LoyaltyCost`;
- supported single or ordered effect bodies;
- existing target and source-reference rules.

Modal scope:

- exact "Choose one" first, then other fixed cardinalities represented by
  `AbilityContent.MinModes` and `MaxModes`;
- lower each mode independently through shared effect lowering;
- preserve mode-local targets, source spans, reminders, and Oracle order;
- use shared targets only when the wording truly shares targets across modes.

Reject variable mode counts, repeatable modes, "choose one or both," entwine,
escalate, hidden mode dependencies, and any target-index arrangement that is not
represented exactly.

Completed with loyalty ability infrastructure and "Choose one" modal spells.
The corpus moved from 2,158 to 2,164 generated cards (+6); all 6 newly
supported cards were "Choose one" modal Instants/Sorceries and were inspected
with no false positives. Zero loyalty-ability cards were generated because every
planeswalker in the corpus has at least one ability body too complex for the
current lowerer, but the full loyalty lowering pipeline (cost parsing, effect
lowering, rendering) is in place.

Choose-N with N≥2 and choose-one-or-both variants remain unsupported (#16).

### 10. Layouts, then frequency-driven mechanics

- [x] Complete and commit step 10.

**Planning signal:** 344 playable layout blockers and 10,288 Oracle constructs.

Add Adventure, split, and the requested prepare-related layout family before
chasing lower-frequency text mechanics. Verify the exact Scryfall layout names
and `game.CardDef` representation first; do not guess layout semantics from a
display name.

For each layout:

- define which faces are independently castable or selectable;
- preserve face names, costs, types, colors, Oracle text, and color identity;
- ensure generated identifiers and paths cannot collide;
- add generated-source round-trip tests and runtime face-selection tests;
- reject the whole card when any required face cannot be represented.

After layouts, use the unsupported report to rank remaining semantic and syntax
constructs by card impact. Select one coherent wording family at a time and run
the same vertical-slice, corpus, Opus review, validation, and commit gates.

Do not optimize for diagnostic count alone. Prefer reusable compiler and game
model depth that safely unlocks several exact families.

Completed with exact Adventure and split layout support in both the typed card
definition pipeline and runtime casting/resolution paths, including
`Alternate`-face rendering, hand-casting of alternate spell faces, split-half
casting, and Adventure exile tracking plus creature recasting from exile.
`prepare` was removed from the generator support list because its prepared-state
and copy-casting runtime semantics are still unimplemented; deferred follow-up
is tracked in issue #18. This step also fixed the `objectWord` mechanical bug
so exact "This artifact enters tapped." replacements now lower successfully.
The corpus moved from 2,164 to 2,176 generated cards (+12): 5 Adventure cards,
2 split cards, and 5 normal artifact mana rocks unlocked by the
object-reference fix. Every newly generated card in
`.cardwork/step-10-delta.json` was inspected after the runtime fixes, and the
delta contains no remaining false positives.

## Known technical limits

- Ordered effect lowering currently permits at most one targeted clause because
  isolated effect lowering preserves target index zero. Add explicit target-index
  remapping before accepting multiple independently targeted clauses.
- Each semantic effect currently emits one Instruction. Revisit sequence
  accounting before adding an effect lowerer that emits several Instructions.
- Reminder text can be removed from exact syntax checks only when its parsed span
  remains consumed by the outer lowering.
- Qualified Enchant and non-color Protection remain unsupported.
- Runtime stack-object target discovery is a prerequisite for counterspell
  lowering; a recognized `Counter target spell` phrase alone is insufficient.
- Hand-written Card Implementations remain an exceptional escape hatch, not an
  alternate generation workflow and not a reason to emit partial compiler output.

## Final rollout gate

After step 10:

1. Run a fresh full-corpus compilation from the committed baseline.
2. Inspect all cards newly supported across the final step.
3. Validate every generated package and the full repository.
4. Request a final independent Opus 4.8 architecture and correctness review.
5. Reconcile this document, `cardgen/oracle/README.md`, and the final report
   counts.
6. Confirm no temporary `cardgen/roundtrippkg*` packages or generated corpus
   trees are tracked.
7. Commit final review fixes separately if any are required.
