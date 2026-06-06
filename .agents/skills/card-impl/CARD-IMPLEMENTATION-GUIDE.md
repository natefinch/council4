# Card Implementation Guide

Reference for parsing Magic: The Gathering oracle text into council4 `CardFace` ability fields.

## Canonical source layout

**Read `mtg/cards/k/karplusan_forest.go` before implementing any card.** It is the canonical
reference for how new card source must be formatted. Key rules:

1. The `CardDef` literal is vertically expanded — never compact on one line.
2. `ColorIdentity` appears before `CardFace` in the struct literal.
3. `CardFace` is vertically expanded; `Name`, `Types`, and other fields are ordinary struct fields.
4. `OracleText` uses an indented raw multiline string: opening backtick on its own field line, one
   oracle paragraph per source line, closing backtick indented on its own line.
5. Every ability body's `Text` field uses the same indented raw multiline string style.
6. Categorized ability slices and bodies are vertically expanded: one brace level per line. Do not
   use compact `{{` forms for card ability bodies.
7. Small truly atomic leaf values may stay one-line, e.g. `[]cost.Additional{{Kind: ...}}` and
   simple single-field `Effect` literals. Complex `Effect` values are vertically expanded.
8. Use categorized `CardFace` fields (`ManaAbilities`, `ActivatedAbilities`, etc.), not the legacy
   `CardFace.Abilities` slice.
9. Preserve oracle order naturally when one categorized slice suffices. If categories are mixed and
   field grouping would obscure oracle order, use an initializer function with categorized appends,
   but still format the base `CardDef` and appended bodies in this expanded/raw-text style.
10. Keep the top oracle comment block.
11. Run `gofmt` after writing, but write this layout explicitly — do not rely on `gofmt` to create it.

The generator (`go run .agents/skills/card-impl/main.go`) already emits mechanical fields in this
style. Preserve that layout when filling in ability bodies.

## Go Type Definitions

These are the exact types you must use. Do not invent new enum values.

### opt.V — optional values

Optional fields use `opt.V[T]` instead of `*T`. To set an optional field use
`opt.Val(value)`; to leave it absent, simply omit it (zero value means absent).
Check presence with `.Exists`. Import path: `"github.com/natefinch/council4/opt"`.

```go
// Set a value:
ManaCost: opt.Val(cost.Mana{cost.R})
// Absent (default): just omit the field
```

### Card type vocabulary

Card supertypes, card types, and subtypes live in
`"github.com/natefinch/council4/mtg/game/types"`. Use
`types.Super`, `types.Card`, and `types.Sub` values such as `types.Legendary`,
`types.Creature`, and `types.Forest`. Do not use old `game.Type*`,
`game.*Subtype*`, or `game.CardType` names.

`mtg/game/types` has named constants for every Comprehensive Rules 205.3
subtype. Prefer those constants in card definitions instead of
`types.Sub("...")`. Examples: use `types.Warrior`, `types.Rogue`,
`types.TimeLord`, `types.Omen`, `types.Siege`, and
`types.BolassMeditationRealm`. The only duplicated printed subtype currently has
family-qualified identifiers: `types.ArtifactSpacecraft` and
`types.PlanarSpacecraft`, both with the string value `"Spacecraft"`. The subtype
lists are organized by card-type family in files under `mtg/game/types`, and
card definitions should import that parent package.

Integer comparisons live in
`"github.com/natefinch/council4/mtg/game/compare"`. Use `compare.Int` with
`compare.Equal`, `compare.LessOrEqual`, `compare.GreaterOrEqual`,
`compare.LessThan`, or `compare.GreaterThan`.

Double-faced cards use `CardDef` root fields for the front face and
`Back: opt.Val(game.CardFace{...})` for the optional back face. Do not add a
`Faces` slice.

### Ability fields on CardFace

Card source definitions use the **categorized fields** on `CardFace` directly. Do **not** populate the legacy `Abilities []AbilityDef` slice in registered card definitions.

```go
// CardFace categorized ability fields:
SpellAbility      opt.V[SpellAbilityBody]    // instants/sorceries — at most one
ActivatedAbilities []ActivatedAbilityBody   // "[Cost]: [Effect]" abilities
ManaAbilities      []ManaAbilityBody        // mana abilities (subset of activated, no targets)
LoyaltyAbilities   []LoyaltyAbilityBody     // planeswalker +/−/0 abilities
TriggeredAbilities []TriggeredAbilityBody   // When/Whenever/At abilities
ReplacementAbilities []ReplacementAbilityDef // "if … would … instead …" abilities
StaticAbilities    []StaticAbilityBody      // declarative continuous effects, keywords
```

