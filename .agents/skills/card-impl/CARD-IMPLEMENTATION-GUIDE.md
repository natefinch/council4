# Card Implementation Guide

Reference for parsing Magic: The Gathering oracle text into council4 `game.AbilityDef` structs.

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

### AbilityDef

```go
type AbilityDef struct {
    Text             string
    KeywordAbilities []KeywordAbility // Sealed keyword variants this provides.
    Body             AbilityBody       // Spell, activated, mana, loyalty, triggered, replacement, or static body.

    // Legacy flat fields remain for compatibility while rules consumers migrate.
    Kind    AbilityKind
    Effects []Effect
    Targets []TargetSpec
    Modes   []Mode
}
```

Prefer the categorized `CardFace` fields (`SpellAbility`, `ActivatedAbilities`,
`ManaAbilities`, `LoyaltyAbilities`, `TriggeredAbilities`,
`ReplacementAbilities`, and `StaticAbilities`) with explicit body structs. Legacy
`Abilities` literals are still accepted; `CardFace.AbilityDefs()` normalizes them
to body-backed `AbilityDef` values for rules code.

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

### EffectType (all valid values)

```go
const (
    EffectUnknown EffectType = iota
    EffectDamage; EffectDestroy; EffectExile; EffectBounce; EffectCounter
    EffectDraw; EffectDiscard; EffectMill; EffectSearch; EffectCreateToken
    EffectGainLife; EffectLoseLife; EffectAddMana; EffectModifyPT
    EffectAddCounter; EffectRemoveCounter; EffectPutOnBattlefield
    EffectSacrifice; EffectTap; EffectUntap; EffectGainControl; EffectCopy
    EffectScry; EffectSurveil; EffectFight; EffectTransform; EffectAttach
    EffectReplace; EffectPrevent; EffectCreateDelayedTrigger
    EffectRegenerate; EffectSkipStep; EffectPhaseOut; EffectCreateEmblem
    EffectApplyContinuous; EffectMoveCounters; EffectChoose; EffectPay
    EffectApplyRule; EffectProliferate; EffectGoad
)
```

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
    LinkID      string
}
```

### Effect

```go
type Effect struct {
    Type            EffectType
    Amount          int                        // Numeric amount (damage, cards drawn, etc.)
    DynamicAmount   opt.V[DynamicAmount]       // Amount determined on resolution (CR 107.3, CR 608.2c)
    TargetIndex     int                        // Index into runtime targets; -1 = controller
    Optional        bool                       // Ask whether to apply this single instruction (CR 608.2c)
    ResultCondition opt.V[EffectResultCondition] // Gate on prior linked effect result
    Condition       opt.V[EffectCondition]
    PowerDelta      int                        // For EffectModifyPT
    ToughnessDelta  int                        // For EffectModifyPT
    CounterKind     counter.Kind               // For EffectAddCounter/EffectRemoveCounter
    CounterSource   CounterSourceSpec          // For EffectMoveCounters
    ManaColor       mana.Color                 // For EffectAddMana
    Choice          opt.V[ResolutionChoice]    // Value chosen during resolution (CR 608.2c, CR 609.3)
    ChoiceLinkID    string                     // Consume a prior choice value
    Payment         opt.V[ResolutionPayment]   // Optional "you may pay..." during resolution (CR 608.2c, CR 117.12)
    UntilEndOfTurn  bool                       // Duration flag
    Duration        EffectDuration
    Step            Step                       // For step-related effects
    Selector        EffectSelector             // For mass effects
    Token           opt.V[*CardDef]            // For EffectCreateToken
    ContinuousEffects []ContinuousEffect       // For EffectApplyContinuous
    DelayedTrigger  opt.V[DelayedTriggerDef]
    EmblemAbilities []AbilityDef
    Replacement     opt.V[ReplacementEffect]   // For EffectReplace
    RuleEffects     []RuleEffect               // For EffectApplyRule
    LinkID          string
    Description     string                     // Human-readable description
}
```

### EffectResultCondition

```go
type EffectResultCondition struct {
    LinkID    string
    Accepted  TriState
    Succeeded TriState
}
```

Use `LinkID` on an earlier effect and `ResultCondition` on later effects for
"if you do" / "if you don't" branches. `Succeeded` checks whether the previous
effect actually did anything, so a failed draw from an empty library does not
count as "if you do" (CR 608.2c, CR 101.3).

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
    Zone           ZoneType
}

type ResolutionPayment struct {
    Prompt          string
    ManaCost        opt.V[cost.Mana]
    AdditionalCosts []AdditionalCost
    XValue          int
}
```

