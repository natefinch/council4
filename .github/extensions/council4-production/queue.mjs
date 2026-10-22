import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

export async function prepareCardBatch({ workPath, task, corpus, owner, planPath, command, signal }) {
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

export async function verifyCardBatch({ workPath, batchId, corpus, owner, command, signal }) {
  const manifest = path.join(workPath, "testdata/card-authoring/production.json");
  const readMembers = async () => {
    const queue = JSON.parse(await readFile(manifest, "utf8"));
    const batch = queue.batches.find(candidate => candidate.id === batchId);
    if (!batch) throw new Error("Claimed batch disappeared");
    const members = batch.oracle_ids.map(id => queue.records.find(card => card.oracle_id === id));
    if (members.some(card => !card || card.batch !== batchId)) throw new Error("Claimed batch membership changed");
    return members;
  };
  const members = await readMembers();
  let operation;
  if (members.every(card => card.status === "claimed" && card.owner === owner)) operation = "verify";
  else if (members.every(card => card.status === "implemented-verified" && !card.owner)) operation = "check";
  else throw new Error("Workers changed queue ownership or partially promoted the card batch");
  await command(workPath, "go", ["run", "./cmd/cardqueue", "-operation", operation,
    "-manifest", manifest, "-owner", owner, "-corpus", corpus, "-batch", batchId], signal);
  const admitted = await readMembers();
  if (admitted.some(card => card.status !== "implemented-verified" || card.owner || !card.evidence)) {
    throw new Error("Completed card family has not passed checked registry/source admission");
  }
  return { batchId, verified: admitted.map(card => card.oracle_id) };
}
