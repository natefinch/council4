# A hard text-interpretation boundary at the executable card model

Oracle card text is interpreted in exactly one place — the parser — and the
result flows through the translation pipeline into a single well-known,
serializable, behavioral data structure, the executable **card model**
(`game.CardDef`). Nothing downstream of that model is allowed to know that Oracle
card text exists. The card model is the artifact a consumer works off of; in this
repository the consumer is the playtester (`mtg/rules`), but the same model could
be serialized to JSON and consumed by any other tool (deck analyzer, alternate
rules engine, search). We chose this because a clean text→data boundary lets the
text-aware translator and the card-consuming runtime evolve independently, makes
the card data portable, and turns "no card understands its own text" into a
property we can mechanically enforce rather than merely intend.

## Context

The Oracle pipeline is a chain of progressively more abstract representations:

    shared <- lexer <- parser <- compiler <- cardgen lowering -> game.CardDef -> runtime

- **lexer**: Oracle text into tokens.
- **parser**: tokens into `parser.Document` — *the text, structured*. Its shape
  mirrors Oracle grammar (abilities, sentences, parsed effects) and it retains
  tokens and source spans. The parser owns all Oracle vocabulary, spelling, and
  grammar.
- **compiler**: `parser.Document` into compiled semantics (`CompiledAbility`) — an
  intermediate representation used only during translation.
- **cardgen lowering**: compiled semantics into `game.CardDef` — *the card,
  modeled*. Typed, declarative game structures describing what the card is and
  does, emitted as generated Go and consumed by the runtime.

Epic #410 made the **compiler** text-blind: it derives no meaning from the
spelling of Oracle tokens, enforced by an AST analyzer
(`TestCompilerIsTextBlind`). Follow-ups removed the compiler's raw-token
ingestion (#429) and span-offset reasoning (#428, `TestCompilerIsPositionBlind`),
and made `parser.Document` JSON-serializable (#431). The compiler is now a clean
typed-data consumer.

Two questions then arose:

1. *Where is the hard line between "card data" and "card consumer"?* The line a
   third party would build against — a stable, text-free, serializable structure
   that fully describes a card's behavior.
2. *Is the parser that line?*

The parser is **not** the line. `parser.Document` is *the text, structured*: it is
shaped like the Oracle sentence rather than like the card, it is expressed in
grammar vocabulary rather than game vocabulary, and it still contains tokens and
spans by construction. It is an excellent debugging artifact (the `mage parse`
command dumps it as JSON) but it is the wrong abstraction level for a consumer
that wants behavior, not grammar.

The right line is the **executable card model**, `game.CardDef`: typed, declarative
game structures (`CardFace` with its activated/triggered/static/spell abilities,
costs, types, power/toughness). These are pure data — no function values or
interfaces — so they can be serialized. The runtime already consumes only
`CardDef`. The translator (parser → compiler → lowering) is a build-time tool that
produces this artifact.

The boundary is not yet *hard*, for two reasons:

- **Text still leaks to the boundary.** The compiler is text-blind, but the final
  stage that produces `CardDef` — cardgen lowering — still interprets Oracle
  wording. It reads raw parser tokens to find clause boundaries, re-slices and
  re-interprets ability body text, classifies trigger events by substring
  (`strings.Contains(event, ...)`), and walks tokens and spans for a fail-closed
  source-coverage check. Until lowering is text-blind, `CardDef` is the
  *aspirational* boundary, not the enforced one.
- **The artifact is emitted as generated Go, not data.** `CardDef` is declarative,
  but it ships as compiled Go literals, not as a portable serialized registry.

## Decision

Draw the hard text-interpretation boundary at the executable card model and make
it enforceable:

1. **The parser is the sole interpreter of Oracle text.** All vocabulary,
   spelling, normalization, and grammar live there. This is already true and
   already enforced for the compiler.

2. **`game.CardDef` is the card-model boundary.** It is the well-known,
   declarative, serializable behavioral structure that consumers depend on.
   Everything upstream of it (lexer, parser, compiler, lowering) is the
   text-aware *translator*; everything downstream (the playtester and any other
   consumer) depends only on the model and never on Oracle text, tokens, spans, or
   parser internals.

3. **Make cardgen lowering text-blind**, so the boundary is real:
   - Move token-level structural work (clause/cost/body boundary finding) into
     typed data the parser emits, mirroring the compiler's #429 migration.
   - Replace the source-coverage safety gate (which walks tokens and spans to
     reject cards with un-accounted-for source text) with a parser-emitted
     "fully consumed" assertion, so the fail-closed guarantee is preserved rather
     than dropped.
   - Type the residual behavioral text interpretation (body-text re-parsing,
     substring event classifiers), leaving only an explicitly justified
     allowlist of diagnostic-only reads, as the compiler's #418 gate did.

4. **Enforce it.** Add an AST analyzer gate proving cardgen lowering performs no
   Oracle-wording interpretation outside the allowlist, mirroring
   `TestCompilerIsTextBlind` / `TestCompilerIsPositionBlind`. This keeps the
   property true permanently rather than only at the moment of the migration.

5. **Make the model portable.** Confirm `CardDef` is JSON-round-trippable and emit
   a JSON card registry as a build artifact alongside the generated Go, so any
   consumer can work off the data without the translator.

Descriptive, human-readable text may remain on the model as labels (for example a
rendered `Text` field), provided nothing reads it to drive behavior. "Text-blind"
means no text *determines behavior*, not that no strings exist.

## Considered Options

- **Draw the line at `parser.Document` (chosen against).** Tempting because the
  parser already owns text and #431 made the document JSON-serializable. Rejected
  because the document is grammar-shaped and text-coupled: it describes how a
  sentence reads, not what a card does, and it still carries tokens and spans. A
  consumer building a playtester or analyzer wants game behavior in game
  vocabulary, which is the executable model, not the parsed sentence tree. The
  serializable document remains useful, but as a debugging tool, not the consumer
  boundary.

- **Draw the line at the compiler's `CompiledAbility` (chosen against).** It is
  semantic and text-blind, but it is an internal intermediate representation tied
  to the translation process — it carries spans for the coverage gate and is not
  designed as a stable public model. Exposing it would couple consumers to
  compiler internals.

- **Leave lowering as-is and call `CardDef` the boundary anyway (chosen against).**
  The model would still be reachable only through a stage that interprets text, so
  the "nothing downstream knows about text" guarantee would be informal and would
  silently erode. Without making lowering text-blind and enforcing it, the
  boundary is a comment, not a contract.

- **Draw the line at the executable card model and enforce it (chosen).** Gives a
  stable, portable, behavioral data model with a mechanically enforced text-free
  contract, cleanly separating the text-aware translator from the card-consuming
  runtime.

## Consequences

- The card model and the text→model translator become independently evolvable. A
  consumer (the playtester, or a third party from JSON) depends only on `CardDef`.
- Oracle-text interpretation lives in exactly one auditable place, removing the
  current split between the parser and ~60 residual interpretation sites in
  lowering and eliminating a class of fragile substring-matching and
  double-parsing bugs.
- The fail-closed coverage guarantee is preserved but relocated upstream into a
  parser-emitted assertion; net complexity moves rather than disappears.
- A few diagnostic-only text reads (unsupported-reason messages that run after a
  card has already failed) are expected to remain behind a justified allowlist,
  exactly as the compiler's text-blindness gate allows.
- This is a multi-issue effort; it is tracked as an epic with children for the
  lowering migration, the coverage-assertion replacement, the enforcement gate,
  and the JSON registry.