For a card whose abilities all belong to one category, set the field directly:

```go
// Single-category: direct field
ManaAbilities: []game.ManaAbilityBody{
    {
        Text: `
            {T}: Add {G}.
        `,
        // ...
    },
},
```

For a card with **mixed categories**, use an immediately-invoked initializer function
and `append` each ability to the correct field **in oracle-text order**:

```go
// Mixed categories: initializer function preserves oracle order
var KessigWolfRun = func() *game.CardDef {
    card := &game.CardDef{
        ColorIdentity: color.NewIdentity(color.Green, color.Red),
        CardFace: game.CardFace{
            // Mechanical fields...
        },
    }
    card.ManaAbilities = append(card.ManaAbilities, game.ManaAbilityBody{
        // Ability fields...
    })
    card.ActivatedAbilities = append(card.ActivatedAbilities, game.ActivatedAbilityBody{
        // Ability fields...
    })
    return card
}()
```

The `CardFace` struct-field order is: `SpellAbility`, `ActivatedAbilities`, `ManaAbilities`,
`LoyaltyAbilities`, `TriggeredAbilities`, `ReplacementAbilities`, `StaticAbilities`. Because
oracle text commonly prints static/keyword abilities **before** activated/triggered abilities,
you will often need an initializer function.

`AbilityDef` and its `Body` field are a **compatibility view** consumed by existing rules
paths via `CardFace.AbilityDefs()`. Nested granted abilities inside effect data (e.g.
`ContinuousEffect.AddAbilities`, `EmblemAbilities`) still use `AbilityDef{Body: ...}` because
no categorized container exists there; that is intentional and is not the same as a top-level
card-face ability.

Card color identity lives in `mtg/game/color`, not `mtg/game/mana`. Use
`color.NewIdentity(color.Green, color.Red)` for `CardDef.ColorIdentity`.

### AbilityKind

```go
const (
    SpellAbility     AbilityKind = iota  // Instant/sorcery instructions
    ActivatedAbility                      // "[Cost]: [Effect]"
    TriggeredAbility                      // "When/Whenever/At..."
    StaticAbility                         // Declarative continuous effect
)
```

### Keyword (all valid values)

```go
const (
    KeywordNone Keyword = iota
    Deathtouch; Defender; DoubleStrike; FirstStrike; Flash; Flying
    Haste; Hexproof; Indestructible; Lifelink; Menace; Protection
    Reach; Shroud; Trample; Vigilance; Ward; SplitSecond
    Equip; Enchant; Cycling; Flashback; Kicker; Madness; Morph
    Disguise; Convoke; Delve; Suspend; Storm; Cascade; Prowess
    Mutate; Companion; Ninjutsu; Escape; Foretell; Craft; Discover
    Affinity; Improvise; Emerge; Undying; Persist; Wither; Infect
    Toxic; Annihilator; Exalted
)
```

### Instructions and Effect Primitives

Resolving abilities use `[]game.Instruction`. Each instruction contains exactly
one sealed `game.Primitive`:

```go
type Instruction struct {
    Primitive     Primitive
    Condition     opt.V[EffectCondition]
    CardCondition opt.V[CardCondition]
    ResultGate    opt.V[InstructionResultGate]
    Optional      bool
    PublishResult ResultKey
    Description   string
}
```

The supported Card Implementation primitives are:

```text
Damage, Draw, Discard, Destroy, AddMana, AddCounter, MoveCounters,
ApplyContinuous, ApplyRule, ModifyPT, Fight, Tap, Search, Reveal,
PutOnBattlefield, CreateToken, ShufflePermanentIntoLibrary,
StartEngines, SetClassLevel, Monstrosity, DiscoverCards, Pay, Choose
```

Do not author the legacy wide `game.Effect` struct. If no typed primitive
expresses the oracle text, use `ImplementationID` and document the required
hand-written behavior.

Use `game.Fixed(N)` or `game.Dynamic(game.DynamicAmount{...})` for numeric
primitive fields. Use `game.TargetRecipient`, `game.SelectorRecipient`, or
`game.PlayerSelectorRecipient` to construct `Damage.Recipient`. Use
`game.TokenDef` / `game.TokenCopyOf` for `CreateToken.Source`, and
`game.CardBattlefieldSource` / `game.LinkedBattlefieldSource` for
`PutOnBattlefield.Source`.

Sequencing keys have distinct types:

- `ResultKey` — instruction outcomes and result gates
- `ChoiceKey` — values published by `Choose`
- `LinkedKey` — cards or objects published by reveal-like primitives

