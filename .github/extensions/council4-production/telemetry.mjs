import { appendFileSync, existsSync, readFileSync, writeFileSync, renameSync } from "node:fs";
import { randomUUID } from "node:crypto";

export function writeJSON(destination, value) {
  const temporary = `${destination}.${randomUUID()}.tmp`;
  writeFileSync(temporary, JSON.stringify(value, null, 2) + "\n");
  renameSync(temporary, destination);
}

// Persist only attribution and billing facts, never prompts, responses or tokens'
// contents. Sequence identity deduplicates live and replayed usage receipts.
export function observeUsage(session, runId, file, onError) {
  const agents = new Set();
  const keys = new Set();
  if (existsSync(file)) {
    for (const row of readReceipts(file)) {
      if (row.kind === "agent") agents.add(row.agentId);
      if (row.kind === "usage") keys.add(row.key);
    }
  }
  return session.on(event => {
    try {
      if (event.type === "subagent.started" &&
          (event.data?.workflowRunId === runId || agents.has(event.data?.parentId))) {
        if (!event.agentId) throw new Error("Workflow agent event has no agent ID");
        agents.add(event.agentId);
        appendFileSync(file, JSON.stringify({
          kind: "agent", runId, agentId: event.agentId, parentId: event.data.parentId ?? null,
        }) + "\n");
      }
      if (!agents.has(event.agentId) ||
          !["assistant.usage", "session.usage_record"].includes(event.type)) return;
      const usage = event.data?.usage ?? event.data;
      const identity = usage.accounting;
      const key = identity
        ? `${identity.sourceSessionId}:${identity.sequence}`
        : event.id;
      if (!key) throw new Error("Usage event has no deduplication identity");
      if (keys.has(key)) return;
      keys.add(key);
      const nanoAiu = usage.copilotUsage?.totalNanoAiu;
      appendFileSync(file, JSON.stringify({
        kind: "usage", runId, key, agentId: event.agentId, at: event.timestamp,
        model: usage.model ?? null, reasoningEffort: usage.reasoningEffort ?? null,
        nanoAiu: Number.isSafeInteger(nanoAiu) && nanoAiu >= 0 ? nanoAiu : null,
        inputTokens: usage.inputTokens ?? null, outputTokens: usage.outputTokens ?? null,
        cacheReadTokens: usage.cacheReadTokens ?? null, cacheWriteTokens: usage.cacheWriteTokens ?? null,
      }) + "\n");
    } catch (error) {
      onError(error);
    }
  });
}

export function readReceipts(file) {
  return existsSync(file) ? readFileSync(file, "utf8").split("\n").filter(Boolean).map(JSON.parse) : [];
}

