import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { createHash } from "node:crypto";
import { mkdir, readFile, readdir, stat, lstat } from "node:fs/promises";
import path from "node:path";
import { pullRequestAPI, landPullRequest } from "./pull-request.mjs";

const execute = promisify(execFile);

export async function command(cwd, executable, args, signal) {
  try {
    const result = await execute(executable, args, {
      cwd, signal, maxBuffer: 16 * 1024 * 1024, env: { ...process.env, GIT_PAGER: "cat", GIT_TERMINAL_PROMPT: "0" },
    });
    return result.stdout.trim();
  } catch (error) {
    if (signal?.aborted) throw error;
    throw new Error(`${executable} ${args.join(" ")} failed (${error.code ?? "unknown"}):\n${error.stdout ?? ""}\n${error.stderr ?? ""}`, { cause: error });
  }
}

export const git = (cwd, args, signal) => command(cwd, "git", args, signal);

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
  const tracked = await git(cwd, ["diff", "--name-only", "HEAD"], signal);
  const untracked = await git(cwd, ["ls-files", "--others", "--exclude-standard"], signal);
  const files = [...new Set([...tracked.split("\n"), ...untracked.split("\n")].filter(Boolean))].sort();
  const hash = createHash("sha256");
  hash.update(await git(cwd, ["diff", "--binary", "HEAD"], signal));
  for (const file of files) {
    if (path.isAbsolute(file) || file.split("/").includes("..") || !/^[a-zA-Z0-9_./-]+$/.test(file)) {
      throw new Error("Changed path requires manual integration");
    }
    hash.update(file);
    try {
      if ((await lstat(path.join(cwd, file))).isSymbolicLink()) throw new Error("Changed symlink requires manual integration");
      hash.update(await readFile(path.join(cwd, file)));
    }
    catch (error) { if (error.code !== "ENOENT") throw error; hash.update("<deleted>"); }
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
