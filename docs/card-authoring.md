# Authoring complete cards in Go

This is the canonical new-card workflow. Legacy compiler maintenance is separate.
The entry skill is [card-authoring](../.agents/skills/card-authoring/SKILL.md).
Programs and declarations share the production Engine and registry.

Battlefield-entry declarations are evaluated against the pending permanent's
current would-be battlefield characteristics. The Engine recomputes applicability
after each replacement, applies self-replacement/copy precedence where represented,
and records the affected controller's choice among competing effects. Copied
entry abilities participate in the same process; lost abilities do not survive
on a frozen list. Entry payments are offered only when their replacement is
selected; an unpaid entry-to-zone replacement prevents battlefield publication.
Entry choices run before publication of the permanent or its
simultaneous cohort, and entering tapped/untapped is not a tap/untap event.

## A bounded worker loop

1. Read the selected live issue and real prerequisites. Claim a small eligible
   batch; respect active ownership, one main writer, scope and commit/push
   permissions. For a direct card request, do the work without requiring an issue.
2. Resolve exact Oracle ID/face and current text from the
   [pinned corpus](../testdata/card-preservation/corpus.json). Read that printing's
   Scryfall `rulings_uri` and the applicable official Comprehensive Rules; record
   revisions. Check the relevant production registry entries and scenario evidence,
   not compiler acceptance or a name match alone. The existing
   [inventory](../support-inventory.md) is a report, not a required refresh.
3. Inspect the nearest authored card and
   [shared operation API](../mtg/cardprogram/README.md). Draft normal Go:
   printed costs, targets, modes, triggers and static/replacement applicability
   stay inspectable; synchronous resolution may use ordinary control flow.
   Mutable links, counters and delayed references stay Engine-owned and clone-safe.
4. Execute named-card scenarios through the real Engine. Exercise choices,
   target changes/all-illegal targets, source departure, relevant replacements
   and timing, not just CardDef shapes or a smoke game's winner.
   Follow the [validation and review policy](../.github/copilot-instructions.md#verify-proportionally).
5. Register only complete behavior-verified entries in
   `mtg/cards/authored/cards.go`, with exact Oracle identity and program versions.
   Set specific `Verification` evidence and retain program dependencies.
   Program-free static cards are valid; do not add a dummy body.
   Explicit legacy replacement is permitted
   only for a verified unique identity, never to hide an ambiguous name.
6. Complete the policy's affected checks and inspect the diff; no corpus,
   fingerprint or global report refresh is needed. Commit intended files to main
   only when authorized; push only with
   separate permission. Record meaningful behavior, blockers and commit once,
   explicitly distinguishing local from published support.

Stop after the assigned batch, a genuine blocker, or the authorized budget.
Continue a production queue only when assigned. Never turn budget exhaustion,
unsupported behavior or partial cards into success. No automatic schedules,
mandatory PRs, per-file agents or per-card engine frameworks.

## Saved production queue

The [production manifest](../testdata/card-authoring/production.json) retains the
entire pinned remaining population, policy exclusions, name-only baseline
associations, claims, verified admissions and precise blockers. Use
`go run ./cmd/cardqueue -operation next -owner YOUR_SESSION_ID` to resume a saved
claim or show the next reviewed batch, then `-operation claim -batch BATCH_ID
-owner YOUR_SESSION_ID` **before editing**. This command does not execute cards
or enable automation. The [production evidence and continuation](card-production.md)
record the completed first three batches, the next three exact identity lists,
live issue reconciliation, measurements and stopping conditions. Keep this guide's
ordinary authoring/scaffold/gates; the queue is work accounting, not another DSL.

The completed legacy compiler import marked newly loadable records
`installed-compiled`, not `implemented-verified`. These records need no duplicate
source generation and are skipped by the saved authored batches. Known runtime
blocks remain blocked even when their generated sources exist. This historical
import is not a step to repeat for new support.

Named issue follow-ups can record `implemented-verified` admissions without
inventing or resizing a saved queue batch. Exact loadable, behavior-verified
registry identity and scenario evidence are still required; existing batch
membership and claim validation remain unchanged.

## Executable draft scaffold

```sh
go run ./cmd/cardauthor \
  -oracle-id fae37e28-e137-4177-b973-fa8b4dd8f409 \
  -symbol GenerousGift -out .cardwork/scaffold/generous-gift
```

The command reads metadata verbatim from the pinned input and emits ordinary
formatted Go: an unregistered entry/constructor, one program declaration/body,
and a real Engine/stack scenario skeleton. It does **not** interpret Oracle
English or decide card behavior. Fill metadata/declarations, replace the explicit
unsupported body and final failing assertion, then copy the draft to authored
and rules packages. The scratch output is not compiled into production.
No automatic registry edit or false successful scenario is generated.
Existing output is never overwritten. Multiface input fails explicitly; author
its exact faces using current CardDef APIs instead of guessing.

For static/triggered cards, replace the example spell declaration with the
ordinary declaration they actually need. The template is not a new DSL and
requires no per-card renderer. `go test ./cmd/cardauthor` checks its input and
ordinary Go output. The held-out challenges are not scaffold examples.

## Missing rules versus authoring friction

Fix a broken shared authoring seam instead of selecting easier cards to hide
low yield. If actual rules are missing, capture the concrete operation/timing
limitation and observable completion condition; search existing issues before
filing through the issue tool. Use
[card-support-todo](../.agents/skills/card-support-todo/SKILL.md).
Leave the **whole** card unsupported; compiler rejection alone is not a blocker.

Record exact identities/revisions, implemented/blocked outcomes, reused
operations, meaningful shared changes, scenario references and changed files.
Report observed effort only; omit guesses about time, popularity and coverage.
The frozen [step-13 pilot](../testdata/card-authoring/pilot.json) is evidence,
not a growing production backlog or an acceptance waiver.

## Fresh worker prompt

> Add the assigned cards using the authored-Go library. Read current Oracle
> text/rulings and the applicable Comprehensive Rules. Reuse shared Engine
> operations and inspectable declarations; write normal Go control flow where
> needed. Do not modify the Oracle parser/compiler/lowerer to gain coverage.
> Do not mutate raw Game state, fake results, or omit abilities. Test actual
> behavior, including relevant replacements, target changes, source departure
> and choices. If a required Engine capability is missing, link a precise shared
> blocker and leave that whole card unsupported; continue other eligible cards.
> Follow the repository validation policy. Keep changes
> reviewable and honor the run's commit/push permissions.
