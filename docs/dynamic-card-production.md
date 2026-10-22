# Observable card and Engine production

The repository extension
[`council4-production`](../.github/extensions/council4-production/extension.mjs)
registers two Copilot Dynamic Workflows:

- **`council4-production`**: select, implement, validate, review when required,
  repair/re-review, and publish accepted units to `origin/main`.
- **`council4-production-report`**: refresh a run's native telemetry and local
  task/stage breakdown without spawning agents or modifying Git.

This is an extension, not a GitHub Actions workflow or a scheduled automation.
By user choice, production runs stay in Copilot, **not GitHub Actions**.
Production mode refuses to start when `GITHUB_ACTIONS=true`; CI only runs
ordinary regression tests and does not launch agents or production work.
Copilot discovers `.github/extensions/*/extension.mjs` in a trusted repository.
The SDK import is supplied by Copilot; no npm install is required. A new session
or extension reload loads the repository definitions. Session-scoped registrations
used while authoring are temporary entry points to the same repository code.

## Main is the delivery destination

**Successfully implemented production cards and Engine fixes land on `main`,
not indefinitely on local work branches.** This is the user-directed update to
the original #3311 plan. A production run authorizes this publication within its
explicit scope; preflight, reporting and telemetry-smoke do not.

The workflow creates one isolated worktree per unit from fetched `origin/main`.
Only the orchestrator commits and publishes. It checks the actual diff, runs
affected packages, applies the review policy, records the reviewed tree, and
uses an ordinary fast-forward push to `refs/heads/main`. A fresh fetch verifies
the commit is on the remote branch before a unit becomes `landed`. The user's
main checkout is never read or changed.

There is one shared-repository integration lock, plus serialized units. A claim
is not a writer lock. Non-fast-forward races, repository protections, unexpected
commits or changed reviewed content stop publication; no force push, automatic
reset, bypass or unreviewed conflict resolution. A protection requiring a PR
must be handled explicitly. Local work is preserved and is not counted as
delivered.

The workflow inspects main's classic protection and rulesets before spending on
production and again before pushing. Required CI checks, PR reviews, restricted
writers, or other unsupported integration rules block direct publication even
when the signed-in account has bypass privileges. This version does not create
PRs or manufacture CI statuses to get around that block.

**Current repository blocker (2026-10-22):** classic protection still requires
`Lint`, `Race Tests`, `Vet`, and `Generated cards compile`, while current CI uses
`Tests`. Those protections need an explicitly approved reconciliation and a
checked integration path before automatic main landing can run. They have not
been disabled or changed by this setup. The initial normal setup push succeeded
but GitHub reported an account-level bypass of the four missing checks; this
explicit preflight guard prevents repeating that behavior.

Start from a clean coordinator checkout whose HEAD is on the ancestry of
`origin/main`. Unlanded setup commits must be deliberately integrated first.
Each unit's worktree and failed drafts remain in
`.cardwork/workflows/<run-id>/work/`; do not archive/delete unresolved work.
After an abnormal process kill, inspect the lock's `owner.json` and the native
run state before manually clearing a stale lock. The workflow never steals it.

## Run a bounded production scope

Ask Copilot:

> Run `council4-production` in produce mode for up to three work units, using
> the verified Oracle corpus at `/absolute/path/oracle-cards.json`, with scope
> “remaining cards and their required Engine prerequisites.”

Equivalent arguments:

```json
{
  "mode": "produce",
  "maxTasks": 3,
  "corpusPath": "/absolute/path/oracle-cards.json",
  "scope": "remaining cards and their required Engine prerequisites"
}
```

`maxTasks` is an explicit work scope, **not** an AIC or time ceiling. Set resource
limits in the invocation/app when desired. No guessed limits are embedded.
Runtime limits count resumed attempts and are soft/post-paid for AIC. No new
schedule or background recurrence is enabled by registering this extension.

Two non-delivery checks are available:

```json
{"mode":"preflight"}
```

This checks registration/policy/repository and the supported protection policy
without agents or Git publication. It does not certify write permissions,
production tests or all production inputs. A protection failure is a real
preflight blocker, not a passing production result.

```json
{"mode":"telemetry-smoke"}
```

This runs one short Luna response with no tools, code edits or publication to
exercise real usage attribution. It spends the response's actual AIC.

## Separate work types and stages

Selection is a separately measured coordination agent. Each production unit has
a stable ID and one task type:

| Task type | Scope |
| --- | --- |
| `card` | A ready exact-identity family or concrete card implementation, with real card behavior and admission |
| `engine` | A demonstrated shared rules gap, acceptance scenarios and capability evidence; dependent new-card admissions belong to a later card unit |

A card agent may not silently include shared Engine production changes. When it
finds a gap, its attempt is retained as blocked card work. The next selection sees
the exact gap and retained evidence and can choose an Engine unit. After the
Engine unit lands, later selections reassess dependent cards on the new runtime.
An exhausted saved batch is a reason to prepare the next bounded batch, not a
reason to restart strategy research.
Landing also rejects unfinished queue claims and separates newly verified
uninstalled identities from verification/migration of existing installed support.
Engine units cannot quietly include new card admissions in their cost account.

