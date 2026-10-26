import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { createHash, randomUUID } from "node:crypto";
import { mkdir, readFile, readdir, stat, lstat, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { pullRequestAPI, landPullRequest } from "./pull-request.mjs";

const execute = promisify(execFile);

export async function command(cwd, executable, args, signal, env = {}) {
  try {
    const result = await execute(executable, args, {
      cwd, signal, maxBuffer: 16 * 1024 * 1024, env: { ...process.env, GIT_PAGER: "cat", GIT_TERMINAL_PROMPT: "0", ...env },
    });
    return result.stdout.trim();
  } catch (error) {
    if (signal?.aborted) throw error;
    throw new Error(`${executable} ${args.join(" ")} failed (${error.code ?? "unknown"}):\n${error.stdout ?? ""}\n${error.stderr ?? ""}`, { cause: error });
  }
}

export const git = (cwd, args, signal, env) => command(cwd, "git", args, signal, env);

export async function checkCoordinator(root, head, upstream, signal) {
  if (await git(root, ["merge-base", head, upstream], signal) === head) return;
  // A completed squash merge has a different commit identity but the same
  // published tree; do not strand its original clean app worktree.
  if (await git(root, ["rev-parse", `${head}^{tree}`], signal) ===
      await git(root, ["rev-parse", `${upstream}^{tree}`], signal)) return;
  throw new Error("Coordinator has unlanded/divergent content. Integrate it deliberately before production.");
}

export function checkProtection(classic, rules) {
  const blockingRules = rules.filter(rule => ![
    "non_fast_forward", "deletion", "required_linear_history", "required_status_checks", "pull_request",
  ].includes(rule.type));
  const checks = [...new Set([
    ...(classic?.required_status_checks?.contexts ?? []),
    ...(classic?.required_status_checks?.checks ?? []).map(check => check.context),
  ])];
  if (classic?.restrictions ||
      classic?.lock_branch?.enabled || classic?.required_signatures?.enabled || blockingRules.length) {
    throw new Error("Unsupported main protection requires explicit integration; never use administrator bypass.");
  }
  for (const rule of rules.filter(rule => rule.type === "required_status_checks")) {
    for (const check of rule.parameters?.required_status_checks ?? []) checks.push(check.context);
  }
  return {
    checks: [...new Set(["Tests", "Race Tests", ...checks])],
    reviewRequired: Boolean(classic?.required_pull_request_reviews?.required_approving_review_count) ||
      rules.some(rule => rule.type === "pull_request" && rule.parameters?.required_approving_review_count > 0),
  };
}

export async function mainProtection(root, signal) {
  const repo = "repos/natefinch/council4";
  const branch = JSON.parse(await command(root, "gh", ["api", `${repo}/branches/main`], signal));
  const rules = JSON.parse(await command(root, "gh", ["api", `${repo}/rules/branches/main`], signal));
  let classic = null;
  if (branch.protected) {
    // A missing/unreadable classic protection is not treated as unprotected.
    classic = JSON.parse(await command(root, "gh", ["api", `${repo}/branches/main/protection`], signal));
  }
  return { protected: branch.protected, rules: rules.map(rule => rule.type), ...checkProtection(classic, rules) };
}

export async function changes(cwd, signal) {
  const tracked = await git(cwd, ["diff", "--no-renames", "--name-only", "HEAD"], signal);
  const untracked = await git(cwd, ["ls-files", "--others", "--exclude-standard"], signal);
  const files = [...new Set([...tracked.split("\n"), ...untracked.split("\n")].filter(Boolean))].sort();
  const hash = createHash("sha256");
  for (const file of files) {
    if (path.isAbsolute(file) || file.split("/").includes("..") || !/^[a-zA-Z0-9_./-]+$/.test(file)) {
      throw new Error("Changed path requires manual integration");
    }
    try {
      const metadata = await lstat(path.join(cwd, file));
      if (metadata.isSymbolicLink()) throw new Error("Changed symlink requires manual integration");
      const content = await readFile(path.join(cwd, file));
      // Index membership is bookkeeping, not reviewed content. Include framed
      // path/mode/bytes so git add and rename detection cannot alter the hash.
      hash.update(JSON.stringify([file, "file", Boolean(metadata.mode & 0o111), content.length]));
      hash.update(content);
    }
    catch (error) { if (error.code !== "ENOENT") throw error; hash.update(JSON.stringify([file, "deleted"])); }
  }
  return { files, digest: hash.digest("hex") };
}

export async function workspace(root, directory, runId, index, signal) {
  const branch = `council4-production/${runId}/${index}`;
  const target = path.join(directory, "work", String(index));
  await mkdir(path.dirname(target), { recursive: true });
  const listing = await git(root, ["worktree", "list", "--porcelain"], signal);
  if (!listing.split("\n").includes(`worktree ${target}`)) {
    await git(root, ["fetch", "origin", "main"], signal);
    await git(root, ["worktree", "add", "-b", branch, target, "refs/remotes/origin/main"], signal);
  }
  if (await git(target, ["branch", "--show-current"], signal) !== branch) throw new Error("Task worktree branch changed");
  return { path: target, branch, base: await git(target, ["rev-parse", "HEAD"], signal) };
}

// Apply a retained unit's uncommitted work, including new files, to a fresh
// task worktree on current main. The retained worktree and its index are not
// changed, and the result stays uncommitted like any other unit's work. If main
// changed the same lines, the conflicted files are left with conflict markers
// and listed in conflicts for an integration repair.
export async function seedFromRetained(source, work, signal) {
  if ((await changes(work.path, signal)).files.length) throw new Error("Resumed unit worktree is not clean; inspect it before seeding");
  const index = path.join(os.tmpdir(), `council4-seed-${randomUUID()}`);
  try {
    const env = { GIT_INDEX_FILE: index };
    const base = await git(source, ["rev-parse", "HEAD"], signal);
    await git(source, ["read-tree", base], signal, env);
    await git(source, ["add", "-A", "--", "."], signal, env);
    const tree = await git(source, ["write-tree"], signal, env);
    const snapshot = await git(source, ["commit-tree", tree, "-p", base, "-m", "Retained production unit"], signal);
    let output, conflicts = [];
    try {
      output = await git(work.path, ["merge-tree", "--write-tree", "--name-only", "--no-messages", work.base, snapshot], signal);
    } catch (error) {
      // Exit status 1 means conflicts: the written tree contains conflict markers.
      if (error.cause?.code !== 1 || !error.cause?.stdout) throw error;
      output = error.cause.stdout.trim();
      conflicts = output.split("\n").slice(1).filter(Boolean);
    }
    const merged = output.split("\n")[0];
    await git(work.path, ["read-tree", "-u", "-m", "HEAD", merged], signal);
    await git(work.path, ["reset", "-q"], signal);
    return { snapshot, retainedBase: base, tree: merged, conflicts };
  } finally {
    await rm(index, { force: true });
  }
}

// Files that still contain merge conflict markers.
export async function unresolvedConflicts(cwd, files, signal) {
  const remaining = [];
  for (const file of files) {
    let content;
    try { content = await readFile(path.join(cwd, file), "utf8"); }
    catch (error) { if (error.code === "ENOENT") continue; throw error; }
    if (/^(<{7}|>{7})( |$)/m.test(content)) remaining.push(file);
  }
  return remaining;
}

export async function land(root, work, task, expected, sessionId, signal, options = {}) {
  return landPullRequest({
    root, work, task, expected, sessionId, signal, git, changes,
    protection: mainProtection,
    api: pullRequestAPI(command, signal),
    ...options,
  });
}

export async function collectReports(root) {
  const output = [];
  let entries;
  try { entries = await readdir(root); }
  catch (error) { if (error.code === "ENOENT") return output; throw error; }
  for (const name of entries) {
    if (!(await stat(path.join(root, name))).isDirectory()) continue;
    try { output.push(JSON.parse(await readFile(path.join(root, name, "report.json"), "utf8"))); }
    catch (error) { if (error.code !== "ENOENT") throw error; }
  }
  return output;
}
