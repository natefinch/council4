# compilecards

`compilecards` stream-decodes a Scryfall Oracle Cards bulk-data array, compiles
cards in parallel, and writes deterministic Go definitions for only the cards
whose complete rules text is supported by the executable backend.

The initial strict backend supports vanilla faces and plain, non-parameterized
keyword abilities that have reusable `mtg/game` templates. It never emits TODOs
or partial ability implementations. Unsupported cards, unsupported layouts,
source-generation failures, non-ASCII package names, and filename collisions
are written to the report.

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
| `unsupported spell ability` | The text was correctly classified as an instant or sorcery spell ability, but the executable backend cannot lower spell effects yet. |
| `unsupported activated ability` | The parser recognized a cost-and-colon activated ability, but the executable backend cannot emit it yet. |
| `unsupported loyalty ability` | The parser recognized a planeswalker loyalty ability, but the executable backend cannot emit it yet. |
| `unsupported triggered ability` | The parser recognized a `when`, `whenever`, or `at` trigger, but the executable backend cannot emit triggered abilities yet. |
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
