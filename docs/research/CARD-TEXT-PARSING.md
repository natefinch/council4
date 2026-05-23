# Magic: The Gathering — Card Text Parsing Deep Dive

> An AI-agent-oriented guide for converting the printed text of any Magic card
> into a precise model of its in-game effect.
>
> Authority: *Magic: The Gathering Comprehensive Rules* (effective 2026-04-17).
> All rule numbers (`CR xxx.y`) refer to that document
> (the Magic Comprehensive Rules, referenced externally).
>
> Companions:
> - [`magic-the-gathering-a-trading-card-game-rules-and-.md`](./magic-the-gathering-a-trading-card-game-rules-and-.md) — how the game is played overall.
> - [`MTG-GLOSSARY.md`](./MTG-GLOSSARY.md) — slang/jargon.
> - [`COMMANDER-STRATEGY.md`](./COMMANDER-STRATEGY.md) — strategy.

This document is **read-once, decide-fast** reference. The structure mirrors
the order in which an agent should interrogate a card.

---

## 0. TL;DR Decision Procedure

Given the printed text of a card, do these steps in order:

1. **Read the type line** (CR 205, §1 below). It tells you *when* the card can
   be cast and *whether it stays on the battlefield*.
2. **Read the mana cost / color identity** (CR 202).
3. **Split the text box into paragraphs** (CR 113.2c). Each paragraph is *one ability*
   (with rare exceptions: keyword lists separated by commas count as multiple
   abilities on a single line — CR 702.1).
4. **Classify each ability** as one of four kinds (§3):
   - **Spell ability** — only on instants/sorceries, executed on resolution.
   - **Activated ability** — `[Cost]: [Effect].`
   - **Triggered ability** — starts with `When`, `Whenever`, or `At`.
   - **Static ability** — a declarative statement that is just true.
5. For each ability extract the **5-tuple**:
   `(trigger or cost, targets, modes, effect(s), zone-of-function)`.
6. Apply the **keyword dictionary** (§9) for any keyword (italic reminder text
   is not authoritative; CR 702.1 is).
7. Check for **replacement / "instead" / "as ... enters" wording** (§7).
   These do not use the stack and silently rewrite events.
8. Decide **timing/speed** from the type line + Flash + special permissions (§8).
9. Combine: the card's effect on a live game state is the union of all its
   abilities' effects, gated by their respective triggers/costs/zones.

---

## 1. Anatomy of a Card (CR 200–213)

Every Magic card is a tuple of fields. An agent should normalize a card to
this shape before reasoning:

| Field | Rule | Notes |
|------|------|-------|
| **Name** | 201 | Self-reference; "this permanent" === the card itself in its text (CR 201.4). |
| **Mana cost** | 202 | Symbols `{W}{U}{B}{R}{G}{C}{X}{2/W}{W/P}` etc. Mana value (MV) = sum of generic + colored, X=0 unless on the stack (CR 202.3b). |
| **Color** | 105, 202.2 | Determined by mana cost symbols and color indicator (204), *not* by name or art. |
| **Type line** | 205 | `[Supertypes] [Card types] — [Subtypes]`. See §2. |
| **Text box** | 207 | Rules text + reminder text (italic, non-authoritative) + flavor text (italic, ignored). |
| **Power/Toughness** | 208 | Creatures and Vehicles only. `*` means CDA (CR 208.2). |
| **Loyalty** | 209 | Planeswalkers. |
| **Defense** | 210 | Battles. |
| **Hand/Life modifier** | 211–212 | Vanguard only; ignore in standard play. |

When an agent ingests a card, store it as JSON-like:
```
{ name, mana_cost, mv, colors, supertypes, types, subtypes,
  power, toughness, loyalty, defense, abilities: [ ... ] }
```

---

## 2. Type Line — the Master Switch (CR 300)

The type line determines almost everything about *when* and *how* a card is used.

### 2.1 Card types (CR 300.1)

