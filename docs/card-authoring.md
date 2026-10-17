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

Mana abilities must use supported inspectable declarations and instruction
sequences, not `Mode.Program`. Installing a program does not establish mana-ability
classification; program-backed mana bodies are rejected at checked admission.
Ordinary non-mana activations, triggers and resolving spells can use Go programs.

## A bounded worker loop

1. Read the selected live issue and real prerequisites. Claim a small eligible
   batch; respect active ownership, one main writer, scope and commit/push
   permissions. For a direct card request, do the work without requiring an issue.
2. Resolve exact Oracle ID/face and current text from the
   [pinned corpus](../testdata/card-preservation/corpus.json). Read that printing's
   Scryfall `rulings_uri`. **Before choosing an implementation, open and read the
   applicable sections of the official Comprehensive Rules** using the
   [source-review requirements below](#required-comprehensive-rules-read).
   Record the actual sources and revisions consulted. Check the relevant
   production registry entries and scenario evidence, not compiler acceptance
   or a name match alone. The existing
   [inventory](../support-inventory.md) is a report, not a required refresh.
3. Inspect the nearest authored card and
   [capability lookup](../mtg/cardprogram/README.md#capability-lookup), then follow
   the linked declaration, Engine path and scenario. Draft normal Go:
   printed costs, targets, modes, triggers and static/replacement applicability
   stay inspectable; synchronous resolution may use ordinary control flow.
   Mutable links, counters and delayed references stay Engine-owned and clone-safe.
4. Execute named-card scenarios through the real Engine. Exercise choices,
   target changes/all-illegal targets, source departure, relevant replacements
   and timing, not just CardDef shapes or a smoke game's winner.
   Follow the [validation and review policy](../.github/copilot-instructions.md#verify-proportionally).
5. Register only complete behavior-verified entries in
   `mtg/cards/authored/cards.go`, with exact Oracle identity and program versions.
   Set specific `Verification` evidence and retain program dependencies; the
   checked registry rejects missing/blank evidence and the old generic fallback.
   Evidence presence is an admission requirement, not proof of gameplay.
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

For activated, triggered or static abilities, replace the example spell
declaration and program reference with the ordinary declarations they actually
need; do not retain the sample Draw or a dummy spell body. The template is not a
new DSL and requires no per-card renderer. `go test ./cmd/cardauthor` checks its
input and ordinary Go output. The held-out challenges are not scaffold examples.

## Evidence and review handoff

### Required Comprehensive Rules read

This is a required input to implementation, not a citation task performed after
the tests pass. Use an official rules document or a verified local copy, confirm
its effective edition against the run's rules pin, and read the relevant
sections before designing the card. If no verified copy was supplied, obtain
the TXT or PDF from the [official Magic rules page](https://magic.wizards.com/en/rules);
do not guess dated download URLs or silently substitute a different edition.
Read the relevant passages, not just search-result snippets or a table of
contents. Reading the entire document for each card is unnecessary.

For each nontrivial clause, identify which rules govern its implementation and
read the necessary supporting rules: costs and timing, target selection and
revalidation, triggering and resolution, replacements, layers, zone changes
and object identity, as applicable. Record the document URL or local path,
effective edition, specific sections actually read, and the concrete behavior
those passages require. Keep this in the existing source record/handoff; no
separate per-card rules document is required. Rulings complement this read but
do not replace it. Neither remembered rule numbers nor citations copied from
another implementation establish that the document was consulted.

If the official document or verified copy is inaccessible, or its edition
cannot be established, stop verification and report the source-review blocker.
Exploration may be reported as incomplete work, but do not register a new card
as behavior-verified or claim source review complete. A reviewer should request
this source record and check that the cited passages support the behavior,
rather than accepting a list of plausible rule numbers.

Keep a concise source record with the Oracle printing revision, rulings URL and
retrieval result (including an empty list), and official rules document edition
and relevant sections actually read. A verified local copy is sufficient; a
date copied from another card, a repository comment or remembered ruling is not
a source read. If a required source cannot be retrieved or verified, report the
gap instead of claiming complete verification. Reference the card's real named
scenarios in its `Verification`; do not invent evidence or use a generic note
about prior work.

Trace reused operations through the actual runtime branch as well as new code.
A helper's existence or another card's passing tests does not establish every
replacement, information-visibility or lifetime behavior required here. For a
negative legality test, show that the same action is otherwise legal before
applying the restriction; an already illegal action cannot prove the rule.
When independent review is requested or required by repository policy, give the
reviewer the implementation, named scenarios, source record and limitations,
not a claimed passing verdict. Keep gameplay findings separate from evidence
gaps, and rerun the finding's regression after a fix rather than certifying it
from the author's explanation alone.

## Missing rules versus authoring friction

An unfamiliar API is not a missing rule. Before declaring a blocker, read the
**complete relevant declarations**, continuing beyond a tool's first range or
truncated output. Search for the rules concept as well as candidate identifiers;
do not infer absence from guessed field names. Follow promising declarations
through their Engine implementation and the nearest actual behavioral tests.
Check declaration-time legality as well as resolving operations: a resolution
body cannot repair incorrect costs, target selection or revalidation. A general
operation may already supply a specialized card's behavior, while an existing
field may not be honored by the particular runtime path.

Fix a broken shared authoring seam instead of selecting easier cards to hide
low yield. When fixing an existing Engine operation, inspect every touched
branch and preserve its other supported behavior, including choices, visibility,
events and destination handling. Add a regression for the discovered bug and
exercise the neighboring affected paths; the new card's passing scenarios alone
do not establish that a shared change is safe.

For zone movement, keep cards in their actual source zone until the
[shared transition machinery](../mtg/rules/README.md#shared-zone-transitions)
applies replacements and commits the move. Do not bypass it with raw zone
removals/additions or add card-specific commander handling. Commander SBAs
remain at Engine checkpoints, not inside resolving card bodies.

If actual rules are missing, capture the concrete operation/timing
limitation and observable completion condition; search existing issues before
filing through the issue tool. Use
[card-support-todo](../.agents/skills/card-support-todo/SKILL.md).
Leave the **whole** card unsupported; compiler rejection alone is not a blocker.

Support a blocker with a smallest relevant behavioral probe when the existing
API can express the attempted behavior. If it cannot, cite the complete
declaration and runtime boundary that prevent it, plus the nearest alternatives
checked; do not add an incorrect approximation merely to make a test fail.
Distinguish a demonstrated missing capability from an unsuccessful search or an
unfinished implementation. Report which named scenarios actually executed:
passing scaffold tests, other cards' tests, or a selector matching no tests is
not named-card verification.

Record exact identities/revisions, implemented/blocked outcomes, reused
operations, meaningful shared changes, scenario references and changed files.
Report observed effort only; omit guesses about time, popularity and coverage.
The frozen [step-13 pilot](../testdata/card-authoring/pilot.json) is evidence,
not a growing production backlog or an acceptance waiver.

## Fresh worker prompt

> Add the assigned cards using the authored-Go library. Read current Oracle
> text/rulings. Before designing the implementation, open the official
> Comprehensive Rules or a verified local copy, read the applicable sections,
> and record the edition, source and behavior they require; do not substitute
> remembered or copied citations. Reuse shared Engine
> operations and inspectable declarations; write normal Go control flow where
> needed. Do not modify the Oracle parser/compiler/lowerer to gain coverage.
> Do not mutate raw Game state, fake results, or omit abilities. Test actual
> behavior, including relevant replacements, target changes, source departure
> and choices. If a required Engine capability is missing, link a precise shared
> blocker and leave that whole card unsupported; continue other eligible cards.
> Follow the repository validation policy. Keep changes
> reviewable and honor the run's commit/push permissions.
