import test from "node:test";
import assert from "node:assert/strict";
import { readFile, mkdtemp, writeFile, mkdir, rm, chmod, rename } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { preexistingRule, tryClaim, claimFailure, defaultCorpusPath, executeUnit, carryGaps, retainedUnit, checkArgs, checkPolicy, checkSelection, assertCardBoundary, cardBoundaryOutcome, admissions, runProduction, validationPackagesForTask } from "./workflow.mjs";
import { git, changes, land, checkProtection, checkCoordinator, seedFromRetained } from "./git.mjs";
import { assessChecks, pullRequestAPI } from "./pull-request.mjs";
import { prepareCardBatch, verifyCardBatch } from "./queue.mjs";
import { chooseWork, knownGaps, gapTask, batchTask, issueNumber, engineIssue } from "./selection.mjs";
import { gapEvidence, earlierUnits } from "./workflow.mjs";
import { observeUsage, readReceipts, summarize, writeReport, reviewHistory } from "./telemetry.mjs";

const policy = JSON.parse(await readFile(new URL("./policy.json", import.meta.url)));
const complete = { status: "complete", reviewRequired: false, evidence: ["actual test"], gaps: [] };

test("Engine validation includes renderer type drift without adding a card gate or mutating policy", () => {
  const original = [...policy.validationPackages];
  const renderer = "./cardgen/cmd/genrender";
  assert.deepEqual(validationPackagesForTask(policy, "card"), original);
  assert.deepEqual(validationPackagesForTask(policy, "engine"), [...original, renderer]);
  assert.deepEqual(policy.validationPackages, original);
  assert.deepEqual(validationPackagesForTask({ ...policy, validationPackages: [...original, renderer] }, "engine"),
    [...original, renderer]);
});

function fixture(overrides = {}) {
  const calls = [];
  const setup = {
    task: { taskType: "engine", id: "3312", reviewRequired: false },
    policy, calls,
    stage: async phase => { calls.push(phase); return complete; },
    validate: async () => { calls.push("validate"); return { ok: true }; },
    snapshot: async () => ({ digest: "fixed", files: ["mtg/rules/scry.go"] }),
    review: async round => { calls.push(`review-${round}`); return { verdict: "approve", findings: [] }; },
    publish: async () => { calls.push("prepare"); return { tree: "reviewed" }; },
    ...overrides,
  };
  return setup;
}

test("Engine review is mandatory; cards follow tunable risk policy", async () => {
  const engine = fixture();
  assert.equal((await executeUnit(engine)).status, "accepted");
  assert.deepEqual(engine.calls, ["implementation", "validate", "review-0", "prepare"]);
  const card = fixture({ task: { taskType: "card", reviewRequired: false } });
  assert.equal((await executeUnit(card)).status, "accepted");
  assert.deepEqual(card.calls, ["implementation", "validate", "prepare"]);
});

test("Review fixes are separate work and require validation and re-review", async () => {
  const f = fixture();
  f.review = async round => {
    f.calls.push(`review-${round}`);
    return { verdict: round ? "approve" : "changes_needed", findings: round ? [] : ["wrong controller"] };
  };
  assert.equal((await executeUnit(f)).status, "accepted");
  assert.deepEqual(f.calls, ["implementation", "validate", "review-0", "review_fix", "validate", "review-1", "prepare"]);
});

test("Resumed units repair final findings first and keep the review-fix budget", async () => {
  const calls = [], fixes = [], reviews = [];
  const earlier = [{ round: 0, verdict: "changes_needed", findings: ["first"] }];
  const f = fixture({
    stage: async (phase, round, feedback) => {
      calls.push(`${phase}-${round}`);
      fixes.push(feedback);
      return { ...complete, summary: `fix ${round}`, gaps: [] };
    },
    review: async (round, _result, _snapshot, history) => {
      calls.push(`review-${round}`);
      reviews.push(history);
      return { verdict: "changes_needed", findings: [`still ${round}`] };
    },
    resume: { history: earlier, verdict: { verdict: "changes_needed", findings: ["final"] },
      result: { ...complete, gaps: [{ id: "#3333", description: "cascade trigger" }] } },
  });
  f.calls = calls;
  f.validate = async () => { calls.push("validate"); return { ok: true }; };
  const outcome = await executeUnit(f);
  assert.equal(outcome.status, "review_rejected");
  assert.deepEqual(calls, ["review_fix-0", "validate", "review-0", "review_fix-1", "validate", "review-1"]);
  assert.deepEqual(fixes[0], { findings: ["final"], earlierRounds: earlier });
  assert.deepEqual(reviews[0], [...earlier, { round: 1, verdict: "changes_needed", findings: ["final"], response: "fix 0" }]);
  assert.deepEqual(outcome.result.gaps, [{ id: "#3333", description: "cascade trigger" }]);
});

test("Re-reviews and fixes see earlier rounds, and repairs keep earlier gaps", async () => {
  const reviews = [], fixes = [];
  const gap = (id, description = id) => ({ id, description });
  const f = fixture({
    stage: async (phase, round, feedback) => {
      if (phase === "implementation") return { ...complete, gaps: [gap("#1"), gap("#2")] };
      fixes.push(feedback);
      return round === 1
        ? { ...complete, summary: "fixed controller", gaps: [] }
        : { ...complete, summary: "fixed zone", gaps: [gap("#2", "narrower"), gap("#3")], resolvedGaps: ["#1"] };
    },
    review: async (round, _result, _snapshot, history) => {
      reviews.push(history);
      return round < 2 ? { verdict: "changes_needed", findings: [`finding ${round}`] } : { verdict: "approve", findings: [] };
    },
  });
  const outcome = await executeUnit(f);
  assert.equal(outcome.status, "accepted");
  assert.deepEqual(reviews, [
    [],
    [{ round: 0, verdict: "changes_needed", findings: ["finding 0"], preexisting: [], response: "fixed controller" }],
    [{ round: 0, verdict: "changes_needed", findings: ["finding 0"], preexisting: [], response: "fixed controller" },
      { round: 1, verdict: "changes_needed", findings: ["finding 1"], preexisting: [], response: "fixed zone" }],
  ]);
  assert.deepEqual(fixes, [
    { findings: ["finding 0"], earlierRounds: [] },
    { findings: ["finding 1"], earlierRounds: [{ round: 0, verdict: "changes_needed", findings: ["finding 0"], preexisting: [], response: "fixed controller" }] },
  ]);
  assert.deepEqual(outcome.result.gaps, [gap("#2", "narrower"), gap("#3")]);
  assert.deepEqual(carryGaps({ gaps: [gap("#1")] }, { status: "blocked", gaps: [] }).gaps, [gap("#1")]);
  assert.equal(carryGaps({ gaps: [gap("#1")] }, null), null);
});

test("The corpus defaults to the shared pinned cache location", () => {
  const sha = "6da9ced0";
  assert.equal(defaultCorpusPath(sha, {}, "darwin", "/Users/me"), "/Users/me/Library/Caches/council4/pinned/6da9ced0.json");
  assert.equal(defaultCorpusPath(sha, {}, "linux", "/home/me"), "/home/me/.cache/council4/pinned/6da9ced0.json");
  assert.equal(defaultCorpusPath(sha, { XDG_CACHE_HOME: "/xdg" }, "linux", "/home/me"), "/xdg/council4/pinned/6da9ced0.json");
  assert.equal(defaultCorpusPath(sha, { COUNCIL4_ORACLE_CARDS: "/mine.json" }, "darwin", "/Users/me"), "/mine.json");
  assert.throws(() => defaultCorpusPath(sha, {}, "win32", "C:/Users/me"), /user cache directory/);
});

