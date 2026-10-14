---
name: card-support-todo
description: Reuse or file one precise shared blocker for proven deferred card behavior. Use when asked to track a card limitation or when authored-card work encounters a missing Engine capability.
---

# Card Support TODO

Keep complete cards supported and incomplete cards explicitly blocked.
Follow `.github/copilot-instructions.md`; new support is authored Go, not
Oracle compiler expansion.

1. Capture the demonstrated missing behavior, representative card(s), owning
   rule/operation if known, originating issue and observable completion condition.
2. Briefly search existing issues. Reuse the canonical capability issue even
   when several cards need it; link the affected work's blocking dependency.
3. If distinct work is proven, create one concise issue using the available
   issue-creation tool with `To Be Triaged` and `Card Support TODO`.
   Use `Legacy Compiler` for an explicitly requested compiler maintenance defect.
4. Leave the affected card unsupported whole; remove stale readiness and
   explain its blocker once. Resume other eligible assigned work.

Do not enumerate speculative neighboring wording, create one issue per
unexamined card, perform lengthy solution research just to file a blocker, or
mirror the same note into an epic's future-work subsections. Add an epic link
only when it changes execution/dependencies. Follow the user's scope rather than
adding discovered work to the current implementation automatically.

## Minimal body

```markdown
## Missing behavior
<actual limitation and representative card>

## Discovered while
<link to the active issue>

## Done when
<observable behavior and relevant rules scenario>
```

Report failures explicitly. Never pretend a blocker was filed or that an
unsupported card is usable when the issue operation or implementation failed.
