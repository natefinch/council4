import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

// prerequisiteEvidence maps a filled gap issue to {runtime_revision, evidence};
// plan members still blocked on that issue are reassessed before batching.
export async function prepareCardBatch({ workPath, task, corpus, owner, planPath, command, signal, prerequisiteEvidence = {} }) {
  const manifest = path.join(workPath, "testdata/card-authoring/production.json");
  const run = (operation, args = []) => command(workPath, "go", ["run", "./cmd/cardqueue",
    "-operation", operation, "-manifest", manifest, "-owner", owner, "-corpus", corpus, ...args], signal);
  const queue = JSON.parse(await readFile(manifest, "utf8"));
  const saved = queue.batches.find(batch => batch.id === task.id);
  if (!saved) {
    if (!task.batchPlan?.trim()) throw new Error("Card selection needs a saved batch or an explicit reviewed batch plan");
    const plan = JSON.parse(task.batchPlan);
    if (!plan || plan.corpus_sha256 !== queue.corpus_sha256 ||
        !Array.isArray(plan.batches) || plan.batches.length !== 1 || plan.batches[0].id !== task.id) {
      throw new Error("New card selection has an invalid or mismatched reviewed batch plan");
    }
    await mkdir(path.dirname(planPath), { recursive: true });
    const current = new Map(queue.records.map(record => [record.oracle_id, record]));
    const blocked = (plan.records ?? []).map(record => current.get(record.oracle_id))
      .filter(record => record?.status === "blocked");
    if (blocked.length) {
      const records = blocked.map(record => {
        const evidence = prerequisiteEvidence[record.blocker];
        if (!evidence) throw new Error(`${record.name} is still blocked by #${record.blocker}; that gap has not been filled`);
        return { oracle_id: record.oracle_id, revision: record.revision, name: record.name,
          reassessment: { issue: record.blocker, ...evidence } };
      });
      const reassessPath = planPath.replace(/\.json$/, "-reassess.json");
      await writeFile(reassessPath, JSON.stringify({ corpus_sha256: queue.corpus_sha256, records }, null, 2) + "\n");
      await run("reassess", ["-plan", reassessPath]);
    }
    await writeFile(planPath, JSON.stringify(plan, null, 2) + "\n");
    // The existing CLI validates exact identity, prerequisites and ownership;
    // this helper never independently promotes a candidate as ready.
    await run("add-reviewed-batch", ["-batch", task.id, "-plan", planPath]);
  } else if (task.batchPlan?.trim()) {
    throw new Error("A reviewed plan cannot rewrite an existing saved batch");
  }
  const next = JSON.parse(await run("next"));
  if (next.batch?.id !== task.id) throw new Error("Selected card family is not the next executable saved batch");
  await run("claim", ["-batch", task.id]);
  const resumed = JSON.parse(await run("next"));
  if (resumed.batch?.id !== task.id || !resumed.cards?.length ||
      resumed.cards.some(card => card.status !== "claimed" || card.owner !== owner)) {
    throw new Error("Coordinator could not verify the durable whole-batch claim");
  }
  return { batchId: task.id, owner, cards: resumed.cards.map(card => ({
    oracleId: card.oracle_id, revision: card.revision, name: card.name,
  })) };
}

// blocked lists members the worker left unregistered, each on its canonical
// Engine gap: [{ oracleId, gap, reason }]. The rest of the batch is admitted
// and those members are blocked, so finished cards land without waiting.
export async function verifyCardBatch({ workPath, batchId, corpus, owner, command, signal, blocked = [], planPath }) {
  const manifest = path.join(workPath, "testdata/card-authoring/production.json");
  const gaps = new Map(blocked.map(card => [card.oracleId, card]));
  const readMembers = async () => {
    const queue = JSON.parse(await readFile(manifest, "utf8"));
    const batch = queue.batches.find(candidate => candidate.id === batchId);
    if (!batch) throw new Error("Claimed batch disappeared");
    const members = batch.oracle_ids.map(id => queue.records.find(card => card.oracle_id === id));
    if (members.some(card => !card || card.batch !== batchId)) throw new Error("Claimed batch membership changed");
    return { queue, members };
  };
  // Settling removes blocked cards from the batch; they stay blocked on their gap.
  const settledOut = queue => [...gaps.values()].every(({ oracleId, gap }) => queue.records.some(card =>
    card.oracle_id === oracleId && card.status === "blocked" && card.blocker === gap && !card.batch && !card.owner));
  const admitted = members => members.length && members.every(card =>
    !gaps.has(card.oracle_id) && card.status === "implemented-verified" && !card.owner && card.evidence);
  const run = (operation, args = []) => command(workPath, "go", ["run", "./cmd/cardqueue", "-operation", operation,
    "-manifest", manifest, "-owner", owner, "-corpus", corpus, "-batch", batchId, ...args], signal);
  const { queue, members } = await readMembers();
  if (members.every(card => card.status === "claimed" && card.owner === owner)) {
    if ([...gaps.keys()].some(id => !members.some(card => card.oracle_id === id))) {
      throw new Error("Blocked cards must be members of the claimed batch");
    }
    if (gaps.size) {
      const file = planPath ?? path.join(workPath, ".cardwork", `settle-${batchId}.json`);
      await mkdir(path.dirname(file), { recursive: true });
      await writeFile(file, JSON.stringify({ corpus_sha256: queue.corpus_sha256, records: members
        .filter(card => gaps.has(card.oracle_id)).map(card => ({ oracle_id: card.oracle_id, revision: card.revision,
          name: card.name, blocker: gaps.get(card.oracle_id).gap, reason: gaps.get(card.oracle_id).reason })) }, null, 2) + "\n");
      await run("settle", ["-plan", file]);
    } else {
      await run("verify");
    }
  } else if (admitted(members) && settledOut(queue)) {
    await run("check");
  } else {
    throw new Error("Workers changed queue ownership or partially promoted the card batch");
  }
  const after = await readMembers();
  if (!admitted(after.members) || !settledOut(after.queue)) {
    throw new Error("Completed card family has not passed checked registry/source admission");
  }
  return { batchId, verified: after.members.map(card => card.oracle_id), blocked: [...gaps.keys()] };
}
