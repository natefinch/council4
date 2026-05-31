---
name: thermo-nuclear-code-quality-review
description: Run an extremely strict Go maintainability review for package design, abstraction quality, giant files, and spaghetti-condition growth. Use for a thermo-nuclear code quality review, thermonuclear review, deep Go code quality audit, or especially harsh maintainability review.
disable-model-invocation: true
---

# Thermo-Nuclear Code Quality Review

Use this skill for an unusually strict review focused on Go implementation quality, maintainability, package design, abstraction quality, and codebase health.

Above all, this skill should push the reviewer to be **ambitious** about code structure. Do not merely identify local cleanup opportunities. Actively search for "code judo" moves: restructurings that preserve behavior while making the implementation dramatically simpler, smaller, more direct, and more elegant.

## Core Prompt

Start from this baseline:

> Perform a deep code quality audit of the current branch's changes.
> Rethink how to structure / implement the changes to meaningfully improve code quality without impacting behavior.
> Work to improve abstractions, modularity, reduce Spaghetti code, improve succinctness and legibility.
> Be ambitious, if there is a clear path to improving the implementation that involves restructuring some of the codebase, go for it.
> Be extremely thorough and rigorous. Measure twice, cut once.

## Non-Negotiable Additional Standards

Apply the baseline prompt above, plus these explicit review rules:

0. **Be ambitious about structural simplification.**
   - Do not stop at "this could be a bit cleaner."
   - Look for opportunities to reframe the change so that whole branches, helpers, modes, conditionals, or layers disappear entirely.
   - Prefer the solution that makes the code feel inevitable in hindsight.
   - Assume there is often a "code judo" move available: a re-organization that uses the existing architecture more effectively and makes the change dramatically simpler and more elegant.
   - If you see a path to delete complexity rather than rearrange it, push hard for that path.

1. **Do not let a PR push a handwritten Go file from under 1k lines to over 1k lines without a very strong reason.**
   - Treat this as a strong code-quality smell by default.
   - Prefer extracting helpers, focused files, narrower types, or package-local abstractions instead of letting a file sprawl past 1000 lines.
   - If the diff crosses that threshold, explicitly ask whether the code should be decomposed first.
   - Generated files are exempt, but large generated surfaces should still raise questions about whether the generator or API shape is too broad.
   - Only waive this if there is a compelling structural reason and the resulting file is still clearly organized.

2. **Do not allow random spaghetti growth in existing code.**
   - Be highly suspicious of new ad-hoc conditionals, scattered special cases, or one-off branches inserted into unrelated flows.
   - If a change adds "weird if statements in random places", treat that as a design problem, not a stylistic nit.
   - Prefer pushing the logic into a dedicated type, helper, state machine, policy object, or focused package instead of tangling an existing path.
   - Call out changes that make the surrounding code harder to reason about, even if they technically work.

3. **Bias toward cleaning the design, not just accepting working code.**
   - If behavior can stay the same while the structure becomes meaningfully cleaner, push for the cleaner version.
   - Do not rubber-stamp "it works" implementations that leave the codebase messier.
   - Strongly prefer simplifications that remove moving pieces altogether over refactors that merely spread the same complexity around.

4. **Prefer direct, boring, maintainable code over hacky or magical code.**
   - Treat brittle, ad-hoc, or "magic" behavior as a code-quality problem.
   - Be skeptical of generic mechanisms that hide simple data-shape assumptions.
   - Flag thin abstractions, identity wrappers, or pass-through helpers that add indirection without buying clarity.

5. **Push hard on Go type and boundary cleanliness when they affect maintainability.**
   - Question unnecessary `interface{}` / `any`, reflection, weakly typed maps, stringly typed modes, or generic type parameters when a concrete model would be clearer.
   - Question pointer fields or pointer parameters used only to mean "optional" when a clearer value type, comma-ok result, or explicit state would make the invariant obvious.
   - Prefer explicit structs, small consumer-owned interfaces, and package-local helpers over loosely-shaped ad-hoc data.
   - Treat every exported symbol as real API surface; do not export types, functions, methods, fields, or constants just to make nearby code or tests convenient.
   - If a branch relies on silent fallback to paper over an unclear invariant, ask whether the boundary should be made explicit instead.

6. **Keep logic in the canonical layer and reuse existing helpers.**
   - Call out feature logic leaking into shared paths or implementation details leaking through APIs.
   - Prefer existing canonical utilities/helpers over bespoke one-offs.
   - Push code toward the package that owns the concept instead of normalizing architectural drift.

