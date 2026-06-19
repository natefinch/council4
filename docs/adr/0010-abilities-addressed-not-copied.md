# Abilities are addressed, not copied, through the Ability interface

The `game.Ability` marker interface is implemented on **pointer receivers**, and
the ability accessors (`CardFace.BodyAt` and the `Body*` helpers) hand back the
address of an ability stored in a `CardFace`, never a copy.

## Context

A card face stores its abilities in typed value slices — `[]ActivatedAbility`,
`[]ManaAbility`, `[]StaticAbility`, and so on. `Ability` is a sealed marker
interface (`isAbility()`), and `CardFace.BodyAt(i)` returns the ability at a
canonical index as an `Ability`.

Originally every ability type satisfied `Ability` with a **value receiver**, so
`BodyAt` returned the ability *value*. Returning a concrete struct as an
interface boxes a heap copy of that struct. `BodyAt` is on the hottest path in
the engine: `basePermanentValues` rebuilds a permanent's effective ability list
by calling `BodyAt` for every ability, and effective values are computed for
every permanent at effectively every priority point. In a profiled long game
this boxing was the single largest source of allocation — billions of
allocations and hundreds of gigabytes of garbage — and it dominated runtime via
garbage-collection pressure. (The static-source frame cache from ADR-adjacent
work, PR #717, removed the *repeated* recomputation; this addresses the
per-ability allocation that remains.)

## Decision

Implement `isAbility()` (and, for consistency, the other `AbilityContent`
methods) on pointer receivers, and have `BodyAt` return `&face.Slice[i]`. The
ability slice element is addressable, so the returned `Ability` wraps the
existing element pointer and allocates nothing.

We deliberately kept the storage as **value slices** (`[]ActivatedAbility`)
rather than switching to pointer slices (`[]*ActivatedAbility`). Pointer slices
would have achieved the same zero-boxing result but would have broken every one
of the ~280 committed generated corpus card files, forcing a full corpus
regeneration. Returning the address of a value-slice element gets the identical
benefit with no change to how cards are written or generated.

## Consequences

- `BodyAt` and effective-ability construction allocate nothing for the ability
  bodies themselves. In the heavy baseline game (`docs/perf`) per-game
  allocations dropped substantially.
- Only `*ActivatedAbility` (etc.) satisfies `Ability`; a bare value no longer
  does. Type switches and assertions on `Ability` use pointer cases, and the
  handful of sites that built an `Ability` from a value now take its address.
- **The returned pointer aliases into the card definition.** This is sound only
  because a `*CardDef` is immutable once a `CardInstance` references it: all
  card modifications go through `copyCardDef` on a fresh copy, and `Game.Clone`
  deliberately shares `*CardDef` for exactly this reason. Callers must treat the
  returned `Ability` as read-only and must not retain it across a mutation of the
  owning definition. A prominent comment at the `isAbility` receivers records
  this contract.

## Alternatives considered

- **Pointer slices (`[]*ActivatedAbility`).** Rejected: breaks the committed
  corpus and the cardgen renderer for no additional benefit over addressing
  value-slice elements.
- **Caching a boxed `[]Ability` per card definition.** Rejected earlier (PR
  #717 follow-up discussion): a per-def cache assumes the def's abilities never
  change after first read, which is the same immutability contract this change
  relies on but with a larger, silent failure mode if violated. Addressing the
  slice element needs no cache and no extra state.
