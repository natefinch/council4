# Card Implementation Guide

Reference for parsing Magic: The Gathering oracle text into council4 `game.AbilityDef` structs.

## Go Type Definitions

These are the exact types you must use. Do not invent new enum values.

### AbilityDef

```go
type AbilityDef struct {
    Kind               AbilityKind
    Text               string              // Full oracle text of this ability paragraph
    Keywords           []Keyword           // Keyword abilities this provides
    ProtectionFromColors []mana.Color       // For Protection keyword
    ManaCost           *mana.Cost           // Mana component of activated ability cost
    AdditionalCosts    []AdditionalCost    // Typed non-mana costs
    AlternativeCosts   []AlternativeCost   // Optional replacement costs
    KickerCost         *mana.Cost          // Optional Kicker mana cost
    KickerEffects      []Effect            // Additional effects if kicked
    Trigger            *TriggerCondition   // When triggered ability fires (nil for non-triggered)
    Optional           bool                // True for "you may" abilities
    Effects            []Effect            // Effects this ability produces
    Targets            []TargetSpec        // Targeting requirements
    Modes              []Mode              // Modal spell/ability modes
    MinModes           int                 // Modal choice minimum (CR 601.2d, CR 700.2)
    MaxModes           int                 // Modal choice maximum; 0/0 with Modes = choose one
    AllowDuplicateModes bool               // True for "choose the same mode more than once" (CR 700.2d)
    ZoneOfFunction     ZoneType            // Zone where ability functions (default: Battlefield)
    Timing             TimingRestriction   // When activated ability can be used
    IsLoyaltyAbility   bool                // True for planeswalker loyalty abilities
    LoyaltyCost        int                 // Loyalty cost for loyalty abilities
    IsManaAbility      bool                // True for mana abilities (CR 605.1)
}
```

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
    PermanentType      CardType
    Negate             bool
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
    Amount          int           // Numeric amount (damage, cards drawn, etc.)
    DynamicAmount   *DynamicAmount // Amount determined on resolution (CR 107.3, CR 608.2c)
    TargetIndex     int           // Index into runtime targets; -1 = controller
    Optional        bool          // Ask whether to apply this single instruction (CR 608.2c)
    ResultCondition *EffectResultCondition // Gate on prior linked effect result
    Condition       *EffectCondition
    PowerDelta      int           // For EffectModifyPT
    ToughnessDelta  int           // For EffectModifyPT
    CounterKind     counter.Kind  // For EffectAddCounter/EffectRemoveCounter
    CounterSource   CounterSourceSpec  // For EffectMoveCounters
    ManaColor       mana.Color    // For EffectAddMana
    Choice          *ResolutionChoice  // Value chosen during resolution (CR 608.2c, CR 609.3)
    ChoiceLinkID    string        // Consume a prior choice value
    Payment         *ResolutionPayment // Optional "you may pay..." during resolution (CR 608.2c, CR 117.12)
    UntilEndOfTurn  bool          // Duration flag
    Duration        EffectDuration
    Step            Step          // For step-related effects
    Selector        EffectSelector  // For mass effects
    Token           *CardDef      // For EffectCreateToken
    ContinuousEffects []ContinuousEffect  // For EffectApplyContinuous
    DelayedTrigger  *DelayedTriggerDef
    EmblemAbilities []AbilityDef
    Replacement     *ReplacementEffect // For EffectReplace
    LinkID          string
    Description     string        // Human-readable description
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
    Colors         []mana.Color
    CardTypes      []CardType
    PlayerRelation PlayerRelation
    Zone           ZoneType
}

