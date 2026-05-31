# Keep Combat as an in-place rules module

Combat behavior remains in `mtg/rules` behind the package-local `combatEngine`
module rather than moving to a `mtg/rules/combat` subpackage.

## Context

Combat is one of the largest behavior areas in the rules engine. It owns combat
phase orchestration, attacker and blocker declaration, goad legality,
evasion restrictions, attack taxes, combat damage assignment, lifelink,
deathtouch, trample, first strike, and the event/log mutations that follow.

We considered extracting this behavior to a smaller `mtg/rules/combat`
subpackage after extracting the Payment Planner package. Unlike payment,
combat does not yet have a narrow state seam. Combat still needs close access to
priority loops, player observations, action comparison, action construction,
payment orchestration, mutation helpers, continuous/effective values,
replacement/prevention-aware damage helpers, last-known information, event
emission, and combat logs.

## Decision

Keep Combat as an in-place rules module for now.

The durable module is `combatEngine` in `mtg/rules/combat_engine.go`.
`Engine.runCombatPhase` delegates to `combatEngine.runPhase`, while
`combatEngine` owns declaration enumeration/application, combat damage passes,
priority-window orchestration for combat, and attack-tax integration.

Do not create `mtg/rules/combat` until the adapter surface is meaningfully
smaller than the implementation it hides.

## Consequences

The current design improves locality without creating a shallow package seam.
Combat code is concentrated around `combatEngine`, while shared data types such
as `game.AttackDeclaration`, `game.BlockDeclaration`, and `game.CombatState`
remain in `mtg/game`.

A future extraction should revisit this decision only when Combat can satisfy
the deletion test: deleting the package would scatter real combat complexity
back into callers, rather than merely removing a pass-through adapter.

The README in `mtg/rules` records promotion criteria for a future combat
subpackage.
