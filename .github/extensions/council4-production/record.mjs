import { mkdir, readFile } from "node:fs/promises";
import path from "node:path";
import { git } from "./git.mjs";
import { observeUsage, readReceipts, writeJSON } from "./telemetry.mjs";

// Workflows other than production keep no ledger of their own. recorded wraps
// such a workflow's run so each run leaves .cardwork/runs/<run>/run.json (plus
// usage receipts) in its worktree, mirrored to the repository's common Git
// directory so the dashboard still finds it after the worktree is removed.

export async function sharedRecordsDirectory(root) {
  const common = await git(root, ["rev-parse", "--path-format=absolute", "--git-common-dir"]);
  return path.join(common, "council4-production", "workflow-runs");
}

// Observed spend by model from usage receipts, in AIC (nano-AIU / 1e9).
export function costsByModel(receipts) {
  const models = new Map();
  const seen = new Set();
  for (const receipt of receipts) {
    if (receipt.kind !== "usage" || seen.has(receipt.key)) continue;
    seen.add(receipt.key);
    const model = receipt.model ?? "unknown";
    const row = models.get(model) ?? { model, calls: 0, aic: 0 };
    row.calls++;
    row.aic += (receipt.nanoAiu ?? 0) / 1e9;
    models.set(model, row);
  }
  return [...models.values()].sort((a, b) => b.aic - a.aic);
}

export function recorded(workflow, run) {
  return async ctx => {
    const root = await git(process.cwd(), ["rev-parse", "--show-toplevel"], ctx.signal);
    const directory = path.join(root, ".cardwork", "runs", ctx.runId);
    const shared = path.join(await sharedRecordsDirectory(root), ctx.runId);
    await mkdir(directory, { recursive: true });
    await mkdir(shared, { recursive: true });
    const receipts = path.join(directory, "usage.jsonl");
    const file = path.join(directory, "run.json");
    const now = () => new Date().toISOString();
    let record;
    try {
      record = JSON.parse(await readFile(file, "utf8"));
      record.resumedAt = [...(record.resumedAt ?? []), now()];
    } catch (error) {
      if (error.code !== "ENOENT") throw error;
      record = { workflow, runId: ctx.runId, args: ctx.args ?? {}, startedAt: now(), phases: [] };
    }
    Object.assign(record, { worktree: root, outcome: null, finishedAt: null, error: null });

    const save = async () => {
      record.updatedAt = now();
      record.models = costsByModel(readReceipts(receipts));
      try {
        const detail = await ctx.session.workflow.getRunDetail(ctx.runId);
        record.aic = detail.consumed.nanoAiu / 1e9;
        record.activeMs = detail.consumed.activeMs;
        record.agents = detail.agents.length;
      } catch (error) {
        record.telemetryError = error.message;
      }
      writeJSON(file, record);
      writeJSON(path.join(shared, "run.json"), record);
    };
    // Saves are serialized so a slow telemetry read never overwrites a later state.
    let saving = Promise.resolve();
    const queueSave = () => (saving = saving.then(save, save));

    // ctx is forwarded unchanged except phase(), which is also recorded. The
    // proxy target is a blank object so ctx's own property invariants don't apply.
    const phase = title => {
      record.phase = title;
      record.phases.push({ title, at: now() });
      queueSave().catch(() => {});
      return ctx.phase(title);
    };
    const forwarded = new Proxy({}, {
      get(_, key) {
        if (key === "phase") return phase;
        const value = Reflect.get(ctx, key);
        return typeof value === "function" ? value.bind(ctx) : value;
      },
      has: (_, key) => key in ctx,
    });

    const off = observeUsage(ctx.session, ctx.runId, receipts, () => {});
    await queueSave();
    try {
      const result = await run(forwarded);
      record.outcome = "finished";
      return result;
    } catch (error) {
      record.outcome = ctx.signal?.aborted ? "cancelled" : "error";
      record.error = error?.message ?? String(error);
      throw error;
    } finally {
      record.finishedAt = now();
      await queueSave().catch(() => {});
      off();
    }
  };
}
