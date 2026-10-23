import test from "node:test";
import assert from "node:assert/strict";
import { readFile, mkdtemp, writeFile, mkdir, rm, chmod, rename } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { executeUnit, checkArgs, checkPolicy, checkSelection, assertCardBoundary, admissions, runProduction, validationPackagesForTask } from "./workflow.mjs";
import { git, changes, land, checkProtection, checkCoordinator } from "./git.mjs";
import { assessChecks, pullRequestAPI } from "./pull-request.mjs";
import { prepareCardBatch, verifyCardBatch } from "./queue.mjs";
import { chooseWork, knownGaps, gapTask, batchTask, issueNumber, engineIssue } from "./selection.mjs";
import { gapEvidence, earlierUnits } from "./workflow.mjs";
import { observeUsage, readReceipts, summarize } from "./telemetry.mjs";

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
  for (const args of [{}, { mode: "produce" }, { mode: "produce", maxTasks: 0, corpusPath: "x" }]) {
    assert.throws(() => checkArgs(args));
  }
  assert.throws(() => checkSelection({ taskType: "card", id: "../bad", title: "bad", instructions: "bad" }));
  assert.throws(() => checkPolicy({ ...policy, validationPackages: ["--help"] }));
  assert.throws(() => assertCardBoundary({ taskType: "card" }, ["mtg/rules/effects.go"]));
  assert.doesNotThrow(() => assertCardBoundary({ taskType: "card" }, ["mtg/rules/my_card_test.go", "mtg/cards/authored/my_card.go"]));
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

test("Configuration identity rejects changed workflow resume before execution", async () => {
  // This is read-only preflight, so it is valid even while CI runs the tests.
  await assert.rejects(runProduction({
    args: { mode: "preflight" },
    step: async () => ({ revision: "old-incompatible-code" }),
  }), /changed since this run/);
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