### CounterSourceSpec

```go
const (
    CounterSourceNone CounterSourceKind = iota
    CounterSourceTarget
    CounterSourceEventPermanent
)

type CounterSourceSpec struct {
    Kind        CounterSourceKind
    TargetIndex int
}
```

### EffectCondition

```go
type EffectCondition struct {
    Text               string
    TargetIndex        int
    MatchPermanentType bool
    PermanentType      types.Card
    Negate             bool
    Condition          opt.V[Condition]
}
```

### DynamicAmount

```go
const (
    DynamicAmountNone DynamicAmountKind = iota
    DynamicAmountConstant
    DynamicAmountX
    DynamicAmountTargetPower
    DynamicAmountTargetToughness
    DynamicAmountTargetManaValue
    DynamicAmountTargetCounters
    DynamicAmountControllerLife
    DynamicAmountControllerHandSize
    DynamicAmountControllerGraveyardSize
    DynamicAmountCountSelector
    DynamicAmountPreviousEffectResult
)

type DynamicAmount struct {
    Kind        DynamicAmountKind
    Constant    int
    Multiplier  int
    TargetIndex int
    CounterKind counter.Kind
    Selector    EffectSelector
    ResultKey   ResultKey
}
```

Set `Instruction.PublishResult` on the producing instruction and use an
`InstructionResultGate` with the same `ResultKey` on later instructions for
"if you do" / "if you don't" branches. `Succeeded` checks whether the previous
primitive actually did anything, so a failed draw from an empty library does
not count as "if you do" (CR 608.2c, CR 101.3).

### ResolutionChoice and ResolutionPayment

```go
const (
    ResolutionChoiceNone ResolutionChoiceKind = iota
    ResolutionChoiceColor
    ResolutionChoiceCardType
    ResolutionChoicePlayer
    ResolutionChoiceCard
)

type ResolutionChoice struct {
    Kind           ResolutionChoiceKind
    Prompt         string
    Player         PlayerID
    UsePlayer      bool
    Colors         []mana.Color                 // Mana colors for mana choices
    CardTypes      []types.Card
    PlayerRelation PlayerRelation
    Zone           zone.Type
}

type ResolutionPayment struct {
    Prompt          string
    ManaCost        opt.V[cost.Mana]
    AdditionalCosts []cost.Additional
    XValue          int
}
```

Use a `Choose` primitive with `PublishChoice` for supported resolution choices.
An `AddMana` primitive consumes a mana choice through `ChoiceFrom`. Use a `Pay`
primitive plus `Instruction.PublishResult` for "you may pay..." during
resolution; follow-up "if you do" instructions use `InstructionResultGate`
with `Accepted: game.TriTrue` and `Succeeded: game.TriTrue`.

### ReplacementEffect

```go
type ReplacementEffect struct {
    Controller       PlayerID
    SourceObjectID   id.ID
    SourceCardID     id.ID
    Description      string
    Duration         EffectDuration
    CreatedTurn      int
    MatchEvent       EventKind
    ControllerFilter TriggerControllerFilter
    MatchFromZone    bool
    FromZone         zone.Type
    MatchToZone      bool
    ToZone           zone.Type
    ReplaceToZone    zone.Type
    EntersTapped     bool
    EntersWithCounters []CounterPlacement
}
```

Use `ReplacementAbilityDef` and the constructors
`EntersTappedReplacement`, `EntersTappedIfReplacement`,
`EntersTappedUnlessPaidReplacement`, and `EntersWithCountersReplacement` for
supported enters-the-battlefield replacement text (CR 614). Other replacement
effects require an `ImplementationID`.

### RuleEffect

```go
const (
    RuleEffectNone RuleEffectKind = iota
    RuleEffectCantGainLife
    RuleEffectCantAttack
    RuleEffectCantBlock
    RuleEffectCostModifier
    RuleEffectCastFromZone
)

type RuleEffect struct {
    Kind               RuleEffectKind
    Controller         PlayerID
    Duration           EffectDuration
    AffectedPlayer     PlayerRelation
    AffectedController ControllerRelation
    PermanentTypes     []types.Card
    SpellTypes         []types.Card
    DefendingPlayer    PlayerRelation
    CostModifier       CostModifier
    CastFromZone       zone.Type
}
```

Put `RuleEffects` directly on `StaticAbilityBody` for static rule-changing text
such as "players can't gain life" (CR 119.6), "creatures can't attack/block"
(CR 506.2, CR 509.1b), "spells cost N more/less", and graveyard-cast
permissions (CR 601.3). Use an `ApplyRule` primitive only when a resolving
ability creates a temporary rule effect.

