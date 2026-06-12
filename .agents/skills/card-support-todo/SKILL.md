---
name: card-support-todo
description: Quickly file a traceable GitHub issue labeled "To Be Triaged" and "Card Support TODO" when card implementation or compiler work discovers known work that is intentionally deferred. Use when the user says "file a card support TODO", "track this card limitation", "create an issue for this deferred work", or whenever card support work deliberately leaves a known gap behind.
---

# Card Support TODO

Quickly preserve deferred card-support work without interrupting implementation.

## Process

1. Capture only the minimum useful context:
   - what card, wording family, mechanic, or compiler limitation was found;
   - why it is being deferred;
   - the obvious completion condition, if known;
   - the active issue where the deferred work was discovered;
   - the active epic, if the issue is an epic child.
2. Search briefly for an existing issue with the same specific scope:

   ```bash
   gh issue list --state all --search '"<specific key terms>"' \
     --limit 10 --json number,title,url
   ```

3. If no matching issue exists, create one immediately:

   ```bash
   gh issue create \
     --title "<short actionable title>" \
     --label "Card Support TODO" \
     --label "To Be Triaged" \
     --body $'## Deferred work\n<one or two sentences>\n\n## Discovered while working on\n- #<active-issue-number>\n\n## Done when\n<one sentence>'
   ```

4. If the active issue is an epic child, add the new issue link to the active
   child's subsection under `## Future Work Discovered` in the epic body. Do not
   make the new issue an epic child.
5. Return the issue URL and resume the original work. Do not turn issue filing
   into a research task.

## Speed Rules

- Spend no more than a couple of minutes filing the issue.
- Prefer the short required body sections over a comprehensive design.
- Include file names, card examples, or counts only when already known.
- Do not investigate solutions merely to improve the issue description.
- Do not block current work after the issue URL is available.

## Boundaries

**Will:**

- Use the current repository through `gh`.
- Apply the exact `To Be Triaged` and `Card Support TODO` labels.
- File intentionally deferred card-support work.
- Reuse an obvious existing issue instead of duplicating it.
- Link new issues to the issue where the work was discovered.
- Record new issues in the active epic without making them epic children.

**Will not:**

- File speculative ideas with no known work attached.
- create long plans, implementation designs, or exhaustive card inventories;
- silently defer discovered work without either fixing it or filing an issue.

## Error Handling

- If either label is missing, create it once:

  ```bash
  gh label create "Card Support TODO" \
    --description "Known work that still needs to be done to support a corpus of cards." \
    --color "BCEA7A"
  gh label create "To Be Triaged" \
    --description "Newly discovered work awaiting triage." \
    --color "FBCA04"
  ```

- If `gh` is unauthenticated or issue creation fails, report the command error
  and continue the original implementation without pretending the TODO was
  filed.
