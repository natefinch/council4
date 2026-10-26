import { meta, runProduction, reportMeta, runReport } from "./workflow.mjs";
import { fileURLToPath } from "node:url";
import { triageMeta, runTriage } from "./triage.mjs";

// The workflows this extension registers, with example arguments for starting
// them. The production dashboard lists these so a run can be copied into a new
// session. Placeholders in angle brackets must be replaced before running.
export const workflows = [
  {
    meta, run: runProduction,
    examples: [
      { title: "Produce", args: { mode: "produce", maxTasks: 20 } },
      { title: "Preflight only", args: { mode: "preflight" } },
      { title: "Resume retained units", args: { mode: "produce", resumeUnits: [{ runId: "<run id>", index: 0 }] } },
      { title: "Retry parked gaps", args: { mode: "produce", maxTasks: 5, retryGaps: ["<issue number>"] } },
    ],
  },
  {
    meta: reportMeta, run: runReport,
    examples: [{ title: "Refresh a run's cost report", args: { runId: "<run id>" } }],
  },
  {
    meta: triageMeta, run: runTriage,
    examples: [
      { title: "Triage the backlog (read-only)", args: {} },
      { title: "Triage selected issues", args: { issues: ["<issue number>"] } },
      { title: "Apply a reviewed triage run", args: { apply: true, fromRunId: "<run id>" } },
    ],
  },
];

// The text to paste into a Copilot session to start a workflow.
// Issue-number placeholders are shown unquoted because they become numbers.
export const prompt = (name, args) =>
  `Run the \`${name}\` workflow with \`${JSON.stringify(args).replaceAll('"<issue number>"', "<issue number>")}\``;

// What the dashboard shows: names, descriptions, argument types and examples.
export function describe() {
  return workflows.map(({ meta: m, examples }) => ({
    name: m.name,
    description: m.description,
    args: Object.entries(m.argsSchema?.properties ?? {}).map(([name, schema]) => ({
      name,
      type: schema.enum ? schema.enum.join(" | ") : schema.type === "array" ? `${schema.items?.type ?? "value"}[]` : schema.type,
      required: (m.argsSchema.required ?? []).includes(name),
    })),
    examples: examples.map(example => ({ ...example, prompt: prompt(m.name, example.args) })),
  }));
}

if (process.argv[1] === fileURLToPath(import.meta.url)) console.log(JSON.stringify(describe()));
