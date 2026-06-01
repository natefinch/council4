# compare

`mtg/game/compare` contains small reusable comparison predicates shared by game
data structures and rules evaluation.

Use `compare.Int` when a declarative card field needs to express an integer
predicate such as "power 4 or greater" or "mana value 3 or less". The data stays
in `mtg/game`-reachable packages so generated card definitions can describe the
predicate without importing `mtg/rules`.