| Type | Permanent? | Default cast timing | Key rules |
|------|------------|---------------------|-----------|
| Land | yes | Special action, not cast (CR 305, 116.1) | One per turn; produces mana via mana abilities. |
| Creature | yes | Sorcery speed | Has P/T; has summoning sickness (CR 302.1, 302.6). |
| Artifact | yes | Sorcery speed | Usually colorless; subtypes drive most behavior (Equipment, Vehicle, Food…). |
| Enchantment | yes | Sorcery speed | Aura subtype attaches (CR 303.4); Saga has chapter triggers (CR 714). |
| Planeswalker | yes | Sorcery speed | Loyalty abilities once/turn at sorcery speed (CR 606). |
| Battle | yes | Sorcery speed | Has defense counters; transforms when defeated (CR 310). |
| Instant | no  | **Any time you have priority** (CR 117) | Goes to graveyard on resolve (CR 304). |
| Sorcery | no  | Sorcery speed | Goes to graveyard on resolve (CR 307). |
| Kindred | mod | Combined with another type (CR 308) | Lets non-creature cards have creature subtypes. |

> **"Sorcery speed"** = your main phase, your turn, with the stack empty
> (CR 117.1a). **Instant speed** = anytime you have priority.

### 2.2 Supertypes (CR 205.4)

`Basic`, `Legendary`, `Snow`, `World`, `Ongoing`. Most relevant:
- **Legendary** — "legend rule" state-based action: you can't control two
  permanents with the same name; one must go to graveyard (CR 704.5j).
- **Basic** — basic lands have intrinsic mana abilities (CR 305.6).

### 2.3 Subtypes (CR 205.3)

Drive ability hooks: an effect that says "Equipment you control" only sees
artifacts with the Equipment subtype. Creature subtypes (Goblin, Wizard, …)
matter for "tribal" effects.

---

## 3. The Four Kinds of Abilities (CR 113.3)

Identifying which kind an ability is governs **stack/priority behavior**,
**when it does anything**, and **whether it can be countered or responded to**.

### 3.1 Spell abilities (CR 113.3a)

- Only on **instants and sorceries**.
- The whole text box is the spell's instructions, executed when the spell
  resolves (CR 608.2).
- A spell ability may *create* lasting effects (e.g., "until end of turn"),
  but the card itself goes to the graveyard on resolution (CR 608.2k).

### 3.2 Activated abilities (CR 113.3b, 602)

Form: **`[Cost]: [Effect].`** — the colon is the syntactic marker.

