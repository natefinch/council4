# compilecards

`compilecards` stream-decodes a Scryfall Oracle Cards bulk-data array, applies
the repository corpus-eligibility policy, compiles eligible cards in parallel,
and writes deterministic Go definitions for only the cards whose complete rules
text is supported by the executable backend.

The strict backend supports the mechanic families listed in the package
[`README`](../../README.md), including ordered spell effects, supported keyword
templates, mana abilities, fixed quantities and targets, Surveil, Investigate,
Proliferate, Regenerate, and Fight. Near-miss wording such as variable
quantities, unsupported conditions or qualifiers, restricted mana, and divided
effects is rejected. The backend never emits TODOs or partial ability
implementations. Unsupported cards, layouts, source-generation failures, and
non-ASCII package names are written to the report. Distinct Oracle cards that
share a generated path or Go identifier receive stable Scryfall-derived suffixes
without changing their printed names. Playable `token` and `double_faced_token`
records instead always go under `tokens/<letter>` and use their complete
normalized Oracle UUID in the filename and Go identifier. A missing or malformed
token Oracle ID fails closed as an unsupported generated identity.

The eligible corpus contains cards that are legal, restricted, or banned in
Standard, Pioneer, Modern, Legacy, Pauper, Vintage, or Commander, plus playable
paper token definitions. The report explicitly excludes Alchemy, digital-only
identities, memorabilia, cards with no qualifying paper legality, minigames,
art-series records, emblems, planes, schemes, and Vanguard cards.

Writes are serialized after compilation. Existing files at matching generated
paths are overwritten. Each affected letter package's `cards.go` list is then
regenerated from all CardDef declarations in that directory. Token generation
also writes package documentation. When targeting the repository's `mtg/cards`
tree it writes a `tokens.Cards` aggregate; temporary output keeps independently
buildable token letter packages. Tokens remain outside the ordinary card-name
registry.

For a safe full-corpus trial, target a temporary cards root:

```bash
go run ./cardgen/oracle/cmd/compilecards \
  -in cardgen/oracle/oracle-cards-20260608090247.json \
  -out .cardwork/generated-cards \
  -report .cardwork/oracle-compile-report.json
```

During compiler expansion work, prefer
[`corpusdelta`](../corpusdelta/README.md), which runs this command and
automatically prepares the corpus delta, supported-card list, generated-package
validation, and review manifest.

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

The report keeps `card_count` as the total input count and separates it into
`eligible_count` and `excluded_count`. Eligible cards are then partitioned into
`generated_count` and `unsupported_count`. Each excluded record includes a
stable reason. This keeps non-playable objects out of compiler-support metrics
without silently dropping them.

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
| `unsupported surveil spell` | A surveil effect was recognized, but it is not an exact fixed amount performed by the controller. |
| `unsupported investigate spell` | Investigate was recognized, but the instruction is repeated, qualified, or otherwise outside the exact supported form. |
| `unsupported proliferate spell` | Proliferate was recognized, but the instruction is repeated, qualified, or otherwise outside the exact supported form. |
| `unsupported regenerate spell` | Regenerate was recognized, but it does not target exactly one supported permanent. |
| `unsupported fight spell` | Fight was recognized, but its two creature targets or controller restrictions cannot be represented exactly. |
| `unsupported discard spell` | A discard effect was recognized, but it is not an exact fixed number of cards discarded by the controller or one target player. |
| `unsupported tap spell` | A tap effect was recognized, but it is not exact tapping of one artifact, creature, enchantment, land, or permanent target. |
| `unsupported untap spell` | An untap effect was recognized, but it is not exact untapping of one artifact, creature, enchantment, land, or permanent target. |
| `unsupported mill spell` | A mill effect was recognized, but it is not an exact fixed number of cards milled by the controller or one target player. |
| `unsupported enter trigger` | A self-enter trigger was recognized, but its event, condition, optionality, structure, or number of effects is outside the exact supported template. |
| `unsupported enter trigger effect` | The trigger clause is supported, but its effect sequence does not match supported complete spell-like effect templates. |
| `unsupported dies trigger` | A self-dies trigger was recognized, but its event, condition, structure, or number of effects is outside the exact supported template. |
| `unsupported dies trigger effect` | The self-dies trigger clause is supported, but its effect sequence does not match supported complete spell-like effect templates. |
| `unsupported enters-tapped replacement` | Replacement wording was recognized, but it is not an exact unconditional supported self enters-tapped sentence. |
| `unsupported Cycling ability` | Cycling was recognized, but it is not an exact ordinary Cycling ability with a representable mana cost. |
| `unsupported activated ability` | The parser recognized a cost-and-colon activated ability, but it is neither a supported tap mana ability nor an ordinary battlefield activation with exact mana-only, tap-only, or mana-then-tap costs and a supported effect body. |
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
| `unsupported mixed keyword ability` | One or more supported keywords were recognized, but the same ability contains additional rules text outside supported complete mixed families. Exact fixed attached-creature power/toughness buffs may also grant supported simple keywords; other mixed forms reject the whole ability rather than emitting only the keyword portion. |
| `unsupported card layout` | The Scryfall layout cannot be represented safely by the current `CardDef` source generator. |
| `unsupported type line` | The type line contains no card type known to `cardgen`, so emitting a mechanically incomplete `CardFace` would be unsafe. |
| `unsupported package letter` | The card name does not begin with an ASCII letter and therefore cannot be routed to an `mtg/cards/[a-z]` package. |
| `generated path collision` | Multiple corpus records still map to the same generated filename after stable identity disambiguation. |
| `generated identifier collision` | Multiple generated files still declare the same Go `CardDef` variable after stable identity disambiguation. |
| `generated identity collision` | A colliding card has neither an Oracle ID nor Scryfall ID from which to derive a stable suffix. |
| `invalid generated identity` | A playable token is missing the valid Oracle UUID required for its stable token namespace identity. |
| `source generation failed` | Semantic support checks passed, but mechanical source formatting or generation failed, commonly because a mana symbol or another mechanical field is unsupported. |

Lexer or parser errors such as `unclosed quote`, `unclosed parenthesis`, and
`modal ability has no options` mean the Oracle text itself is malformed or its
syntax is not yet recognized. These errors are passed through unchanged and
prevent source generation.
