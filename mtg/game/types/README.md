# types

`mtg/game/types` defines Magic card supertypes, card types, and subtypes as
named string types for card definitions and rules predicates. The public API for
card definitions is this package: use `types.<Name>` constants from
`mtg/game/types`.

Use `types.Super` for supertypes such as `types.Legendary` and `types.Host`,
`types.Card` for primary card types such as `types.Creature`, and `types.Sub`
for subtypes such as `types.Angel` or `types.Forest`. The values are strings so
they can be logged and rendered directly while still giving card definitions
type-safe fields.

The subtype source lists are organized by card-type family in separate files in
this directory (`artifact.go`, `creature.go`, `land.go`, and so on). `types.go`
keeps the shared type definitions and consolidates the subtype lists into
`KnownSubtypeForType`.

The constants cover the Comprehensive Rules 205.3 subtype lists:

- artifact, enchantment, land, planeswalker, instant/sorcery spell, creature and
  kindred, plane, dungeon, and battle subtypes;
- `types.Kindred` shares the creature subtype list;
- `types.Instant` and `types.Sorcery` share the spell subtype list;
- `types.Plane` subtypes may be multi-word values such as
  `types.BolassMeditationRealm`;
- `types.TimeLord` is the single two-word creature subtype.

When the same printed subtype exists in more than one family, the Go identifier
is disambiguated while preserving the same string value. For example,
`types.ArtifactSpacecraft` and `types.PlanarSpacecraft` both have the value
`"Spacecraft"`.

Use `KnownSubtypeForType` when code needs to confirm that a subtype is legal for
a card type before generating a named constant reference.