type ResolutionPayment struct {
    Prompt          string
    ManaCost        *mana.Cost
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
    Player     TriggerPlayerFilter      // TriggerPlayerAny/You/Opponent
    MatchPermanentType bool
    PermanentType      CardType
    RequirePermanentTypes []CardType
    ExcludePermanentTypes []CardType
    RequireCardTypes []CardType
    ExcludeCardTypes []CardType
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
)
```

### Mana construction helpers

```go
mana.ColoredMana(mana.White)    // {W}
mana.ColoredMana(mana.Blue)     // {U}
mana.ColoredMana(mana.Black)    // {B}
mana.ColoredMana(mana.Red)      // {R}
mana.ColoredMana(mana.Green)    // {G}
mana.GenericMana(3)             // {3}
mana.ColorlessMana()            // {C}
mana.VariableMana()             // {X}
mana.HybridMana(mana.White, mana.Blue)  // {W/U}
mana.MonoHybridMana(mana.White)         // {2/W}
mana.PhyrexianMana(mana.White)          // {W/P}
mana.SnowMana()                         // {S}
```

---

## Classification Rules

### Step 1: Split oracle text into paragraphs

Each paragraph (separated by `\n`) is one ability. Exception: a comma-separated list of keywords on a single line counts as multiple keyword abilities grouped into one `AbilityDef`.

### Step 2: Classify each paragraph

| Test | Kind |
|------|------|
| Contains `:` outside of `{...}` braces | `ActivatedAbility` |
| Starts with `When`, `Whenever`, or `At` | `TriggeredAbility` |
| On an instant or sorcery and not activated/triggered | `SpellAbility` |
| Otherwise | `StaticAbility` |

### Step 3: Extract fields per ability kind

#### Keywords (any ability kind)

If a paragraph is just a keyword name (or comma-separated keywords), create one `AbilityDef` with:
- `Kind: StaticAbility`
- `Keywords: []game.Keyword{game.Flying, game.Vigilance, ...}`
- `Text:` the full oracle text line

For keywords with parameters:
- **Protection from [color]**: `Keywords: []game.Keyword{game.Protection}`, `ProtectionFromColors: []mana.Color{mana.Red}`
- **Ward {N}**: `Keywords: []game.Keyword{game.Ward}`, `ManaCost: &mana.Cost{mana.GenericMana(N)}`
- **Equip {N}**: `Kind: ActivatedAbility`, `Keywords: []game.Keyword{game.Equip}`, `ManaCost: &mana.Cost{mana.GenericMana(N)}`, `Timing: game.SorceryOnly`
- **Cycling {N}**: `Kind: ActivatedAbility`, `Keywords: []game.Keyword{game.Cycling}`, `ManaCost: &mana.Cost{...}`, `AdditionalCosts` with discard self

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
  - "At the beginning of your upkeep/draw step/beginning of combat/end step" → `EventBeginningOfStep` with explicit `Step`
  - "Whenever ... attacks" → `EventAttackerDeclared`
  - "Whenever ... deals damage" → `EventDamageDealt`
  - "Whenever ... is cast" → `EventSpellCast`
- Set controller/source filters based on "you", "an opponent", "another creature", "this creature"
- If "you may" appears → `Optional: true`
- For "At the beginning of your upkeep/draw step/beginning of combat/end step",
  use `EventBeginningOfStep` with `Step: game.StepUpkeep`, `game.StepDraw`,
  `game.StepBeginningOfCombat`, or `game.StepEnd`.
- For cast triggers such as "Whenever an opponent casts a noncreature spell",
  use `EventSpellCast`, `Controller: game.TriggerControllerOpponent`, and
  `ExcludeCardTypes: []game.CardType{game.TypeCreature}`.
- For state triggers, set `Trigger.Type: game.TriggerState` and fill
  `Trigger.State`; do not set an event pattern.

#### Static abilities

- `Kind: StaticAbility`
- Common patterns:
  - "Creatures you control get +N/+M" → `Effects` with `EffectModifyPT`, `Selector: game.EffectSelectorCreaturesYouControl`
  - "Other creatures you control get +N/+M" → same with `EffectSelectorOtherCreaturesYouControl`

---

## Effect Mapping

| Oracle text pattern | EffectType | Notes |
|---------------------|------------|-------|
| "deals N damage to" | `EffectDamage` | `Amount: N`, set `TargetIndex` |
| "deals X damage to" | `EffectDamage` | `DynamicAmount: &game.DynamicAmount{Kind: game.DynamicAmountX}` |
| "deals damage equal to [target]'s power" | `EffectDamage` | `DynamicAmount: &game.DynamicAmount{Kind: game.DynamicAmountTargetPower, TargetIndex: N}` |
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
| "fight" | `EffectFight` | |
| "if [zone change] would happen, instead..." | `EffectReplace` | Set `Replacement` with match zones and `ReplaceToZone` |
| "enters tapped / with counters" as a runtime effect | `EffectReplace` | Set `Replacement.EntersTapped` / `EntersWithCounters` |
| "counter target spell" | `EffectCounter` | |

### TargetIndex convention

- `TargetIndex: 0` = first target declared
- `TargetIndex: 1` = second target declared
- `TargetIndex: -1` = the ability's controller (for "you draw", "you gain life")

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
    {
        Kind:     game.StaticAbility,
        Text:     "Flying\nVigilance",
        Keywords: []game.Keyword{game.Flying, game.Vigilance},
    },
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
        Trigger: &game.TriggerCondition{
            Type: game.TriggerWhenever,
            Pattern: game.TriggerPattern{
                Event:              game.EventPermanentEnteredBattlefield,
                Source:             game.TriggerSourceAny,
                MatchPermanentType: true,
                PermanentType:      game.TypeCreature,
            },
        },
        Effects: []game.Effect{
            {Type: game.EffectGainLife, Amount: 1, TargetIndex: -1},
        },
    },
},
```

Note: "another creature" means the trigger should not fire for Soul Warden itself entering. The current `TriggerPattern` does not have an explicit "not self" filter — the rules engine's trigger matching handles this by comparing the entering permanent's object ID against the trigger source. If this exclusion is not working correctly at runtime, use `ImplementationID` as a fallback.

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