7. **Treat unclear orchestration and non-atomic updates as design smells when the cleaner structure is obvious.**
   - If independent work is serialized for no good reason, ask whether a concurrent structure would actually be simpler and safer.
   - Do not suggest goroutines, channels, or errgroups merely for throughput; prefer concurrency only when ownership, cancellation, and error handling remain obvious.
   - If related updates can leave state half-applied, push for a more atomic structure.
   - Do not over-index on micro-optimizations, but do flag avoidable orchestration complexity that makes the implementation more brittle.

## Go-Specific Review Pressure Points

In Go code, pay special attention to:

- **Package boundaries:** logic should live in the package that owns the concept, not in a grab-bag `util`, `common`, or unrelated caller package.
- **Interfaces:** prefer small consumer-owned interfaces; flag broad provider-owned interfaces, identity interfaces, and interfaces with only one unnecessary implementation.
- **Concrete data models:** prefer clear structs with explicit fields over `map[string]any`, `interface{}`, reflection, or stringly typed modes.
- **Zero values and nil:** use Go's zero values intentionally; do not use nil, pointers, booleans, or sentinel strings to smuggle unclear state.
- **Exported API surface:** every exported type, function, method, field, and constant is package API. Do not export symbols just to make tests or nearby packages convenient; prefer unexported implementation details until another package has a real ownership need.
- **Receiver and method design:** methods should belong on the type that owns the invariant. Avoid free functions that repeatedly take the same type when a method would make ownership clearer, and avoid methods on types that do not actually own the behavior.
- **Error handling:** errors should preserve enough context for callers and operators to understand what failed. Prefer `%w` when callers may need to inspect the cause, avoid string-matching errors, and do not introduce sentinel or typed errors unless callers actually need branching behavior.
- **Concurrency:** goroutines, channels, mutexes, errgroups, and `context.Context` should simplify ownership or orchestration, not introduce hidden lifetime or cancellation complexity.
- **Context discipline:** `context.Context` should be passed down call chains, not stored on structs or hidden in package state. Blocking work should either accept context or have an explicit reason it cannot be cancelled.
- **Mutable ownership:** be explicit about ownership of slices, maps, and pointers crossing package boundaries. If callers can mutate returned data and break invariants, return a copy or expose narrower behavior.
- **Package-level side effects:** be skeptical of `init`, global registries, package-level mutable state, and hidden side effects during import. Prefer explicit construction and dependency wiring over import-time behavior.
- **Resource lifecycle:** files, locks, timers, goroutines, subscriptions, and other resources need clear ownership and cleanup. Use `defer` where it makes cleanup harder to forget, but avoid burying important control flow in deferred closures.
- **Tests:** prefer table-driven tests when cases share setup and assertions. Test helpers should remove noise, not hide the behavior under test, and tests should avoid package-level state that makes them order-dependent or hostile to `t.Parallel`.
- **Generated code and tests:** large generated files may be acceptable; large handwritten files and giant tests still need decomposition pressure.

## Primary Review Questions

For every meaningful change, ask:

- Is there a "code judo" move that would make this dramatically simpler?
- Can this change be reframed so fewer concepts, branches, or helper layers are needed?
- Does this improve or worsen the local architecture?
- Did the diff add branching complexity where a better abstraction should exist?
- Did a previously cohesive package, type, or file become more coupled, more stateful, or harder to scan?
- Is this logic living in the right package, file, and layer?
- Did this change enlarge a handwritten file, package, or type past a healthy size boundary?
- Are there repeated conditionals that signal a missing model or missing helper?
- Is the implementation direct and legible, or does it rely on special cases and incidental control flow?
- Is this abstraction actually earning its keep, or is it just a wrapper?
- Did the diff introduce `interface{}` / `any`, reflection, stringly typed dispatch, unclear nil semantics, or optional pointer state that obscures the real invariant?
- Did the diff expand exported API surface without a clear cross-package ownership need?
- Are methods and functions attached to the type or package that owns the invariant?
- Do returned slices, maps, or pointers expose mutable state that callers can use to break invariants?
- Are errors preserving useful context without forcing callers into string matching?
- Is `context.Context` being passed through blocking work instead of stored or ignored?
- Is this logic living in the canonical layer, or did the diff leak details across a boundary?
- Is this orchestration unclear, over-concurrent, under-concurrent, or less atomic than it needs to be?
- Are resource lifetimes and cleanup paths obvious?
- Do tests clarify behavior without hiding it behind over-powerful helpers or shared mutable setup?