test("Every agent's model settings must be explicit", () => {
  assert.doesNotThrow(() => checkPolicy(policy));
  assert.equal(policy.models.engine.implementation.model, "claude-opus-5.5");
  const broken = mutate => { const copy = structuredClone(policy); mutate(copy.models); return copy; };
  for (const copy of [
    broken(models => { models.selection = null; }),
    broken(models => { models.engine.review = null; }),
    broken(models => { delete models.card.ci_fix.reasoningEffort; }),
    broken(models => { models.engine.implementation.contextTier = "huge"; }),
    broken(models => { models.card.implementation = "gpt-6-luna"; }),
  ]) {
    assert.throws(() => checkPolicy(copy), /model settings/);
  }
});

test("Engine work gets two review-fix rounds; card work gets one", async () => {
  const twice = round => ({ verdict: round < 2 ? "changes_needed" : "approve", findings: round < 2 ? [`finding ${round}`] : [] });
  const engine = fixture();
  engine.review = async round => { engine.calls.push(`review-${round}`); return twice(round); };
  assert.equal((await executeUnit(engine)).status, "accepted");
  assert.deepEqual(engine.calls.filter(call => call === "review_fix").length, 2);
  const card = fixture({ task: { taskType: "card", id: "batch", reviewRequired: true } });
  card.review = async round => { card.calls.push(`review-${round}`); return twice(round); };
  card.publish = async () => assert.fail("card exceeded its review-fix rounds");
  assert.notEqual((await executeUnit(card)).status, "accepted");
  assert.deepEqual(card.calls.filter(call => call.startsWith("review-")), ["review-0", "review-1"]);
  for (const reviewFixRounds of [1, { card: 1 }, { card: 1, engine: -1 }]) {
    assert.throws(() => checkPolicy({ ...policy, reviewFixRounds }), /reviewFixRounds/);
  }
});

test("Validation failures, missing agent output and unresolved reviews never land", async () => {
  for (const f of [
    fixture({ stage: async () => null }),
    fixture({ stage: async () => ({ status: "blocked", gaps: [{ id: "x", description: "missing Engine operation" }] }) }),
    fixture({ validate: async () => ({ ok: false, error: "actual test failed" }) }),
    fixture({ review: async () => null }),
    fixture({ review: async () => ({ verdict: "changes_needed", findings: ["still wrong"] }) }),
    fixture({ review: async () => ({ verdict: "approve", findings: ["unresolved material defect"] }) }),
  ]) {
    f.publish = async () => assert.fail("unaccepted work reached publication");
    assert.notEqual((await executeUnit(f)).status, "accepted");
  }
});

test("Reviewer cannot approve a changed submission", async () => {
  let snapshots = 0;
  const f = fixture({ snapshot: async () => ({ digest: String(snapshots++), files: ["x"] }) });
  await assert.rejects(executeUnit(f), /Reviewer modified/);
});

test("Inputs fail closed and card tasks cannot hide Engine work", () => {
  checkPolicy(policy);
  checkArgs({ mode: "produce", maxTasks: 2, corpusPath: "/verified/corpus.json" });
  checkArgs({ mode: "produce", maxTasks: 2 });
  for (const args of [{}, { mode: "produce" }, { mode: "produce", maxTasks: 0, corpusPath: "x" }, { mode: "produce", maxTasks: 1, corpusPath: "" }]) {
    assert.throws(() => checkArgs(args));
  }
  assert.throws(() => checkSelection({ taskType: "card", id: "../bad", title: "bad", instructions: "bad" }));
  assert.throws(() => checkPolicy({ ...policy, validationPackages: ["--help"] }));
  assert.throws(() => assertCardBoundary({ taskType: "card" }, ["mtg/rules/effects.go"]));
  assert.doesNotThrow(() => assertCardBoundary({ taskType: "card" }, ["mtg/rules/my_card_test.go", "mtg/cards/authored/my_card.go"]));
  assert.doesNotThrow(() => assertCardBoundary({ taskType: "card" },
    ["mtg/rules/testdata/my-card/rulings.json", "cardgen/testdata/source.txt"]));
  assert.throws(() => assertCardBoundary({ taskType: "card" }, ["mtg/rules/testdata.go"]));
});

test("A card unit's Engine changes block that unit without aborting the run", async () => {
  const task = { taskType: "card", id: "batch" };
  const answer = { ...complete, gaps: [{ id: "#3285", description: "missing operation" }] };
  assert.equal(cardBoundaryOutcome(task, ["mtg/cards/authored/my_card.go", "mtg/rules/testdata/x.json"], answer), answer);
  const blocked = cardBoundaryOutcome(task, ["mtg/rules/effects.go", "mtg/rules/my_card_test.go"], answer);
  assert.equal(blocked.status, "blocked");
  assert.deepEqual(blocked.boundaryViolation, ["mtg/rules/effects.go"]);
  assert.deepEqual(blocked.gaps, answer.gaps);
  assert.equal(cardBoundaryOutcome({ taskType: "engine" }, ["mtg/rules/effects.go"], answer), answer);
  const f = fixture({ task, stage: async () => cardBoundaryOutcome(task, ["mtg/game/state.go"], complete) });
  f.publish = async () => assert.fail("boundary violation reached publication");
  const outcome = await executeUnit(f);
  assert.equal(outcome.status, "blocked");
  assert.deepEqual(outcome.result.boundaryViolation, ["mtg/game/state.go"]);
});

test("A rejected batch plan fails its unit, and a second rejection in a row stops the run", async () => {
  const rejected = await tryClaim(async () => { throw new Error("invalid batch x: missing family"); });
  assert.deepEqual(rejected, { error: "invalid batch x: missing family" });
  assert.deepEqual(await tryClaim(async () => ({ batchId: "b" })), { claim: { batchId: "b" } });
  const controller = new AbortController();
  controller.abort();
  await assert.rejects(tryClaim(async () => { throw new Error("cancelled"); }, controller.signal), /cancelled/);
  const first = claimFailure({ index: 1 }, rejected.error, [{ status: "landed" }]);
  assert.equal(first.status, "claim_failed");
  assert.equal(first.stop, undefined);
  assert.equal(claimFailure({ index: 2 }, "again", [first]).stop, true);
});

test("Actual usage receipts deduplicate, survive reload, exclude unrelated agents and preserve model attribution", async t => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "council4-telemetry-"));
  t.after(() => rm(directory, { recursive: true }));
  const file = path.join(directory, "usage.jsonl");
  let handler;
  const session = { on(fn) { handler = fn; return () => {}; } };
  const errors = [];
  observeUsage(session, "run", file, e => errors.push(e));
  handler({ type: "subagent.started", agentId: "a", data: { workflowRunId: "run" } });
  handler({ type: "subagent.started", agentId: "child", data: { parentId: "a" } });
  const usage = {
    type: "assistant.usage", id: "e1", agentId: "a", timestamp: "now",
    data: { accounting: { sourceSessionId: "s", sequence: 1 }, model: "actual-luna",
      inputTokens: 100, outputTokens: 10, copilotUsage: { totalNanoAiu: 500_000_000 } },
  };
  handler(usage);
  handler(usage);
  handler({ ...usage, agentId: "unrelated", id: "other", data: { ...usage.data, accounting: { sourceSessionId: "o", sequence: 1 } } });
  observeUsage(session, "run", file, e => errors.push(e));
  handler(usage);
  handler({ ...usage, agentId: "child", data: { ...usage.data, accounting: { sourceSessionId: "child", sequence: 1 }, copilotUsage: { totalNanoAiu: 250_000_000 } } });
  const stages = [{ label: "engine/gap/review_fix/1", taskType: "engine", taskId: "gap", phase: "review_fix" }];
  const detail = { runId: "run", consumed: { nanoAiu: 750_000_000, activeMs: 123 },
    agents: [{ agentId: "a", label: stages[0].label, resolvedModel: "actual-luna", activeMs: 120 }] };
  const report = summarize(detail, stages, readReceipts(file));
  assert.equal(errors.length, 0);
  assert.equal(report.groups.length, 1);
  assert.equal(report.groups[0].observedAic, 0.75);
  assert.equal(report.groups[0].activeMs, 120);
  assert.equal(report.groups[0].phase, "review_fix");
  assert.equal(report.reconciliation, "matched");
  const late = summarize({ ...detail, consumed: { nanoAiu: 1e9, activeMs: 150 } }, stages, readReceipts(file));
  assert.equal(late.unallocatedNanoAiu, 250_000_000);
  assert.equal(late.reconciliation, "incomplete");
  assert.equal(late.groups[0].complete, false);
});

