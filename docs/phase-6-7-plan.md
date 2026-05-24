# Phase 6/7 Plan: Core Combat Completion and Permanent Interaction

This plan covers the next two roadmap phases after the current minimal combat/death foundation.

The product goal remains: accept four Commander decklists, run repeated AI-controlled games with those decks, and produce an analytics report for the individual deck being tested.

## Current foundation

Already implemented:

- Four-player Commander game state.
- Deterministic setup and seeded engine RNG.
- Turn loop, priority, land play, simple mana payment, and simple spell casting.
- Runtime player targets for simple targeted spells.
- Basic effect primitives for draw, life gain/loss, and player damage.
- State-based player elimination.
- Combat phase structure with attacker and blocker declarations.
- Combat damage to players and creatures.
- Lethal and 0-toughness creature cleanup.
- Battlefield zone-change helpers for card-backed permanents and tokens.
- CLI smoke modes for land, spells, and combat.

## Planning review notes

This plan incorporates Opus 4.7 review feedback:

- Build effective power/toughness and counter-aware lethal damage before advanced combat keywords.
- Defer Protection until targeting, prevention, and attachment systems exist.
- Defer regeneration and prevention/replacement effects until the later event/replacement architecture.
- Defer planeswalker and battle attacks until permanent damage exists.
- Defer attack taxes until advanced costs are available.
- Shape damage assignment as an agent-choice seam even if the first implementation uses deterministic assignment.

## Phase 6 — Complete core combat

### Goal

Make creature combat mechanically faithful enough for common Commander games while avoiding broader event, layer, trigger, and replacement systems until later phases.

### Explicit non-goals

- No Protection.
- No regeneration.
- No general damage prevention or replacement effects.
- No planeswalker or battle attacks until permanent damage exists.
- No attack taxes requiring payment choices.
- No triggered abilities from attacking, blocking, damage, or death.

### Step 6.1 — Effective power/toughness foundation

Work:

- Add rules-level helpers for effective power, effective toughness, and lethal damage needed.
- Include base P/T and +1/+1 / -1/-1 counters.
- Route existing combat damage and creature SBAs through these helpers.
- Keep dynamic star P/T unsupported for now unless a card implementation supplies a value.

Files:

- `mtg/rules/combat.go`
- `mtg/rules/sba.go`
- `mtg/rules/*_test.go`

Tests:

- +1/+1 counters increase combat damage and lethal threshold.
- -1/-1 counters reduce combat damage and can cause 0-toughness death.
- Existing lethal damage tests still pass through effective toughness.

### Step 6.2 — Multi-blocking and damage-assignment seam

Work:

- Allow multiple blockers per attacker.
- Keep each blocker assigned to only one attacker.
- Add blocker order for attackers with multiple blockers.
- Add a damage-assignment choice shape or helper boundary so agent-driven damage assignment can be added later.
- Use deterministic default assignment for now: assign lethal damage to blockers in order, then move to the next blocker.

Files:

- `mtg/game/combat.go`
- `mtg/game/action/action.go` if the action payload needs assignment data.
- `mtg/rules/combat.go`
- `mtg/rules/combat_test.go`

Tests:

- Multiple blockers can legally block one attacker.
- Duplicate blocker declarations are rejected.
- Blocker order is recorded.
- A large attacker assigns damage across blockers in order.
- Insufficient damage is assigned to the first blocker only.

### Step 6.3 — Evasion and block restrictions

Work:

- Implement Flying and Reach block legality.
- Implement Menace requiring at least two blockers.
- Preserve Defender as an attack restriction.
- Add simple cannot-block checks only when represented by existing data.
- Do not implement Protection here.

Files:

- `mtg/rules/combat.go`
- `mtg/rules/combat_test.go`
- `mtg/rules/README.md`

Tests:

- Flying can be blocked by Flying or Reach.
- Flying cannot be blocked by a normal ground creature.
- Menace requires two or more blockers.
- Menace works with multi-block damage assignment.

### Step 6.4 — First strike and double strike

Work:

- Use `StepFirstStrikeDamage` only when at least one attacker or blocker has First Strike or Double Strike.
- Resolve first-strike damage first.
- Apply SBAs after first-strike damage before normal combat damage.
- Double Strike creatures deal damage in both combat damage passes.

Files:

- `mtg/rules/combat.go`
- `mtg/rules/phases.go` if step flow changes.
- `mtg/rules/combat_test.go`

Tests:

- First striker kills a blocker before it can deal normal combat damage.
- Double striker deals damage in both passes.
- If no first/double strike exists, no extra first-strike step or priority window runs.

### Step 6.5 — Trample and deathtouch

Work:

- Extend lethal damage calculation so Deathtouch means 1 damage is lethal for assignment purposes.
- Implement Trample remainder assignment to defending player after lethal damage is assigned to blockers.
- Combine Trample and Deathtouch correctly.

