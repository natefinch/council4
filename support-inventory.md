# Production support inventory

Production registry: **2790 entries, 2790 loadable, 11 behavior-verified authored identities**. Frozen compiler: **18700/33016 eligible Oracle records accepted**. These are separate measurements; name-only legacy entries are not certified exact Oracle/face identities or behavior-verified games.

Pinned corpus SHA256 `6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4`; compiler evidence SHA256 `f8f1f1447e8b56484b40ab91f86bd5d3e427b71991a03b3926e64972c1b78601`; Comprehensive Rules 2026-09-25. Reproduce with `mage supportInventory` (no corpus compilation).

## Exact authored Oracle/front admissions

Source revisions bind authored Go (all handlers and dependency declarations), not just CardDefs. Shared-rules revisions bind non-test Go in `mtg/{rules,game,cardprogram}` and `opt`. `go generate ./mtg/cards/authored` updates them; a native test rejects stale bindings. Digests detect changes; they do not prove behavior.

| Card | Oracle identity | Compiler accepted | Loadable | Behavior evidence | Program versions |
| --- | --- | --- | --- | --- | --- |
| Abundance | `eccc9a54-2b56-4a02-9926-258d5b2e25fb/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{eccc9a54-2b56-4a02-9926-258d5b2e25fb/replacement/0 1}]` |
| Blood Moon | `94fac5fe-97d5-4c12-a80c-8efff9d853ae/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[]` |
| Gifts Ungiven | `58aec411-167d-4709-8560-793eaaed62c5/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{58aec411-167d-4709-8560-793eaaed62c5/spell/0 1}]` |
| Gilded Drake | `7f06c098-6482-4bf3-a9a1-110d6d5b5703/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{7f06c098-6482-4bf3-a9a1-110d6d5b5703/enter/0 1}]` |
| Hive Mind | `f97e405d-4c24-4d4d-9e19-e349c073113c/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{f97e405d-4c24-4d4d-9e19-e349c073113c/cast/0 1}]` |
| Mindslaver | `a806f4ee-f48a-46c3-8772-5aaabebe4c7e/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{a806f4ee-f48a-46c3-8772-5aaabebe4c7e/activate/0 1}]` |
| Muldrotha, the Gravetide | `e4625704-1d52-44e4-804f-2f45644d76ac/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[]` |
| Nim Deathmantle | `66d41377-626d-4ae6-ba86-17bf0c8b3362/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{66d41377-626d-4ae6-ba86-17bf0c8b3362/death/0 1}]` |
| Skullbriar, the Walking Grave | `088ac883-6f44-4046-bf19-81c7fd952e55/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[]` |
| Terastodon | `17975a20-2c13-4325-be1e-0bcda5063a2d/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{17975a20-2c13-4325-be1e-0bcda5063a2d/enter/0 1}]` |
| Worldgorger Dragon | `a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c/enter/0 1} {a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c/leave/0 1}]` |

- Abundance: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Blood Moon: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Gifts Ungiven: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Gilded Drake: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Hive Mind: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Mindslaver: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Muldrotha, the Gravetide: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Nim Deathmantle: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Skullbriar, the Walking Grave: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Terastodon: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

- Worldgorger Dragon: authored `sha256:92fff125ef534095ef187ebecfdfb5bbd6fefd3952f6244bfd50c2cb8bf8bc4b`; shared rules `sha256:86064a3c9b9336e183c5d90f9d698a934025514aa4d0b20c3a002a239ba0f091`.

## Fidelity and consumers

`cards.Registry.Inventory/SupportFor` is authoritative for installed entries; deck loading returns the same evidence. JSON inventory output joins corpus IDs without guessing legacy identity. Engine, Simulator, CLI and WASM share the installed catalog. A deliberately absent handler fails checked loading or produces `unsupported-execution`, never a rules draw.

Rules wins/draws, turn budgets, work budgets, unsupported and invalid execution are distinct. Action/choice fallbacks and invalid answers are counted separately. `GameResult.Exact()` rejects declared restrictions and policy fallbacks. Limits include X/multikicker 20 ([#3284](https://github.com/natefinch/council4/issues/3284)), representative combat declarations ([#3126](https://github.com/natefinch/council4/issues/3126)) and incomplete evaluation knowledge. Search action width, determinization count, lookahead and deep width are policy budgets, not rules. Unknown authored/effect valuation is explicit; search executes real programs, not metadata.

See [mixed production validation](docs/perf/mixed-production.md) for executed scenarios and measurements. CI runs code/scenario checks and registry preservation for authored edits, not frozen-corpus regeneration. Legacy compiler diagnostics remain in `unsupported-reasons.md` and `card-backlog.md`; they are not the automatic authoring roadmap.