test("Unavailable billing stays incomplete rather than becoming zero-cost success", () => {
  const detail = { runId: "r", consumed: { nanoAiu: 0, activeMs: 20 }, agents: [] };
  const result = summarize(detail, [], [{ kind: "usage", key: "a", agentId: "unknown", nanoAiu: null }]);
  assert.equal(result.missingBillingReceipts, 1);
  assert.equal(result.reconciliation, "incomplete");
});

test("Repeating the same card family in different units retains separate cost/time rows", () => {
  const stages = [0, 2].map(unitIndex => ({
    label: `card/batch/unit-${unitIndex}/implementation/0`, taskType: "card",
    taskId: "batch", phase: "implementation", unitIndex,
  }));
  const agents = stages.map((stage, index) => ({
    agentId: `agent-${index}`, label: stage.label, resolvedModel: "luna", activeMs: index + 10,
  }));
  const receipts = agents.map((agent, index) => ({
    kind: "usage", key: String(index), agentId: agent.agentId, model: "luna", nanoAiu: 1e9,
  }));
  const report = summarize({ runId: "r", consumed: { nanoAiu: 2e9, activeMs: 30 }, agents }, stages, receipts);
  assert.deepEqual(report.groups.map(group => group.unitIndex), [0, 2]);
  assert.deepEqual(report.groups.map(group => group.observedAic), [1, 1]);
  assert.deepEqual(report.groups.map(group => group.activeMs), [10, 11]);
  assert.equal(report.reconciliation, "matched");
});

test("Admission accounting separates new cards from migration and refuses unfinished claims", () => {
  const before = { records: [
    { oracle_id: "new", name: "New", status: "unclaimed" },
    { oracle_id: "old", name: "Old", status: "installed-compiled" },
  ] };
  const after = { records: before.records.map(r => ({ ...r, status: "implemented-verified" })) };
  const result = admissions(before, after, "card");
  assert.deepEqual(result.newlySupported, [{ oracleId: "new", name: "New" }]);
  assert.deepEqual(result.existingSupportVerified, [{ oracleId: "old", name: "Old" }]);
  assert.throws(() => admissions(before, after, "engine"), /mixed in card admissions/);
  assert.throws(() => admissions(before, { records: [{ status: "claimed", owner: "writer" }] }, "card"), /active queue claim/);
});

async function queueFixture(t) {
  const workPath = await mkdtemp(path.join(os.tmpdir(), "council4-claim-"));
  t.after(() => rm(workPath, { recursive: true }));
  const manifest = path.join(workPath, "testdata/card-authoring/production.json");
  await mkdir(path.dirname(manifest), { recursive: true });
  const queue = {
    corpus_sha256: "pin",
    batches: [{ id: "batch", oracle_ids: ["a", "b", "c"] }],
    records: ["a", "b", "c"].map(id => ({
      oracle_id: id, revision: `printing-${id}`, name: id, batch: "batch", status: "unclaimed",
    })),
  };
  await writeFile(manifest, JSON.stringify(queue));
  const calls = [];
  const command = async (_cwd, _executable, args) => {
    const operation = args[args.indexOf("-operation") + 1];
    calls.push(operation);
    const current = JSON.parse(await readFile(manifest, "utf8"));
    if (operation === "next") {
      if (current.records.some(card => card.owner && card.owner !== "run")) throw new Error("Another owner holds the claim");
      return JSON.stringify({ batch: current.batches[0], cards: current.records });
    }
    if (operation === "claim") {
      current.records.forEach(card => { card.status = "claimed"; card.owner = "run"; });
    } else if (operation === "verify") {
      current.records.forEach(card => { card.status = "implemented-verified"; card.owner = ""; card.evidence = "actual named scenario"; });
    } else if (operation !== "check") throw new Error("Unexpected queue operation");
    await writeFile(manifest, JSON.stringify(current));
    return "";
  };
  const options = { workPath, task: { id: "batch", batchPlan: "" }, corpus: "/corpus",
    owner: "run", planPath: path.join(workPath, ".cardwork/plan.json"), command };
  return { options, manifest, calls };
}

test("Coordinator claims the saved batch before card work and owns verified admission", async t => {
  const f = await queueFixture(t);
  const claim = await prepareCardBatch(f.options);
  assert.deepEqual(f.calls, ["next", "claim", "next"]);
  assert.equal(claim.owner, "run");
  assert.equal(claim.cards.length, 3);
  assert.ok(JSON.parse(await readFile(f.manifest)).records.every(card => card.status === "claimed" && card.owner === "run"));
  await verifyCardBatch({ ...f.options, batchId: claim.batchId });
  assert.equal(f.calls.at(-1), "verify");
  await verifyCardBatch({ ...f.options, batchId: claim.batchId });
  assert.equal(f.calls.at(-1), "check");
});

test("Coordinator refuses foreign claims, invented batches, rewrites and partial worker promotion", async t => {
  for (const failure of ["foreign owner", "unknown batch", "saved rewrite", "wrong plan", "partial"]) {
    const f = await queueFixture(t);
    const queue = JSON.parse(await readFile(f.manifest, "utf8"));
    if (failure === "foreign owner") queue.records.forEach(card => { card.owner = "other"; card.status = "claimed"; });
    if (failure === "unknown batch") f.options.task.id = "unknown";
    if (failure === "saved rewrite") f.options.task.batchPlan = "{}";
    if (failure === "wrong plan") {
      f.options.task.id = "unknown";
      f.options.task.batchPlan = JSON.stringify({ corpus_sha256: "stale", batches: [{ id: "unknown" }] });
    }
    if (failure === "partial") {
      queue.records[0].status = "implemented-verified";
      queue.records[1].status = "claimed";
      queue.records[1].owner = "run";
    }
    await writeFile(f.manifest, JSON.stringify(queue));
    const before = await readFile(f.manifest, "utf8");
    if (failure === "partial") await assert.rejects(verifyCardBatch({ ...f.options, batchId: "batch" }), /partially promoted/);
    else await assert.rejects(prepareCardBatch(f.options));
    assert.equal(await readFile(f.manifest, "utf8"), before);
    assert.ok(!f.calls.includes("claim") && !f.calls.includes("verify"));
  }
});

const open = number => ({ number, state: "OPEN", title: `Gap ${number}`, url: `https://github.com/natefinch/council4/issues/${number}` });
const readyBatch = { id: "11-ready", family: "draw", operations: ["draw"], prerequisites: [12] };
const blockedCard = (issue, ids = ["x", "y", "z"]) => ({
  status: "blocked", task: { taskType: "card", id: "new-family" },
  claim: { cards: ids.map(oracleId => ({ oracleId })) },
  result: { status: "blocked", gaps: [{ id: `#${issue}`, description: "missing operation" }] },
});
const landedGap = (issue, id = `gap-${issue}`) => ({
  status: "landed", task: { taskType: "engine", id },
  landing: { commit: `commit-${issue}`, pr: { url: `https://github.com/natefinch/council4/pull/${issue}0` } },
});

