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
    AdditionalCost     string              // Deprecated: use AdditionalCosts
    AdditionalCosts    []AdditionalCost    // Typed non-mana costs
    AlternativeCosts   []AlternativeCost   // Optional replacement costs
    KickerCost         *mana.Cost          // Optional Kicker mana cost
    KickerEffects      []Effect            // Additional effects if kicked
    Trigger            *TriggerCondition   // When triggered ability fires (nil for non-triggered)
    Optional           bool                // True for "you may" abilities
    Effects            []Effect            // Effects this ability produces
    Targets            []TargetSpec        // Targeting requirements
    Modes              []Mode              // Modal spell/ability modes
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
)
```

### Effect

```go
type Effect struct {
    Type            EffectType
    Amount          int           // Numeric amount (damage, cards drawn, etc.)
    TargetIndex     int           // Index into runtime targets; -1 = controller
    PowerDelta      int           // For EffectModifyPT
    ToughnessDelta  int           // For EffectModifyPT
    ManaColor       mana.Color    // For EffectAddMana
    UntilEndOfTurn  bool          // Duration flag
    Duration        EffectDuration
    Step            Step          // For step-related effects
    Selector        EffectSelector  // For mass effects
    Token           *CardDef      // For EffectCreateToken
    DelayedTrigger  *DelayedTriggerDef
    EmblemAbilities []AbilityDef
    LinkID          string
    Description     string        // Human-readable description
}
```

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

### TargetSpec

```go
type TargetSpec struct {
    MinTargets int     // 0 for "up to"
    MaxTargets int
    Constraint string  // e.g., "creature", "creature or planeswalker", "player",
                       // "any target", "creature or player"
}
```

### TriggerCondition

```go
type TriggerCondition struct {
    Type          TriggerType     // TriggerWhen, TriggerWhenever, or TriggerAt
    Pattern       TriggerPattern  // Structured event pattern
    Event         string          // Deprecated: use Pattern
    InterveningIf string          // "if" condition (CR 603.4)
    InterveningIfControllerLifeAtLeast int
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
    MatchFromZone bool
    FromZone      ZoneType
    MatchToZone   bool
    ToZone        ZoneType
    DamageRecipient DamageRecipientKind
}
```

### EventKind (for trigger patterns)

```go
const (
    EventSpellCast EventKind = iota
    EventAbilityActivated; EventAbilityTriggered
    EventPermanentETB; EventPermanentLTB
    EventCreatureDied; EventZoneChange
    EventDamageDealt; EventLifeGained; EventLifeLost
    EventDrawCard; EventDiscardCard; EventMillCard
    EventCounterAdded; EventCounterRemoved
    EventAttackDeclared; EventBlockDeclared
    EventCombatDamageDealt
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
  - "enters" / "enters the battlefield" → `EventPermanentETB`
  - "dies" → `EventCreatureDied`
  - "leaves the battlefield" → `EventPermanentLTB`
  - "At the beginning of your upkeep" → `EventBeginningOfStep`
  - "Whenever ... attacks" → `EventAttackDeclared`
  - "Whenever ... deals damage" → `EventDamageDealt`
  - "Whenever ... is cast" → `EventSpellCast`
- Set controller/source filters based on "you", "an opponent", "another creature", "this creature"
- If "you may" appears → `Optional: true`

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
| "destroy target" | `EffectDestroy` | `TargetIndex` from target order |
| "exile target" | `EffectExile` | |
| "return target ... to its owner's hand" | `EffectBounce` | |
| "draw N card(s)" | `EffectDraw` | `Amount: N` |
| "discard N card(s)" | `EffectDiscard` | `Amount: N` |
| "gain(s) N life" | `EffectGainLife` | `Amount: N` |
| "lose(s) N life" | `EffectLoseLife` | `Amount: N` |
| "add {C}{C}" / "add {G}" | `EffectAddMana` | `Amount: N`, `ManaColor` |
| "gets +N/+M" | `EffectModifyPT` | `PowerDelta: N`, `ToughnessDelta: M` |
| "put N +1/+1 counter(s)" | `EffectAddCounter` | `Amount: N` |
| "create a N/M token" | `EffectCreateToken` | Set `Token` field |
| "sacrifice" (as effect) | `EffectSacrifice` | |
| "tap target" | `EffectTap` | |
| "untap target" | `EffectUntap` | |
| "scry N" | `EffectScry` | `Amount: N` |
| "surveil N" | `EffectSurveil` | `Amount: N` |
| "mill N" | `EffectMill` | `Amount: N` |
| "fight" | `EffectFight` | |
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

Note: Swords to Plowshares has a variable life gain amount (equal to the creature's power). Since the current `Effect.Amount` is a static int, we use `Description` to document the dynamic behavior. This card may need an `ImplementationID` for full rules accuracy.

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
                Event:              game.EventPermanentETB,
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

3. **"Any target"**: This means "target creature, player, or planeswalker" (CR 115.4). Use `Constraint: "any target"`.

4. **"You" as target vs. controller**: When the text says "you gain 1 life", "you" is the controller, not a target. Use `TargetIndex: -1`. Only use `TargetIndex: 0+` when there's an explicit "target" word.

5. **Tap symbol in costs**: `{T}` in an activated ability cost means the permanent taps itself. This goes in `AdditionalCosts` as `{Kind: game.CostTap}`, not in `ManaCost`.

6. **Multiple paragraphs = multiple abilities**: Each `\n`-separated paragraph is a separate ability and gets its own `AbilityDef`, unless it's a comma-separated keyword list.

7. **Variable amounts**: When an effect says "equal to its power" or "equal to the number of...", the current `Effect.Amount` field can't express this. Use `Description` to document it, and consider setting `ImplementationID` if the card needs full rules accuracy.

8. **Cards that need ImplementationID**: If a card does things the declarative system can't express (variable amounts based on game state, choices within resolution, complex conditional logic), set `ImplementationID` to a descriptive kebab-case string like `"swords-to-plowshares"` and leave a comment explaining what the hand-written code needs to do.