### TargetSpec

```go
type TargetSpec struct {
    MinTargets int
    MaxTargets int
    Constraint string
    Allow      TargetAllow
    Predicate  TargetPredicate
}

const (
    TargetAllowUnspecified TargetAllow = 0
    TargetAllowPermanent   TargetAllow = 1 << 0
    TargetAllowPlayer      TargetAllow = 1 << 1
    TargetAllowStackObject TargetAllow = 1 << 2
)

type TargetPredicate struct {
    PermanentTypes []types.Card
    ExcludedTypes  []types.Card
    Colors         []color.Color
    ExcludedColors []color.Color
    Controller     ControllerRelation
    Player         PlayerRelation
    Tapped         TriState
    CombatState    CombatStateFilter
    Keyword        Keyword
    ExcludedKeyword Keyword
    ManaValue      opt.V[compare.Int]
    Power          opt.V[compare.Int]
    Toughness      opt.V[compare.Int]
    Another        bool
}
```

Use structured `Allow` and `Predicate` for common constraints such as nonblack,
tapped/untapped, attacking/blocking, mana value, power/toughness, "another",
and "with flying". Keep `Constraint` as human-readable oracle wording. Targets
must be legal when chosen and again on resolution (CR 115, CR 601.2c,
CR 603.3d, CR 608.2b).

### EffectSelector (for mass effects)

```go
const (
    EffectSelectorNone                     EffectSelector = ""
    EffectSelectorAllCreatures             EffectSelector = "all creatures"
    EffectSelectorAllArtifacts             EffectSelector = "all artifacts"
    EffectSelectorAllEnchantments          EffectSelector = "all enchantments"
    EffectSelectorAllNonlandPermanents     EffectSelector = "all nonland permanents"
    EffectSelectorAllPermanents            EffectSelector = "all permanents"
    EffectSelectorCreaturesYouControl      EffectSelector = "creatures you control"
    EffectSelectorOtherCreaturesYouControl EffectSelector = "other creatures you control"
)
```

### TriggerCondition

```go
type TriggerCondition struct {
    Type          TriggerType     // TriggerWhen, TriggerWhenever, or TriggerAt
    Pattern       TriggerPattern  // Structured event pattern
    InterveningIf string          // "if" condition (CR 603.4)
    InterveningIfControllerLifeAtLeast int
    InterveningIfEventPermanentHadCounters bool
    State         *StateTriggerCondition
}

type StateTriggerCondition struct {
    MatchControllerLifeLessOrEqual bool
    ControllerLifeLessOrEqual      int
}
```

### TriggerPattern

```go
type TriggerPattern struct {
    Event      EventKind
    Controller TriggerControllerFilter  // TriggerControllerAny/You/Opponent
    Source     TriggerSourceFilter      // TriggerSourceAny/Self
    ExcludeSelf bool
    Player     TriggerPlayerFilter      // TriggerPlayerAny/You/Opponent
    RequirePermanentTypes []types.Card
    ExcludePermanentTypes []types.Card
    RequireCardTypes []types.Card
    ExcludeCardTypes []types.Card
    MatchFromZone bool
    FromZone      zone.Type
    MatchToZone   bool
    ToZone        zone.Type
    DamageRecipient DamageRecipientKind
    Step Step
}
```

Use `RequireCardTypes` / `ExcludeCardTypes` for cast triggers such as
"Whenever an opponent casts a noncreature spell" (CR 603.2). Use
`RequirePermanentTypes` / `ExcludePermanentTypes` for ETB/LTB/dies triggers;
LTB and dies triggers use last-known information (CR 603.10). Use explicit
`Step` with `EventBeginningOfStep` for "At the beginning of your upkeep/draw
step/beginning of combat/end step" (CR 603.6c); broad beginning-of-step
patterns with `StepNone` do not match. Use `State` for state triggers; the
rules engine latches them until the condition becomes false (CR 603.8).

### EventKind (for trigger patterns)

```go
const (
    EventUnknown EventKind = iota
    EventSpellCast; EventSpellResolved
    EventPermanentEnteredBattlefield; EventPermanentDied
    EventDamageDealt; EventCardDrawn; EventZoneChanged
    EventAttackerDeclared; EventBlockerDeclared; EventCardDiscarded
    EventDamagePrevented; EventDestroyReplaced
    EventBeginningOfStep
    EventLifeGained; EventLifeLost
    EventPermanentTapped; EventPermanentUntapped
    EventObjectBecameTarget
)
```

