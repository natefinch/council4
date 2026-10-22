import { mkdir, readFile } from "node:fs/promises";
import path from "node:path";
import { writeJSON } from "./telemetry.mjs";

const repository = "natefinch/council4";

export function assessChecks(pr, checks, required, expectedHead) {
  if (pr.headRefOid !== expectedHead || pr.baseRefName !== "main" || pr.isDraft || pr.state !== "OPEN") {
    throw new Error("PR is not the expected open main-bound reviewed head");
  }
  if (checks.head !== expectedHead) throw new Error("CI observations belong to a stale commit");
  const latest = new Map();
  for (const check of checks.runs) {
    const previous = latest.get(check.name);
    if (!previous || check.id > previous.id) latest.set(check.name, check);
  }
  const failed = checks.runs.filter(check => latest.get(check.name) === check &&
    check.status === "completed" && !["success", "neutral", "skipped"].includes(check.conclusion));
  const statusFailures = checks.statuses.filter(check => ["failure", "error"].includes(check.state));
  if (failed.length || statusFailures.length) return { status: "failed", failed: [...failed, ...statusFailures] };
  for (const name of required.checks) {
    const run = latest.get(name);
    const status = checks.statuses.find(check => check.context === name);
    if (run) {
      if (run.status !== "completed") return { status: "pending" };
      if (run.conclusion !== "success") return { status: "failed", failed: [run] };
      if (["Tests", "Race Tests"].includes(name) && run.app?.slug !== "github-actions") {
        throw new Error(`Required ${name} result is not a GitHub Actions check run`);
      }
    } else {
      if (["Tests", "Race Tests"].includes(name) || status?.state !== "success") return { status: "pending" };
    }
  }
  if ([...latest.values()].some(check => check.status !== "completed") ||
      checks.statuses.some(check => check.state === "pending")) return { status: "pending" };
  if (required.reviewRequired && pr.reviewDecision !== "APPROVED") return { status: "review_required" };
  if (["CHANGES_REQUESTED", "REVIEW_REQUIRED"].includes(pr.reviewDecision)) return { status: "review_required" };
  if (pr.mergeStateStatus === "BEHIND") return { status: "behind" };
  if (pr.mergeStateStatus !== "CLEAN") return { status: "not_mergeable", mergeStateStatus: pr.mergeStateStatus };
  return { status: "ready" };
}

export function pullRequestAPI(command, signal) {
  const gh = (cwd, args) => command(cwd, "gh", args, signal);
  const fields = "number,url,state,isDraft,headRefOid,baseRefName,baseRefOid,mergeStateStatus,reviewDecision,mergeCommit";
  return {
    async find(cwd, branch) {
      const all = JSON.parse(await gh(cwd, ["pr", "list", "--repo", repository, "--head", branch,
        "--base", "main", "--state", "all", "--json", fields, "--limit", "10"]));
      if (all.length > 1) throw new Error("Ambiguous existing workflow PRs");
      return all[0] ?? null;
    },
    async create(cwd, branch, task) {
      // Unit worktrees aren't the app session's tracked head; use the REST API
      // rather than the current-session PR tool. No admin/auto-merge flags.
      const pr = JSON.parse(await gh(cwd, ["api", `repos/${repository}/pulls`, "--method", "POST",
        "-f", `head=${branch}`, "-f", "base=main", "-f", `title=${task.title}`,
        "-f", `body=Production ${task.taskType} unit \`${task.id}\`.\n\nLocal behavioral validation and applicable independent review completed. The workflow waits for Tests, Race Tests, every other reported check, and repository-required approval before merging. CI fixes and integration re-review are separately attributed. No protection bypass.`]));
      return this.view(cwd, pr.number);
    },
    async view(cwd, number) {
      return JSON.parse(await gh(cwd, ["pr", "view", String(number), "--repo", repository, "--json", fields]));
    },
    async checks(cwd, head) {
      const pages = JSON.parse(await gh(cwd, ["api", `repos/${repository}/commits/${head}/check-runs?per_page=100`, "--paginate", "--slurp"]));
      const statusPages = JSON.parse(await gh(cwd, ["api", `repos/${repository}/commits/${head}/statuses?per_page=100`, "--paginate", "--slurp"]));
      const byContext = new Map();
      for (const status of statusPages.flat()) {
        if (!byContext.has(status.context)) byContext.set(status.context, status);
      }
      return { head, runs: pages.flatMap(page => page.check_runs), statuses: [...byContext.values()] };
    },
    async watch(cwd, number) {
      try {
        await gh(cwd, ["pr", "checks", String(number), "--repo", repository, "--watch", "--fail-fast", "--interval", "15"]);
      } catch (error) {
        // gh uses 1 for a failed check, 8 for pending. Re-read authoritative
        // head-specific checks; authentication/network/other failures propagate.
        if (![1, 8].includes(error.cause?.code)) throw error;
      }
    },
    async failedLogs(cwd, checks) {
      const urls = checks.failed.map(check => check.html_url ?? check.target_url ?? check.details_url).filter(Boolean);
      const runIds = [...new Set(urls.map(url => /\/actions\/runs\/(\d+)/.exec(url)?.[1]).filter(Boolean))];
      const logs = [];
      for (const run of runIds) logs.push(await gh(cwd, ["run", "view", run, "--repo", repository, "--log-failed"]));
      return { checks, logs: logs.map(log => log.slice(-32000)) };
    },
    async merge(cwd, number, head) {
      await gh(cwd, ["pr", "merge", String(number), "--repo", repository, "--squash", "--match-head-commit", head]);
      return this.view(cwd, number);
    },
  };
}

