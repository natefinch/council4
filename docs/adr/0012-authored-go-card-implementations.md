# Authored Go card implementations and a direct-to-main workflow

## Decision

New Card Implementations are authored Go using shared Engine operations.
AI may assist offline authoring, but correct behavior is established by rules
research and executable scenarios, not generated code or an AI approval.
Oracle recognition is no longer a prerequisite for supporting a card.

Use data for printed metadata, costs, targets, modes, triggers, static effects
and replacement applicability. Use ordinary Go variables, branches and iteration
for resolving behavior when clearer than declarative Instructions. Keep useful
existing primitives and definitions; do not create another universal DSL or
force every composition through parser, lowering and renderer changes.

The Engine remains the single rules implementation. Card programs use a
controlled execution context, never raw mutable Game state. Simultaneous
operations, actual outcomes, replacement ordering, payments, targeting, priority
and events stay engine-owned. Immutable programs may be shared; mutable links,
references, usage and delayed state must remain game-owned and clone-safe.
Choices use the existing policy callbacks and search branches at stable Simulator
seams, not by copying an ordinary Go call stack.

The repository is a one-maintainer project. Authorized issue implementations
normally produce small direct commits to main, with one writer, rather than
mandatory branches, PRs, approvals and merges. Run scoped checks during
development, the existing suite/linter once before landing code, and inspect the
diff. Docs-only changes use documentation checks. Push permission is separate;
never overwrite unrelated work, force-push or bypass protections.
Independent review is focused on substantial shared-rule changes, not required
for every routine issue. Deep architectural review is explicitly opt-in.

## Why

The long tail needs correlated actors/results, collection behavior, scoped
permissions and rule exceptions. Teaching a general English translator every
composition adds work even when the behavior can already be expressed in Go.
Repeated corpus regeneration, generated-source assertions, nested coordination
and per-issue review ceremony also distract from actual rules correctness.

This changes the authoring and execution defaults, not the requirement to
implement all a card's abilities. Partial cards remain unsupported. Named-card
interaction scenarios and shared-operation tests are both useful; a source or
CardDef fingerprint cannot prove runtime semantics.

## Preserved and superseded decisions

- **ADR 0001:** retain hybrid behavior and engine-owned mutation. Supersede
  compiler-first/declarative-by-default authoring and the claim that Go is only
  for a small exceptional tail. Controlled Go resolution is not arbitrary
  generated state-mutation code.
- **ADR 0008:** retain typed validation, deterministic rendering and conservative
  consumption for the existing compiler. That compiler is now legacy
  correctness/compatibility tooling, not the route to new card coverage.
- **ADR 0009:** retain text-blind runtime behavior and descriptive text as
  nonbehavioral metadata. Supersede the requirement that a serializable CardDef
  alone contains every card's behavior. Authored program identities need an
  installed immutable handler catalog; Go functions do not serialize.
- **ADR 0004:** retain pure game state, one rules Engine, reproducibility and
  fog-of-war observations. Both existing declarative and authored cards use it.

## Migration and current limits

[Epic #3269](https://github.com/natefinch/council4/issues/3269) defines ordered
implementation and acceptance work. This decision does not claim that its future
interfaces already exist. `rules.CardImplementation` currently supports only
`ResolveSpell`; a general safe program catalog/context is step 03 (#3272).
Step 02 (#3270) supplies pinned compiler/registry preservation and read-only,
proportional CI. Mixed support reporting is still step 12 (#3278). Existing
compiler reports measure recognition, not all deployed or behavior-verified cards.

Keep generated definitions, tests, merged improvements and unmerged work.
Freeze compiler coverage expansion; allow explicitly assigned correctness fixes
and regeneration. Preserve genuine runtime bugs when parking a compiler-only
approach. Do not turn this migration into wholesale card regeneration.