Files:

- `mtg/rules/combat.go`
- `mtg/rules/combat_test.go`

Tests:

- 5/5 Trample blocked by 2/2 deals 2 to blocker and 3 to player.
- Deathtouch attacker needs only 1 damage assigned as lethal.
- Deathtouch + Trample can assign 1 to blocker and the rest to player.
- First Strike + Deathtouch kills before normal damage.

### Step 6.6 — Lifelink and commander combat damage

Work:

- Apply Lifelink life gain equal to combat damage dealt by the source.
- Track Commander combat damage when the source permanent's `CardInstanceID` is that controller's commander.
- Scope commander-damage tracking to actual commander card instances; defer copies/tokens.

Files:

- `mtg/rules/combat.go`
- `mtg/rules/result.go` if logs need more fields.
- `mtg/rules/combat_test.go`

Tests:

- Lifelink gains life from damage to players.
- Lifelink gains life from damage to creatures.
- 21 commander combat damage from one commander eliminates a player.
- Non-commander creatures do not add commander damage.

### Step 6.7 — Indestructible

Work:

- Prevent destroy effects and lethal-damage SBAs from destroying indestructible permanents.
- Marked damage remains on indestructible creatures until cleanup.
- Do not implement regeneration here.

Files:

- `mtg/rules/sba.go`
- `mtg/rules/zones.go`
- `mtg/rules/combat_test.go`

Tests:

- Indestructible creature with lethal marked damage survives SBA.
- Indestructible creature keeps marked damage until cleanup.
- Destroy effect does not move an indestructible permanent to graveyard.

### Step 6.8 — Attack requirements, restrictions, and goad basics

Work:

- Use existing `Permanent.Goaded` data.
- Enforce goaded creatures attacking if able.
- Prefer legal attacks against non-goading players when possible.
- Keep attack taxes deferred until advanced costs.

Files:

- `mtg/rules/combat.go`
- `mtg/rules/combat_test.go`

Tests:

- Goaded creature must attack if able.
- Goaded by one player attacks a different player if possible.
- Goaded by two of three opponents must attack the remaining non-goading opponent if possible.
- Goad does not force illegal attacks.

### Step 6.9 — Documentation and roadmap updates

Work:

- Update `ROADMAP.md` checkboxes as features land.
- Update `mtg/rules/README.md`, `mtg/game/README.md`, and `mtg/game/action/README.md`.
- Document explicit deferrals: Protection, regeneration, prevention/replacement, planeswalker/battle attacks, and attack taxes.

Validation:

```bash
go test ./mtg/rules
go test ./...
go vet ./...
go run ./cmd/council4 -mode combat -verbose -nopass
```

## Phase 7 — Permanent interaction and richer state-based actions

### Goal

Add common non-combat permanent interaction and richer battlefield SBAs needed by removal, counters, tokens, board wipes, and simple Commander staples.

### Explicit non-goals

- No full trigger system.
- No full continuous-effect layer system.
- No broad replacement/prevention framework.
- No strategic choice framework beyond simple deterministic choices unless a narrow action already exists.

### Step 7.1 — Runtime permanent targets

Work:

- Extend target choice generation to include permanents by object ID.
- Support common constraints: creature, artifact, enchantment, land, nonland permanent, any permanent, controlled by opponent, controlled by you.
- Re-check target legality on resolution.
- Counter by rules if all targets are illegal.

Files:

- `mtg/game/target.go`
- `mtg/rules/targets.go`
- `mtg/rules/actions.go`
- `mtg/rules/targets_test.go`

Tests:

- Target lists include only matching public permanents.
- Illegal permanent targets are rejected during action application.
- A spell with all targets illegal on resolution does not apply effects.

### Step 7.2 — Permanent effect primitives

Work:

- Add effect primitives for destroy, exile, bounce to hand, sacrifice, tap, untap, damage to permanent, and create token.
- Reuse battlefield zone-change helpers from Phase 5.
- Add logs where useful for future reporting.

Files:

- `mtg/game/ability.go`
- `mtg/rules/effects.go`
- `mtg/rules/zones.go`
- `mtg/rules/result.go`
- `mtg/rules/effects_test.go`

Tests:

- Destroy moves a card-backed permanent to owner's graveyard.
- Exile moves to owner's exile.
- Bounce moves to owner's hand.
- Sacrifice moves controller-chosen permanent to owner's graveyard.
- Tap/untap changes tapped state.
- Damage-to-permanent marks damage and triggers SBA when lethal.
- Create-token creates a permanent with `TokenDef`.

### Step 7.3 — Planeswalker and battle attack targets

Work:

- Allow attack targets to include planeswalker or battle object IDs.
- Combat damage to planeswalkers removes loyalty.
- Combat damage to battles removes defense.
- Add SBAs for 0 loyalty and 0 defense if the data model supports them.

Files:

