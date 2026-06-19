# Playtest rule coverage and known limitations

This page summarizes what the rules engine supports along the **playtest path**
— the flow exercised when [`mtg/sim`](../mtg/sim/README.md) runs many games over
a set of decks and [`mtg/report`](../mtg/report/README.md) aggregates the
outcomes. It is the runtime counterpart to the card-compile coverage tracked in
[`supported.md`](../supported.md), [`unsupported.md`](../unsupported.md), and
[`unsupported-reasons.md`](../unsupported-reasons.md): those describe which cards
cardgen can *generate*, while this page describes which mechanics the engine can
*resolve* once a generated card is in play.

For the authoritative, frequently-updated capability list, see the **Current
implementation status** section of [`mtg/rules/README.md`](../mtg/rules/README.md);
this page is a stable, higher-level orientation that links into it.

## How the engine resolves effects

Card effects lower to a sequence of typed **primitives** (`game.Primitive`). At
runtime the engine dispatches each primitive through a handler registry
(`mtg/rules/instruction_registry.go`); each `PrimitiveKind` has exactly one
registered handler. This typed-IR lowering design is recorded in
[ADR 0008](./adr/0008-typed-ir-lowering.md), and the surrounding engine
architecture in [ADR 0004](./adr/0004-engine-architecture.md) and
[ADR 0005](./adr/0005-event-centered-ability-architecture.md).

## Supported rule areas

The playtest path supports, among others:

- **Zones and turn structure** — 4-player Commander game state, opening hands,
  draws, and progression through beginning, main, combat, ending, and cleanup
  steps, including extra turns in LIFO order and Saga lore advancement.
- **Priority** — multiplayer pass-around-the-table priority with stack-aware
  all-pass handling. Empty priority passes are skipped per
  [ADR 0002](./adr/0002-smart-priority-skips-empty-passes.md).
- **The stack** — casting supported spells from hand, graveyard, command zone,
  and prepared battlefield permanents; modal spells; copies; and resolution of
  creature, instant/sorcery, Mutate, Cycling, Ninjutsu, Kicker, and activated
  abilities.
- **Combat** — attacker/blocker declarations, evasion (Flying/Reach/Menace),
  first/double strike, Trample, Deathtouch, Lifelink, Toxic, Wither/Infect,
  commander combat damage, and lethal permanent cleanup. Combat is kept in the
  rules package per [ADR 0007](./adr/0007-keep-combat-in-rules.md).
- **Continuous effects** — effective characteristics through the standard
  layer system (copy/control/text/type/color/ability/P-T), with timestamps and
  dependencies, counters, and temporary modifiers.
- **Durations, replacement, and prevention** — runtime duration expiry,
  ETB tapped/counter/payment replacements, Commander command-zone replacement,
  damage prevention, and regeneration.
- **Cost modifiers** — colored/generic/colorless/X payment via pool mana and
  auto-tapped sources, generic cost increases, sacrifice and turn-face-up costs,
  and attack taxes.
- **Triggered and activated abilities** — event-driven trigger detection over a
  typed `game.Event` stream (see
  [ADR 0005](./adr/0005-event-centered-ability-architecture.md)), APNAP stack
  placement, intervening-if conditions, optional triggers, and mana/equip/
  loyalty/general activated abilities.
- **State-based actions** — player elimination (0 life, poison, commander
  damage, failed draw) and permanent SBAs (0 toughness, lethal/deathtouch
  damage, 0 loyalty/defense, completed Sagas, illegal attachments, legendary
  rule, counter cancellation, token cleanup).

The hand-written card escape hatch (`CardDef.ImplementationID`) supplements the
declarative path for effects that do not yet lower cleanly; see
[ADR 0001](./adr/0001-hybrid-declarative-card-implementations.md).

## Known limitations

- **Unsupported mechanics surface as a typed runtime error.** When the engine
  reaches a primitive whose `PrimitiveKind` has no registered handler (a mechanic
  it does not yet resolve), the dispatch path panics with a
  `rules.UnsupportedError` carrying the offending kind and a human-readable
  reason. This is intentional and attributable rather than an opaque string
  panic or a silent wrong result. Most unsupported mechanics are instead caught
  earlier, at card-generation time: cardgen rejects card shapes it cannot lower,
  so those cards never enter a deck or reach the playtest path. `UnsupportedError`
  covers the narrower runtime case of a structurally-recognized primitive that
  reaches the engine with no registered handler.
- **Failures are captured, not fatal.** A single game must not abort a long
  batch. `sim.Run` plays each game under a recover, so any panic — an engine
  bug, an unsupported mechanic, or an illegal applied action — is caught and
  recorded as a `sim.GameFailure{Index, Seed, Reason, Stack, Unsupported}` while
  the rest of the batch completes (see the *Failure capture* section of
  [`mtg/sim/README.md`](../mtg/sim/README.md)). The `Unsupported` boolean is set
  when the recovered value is a `rules.UnsupportedError`, letting a batch
  distinguish "not implemented yet" from genuine defects. Every failure carries
  its seed, so the game is reproducible via `RunOne(cfg, i)` or replay.
- **Card-compile coverage is partial.** Roughly a quarter of paper-eligible
  cards currently generate; the rest are blocked at compile time and never reach
  the playtest path. See [`unsupported-reasons.md`](../unsupported-reasons.md)
  for the capability-aware blocker breakdown.
- **Engine features still in progress.** The rules README's *Not implemented
  yet* list is authoritative; current gaps include full attachment legality,
  agent-driven mulligan/discard/sacrifice/reveal/tutor decisions, several
  alternative-cost keywords (Escape, Foretell, Evoke, cast-without-paying,
  copy-on-stack), DFC back-face characteristics, day/night transitions, and some
  nonstandard Saga and play-vs-cast timing.

## Related documentation

- [`mtg/rules/README.md`](../mtg/rules/README.md) — authoritative implementation
  status.
- [`mtg/sim/README.md`](../mtg/sim/README.md) — determinism, replay, and failure
  capture.
- [`mtg/report/README.md`](../mtg/report/README.md) — how outcomes and failures
  are aggregated and reported.
- [`docs/adr/`](./adr/) — architectural decision records referenced above.
