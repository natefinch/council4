import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { createHash } from "node:crypto";
import { mkdir, readFile, readdir, stat, lstat } from "node:fs/promises";
import path from "node:path";

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
  const blockingRules = rules.filter(rule => !["non_fast_forward", "deletion", "required_linear_history"].includes(rule.type));
  const checks = [...new Set([
    ...(classic?.required_status_checks?.contexts ?? []),
    ...(classic?.required_status_checks?.checks ?? []).map(check => check.context),
  ])];
  if (checks.length || classic?.required_pull_request_reviews || classic?.restrictions ||
      classic?.lock_branch?.enabled || classic?.required_signatures?.enabled || blockingRules.length) {
    throw new Error(`Main protection requires a checked integration path, not an administrator bypass. Required checks: ${checks.join(", ") || "see ruleset"}. No main push attempted.`);
  }
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
  checkProtection(classic, rules);
  return { protected: branch.protected, rules: rules.map(rule => rule.type) };
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

export async function land(root, work, task, expected, sessionId, signal, protection = mainProtection) {
  let head = await git(work.path, ["rev-parse", "HEAD"], signal);
  if (head === work.base) {
    const current = await changes(work.path, signal);
    if (current.digest !== expected.digest || !current.files.length) throw new Error("Reviewed content changed or is empty");
    await git(work.path, ["add", "--", ...current.files], signal);
    if (await git(work.path, ["write-tree"], signal) !== expected.tree) throw new Error("Prepared tree no longer matches reviewed content");
    await git(work.path, ["commit", "-m", `${task.title}\n\nCo-authored-by: Copilot App <223556219+Copilot@users.noreply.github.com>\nCopilot-Session: ${sessionId}`], signal);
    head = await git(work.path, ["rev-parse", "HEAD"], signal);
  } else {
    // Crash recovery may find the workflow's commit but no journaled landing.
    if (await git(work.path, ["rev-parse", "HEAD^"], signal) !== work.base ||
        !await git(work.path, ["log", "-1", "--format=%B"], signal).then(text => text.includes(`Copilot-Session: ${sessionId}`))) {
      throw new Error("Unexpected task commits; refusing unreviewed publication");
    }
    if (await git(work.path, ["rev-parse", "HEAD^{tree}"], signal) !== expected.tree ||
        await git(work.path, ["status", "--porcelain"], signal)) throw new Error("Task changed after its landing commit");
  }
  await git(root, ["fetch", "origin", "main"], signal);
  const main = await git(root, ["rev-parse", "refs/remotes/origin/main"], signal);
  const contains = await git(root, ["merge-base", main, head], signal);
  if (contains === head) return { commit: head, main, published: true, recovered: true };
  if (main !== work.base) throw new Error("origin/main advanced. Preserve work; rebase/integration requires new validation and review in a new run.");
  await protection(root, signal);
  await git(work.path, ["push", "origin", `${head}:refs/heads/main`], signal);
  await git(root, ["fetch", "origin", "main"], signal);
  const confirmed = await git(root, ["rev-parse", "refs/remotes/origin/main"], signal);
  if (await git(root, ["merge-base", confirmed, head], signal) !== head) throw new Error("Remote main does not contain the landing commit");
  return { commit: head, main: confirmed, published: true, recovered: false };
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