test("Gap references parse canonical issue forms only", () => {
  for (const [value, issue] of [["#3285", 3285], ["3285", 3285], ["gap-3285", 3285], ["issue-12", 12],
    ["https://github.com/natefinch/council4/issues/77", 77], ["08-scry-interaction-claim", null], ["#12-x", null], ["", null]]) {
    assert.equal(issueNumber(value), issue, value);
  }
  assert.equal(engineIssue({ task: { taskType: "engine", id: "1906-simultaneous-choice" } }), 1906);
  assert.equal(engineIssue({ task: { taskType: "card", id: "1906-cards" } }), null);
  // Blocked results naming batch/claim problems are not schedulable Engine gaps.
  assert.deepEqual(knownGaps({ units: [{ ...blockedCard(1), result: { gaps: [{ id: "08-scry-interaction-claim" }] } }] }), []);
});

test("A landed Engine slice with a remainder keeps its gap scheduled; a complete one fills it", () => {
  const next = { batch: readyBatch, pending_prerequisites: [{ blocked_oracle_ids: ["q"], issues: [3285] }] };
  const issues = { 3285: open(3285), 3400: open(3400) };
  const slice = { ...landedGap(3285), task: { taskType: "engine", id: "gap-3285", issue: 3285, oracleIds: ["q"] },
    result: { status: "complete", gaps: [{ id: "#3285", description: "linked Aura transition" }] } };
  let choice = chooseWork({ next, units: [slice], issues });
  assert.equal(choice.kind, "engine");
  assert.equal(choice.gap.issue, 3285);
  assert.ok(choice.gap.sources.includes("this run remainder of gap-3285"));
  const subIssue = { ...slice, result: { status: "complete", gaps: [{ id: "#3400", description: "sub-issue" }] } };
  choice = chooseWork({ next, units: [subIssue], issues });
  assert.equal(choice.gap.issue, 3400);
  assert.deepEqual(choice.gap.oracleIds, ["q"]);
  assert.equal(chooseWork({ next, units: [landedGap(3285)], issues }).kind, "card");
  assert.match(gapTask({ gap: { issue: 3285, oracleIds: [], sources: [] }, issue: open(3285) }, {}).instructions,
    /first shared piece[\s\S]*status complete/);
});

test("Pre-existing defects are filed, carried forward and never block approval", async () => {
  const filed = [{ issue: 4100, description: "face-down graveyard casts ignore granted abilities", evidence: "path unchanged by diff" }];
  const seen = [];
  const f = fixture({
    review: async (round, _result, _snapshot, history) => {
      seen.push(history);
      return round === 0
        ? { verdict: "changes_needed", findings: ["introduced regression"], preexisting: filed }
        : { verdict: "approve", findings: [], preexisting: [] };
    },
  });
  assert.equal((await executeUnit(f)).status, "accepted");
  assert.deepEqual(seen[1][0].preexisting, filed);
  const onlyPreexisting = fixture({ review: async () => ({ verdict: "approve", findings: [], preexisting: filed }) });
  assert.equal((await executeUnit(onlyPreexisting)).status, "accepted");
  assert.match(preexistingRule("abc123"), /base commit abc123[\s\S]*never block approval/);
});