Each model call has a unique durable label:

```text
card/08-scry-interaction/implementation/0
card/08-scry-interaction/review/0
card/08-scry-interaction/review_fix/1
card/08-scry-interaction/re_review/1
engine/3313/implementation/0
engine/3313/review/0
```

`validation_fix` is distinct from fixes requested by a reviewer. Deterministic
test commands and landing are not falsely attributed to a model call. A new
independent reviewer is used for re-review; repeated identical workflow calls
are memoized by the runtime rather than spawning accidental duplicates.

Engine units always get independent review. Card review defaults to risk-based:
the selection or implementation flags material novelty, and a focused review
then runs. Known-supported routine cards still receive real behavioral tests
and self-inspection. Set `cardReview` to `always` to compare a stricter policy.
The workflow does not treat model approval, strings in Verification, or passing
constructor assertions as sufficient proof of correctness.

Failed tests trigger bounded `validation_fix` work. Review findings trigger
bounded `review_fix`, validation, and re-review. Exhausting those rounds retains
the unsuccessful result and its cost; it never turns rejection into admission.
Null/invalid agent responses and reviewer edits to the submission fail closed.

## See time, AIC and model

The app's **Workflows** view shows named phases, live agents, active time,
requested/resolved model and aggregate AIC. The extension also writes:

| File under `.cardwork/workflows/<run-id>/` | Purpose |
| --- | --- |
| `state.json` | Pinned code/policy revision, stage identities, task outcomes and landing commits |
| `usage.jsonl` | Sanitized authoritative usage receipts keyed by source-session/sequence and agent ID; no prompts or model reasoning |
| `report.json` | Per-task-type/task/stage/model AIC and agent time, outcomes, native totals and reconciliation |
| `report.md` | Readable task/stage/model table and landed/blocked outcomes, suitable for the app's editor panel |
| `costs.csv` | Copyable task/stage/model credit and active-time rows |
| `validation-*.json` | Actual deterministic check results |

**Run AIC** is the native workflow-wide bill, including its subagent subtree.
The native agent-summary API supplies model and timing but no per-agent AIC.
The extension therefore joins actual `assistant.usage` receipts to workflow
agent IDs and durable task labels. It counts each accounting receipt once,
even after replay, and attributes unexpected descendants to their direct stage.
Agents are instructed not to spawn descendants.

Requested and resolved models are both retained; receipt model identifies
actual billed calls. `nanoAiu / 1e9` is AIC, **not USD**. Missing/delayed receipts
remain explicitly incomplete/unallocated rather than becoming guessed costs
or zero-cost successes. `runtimeNanoAiu` is the authoritative total even when
the local itemization is incomplete. A local process crash can lose live receipt
delivery; report refresh cannot invent those missing receipts.

Agent `activeMs` comes from the runtime. Stage timestamps represent elapsed
stage time, which may include a pause/retry interval; run active time excludes
time paused by the runtime. These are not interchangeable and concurrent agent
time would not equal wall time. AIC excludes this outer setup/coordinator chat,
CI compute, and non-model shell cost. Do not omit those from a full business-cost
estimate or pretend they belong to a particular model.

The workflow writes snapshots after stages. Once it settles, ask:

> Run `council4-production-report` with runId `<the-production-run-id>`.

This reads the owning session's native run detail, includes any settled native
totals, and refreshes the report. Reports are session-scoped observations; another
session can read saved JSON but cannot query this session's native run history.

Compare persisted runs without another model:

```sh
node .github/extensions/council4-production/report.mjs
```

The output includes workflow/policy revisions, all stages and failed attempts,
not just accepted cards. No reports are committed automatically.

## Tune without rewriting history

Edit [`policy.json`](../.github/extensions/council4-production/policy.json):

- Card author/fix defaults: `gpt-6-luna`.
- Engine, selection and reviewer defaults: `null`, meaning the session/runtime
  model preference is inherited; the actual model is measured, not assumed.
- Independent model choices for each task type's implementation, review,
  review fixes and validation fixes.
- Risk-based versus always-on card review, bounded repair rounds, and affected
  package validation.

Keep task/phase labels stable. Bump the policy version when changing policy;
every run also records a hash of executable workflow code and policy. Tune between
runs, validate with `node --test .github/extensions/council4-production/workflow.test.mjs`,
review/commit the change, then reload the extension. No corpus rebuild or lint
is required.

Resuming reuses completed durable steps and agent calls. **Changing code/policy
does not retroactively change a paused run**: the workflow rejects a mismatched
revision instead of silently replaying old results under new meanings. Restore
the old revision to resume that run, or start a new one and reconcile preserved
work first. Git publication is idempotent across a crash after commit/push:
only the previously prepared reviewed tree may be recovered.

The definition is versioned in the repository and can improve alongside
production. It is not another Oracle compiler or an autonomous guarantee of
correctness. `scope_finished` means the requested number of units finished;
`partial`, retained blockers and failures remain visible, and neither result
means the entire corpus is done.

References: [Dynamic Workflows](https://docs.github.com/en/copilot/concepts/agent-capabilities/agent-powered-workflows-and-operations/dynamic-workflows),
[using and monitoring workflows](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/use-dynamic-workflows).
