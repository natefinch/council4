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

For a bounded assignment, stop after the assigned batch, a genuine out-of-scope
blocker, or the authorized budget. Broad continuous production follows the
[continuation scope below](#continuation-scope). Never turn budget exhaustion,
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

### Reviewed continuation commands

The coordinator is the **only queue writer**. Workers use isolated code
workspaces and report evidence; they do not compete for manifest writes. Claims
are durable work accounting, not an atomic Git lock. Finish/release the current
claim before replanning; even its owner cannot replace it with a new plan.
`next` resumes the same owner's claim and refuses another owner's takeover.
Ordinary mutations write a complete temporary checkpoint and rename it into
place, refusing an observed intervening edit; exclusive first creation remains
separate. This protects against truncation, not concurrent-writer races.

When no saved batch is executable, `next` returns `batch: null`, saved
`pending_prerequisites`, up to five exact `unreviewed_candidates` in saved order,
and `next_action`. Candidates are only a bounded discovery slice, not a
capability/readiness verdict. Review actual deck priorities, primary sources,
complete declarations, runtime paths and behavioral prior art before selection.
No global inventory refresh or corpus compilation is needed.
Pending output also retains canonical whole-card blockers and their
`blocked_oracle_ids`, including cards outside a saved batch; checked prerequisite
evidence never silently reopens these cards.

Use `-operation add-reviewed-batch -plan PATH -batch NEW_ID -owner SESSION_ID`
to append **one** reviewed three-to-five-card batch. The plan is a small JSON
object, not a replacement manifest:

```json
{
  "corpus_sha256": "EXACT_SAVED_CORPUS_HASH",
  "batches": [{
    "id": "NEW_ID",
    "family": "reviewed-family",
    "oracle_ids": ["ORACLE_ID_1", "ORACLE_ID_2", "ORACLE_ID_3"],
    "prerequisites": [3273],
    "operations": ["Context.Destroy"],
    "review": "source/capability packet reference",
    "prerequisite_evidence": [{
      "issue": 3273,
      "runtime_revision": "CHECKED_COMMIT",
      "evidence": "actual Engine path and executed acceptance scenario/result references"
    }]
  }],
  "records": [
    {"oracle_id": "ORACLE_ID_1", "revision": "PRINTING_ID_1", "name": "Exact Name 1"},
    {"oracle_id": "ORACLE_ID_2", "revision": "PRINTING_ID_2", "name": "Exact Name 2"},
    {"oracle_id": "ORACLE_ID_3", "revision": "PRINTING_ID_3", "name": "Exact Name 3"}
  ]
}
```

The command rejects duplicate batch/member IDs, unknown or stale identities,
non-unclaimed or already-assigned cards, missing review references and active
claims. Existing batch order/history, exclusions, installed/verified states,
baseline associations and record evidence are preserved. Follow with:

```sh
go run ./cmd/cardqueue -operation next -owner SESSION_ID
go run ./cmd/cardqueue -operation claim -batch NEW_ID -owner SESSION_ID
```

`verify` requires the existing exact loadable behavioral admission, then binds
that batch's pinned Oracle ID, printing revision, name, text hash and rulings URI
into `rules_evidence.cards`. It preserves full source records and other rules
fields, rejects conflicting metadata, and never invents source retrieval,
rules-reading or review claims. Record those separately from actual sources.

`prerequisites` lists canonical issues, not an assumption that their closure
establishes behavior. Each needs explicit `prerequisite_evidence` with the
runtime revision and references to the actual checked path and executed
behavioral evidence. Missing entries keep the saved batch pending and
unclaimable; operation names alone never make it ready. Older batch JSON
remains readable without this field, but a fresh claim needs explicit evidence.
Existing same-owner claims remain resumable. Evidence fields are the
coordinator's checked record, not automated proof of rules correctness.
Recheck that saved evidence applies to the current runtime before using it.

After completing a pending prerequisite, use `-operation record-prerequisites
-plan PATH -batch BATCH_ID -owner SESSION_ID`. Supply the same saved batch
definition in `batches`, only newly checked `prerequisite_evidence`, and no
`records`. This appends evidence without changing batch membership or replacing
earlier checks; duplicate/unknown issue evidence is rejected.

Reopen blocked cards only with `-operation reassess -plan PATH -owner SESSION_ID`,
no `-batch`, and a plan containing the corpus hash plus one-to-five exact
`records`, each with:

```json
{
  "oracle_id": "BLOCKED_ORACLE_ID",
  "revision": "EXACT_PRINTING_ID",
  "name": "Exact Name",
  "reassessment": {
    "issue": 999,
    "runtime_revision": "FIXED_COMMIT",
    "evidence": "canonical blocker acceptance probe passed; complete card behavior now expressible"
  }
}
```

The issue must match the card's saved canonical blocker. An assigned blocked
batch must be reassessed as a whole, never silently resized; unbatched cards can
then enter a new reviewed batch. Reassessment retains the original blocker and
reason plus an appended `reassessments` trace. It changes only eligibility to
`unclaimed`, never to supported/verified, and does not satisfy a saved batch's
other prerequisites. A closed issue alone is insufficient.

`block` is a **whole-batch** transition: use it only if the same observed blocker
actually applies to every claimed member. For a gap affecting only some cards,
do not falsely classify ready siblings or partially verify a claim. Instead:

1. `-operation release -batch OLD_ID -owner SESSION_ID` releases the entire
   claim without changing identities or history.
2. `-operation retire-batch -batch OLD_ID -plan PATH -owner SESSION_ID` takes
   the corpus hash and a `records` list containing **only** the affected exact
   identities, with `blocker` (canonical issue number) and `reason` (observed
   missing behavior), not `batches` or `reassessment`. It requires every old
   member to be released/unclaimed, archives the original batch intact in
   `retired_batches`, clears those released memberships and blocks only the
   specified cards. Retired batch IDs cannot be reused.
3. Review eligible siblings and additional unassigned candidates, then
   `add-reviewed-batch` under a new ID with three-to-five exact identities and
   checked prerequisites. Preserve any existing draft/scenario work; retirement
   neither admits it nor discards it. Blocked members remain available for
   explicit reassessment after their prerequisite, never automatic admission.

All commands accept the existing `-manifest`, `-corpus` and `-compiler-report`
paths. If the historical frozen compiler report is absent, the queue joins
the pinned corpus, tracked
`testdata/card-preservation/installed-compiler.json` and actual registry evidence.
It checks the installation's corpus hash and exact identities; it neither
compiles nor invents compiler results for the uninstalled tail. A supplied
malformed report fails rather than silently falling back. Loadability and
behavior verification still come from checked production registrations;
untracked promotions remain errors requiring explicit evidence reconciliation.

### Continuation scope

Establish scope, integration destination, commit/push permissions and resource
limits once per run. A request for a bounded issue or card batch does not
authorize replenishment or unrelated Engine work. Under an explicit assignment
to implement remaining cards **and their required Engine support**, finish and
record a batch, replenish the next reviewed small family, and continue without
another card-selection or strategy conversation.

A demonstrated capability gap routes to its canonical executable prerequisite:
check ownership, retain an expected-correct probe or a complete declaration/
runtime boundary, define the shared invariant and one real proving card, then
implement, validate and independently review the shared change proportionally.
Keep the whole dependent card unsupported. Continue independent ready-card work
only with isolated workspaces and serialized integration, and do not indefinitely
replace an active Engine task with easier batches. After the change, explicitly
reassess exact dependents and all remaining abilities, then return to card
scenarios and production admission. Blocker discovery or an API alone is not
delivered card support.

Tune observed prompt/tooling/batching friction in bounded changes while
delivering; another model comparison, scheduler, corpus census or new framework
is not a prerequisite. Stop at completed scope, an explicit budget, inaccessible
required sources/access, conflicting ownership/integration permissions, an
external rules/product decision, or inability to proceed safely after bounded
investigation/escalation. Preserve exact claims, blockers, evidence, actual
integration status and next action. No saved batch or a needed Engine update
alone does not end broadly authorized production. Durable queue state does not
authorize automation or promise unattended restart.

For explicitly requested Dynamic Workflow execution, use the
[observable production workflow](dynamic-card-production.md). Successful units
are validated, reviewed according to risk and landed on remote `main`; local
commits are not the final delivery condition. Separate card implementation,
shared Engine work, reviews and review fixes in the workflow's measured stages.
The repository workflow/policy can be tuned between runs without relabeling
historical results or weakening the required source/behavior evidence.

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

### Independent reviewer contract

Review the whole printed card against the pinned rules, not just the changed
Go or the author's passing tests. The author's handoff is evidence to inspect,
not an acceptance verdict. Record the baseline and exact patch reviewed, and
keep your initial verdict separate from any later guided repair.
For a delegated review, keep the assigned card and reviewer role: a quoted
overall request is context, not authorization to choose or implement more cards.

1. **Verify primary sources yourself.** Read the exact Oracle printing, its
   full rulings response (including an empty list), and the relevant official
   CR passages. A verified saved primary-source response is sufficient; a new
   network fetch is not mandatory. Record source/edition, retrieval or snapshot
   identity, passages actually read and the behavior they require. An author's
   assertion, copied citation, search snippet or third-party summary is not this
   verification. If a necessary source remains unavailable, say so and withhold
   complete approval rather than substituting memory.
2. **Trace the complete rules path.** Map each nontrivial printed clause to its
   declaration and actual Engine branch: costs/timing, target choice and
   revalidation, resolving choices, mutation and events, replacements, identity,
   source lifetime, visibility and duration as applicable. Check reused helpers
   on the route this card actually takes. Separate a wrong card body from a
   shared runtime defect, false missing-capability report, or bad existing test.
   Neither an API name nor a green suite establishes the rule.
3. **Try to falsify the implementation independently.** Choose the card's
   highest-risk interactions from the rules and runtime trace, not only from the
   author's test list. Add and retain focused expected-correct behavioral
   scenarios that address those risks through the real Engine and exact
   checked card identity. Use faithful Engine-owned identity/creation fixtures
   for tokens and copies, not impossible states manufactured by flipping flags
   on card-backed objects. Establish positive controls for negative-legality
   claims. Assert meaningful observable results, not just declarations, hashes,
   loadability or the absence of a panic. For a nontrivial card, rerunning only
   the author's scenarios is not independent behavioral verification.
4. **Resolve material gaps before approval.** A gap is material when it could
   change a printed ability, legality, payment, information exposure, relevant
   replacement, identity or lifetime outcome. Investigate or execute a focused
   probe rather than listing such a gap under an otherwise complete approval.
   Exhaustive combinations are not required; explain why the chosen probes and
   existing shared regressions address the relevant risks. A redundant network
   fetch or an unrelated scenario is not an approval gate.
5. **Deliver an evidence-backed verdict.** Use `approve` only when the assigned
   review is complete with no known material gap, `changes needed` for
   demonstrated defects or missing required verification, and `inconclusive`
   when a blocker prevents a justified verdict. Distinguish gameplay, source,
   test-contract and evidence findings; cite file/line, expected versus actual
   behavior, primary basis and executable evidence. Report exact commands,
   executed test names/results, your added scenarios and remaining nonmaterial
   limits. A selector matching zero tests or a partial trace is not success.

Respect the review assignment's editing boundary. In a blind trial, retain
reviewer tests without modifying the author's implementation or original tests,
contacting the author, or reading evaluator-only probes. A coordinator validates
findings separately and records any prompted repair without rewriting the
initial outcome. Do not silently lower assertions to accommodate a suspected
bad test: establish the rules-correct counterexample first.

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
> blocker and leave that whole card unsupported. Honor the assigned scope:
> a bounded worker does not expand its role; broad continuous production routes
> the prerequisite, explicitly reassesses dependents and replenishes reviewed
> batches through the coordinator-owned saved queue.
> Follow the repository validation policy. Keep changes
> reviewable and honor the run's commit/push permissions.