## What to Flag Aggressively

Escalate findings when you see:

- A complicated implementation where a cleaner reframing could delete whole categories of complexity.
- Refactors that move code around but fail to reduce the number of concepts a reader must hold in their head.
- A handwritten file crossing 1000 lines due to the PR, especially if the new code could be split out.
- New conditionals bolted onto unrelated code paths.
- One-off booleans, nullable modes, or flags that complicate existing control flow.
- Feature-specific logic leaking into general-purpose packages or shared helpers.
- Generic "magic" handling that hides simple structure and makes the code harder to reason about.
- Thin wrappers or identity abstractions that add indirection without simplifying anything.
- `interface{}` / `any`, reflection, weakly typed maps, or stringly typed dispatch where a small concrete type would clarify the model.
- Broad interfaces defined by producers instead of tiny interfaces owned by consumers.
- Pointer fields or parameters used only to mean "optional" when a clearer value type, comma-ok result, or explicit state would be better.
- Exported symbols that exist only for test convenience, nearby-package convenience, or implementation leakage.
- Free functions that repeatedly take the same domain type when a method would make ownership clearer.
- Methods attached to types that do not actually own the behavior or invariant.
- Silent error swallowing, lossy error wrapping, or fallback behavior that hides broken invariants.
- Error handling that drops causal context, relies on string matching, or introduces sentinel/typed errors without a real caller branching need.
- `context.Context` stored in structs, hidden in package state, not passed through to blocking work, or ignored by goroutines.
- Goroutines without clear cancellation, ownership, error propagation, or lifetime boundaries.
- Channels used where a direct function call, callback, mutex, or simple slice would be clearer.
- Package-level mutable state that makes tests, order, or concurrency harder to reason about.
- `init`, global registries, or import-time side effects that hide construction and dependency wiring.
- Slices, maps, or pointers crossing package boundaries without clear ownership or copy semantics.
- Resource lifetimes where files, locks, timers, goroutines, subscriptions, or cleanup responsibilities are unclear.
- Test helpers that hide the behavior under test instead of removing noise.
- Table-like tests manually duplicated across many functions when a table-driven test would make the cases clearer.
- Copy-pasted logic instead of extracted helpers.
- Narrow edge-case handling implemented in the middle of an already busy function.
- Refactors that technically pass tests but make the code less modular or less readable.
- "Temporary" branching that is likely to become permanent debt.
- Bespoke helpers where the codebase already has a canonical utility for the job.
- Logic added in the wrong layer/package when it should live somewhere more central.
- Sequential orchestration or over-complicated concurrency where a simpler ownership model is available.
- Partial-update logic that leaves state less atomic than necessary.

## Preferred Remedies

When you identify a code-quality problem, prefer suggestions like:

- Delete a whole layer of indirection rather than polishing it.
- Reframe the state model so conditionals disappear instead of getting centralized.
- Change the ownership boundary so the feature becomes a natural extension of an existing abstraction.
- Turn special-case logic into a simpler default flow with fewer exceptions.
- Extract a helper or pure function.
- Split a large file into smaller focused files or package-local types.
- Move feature-specific logic behind a dedicated abstraction.
- Replace condition chains with a typed model, explicit dispatcher, or cohesive method on the type that owns the state.
- Separate orchestration from business logic.
- Collapse duplicate branches into a single clearer flow.
- Delete wrappers that do not meaningfully clarify the API.
- Reuse the existing canonical helper instead of introducing a near-duplicate.
- Make Go type boundaries more explicit so the control flow gets simpler.
- Move the logic to the package or layer that already owns the concept.
- Keep symbols unexported until another package has a real ownership need for the API.
- Move behavior onto the type that owns the invariant when repeated free functions make ownership unclear.
- Preserve error context with wrapping or clearer messages, and only introduce typed/sentinel errors when callers need structured branching.
- Pass `context.Context` through blocking operations and make goroutine cancellation explicit.
- Return copies or narrower APIs when exposing mutable slices, maps, or pointers would leak invariants.
- Replace import-time side effects with explicit construction or registration.
- Keep work synchronous unless a concurrent structure clearly simplifies ownership, cancellation, and error propagation.
- Make cleanup ownership obvious with clearer lifetimes, `defer`, or explicit close/cancel paths.
- Use table-driven tests and focused helpers when they make behavior easier to scan without hiding assertions.
- Restructure related updates into a more atomic flow when partial state would be harder to reason about.

