# Card Features Roadmap

This roadmap tracks card-text features needed by generated `CardDef` implementations beyond the current mechanical Scryfall import. It complements `docs/research/CARD-TEXT-PARSING.md` by mapping common oracle-text patterns to concrete engine/model work.

## Recently covered

- [x] Type-changing animation effects, such as "becomes a 0/0 Robot creature in addition to its other types", via the typed `ApplyContinuous` primitive and runtime `ContinuousEffect` layer entries.
- [x] Moving counters from a target or triggering event permanent, including last-known information for objects that left the battlefield, via the typed `MoveCounters` primitive.
- [x] Optional single targets (`MinTargets: 0`, `MaxTargets: 1`) for "up to one target" templates.
- [x] Generalized target selection for multiple target counts and mixed target slots.
- [x] Structured target predicates for common type, color, controller, tapped, combat-state, keyword, mana value, P/T, and "another" constraints.
- [x] Dynamic instruction amounts for X, target characteristics, controller zone/life counts, selector counts, counter counts, and linked "that much" amounts.
- [x] Sealed typed primitives with exhaustive validation and handler registration.
- [x] Modal spell variants for choose N, choose up to N, one-or-both, choose three, all modes, and duplicate-mode templates.
- [x] Resolution-time optional effects and success-aware "if you do" / "if you don't" branches.
- [x] Common trigger-pattern slice for upkeep/end beginning-of-step triggers, spell type include/exclude filters, and permanent type include/exclude filters.
- [x] Resolution choices and payments for "choose a color/card type/player/card" and "you may pay..." during resolution, with linked choice/payment results.
- [x] Broader beginning-of-step and state trigger support, including explicit draw/beginning-of-combat step events and CR 603.8 state-trigger latching.
- [x] Generic runtime replacement effects for zone-destination replacement and simple ETB tapped/counter modifiers, with deterministic fallback ordering for matching generic replacements.
- [x] Static rule-changing effects for life-gain prohibitions, attack/block prohibitions, generic spell cost modifiers, and graveyard cast permissions via `RuleEffect`.
- [x] Prowess implicit triggers for noncreature spells.
- [x] Flashback-style graveyard casting and exile-on-leave-stack handling.
- [x] Proliferate counter-kind choices for permanents and players.
- [x] Goad effects with expiry on the goading player's next turn.

## High-priority parser/model gaps

- [x] Generalized target selection for multiple target counts: "up to two target creatures", "two target players", "any number of target...", and mixed target slots.
- [x] Richer target predicates beyond type/controller: color, tapped/untapped, attacking/blocking, mana value, power/toughness, "another", "nonblack", "with flying", and similar qualities.
- [x] Dynamic instruction amounts for "equal to...", "for each...", "that much", X-based effects, and counts determined on resolution.
- [x] Typed primitives for supported resolution actions, with unsupported card behavior routed explicitly through `ImplementationID`.
- [x] Modal variants beyond choose-one: "choose two", "choose one or both", "choose up to one", "choose three", and all-mode selection. Carry-forward: Entwine as a cost-gated mode expansion.
- [x] Resolution choices inside instructions: "you may...", "if you do", "if you don't", "choose a color/card/type/player", and "you may pay..." during resolution. Carry-forward: add more consumers for chosen values as new card templates need them.
- [x] Common trigger patterns beyond the initial event filters: beginning-of-step triggers, state triggers, noncreature/artifact/permanent spell casts, opponent-controlled objects, and richer negated conditions. Carry-forward: richer negated trigger conditions beyond existing include/exclude type filters.
- [x] General replacement effects: zone-destination replacement and ETB tapped/counter modifiers via runtime `ReplacementEffect`. Carry-forward: ETB-as-copy/as-choice, full APNAP/self-replacement ordering choices, and richer arbitrary event rewrites.
- [x] Static rule-changing effects: "players can't gain life", "creatures can't attack/block", "spells cost less/more", "you may cast/play from...", and other permission/prohibition effects. Carry-forward: richer conditional/tax/permission variants.
- [ ] High-frequency keyword mechanics from the parsing guide and rules README:
  - [x] Deathtouch
  - [x] Defender
  - [x] Double strike
  - [x] First strike
  - [x] Flash
  - [x] Flying
  - [x] Haste
  - [x] Indestructible
  - [x] Lifelink
  - [x] Menace
  - [x] Protection
  - [x] Reach
  - [x] Trample
  - [x] Vigilance
  - [x] Equip
  - [x] Cycling
  - [x] Kicker
  - [x] Prowess
  - [x] Flashback
  - [x] Proliferate
  - [x] Goad
  - [x] Enchant
  - [x] Hexproof
  - [x] Ward
  - [x] Madness
  - [x] Suspend
  - [x] Storm
  - [x] Cascade
  - [x] Convoke
  - [x] Delve
  - [x] Morph/Disguise: face-down 2/2 cast for `{3}`, turn face up for morph/disguise cost, including Disguise Ward `{2}`/shield counters.
  - [x] Related keyword actions not listed above: counter, discard, supported library search, reveal, and investigate now have first-class effect support.

## Lower-level infrastructure follow-ups

- [ ] Convert natural-language target constraints into structured predicates instead of relying on string matching.
- [x] Add a dynamic value model for effect amounts, separate from the existing P/T-only `DynamicValue`.
- [ ] Ensure generated card definitions cannot appear fully supported when they use an effect enum that the resolver does not execute. Carry-forward: search variants outside explicit library-to-hand `SearchSpec` remain unsupported and logged until modeled; "can't be countered" is not modeled yet.
- [ ] Add fixture cards for each supported card-text pattern so generation regressions are caught before broad Commander staple ingestion.