Recognize them by the colon. Examples:
- `{T}: Add {G}.` — mana ability (CR 605, doesn't use the stack).
- `{2}, {T}: Draw a card.` — multi-component cost.
- `{1}{B}, Sacrifice a creature: Target player loses 2 life.` — additional non-mana cost.

Special subclasses:
- **Mana abilities** (CR 605.1): could produce mana, no target, isn't a loyalty ability. Resolve immediately, can't be responded to.
- **Loyalty abilities** (CR 606): `+N`, `0`, `-N`. Once per turn, sorcery speed only, only by controller.

### 3.3 Triggered abilities (CR 113.3c, 603)

Form: **`[Trigger condition], [effect].`** — begins with `When`, `Whenever`,
or `At`.

| Wording | Meaning |
|--------|---------|
| `When ...` | Fires once on a specific event (CR 603.2). |
| `Whenever ...` | Fires every time the event occurs. |
| `At the beginning of [step] ...` | Fires at a turn-based point (CR 603.6c). |

Variants the parser must recognize:
- **ETB** — "When this enters" / "When [name] enters" (CR 603.6a).
- **LTB / Dies** — "When this dies" = "from the battlefield to a graveyard" (CR 700.4, 603.6f).
- **Conditional triggers** — "When..., if..., do..." (intervening *if*, CR 603.4): condition checked when the event happens AND on resolution.
- **State triggers** — "Whenever you control five or more lands..." (CR 603.8): trigger as long as the condition is true and re-trigger only after becoming false then true again.
- **Delayed triggers** — created by spells/abilities that say "When ... next ...", "At the beginning of the next end step ..." (CR 603.7).
- **Ability triggers on its own** when the event occurs; goes on the stack the next time *any* player would get priority (CR 603.3). This is why a triggered ability can be responded to but the trigger itself can't be prevented.

### 3.4 Static abilities (CR 113.3d, 604)

Declarative sentences with **no `:`** and not starting with `When/Whenever/At`. They
generate **continuous effects** that are simply true while the source is in
the appropriate zone (default: battlefield, CR 113.6). Examples:

- `Creatures you control get +1/+1.`
- `Players can't gain life.`
- `Lightning Bolt costs {1} less to cast.`

Subclass: **Replacement / prevention effects** (§7) are static abilities
written as `if ... would ..., instead ...` or `if ... would ..., ... that
much/many ... [modifier]` (CR 614, 615).

> **Heuristic — quickly classify a paragraph:**
> 1. Contains `:` not inside `{...}`? → Activated.
> 2. Starts with `When|Whenever|At`? → Triggered.
> 3. On an instant or sorcery and not (1) or (2)? → Spell.
> 4. Otherwise → Static.

---

## 4. Costs (CR 118, 601.2f)

### 4.1 Mana symbols (CR 107.4)

| Symbol | Means |
|--------|-------|
| `{W} {U} {B} {R} {G}` | One mana of that color. |
| `{C}` | One colorless. |
| `{N}` (digit) | N generic mana, payable with any. |
| `{X}` | Variable; chosen on cast (CR 107.3). |
| `{W/U}` | Hybrid: pay either color. |
| `{2/W}` | Monocolored hybrid: 2 generic *or* {W}. |
| `{W/P}` | Phyrexian: pay {W} or 2 life (CR 107.4f). |
| `{S}` | Snow mana (must come from a snow source, CR 107.4g). |
| `{T}` `{Q}` | Tap / untap symbol — used in costs, not a mana symbol. |

**Mana value (MV)** = numeric sum of all mana symbols in the cost. `{X}` counts
as the chosen X *only on the stack*; otherwise as 0 (CR 202.3b).

### 4.2 Additional / alternative costs (CR 118.8–118.9)

Watch for:
- `As an additional cost to cast this spell, [pay X / sacrifice Y / discard Z].`
  → must be paid in addition to the mana cost.
- `You may [pay X] rather than pay this spell's mana cost.`
  → an *alternative* cost (Flashback, Madness, Evoke, etc.).
- `[Spell] costs {N} less to cast` / `{N} more to cast.` → cost modifiers
  applied during the "determine total cost" step of casting (CR 601.2f).

### 4.3 Special non-mana costs (CR 118.12)

`Tap`, `Sacrifice X`, `Discard X`, `Pay N life`, `Exile X`, `Remove a counter`,
`Reveal a card`. All are paid only when announcing/casting/activating, never
on resolution.

---

## 5. Targets (CR 115)

**Targets are chosen on cast/activation/trigger**, not on resolution
(CR 601.2c, 603.3d).

Recognize targeting by the literal word **"target"**:
- `Destroy target creature.` → 1 target, declared at cast time.
- `Target player draws two cards.` → 1 target.
- `Up to two target creatures.` → 0–2 targets.
- `Each opponent` / `each creature` (no "target") → **not targeted**; can't be
  responded to with hexproof/shroud/protection-from-being-targeted (CR 702.11, 702.18, 702.16).

When parsing, count `target` occurrences. Each occurrence is a target slot.

> If on resolution every target has become illegal, the spell/ability is
> countered by the rules and does nothing (CR 608.2b).

---

## 6. Modes and Choices

### 6.1 Modal spells (CR 700.2)

`Choose one —` / `Choose two —` / `Choose one or both —`. Each bullet is a
**mode**. The chosen modes are locked in on cast (CR 601.2d). Examples:

```
Choose one —
• Destroy target creature.
• Target player draws two cards.
```

`Entwine [cost]` (CR 702.42) lets you choose all modes by paying extra.

### 6.2 "Up to" / "as many as" (CR 109.4)

`Up to N` means 0..N targets; `up to N times` means 0..N executions.

### 6.3 "May" vs imperative

- `You may [do X].` → optional. Choice on resolution unless the text says
  otherwise.
- `[Do X].` (no may) → mandatory (only the legal subset is performed; if a
  step is impossible, do as much as you can — CR 101.3).

### 6.4 "If [condition]" (CR 614.13, 700.5)

- `If you do, ...` → only fires when the previous optional was taken.
- `If you don't, ...` → fires when it wasn't taken.
- `Then ...` → strict ordering: do first part, **then** do second part.

---

## 7. Replacement, Prevention, "Instead", and "As ... Enters"

These are the silent rewriters of game events. They do **not** use the stack.

### 7.1 Replacement effects (CR 614)

Pattern markers:
- `If [event] would [happen], [different thing happens] instead.` (CR 614.1a)
- `[Permanent] enters with [counters / tapped / etc.]` (CR 614.1c, "as ... enters" or "enters with").
- `[Permanent] enters as a [thing].` (CR 614.1d).
- `Skip [step/phase].` (CR 614.1b).
- Self-replacement of a spell's resolution: `... draws a card. If you do, ...`.

When a replacement effect applies, the original event never happens —
triggers that watch for the original event don't see it (CR 614.6). Triggers
that watch the replacement do.

### 7.2 Prevention effects (CR 615)

- `Prevent the next N damage that would be dealt to [target] [duration].`
- Apply to specific damage events; consumed when matched.

### 7.3 Multiple replacements / prevention (CR 616)

If several apply to the same event, the **affected player or controller**
chooses the order. This matters for:
- Lifegain replaced by something that triggers on lifegain.
- ETB replacements that change counters and ETB-as-tapped at the same time.

### 7.4 Layer system for continuous effects (CR 613) — short version

For static abilities that change characteristics, apply effects in this order:

1. Copy effects.
2. Control change.
3. Text change.
4. Type/subtype/supertype.
5. Color.
6. Add/remove abilities + keyword counters.
7. Power/Toughness, in sublayers:
   - 7a CDAs setting P/T.
   - 7b Effects setting P/T to a specific value (`becomes a 3/3`, base P/T).
   - 7c Modifying effects (`+N/+N`, `-N/-N`, +1/+1 counters).
   - 7d P/T switch.

Within a layer, apply effects in **timestamp order** (CR 613.7), with
**dependency** overrides (CR 613.8). Most simple cards never need this;
two-card interactions involving "becomes a Creature" effects do.

---

## 8. Timing and Speed (CR 117)

Speed is determined per-action:

| Action | Speed |
|--------|-------|
| Casting a creature/sorcery/artifact/enchantment/planeswalker/battle (default) | Sorcery |
| Casting an instant | Instant |
| Casting a card with **Flash** (CR 702.8) | Instant |
| Activating an ability (default) | Instant |
| Activating an ability whose cost includes `{T}` on a creature with summoning sickness | Forbidden until next turn (CR 302.1) |
| Activating a planeswalker loyalty ability | Sorcery, your turn, once per planeswalker (CR 606.5) |
| Playing a land | Sorcery, your main phase, empty stack, once/turn (CR 305.2) — special action, no stack (CR 116.2a) |
| Special actions (turn face up, suspend exile cast, conspire, etc.) | Per the rule for that action (CR 116) |

Important secondary timing constraints to look for in text:
- `Activate only as a sorcery.` → forces sorcery speed (CR 113.6e).
- `Activate only any time you could cast a sorcery.` → same effect.
- `Activate only during combat.` / `... during your upkeep.` → step-restricted.
- `Activate only once each turn.` → frequency cap.

---

## 9. Keyword Abilities — Lookup Table

> Reminder text is *not* the rule; CR 702.x is. The list is huge (192 entries,
> CR 702.2–702.192). The table below covers all **evergreen** and high-frequency
> keywords. For anything else, look up `702.NNN` in the comprehensive rules.

### 9.1 Evergreen combat / interaction keywords

| Keyword | CR | One-line semantics |
|--------|----|--------------------|
| **Deathtouch** | 702.2 | Any nonzero damage dealt by this to a creature is lethal. |
| **Defender** | 702.3 | Can't attack. |
| **Double strike** | 702.4 | Deals first-strike *and* normal combat damage. |
| **First strike** | 702.7 | Deals combat damage in a separate, earlier damage step. |
| **Flash** | 702.8 | May be cast at instant speed. |
| **Flying** | 702.9 | Only blocked by creatures with flying or reach. |
| **Haste** | 702.10 | Ignore summoning sickness; can attack and `{T}` the turn it enters. |
| **Hexproof** | 702.11 | Opponents can't *target* it with spells or abilities. |
| **Indestructible** | 702.12 | "Destroy" effects and lethal damage don't kill it (still dies to 0 toughness, exile, sacrifice). |
| **Lifelink** | 702.15 | Damage dealt by this also causes its controller to gain that much life. |
| **Menace** | 702.111 | Can't be blocked except by 2+ creatures. |
| **Protection from X** | 702.16 | DEBT: **D**amage prevented from X, **E**nchanted/equipped by X impossible, **B**locked by X impossible, **T**argeted by X impossible. |
| **Reach** | 702.17 | Can block creatures with flying. |
| **Shroud** | 702.18 | *No one* can target it (rare; mostly replaced by hexproof). |
| **Trample** | 702.19 | Excess combat damage past blockers' toughness goes to the defending player/PW/battle. |
| **Vigilance** | 702.20 | Doesn't tap to attack. |
| **Ward [cost]** | 702.21 | When this becomes the target of a spell/ability an opponent controls, counter it unless they pay [cost]. |

### 9.2 Common non-combat keyword abilities

| Keyword | CR | Semantics |
|--------|----|-----------|
| **Equip [cost]** | 702.6 | Activated: attach Equipment to a creature you control. Sorcery speed. |
| **Enchant [object]** | 702.5 | Restricts what an Aura can be attached to. |
| **Cycling [cost]** | 702.31 | `[Cost], Discard this card: Draw a card.` Activated, any time. |
| **Flashback [cost]** | 702.34 | Cast from graveyard for [cost] instead of mana cost; exile on resolution. |
| **Kicker [cost]** | 702.33 | Optional additional cost on cast; enables extra effects ("if this was kicked, ..."). |
| **Madness [cost]** | 702.35 | When discarded, exile; may be cast from exile for [cost]; otherwise to graveyard. |
| **Morph / Disguise [cost]** | 702.37 / 702.168 | May cast face-down as a 2/2 for {3}; turn face up by paying [cost]. |
| **Convoke** | 702.51 | Tap any creatures to pay {1}/colored mana per creature. |
| **Delve** | 702.66 | Exile cards from your graveyard to pay {1} each. |
| **Suspend [N] [cost]** | 702.62 | Exile from hand with N time counters; remove one each upkeep; cast for free when last is removed. |
| **Storm** | 702.40 | When cast, copy this for each spell cast before it this turn. |
| **Cascade** | 702.85 | When cast, exile cards until you hit a nonland with lower MV; may cast it for free. |
| **Prowess** | 702.108 | Whenever you cast a noncreature spell, this gets +1/+1 until end of turn. |
| **Mutate** | 702.140 | May cast for the mutate cost; merge with non-Human creature. |
| **Companion** | 702.139 | Deck-construction restriction; cast once from sideboard via the companion zone. |

### 9.3 Common keyword *actions* (CR 701)

These appear inside ability text rather than as standalone keywords on their own line.

| Action | CR | Definition |
|--------|----|------------|
| **Tap** | 701.26 | Turn 90°. |
| **Untap** | 701.26 | Return upright. |
| **Destroy** | 701.8 | Move from battlefield to graveyard. Stopped by Indestructible. |
| **Exile** | 701.13 | Move to the exile zone. Bypasses Indestructible. |
| **Sacrifice** | 701.21 | Owner moves their own permanent to graveyard. Bypasses Indestructible/Hexproof. |
| **Discard** | 701.9 | Move from hand to graveyard (or as the spell directs). |
| **Counter** | 701.6 | Remove a spell/ability from the stack; spell goes to graveyard (CR 701.6c). |
| **Mill N** | 701.17 | Put top N of your library into your graveyard. |
| **Scry N** | 701.22 | Look at top N; put any number on the bottom in any order, the rest on top in any order. |
| **Surveil N** | 701.25 | Look at top N; put any in graveyard, the rest on top in any order. |
| **Search** | 701.23 | Look through a zone (usually your library) for a card matching criteria. Library searches require a shuffle if you searched a hidden zone (CR 701.23d). |
| **Reveal** | 701.20 | Show without changing zone. |
| **Create [token]** | 701.7 | Put a token onto the battlefield. |
| **Fight** | 701.14 | Each fighting creature deals damage equal to its power to the other. |
| **Investigate** | 701.16 | Create a Clue token. |
| **Proliferate** | 701.34 | Choose any number of permanents/players with counters; put another counter of an existing kind on each. |
| **Regenerate** | 701.19 | Replacement: next time it would be destroyed, instead tap, remove from combat, remove all damage. |
| **Goad** | 701.15 | Until your next turn, that creature attacks each combat if able and not you. |
| **Transform** | 701.27 | Turn a double-faced permanent to its other face. |

---

## 10. Counters (CR 122)

- `+1/+1`, `-1/-1`, charge, loyalty, time, etc. Counters are tracked per
  permanent or player.
- `+1/+1` and `-1/-1` cancel as a state-based action (CR 704.5r).
- Loyalty counters double as a planeswalker's "life" — at 0, it dies (CR 704.5i).
- Defense counters do the same for battles (CR 704.5x).
- "Modular", "Persist", "Undying", "Adapt", "Bolster" all manipulate +1/+1
  counters; "Wither", "Infect", "Toxic" use -1/-1 counters or poison.

---

## 11. Damage (CR 120)

- Damage is *dealt* by sources (CR 120.1). Combat damage and noncombat damage
  follow the same rule once dealt.
- Damage to a creature/PW becomes a marked-damage value, removed at cleanup (CR 120.3).
- Damage to a planeswalker is dealt to the planeswalker; loyalty is removed (CR 120.4).
- Combat damage assignment honors blocker order (CR 510.1c). Trample is a
  damage-assignment exception (CR 702.19b).
- Lethal damage and "destroy" are different events. Indestructible blocks
  destroy and lethal-damage destruction, not 0-toughness or exile.

---

## 12. Zones (CR 400)

| Zone | Public? | CR |
|------|---------|----|
| Library | hidden, ordered | 401 |
| Hand | hidden | 402 |
| Battlefield | public | 403 |
| Graveyard | public, ordered | 404 |
| Stack | public, ordered | 405 |
| Exile | usually public, sometimes face-down | 406 |
| Command | public | 408 |

When parsing a card, identify **the zone its abilities function in** (CR 113.6).
Default = battlefield for permanents, stack for instants/sorceries. Watch for
exceptions:

- `You may cast this card from your graveyard ...` → functions in graveyard.
- `When this is put into a graveyard from anywhere ...` → triggers from
  graveyard (the trigger looks back at the moment it was on the battlefield via
  last-known information, CR 603.10).
- `As long as this is in exile ...` → functions from exile.

---

## 13. Stack and Resolution (CR 405, 608)

1. Casting/activating goes on the stack (mana abilities and special actions
   excepted).
2. Players pass priority. With both players passing on an empty stack, the
   active player advances steps; with both passing on a non-empty stack, the
   top object resolves (CR 117.4).
3. On resolution: follow instructions in order (CR 608.2c). If a target is
   illegal at resolve, the *spell or ability is countered by the game rules*
   and any non-target effects don't happen unless they're explicitly free of
   the countered branch (CR 608.2b).
4. On resolution of an instant/sorcery, after all effects, the card goes to
   its owner's graveyard (CR 608.2k).

---

## 14. Common Templating Patterns

The agent should recognize these idioms verbatim:

- `Enters tapped.` — replacement (CR 614.1c).
- `Enters with N +1/+1 counters on it.` — replacement.
- `If [perm] would enter, instead it enters with ...` — explicit replacement.
- `Until end of turn` / `until your next turn` — duration on a continuous effect (CR 611).
- `This turn` — duration anchored to current turn.
- `For each [thing], do X.` — repeat or scale by count; count is determined on resolution (CR 107.1, 107.3).
- `Equal to [value]` — value is determined when the effect applies (CR 107.1b).
- `Up to N` — choose 0..N.
- `Among them` / `among those` — restrict scope to the selection, not the whole battlefield.
- `That player` / `its controller` / `that creature's controller` — pronouns refer to the most recent matching antecedent (CR 109.5 broadly).
- `Cast this only [restriction]` — casting restriction on the stack & in source zone (CR 113.6e).
- `You may have [X] enter as a copy of [Y]` — copy effect (CR 707).
- `[Permanent] becomes a [type] [P/T] in addition to its other types/abilities` — animation effect; layers 4 and 7b.

---

## 15. Algorithm — Turning a Card into Game Effects

Pseudocode for an agent:

```text
function parse(card):
  card.fields = read_frame(card)             # §1
  apply_typeline(card)                        # §2
  paragraphs = split_text_box(card.text)
  card.abilities = []
  for p in paragraphs:
    for sub in split_keyword_run(p):          # commas in keyword line → multiple
      card.abilities.append(classify(sub))    # §3.heuristic

function evaluate(card, game_state):
  effects = []
  for ab in card.abilities:
    if ab.kind == 'static':
      if zone_ok(card, ab):                   # CR 113.6
        effects += continuous_effect(ab, layers=613)
    elif ab.kind == 'triggered':
      if event_matches(game_state, ab.trigger):
        stack.push(instantiate(ab))           # CR 603.3
    elif ab.kind == 'activated':
      if can_activate(card, ab, game_state):  # priority + zone + cost payable
        offer_action(ab)
    elif ab.kind == 'spell':
      pass                                    # only relevant during cast/resolve
  return effects
```

When resolving any spell or ability:

```text
function resolve(obj):
  for target in obj.targets:                  # CR 608.2b
    if not legal(target): mark_illegal(target)
  if all_targets_illegal(obj):
    counter_by_rules(obj); return
  for instruction in obj.effect:
    apply(instruction, skipping_illegal_targets=True)
  if obj.is_instant_or_sorcery:
    move_to_graveyard(obj)
```

---

## 16. Worked Examples

### 16.1 Lightning Bolt — `{R}` — Instant

> Lightning Bolt deals 3 damage to any target.

- Type: Instant → can be cast anytime with priority.
- Mana cost `{R}`, MV 1, color: red.
- 1 paragraph, on instant → spell ability.
- `any target` = creature, player, planeswalker, or battle (CR 115.4) — 1 target slot.
- Effect: deal 3 damage to the chosen target on resolution.
- Goes to graveyard after resolving (CR 608.2k).

Agent representation:
```
{ name: "Lightning Bolt",
  cast_speed: instant,
  abilities: [
    { kind: spell, targets: [{type: any_target}],
      effect: deal_damage(source=self, target=$1, amount=3) } ] }
```

### 16.2 Llanowar Elves — `{G}` — Creature — Elf Druid 1/1

> {T}: Add {G}.

- Type: Creature → sorcery-speed cast, summoning sickness (CR 302.1).
- One ability paragraph with `:` → activated.
- Cost: `{T}`. Effect: add `{G}`. Could produce mana, no target → **mana ability** (CR 605.1) → doesn't use the stack.
- P/T 1/1 baseline; can attack/block once summoning sickness is gone or with haste.

### 16.3 Doom Blade — `{1}{B}` — Instant

> Destroy target nonblack creature.

- Spell ability, 1 target slot with restriction `nonblack creature`.
- Target chosen on cast; if target becomes black or otherwise illegal at
  resolution → countered by the rules.
- Effect: `destroy` (CR 701.8); blocked by **Indestructible** (CR 702.12).

### 16.4 Birds of Paradise — `{G}` — Creature — Bird 0/1

> Flying
> {T}: Add one mana of any color.

- Two paragraphs.
- Paragraph 1: `Flying` → static keyword (CR 702.9).
- Paragraph 2: activated mana ability (CR 605.1). The "any color" is a
  choice on activation (CR 605.1c).
- Combat: 0/1 flier; can block fliers/reach.

### 16.5 Wrath of God — `{2}{W}{W}` — Sorcery

> Destroy all creatures. They can't be regenerated.

- Sorcery → main phase, your turn.
- Untargeted (no `target` keyword); affects every creature simultaneously.
- "Can't be regenerated" turns off the regenerate replacement (CR 701.19c).
- Indestructible creatures are unaffected.

### 16.6 Dryad Arbor — Land — Forest, Creature — Dryad 1/1

> (Dryad Arbor isn't a spell, it's affected by summoning sickness, and it has
> "{T}: Add {G}.")

- Type line: Land *and* Creature → both rules apply.
- Played as a land (special action, CR 305.2). Has summoning sickness as a
  creature (CR 302.1).
- Has the intrinsic `{T}: Add {G}.` ability of Forests (CR 305.6).

### 16.7 Saga — Example chapter card

> *(Each chapter ability is a triggered ability that triggers when this Saga
> enters and after your draw step.)*
>
> I — Each opponent loses 2 life.
> II — Draw a card.
> III — Create a 4/4 token.

- Type: Enchantment — Saga (CR 714).
- Three triggered abilities, with chapter counters (CR 714.2 – 714.4).
- After last chapter ability resolves, Saga is sacrificed (CR 714.5).

---

## 17. Edge Cases an Agent Must Handle

1. **Replacement vs trigger ordering** — Replacements apply *before* triggers
   see the event (CR 614.6). Example: a creature with "When this enters, draw
   a card" + a static "Creatures you control enter tapped" both apply, but
   "enters tapped" never *triggers* — it just modifies the entry event.
2. **Last known information** — When an object leaves a zone, abilities that
   look back at it (e.g., "When this dies, return a creature with MV ≤ this
   creature's power") use its values right before it left (CR 603.10).
3. **Layer dependency** — `Mycosynth Lattice` makes everything an artifact,
   then `March of the Machines` turns each noncreature artifact into a
   creature. These resolve in dependency order, not strict timestamp (CR 613.8).
4. **"Cast" vs "play"** — "Play" includes both casting spells and playing
   lands (CR 601.1, 305.1). Some effects let you "play" cards from elsewhere
   (e.g., Bolas's Citadel) — this includes lands.
5. **"Whenever ... cast"** vs **"Whenever ... enters"** — cast triggers fire
   when the spell goes onto the stack (CR 603.6b); ETB triggers fire after
   the spell resolves and the permanent is on the battlefield.
6. **Targets locked in vs replaced** — Targets are chosen once at cast
   (CR 601.2c). "Change targets" effects (Misdirection-style) modify them
   later but must result in legal targets (CR 115.7).
7. **State-based actions** (CR 704) — Run *before any player gets priority*:
   0-toughness/lethal-damage destruction, 0-loyalty death, legend rule, world
   rule, planeswalker uniqueness, Aura with no legal attachment, +1/+1 vs
   -1/-1 cancel, players at 0 life lose, 10-poison rule, etc. Many subtle
   "why did my creature die?" puzzles resolve here.
8. **Activated vs triggered look-alikes** —
   `When [thing happens], you may pay {2}: do X.` is a triggered ability with
   a payment as part of its effect, NOT an activated ability.
9. **"Can't be countered"** — applies to the spell/ability, not to its
   effects (CR 113.6g, 701.6e). Counterspells fail; cards that "exile target
   spell" still work because they don't say "counter".
10. **Protection nuances (DEBT)** — Protection from a quality means: can't be
    Dealt damage by, Equipped/Enchanted/Fortified/Attached by, Blocked by,
    Targeted by sources of that quality (CR 702.16b).

---

## 18. Quick Reference — "Where in the Rules?"

| Topic | CR section |
|-------|-----------|
| Card types | 300, 301–315 |
| Casting spells | 601 |
| Activating activated abilities | 602 |
| Handling triggered abilities | 603 |
| Handling static abilities | 604 |
| Mana abilities | 605 |
| Loyalty abilities | 606 |
| Resolving spells & abilities | 608 |
| Effects (one-shot vs continuous) | 609–611 |
| Layered continuous effects | 613 |
| Replacement effects | 614 |
| Prevention effects | 615 |
| Multiple replacement/prevention | 616 |
| Keyword actions | 701 |
| Keyword abilities | 702 |
| State-based actions | 704 |
| Copying objects | 707 |
| Double-faced cards | 712 |
| Sagas | 714 |

---

## 19. Glossary of Templating Verbs

| Verb / phrase | Meaning |
|---------------|---------|
| Add `{X}` | Produce mana into mana pool. |
| Attach | Move an Aura/Equipment/Fortification onto a permanent. |
| Cast | Put a card on the stack as a spell, paying costs (CR 601). |
| Counter | Remove an object from the stack (CR 701.6). |
| Create [token] | Make a new token permanent on the battlefield. |
| Deal damage | Cause damage from a source to a target/creature/player. |
| Destroy | Move a permanent from battlefield to graveyard (stopped by Indestructible). |
| Discard | Move a card from a hand to that owner's graveyard. |
| Draw | Move the top card of a library to its owner's hand (CR 121). |
| Exchange | Swap control/zone of two things (CR 701.12). |
| Exile | Move to the exile zone. |
| Fight | Two creatures deal damage equal to power to each other. |
| Mill | Top of library to graveyard. |
| Pay [cost] | Apply the cost requirement (mana, life, sacrifice, etc.). |
| Play | Cast a spell *or* play a land. |
| Proliferate | Add another counter of an existing kind to chosen permanents/players. |
| Regenerate | Replace next destruction this turn with tap+remove from combat+heal. |
| Return | Move from one zone to another (usually graveyard → hand or battlefield). |
| Reveal | Show a card without changing its zone. |
| Sacrifice | Owner moves their own permanent to their graveyard. |
| Scry | Sort the top of your library. |
| Search | Look through a hidden zone for matching cards. |
| Shuffle | Randomize a library. |
| Tap / Untap | Toggle the tapped state. |
| Transform | Flip a double-faced permanent. |

---

*End of deep dive. For any phrasing this guide doesn't cover, search the
rules file (the Magic Comprehensive Rules) by the specific verb or
keyword — the comprehensive rules define every term used on a card.*