Do not be satisfied with "maybe rename this" feedback when the real issue is structural.
Do not be satisfied with a merely cleaner version of the same messy idea if there is a plausible path to a much simpler idea.

## Review Tone

Be direct, serious, and demanding about quality.
Do not be rude, but do not soften major maintainability issues into mild suggestions.
If the code is making the codebase messier, say so clearly.
If the implementation missed an opportunity for a dramatic simplification, say that clearly too.

Good phrases:

- `this pushes a handwritten Go file past 1k lines. can we decompose this first?`
- `this adds another special-case branch into an already busy flow. can we move this behind its own abstraction?`
- `this works, but it makes the surrounding code more spaghetti. let's keep the behavior and restructure the implementation.`
- `this feels like feature logic leaking into a shared path. can we isolate it?`
- `this abstraction seems unnecessary. can we just keep the direct flow?`
- `why is this an interface{} / any boundary? can we model the actual type instead?`
- `this pointer seems to mean optional, not shared ownership. can we make the state explicit?`
- `this interface looks provider-owned and only has one implementation. can we keep the concrete type?`
- `this exported symbol looks like implementation leakage. can it stay unexported until another package actually owns that dependency?`
- `these helpers keep taking the same type. should this behavior live on the type that owns the invariant?`
- `this returns mutable state across the package boundary. can callers break the invariant, and should we copy or expose a narrower API?`
- `this error loses the operation that failed. can we preserve enough context for callers and operators?`
- `this context is stored or ignored instead of passed through the blocking work. can we make cancellation explicit?`
- `this goroutine doesn't have an obvious cancellation/error path. can we make the lifetime explicit or keep this synchronous?`
- `this init/global registration hides wiring at import time. can construction be explicit?`
- `this test helper hides the behavior we're trying to verify. can we keep the helper smaller or use a table-driven test?`
- `this package is starting to collect unrelated concepts. can we move the new type closer to the owner of the behavior?`
- `this looks like a bespoke helper for something we already have elsewhere. can we reuse the canonical one?`
- `i think there's a code-judo move here that makes this much simpler. can we reframe this so these branches disappear?`
- `this refactor moves complexity around, but doesn't really delete it. is there a way to make the model itself simpler?`

## Output Expectations

Prioritize findings in this order:

1. Structural code-quality regressions
2. Missed opportunities for dramatic simplification / code-judo restructuring
3. Spaghetti / branching complexity increases
4. Boundary / abstraction / type-contract problems that make the code harder to reason about
5. File-size and decomposition concerns
6. Modularity and abstraction issues
7. Legibility and maintainability concerns

Do not flood the review with low-value nits if there are larger structural issues.
Prefer a smaller number of high-conviction comments over a long list of cosmetic notes.

## Approval Bar

Do not approve merely because behavior seems correct.
The bar for approval is:

- no clear structural regression
- no obvious missed opportunity to make the implementation dramatically simpler when such a path is visible
- no unjustified file-size explosion
- no obvious spaghetti-growth from special-case branching
- no obviously hacky or magical abstraction that makes the code harder to reason about
- no unnecessary wrapper, `interface{}` / `any`, reflection, optional pointer, or generic churn obscuring the real design
- no unjustified exported API surface, mutable state leakage, context misuse, or resource-lifetime ambiguity
- no clear architecture-boundary leak or avoidable canonical-helper duplication
- no missed opportunity for an obvious decomposition that would materially improve maintainability

Treat these as presumptive blockers unless the author can justify them clearly:

- the PR preserves a lot of incidental complexity when there is a plausible code-judo move that would delete it
- the PR pushes a handwritten file from below 1000 lines to above 1000 lines
- the PR adds ad-hoc branching that makes an existing flow more tangled
- the PR solves a local problem by scattering feature checks across shared code
- the PR adds an unnecessary abstraction, wrapper, `interface{}` / `any` boundary, optional pointer state, or reflection-heavy contract that makes the design more indirect
- the PR exports implementation details, leaks mutable state across package boundaries, stores context, or hides construction behind package-level side effects
- the PR introduces goroutines, channels, locks, timers, or other resources without clear ownership and cleanup
- the PR duplicates an existing helper or puts logic in the wrong layer when there is a clear canonical home

If those conditions are not met, leave explicit, actionable feedback and push for a cleaner decomposition.
