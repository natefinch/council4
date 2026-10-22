import { defineWorkflow, joinSession } from "@github/copilot-sdk/extension";
import { meta, runProduction, reportMeta, runReport } from "./workflow.mjs";

await joinSession({
  workflows: [
    defineWorkflow({ meta, run: runProduction }),
    defineWorkflow({ meta: reportMeta, run: runReport }),
  ],
});
