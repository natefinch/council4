# compilecards

`compilecards` stream-decodes a Scryfall Oracle Cards bulk-data array, compiles
cards in parallel, and writes deterministic Go definitions for only the cards
whose complete rules text is supported by the executable backend.

The strict backend supports vanilla faces, plain non-parameterized keywords,
mana-cost Ward and Cycling, supported tap mana choices, unconditional
enters-tapped replacements, fixed single-target damage, destruction, exile,
return-to-hand, and power/toughness changes, narrow mass destruction, fixed draw
and life changes, fixed controller scry, fixed controller or target-player
discard and mill, and one-target tap and untap. It also supports exact
self-enter and self-dies triggers containing one supported effect. Near-miss
wording such as variable quantities, compound or conditional effects, qualified
targets, optional triggers, restricted mana, and divided or unsupported mass
effects is rejected. The backend never emits TODOs or partial ability
implementations. Unsupported cards, unsupported layouts, source-generation
failures, non-ASCII package names, and filename collisions are written to the
report.

Writes are serialized after compilation. Existing files at matching generated
paths are overwritten. Each affected letter package's `cards.go` registry is
then regenerated from all CardDef declarations in that directory.

For a safe full-corpus trial, target a temporary cards root:

```bash
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/generated-cards \
  -report .cardwork/oracle-compile-report.json
```

To overwrite matching repository card files:

```bash
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out mtg/cards \
  -report .cardwork/oracle-compile-report.json
```

Flags:

- `-in`: required Scryfall Oracle Cards JSON array.
- `-out`: cards package root. Default `mtg/cards`.
- `-report`: unsupported report path, or `-` for standard output. Default `-`.
- `-format`: `json` or `text`. Default `json`.
- `-workers`: compiler/source-generator worker count. Default
  `runtime.NumCPU()`.

## Report diagnostics

Each unsupported card has one or more source-spanned diagnostics. A card can
have several diagnostics when it has several unsupported abilities or when the
semantic compiler and executable backend both identify limitations.

| Summary | Meaning |
| --- | --- |
| `unsupported Oracle construct` | The semantic compiler could not identify a supported action or keyword in the indicated text. This is an Oracle-language recognition gap, before executable source generation. |
| `unsupported cost` | An activated or loyalty cost was preserved as text but was not assigned complete typed cost semantics. |
| `unsupported spell ability` | The text was correctly classified as an instant or sorcery spell ability, but it does not match a supported complete spell template. |
| `unsupported multiple spell abilities` | A face has more than one separately parsed spell ability. The current backend emits only one `SpellAbility` value per face. |
| `unsupported damage spell` | A damage effect was recognized, but its source, amount, recipient, targeting, or surrounding wording is outside the exact fixed single-target templates. |
| `unsupported draw spell` | A draw effect was recognized, but its amount, recipient, targeting, or surrounding wording is outside the exact fixed draw templates. |
| `unsupported destroy spell` | A destroy effect was recognized, but it is neither exact unconditional destruction of one supported target permanent nor an exact supported destroy-all form. |
| `unsupported exile spell` | An exile effect was recognized, but it is not exact exile of one supported target permanent. |
| `unsupported return spell` | A return effect was recognized, but it is not exact return of one supported target permanent to its owner's hand. |
| `unsupported power/toughness spell` | A power/toughness change was recognized, but it is not an exact fixed signed change to one target creature until end of turn. |
| `unsupported life spell` | A gain-life or lose-life effect was recognized, but its amount, affected player, or surrounding wording is outside the exact fixed templates. |
| `unsupported scry spell` | A scry effect was recognized, but it is not an exact fixed amount performed by the controller. |
| `unsupported discard spell` | A discard effect was recognized, but it is not an exact fixed number of cards discarded by the controller or one target player. |
| `unsupported tap spell` | A tap effect was recognized, but it is not exact tapping of one artifact, creature, enchantment, land, or permanent target. |
| `unsupported untap spell` | An untap effect was recognized, but it is not exact untapping of one artifact, creature, enchantment, land, or permanent target. |
| `unsupported mill spell` | A mill effect was recognized, but it is not an exact fixed number of cards milled by the controller or one target player. |
| `unsupported enter trigger` | A self-enter trigger was recognized, but its event, condition, optionality, structure, or number of effects is outside the exact supported template. |
| `unsupported enter trigger effect` | The trigger clause is supported, but its single effect does not match a supported complete spell-like effect template. |
| `unsupported dies trigger` | A self-dies trigger was recognized, but its event, condition, structure, or number of effects is outside the exact supported template. |
| `unsupported dies trigger effect` | The self-dies trigger clause is supported, but its single effect does not match a supported complete spell-like effect template. |
| `unsupported enters-tapped replacement` | Replacement wording was recognized, but it is not an exact unconditional supported self enters-tapped sentence. |
| `unsupported Cycling ability` | Cycling was recognized, but it is not an exact ordinary Cycling ability with a representable mana cost. |
| `unsupported activated ability` | The parser recognized a cost-and-colon activated ability, but it is not an exact supported single-color tap mana ability. |
| `unsupported mana symbol` | A mana effect otherwise matched the supported template, but its output symbol is not one of `{W}`, `{U}`, `{B}`, `{R}`, `{G}`, or `{C}`. |
| `incomplete executable lowering` | A lowering path did not account for every semantic element or meaningful source token. This internal safety check rejects the whole card rather than emitting a partial implementation. |
| `unsupported loyalty ability` | The parser recognized a planeswalker loyalty ability, but the executable backend cannot emit it yet. |
| `unsupported replacement ability` | The parser recognized replacement wording, but the executable backend cannot emit replacement abilities yet. |
| `unsupported static ability` | The text is a non-keyword static ability. The current backend supports only static abilities composed entirely of supported keyword templates. |
| `unsupported reminder ability` | The entire Oracle text is reminder text. This is valid syntax, but it does not map to an executable ability emitted by the backend. |
| `unsupported modal ability` | The parser recognized a bullet or inline `Choose` ability, but the executable backend cannot emit its modes yet. |
| `unsupported ability word` | The ability has a prefix such as landfall or threshold. The semantic text is preserved, but the backend does not yet lower ability-word abilities. |
| `unsupported keyword ability` | The keyword was recognized, but `mtg/game` has no reusable static-ability template for it. |
| `unsupported parameterized keyword` | The keyword requires a value, cost, selector, or other parameter, such as `Toxic 1`; the strict backend currently emits only non-parameterized keyword templates. |
| `unsupported mixed keyword ability` | One or more supported keywords were recognized, but the same ability also contains additional rules text. The backend rejects the whole ability rather than emitting only the keyword portion. |
| `unsupported card layout` | The Scryfall layout cannot be represented safely by the current `CardDef` source generator. |
| `unsupported type line` | The type line contains no card type known to `cardgen`, so emitting a mechanically incomplete `CardFace` would be unsafe. |
| `unsupported package letter` | The card name does not begin with an ASCII letter and therefore cannot be routed to an `mtg/cards/[a-z]` package. |
| `generated path collision` | Multiple corpus records map to the same generated filename. All colliding records are rejected to avoid order-dependent overwrites. |
| `generated identifier collision` | Multiple generated files would declare the same Go `CardDef` variable in one package. All colliding records are rejected. |
| `source generation failed` | Semantic support checks passed, but mechanical source formatting or generation failed, commonly because a mana symbol or another mechanical field is unsupported. |

Lexer or parser errors such as `unclosed quote`, `unclosed parenthesis`, and
`modal ability has no options` mean the Oracle text itself is malformed or its
syntax is not yet recognized. These errors are passed through unchanged and
prevent source generation.
