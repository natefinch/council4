# Example: Commander Staples Rollout

## Scenario

You have a card list at `testdata/commander-staples.txt` and want the agent to
implement the cards that fit the current council4 rules model, then report the
rest.

## User Request

> Use card-rollout on `testdata/commander-staples.txt`, three cards at a time.

## Agent Flow

```bash
go run ./cardgen/cmd/cardbatch parse \
  -in testdata/commander-staples.txt \
  -out .cardwork/commander-staples/cards.json

go run ./cardgen/cmd/cardbatch fetch \
  -manifest .cardwork/commander-staples/cards.json

go run ./cardgen/cmd/cardbatch worklist \
  -manifest .cardwork/commander-staples/cards.json \
  -repo . \
  -limit 3
```

The agent then sends those cards to implementation subagents. After they return:

```bash
go generate ./mtg/cards/...

go run ./cardgen/cmd/cardbatch validate \
  -manifest .cardwork/commander-staples/cards.json \
  -repo .

go run ./cardgen/cmd/cardbatch report \
  -manifest .cardwork/commander-staples/cards.json \
  -repo . \
  -md .cardwork/commander-staples/unsupported.md \
  -json .cardwork/commander-staples/unsupported.json
```

The agent then appends a functionality rollup to the Markdown report:

```markdown
## Missing functionality rollup

### Equipped-creature selector

- Cards: Basilisk Collar, Blazing Sunsteel
- Needed for: "Equipped creature has ..."; "Equipped creature gets ..."
- Current state: delegated to `ImplementationID`
- Likely area: effect selectors / attachment-aware continuous effects
```

## Outcome

The final response lists implemented cards, blocked cards, validation failures,
missing functionality grouped by reusable capability, and report paths. The
report becomes input for future rules-roadmap work.