test("A split gap is worked through its first open sub-issue, in order", () => {
  const next = { batch: readyBatch, pending_prerequisites: [{ blocked_oracle_ids: ["q"], issues: [3285] }] };
  const subs = states => states.map(([number, state]) => ({ number, state, title: `Part ${number}`, url: `u/${number}` }));
  const issues = { 3285: { ...open(3285), subIssues: subs([[4001, "CLOSED"], [4002, "OPEN"], [4003, "OPEN"]]) } };
  const failedParent = { status: "review_rejected", task: { taskType: "engine", id: "gap-3285", issue: 3285 } };
  let choice = chooseWork({ next, priorUnits: [failedParent], issues });
  assert.equal(choice.kind, "engine");
  assert.equal(choice.gap.issue, 4002);
  assert.equal(choice.gap.parent, 3285);
  assert.deepEqual(choice.gap.oracleIds, ["q"]);
  const task = gapTask(choice, { q: "Necromancy" });
  assert.equal(task.id, "gap-4002");
  assert.match(task.title, /#4002: Part 4002/);
  assert.match(task.instructions, /sub-issue of #3285/);
  // A landed sub-issue moves on to the next; a failed one parks the parent.
  choice = chooseWork({ next, priorUnits: [failedParent, landedGap(4002)], issues });
  assert.equal(choice.gap.issue, 4003);
  const failedSub = { status: "review_rejected", task: { taskType: "engine", id: "gap-4002", issue: 4002 } };
  assert.equal(chooseWork({ next, priorUnits: [failedParent, failedSub], issues }).kind, "card");
  // A failed Engine unit whose gap came from neither a card nor the queue is
  // still worked through its sub-issues once split.
  const selectorGap = { status: "review_rejected", task: { taskType: "engine", id: "gap-170" } };
  const adventure = { 170: { ...open(170), subIssues: subs([[3353, "OPEN"], [3354, "OPEN"]]) } };
  choice = chooseWork({ next: { batch: readyBatch }, priorUnits: [selectorGap], issues: adventure });
  assert.equal(choice.gap.issue, 3353);
  assert.equal(chooseWork({ next: { batch: readyBatch }, priorUnits: [selectorGap], issues: { 170: open(170) } }).kind, "card");
  // Without sub-issues a failed parent stays parked.
  assert.equal(chooseWork({ next, priorUnits: [failedParent], issues: { 3285: open(3285) } }).kind, "card");
});

test("Known open Engine gaps are selected before ready card batches, demonstrated gaps first", () => {
  const next = { batch: readyBatch, cards: [{ name: "A" }], pending_prerequisites: [{ blocked_oracle_ids: ["q"], issues: [3285] }] };
  const issues = { 3285: open(3285), 2161: open(2161) };
  let choice = chooseWork({ next, units: [blockedCard(2161)], issues });
  assert.equal(choice.kind, "engine");
  assert.equal(choice.gap.issue, 2161);
  assert.deepEqual(choice.gap.oracleIds, ["x", "y", "z"]);
  choice = chooseWork({ next, issues });
  assert.equal(choice.kind, "engine");
  assert.equal(choice.gap.issue, 3285);
  const task = gapTask(choice, { q: "Necromancy" });
  assert.equal(task.id, "gap-3285");
  assert.equal(task.taskType, "engine");
  assert.equal(task.issue, 3285);
  assert.match(task.instructions, /Necromancy \(q\)/);
  assert.doesNotThrow(() => checkSelection(task));
});

test("Ready saved batches are selected deterministically when no open gap remains", () => {
  const next = { batch: readyBatch, cards: [{ name: "A", oracle_id: "a", revision: "r" }],
    pending_prerequisites: [{ issues: [3285] }, { issues: [999] }] };
  // A filled gap, a closed issue and an unreadable reference never outrank ready work.
  const choice = chooseWork({ next, units: [landedGap(3285)], issues: { 3285: open(3285), 999: { state: "CLOSED" } } });
  assert.equal(choice.kind, "card");
  const task = batchTask(choice);
  assert.equal(task.id, "11-ready");
  assert.equal(task.batchPlan, "");
  assert.doesNotThrow(() => checkSelection(task));
});

test("A failed gap attempt is parked for this run and batch preparation avoids open gaps", () => {
  const failed = { status: "blocked", task: { taskType: "engine", id: "gap-3285", issue: 3285 } };
  const next = { batch: null, pending_prerequisites: [{ blocked_oracle_ids: ["q"], issues: [3285] }, { blocked_oracle_ids: ["n"], issues: [500] }] };
  const choice = chooseWork({ next, units: [failed], priorUnits: [landedGap(500)], issues: { 3285: open(3285), 500: open(500) } });
  assert.equal(choice.kind, "prepare");
  assert.deepEqual(choice.parked.map(gap => gap.issue), [3285]);
  assert.deepEqual(choice.avoid.map(gap => gap.issue), [3285]);
  assert.deepEqual(choice.unblocked.map(gap => gap.issue), [500]);
  // Earlier-run failures stay parked unless explicitly retried.
  const earlier = { next: { batch: readyBatch, pending_prerequisites: [{ issues: [3285] }] }, priorUnits: [failed], issues: { 3285: open(3285) } };
  assert.equal(chooseWork(earlier).kind, "card");
  assert.equal(chooseWork({ ...earlier, retry: [3285] }).gap.issue, 3285);
  assert.deepEqual(gapEvidence(choice.unblocked, [landedGap(500)], {}, "base"), {
    500: { runtime_revision: "commit-500", evidence: "Engine unit gap-500 landed in https://github.com/natefinch/council4/pull/5000" },
  });
  assert.deepEqual(gapEvidence([{ issue: 7, oracleIds: ["c"] }], [], { 7: { state: "CLOSED" } }, "base"), {
    7: { runtime_revision: "base", evidence: "Issue #7 closed; reassessed on runtime base" },
  });
});

test("Earlier local runs carry blocked-card gaps and landed gap fixes forward", async t => {
  const root = await mkdtemp(path.join(os.tmpdir(), "council4-history-"));
  t.after(() => rm(root, { recursive: true }));
  for (const [run, units] of [["old", [blockedCard(41)]], ["fix", [landedGap(42)]], ["current", [blockedCard(43)]]]) {
    await mkdir(path.join(root, ".cardwork/workflows", run), { recursive: true });
    await writeFile(path.join(root, ".cardwork/workflows", run, "state.json"), JSON.stringify({ units }));
  }
  await mkdir(path.join(root, ".cardwork/workflows/empty"));
  const priorUnits = await earlierUnits(root, "current");
  assert.equal(priorUnits.length, 2);
  const gaps = knownGaps({ next: { pending_prerequisites: [{ issues: [42], blocked_oracle_ids: ["w"] }] }, priorUnits });
  assert.deepEqual(gaps.map(gap => gap.issue), [41, 42]);
  const choice = chooseWork({ next: { batch: readyBatch, pending_prerequisites: [{ issues: [42], blocked_oracle_ids: ["w"] }] },
    priorUnits, issues: { 41: open(41), 42: open(42) } });
  assert.equal(choice.gap.issue, 41);
  assert.deepEqual(await earlierUnits(path.join(root, "missing"), "x"), []);
});

test("Coordinator reassesses only filled-gap blocked plan members before batching", async t => {
  const workPath = await mkdtemp(path.join(os.tmpdir(), "council4-reassess-"));
  t.after(() => rm(workPath, { recursive: true }));
  const manifest = path.join(workPath, "testdata/card-authoring/production.json");
  await mkdir(path.dirname(manifest), { recursive: true });
  const cards = ["a", "b", "c"].map(id => ({ oracle_id: id, revision: `r-${id}`, name: id, status: "unclaimed" }));
  cards[0] = { ...cards[0], status: "blocked", blocker: 3285, reason: "missing" };
  await writeFile(manifest, JSON.stringify({ corpus_sha256: "pin", batches: [], records: cards }));
  const plan = { corpus_sha256: "pin", batches: [{ id: "new" }], records: cards.map(({ oracle_id, revision, name }) => ({ oracle_id, revision, name })) };
  for (const evidence of [{}, { 3285: { runtime_revision: "commit", evidence: "landed" } }]) {
    const calls = [];
    const command = async (_cwd, _executable, args) => {
      const operation = args[args.indexOf("-operation") + 1];
      calls.push(operation);
      if (operation === "reassess") {
        const reassess = JSON.parse(await readFile(args[args.indexOf("-plan") + 1], "utf8"));
        assert.deepEqual(reassess.records, [{ oracle_id: "a", revision: "r-a", name: "a",
          reassessment: { issue: 3285, runtime_revision: "commit", evidence: "landed" } }]);
      }
      if (operation === "next") return JSON.stringify({ batch: { id: "new" }, cards: cards.map(card => ({ ...card, status: "claimed", owner: "run" })) });
      return "";
    };
    const options = { workPath, task: { id: "new", batchPlan: JSON.stringify(plan) }, corpus: "/c", owner: "run",
      planPath: path.join(workPath, ".cardwork/plan.json"), command, prerequisiteEvidence: evidence };
    if (!evidence[3285]) {
      await assert.rejects(prepareCardBatch(options), /still blocked by #3285/);
      assert.deepEqual(calls, []);
    } else {
      await prepareCardBatch(options);
      assert.deepEqual(calls, ["reassess", "add-reviewed-batch", "next", "claim", "next"]);
    }
  }
});

async function repositoryFixture(t) {
  const directory = await mkdtemp(path.join(os.tmpdir(), "council4-landing-"));
  t.after(() => rm(directory, { recursive: true }));
  const remote = path.join(directory, "remote.git");
  const repo = path.join(directory, "repo");
  await mkdir(repo);
  await git(directory, ["init", "--bare", remote]);
  await git(repo, ["init", "-b", "main"]);
  await git(repo, ["config", "user.name", "Workflow tests"]);
  await git(repo, ["config", "user.email", "workflow-tests@example.invalid"]);
  await writeFile(path.join(repo, ".gitignore"), ".cardwork/\n");
  await writeFile(path.join(repo, "card.txt"), "baseline\n");
  await git(repo, ["add", "card.txt", ".gitignore"]);
  await git(repo, ["commit", "-m", "baseline"]);
  await git(repo, ["remote", "add", "origin", remote]);
  await git(repo, ["push", "-u", "origin", "main"]);
  const base = await git(repo, ["rev-parse", "HEAD"]);
  const work = { path: path.join(directory, "task"), base, branch: "task" };
  await git(repo, ["worktree", "add", "-b", "task", work.path, base]);
  await writeFile(path.join(work.path, "card.txt"), "implemented\n");
  const expected = await changes(work.path);
  await git(work.path, ["add", "card.txt"]);
  expected.tree = await git(work.path, ["write-tree"]);
  return { repo, remote, base, work, expected };
}

const remoteHead = f => git(f.repo, ["ls-remote", "origin", "refs/heads/main"]).then(out => out.split(/\s+/)[0]);

test("Resume arguments, retained units and seeding fail closed and preserve work", async t => {
  checkArgs({ mode: "produce", maxTasks: 2, resumeUnits: [{ runId: "run-1", index: 0 }, { runId: "run-1", index: 3 }] });
  for (const resumeUnits of [[], [{ runId: "../x", index: 0 }], [{ runId: "r", index: -1 }],
    [{ runId: "r", index: 0 }, { runId: "r", index: 0 }], [{ runId: "r", index: 0 }, { runId: "r", index: 1 }, { runId: "r", index: 2 }]]) {
    assert.throws(() => checkArgs({ mode: "produce", maxTasks: 2, resumeUnits }));
  }

  const f = await repositoryFixture(t);
  const runId = "old-run";
  const run = path.join(f.repo, ".cardwork", "workflows", runId);
  const retained = path.join(run, "work", "0");
  await git(f.repo, ["worktree", "add", "-b", "retained", retained, f.base]);
  await writeFile(path.join(retained, "card.txt"), "retained fix\n");
  await writeFile(path.join(retained, "new_test.go"), "package rules\n");
  const unit = { index: 0, task: { taskType: "engine", id: "gap-3287", issue: 3287, unitIndex: 0 }, status: "review_rejected",
    worktree: retained, result: { ...complete, gaps: [] }, verdict: { verdict: "changes_needed", findings: ["final"] } };
  const stages = [
    { unitIndex: 0, phase: "review", verdict: "changes_needed", findings: ["first"] },
    { unitIndex: 0, phase: "re_review", verdict: "changes_needed", findings: ["final"] },
  ];
  const writeState = units => writeFile(path.join(run, "state.json"), JSON.stringify({ units, stages }));
  await writeState([unit, { ...unit, index: 1, task: { taskType: "card", id: "b" } }, { ...unit, index: 2, status: "landed" },
    { ...unit, index: 3, worktree: f.work.path }]);
  const loaded = await retainedUnit(f.repo, { runId, index: 0 });
  assert.deepEqual(loaded.history, [{ round: 0, verdict: "changes_needed", findings: ["first"], preexisting: [] }]);
  for (const index of [1, 2, 3, 4]) await assert.rejects(retainedUnit(f.repo, { runId, index }));
  // Another top-level worktree of the repository (a new session) finds the run.
  assert.equal((await retainedUnit(f.work.path, { runId, index: 0 })).unit.worktree, retained);
  assert.equal((await earlierUnits(f.work.path, "current")).length, 4);
  await assert.rejects(retainedUnit(f.repo, { runId: "missing", index: 0 }), /No retained run/);

  // Main advanced independently; the resumed worktree gets both.
  await writeFile(path.join(f.repo, "other.txt"), "main moved\n");
  await git(f.repo, ["add", "other.txt"]);
  await git(f.repo, ["commit", "-m", "main moved"]);
  const head = await git(f.repo, ["rev-parse", "HEAD"]);
  const target = { path: path.join(path.dirname(f.repo), "resumed"), base: head };
  await git(f.repo, ["worktree", "add", "-b", "resumed", target.path, head]);
  const sourceStatus = await git(retained, ["status", "--porcelain"]);
  const seeded = await seedFromRetained(retained, target);
  assert.equal(seeded.retainedBase, f.base);
  assert.equal(await git(target.path, ["rev-parse", "HEAD"]), head);
  assert.deepEqual((await changes(target.path)).files, ["card.txt", "new_test.go"]);
  assert.equal(await readFile(path.join(target.path, "card.txt"), "utf8"), "retained fix\n");
  assert.equal(await readFile(path.join(target.path, "other.txt"), "utf8"), "main moved\n");
  assert.equal(await git(target.path, ["diff", "--cached", "--name-only"]), "");
  assert.equal(await git(retained, ["status", "--porcelain"]), sourceStatus);
  await assert.rejects(seedFromRetained(retained, target), /not clean/);

  // Conflicting main changes stop instead of overwriting retained work.
  const clash = { path: path.join(path.dirname(f.repo), "clash"), base: null };
  await writeFile(path.join(f.repo, "card.txt"), "main changed card\n");
  await git(f.repo, ["commit", "-am", "conflict"]);
  clash.base = await git(f.repo, ["rev-parse", "HEAD"]);
  await git(f.repo, ["worktree", "add", "-b", "clash", clash.path, clash.base]);
  await assert.rejects(seedFromRetained(retained, clash), /conflicts with current main/);
  assert.deepEqual((await changes(clash.path)).files, []);
});
const required = { checks: ["Tests", "Race Tests"], reviewRequired: false };
const successfulChecks = head => ({
  head, statuses: [],
  runs: ["Tests", "Race Tests"].map((name, index) => ({
    id: index + 1, name, status: "completed", conclusion: "success", app: { slug: "github-actions" },
  })),
});

function landingServices(f) {
  const calls = { created: 0, merged: 0, validation: 0, review: 0, commands: [] };
  let pr = null;
  const services = {
    protection: async () => required,
    validate: async () => { calls.validation++; },
    reviewIntegration: async () => { calls.review++; },
    api: {
      find: async () => pr,
      create: async () => {
        calls.created++;
        pr = { number: 1, url: "https://github.com/test/repo/pull/1", state: "OPEN", isDraft: false,
          baseRefName: "main", mergeStateStatus: "CLEAN", reviewDecision: "" };
        return services.api.view();
      },
      view: async () => {
        if (!pr) throw new Error("PR does not exist");
        return { ...pr, headRefOid: await git(f.work.path, ["rev-parse", "HEAD"]),
          baseRefOid: await remoteHead(f) };
      },
      checks: async (_cwd, head) => successfulChecks(head),
      watch: async () => {},
      failedLogs: async (_cwd, checks) => ({ checks, logs: ["Actual failing check log"] }),
      merge: async (_cwd, _number, expectedHead) => {
        assert.equal(expectedHead, await git(f.work.path, ["rev-parse", "HEAD"]));
        calls.merged++;
        // Simulate GitHub's squash-merge server with a real temporary remote.
        await git(f.repo, ["fetch", "origin", f.work.branch]);
        await git(f.repo, ["merge", "--squash", expectedHead]);
        await git(f.repo, ["commit", "-m", "GitHub squash merge"]);
        await git(f.repo, ["push", "origin", "main"]);
        pr = { ...pr, state: "MERGED", mergeCommit: { oid: await remoteHead(f) } };
        return services.api.view();
      },
    },
    git: async (...args) => {
      calls.commands.push(args[1]);
      return git(...args);
    },
    calls,
  };
  return services;
}

test("CI watch waits for required check registration even when unrelated checks appear first", async () => {
  const waits = [];
  const commands = [];
  const api = pullRequestAPI(async (_cwd, executable, args) => {
    commands.push([executable, ...args]);
    return "";
  }, undefined, async ms => { waits.push(ms); });
  const head = "reviewed-head";
  api.view = async () => ({ headRefOid: head, state: "OPEN", baseRefName: "main", mergeStateStatus: "CLEAN" });
  const running = successfulChecks(head).runs.map(check => ({ ...check, status: "in_progress", conclusion: null }));
  const manual = { id: 3, name: "Manual", status: "completed", conclusion: "skipped" };
  const observations = [[], [manual], [manual, running[0]], running];
  let reads = 0;
  api.checks = async () => ({
    head, runs: observations[reads++],
    statuses: required.checks.map(context => ({ context, state: "success" })),
  });
  await api.watch("/repo", 42, head, required);
  assert.equal(reads, 4);
  assert.deepEqual(waits, [15000, 15000, 15000]);
  assert.deepEqual(commands, [["gh", "pr", "checks", "42", "--repo", "natefinch/council4",
    "--watch", "--fail-fast", "--interval", "15"]]);
});

test("CI registration wait is bounded and preserves missing checks as pending", async () => {
  let waits = 0;
  let reads = 0;
  const api = pullRequestAPI(async () => { assert.fail("Must not watch nonexistent checks"); },
    undefined, async () => { waits++; });
  const head = "reviewed-head";
  const pr = { headRefOid: head, state: "OPEN", baseRefName: "main", mergeStateStatus: "CLEAN" };
  const checks = { head, statuses: [], runs: [] };
  api.view = async () => pr;
  api.checks = async () => { reads++; return checks; };
  await api.watch("/repo", 42, head, required);
  assert.equal(reads, 9);
  assert.equal(waits, 8);
  assert.equal(assessChecks(pr, checks, required, head).status, "pending");
});

test("CI registration wait propagates stale heads, read errors and cancellation", async () => {
  for (const mode of ["head", "read", "cancel"]) {
    const head = "reviewed-head";
    const failure = new Error(mode);
    const controller = new AbortController();
    const api = pullRequestAPI(async () => { assert.fail("Must not start gh watch"); },
      controller.signal);
    api.view = async () => ({
      headRefOid: mode === "head" ? "changed-head" : head,
      state: "OPEN", baseRefName: "main", mergeStateStatus: "CLEAN",
    });
    api.checks = async () => {
      if (mode === "read") throw failure;
      if (mode === "cancel") controller.abort();
      return { head, statuses: [], runs: [] };
    };
    await assert.rejects(api.watch("/repo", 42, head, required),
      mode === "head" ? /expected open main-bound reviewed head/ :
        mode === "cancel" ? { name: "AbortError" } : failure);
  }
});

test("Landing watches initially absent CI and only merges after exact-head checks register", async t => {
  const f = await repositoryFixture(t);
  const services = landingServices(f);
  let registered = false;
  let watched = 0;
  services.api.checks = async (_cwd, head) =>
    registered ? successfulChecks(head) : { head, runs: [], statuses: [] };
  services.api.watch = async (_cwd, _number, head, checksRequired) => {
    assert.equal(head, await git(f.work.path, ["rev-parse", "HEAD"]));
    assert.deepEqual(checksRequired, required);
    watched++;
    registered = true;
  };
  const result = await land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services);
  assert.equal(watched, 1);
  assert.equal(result.published, true);
  assert.equal(services.calls.merged, 1);
});

test("Landing publishes only a PR branch and verifies CI before squash merge; replay is idempotent", async t => {
  const f = await repositoryFixture(t);
  const services = landingServices(f);
  const first = await land(f.repo, f.work, { title: "Implement card" }, f.expected, "session", undefined, services);
  assert.equal(first.published, true);
  assert.equal(await remoteHead(f), first.commit);
  const replay = await land(f.repo, f.work, { title: "Implement card" }, f.expected, "session", undefined, services);
  assert.equal(replay.commit, first.commit);
  assert.equal(services.calls.created, 1);
  assert.equal(services.calls.merged, 1);
  assert.equal(services.calls.validation, 1);
  const pushes = services.calls.commands.filter(command => command[0] === "push");
  assert.ok(pushes.length);
  assert.ok(pushes.every(command => command.at(-1) === "HEAD:refs/heads/task"));
  await checkCoordinator(f.work.path, first.head, first.main);
  await writeFile(path.join(f.work.path, "unpublished.txt"), "unlanded work\n");
  await git(f.work.path, ["add", "unpublished.txt"]);
  await git(f.work.path, ["commit", "-m", "unpublished"]);
  await assert.rejects(checkCoordinator(f.work.path, await git(f.work.path, ["rev-parse", "HEAD"]), first.main), /unlanded/);
});

test("Changed submission and unreviewed crash commits reject publication", async t => {
  const f = await repositoryFixture(t);
  const services = landingServices(f);
  await writeFile(path.join(f.work.path, "card.txt"), "different\n");
  await assert.rejects(land(f.repo, f.work, { title: "bad" }, f.expected, "session", undefined, services), /Reviewed content/);
  assert.equal(await remoteHead(f), f.base);
  await git(f.work.path, ["add", "card.txt"]);
  await git(f.work.path, ["commit", "-m", "unreviewed\n\nCopilot-Session: session"]);
  await assert.rejects(land(f.repo, f.work, { title: "bad" }, f.expected, "session", undefined, services), /Unexpected task/);
});

test("Review fingerprints do not change when new, renamed, binary and executable files are staged", async t => {
  const f = await repositoryFixture(t);
  await writeFile(path.join(f.work.path, "new.txt"), "new card source\n");
  await writeFile(path.join(f.work.path, "empty.txt"), "");
  await writeFile(path.join(f.work.path, "binary.dat"), Buffer.from([0, 255, 0, 128]));
  await writeFile(path.join(f.work.path, "tool.sh"), "#!/bin/sh\nexit 0\n");
  await chmod(path.join(f.work.path, "tool.sh"), 0o755);
  await rename(path.join(f.work.path, "card.txt"), path.join(f.work.path, "renamed.txt"));
  const before = await changes(f.work.path);
  await git(f.work.path, ["add", "--", ...before.files]);
  assert.deepEqual(await changes(f.work.path), before);
  await git(f.work.path, ["restore", "--staged", "--", ...before.files]);
  assert.deepEqual(await changes(f.work.path), before);
  await chmod(path.join(f.work.path, "tool.sh"), 0o644);
  assert.notEqual((await changes(f.work.path)).digest, before.digest);
  await chmod(path.join(f.work.path, "tool.sh"), 0o755);
  assert.deepEqual(await changes(f.work.path), before);
  await rm(path.join(f.work.path, "empty.txt"));
  assert.notEqual((await changes(f.work.path)).digest, before.digest);
  await writeFile(path.join(f.work.path, "empty.txt"), "");
  assert.deepEqual(await changes(f.work.path), before);
  await writeFile(path.join(f.work.path, "new.txt"), "different, unreviewed source\n");
  assert.notEqual((await changes(f.work.path)).digest, before.digest);
});

test("A prepared new-file tree lands without losing its review fingerprint", async t => {
  const f = await repositoryFixture(t);
  await writeFile(path.join(f.work.path, "new-card.txt"), "real new card\n");
  const expected = await changes(f.work.path);
  await git(f.work.path, ["add", "--", ...expected.files]);
  expected.tree = await git(f.work.path, ["write-tree"]);
  const services = landingServices(f);
  const result = await land(f.repo, f.work, { title: "new card" }, expected, "session", undefined, services);
  assert.equal(result.published, true);
});

test("New main is merged locally and changed integration is reviewed and validated before PR", async t => {
  const race = await repositoryFixture(t);
  const services = landingServices(race);
  await writeFile(path.join(race.repo, "other.txt"), "other writer\n");
  await git(race.repo, ["add", "other.txt"]);
  await git(race.repo, ["commit", "-m", "concurrent"]);
  await git(race.repo, ["push", "origin", "main"]);
  const main = await remoteHead(race);
  const result = await land(race.repo, race.work, { title: "card" }, race.expected, "session", undefined, services);
  assert.equal(result.published, true);
  assert.equal(services.calls.review, 1);
  assert.equal(services.calls.validation, 1);
  assert.equal(await git(race.repo, ["merge-base", result.commit, main]), main);
});

test("CI fixes change the PR head, rerun validation and require fresh successful CI", async t => {
  const f = await repositoryFixture(t);
  const services = landingServices(f);
  let initialHead;
  let fixed = false;
  services.api.checks = async (_cwd, head) => {
    initialHead ??= head;
    const checks = successfulChecks(head);
    if (head === initialHead) checks.runs[1].conclusion = "failure";
    return checks;
  };
  services.fixCI = async feedback => {
    assert.equal(feedback.head, initialHead);
    assert.ok(feedback.logs[0].includes("failing"));
    await writeFile(path.join(f.work.path, "card.txt"), "fixed actual CI problem\n");
    fixed = true;
  };
  const result = await land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services);
  assert.ok(fixed);
  assert.equal(result.ciFixes, 1);
  assert.notEqual(result.head, initialHead);
  assert.equal(services.calls.validation, 2);
  assert.equal(services.calls.created, 1);
});

test("Failed or missing CI, external review and nonmergeable PRs never merge", async t => {
  for (const mode of ["failure", "missing", "pending", "approval", "conflict", "stale"]) {
    const f = await repositoryFixture(t);
    const services = landingServices(f);
    if (mode === "failure") services.api.checks = async (_cwd, head) => {
      const checks = successfulChecks(head);
      checks.runs[1].conclusion = "failure";
      return checks;
    };
    if (mode === "missing") services.api.checks = async (_cwd, head) => ({ head, runs: [], statuses: [] });
    if (mode === "pending") services.api.checks = async (_cwd, head) => {
      const checks = successfulChecks(head);
      checks.runs[1].status = "in_progress";
      return checks;
    };
    const view = services.api.view;
    if (mode === "approval") services.api.view = async () => ({ ...await view(), reviewDecision: "CHANGES_REQUESTED" });
    if (mode === "conflict") services.api.view = async () => ({ ...await view(), mergeStateStatus: "DIRTY" });
    if (mode === "stale") services.api.checks = async () => successfulChecks("some-other-head");
    if (mode === "stale") {
      await assert.rejects(land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services), /stale commit/);
    } else {
      const result = await land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services);
      assert.equal(result.published, false);
    }
    assert.equal(services.calls.merged, 0);
    assert.equal(await remoteHead(f), f.base);
  }
});

