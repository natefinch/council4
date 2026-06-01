# Card functionality implementation plan

This plan turns the missing functionality reported by
`.cardwork/deck/unsupported.md` into reusable rules and parser primitives. The
goal is to remove card-specific `ImplementationID` escape hatches only after the
underlying behavior is represented by tested, reusable data and rules support.

## Principles

1. Land each primitive with at least one consuming card conversion.
2. Use the unsupported-card report as the acceptance check: the relevant card or
   functionality gap should disappear when a primitive is complete.
3. Prefer reusable model shapes over card-specific fields, but keep evaluation
   contexts explicit so conditions mean the same thing at each rules site.
4. Do not auto-derive land subtype mana abilities in this slice; existing cards
   use explicit generated mana abilities, and intrinsic subtype mana is a larger
   behavior migration.

## Implementation sequence

### 1. Add report-driven checks

Create a baseline fixture or targeted tests from the current unsupported report.
Each later step should update the expected report so regressions are visible.

### 2. Add reusable conditions with explicit contexts

Add a shared condition model plus rules-side evaluators for:

- static abilities, with source and controller bindings;
- activation restrictions, with activating player and source bindings;
- trigger intervening-if checks, with event bindings;
- effect conditions, with target and linked-result bindings;
- replacement effects, with the in-flight replacement event binding.

Initial condition predicates should cover “controller controls matching
permanent” with a reusable permanent filter: card types, supertypes, subtypes,
minimum count, and power/toughness thresholds.

First consumers: Anger, Blazemire Verge, Bugenhagen, Cinder Glade.

### 3. Add reusable object/player/card references

Introduce reference specs for common bindings while keeping `TargetIndex` as a
legacy shortcut:

- controller;
- target permanent/player;
- target permanent’s controller;
- target permanent’s owner;
- source permanent;
- attached/equipped object;
- linked/revealed card.

First consumers: Beast Within token recipient and Bite Down creature-sourced
damage.

### 4. Expand selectors and dynamic amounts

Add reusable selectors and amount sources:

- attached/equipped object selector;
- all creatures except a target/source;
- player selectors such as opponents;
- opponent count dynamic amount;
- event-derived dynamic amounts such as damage from the triggering event.

First consumers: Basilisk Collar, Blazing Sunsteel, Chandra’s Ignition.

### 5. Implement linked reveal and library-zone effects

Split Chaos Warp into reusable primitives:

- shuffle target permanent into its owner’s library;
- reveal the top card of a player’s library and store it under `LinkID`;
- condition on linked/revealed card characteristics;
- put a linked card onto the battlefield.

These primitives should be usable by later reveal-and-act, cascade-like, and
library-manipulation cards.

### 6. Split ETB replacement support

Implement conditional ETB tapped first, reusing the condition model. Then add
pay-life replacement/payment support for “you may pay N life; if you don’t, this
enters tapped.”

First consumers: Cinder Glade and Tanglespan Bridgeworks.

### 7. Add commander color identity choices

Extend resolution choices so color options can be derived at choice time from
the controller’s commander color identity.

First consumer: Command Tower.

### 8. Add search supertype filtering

Extend `SearchSpec` with supertype filtering so “basic land” can be represented
as a reusable library-search constraint.

First consumer: Bushwhack.

### 9. Add announcement-time target chooser support

Extend target selection so a target slot can be chosen by another player during
spell/ability announcement, not during resolution.

First consumer: Arena.

### 10. Convert cards incrementally

For each completed primitive:

1. Update the generated card definition to use the new reusable model.
2. Remove `ImplementationID` only when all unsupported behavior for that card is
   represented.
3. Regenerate card lists.
4. Run `cardbatch validate` and `cardbatch report`.
5. Confirm the unsupported report shrinks for the targeted card/capability.

## Expected outcome

The rollout report should evolve from “card has generated source but needs
hand-written support” toward a smaller list of genuinely unsupported mechanics.
Each removed report item should correspond to a general capability that future
card implementations can reuse.
