---
name: card-support-todo
description: Quickly file a GitHub issue labeled "Card Support TODO" when card implementation or compiler work discovers known work that is intentionally deferred. Use when the user says "file a card support TODO", "track this card limitation", "create an issue for this deferred work", or whenever card support work deliberately leaves a known gap behind.
---

# Card Support TODO

Quickly preserve deferred card-support work without interrupting implementation.

## Process

1. Capture only the minimum useful context:
   - what card, wording family, mechanic, or compiler limitation was found;
   - why it is being deferred;
   - the obvious completion condition, if known.
2. Search briefly for an existing open issue with the same specific title:

   ```bash
   gh issue list --state open --search 'in:title "<short title>"' \
     --limit 10 --json number,title,url
   ```

3. If no matching issue exists, create one immediately:

   ```bash
   gh issue create \
     --title "<short actionable title>" \
     --label "Card Support TODO" \
     --body $'## Deferred work\n<one or two sentences>\n\n## Done when\n<one sentence>'
   ```

4. Return the issue URL and resume the original work. Do not turn issue filing
   into a research task.

## Speed Rules

- Spend no more than a couple of minutes filing the issue.
- Prefer a clear two-section body over a comprehensive design.
- Include file names, card examples, or counts only when already known.
- Do not investigate solutions merely to improve the issue description.
- Do not block current work after the issue URL is available.

## Boundaries

**Will:**

- Use the current repository through `gh`.
- Apply the exact `Card Support TODO` label.
- File intentionally deferred card-support work.
- Reuse an obvious existing issue instead of duplicating it.

**Will not:**

- File speculative ideas with no known work attached.
- create long plans, implementation designs, or exhaustive card inventories;
- silently defer discovered work without either fixing it or filing an issue.

## Error Handling

- If the label is missing, create it once:

  ```bash
  gh label create "Card Support TODO" \
    --description "Known work that still needs to be done to support a corpus of cards." \
    --color "BCEA7A"
  ```

- If `gh` is unauthenticated or issue creation fails, report the command error
  and continue the original implementation without pretending the TODO was
  filed.
