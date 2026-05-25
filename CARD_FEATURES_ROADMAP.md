# Card Features Roadmap

This roadmap tracks card-text features needed by generated `CardDef` implementations beyond the current mechanical Scryfall import. It complements `docs/research/CARD-TEXT-PARSING.md` by mapping common oracle-text patterns to concrete engine/model work.

## Recently covered

- [x] Type-changing animation effects, such as "becomes a 0/0 Robot creature in addition to its other types", via `EffectApplyContinuous` and runtime `ContinuousEffect` layer entries.
- [x] Moving counters from a target or triggering event permanent, including last-known information for objects that left the battlefield, via `EffectMoveCounters`.
- [x] Optional single targets (`MinTargets: 0`, `MaxTargets: 1`) for "up to one target" templates.
- [x] Generalized target selection for multiple target counts and mixed target slots.
- [x] Structured target predicates for common type, color, controller, tapped, combat-state, keyword, mana value, P/T, and "another" constraints.
- [x] Dynamic effect amounts for X, target characteristics, controller zone/life counts, selector counts, counter counts, and linked "that much" amounts.
- [x] Unsupported effect routing/logging for effect enum values that the generic resolver does not execute yet.
- [x] Modal spell variants for choose N, choose up to N, one-or-both, choose three, all modes, and duplicate-mode templates.
- [x] Resolution-time optional effects and success-aware "if you do" / "if you don't" branches.
- [x] Common trigger-pattern slice for upkeep/end beginning-of-step triggers, spell type include/exclude filters, and permanent type include/exclude filters.

## High-priority parser/model gaps

- [x] Generalized target selection for multiple target counts: "up to two target creatures", "two target players", "any number of target...", and mixed target slots.
- [x] Richer target predicates beyond type/controller: color, tapped/untapped, attacking/blocking, mana value, power/toughness, "another", "nonblack", "with flying", and similar qualities.
- [x] Dynamic effect amounts for "equal to...", "for each...", "that much", X-based effects, and counts determined on resolution.
- [x] Resolver coverage or explicit unsupported routing for existing effect enum names that are not generally executed yet: `EffectCounter`, `EffectDiscard`, `EffectSearch`, `EffectGainControl`, `EffectCopy`, `EffectAttach`, and `EffectReplace`.
- [x] Modal variants beyond choose-one: "choose two", "choose one or both", "choose up to one", "choose three", and all-mode selection. Carry-forward: Entwine as a cost-gated mode expansion.
- [ ] Resolution choices inside effects: "you may...", "if you do", "if you don't", "choose a color/card/type/player", and "you may pay..." during resolution. Partial: optional effects and success-aware linked branches are supported; value-producing choices and resolution-time payments remain.
- [ ] Common trigger patterns beyond the initial event filters: beginning-of-step triggers, state triggers, noncreature/artifact/permanent spell casts, opponent-controlled objects, and richer negated conditions. Partial: upkeep/end beginning-of-step triggers plus spell/permanent type include/exclude filters are supported; state triggers and broader step coverage remain.
- [ ] General replacement effects: arbitrary "if X would happen, instead Y", ETB-as-copy/as-a-choice, replacement ordering choices, and self-replacement effects during spell resolution.
- [ ] Static rule-changing effects: "players can't gain life", "creatures can't attack/block", "spells cost less/more", "you may cast/play from...", and other permission/prohibition effects.
- [ ] High-frequency keyword mechanics from the parsing guide and rules README: Ward, Prowess, Flashback, Madness, Suspend, Storm, Cascade, Convoke, Delve, Proliferate, Goad, Morph/Disguise, and related keyword actions.

## Lower-level infrastructure follow-ups

- [ ] Convert natural-language target constraints into structured predicates instead of relying on string matching.
- [x] Add a dynamic value model for effect amounts, separate from the existing P/T-only `DynamicValue`.
- [ ] Ensure generated card definitions cannot appear fully supported when they use an effect enum that the resolver does not execute.
- [ ] Add fixture cards for each supported card-text pattern so generation regressions are caught before broad Commander staple ingestion.
