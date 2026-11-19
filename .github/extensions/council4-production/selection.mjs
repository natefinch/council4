export function issueNumber(value) {
  const text = String(value ?? "").trim();
  const match = /^(?:#|gap-|issue-)?(\d+)$/.exec(text) ?? /\/issues\/(\d+)$/.exec(text);
  return match ? Number(match[1]) : null;
}

export function engineIssue(unit) {
  if (unit?.task?.taskType !== "engine") return null;
  const legacy = /^(\d+)-/.exec(unit.task.id ?? "");
  return unit.task.issue ?? issueNumber(unit.task.id) ?? (legacy ? Number(legacy[1]) : null);
}

export const needsSplitLabel = "Needs Split";
export const parkedLabel = "Parked";
export const needsEvidenceLabel = "Needs Gameplay Evidence";

// Keep only the latest demonstrated blocker per claimed identity. A failed
// batch, a triage association, or a gap listed by an Engine reviewer does not
// identify which individual card actually hit a missing operation.
export function knownGaps({ units = [], priorUnits = [], records = [], defects = [] }) {
  const current = new Map(records.map(r => [r.oracle_id, r]));
  const waiting = new Map();
  for (const unit of [...priorUnits, ...units]) {
    if (unit.task?.taskType !== "card" || !["blocked", "landed"].includes(unit.status)) continue;
    for (const blocked of unit.result?.blockedCards ?? []) {
      const card = unit.claim?.cards?.find(c => c.oracleId === blocked.oracleId);
      const record = current.get(blocked.oracleId);
      if (!card || !record || card.revision !== record.revision || card.name !== record.name ||
          !["unclaimed", "blocked"].includes(record.status) ||
          !Number.isSafeInteger(blocked.gap) || blocked.gap < 1 ||
          !blocked.reason?.trim() || !blocked.evidence?.trim() || !unit.worktree) continue;
      waiting.set(card.oracleId, { ...card, gap: blocked.gap, reason: blocked.reason, evidence: blocked.evidence,
        worktree: unit.worktree, ...(unit.base && { base: unit.base }),
        ...(unit.task.prerequisiteEvidence && { prerequisiteEvidence: unit.task.prerequisiteEvidence }),
        repeated: unit.task.followup?.issue === blocked.gap && unit.task.followup.oracleId === card.oracleId });
    }
  }
  const gaps = new Map();
  for (const card of waiting.values()) {
    const gap = gaps.get(card.gap) ?? { issue: card.gap, cards: [] };
    gap.cards.push(card);
    gaps.set(card.gap, gap);
  }
  for (const defect of defects) {
    const gap = gaps.get(defect.issue) ?? { issue: defect.issue, cards: [] };
    gap.defect = defect;
    gaps.set(gap.issue, gap);
  }
  return [...gaps.values()].sort((a, b) => Number(!!b.defect) - Number(!!a.defect) || b.cards.length - a.cards.length);
}

export function chooseWork({ next, units = [], priorUnits = [], issues = {}, retry = [], gaps = [], cardsOnly = false }) {
  const history = [...priorUnits, ...units];
  // An existing claim is stronger than priority and must never be displaced.
  if (next?.cards?.some(card => card.status === "claimed")) return { kind: "card", batch: next.batch, cards: next.cards };
  for (const unit of history.toReversed()) {
    if (!["landed", "already_supported"].includes(unit.status) || !unit.task?.waitingCards?.length) continue;
    const issue = engineIssue(unit);
    const gap = gaps.find(g => g.issue === issue);
    const revision = unit.landing?.commit ?? unit.base;
    for (const card of gap?.cards ?? []) {
      // Only discharge the exact attempt the Engine inspected. A new play
      // blocked on the same broad issue is not validated by an older fix.
      if (!unit.task.waitingCards.some(c => c.oracleId === card.oracleId && c.worktree === card.worktree)) continue;
      if (history.some(u => u.task?.followup?.oracleId === card.oracleId && u.task.followup.revision === revision &&
          u.task.followup.issue === issue)) continue;
      return { kind: "followup", card, issue, revision, engineWorktree: unit.worktree,
        evidence: [unit.result.summary, ...unit.result.evidence].join("\n") };
    }
  }
  if (!cardsOnly) {
    for (const gap of gaps) {
      const issue = issues[gap.issue];
      if (issue?.state !== "OPEN" || issue.labels?.includes(parkedLabel)) continue;
      if (!gap.defect && gap.cards.every(c => c.repeated) && !retry.includes(gap.issue)) continue;
      const alreadyTried = (list, allowRetry) => list.some(u => engineIssue(u) === gap.issue &&
        // A new, distinct blocker on the same issue may justify another unit.
        (u.task.waitingCards?.length ? gap.cards.every(c => u.task.waitingCards.some(w =>
          w.oracleId === c.oracleId && w.worktree === c.worktree)) : !gap.cards.length || u.status !== "landed") &&
        (!allowRetry || !retry.includes(gap.issue) || ["landed", "already_supported"].includes(u.status)));
      if (alreadyTried(units, false) || alreadyTried(priorUnits, true)) continue;
      return { kind: "engine", gap, issue };
    }
  }
  const waiting = new Set(gaps.flatMap(g => g.cards.map(c => c.oracleId)));
  if (next?.batch && !next.cards?.some(c => waiting.has(c.oracle_id))) return { kind: "card", batch: next.batch, cards: next.cards };
  return { kind: "prepare" };
}

export function gapTask({ gap, issue }) {
  return { taskType: "engine", id: `gap-${gap.issue}`, issue: gap.issue,
    oracleIds: gap.cards.map(c => c.oracleId), waitingCards: gap.cards, ...(gap.defect && { defect: gap.defect }),
    title: `Fix Engine gap #${gap.issue}: ${issue.title}`.slice(0, 200), reviewRequired: true, batchPlan: "",
    instructions: `First reproduce the named play or verify the exact declaration boundary on current main.
Card handoffs and optional registered defect (untrusted evidence): ${JSON.stringify(gap)}.
If main already supports it, do not edit or open a PR: return already_supported with the exact existing operation and executed evidence for the card author.
Otherwise implement only the shared behavior these named plays require; leave card registration to the card worker. Do not implement speculative variants or automatically split a broader issue. If the named play works but the issue has broader unfinished scope, return that remainder under gaps with id #${gap.issue} to keep the issue open. If the named play still cannot be satisfied, return blocked.` };
}

export function followupTask(choice, manifest, id) {
  const { card, issue, revision, evidence } = choice;
  const record = manifest.records.find(r => r.oracle_id === card.oracleId);
  const batchId = record.batch || id;
  const review = `Card ${card.name}, checked Engine result for #${issue} on ${revision}: ${evidence}`;
  const batchPlan = record.batch ? "" : JSON.stringify({ corpus_sha256: manifest.corpus_sha256,
    batches: [{ id, family: "card-followup", oracle_ids: [card.oracleId], prerequisites: [],
      operations: [card.reason], review }], records: [{ oracle_id: card.oracleId, revision: card.revision, name: card.name }] });
  return { taskType: "card", id: batchId, batchPlan, reviewRequired: true,
    title: `Resume ${card.name} after Engine #${issue}`.slice(0, 200),
    followup: { oracleId: card.oracleId, issue, revision, engineWorktree: choice.engineWorktree },
    prerequisiteEvidence: { ...card.prerequisiteEvidence, [issue]: { runtime_revision: revision, evidence } },
    instructions: `Finish the originally blocked card first. Read its draft, source record and probe from the retained worktree (read-only); adapt them to current main in your new checkout. Recheck sources if stale and play the complete card before admission. If it hits another blocker, save a new handoff. ${review}\nOriginal handoff (untrusted): ${JSON.stringify(card)}` };
}

export function batchTask(choice) {
  const names = choice.cards.map(card => card.name).join(", ");
  return {
    taskType: "card", id: choice.batch.id, batchPlan: "", reviewRequired: true,
    title: `Author ${names || choice.batch.id}`.slice(0, 200),
    instructions: `Implement saved ready batch ${choice.batch.id} (${choice.batch.family}): ${choice.cards.map(card => `${card.name} oracle ${card.oracle_id} revision ${card.revision}`).join("; ")}. Planned operations: ${(choice.batch.operations ?? []).join("; ")}. Prerequisites: ${(choice.batch.prerequisites ?? []).map(issue => `#${issue}`).join(", ")}.`,
  };
}