Use `EventZoneChanged` for "leaves [zone]" triggers by setting `FromZone`; set
`ToZone` as well when the destination matters. Use `EventLifeGained` and
`EventLifeLost` for life-total-change triggers; player damage also emits
`EventLifeLost` in addition to `EventDamageDealt`. Use `EventPermanentTapped`
and `EventPermanentUntapped` for tap/untap triggers. Use
`EventObjectBecameTarget` for "becomes the target of..." triggers; the event's
`Target` field identifies whether a permanent, player, or stack object became
the target.

### Mana Cost construction helpers

```go
cost.W    // {W}
cost.U     // {U}
cost.B    // {B}
cost.R      // {R}
cost.G    // {G}
cost.O(3)             // {3}
cost.C            // {C}
cost.X             // {X}
cost.HybridMana(mana.W, mana.U)  // {W/U}
cost.Twobrid(mana.W)             // {2/W}
cost.PhyrexianMana(mana.W)          // {W/P}
cost.S                         // {S}
```

---

## Classification Rules

### Step 1: Split oracle text into paragraphs

Each paragraph (separated by `\n`) is one ability. Exception: a comma-separated list of plain keywords on a single line counts as multiple keyword abilities, one `AbilityDef` per keyword.

### Step 2: Classify each paragraph

| Test | Kind |
|------|------|
| Contains `:` outside of `{...}` braces | `ActivatedAbility` |
| Starts with `When`, `Whenever`, or `At` | `TriggeredAbility` |
| On an instant or sorcery and not activated/triggered | `SpellAbility` |
| Otherwise | `StaticAbility` |

### Step 3: Extract fields per ability kind

#### Keywords (any ability kind)

If a paragraph is just one or more plain non-parameterized keywords, use the reusable typed
`StaticAbilityBody` templates for each keyword:
- `game.FlyingStaticBody`
- `game.DeathtouchStaticBody`
- `game.IndestructibleStaticBody`
- etc.

For comma-separated keyword lines such as `Deathtouch, indestructible`, add each template separately:

```go
// Single category — direct slice:
StaticAbilities: []game.StaticAbilityBody{
    game.DeathtouchStaticBody,
    game.IndestructibleStaticBody,
},
```

When keywords appear on a card that also has activated or triggered abilities, use an initializer function
and `append` in oracle order:

```go
// Mixed categories — initializer preserves oracle order:
var MyCard = func() *game.CardDef {
    card := &game.CardDef{/* ... */}
    card.StaticAbilities = append(card.StaticAbilities, game.FlyingStaticBody)
    card.TriggeredAbilities = append(card.TriggeredAbilities, game.TriggeredAbilityBody{/* ... */})
    return card
}()
```

Do not smash multiple plain keywords into one ad-hoc body; use `game.SimpleKeywords(...)` only when a helper template is not suitable. Use explicit body structs for keyword abilities that need card-specific parameters or costs.

For keywords with parameters, use the typed field on a `StaticAbilityBody` (or the appropriate body type):
- **Protection from [color]**: `StaticAbilityBody{KeywordAbilities: []game.KeywordAbility{game.ProtectionKeyword{FromColors: []color.Color{color.Red}}}}`
- **Ward {N}**: `StaticAbilityBody{KeywordAbilities: []game.KeywordAbility{game.WardKeyword{Cost: cost.Mana{cost.O(N)}}}}`
- **Equip {N}**: `ActivatedAbilityBody{Text: "Equip {N}", ManaCost: opt.Val(cost.Mana{cost.O(N)}), Timing: game.SorceryOnly, KeywordAbilities: []game.KeywordAbility{game.EquipKeyword{Cost: cost.Mana{cost.O(N)}}}, Content: game.PlainAbilityContent{Targets: []game.TargetSpec{{...}}}}`
- **Cycling {N}**: `ActivatedAbilityBody{Text: "Cycling {N}", ManaCost: opt.Val(cost.Mana{...}), AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalDiscard, Text: "Discard this card", Source: zone.Hand}}, KeywordAbilities: []game.KeywordAbility{game.CyclingKeyword{Cost: cost.Mana{...}}}}`
- **Prowess**: `StaticAbilityBody{KeywordAbilities: game.SimpleKeywords(game.Prowess)}`; the rules engine creates the implicit trigger (CR 702.108)
- **Flashback {cost}**: `StaticAbilityBody{KeywordAbilities: game.SimpleKeywords(game.Flashback)}` plus a spell `cost.Alternative{Label: "Flashback", ManaCost: ...}`; flashback costs are usable only from graveyard and exile the spell when it leaves the stack (CR 702.34)

