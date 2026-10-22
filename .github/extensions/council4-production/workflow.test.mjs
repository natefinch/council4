import test from "node:test";
import assert from "node:assert/strict";
import { readFile, mkdtemp, writeFile, mkdir, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { executeUnit, checkArgs, checkPolicy, checkSelection, assertCardBoundary, admissions, runProduction } from "./workflow.mjs";
import { git, changes, land, checkProtection } from "./git.mjs";
import { observeUsage, readReceipts, summarize } from "./telemetry.mjs";

const policy = JSON.parse(await readFile(new URL("./policy.json", import.meta.url)));
const complete = { status: "complete", reviewRequired: false, evidence: ["actual test"], gaps: [] };

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
  await writeFile(path.join(repo, "card.txt"), "baseline\n");
  await git(repo, ["add", "card.txt"]);
  await git(repo, ["commit", "-m", "baseline"]);
  await git(repo, ["remote", "add", "origin", remote]);
  await git(repo, ["push", "-u", "origin", "main"]);
  const base = await git(repo, ["rev-parse", "HEAD"]);
  const work = { path: path.join(directory, "task"), base };
  await git(repo, ["worktree", "add", "-b", "task", work.path, base]);
  await writeFile(path.join(work.path, "card.txt"), "implemented\n");
  const expected = await changes(work.path);
  await git(work.path, ["add", "card.txt"]);
  expected.tree = await git(work.path, ["write-tree"]);
  return { repo, remote, base, work, expected };
}

const remoteHead = f => git(f.repo, ["ls-remote", "origin", "refs/heads/main"]).then(out => out.split(/\s+/)[0]);
const noProtection = async () => {};

test("Landing verifies reviewed tree and remote main; crash replay does not duplicate commits", async t => {
  const f = await repositoryFixture(t);
  const first = await land(f.repo, f.work, { title: "Implement card" }, f.expected, "session", undefined, noProtection);
  assert.equal(first.published, true);
  assert.equal(await remoteHead(f), first.commit);
  const replay = await land(f.repo, f.work, { title: "Implement card" }, f.expected, "session", undefined, noProtection);
  assert.equal(replay.commit, first.commit);
  assert.equal(replay.recovered, true);
});

test("Changed submission, unreviewed crash commit and concurrent main reject publication", async t => {
  const f = await repositoryFixture(t);
  await writeFile(path.join(f.work.path, "card.txt"), "different\n");
  await assert.rejects(land(f.repo, f.work, { title: "bad" }, f.expected, "session"), /Reviewed content/);
  assert.equal(await remoteHead(f), f.base);
  await git(f.work.path, ["add", "card.txt"]);
  await git(f.work.path, ["commit", "-m", "unreviewed\n\nCopilot-Session: session"]);
  await assert.rejects(land(f.repo, f.work, { title: "bad" }, f.expected, "session"), /changed after/);
  const race = await repositoryFixture(t);
  await writeFile(path.join(race.repo, "other.txt"), "other writer\n");
  await git(race.repo, ["add", "other.txt"]);
  await git(race.repo, ["commit", "-m", "concurrent"]);
  await git(race.repo, ["push", "origin", "main"]);
  const main = await remoteHead(race);
  await assert.rejects(land(race.repo, race.work, { title: "card" }, race.expected, "session"), /main advanced/);
  assert.equal(await remoteHead(race), main);
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
  assert.throws(() => checkProtection({ required_status_checks: { contexts: ["Tests"] } }, []), /administrator bypass/);
  assert.throws(() => checkProtection({ required_status_checks: { checks: [{ context: "Tests", app_id: 1 }] } }, []), /administrator bypass/);
  assert.throws(() => checkProtection({ required_pull_request_reviews: {} }, []), /checked integration/);
  assert.throws(() => checkProtection(null, [{ type: "pull_request" }]), /checked integration/);
  assert.throws(() => checkProtection(null, [{ type: "required_signatures" }]), /checked integration/);
  assert.doesNotThrow(() => checkProtection(null, []));
  const f = await repositoryFixture(t);
  await assert.rejects(land(f.repo, f.work, { title: "protected" }, f.expected, "session", undefined,
    async () => { throw new Error("required checks missing"); }), /required checks missing/);
  assert.equal(await remoteHead(f), f.base);
});