Use `EffectChoose` with `Choice` and `LinkID` for "choose a color/card
type/player/card" instructions. Later effects can consume a chosen color for
`EffectAddMana` or a chosen player for player effects by setting
`ChoiceLinkID`. Use `EffectPay` with `Payment` and `LinkID` for "you may pay..."
during resolution; follow-up "if you do" effects should use `ResultCondition`
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
    FromZone         ZoneType
    MatchToZone      bool
    ToZone           ZoneType
    ReplaceToZone    ZoneType
    EntersTapped     bool
    EntersWithCounters []CounterPlacement
}
```

Use `EffectReplace` with `Replacement` to create runtime replacement effects
for zone-change destination replacement and simple enters-the-battlefield
modifiers (CR 614). The generic replacement slice applies each matching effect
at most once to the event and records deterministic fallback ordering when
multiple generic replacements apply (CR 614.5, CR 616). ETB-as-copy,
ETB-as-choice, and full APNAP replacement ordering are still follow-ups.

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
    CastFromZone       ZoneType
}
```

Use `EffectApplyRule` with `RuleEffects` for static rule-changing text such as
"players can't gain life" (CR 119.6), "creatures can't attack/block" (CR 506.2,
CR 509.1b), "spells cost N more/less", and "you may cast [cards] from your
graveyard" permissions (CR 601.3). Rule effects in static abilities are derived
while their source is on the battlefield; rule effects created by resolving an
effect can use normal duration fields.

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
    FromZone      ZoneType
    MatchToZone   bool
    ToZone        ZoneType
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

If a paragraph is just one or more plain non-parameterized keywords, use the reusable helper ability for each keyword:
- `game.FlyingAbility`
- `game.DeathtouchAbility`
- `game.IndestructibleAbility`
- etc.

For comma-separated keyword lines such as `Deathtouch, indestructible`, add each helper separately:

```go
Abilities: []game.AbilityDef{
    game.DeathtouchAbility,
    game.IndestructibleAbility,
}
```

Do not smash multiple plain keywords into one ad-hoc `AbilityDef`; use `game.SimpleKeywords(...)` only when a helper template is not suitable. Use explicit `AbilityDef` values for keyword abilities that need card-specific parameters or costs.

For keywords with parameters:
- **Protection from [color]**: `KeywordAbilities: []game.KeywordAbility{game.ProtectionKeyword{FromColors: []color.Color{color.Red}}}`
- **Ward {N}**: `KeywordAbilities: []game.KeywordAbility{game.WardKeyword{Cost: cost.Mana{cost.O(N)}}}`
- **Equip {N}**: `Kind: ActivatedAbility`, `KeywordAbilities: []game.KeywordAbility{game.EquipKeyword{Cost: cost.Mana{cost.O(N)}}}`, `ManaCost: opt.Val(cost.Mana{cost.O(N)})`, `Timing: game.SorceryOnly`
- **Cycling {N}**: `Kind: ActivatedAbility`, `KeywordAbilities: []game.KeywordAbility{game.CyclingKeyword{Cost: cost.Mana{...}}}`, `ManaCost: opt.Val(cost.Mana{...})`, `AdditionalCosts` with discard self
- **Prowess**: `KeywordAbilities: game.SimpleKeywords(game.Prowess)` on a static ability; the rules engine creates the implicit trigger (CR 702.108)
- **Flashback {cost}**: `KeywordAbilities: game.SimpleKeywords(game.Flashback)` plus a spell `AlternativeCost{Label: "Flashback", ManaCost: ...}`; flashback costs are usable only from graveyard and exile the spell when it leaves the stack (CR 702.34)

