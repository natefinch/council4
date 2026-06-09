# Typed intermediate representation for executable card lowering

The `cardgen` executable backend lowers Oracle text into a typed intermediate
representation (IR) of `game.*` ability values, assembles and validates a
`game.CardDef` with `game.ValidateCardDef`, and only then renders deterministic
Go source. It no longer builds card source by concatenating Go code strings.

## Context

The original executable backend in `cardgen/executable.go` lowered each
recognized Oracle ability directly into **Go source-code strings** stored on an
`abilityLowering` struct (`staticBodies []string`, `activatedAbility string`,
and so on). Those fragments were injected into the card generator through a
`generatedAbilityFields` plumbing path and stitched together with hand-managed
import detection.

This string-first design had structural problems:

- **No validation before emission.** The generator could emit source for a card
  whose assembled shape would never satisfy `game.ValidateCardDef`. The only
  feedback was a downstream compile or test failure on generated output.
- **Recognition and rendering were entangled.** A single helper both decided
  *what* a piece of Oracle text meant and decided *how* to spell it in Go.
  Changing the rendered form of a `game` type meant editing recognition code,
  and vice versa.
- **Import detection was string-scraping.** Needed packages were inferred by
  inspecting the generated text rather than the values being rendered.
- **Determinism was incidental, not enforced.** Ordering depended on
  string-assembly order rather than a deliberate, testable contract.

The runtime already owns a canonical typed card model (`game.CardDef`,
`game.CardFace`, the categorized ability types) and the authoritative structural
validator `game.ValidateCardDef`. The string backend duplicated knowledge of
that model in `fmt`-style templates instead of using the types directly.

## Decision

Split the executable backend into three explicit stages, each owned by the
correct layer:

1. **Recognition / lowering** (`cardgen/lower.go`). `lowerExecutableFaces`
   compiles Oracle text and dispatches each recognized ability to a `lowerXxx`
   helper that returns a **typed** `game.*` value rather than a string. The
   per-face result is a `loweredFaceAbilities` holding the categorized typed
   values in Oracle order. Unsupported text yields a source-spanned
   `oracle.Diagnostic`, never a guessed value.
2. **Assembly + validation** (`cardgen/executable.go`). `assembleCardDefs`
   combines parsed Scryfall fields with the lowered typed abilities into one or
   more `game.CardDef` values and calls `game.ValidateCardDef`. Any structural
   issue becomes a diagnostic and the card is failed *before* any source is
   emitted.
3. **Deterministic rendering** (`cardgen/render.go`). A zero-value `Renderer`
   walks the validated `[]*game.CardDef` values — the sole source of every
   mechanical and ability value — and emits Go source. It tracks needed imports
   through a `renderCtx`: each method that emits a package's identifiers calls
   `ctx.need(importXxx)` during traversal, so the import set is derived from the
   values being rendered rather than by scraping the generated text. The import
   list is sorted, maps are never iterated for ordering, and identical input
   produces byte-identical output.

The renderer no longer renders printed fields from Scryfall-derived
`generatedCardFields`; the `ScryfallCard` survives only as comment, variable-name,
and layout metadata. Presentation choices that are not derivable from the typed
model — currently just the package-level variable reference a static ability
should render as, e.g. `game.FlyingStaticBody` instead of a struct literal — are
passed in a narrow `faceRenderHints` value. Each hint carries the expected
`game.StaticAbility` body and is verified with `reflect.DeepEqual` against the
CardDef value before the `VarName` is used; a mismatch is a render error
(divergence), never a silently wrong emission.

The string-building helpers, the `generatedAbilityFields` injection path, and
the non-executable skeleton generator are removed. Card Generation has one
source path: complete recognition, typed lowering, validation, and rendering.

This keeps the layer boundaries crisp: **`mtg/game` owns the typed data and what
makes it structurally valid, `mtg/rules` owns behavior, and `cardgen` owns
recognition (Oracle text → typed values) and rendering (typed values → Go
source).** The typed model is used as a compiler IR precisely because it is the
runtime's own validated representation — there is no second, parallel notion of
"a valid card."

Two constraints shape the renderer. First, sealed interfaces are rendered by
switching on the value's `Kind()` and performing a single type assertion per
case, never a Go type switch, matching the dispatch style used elsewhere in the
codebase. Second, all generated Go is run through `go/format` so the committed
shape is gofmt-stable.

## Considered Options

- **Keep string templates, add a post-emit validation pass.** Rejected: it
  validates the *text*, not the model, and leaves recognition and rendering
  entangled. Import detection stays string-based.
- **Lower directly to `game.CardDef` and rely on `go/format` of a `%#v`-style
  dump.** Rejected: `%#v` does not produce package-qualified, constructor-based,
  human-reviewable source, and cannot emit shared vars like
  `game.FlyingStaticBody` or compact forms the tests and reviewers expect.
- **Typed IR + validate-before-render (chosen).** Recognition produces typed
  values, assembly validates them against the runtime's own validator, and a
  dedicated deterministic renderer owns spelling. Each stage is independently
  testable.

## Consequences

- Cards that cannot pass `game.ValidateCardDef` are reported as unsupported and
  never emitted, so generated source is structurally valid by construction.
- Recognition changes (new Oracle patterns) and rendering changes (how a type is
  spelled in Go) are now independent edits in `lower.go` versus `render.go`.
- The renderer's determinism is a tested contract (`render_test.go` asserts
  byte-identical repeated output and gofmt-stability, and that unsupported typed
  values and divergent hints return errors; `roundtrip_test.go` both compiles
  generated source with `go build` and runs a generated semantic test that
  asserts the emitted vars round-trip to the expected typed structure).
- The handwritten **Card Implementation** escape hatch remains available for
  exceptional mechanics, but it is not a fallback source-generation path.
  Card Generation emits only cards it can fully recognize, assemble, and
  validate; everything else is reported as unsupported.
- Existing generated cards in `mtg/cards/` are unaffected; the change is to how
  new executable source is produced, not to already-committed definitions.