export async function landPullRequest({
  root, work, task, expected, sessionId, signal, git, changes, protection, api,
  validate, reviewIntegration, fixCI, ciFixRounds = 1, checkpoint,
  watchCI = watch => watch(),
}) {
  if (!validate || !reviewIntegration) throw new Error("PR landing requires local validation and integration-review hooks");
  const file = checkpoint ?? path.join(work.path, ".cardwork", "production-landing.json");
  await mkdir(path.dirname(file), { recursive: true });
  let state;
  try { state = JSON.parse(await readFile(file, "utf8")); }
  catch (error) { if (error.code !== "ENOENT") throw error; }
  if (state && (state.base !== work.base || state.reviewedTree !== expected.tree || state.sessionId !== sessionId)) {
    throw new Error("Landing checkpoint does not match the reviewed unit");
  }
  const save = () => writeJSON(file, state);
  const head = () => git(work.path, ["rev-parse", "HEAD"], signal);
  const fetchMain = async () => {
    await git(root, ["fetch", "origin", "main"], signal);
    return git(root, ["rev-parse", "refs/remotes/origin/main"], signal);
  };
  const clean = async () => {
    if (await git(work.path, ["status", "--porcelain"], signal)) throw new Error("Landing checkout contains unreviewed edits");
  };
  const finishPending = async () => {
    const pending = state.pending;
    if (!pending) return;
    if (pending.kind === "merge") {
      if (await head() === pending.parent) {
        await clean();
        await git(work.path, ["merge", "--no-edit", pending.main], signal);
      }
      const parents = (await git(work.path, ["rev-list", "--parents", "-n", "1", "HEAD"], signal)).split(/\s+/);
      if (parents.length !== 3 || parents[1] !== pending.parent || parents[2] !== pending.main) {
        throw new Error("Interrupted integration does not match the saved merge parents");
      }
      state.head = await head();
      state.validatedHead = null;
      state.reviewedHead = await git(work.path, ["rev-parse", "HEAD^{tree}"], signal) === pending.beforeTree ? state.head : null;
    } else if (pending.kind === "ci_commit") {
      if (await head() === pending.parent) {
        if (await git(work.path, ["write-tree"], signal) !== pending.tree ||
            await git(work.path, ["diff", "--name-only"], signal) ||
            await git(work.path, ["ls-files", "--others", "--exclude-standard"], signal)) {
          throw new Error("Interrupted CI repair changed after review");
        }
        await git(work.path, ["commit", "-m", pending.message], signal);
      }
      if (await git(work.path, ["rev-parse", "HEAD^"], signal) !== pending.parent ||
          await git(work.path, ["rev-parse", "HEAD^{tree}"], signal) !== pending.tree ||
          (await git(work.path, ["log", "-1", "--format=%B"], signal)).trim() !== pending.message.trim()) {
        throw new Error("Interrupted CI commit does not match the reviewed tree and parent");
      }
      state.head = await head();
      state.validatedHead = null;
      state.reviewedHead = state.head;
    } else {
      throw new Error("Unknown pending landing mutation");
    }
    delete state.pending;
    save();
  };
  if (!state) {
    const currentHead = await head();
    if (currentHead === work.base) {
      const current = await changes(work.path, signal);
      if (current.digest !== expected.digest || !current.files.length) throw new Error("Reviewed content changed or is empty");
      await git(work.path, ["add", "--", ...current.files], signal);
      if (await git(work.path, ["write-tree"], signal) !== expected.tree) throw new Error("Prepared tree changed");
      await git(work.path, ["commit", "-m", `${task.title}\n\nCo-authored-by: Copilot App <223556219+Copilot@users.noreply.github.com>\nCopilot-Session: ${sessionId}`], signal);
    } else {
      if (await git(work.path, ["rev-parse", "HEAD^"], signal) !== work.base ||
          await git(work.path, ["rev-parse", "HEAD^{tree}"], signal) !== expected.tree ||
          !(await git(work.path, ["log", "-1", "--format=%B"], signal)).includes(`Copilot-Session: ${sessionId}`)) {
        throw new Error("Unexpected task commits; refusing unreviewed publication");
      }
      await clean();
    }
    state = { base: work.base, reviewedTree: expected.tree, sessionId, head: await head(),
      reviewedHead: await head(), round: 0, ciFixes: 0 };
    save();
  }
  await finishPending();
  await clean();
  if (await head() !== state.head) throw new Error("Task head differs from the durable landing checkpoint");
  let pr = state.pr ? await api.view(work.path, state.pr.number) : await api.find(work.path, work.branch);
  if (pr?.state === "MERGED") return confirmMerged(pr);
  if (pr && pr.state !== "OPEN") throw new Error("Workflow PR was closed without merging; preserve the work");

  async function confirmMerged(merged) {
    if (merged.headRefOid !== state.head || !merged.mergeCommit?.oid) throw new Error("Merged PR does not match the validated workflow head");
    const main = await fetchMain();
    if (await git(root, ["merge-base", main, merged.mergeCommit.oid], signal) !== merged.mergeCommit.oid) {
      throw new Error("PR merge commit is not on remote main");
    }
    return { commit: merged.mergeCommit.oid, head: state.head, main, published: true,
      pr: { number: merged.number, url: merged.url }, ciFixes: state.ciFixes };
  }

  // One pass either merges, fixes a real failure, integrates newer main, or
  // returns a precise pending state. GitHub's watch handles active CI waits.
  for (;;) {
    const required = await protection(root, signal);
    const main = await fetchMain();
    if (await git(work.path, ["merge-base", main, state.head], signal) !== main) {
      const before = await git(work.path, ["rev-parse", "HEAD^{tree}"], signal);
      state.round++;
      state.pending = { kind: "merge", parent: state.head, main, beforeTree: before };
      save();
      await finishPending();
    }
    if (state.reviewedHead !== state.head) {
      await reviewIntegration({ round: state.round, head: state.head, main });
      await clean();
      if (await head() !== state.head) throw new Error("Integration reviewer committed outside landing");
      state.reviewedHead = state.head;
      save();
    }
    if (state.validatedHead !== state.head) {
      await validate({ round: state.round, head: state.head, main });
      await clean();
      if (await head() !== state.head) throw new Error("Validation changed the landing head");
      state.validatedHead = state.head;
      save();
    }
    // Push only the task branch. Never write refs/heads/main.
    await git(work.path, ["push", "origin", `HEAD:refs/heads/${work.branch}`], signal);
    pr ??= await api.create(work.path, work.branch, task);
    state.pr = { number: pr.number, url: pr.url };
    save();
    pr = await api.view(work.path, pr.number);
    if (pr.state === "MERGED") return confirmMerged(pr);
    let checks = await api.checks(work.path, state.head);
    let assessed = assessChecks(pr, checks, required, state.head);
    if (assessed.status === "pending" && (checks.runs.length || checks.statuses.length)) {
      await watchCI(() => api.watch(work.path, pr.number), { pr: state.pr, head: state.head });
      pr = await api.view(work.path, pr.number);
      if (pr.state === "MERGED") return confirmMerged(pr);
      checks = await api.checks(work.path, state.head);
      assessed = assessChecks(pr, checks, required, state.head);
    }
    if (assessed.status === "failed") {
      if (!fixCI || state.ciFixes >= ciFixRounds) {
        return { status: "ci_failed", published: false, pr: state.pr, checks: assessed };
      }
      const before = state.head;
      state.ciFixes++;
      state.round++;
      save();
      await fixCI({ round: state.ciFixes, head: before, ...(await api.failedLogs(work.path, assessed)) });
      if (await head() !== before) throw new Error("CI fixer committed outside landing");
      const fixed = await changes(work.path, signal);
      if (!fixed.files.length) throw new Error("CI fixer produced no changes; no empty success");
      await git(work.path, ["add", "--", ...fixed.files], signal);
      state.pending = { kind: "ci_commit", parent: before,
        tree: await git(work.path, ["write-tree"], signal),
        message: `Fix CI: ${task.title}\n\nCo-authored-by: Copilot App <223556219+Copilot@users.noreply.github.com>\nCopilot-Session: ${sessionId}` };
      save();
      await finishPending();
      continue;
    }
    if (assessed.status === "behind") {
      if (await fetchMain() !== main) continue;
      return { status: "pending", reason: "GitHub still reports the PR behind main", published: false, pr: state.pr };
    }
    if (assessed.status !== "ready") return { ...assessed, published: false, pr: state.pr };
    if (await fetchMain() !== main) continue;
    // Re-read the PR head and check state immediately before a head-matched,
    // ordinary merge. Never --admin, false statuses or protection mutations.
    pr = await api.view(work.path, pr.number);
    const last = assessChecks(pr, await api.checks(work.path, state.head), await protection(root, signal), state.head);
    if (last.status !== "ready") return { ...last, published: false, pr: state.pr };
    return confirmMerged(await api.merge(work.path, pr.number, state.head));
  }
}