#### Spell abilities (instants/sorceries)

- `Kind: SpellAbility`
- `Text:` full oracle text
- Extract `Targets` from "target [constraint]" phrases
- Extract `Effects` from the action verbs (see Effect Mapping below)
- For modal text (`Choose one —`, `Choose two —`, `Choose one or both —`,
  `Choose up to one —`), fill `Modes` and set `MinModes` / `MaxModes` from the
  choice count. Leave `MinModes` and `MaxModes` at zero only for legacy
  choose-one mode lists.

#### Activated abilities

- `Kind: ActivatedAbility`
- Split on `:` — left side is costs, right side is effects
- Parse mana symbols in cost → `ManaCost`
- Parse non-mana costs → `AdditionalCosts` (e.g., `{T}` = tap, "Sacrifice a creature", "Pay 2 life")
- `IsManaAbility: true` if the effect adds mana, has no targets, and is not a loyalty ability
- `Timing`: set if "Activate only as a sorcery" or "Activate only once each turn"

#### Triggered abilities

- `Kind: TriggeredAbility`
- Parse the trigger word → `Trigger.Type` (TriggerWhen/TriggerWhenever/TriggerAt)
- Parse the event → `Trigger.Pattern`
- Common patterns:
  - "enters" / "enters the battlefield" → `EventPermanentEnteredBattlefield`
  - "dies" → `EventPermanentDied`
  - "leaves the battlefield" → `EventZoneChanged` with `FromZone: game.ZoneBattlefield`
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

- `Kind: StaticAbility`
- Common patterns:
  - "Creatures you control get +N/+M" → `Effects` with `EffectModifyPT`, `Selector: game.EffectSelectorCreaturesYouControl`
  - "Other creatures you control get +N/+M" → same with `EffectSelectorOtherCreaturesYouControl`
  - "Players can't gain life" → `EffectApplyRule` with `RuleEffectCantGainLife`
  - "Creatures can't attack/block" → `EffectApplyRule` with `RuleEffectCantAttack` / `RuleEffectCantBlock`
  - "Spells cost N more/less" → `EffectApplyRule` with `RuleEffectCostModifier`
  - "You may cast ... from your graveyard" → `EffectApplyRule` with `RuleEffectCastFromZone`

---

## Effect Mapping

