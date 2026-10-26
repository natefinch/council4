import { defineWorkflow, joinSession } from "@github/copilot-sdk/extension";
import { workflows } from "./catalog.mjs";

await joinSession({ workflows: workflows.map(({ meta, run }) => defineWorkflow({ meta, run })) });
