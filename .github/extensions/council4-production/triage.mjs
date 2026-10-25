import { readFile, writeFile, mkdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { git, command } from "./git.mjs";
import { observeUsage } from "./telemetry.mjs";
import { defaultCorpusPath, existingCardDefectLabel } from "./workflow.mjs";

// Backlog triage: a cheap read-only pass over open Card Support TODO issues
// that finds stale and duplicate issues, defects in rules already-supported
// cards use, and which queued cards each remaining gap blocks. Its output
// ranks Engine work for the production workflow.

export const triageMeta = {
  name: "council4-engine-triage",
  description: "Rank open Card Support TODO issues: stale/duplicate, existing-card defects, and queued cards each blocks. Read-only agents; apply=true labels defects and closes stale/duplicate issues with evidence.",
  phases: [{ title: "Index" }, { title: "Triage" }, { title: "Report" }, { title: "Apply" }],
  argsSchema: {
    type: "object",
    properties: {
      apply: { type: "boolean" },
      fromRunId: { type: "string" },
      issues: { type: "array", items: { type: "integer" } },
      batchSize: { type: "integer" },
      concurrency: { type: "integer" },
    },
  },
};

const text = { type: "string" };
const verdictSchema = {
  type: "object", required: ["issues"],
  properties: {
    issues: {
      type: "array",
      items: {
        type: "object",
        required: ["issue", "verdict", "evidence", "existingCardDefect", "registeredExamples", "blockedOracleIds", "blockedEstimate", "size"],
        properties: {
          issue: { type: "integer" },
          verdict: { type: "string", enum: ["valid", "stale", "duplicate"] },
          duplicateOf: { type: "integer" },
          evidence: text,
          existingCardDefect: { type: "boolean" },
          registeredExamples: { type: "array", items: text },
          blockedOracleIds: { type: "array", items: text },
          blockedEstimate: { type: "integer" },
          size: { type: "string", enum: ["small", "medium", "large"] },
        },
      },
    },
  },
};

function faces(card) {
  if (card.oracle_text !== undefined) return card.oracle_text;
  return (card.card_faces ?? []).map(face => `${face.name}: ${face.type_line ?? ""}\n${face.oracle_text ?? ""}`).join("\n//\n");
}

// One JSON line per card, searchable with rg by name, type or rules text.
async function writeIndexes(root, directory, signal) {
  const manifest = JSON.parse(await readFile(path.join(root, "testdata/card-authoring/production.json"), "utf8"));
  const corpus = JSON.parse(await readFile(defaultCorpusPath(manifest.corpus_sha256), "utf8"));
  const cards = new Map(corpus.map(card => [card.oracle_id, card]));
  const line = (record, extra) => {
    const card = cards.get(record.oracle_id) ?? {};
    return JSON.stringify({ oracle_id: record.oracle_id, name: record.name, type_line: card.type_line ?? "",
      mana_cost: card.mana_cost ?? "", text: faces(card).replace(/\s+/g, " "), ...extra });
  };
  const queued = manifest.records.filter(record => ["unclaimed", "blocked"].includes(record.status) && !record.batch);
  const registered = manifest.records.filter(record => ["installed-compiled", "implemented-verified"].includes(record.status));
  await writeFile(path.join(directory, "queued-cards.jsonl"), queued.map(record => line(record,
    { status: record.status, blocker: record.blocker ?? null, decks: record.deck_occurrences ?? 0 })).join("\n") + "\n");
  await writeFile(path.join(directory, "registered-cards.jsonl"), registered.map(record => line(record, {})).join("\n") + "\n");
  return { manifest, queuedCount: queued.length, registeredCount: registered.length };
}

async function openIssues(root, signal) {
  return JSON.parse(await command(root, "gh", ["issue", "list", "--repo", "natefinch/council4", "--state", "open",
    "--label", "Card Support TODO", "--json", "number,title,body,labels", "--limit", "1000"], signal))
    .map(issue => ({ number: issue.number, title: issue.title, body: (issue.body ?? "").slice(0, 4000),
      labels: issue.labels.map(label => label.name) }))
    .sort((a, b) => a.number - b.number);
}

export function chunk(items, size) {
  const out = [];
  for (let i = 0; i < items.length; i += size) out.push(items.slice(i, i + size));
  return out;
}

// Deterministic merge of agent verdicts: drop unknown cards, give each queued
// card to the highest-priority issue that claims it, and rank issues.
export function rankTriage(verdicts, issues, queued) {
  const byNumber = new Map(issues.map(issue => [issue.number, issue]));
  const rows = verdicts.filter(v => byNumber.has(v.issue)).map(v => ({
    ...v, title: byNumber.get(v.issue).title,
    blockedOracleIds: [...new Set(v.blockedOracleIds)].filter(id => queued.get(id)?.status === "unclaimed"),
  }));
  const valid = rows.filter(row => row.verdict === "valid")
    .sort((a, b) => (b.existingCardDefect - a.existingCardDefect) || (b.blockedEstimate - a.blockedEstimate) || (a.issue - b.issue));
  const owner = new Map();
  for (const row of valid) {
    row.assignedOracleIds = row.blockedOracleIds.filter(id => !owner.has(id));
    for (const id of row.assignedOracleIds) owner.set(id, row.issue);
  }
  const blockers = valid.flatMap(row => row.assignedOracleIds.map(id => {
    const record = queued.get(id);
    return { oracle_id: id, revision: record.revision, name: record.name, blocker: row.issue,
      reason: `Backlog triage: needs #${row.issue} (${row.title})`.slice(0, 300) };
  }));
  return { valid, stale: rows.filter(row => row.verdict === "stale"), duplicate: rows.filter(row => row.verdict === "duplicate"),
    missing: issues.filter(issue => !rows.some(row => row.issue === issue.number)).map(issue => issue.number), blockers };
}

function markdown(ranked, meta) {
  const cell = value => String(value ?? "").replaceAll("|", "\\|").replaceAll("\n", " ");
  return [
    `# Engine backlog triage`, "",
    `${meta.issues} open Card Support TODO issues; ${meta.queued} queued cards; ${meta.registered} registered cards.`, "",
    `Valid ${ranked.valid.length} (existing-card defects ${ranked.valid.filter(r => r.existingCardDefect).length}), stale ${ranked.stale.length}, duplicate ${ranked.duplicate.length}, not triaged ${ranked.missing.length}. Queue blockers to record: ${ranked.blockers.length}.`, "",
    "## Ranked valid issues", "",
    "| Rank | Issue | Existing-card defect | Blocks (est.) | Assigned queued cards | Size | Title |",
    "| ---: | --- | --- | ---: | ---: | --- | --- |",
    ...ranked.valid.map((row, i) => `| ${i + 1} | #${row.issue} | ${row.existingCardDefect ? `yes: ${cell(row.registeredExamples.slice(0, 3).join(", "))}` : ""} | ${row.blockedEstimate} | ${row.assignedOracleIds.length} | ${row.size} | ${cell(row.title)} |`),
    "", "## Stale (capability exists)", "",
    ...ranked.stale.map(row => `- #${row.issue} ${cell(row.title)}: ${cell(row.evidence)}`),
    "", "## Duplicates", "",
    ...ranked.duplicate.map(row => `- #${row.issue} duplicate of #${row.duplicateOf}: ${cell(row.evidence)}`),
    ...(ranked.missing.length ? ["", `Not triaged (agent failure): ${ranked.missing.map(n => `#${n}`).join(", ")}`] : []),
  ].join("\n") + "\n";
}

async function apply(root, ranked, signal) {
  const gh = args => command(root, "gh", args, signal);
  const labels = JSON.parse(await gh(["label", "list", "--repo", "natefinch/council4", "--json", "name", "--limit", "200"]));
  if (!labels.some(label => label.name === existingCardDefectLabel)) {
    await gh(["label", "create", existingCardDefectLabel, "--repo", "natefinch/council4", "--color", "B60205",
      "--description", "Defect in rules that already-supported cards use; fixed before new card work"]);
  }
  const applied = { labeled: [], closed: [] };
  for (const row of ranked.valid.filter(row => row.existingCardDefect)) {
    await gh(["issue", "edit", String(row.issue), "--repo", "natefinch/council4", "--add-label", existingCardDefectLabel]);
    await gh(["issue", "comment", String(row.issue), "--repo", "natefinch/council4", "--body",
      `Backlog triage: already-supported cards hit this defect (e.g. ${row.registeredExamples.slice(0, 3).join(", ")}). ${row.evidence}`]);
    applied.labeled.push(row.issue);
  }
  for (const row of [...ranked.stale, ...ranked.duplicate]) {
    const why = row.verdict === "duplicate" ? `Duplicate of #${row.duplicateOf}. ${row.evidence}` : `The capability appears to exist on main. ${row.evidence}`;
    await gh(["issue", "close", String(row.issue), "--repo", "natefinch/council4", "--reason", "not planned",
      "--comment", `Backlog triage: ${why}\n\nReopen if this is wrong.`]);
    applied.closed.push(row.issue);
  }
  return applied;
}

export async function runTriage(ctx) {
  const root = await git(process.cwd(), ["rev-parse", "--show-toplevel"], ctx.signal);
  const source = path.dirname(fileURLToPath(import.meta.url));
  const policy = JSON.parse(await readFile(path.join(source, "policy.json"), "utf8"));
  const settings = policy.models.triage;
  if (!settings?.model) throw new Error("policy.models.triage is required");
  // apply=true with fromRunId applies an earlier, reviewed triage without new agents.
  if (ctx.args.fromRunId) {
    if (!ctx.args.apply) throw new Error("fromRunId is only for applying an earlier triage");
    const ranked = JSON.parse(await readFile(path.join(root, ".cardwork", "triage", ctx.args.fromRunId, "triage.json"), "utf8"));
    ctx.phase("Apply");
    return { applied: await apply(root, ranked, ctx.signal), blockersPlan: path.join(root, ".cardwork", "triage", ctx.args.fromRunId, "blockers-plan.json") };
  }
  const directory = path.join(root, ".cardwork", "triage", ctx.runId);
  await mkdir(directory, { recursive: true });
  const off = observeUsage(ctx.session, ctx.runId, path.join(directory, "usage.jsonl"), () => {});
  try {
    ctx.phase("Index");
    const index = await ctx.step("index-v1", async () => {
      const { manifest, queuedCount, registeredCount } = await writeIndexes(root, directory, ctx.signal);
      const issues = (await openIssues(root, ctx.signal)).filter(issue => !ctx.args.issues || ctx.args.issues.includes(issue.number));
      return { corpus: manifest.corpus_sha256, queuedCount, registeredCount, issues };
    });
    const manifest = JSON.parse(await readFile(path.join(root, "testdata/card-authoring/production.json"), "utf8"));
    const queued = new Map(manifest.records.filter(record => ["unclaimed", "blocked"].includes(record.status) && !record.batch)
      .map(record => [record.oracle_id, record]));
    const titles = index.issues.map(issue => `#${issue.number} ${issue.title}`).join("\n");
    const batches = chunk(index.issues, ctx.args.batchSize ?? 8);
    ctx.phase("Triage");
    const verdicts = [];
    for (const wave of chunk(batches.map((batch, i) => ({ batch, i })), ctx.args.concurrency ?? 6)) {
      const results = await ctx.parallel(wave.map(({ batch, i }) => () => ctx.agent(
        `Triage these open Council4 Engine-gap issues. READ ONLY: do not edit files, run git commands that change state, or touch issues.
Council4 is a Go Magic: The Gathering rules engine; read .github/copilot-instructions.md and docs/card-authoring.md briefly for context. The checkout at ${root} is current main.
For EACH issue below decide:
- verdict "stale" if main already supports the capability (cite file:line and a behavioral test that exercises it), "duplicate" if another open issue in the list covers the same capability (set duplicateOf; the lower issue number is the original), otherwise "valid". When unsure, use "valid".
- existingCardDefect: true only if cards already supported (search ${directory}/registered-cards.jsonl, one JSON card per line with name, type_line and text) exercise the defective rule today, so they play wrong now. Name up to 3 such cards in registeredExamples. A missing capability that no supported card uses is not an existing-card defect.
- blockedOracleIds: oracle_id values from ${directory}/queued-cards.jsonl (unbuilt cards) whose printed text needs this capability; search with rg on distinctive rules text. List at most 40, and set blockedEstimate to your estimate of the total count.
- size: small (one rule, under ~200 lines), medium, or large (several rules; would need splitting).
Keep evidence to 1-3 sentences. Be efficient: a few targeted rg searches per issue, no broad code reading.
All open issue titles, for duplicate checks:
${titles}

Issues to triage:
${JSON.stringify(batch)}`,
        { label: `triage/batch-${i}`, schema: verdictSchema, model: settings.model,
          reasoningEffort: settings.reasoningEffort, contextTier: settings.contextTier })));
      for (const result of results) verdicts.push(...(result?.issues ?? []));
      await writeFile(path.join(directory, "verdicts.json"), JSON.stringify(verdicts, null, 2) + "\n");
    }
    ctx.phase("Report");
    const ranked = rankTriage(verdicts, index.issues, queued);
    await writeFile(path.join(directory, "triage.json"), JSON.stringify(ranked, null, 2) + "\n");
    await writeFile(path.join(directory, "blockers-plan.json"), JSON.stringify({ corpus_sha256: index.corpus, records: ranked.blockers }, null, 2) + "\n");
    await writeFile(path.join(directory, "report.md"), markdown(ranked, { issues: index.issues.length, queued: index.queuedCount, registered: index.registeredCount }));
    let applied = null;
    if (ctx.args.apply) { ctx.phase("Apply"); applied = await apply(root, ranked, ctx.signal); }
    return { directory, valid: ranked.valid.length, existingCardDefects: ranked.valid.filter(r => r.existingCardDefect).length,
      stale: ranked.stale.length, duplicate: ranked.duplicate.length, missing: ranked.missing, blockers: ranked.blockers.length, applied,
      next: "Review report.md. Then apply with {apply:true, fromRunId} and record blockers-plan.json with cardqueue -operation record-blockers in a PR." };
  } finally {
    off();
  }
}