- `mtg/game/combat.go`
- `mtg/rules/combat.go`
- `mtg/rules/sba.go`
- `mtg/rules/combat_test.go`

Tests:

- Creature can attack a planeswalker controlled by an opponent.
- Creature can attack a battle.
- Combat damage reduces loyalty/defense instead of player life.
- 0-loyalty planeswalker and defeated battle leave battlefield as appropriate for the current model.

### Step 7.4 — Mass effects and board wipes

Work:

- Add selector helpers for all creatures, all artifacts, all enchantments, all nonland permanents, and all permanents matching a predicate.
- Apply mass effects simultaneously from a snapshot.
- Respect indestructible for destroy effects.

Files:

- `mtg/rules/effects.go`
- `mtg/rules/zones.go`
- `mtg/rules/effects_test.go`

Tests:

- Destroy all creatures destroys every matching creature.
- Nonmatching permanents survive.
- Indestructible permanents survive destroy-based board wipes.
- Multiple simultaneous deaths are all logged.

### Step 7.5 — Token creation and lifecycle

Work:

- Create token permanents with owner, controller, timestamp, `TokenDef`, and summoning sickness.
- Prefer rules-correct lifecycle: token goes to graveyard, then ceases to exist as an SBA.
- If logs need stable identity after removal, use token object ID and token definition name.

Files:

- `mtg/rules/effects.go`
- `mtg/rules/sba.go`
- `mtg/rules/zones.go`
- `mtg/rules/result.go`

Tests:

- Token creation puts token on battlefield.
- Tokens can attack, block, take damage, and die.
- Token moves through death flow and is removed from zones by SBA.
- Token logs remain readable after removal.

### Step 7.6 — Simple additive temporary P/T modifiers

Work:

- Add until-end-of-turn additive P/T modifiers.
- Apply them in effective power/toughness helpers.
- Expire them during cleanup.
- Do not support keyword-granting temporary effects until the layer system exists.

Files:

- `mtg/game/game.go` or `mtg/game/turn.go` for storage.
- `mtg/rules/effects.go`
- `mtg/rules/phases.go`
- `mtg/rules/combat.go`

Tests:

- Giant Growth-like effect changes combat damage and lethal threshold.
- Modifier expires during cleanup.
- Multiple additive modifiers stack deterministically.

### Step 7.7 — Auras and Equipment skeleton

Work:

- Add attach/unattach helpers.
- Add legal attachment checks for basic Aura and Equipment cases.
- Add equip action only if cost support is ready enough.
- Add SBAs for illegal attachments and auras.

Files:

- `mtg/game/permanent.go`
- `mtg/rules/zones.go`
- `mtg/rules/actions.go`
- `mtg/rules/sba.go`

Tests:

- Aura attaches to legal target.
- Aura with illegal target goes to graveyard.
- Equipment remains on battlefield when equipped creature dies.
- Attachment references are cleaned up when objects leave battlefield.

### Step 7.8 — Maximum hand size cleanup discard

Work:

- Implement maximum hand size discard during cleanup.
- Use deterministic discard choice for now.
- Document that this needs choice-framework support later.

Files:

- `mtg/rules/phases.go`
- `mtg/rules/result.go` if discard logs are added.
- `mtg/rules/phases_test.go` or existing setup/phase tests.

Tests:

- Player with more than maximum hand size discards during cleanup.
- Player at or below maximum hand size does not discard.
- Deterministic discard order is stable.

### Step 7.9 — Additional SBAs

Work:

- Legendary rule.
- Planeswalker uniqueness / loyalty death if in scope.
- Battle defense defeat if in scope.
- Illegal attachments and aura legality.
- Token cease-to-exist SBA.

Files:

- `mtg/rules/sba.go`
- `mtg/rules/sba_test.go`

Tests:

- Each SBA mutates state correctly.
- SBA loop converges when multiple SBAs cascade.
- Logs are emitted where useful for reports/debugging.

### Step 7.10 — Documentation and log shape

Work:

- Update `ROADMAP.md` as items land.
- Update package READMEs.
- Add permanent zone-change, token, and counter log fields only where useful for future reports.
- Document remaining deferrals to Phase 8/9.

Validation:

```bash
go test ./mtg/rules
go test ./...
go vet ./...
go run ./cmd/council4 -mode land -verbose -nopass
go run ./cmd/council4 -mode spells -verbose -nopass
go run ./cmd/council4 -mode combat -verbose -nopass
```

## Cross-phase design notes

- Keep mutation logic in `mtg/rules`; keep `mtg/game` as pure data.
- Prefer small, reusable helpers over card-specific logic.
- Avoid broad interfaces unless they isolate strategies or future code escape hatches.
- Keep action payloads as tagged struct data, not action interfaces.
- Add tests at each step before moving to the next one.
- Keep `ROADMAP.md` checkboxes up to date as each feature lands.