test("Pending PR resume reuses its checkpoint without duplicate PR or validation", async t => {
  const f = await repositoryFixture(t);
  const services = landingServices(f);
  services.api.checks = async (_cwd, head) => ({ head, runs: [], statuses: [] });
  const first = await land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services);
  assert.equal(first.status, "pending");
  services.api.checks = async (_cwd, head) => successfulChecks(head);
  const resumed = await land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services);
  assert.equal(resumed.published, true);
  assert.equal(services.calls.created, 1);
  assert.equal(services.calls.validation, 1);
});

test("Interrupted merge and CI-fix commit resume only their exact journaled transition", async t => {
  for (const mode of ["merge", "ci_commit"]) {
    const f = await repositoryFixture(t);
    const services = landingServices(f);
    if (mode === "merge") {
      await writeFile(path.join(f.repo, "other.txt"), "concurrent main\n");
      await git(f.repo, ["add", "other.txt"]);
      await git(f.repo, ["commit", "-m", "main progressed"]);
      await git(f.repo, ["push", "origin", "main"]);
    } else {
      let firstHead;
      services.api.checks = async (_cwd, head) => {
        firstHead ??= head;
        const checks = successfulChecks(head);
        if (head === firstHead) checks.runs[1].conclusion = "failure";
        return checks;
      };
      services.fixCI = async () => writeFile(path.join(f.work.path, "card.txt"), "reviewed repair\n");
    }
    const original = services.git;
    let interrupted = false;
    services.git = async (...args) => {
      const result = await original(...args);
      if (!interrupted && args[0] === f.work.path &&
          (mode === "merge" && args[1][0] === "merge" ||
           mode === "ci_commit" && args[1][0] === "commit" && args[1][2].startsWith("Fix CI:"))) {
        interrupted = true;
        throw new Error("interrupted after Git succeeded");
      }
      return result;
    };
    await assert.rejects(land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services), /interrupted after/);
    const result = await land(f.repo, f.work, { title: "card" }, f.expected, "session", undefined, services);
    assert.equal(result.published, true);
    assert.equal(services.calls.created, 1);
    assert.equal(services.calls.merged, 1);
    if (mode === "merge") assert.equal(services.calls.review, 1);
  }
});

