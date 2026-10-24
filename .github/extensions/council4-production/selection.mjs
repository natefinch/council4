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

// Engine units may land one slice of a large gap and report the remainder as
// gaps naming the same (or a new sub-) issue.
function remainingIssues(unit) {
  return (unit.result?.gaps ?? []).map(gap => issueNumber(gap.id)).filter(Boolean);
}

// Known gaps, in priority order: those demonstrated by blocked card work and
// remainders of landed Engine slices (this run first, then earlier local runs),
// then saved queue prerequisites.
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
      if (unit?.task?.taskType === "engine" && unit.status === "landed") {
        for (const issue of remainingIssues(unit)) add(issue, unit.task.oracleIds, `${source} remainder of ${unit.task.id}`);
      }
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

// A gap is filled once a landed Engine unit targeted it and reported no
// remainder for it. Issues often stay open for broader work, so issue state
// alone is not enough.
export function filledIssues(units = [], priorUnits = []) {
  return new Set([...priorUnits, ...units]
    .filter(u => u.status === "landed" && !remainingIssues(u).includes(engineIssue(u)))
    .map(engineIssue).filter(Boolean));
}

// A gap whose Engine unit failed (this run or an earlier one) is parked rather
// than paid for again, unless the caller explicitly lists it in retry. A landed
// slice with a reported remainder stays eligible so the next slice follows.
export function chooseWork({ next, units = [], priorUnits = [], issues = {}, retry = [] }) {
  const filled = filledIssues(units, priorUnits);
  const failed = unit => unit.status !== "landed";
  const attempted = new Set([
    ...units.filter(failed).map(engineIssue),
    ...priorUnits.filter(failed).map(engineIssue).filter(issue => !retry.includes(issue)),
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
    taskType: "engine", id: `gap-${gap.issue}`, issue: gap.issue, oracleIds: gap.oracleIds,
    title: `Close Engine gap #${gap.issue}: ${issue.title}`.slice(0, 200),
    reviewRequired: true, batchPlan: "",
    instructions: `Known Engine gap #${gap.issue} (${issue.url}) blocks card work: ${cards.join(", ") || "see issue"}.
Evidence sources: ${gap.sources.join("; ")}. Read the issue, its comments and retained blocked-card probes listed in earlier unit evidence.
Deliver working shared Engine code with proving rules scenarios; a design writeup or failing probes alone is not a result. If it touches spell, cost or stack-object state, cover every applicable route in the spell-route checklist in docs/card-authoring.md from the start. Do not admit cards; a later card unit does that.
If the whole capability is too large for one unit, split it: implement and prove the first shared piece that is correct and useful on its own (a complete rule for the situations it handles, not an approximation), then return status complete with each remaining piece as a gap whose id is #${gap.issue} or a new sub-issue you file and link from #${gap.issue}. The coordinator schedules the next piece. Return blocked only when no piece can land without an unresolved design decision, and name that decision.`,
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
