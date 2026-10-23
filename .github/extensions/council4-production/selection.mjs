// Deterministic work selection. Known Engine gaps come first so card work does
// not repeatedly trip over them; then saved ready batches; only when neither
// exists does a model prepare a new reviewed batch.

// Gap references must name an issue explicitly; batch IDs such as
// "08-scry-interaction" are not issue numbers.
export function issueNumber(value) {
  const text = String(value ?? "").trim();
  const match = /^(?:#|gap-|issue-)?(\d+)$/.exec(text) ?? /\/issues\/(\d+)$/.exec(text);
  return match ? Number(match[1]) : null;
}

// The Engine issue a unit targeted: explicit for deterministic gap units,
// otherwise the leading issue number of the selected task ID.
export function engineIssue(unit) {
  if (unit?.task?.taskType !== "engine") return null;
  const legacy = /^(\d+)-/.exec(unit.task.id ?? "");
  return unit.task.issue ?? issueNumber(unit.task.id) ?? (legacy ? Number(legacy[1]) : null);
}

// Known gaps, in priority order: those demonstrated by blocked card work (this
// run first, then earlier local runs), then saved queue prerequisites.
export function knownGaps({ next, units = [], priorUnits = [] }) {
  const gaps = new Map();
  const add = (issue, oracleIds, source) => {
    if (!Number.isInteger(issue) || issue <= 0) return;
    const gap = gaps.get(issue) ?? { issue, oracleIds: [], sources: [] };
    for (const id of oracleIds ?? []) if (!gap.oracleIds.includes(id)) gap.oracleIds.push(id);
    if (!gap.sources.includes(source)) gap.sources.push(source);
    gaps.set(issue, gap);
  };
  for (const [list, source] of [[units, "this run"], [priorUnits, "earlier run"]]) {
    for (const unit of list) {
      if (unit?.task?.taskType !== "card" || unit.status !== "blocked") continue;
      const oracleIds = (unit.claim?.cards ?? []).map(card => card.oracleId);
      for (const gap of unit.result?.gaps ?? []) add(issueNumber(gap.id), oracleIds, `${source} card ${unit.task.id}`);
    }
  }
  for (const pending of next?.pending_prerequisites ?? []) {
    for (const issue of pending.issues ?? []) add(issue, pending.blocked_oracle_ids, pending.batch ? `queue batch ${pending.batch}` : "queue");
  }
  return [...gaps.values()];
}

// A gap is filled once a landed Engine unit targeted it after it was found.
// Issues often stay open for broader work, so issue state alone is not enough.
export function filledIssues(units = [], priorUnits = []) {
  return new Set([...priorUnits, ...units].filter(u => u.status === "landed").map(engineIssue).filter(Boolean));
}

// A gap whose Engine unit failed (this run or an earlier one) is parked rather
// than paid for again, unless the caller explicitly lists it in retry.
export function chooseWork({ next, units = [], priorUnits = [], issues = {}, retry = [] }) {
  const filled = filledIssues(units, priorUnits);
  const attempted = new Set([
    ...units.map(engineIssue),
    ...priorUnits.filter(unit => unit.status !== "landed").map(engineIssue).filter(issue => !retry.includes(issue)),
  ].filter(Boolean));
  const gaps = knownGaps({ next, units, priorUnits });
  const open = gaps.filter(gap => !filled.has(gap.issue) && issues[gap.issue]?.state === "OPEN");
  const gap = open.find(candidate => !attempted.has(candidate.issue));
  if (gap) return { kind: "engine", gap, issue: issues[gap.issue] };
  if (next?.batch) return { kind: "card", batch: next.batch, cards: next.cards ?? [] };
  return {
    kind: "prepare",
    avoid: open,
    unblocked: gaps.filter(candidate => candidate.oracleIds.length &&
      (filled.has(candidate.issue) || issues[candidate.issue]?.state === "CLOSED")),
    parked: open.filter(candidate => attempted.has(candidate.issue)),
  };
}

export function gapTask(choice, names) {
  const { gap, issue } = choice;
  const cards = gap.oracleIds.map(id => names[id] ? `${names[id]} (${id})` : id);
  return {
    taskType: "engine", id: `gap-${gap.issue}`, issue: gap.issue,
    title: `Close Engine gap #${gap.issue}: ${issue.title}`.slice(0, 200),
    reviewRequired: true, batchPlan: "",
    instructions: `Known Engine gap #${gap.issue} (${issue.url}) blocks card work: ${cards.join(", ") || "see issue"}.
Evidence sources: ${gap.sources.join("; ")}. Read the issue, its comments and retained blocked-card probes listed in earlier unit evidence.
Implement the smallest coherent shared Engine capability that unblocks these cards, with proving rules scenarios. If the issue is broader, deliver a bounded slice that covers these cards and comment what remains. Do not admit cards; a later card unit does that.
If the gap cannot be closed in one bounded unit, return blocked with the precise missing design rather than a partial or approximate rule.`,
  };
}

export function batchTask(choice) {
  const names = choice.cards.map(card => card.name).join(", ");
  return {
    taskType: "card", id: choice.batch.id, batchPlan: "", reviewRequired: true,
    title: `Author ${names || choice.batch.id}`.slice(0, 200),
    instructions: `Implement saved ready batch ${choice.batch.id} (${choice.batch.family}): ${choice.cards.map(card => `${card.name} oracle ${card.oracle_id} revision ${card.revision}`).join("; ")}. Planned operations: ${(choice.batch.operations ?? []).join("; ")}. Prerequisites: ${(choice.batch.prerequisites ?? []).map(issue => `#${issue}`).join(", ")}.`,
  };
}