test("Exact-head checks require real Actions tests and reject skipped required checks", () => {
  const pr = { state: "OPEN", baseRefName: "main", headRefOid: "head", isDraft: false, mergeStateStatus: "CLEAN", reviewDecision: "" };
  assert.equal(assessChecks(pr, successfulChecks("head"), required, "head").status, "ready");
  const skipped = successfulChecks("head");
  skipped.runs[1].conclusion = "skipped";
  assert.equal(assessChecks(pr, skipped, required, "head").status, "failed");
  const forged = successfulChecks("head");
  forged.runs[1].app.slug = "not-actions";
  assert.throws(() => assessChecks(pr, forged, required, "head"), /not a GitHub Actions/);
});

test("Reports show every review round, falling back to a legacy final verdict", async t => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "council4-report-"));
  t.after(() => rm(directory, { recursive: true }));
  const stages = [
    { label: "engine/1/unit-0/implementation/0", unitIndex: 0, phase: "implementation", round: 0, outcome: "returned" },
    { label: "engine/1/unit-0/review/0", unitIndex: 0, phase: "review", round: 0, verdict: "changes_needed", findings: ["wrong | controller"] },
    { label: "engine/1/unit-0/re_review/1", unitIndex: 0, phase: "re_review", round: 1, verdict: "approve", findings: [] },
  ];
  const units = [
    { index: 0, task: { taskType: "engine", id: "gap-1" }, status: "landed" },
    { index: 1, task: { taskType: "engine", id: "gap-2" }, status: "review_rejected",
      verdict: { verdict: "changes_needed", findings: ["legacy finding"] } },
    { index: 2, task: { taskType: "card", id: "batch" }, status: "blocked" },
  ];
  assert.deepEqual(reviewHistory(units[0], stages).map(r => [r.phase, r.round, r.verdict]),
    [["review", 0, "changes_needed"], ["re_review", 1, "approve"]]);
  assert.deepEqual(reviewHistory(units[1], stages), [{ phase: "final review", round: null, verdict: "changes_needed", findings: ["legacy finding"] }]);
  assert.deepEqual(reviewHistory(units[2], stages), []);
  writeReport(directory, { runId: "r", policyVersion: "1.10.0", runtimeStatus: "completed", runtimeAic: 0,
    reconciliation: "matched", unallocatedNanoAiu: 0, groups: [], stages, units, note: "" });
  const markdown = await readFile(path.join(directory, "report.md"), "utf8");
  assert.match(markdown, /- engine \/ gap-1: \*\*landed\*\*\n  - review 0: \*\*changes_needed\*\*\n    - wrong \\\| controller\n  - re_review 1: \*\*approve\*\*\n/);
  assert.match(markdown, /- engine \/ gap-2: \*\*review_rejected\*\*\n  - final review: \*\*changes_needed\*\*\n    - legacy finding\n/);
  assert.match(markdown, /- card \/ batch: \*\*blocked\*\*\n\n/);
});

