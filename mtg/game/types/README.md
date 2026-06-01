# types

`mtg/game/types` defines Magic card supertypes, card types, and subtypes as
named string types.

Use `types.Super` for supertypes such as `types.Legendary`, `types.Card` for
primary card types such as `types.Creature`, and `types.Sub` for subtypes such
as `types.Angel` or `types.Forest`. The values are strings so they can be logged
and rendered directly while still giving card definitions type-safe fields.
