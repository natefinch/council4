# Observable card and Engine production

The repository extension
[`council4-production`](../.github/extensions/council4-production/extension.mjs)
registers two Copilot Dynamic Workflows:

- **`council4-production`**: select, implement, validate, review when required,
  repair/re-review, open a PR, repair actual CI failures, and merge into `main`
  only after CI succeeds.
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

**User-directed landing policy:** use PRs and merge after CI, not direct main
pushes. The remaining GitHub `Race Tests` protection stays enabled. CI now has a
PR job named exactly `Race Tests`, running `go test -race ./...`; production
agents themselves still run locally in Copilot, not in Actions.

The workflow creates an isolated worktree per unit from fetched `origin/main`.
Only the orchestrator commits and publishes. After local validation and the
applicable independent review it commits the reviewed tree, fetches the latest
main, and merges any new main changes into the unit branch. Changed integration
receives independent review and the merged head runs affected local tests.
Merge conflicts preserve the work and stop rather than being auto-resolved.

The orchestrator pushes **only the task branch**, creates or reuses its PR, and
watches CI. `Tests`, `Race Tests`, every other reported check and any additional
repository-required checks must finish successfully for that exact PR head.
The two named test checks must be actual GitHub Actions check runs, not manual
status assertions. Required approvals are honored. Failed, missing, pending,
skipped-required, stale-head, draft or unmergeable PRs do not merge.

Actual CI failures can trigger bounded `ci_fix` work, local validation and an
independent `ci_review`. Only then does the workflow commit/push the repair and
wait for fresh CI on its new head. Missing checks or external approvals pause
with the PR and resume checkpoint, not a manufactured success. If main advances,
the workflow integrates it and repeats the affected checks/review before merging.

A final fresh PR/check/protection read precedes ordinary
`gh pr merge --squash --match-head-commit`. There is no `--admin`, direct
`refs/heads/main` push, false check status, protection mutation or force push.
The PR's returned squash commit must appear in fetched remote main before the
unit becomes `landed`; the original branch commit is not incorrectly required
to be an ancestor after squash.

There is one shared-repository integration lock and serialized units. A claim
is not a writer lock. Unexpected commits, changed reviewed content, unsupported
protections or conflicts stop safely. Local/unmerged PR work is preserved and
never counted as delivered. The user's main checkout is never read or changed.

Start from a clean coordinator checkout whose HEAD is on the ancestry of
`origin/main`, or whose tree exactly matches main after a squash merge.
Unlanded setup content must be deliberately integrated first.
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
For a card unit the selector returns the exact saved batch ID, or a bounded
reviewed plan in the existing `add-reviewed-batch` format. The deterministic
coordinator saves that plan if needed and claims the whole batch **before**
launching the card worker. The worker receives confirmed owner/identity evidence
and never waits for a separate claim message or edits the saved queue itself.
After complete implementation, the coordinator runs checked registry/source
admission through `cardqueue verify` before validation/review. Failed or partial
admission cannot become a successful unit.
Landing also rejects unfinished queue claims and separates newly verified
uninstalled identities from verification/migration of existing installed support.
Engine units cannot quietly include new card admissions in their cost account.

Each model call has a unique durable label:

```text
card/08-scry-interaction/unit-0/implementation/0
card/08-scry-interaction/unit-0/review/0
card/08-scry-interaction/unit-0/review_fix/1
card/08-scry-interaction/unit-0/re_review/1
card/08-scry-interaction/unit-0/ci_fix/1
card/08-scry-interaction/unit-0/ci_review/1
card/08-scry-interaction/unit-0/integration_review/1
engine/3313/unit-1/implementation/0
engine/3313/unit-1/review/0
```

The unit index distinguishes retrying the same family after an Engine change
from its earlier failed attempt. Policy versions before 1.2.0 did not record
this dimension and can combine repeated task/stage rows; their native totals
remain valid, but the missing unit attribution is not retroactively invented.

`validation_fix` is distinct from `review_fix` and `ci_fix`. CI fixes have their
own task-type model choice; CI and local-test elapsed time are deterministic
stages, not falsely attributed as model active time or AIC. A new
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
| `state.json` | Pinned code/policy revision, stage identities, task outcomes, PRs and verified merge commits |
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
  review fixes, validation fixes and CI fixes.
- Risk-based versus always-on card review, bounded local/review/CI repair rounds, and affected
  package validation.

Keep task/phase labels stable. Bump the policy version when changing policy;
every run also records a hash of executable workflow code and policy. Tune between
runs, validate with `node --test .github/extensions/council4-production/workflow.test.mjs`,
review/commit the change, then reload the extension. No corpus rebuild or lint
is required.

Reviewed file fingerprints use framed paths, executable state and actual file
bytes (or deletion), independent of staging and rename detection. Staging a
new source/test file must not masquerade as a post-review content change.
The final Git tree is checked separately before commit. Policy 1.2.0 includes
regressions for this failure and deterministic claim/admission handoffs after
the first production run exposed both.

Resuming reuses completed durable steps and agent calls. **Changing code/policy
does not retroactively change a paused run**: the workflow rejects a mismatched
revision instead of silently replaying old results under new meanings. Restore
the old revision to resume that run, or start a new one and reconcile preserved
work first. Git publication is idempotent across a crash after commit/push:
only the previously prepared reviewed tree may be recovered. Integration merge
and reviewed CI-commit intents are recorded before Git mutation; interruption
recovery verifies their exact parents/tree before continuing. An existing PR is
reused, never duplicated, and an already merged PR is checked against remote main.

The definition is versioned in the repository and can improve alongside
production. It is not another Oracle compiler or an autonomous guarantee of
correctness. `scope_finished` means the requested number of units finished;
`partial`, retained blockers and failures remain visible, and neither result
means the entire corpus is done.

References: [Dynamic Workflows](https://docs.github.com/en/copilot/concepts/agent-capabilities/agent-powered-workflows-and-operations/dynamic-workflows),
[using and monitoring workflows](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/use-dynamic-workflows).