test("Configuration identity rejects changed workflow resume before execution", async () => {
  // This is read-only preflight, so it is valid even while CI runs the tests.
  await assert.rejects(runProduction({
    args: { mode: "preflight" },
    step: async () => ({ revision: "old-incompatible-code" }),
  }), /changed since this run/);
});

test("Agents are started with the policy's explicit model, effort and context tier", async t => {
  const runId = `test-explicit-models-${process.pid}`;
  const root = (await git(process.cwd(), ["rev-parse", "--show-toplevel"]));
  t.after(() => rm(path.join(root, ".cardwork/workflows", runId), { recursive: true, force: true }));
  const calls = [];
  await runProduction({
    runId, args: { mode: "telemetry-smoke" }, signal: undefined,
    step: async (_key, fn) => fn(), phase() {}, log() {},
    session: {
      sessionId: "session", on: () => () => {},
      workflow: { getRunDetail: async () => ({ runId, workflowName: "test", status: "running", observedAt: 0,
        agents: [], consumed: { nanoAiu: 0, activeMs: 0 } }) },
    },
    agent: async (_prompt, options) => { calls.push(options); return { ready: true }; },
  });
  const { model, reasoningEffort, contextTier } = policy.models.card.implementation;
  assert.equal(calls.length, 1);
  assert.deepEqual({ model: calls[0].model, reasoningEffort: calls[0].reasoningEffort, contextTier: calls[0].contextTier },
    { model, reasoningEffort, contextTier });
});

test("Production refuses to run in GitHub Actions without launching agents", async () => {
  const original = process.env.GITHUB_ACTIONS;
  process.env.GITHUB_ACTIONS = "true";
  try {
    await assert.rejects(runProduction({
      args: { mode: "produce", maxTasks: 1, corpusPath: "/unused" },
      agent: () => assert.fail("CI launched a production agent"),
    }), /not GitHub Actions/);
  } finally {
    if (original === undefined) delete process.env.GITHUB_ACTIONS;
    else process.env.GITHUB_ACTIONS = original;
  }
});

test("Branch protection cannot silently use admin bypass privileges", async t => {
  assert.deepEqual(checkProtection({ required_status_checks: { contexts: ["Race Tests"] } }, []).checks, ["Tests", "Race Tests"]);
  assert.equal(checkProtection({ required_pull_request_reviews: { required_approving_review_count: 1 } }, []).reviewRequired, true);
  assert.equal(checkProtection(null, [{ type: "pull_request", parameters: { required_approving_review_count: 1 } }]).reviewRequired, true);
  assert.throws(() => checkProtection(null, [{ type: "required_signatures" }]), /explicit integration/);
  const f = await repositoryFixture(t);
  const services = landingServices(f);
  services.protection = async () => { throw new Error("unsupported protection"); };
  await assert.rejects(land(f.repo, f.work, { title: "protected" }, f.expected, "session", undefined, services), /unsupported protection/);
  assert.equal(await remoteHead(f), f.base);
});