| Oracle text pattern | EffectType | Notes |
|---------------------|------------|-------|
| "deals N damage to" | `EffectDamage` | `Amount: N`, set `TargetIndex` |
| "deals X damage to" | `EffectDamage` | `DynamicAmount: opt.Val(game.DynamicAmount{Kind: game.DynamicAmountX})` |
| "deals damage equal to [target]'s power" | `EffectDamage` | `DynamicAmount: opt.Val(game.DynamicAmount{Kind: game.DynamicAmountTargetPower, TargetIndex: N})` |
| "that much" | any amount effect | Use `LinkID` on the producing effect and `DynamicAmountPreviousEffectResult` on the consuming effect |
| "you may [do X]. If you do, [Y]" | any effect(s) | Put `Optional: true` and `LinkID` on X; put `ResultCondition` with `Accepted: game.TriTrue`, `Succeeded: game.TriTrue` on Y |
| "if you don't" | any effect | Put `ResultCondition` with `Accepted: game.TriFalse` on the branch effect |
| "choose a color/player/card type/card" | `EffectChoose` | Set `Choice` and `LinkID`; later effects consume with `ChoiceLinkID` where supported |
| "you may pay [cost]. If you do..." | `EffectPay` | Set `Payment` and `LinkID`; gate the branch with `ResultCondition` |
| "destroy target" | `EffectDestroy` | `TargetIndex` from target order |
| "exile target" | `EffectExile` | |
| "return target ... to its owner's hand" | `EffectBounce` | |
| "draw N card(s)" | `EffectDraw` | `Amount: N` |
| "discard N card(s)" | `EffectDiscard` | `Amount: N` |
| "gain(s) N life" | `EffectGainLife` | `Amount: N` |
| "lose(s) N life" | `EffectLoseLife` | `Amount: N` |
| "add {C}{C}" / "add {G}" | `EffectAddMana` | `Amount: N`, `ManaColor` |
| "gets +N/+M" | `EffectModifyPT` | `PowerDelta: N`, `ToughnessDelta: M` |
| "put N +1/+1 counter(s)" | `EffectAddCounter` | `Amount: N`, `CounterKind: counter.PlusOnePlusOne` |
| "move counters from target ... onto target ..." | `EffectMoveCounters` | `CounterSource: CounterSourceSpec{Kind: CounterSourceTarget, TargetIndex: sourceIndex}`, `TargetIndex: destinationIndex` |
| "put those counters on ..." from a triggered zone-change object | `EffectMoveCounters` | `CounterSource: CounterSourceSpec{Kind: CounterSourceEventPermanent}` reads current/LKI counters from the event permanent |
| "becomes a N/M [subtype] creature in addition to its other types" | `EffectApplyContinuous` | Add `ContinuousEffects` entries for `LayerType` and `LayerPowerToughnessSet` |
| "create a N/M token" | `EffectCreateToken` | Set `Token` field |
| "sacrifice" (as effect) | `EffectSacrifice` | |
| "tap target" | `EffectTap` | |
| "untap target" | `EffectUntap` | |
| "scry N" | `EffectScry` | `Amount: N` |
| "surveil N" | `EffectSurveil` | `Amount: N` |
| "mill N" | `EffectMill` | `Amount: N` |
| "fight" | `EffectFight` | Uses `TargetIndex` and `RelatedTargetIndex`; bare fight effects default to targets 0 and 1 |
| "if [zone change] would happen, instead..." | `EffectReplace` | Set `Replacement` with match zones and `ReplaceToZone` |
| "enters tapped / with counters" as a runtime effect | `EffectReplace` | Set `Replacement.EntersTapped` / `EntersWithCounters` |
| "players can't gain life" | `EffectApplyRule` | Add `RuleEffectCantGainLife` |
| "creatures can't attack/block" | `EffectApplyRule` | Add `RuleEffectCantAttack` / `RuleEffectCantBlock` with controller/type filters |
| "spells cost N more/less" | `EffectApplyRule` | Add `RuleEffectCostModifier` with `CostModifier` |
| "you may cast ... from your graveyard" | `EffectApplyRule` | Add `RuleEffectCastFromZone` with `CastFromZone: game.ZoneGraveyard` |
| "proliferate" | `EffectProliferate` | Chooses one existing counter kind per eligible permanent/player (CR 701.27) |
| "goad target creature" | `EffectGoad` | `TargetIndex` points at the target creature; expires on goading player's next turn (CR 701.38) |
| "counter target spell" | `EffectCounter` | |

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
Abilities: []game.AbilityDef{
    {
        Kind: game.SpellAbility,
        Text: "Lightning Bolt deals 3 damage to any target.",
        Targets: []game.TargetSpec{
            {MinTargets: 1, MaxTargets: 1, Constraint: "any target"},
        },
        Effects: []game.Effect{
            {Type: game.EffectDamage, Amount: 3, TargetIndex: 0},
        },
    },
},
```

### Example 2: Sol Ring (mana ability)

Oracle text: `{T}: Add {C}{C}.`

```go
Abilities: []game.AbilityDef{
    {
        Kind:         game.ActivatedAbility,
        Text:         "{T}: Add {C}{C}.",
        IsManaAbility: true,
        AdditionalCosts: []game.AdditionalCost{
            {Kind: game.CostTap},
        },
        Effects: []game.Effect{
            {Type: game.EffectAddMana, Amount: 2, ManaColor: mana.Colorless},
        },
    },
},
```

### Example 3: Serra Angel (keyword abilities)

Oracle text:
```
Flying
Vigilance (Attacking doesn't cause this creature to tap.)
```

```go
Abilities: []game.AbilityDef{
    game.FlyingAbility,
    game.VigilanceAbility,
},
```

### Example 4: Swords to Plowshares (targeted removal with controller effect)

Oracle text: `Exile target creature. Its controller gains life equal to its power.`

```go
Abilities: []game.AbilityDef{
    {
        Kind: game.SpellAbility,
        Text: "Exile target creature. Its controller gains life equal to its power.",
        Targets: []game.TargetSpec{
            {MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
        },
        Effects: []game.Effect{
            {Type: game.EffectExile, TargetIndex: 0},
            {Type: game.EffectGainLife, TargetIndex: 0, Description: "controller gains life equal to creature's power"},
        },
    },
},
```

Note: Swords to Plowshares needs both a dynamic amount
(`DynamicAmountTargetPower`) and "that permanent's controller" as the life-gain
recipient. The dynamic amount is supported, but target-controller-as-recipient
still needs either a future recipient primitive or `ImplementationID`.

### Example 5: Soul Warden (triggered ability)

Oracle text: `Whenever another creature enters, you gain 1 life.`

```go
Abilities: []game.AbilityDef{
    {
        Kind: game.TriggeredAbility,
        Text: "Whenever another creature enters, you gain 1 life.",
        Trigger: opt.Val(game.TriggerCondition{
            Type: game.TriggerWhenever,
            Pattern: game.TriggerPattern{
                Event:                 game.EventPermanentEnteredBattlefield,
                Source:                game.TriggerSourceAny,
                ExcludeSelf:           true,
                RequirePermanentTypes: []types.Card{types.Creature},
            },
        }),
        Effects: []game.Effect{
            {Type: game.EffectGainLife, Amount: 1, TargetIndex: -1},
        },
    },
},
```

Note: "another creature" means the trigger should not fire for Soul Warden itself entering; `ExcludeSelf` handles that source/event comparison.

### Example 6: Glorious Anthem (static anthem effect)

Oracle text: `Creatures you control get +1/+1.`

```go
Abilities: []game.AbilityDef{
    {
        Kind: game.StaticAbility,
        Text: "Creatures you control get +1/+1.",
        Effects: []game.Effect{
            {
                Type:           game.EffectModifyPT,
                PowerDelta:     1,
                ToughnessDelta: 1,
                Selector:       game.EffectSelectorCreaturesYouControl,
            },
        },
    },
},
```

---

## Common Pitfalls

1. **Self-referencing names**: When oracle text says the card's own name (e.g., "Lightning Bolt deals 3 damage"), it means "this spell/permanent". Don't create a target for the card itself.

2. **Reminder text**: Text in parentheses like "(Attacking doesn't cause this creature to tap.)" is reminder text — it's not rules text. Ignore it when parsing abilities. The keyword itself carries the rules meaning.

3. **"Any target"**: This means target creature, player, planeswalker, or battle (CR 115.4). Use `Constraint: "any target"` or `Allow: game.TargetAllowPermanent | game.TargetAllowPlayer`.

4. **"You" as target vs. controller**: When the text says "you gain 1 life", "you" is the controller, not a target. Use `TargetIndex: -1`. Only use `TargetIndex: 0+` when there's an explicit "target" word.

5. **Tap symbol in costs**: `{T}` in an activated ability cost means the permanent taps itself. This goes in `AdditionalCosts` as `{Kind: game.CostTap}`, not in `ManaCost`.

6. **Multiple paragraphs = multiple abilities**: Each `\n`-separated paragraph is a separate ability and gets its own `AbilityDef`, unless it's a comma-separated keyword list.

7. **Variable amounts**: When an effect says "equal to its power" or "equal to the number of...", the current `Effect.Amount` field can't express this. Use `Description` to document it, and consider setting `ImplementationID` if the card needs full rules accuracy.

8. **Cards that need ImplementationID**: If a card does things the declarative system can't express (unsupported choice consumers, ETB-as-copy/as-choice, full replacement ordering, or complex conditional logic), set `ImplementationID` to a descriptive kebab-case string like `"swords-to-plowshares"` and leave a comment explaining what the hand-written code needs to do.