#### Spell abilities (instants/sorceries)

Set `SpellAbility: opt.Val(game.SpellAbilityBody{...})`:
- `Text:` full oracle text
- `Content`: `PlainAbilityContent{Targets: [...], Sequence: [...]}` or `ModalAbilityContent{Modes: [...]}`
- Extract `Targets` from "target [constraint]" phrases
- Extract typed `Instruction` primitives from the action verbs (see Primitive Mapping below)
- For modal text (`Choose one —`, `Choose two —`, `Choose one or both —`,
  `Choose up to one —`), fill `Modes` and set `MinModes` / `MaxModes` from the
  choice count. Leave `MinModes` and `MaxModes` at zero only for legacy
  choose-one mode lists.

#### Activated abilities

Use `game.ActivatedAbilityBody{...}` in `ActivatedAbilities`, or
`game.ManaAbilityBody{...}` in `ManaAbilities` when the effect adds mana with no targets:
- Split on `:` — left side is costs, right side is effects
- Parse mana symbols in cost → `ManaCost`
- Parse non-mana costs into `[]cost.Additional` (e.g., `{T}` = `cost.AdditionalTap`, "Sacrifice a creature", "Pay 2 life")
- Use `ManaAbilityBody` / `ManaAbilities` if the effect adds mana, has no targets, and is not a loyalty ability
- `Timing`: set if "Activate only as a sorcery" or "Activate only once each turn"

#### Triggered abilities

Use `game.TriggeredAbilityBody{...}` in `TriggeredAbilities`:
- Parse the trigger word → `Trigger.Type` (TriggerWhen/TriggerWhenever/TriggerAt)
- Parse the event → `Trigger.Pattern`
- Common patterns:
  - "enters" / "enters the battlefield" → `EventPermanentEnteredBattlefield`
  - "dies" → `EventPermanentDied`
  - "leaves the battlefield" → `EventZoneChanged` with `FromZone: zone.Battlefield`
  - "leaves your graveyard/hand/library/exile/command zone" → `EventZoneChanged` with the matching `FromZone`
  - "is put into [zone] from [zone]" → `EventZoneChanged` with both `FromZone` and `ToZone`
  - "At the beginning of your upkeep/draw step/beginning of combat/end step" → `EventBeginningOfStep` with explicit `Step`
  - "Whenever ... attacks" → `EventAttackerDeclared`
  - "Whenever ... blocks" / "becomes blocked by ..." → `EventBlockerDeclared`
  - "Whenever ... deals damage" → `EventDamageDealt`
  - "Whenever you gain life" → `EventLifeGained`
  - "Whenever an opponent loses life" → `EventLifeLost`
  - "Whenever ... becomes tapped" → `EventPermanentTapped`
  - "Whenever ... becomes untapped" → `EventPermanentUntapped`
  - "Whenever ... becomes the target of..." → `EventObjectBecameTarget`
  - "Whenever ... is cast" → `EventSpellCast`
- Set controller/source filters based on "you", "an opponent", "another creature", "this creature"; use `ExcludeSelf: true` for "another" trigger wording.
- If "you may" appears → `Optional: true`
- For "At the beginning of your upkeep/draw step/beginning of combat/end step",
  use `EventBeginningOfStep` with `Step: game.StepUpkeep`, `game.StepDraw`,
  `game.StepBeginningOfCombat`, or `game.StepEnd`.
- For cast triggers such as "Whenever an opponent casts a noncreature spell",
  use `EventSpellCast`, `Controller: game.TriggerControllerOpponent`, and
  `ExcludeCardTypes: []types.Card{types.Creature}`.
- For state triggers, set `Trigger.Type: game.TriggerState` and fill
  `Trigger.State`; do not set an event pattern.

#### Static abilities

Use `game.StaticAbilityBody{...}` in `StaticAbilities`:
- Common patterns:
  - "Creatures you control get +N/+M" → `ContinuousEffects` with `LayerPowerToughnessModify` and `EffectSelectorCreaturesYouControl`
  - "Other creatures you control get +N/+M" → same with `EffectSelectorOtherCreaturesYouControl`
  - "Players can't gain life" → `RuleEffects` with `RuleEffectCantGainLife`
  - "Creatures can't attack/block" → `RuleEffects` with `RuleEffectCantAttack` / `RuleEffectCantBlock`
  - "Spells cost N more/less" → `RuleEffects` with `RuleEffectCostModifier`
  - "You may cast ... from your graveyard" → `RuleEffects` with `RuleEffectCastFromZone`