export function summarize(detail, stages, receipts) {
  const parents = new Map(receipts.filter(r => r.kind === "agent").map(r => [r.agentId, r.parentId]));
  const byAgent = new Map(detail.agents.map(a => [a.agentId, a]));
  const byLabel = new Map(stages.map(stage => [stage.label, stage]));
  const assigned = new Map();
  const seen = new Set();
  let observedNanoAiu = 0;
  let missingBillingReceipts = 0;
  for (const receipt of receipts) {
    if (receipt.kind !== "usage" || seen.has(receipt.key)) continue;
    seen.add(receipt.key);
    let id = receipt.agentId;
    const ancestry = new Set();
    while (!byAgent.has(id) && parents.has(id) && !ancestry.has(id)) {
      ancestry.add(id);
      id = parents.get(id);
    }
    const agent = byAgent.get(id);
    const stage = agent && byLabel.get(agent.label);
    if (receipt.nanoAiu === null) missingBillingReceipts++;
    else observedNanoAiu += receipt.nanoAiu;
    if (!stage) continue;
    const key = JSON.stringify([stage.taskType, stage.taskId, stage.unitIndex ?? null, stage.phase, receipt.model ?? agent.resolvedModel ?? null]);
    const group = assigned.get(key) ?? {
      taskType: stage.taskType, taskId: stage.taskId, phase: stage.phase,
      unitIndex: stage.unitIndex ?? null,
      model: receipt.model ?? agent.resolvedModel ?? null,
      observedNanoAiu: 0, calls: 0, missingBillingReceipts: 0,
    };
    group.calls++;
    if (receipt.nanoAiu === null) group.missingBillingReceipts++;
    else group.observedNanoAiu += receipt.nanoAiu;
    assigned.set(key, group);
  }
  const agents = detail.agents.map(agent => ({
    agentId: agent.agentId, label: agent.label, status: agent.status,
    requestedModel: agent.requestedModel ?? null, resolvedModel: agent.resolvedModel ?? null,
    activeMs: agent.activeMs, startedAt: agent.startedAt ?? null, completedAt: agent.completedAt ?? null,
    attribution: byLabel.get(agent.label) ?? null,
  }));
  for (const agent of agents) {
    const stage = agent.attribution;
    if (!stage) continue;
    const key = JSON.stringify([stage.taskType, stage.taskId, stage.unitIndex ?? null, stage.phase, agent.resolvedModel]);
    const group = assigned.get(key) ?? {
      taskType: stage.taskType, taskId: stage.taskId, phase: stage.phase, model: agent.resolvedModel,
      unitIndex: stage.unitIndex ?? null,
      observedNanoAiu: 0, calls: 0, missingBillingReceipts: 0,
    };
    group.activeMs = (group.activeMs ?? 0) + agent.activeMs;
    group.agents = (group.agents ?? 0) + 1;
    assigned.set(key, group);
  }
  const allocated = [...assigned.values()].reduce((sum, g) => sum + g.observedNanoAiu, 0);
  const runtimeNanoAiu = detail.consumed.nanoAiu;
  const missingAgentUsage = agents.some(agent => !receipts.some(r => r.kind === "usage" && r.agentId === agent.agentId));
  const matched = runtimeNanoAiu === allocated && missingBillingReceipts === 0 && !missingAgentUsage;
  return {
    runId: detail.runId, workflowName: detail.workflowName, runtimeStatus: detail.status,
    observedAt: detail.observedAt,
    runtimeActiveMs: detail.consumed.activeMs, runtimeNanoAiu, runtimeAic: runtimeNanoAiu / 1e9,
    observedNanoAiu, unallocatedNanoAiu: runtimeNanoAiu - allocated,
    reconciliation: matched ? "matched" : "incomplete",
    missingBillingReceipts,
    groups: [...assigned.values()].map(group => ({
      ...group, observedAic: group.observedNanoAiu / 1e9,
      complete: matched,
    })),
    agents, stages,
    note: "AIC = nano-AIU / 1e9, not USD. Missing/delayed attribution is unallocated, not zero. Agent active time differs from stage elapsed time and run wall time. Outer chat, CI and provider-independent compute are outside native workflow AIC.",
  };
}

export function writeReport(directory, report) {
  const destination = `${directory}/report.json`;
  writeJSON(destination, report);
  const cells = value => `"${String(value ?? "").replaceAll('"', '""')}"`;
  const columns = ["unitIndex", "taskType", "taskId", "phase", "model", "activeMs", "calls", "observedAic", "complete"];
  const csv = [columns.join(","), ...report.groups.map(row => columns.map(k => cells(row[k])).join(","))];
  writeFileSync(`${directory}/costs.csv`, csv.join("\n") + "\n");
  const cell = value => String(value ?? "unknown").replaceAll("|", "\\|").replaceAll("\n", " ");
  const markdown = [
    "# Council4 production run",
    "",
    `Run: \`${report.runId}\` · policy: \`${report.policyVersion}\` · status: **${report.runtimeStatus}**`,
    "",
    `Native total: **${report.runtimeAic.toFixed(6)} AIC**. Accounting: **${report.reconciliation}**. Unallocated: **${(report.unallocatedNanoAiu / 1e9).toFixed(6)} AIC**.`,
    "",
    "| Unit | Task type | Task | Stage | Actual model | Agent active seconds | Observed AIC | Complete accounting |",
    "| --- | --- | --- | --- | --- | ---: | ---: | --- |",
    ...report.groups.map(row => `| ${[row.unitIndex ?? "unrecorded", row.taskType, row.taskId, row.phase, row.model,
      row.activeMs == null ? "unknown" : (row.activeMs / 1000).toFixed(2),
      row.observedAic.toFixed(6), row.complete ? "yes" : "no"].map(cell).join(" | ")} |`),
    "",
    "## Outcomes",
    "",
    ...(report.units ?? []).map(unit => `- ${cell(unit.task?.taskType)} / ${cell(unit.task?.id)}: **${cell(unit.status)}**${unit.landing?.commit ? ` (\`${unit.landing.commit}\` on origin/main)` : ""}`),
    "",
    report.error ? `**Run error:** ${cell(report.error.message)}` : "",
    report.note,
  ];
  writeFileSync(`${directory}/report.md`, markdown.join("\n") + "\n");
}
