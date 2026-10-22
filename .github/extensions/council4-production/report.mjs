import path from "node:path";
import { collectReports } from "./git.mjs";

const reports = await collectReports(path.resolve(process.argv[2] ?? ".cardwork/workflows"));
console.log(JSON.stringify(reports.map(report => ({
  runId: report.runId, revision: report.revision, policyVersion: report.policyVersion,
  status: report.runtimeStatus, runtimeAic: report.runtimeAic, runtimeActiveMs: report.runtimeActiveMs,
  reconciliation: report.reconciliation, unallocatedNanoAiu: report.unallocatedNanoAiu,
  groups: report.groups, agents: report.agents, stages: report.stages, units: report.units,
  currentUnit: report.currentUnit, error: report.error,
})), null, 2));