Static abilities do not use `Sequence`: continuous and rule effects are
declarations that apply while the ability functions, not resolving
instructions.

---

## Primitive Mapping

| Oracle text pattern | Primitive | Notes |
|---------------------|-----------|-------|
| "deals N damage to" | `Damage` | `Amount: game.Fixed(N)`, `Recipient: game.TargetRecipient(index)` |
| "deals X damage to" | `Damage` | `Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})` |
| "deals damage equal to [target]'s power" | `Damage` | Use `DynamicAmountTargetPower` with the target index |
| "that much" | quantity-bearing primitive | Publish a `ResultKey`; consume it with `DynamicAmount.ResultKey` |
| "you may [do X]. If you do, [Y]" | any primitives | Set `Optional` and `PublishResult` on X; set Y's `ResultGate` to accepted and succeeded |
| "if you don't" | any primitive | Gate on the producer's `ResultKey` with `Accepted: game.TriFalse` |
| "choose a color" | `Choose` | Publish a `ChoiceKey`; `AddMana.ChoiceFrom` consumes it |
| "you may pay [cost]. If you do..." | `Pay` | Publish a `ResultKey`; gate the following instruction |
| "destroy target" | `Destroy` | `TargetIndex` follows target order |
| "draw N card(s)" | `Draw` | `Amount: game.Fixed(N)` |
| "discard N card(s)" | `Discard` | `Amount: game.Fixed(N)` |
| "add {C}{C}" / "add {G}" | `AddMana` | Set `Amount` and `ManaColor` |
| "gets +N/+M" | `ModifyPT` | Use fixed or dynamic `PowerDelta` and `ToughnessDelta` quantities |
| "put N +1/+1 counter(s)" | `AddCounter` | Set `Amount` and `CounterKind` |
| "move counters from ... onto ..." | `MoveCounters` | Set a typed `CounterSourceSpec` and destination `TargetIndex` |
| "becomes a N/M [subtype] creature" | `ApplyContinuous` | Add layer-specific `ContinuousEffects` and a duration |
| "create a N/M token" | `CreateToken` | `Source: game.TokenDef(tokenDef)` |
| "tap target" | `Tap` | Set `TargetIndex` |
| "fight" | `Fight` | Set `TargetIndex` and optional `RelatedTargetIndex` |
| "search your library" | `Search` | Set `SearchSpec`, amount, and player target |
| "reveal the top card" | `Reveal` | Publish a `LinkedKey` when a later instruction consumes the card |
| "put [that card] onto the battlefield" | `PutOnBattlefield` | Use `CardBattlefieldSource` or `LinkedBattlefieldSource` |
| "discover N" | `DiscoverCards` | `Amount: game.Fixed(N)` |

Oracle patterns not represented by the listed primitives require an
`ImplementationID`; do not fall back to `game.Effect`.

### TargetIndex convention

- `TargetIndex: 0` = first target declared
- `TargetIndex: 1` = second target declared
- `TargetIndex: -1` = the ability's controller (for "you draw", "you gain life")
- `TargetIndex: -2` = the source permanent (used internally by Prowess-style source effects)

### TargetSpec.Constraint values

Use natural-language descriptions matching the oracle text:
- `"any target"` — creature, player, or planeswalker
- `"creature"` — target creature
- `"creature or planeswalker"` — target creature or planeswalker
- `"player"` — target player
- `"artifact"` — target artifact
- `"enchantment"` — target enchantment
- `"permanent"` — target permanent
- `"creature or player"` — target creature or player

---

## Worked Examples

### Example 1: Lightning Bolt (simple targeted spell)

Oracle text: `Lightning Bolt deals 3 damage to any target.`

```go
SpellAbility: opt.Val(game.SpellAbilityBody{
    Text: `
        Lightning Bolt deals 3 damage to any target.
    `,
    Content: game.PlainAbilityContent{
        Targets: []game.TargetSpec{
            {MinTargets: 1, MaxTargets: 1, Constraint: "any target"},
        },
        Sequence: []game.Instruction{
            {
                Primitive: game.Damage{
                    Amount:    game.Fixed(3),
                    Recipient: game.TargetRecipient(0),
                },
            },
        },
    },
}),
```

### Example 2: Sol Ring (mana ability)

Oracle text: `{T}: Add {C}{C}.`

```go
ManaAbilities: []game.ManaAbilityBody{
    {
        Text: `
            {T}: Add {C}{C}.
        `,
        AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
        Content: game.PlainAbilityContent{
            Sequence: []game.Instruction{
                {
                    Primitive: game.AddMana{
                        Amount:    game.Fixed(2),
                        ManaColor: mana.Colorless,
                    },
                },
            },
        },
    },
},
```

### Example 3: Serra Angel (keyword abilities — same category, direct slice)

Oracle text:
```
Flying
Vigilance (Attacking doesn't cause this creature to tap.)
```

```go
StaticAbilities: []game.StaticAbilityBody{
    game.FlyingStaticBody,
    game.VigilanceStaticBody,
},
```

### Example 4: Swords to Plowshares (targeted removal with controller effect)

Oracle text: `Exile target creature. Its controller gains life equal to its power.`

This card cannot currently be expressed by the typed primitive set because
there is no exile primitive or life-gain primitive. Set a descriptive
`ImplementationID` and document both operations for the hand-written resolver.

### Example 5: Soul Warden (triggered ability)

Oracle text: `Whenever another creature enters, you gain 1 life.`

The trigger pattern is expressible, but life gain is not yet a typed primitive.
Set `ImplementationID` rather than authoring a legacy `game.Effect`. When a
typed life-gain primitive is added, use `ExcludeSelf: true` for "another."

### Example 6: Glorious Anthem (static anthem effect)

Oracle text: `Creatures you control get +1/+1.`

```go
StaticAbilities: []game.StaticAbilityBody{
    {
        Text: `
            Creatures you control get +1/+1.
        `,
        ContinuousEffects: []game.ContinuousEffect{
            {
                Layer:          game.LayerPowerToughnessModify,
                Selector:       game.EffectSelectorCreaturesYouControl,
                PowerDelta:     1,
                ToughnessDelta: 1,
            },
        },
    },
},
```

### Example 7: Kessig Wolf Run (mixed categories — initializer function)

Oracle text:
```
{T}: Add {R} or {G}.
{X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.
```

Struct field order (ActivatedAbilities before ManaAbilities) differs from oracle order,
so use an initializer function:

```go
var KessigWolfRun = func() *game.CardDef {
    card := &game.CardDef{
        ColorIdentity: color.NewIdentity(color.Red, color.Green),
        CardFace: game.CardFace{
            Name:  "Kessig Wolf Run",
            Types: []types.Card{types.Land},
            OracleText: `
                {T}: Add {R} or {G}.
                {X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.
            `,
        },
    }
    // oracle order: mana ability first, then activated ability
    card.ManaAbilities = append(card.ManaAbilities, game.ManaAbilityBody{
        Text: `
            {T}: Add {R} or {G}.
        `,
        AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
        // ...
    })
    card.ActivatedAbilities = append(card.ActivatedAbilities, game.ActivatedAbilityBody{
        Text: `
            {X}{R}{G}, {T}: Target creature gets +X/+0 and gains trample until end of turn.
        `,
        // ...
    })
    return card
}()
```

---

## Common Pitfalls

1. **Self-referencing names**: When oracle text says the card's own name (e.g., "Lightning Bolt deals 3 damage"), it means "this spell/permanent". Don't create a target for the card itself.

2. **Reminder text**: Text in parentheses like "(Attacking doesn't cause this creature to tap.)" is reminder text — it's not rules text. Ignore it when parsing abilities. The keyword itself carries the rules meaning.

3. **"Any target"**: This means target creature, player, planeswalker, or battle (CR 115.4). Use `Constraint: "any target"` or `Allow: game.TargetAllowPermanent | game.TargetAllowPlayer`.

4. **"You" as target vs. controller**: When the text says "you gain 1 life", "you" is the controller, not a target. Use `TargetIndex: -1`. Only use `TargetIndex: 0+` when there's an explicit "target" word.

5. **Tap symbol in costs**: `{T}` in an activated ability cost means the permanent taps itself. This goes in `AdditionalCosts` as `{Kind: cost.AdditionalTap}`, not in `ManaCost`.

6. **Multiple paragraphs = multiple abilities**: Each `\n`-separated paragraph is a separate ability and gets its own body in the appropriate field, unless it's a comma-separated keyword list.

7. **Variable amounts**: When an effect says "equal to its power" or "equal to the number of...", the current `Effect.Amount` field can't express this. Use `Description` to document it, and consider setting `ImplementationID` if the card needs full rules accuracy.

8. **Cards that need ImplementationID**: If a card does things the declarative system can't express (unsupported choice consumers, ETB-as-copy/as-choice, full replacement ordering, or complex conditional logic), set `ImplementationID` to a descriptive kebab-case string like `"swords-to-plowshares"` and leave a comment explaining what the hand-written code needs to do.
