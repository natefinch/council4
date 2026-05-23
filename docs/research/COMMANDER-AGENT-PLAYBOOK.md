# Commander Playbook for an AI Agent

> Operational instructions for an autonomous agent playing **Magic: The Gathering — Commander** (4-player free-for-all by default) against other agents/humans, given:
>
> 1. **Full knowledge of its own decklist** (100 singleton cards + 1 commander, plus color identity).
> 2. **A query interface** to the game state with all the information a player at the table would have (public zones, life totals, mana pools, stack, priority, opponents' boards/graveyards, the agent's hand, etc.).
>
> This document tells the agent *what to compute, when to act, and how to choose*. It is designed to be read once at game start and consulted at every decision point.
>
> Companion docs (load these into context first):
> - [`magic-the-gathering-a-trading-card-game-rules-and-.md`](./magic-the-gathering-a-trading-card-game-rules-and-.md) — rules of play and high-level strategy.
> - [`CARD-TEXT-PARSING.md`](./CARD-TEXT-PARSING.md) — turn any printed card into a structured effect.
> - [`MTG-GLOSSARY.md`](./MTG-GLOSSARY.md) — slang & jargon.
> - [`COMMANDER-STRATEGY.md`](./COMMANDER-STRATEGY.md) — strategy fundamentals & 11 archetype playbooks.
> - Magic Comprehensive Rules — authoritative rules referenced externally; cite as `CR x.y`.

---

## 0. The Top-Level Agent Loop

```text
on_game_start():
  ingest_deck()                                  # §1
  classify_archetype()                           # §1.4
  build_threat_model_template()                  # §3.1
  load_card_index()                              # §1.3

on_priority(turn, phase, step, stack):
  state  = query_full_game_state()               # §2
  legal  = enumerate_legal_actions(state)        # §4
  if legal == [pass_priority]: return pass_priority()
  scored = [(score(a, state), a) for a in legal] # §6 (rubric)
  return argmax(scored)

on_choice_required(choice_type, options, ctx):
  return resolve_choice(choice_type, options, ctx) # §7

on_combat_step(step, state):
  return combat_decision(step, state)              # §8

on_game_end(result):
  log_lessons(result)                              # §13
```

Every "should I do X?" decision the agent makes flows through `score(action,
state)` (§6). Everything else in this document either defines the inputs to
that score or constrains when it can be invoked.

---

## 1. Pre-Game Preparation (Once Per Match)

### 1.1 Deck Ingestion

Build a structured representation of your own 100 cards. For each card, store:

```text
Card {
  name, mana_cost, mv, colors, color_identity,
  supertypes, types, subtypes,
  power, toughness, loyalty, defense,
  abilities: [parsed via CARD-TEXT-PARSING.md §15],
  tags: [ramp | draw | removal_spot | removal_mass |
         tutor | counterspell | protection | wincon |
         combo_piece | utility | mana_rock | mana_dork |
         landfall_payoff | sac_outlet | drain_payoff | …],
  speed: {sorcery | instant | flash | activated_only},
  legendary: bool,
  combo_partners: [other card names from this deck this enables],
}
```

The `tags` list is the most important deck-level metadata. The agent uses tags
when picking targets, building turn plans, and assessing what it can still
draw.

### 1.2 Commander Profile

Precompute the commander itself:

- **Color identity** (defines what cards you can play).
- **Casting cost trajectory:** `mv`, `mv+2`, `mv+4`, `mv+6` — tax escalates by
  +2 generic per prior cast from command (CR 903.8). Stop recasting once it's
  no longer cost-effective.
- **Role:** is the commander your *primary wincon* (Voltron, combo enabler) or
  a *value engine* (card draw, ramp)? This drives whether resolving it is
  mandatory vs. optional each game.
- **Replacement value:** if removed, can the deck still win? If no →
  protection budget must be high (Lightning Greaves, Heroic Intervention,
  bounce-to-hand outs).

### 1.3 Card Index

Build lookup tables once:

- `by_tag[tag] -> [cards]`.
- `by_mv[n] -> [cards]` (for curve sequencing).
- `by_color_pip[symbol] -> [cards]` (for mana fixing).
- `mana_producers -> [cards with intrinsic mana abilities]`.
- `ramp_cards`, `draw_engines`, `tutors`, `counterspells`,
  `protection_spells`, `removal_spot`, `removal_mass`, `combo_pieces`.
- `enables_alt_win` — Approach of the Second Sun, Thassa's Oracle, etc.

### 1.4 Archetype Classification

Match the deck against the 11 archetypes in
[`COMMANDER-STRATEGY.md`](./COMMANDER-STRATEGY.md) §5 (Voltron, Aristocrats,
Stax, Tokens, Spellslinger, Reanimator, Group Hug, Landfall, +1/+1 Counters,
Combo, Control). The matched archetype unlocks:

- **Default kill plan** — the 1–3 sequences the agent treats as "win the
  game" sequences.
- **Default threat priority** — which opponents become the agent's main
  target (e.g., a stax deck targets combo decks first; a voltron deck
  targets the player who can blow up its commander).
- **Default sequencing template** — what to play turns 1–4
  (`COMMANDER-STRATEGY.md` §4.2).

If the deck fits two archetypes (e.g., Tokens + Aristocrats), pick the
**stronger finisher** as the primary plan and treat the other as backup.

### 1.5 Power Level Self-Assessment

Estimate the deck's average **goldfish kill turn** (turn the deck wins with no
opposition):

| Goldfish turn | Bracket | Behavior |
|---------------|---------|----------|
| ≤ T4 | cEDH | Race; assume opponents have free counterspells. |
| T5–T7 | High power | Set up fast; expect strong interaction. |
| T8–T10 | Mid power | Play long-game; lean on card advantage. |
| T11+ | Casual / battlecruiser | Politics > raw speed. |

This bracket sets **how patient** the agent is and **how aggressively** it
mulligans (§3.4) and trades resources.

---

## 2. State Query Schema

Every time the agent gets priority or is asked for a choice, it must
re-query. Don't trust cached state — opponents may have made hidden moves
during ability resolution.

Minimum fields the agent must read each tick:

```text
GameState {
  turn_number, active_player, phase, step, stack: [StackObject],
  players: [Player {
    seat, name,
    life, poison, commander_damage_received_from: { commander_id -> int },
    mana_pool: { W, U, B, R, G, C, X },
    hand_size, library_size, graveyard: [Card], exile: [Card],
    battlefield: [Permanent],
    command_zone: [Card],
    has_priority: bool, is_active: bool,
    revealed_information: [...],   # delve, foretold, etc.
    available_mana_estimate: int,  # untapped lands + rocks
  }],
  me: index into players,
  opponents: [indices],
  emblems: [Emblem],
  monarch: int | None,
  initiative: int | None,
  day_or_night: 'day' | 'night' | None,
  the_ring_tempted: { player_idx -> level },
  city_blessing: { player_idx -> bool },
  effects: [ActiveContinuousEffect],   # auras, anthems, replacements
  triggered_abilities_pending: [Trigger], # waiting for stack placement order
}
```

`Permanent` includes: `controller, owner, types, subtypes, power, toughness,
counters, attached_to/attachments, tapped, summoning_sick, marked_damage,
abilities, current_zone_timestamp`.

### 2.1 Derived metrics (recompute on demand)

| Metric | How to compute |
|--------|----------------|
| `available_mana(player)` | sum mana from untapped lands + tappable mana rocks + mana dorks (ignore summoning-sick dorks). |
| `available_mana_at_instant_speed(player)` | same but exclude sources that can only tap on the player's own turn (rare) and reserve mana for any "mandatory upkeep" costs. |
| `removal_in_hand(player)` | the agent only knows its own hand; for opponents, infer from open mana, deck archetype hints, and prior plays. |
| `clock(player)` | turns from now until that player would deal lethal to a chosen target with their current board, ignoring blockers. |
| `defended_clock(target)` | clock minus blockers currently available to target. |
| `combo_proximity(player)` | count of named combo pieces visible (battlefield + graveyard) for an inferred combo of theirs; high = imminent. |
| `archenemy_index(player)` | normalized score of "how scary they look": commander on board (+1), 2+ engines (+1 each), 5+ power on board, infinite-mana enabler visible, etc. |

`archenemy_index(me)` is what other tables use against you. Keep it low
unless you're closing the game *this turn*.

---

## 3. Pre-Game Decisions

### 3.1 Threat Model (Updated Continuously)

For each opponent, maintain a record:

```text
OpponentModel {
  seat, life, commander, identified_archetype, suspected_combos,
  cards_played: [Card], cards_in_graveyard: [Card],
  open_mana_history: [int per turn end],
  revealed_hand_info: [...],
  removal_used: [{ name, target, turn }],
  threats_on_board: [Permanent],
  political_disposition: 'hostile' | 'neutral' | 'cooperative',
  reliability: 0..1,   # honors deals?
}
```

Update after every visible action:

- A card played → infer archetype/wincons.
- Open mana at end of turn → infer counterspell / protection.
- Removal used → opponent's interaction count drops by 1.
- A tutor activation → adversary's combo proximity rises sharply.

Anytime an opponent's combo proximity or archenemy index spikes,
**re-rank threat priority** before your next priority pass.

### 3.2 Seat Awareness

In a 4-player free-for-all your two **flanking opponents** matter most: the
player to your right (you attack first into them when going around) and to
your left (they attack you last in the rotation, often after you're tapped
out).

- **Right opponent** is your easiest attack target.
- **Left opponent** can punish you most after you tap out — be cautious
  about open-mana commitments before their turn.

### 3.3 Mulligan Decision (London Mulligan, CR 103.5; free first mull common)

Algorithm:

```text
def keep_or_mull(hand, mulls_taken, archetype, is_on_play):
  lands           = count_lands(hand)
  ramp            = count(hand, tag='ramp', mv<=2)
  early_action    = count(hand, mv<=3) + ramp
  reaches_4_mana  = lands + ramp >= 4
  produces_colors = covers_first_3_turns_of_costs(hand)
  has_path        = any(card in hand has tag in {'wincon','combo_piece','draw'})

  # Hard mulligans
  if lands < 2 and 'Sol Ring' not in hand:           return MULL
  if lands > 5 and ramp == 0:                        return MULL
  if not produces_colors:                            return MULL
  if archetype == 'Combo' and not has_combo_seed(hand): return MULL

  # Soft criteria
  score = (
     2*ramp + 2*has_path + early_action +
     (1 if 'Sol Ring' in hand else 0) +
     (1 if reaches_4_mana else 0)
  )
  threshold = 5 - mulls_taken          # be more lenient with each mull
  return KEEP if score >= threshold else MULL

def bottom_card(hand_to_bottom):
  # Bottom in this order until we've bottomed mulls_taken cards:
  priority = [
    duplicates_of_categories,         # excess lands beyond 5 / excess ramp
    high_mv_no_ramp_to_cast_them,
    narrow_silver_bullets_for_unseen_threats,
    color-poor lands when fixing solid,
  ]
```

Stop mulling at **5 cards** (after bottoming 2) — going further hurts more
than it helps in Commander. With a free first mull (most casual pods),
mulligan more aggressively for the first redraw.

### 3.4 Going First vs. Going Last

If random seat is fixed: no choice. Otherwise:

- **Go first** with low-curve, proactive, combo, or stax decks (you set the
  pace).
- **Go last** with reactive control or grindy decks (extra draw, more
  information).

If the rule "going first skips first draw step" applies (CR 103.8a), prefer
draw-step reset for control; ignore it for combo.

---

## 4. Action Enumeration on Priority

Every priority window, enumerate **legal** actions before scoring them:

```text
def enumerate_legal_actions(state):
  actions = [PASS_PRIORITY]
  if state.priority_holder != me: return actions

  # Special actions (no stack, CR 116)
  if can_play_land(): actions += [PlayLand(L) for L in lands_in_hand]
  if can_turn_face_up(): ...
  if can_use_companion(): ...

  # Sorcery-speed gate
  is_sorcery_speed = (state.active_player == me
                      and state.phase in ('main_pre','main_post')
                      and state.stack.empty())

  # Spells from hand
  for c in hand:
    if can_cast(c, is_sorcery_speed):  actions.append(CastSpell(c, choices))

  # Commander cast from command zone
  if commander_in_command_zone() and can_pay(commander_cost_with_tax()):
    if commander.is_instant or commander.has(Flash) or is_sorcery_speed:
      actions.append(CastCommander())

  # Cast from other zones (flashback, escape, foretell, etc.)
  for c in {graveyard, exile} that allow casting:
    actions += alt_cast_actions(c)

  # Activated abilities
  for p in my_permanents + my_graveyard_with_active_abilities:
    for ab in p.activated_abilities:
      if can_pay(ab.cost) and ab.timing_ok(is_sorcery_speed):
        actions.append(Activate(p, ab))

  # Loyalty abilities
  for pw in my_planeswalkers_not_yet_activated_this_turn:
    if is_sorcery_speed:
      for la in pw.loyalty_abilities:
        if can_pay(la.cost):
          actions.append(LoyaltyAbility(pw, la))

  # Triggered/static reordering choices, mode/target choices on stack: handled in §7
  return actions
```

When `can_cast` is checked, factor in cost modifiers (Convoke, Improvise,
Delve, kicker, additional costs) and verify legal targets exist. Don't enumerate
spells with no legal target unless the spell has "up to" wording (CR 115.4).

---

## 5. Mana Planning

### 5.1 Mana availability check

`can_pay(cost)` must respect color and source constraints. Algorithm:

```text
def can_pay(cost, sources):
  # 1. Solve a bipartite match: each colored pip must consume a source
  #    that produces that color (or hybrid that includes it).
  # 2. Generic pips can be filled by any leftover source.
  # 3. Reserve sources locked to a specific use (e.g., Cabal Coffers needs Swamp count).
  # 4. Account for cost modifiers (Goblin Electromancer −{1} for instants/sorceries).
  return solve(cost, sources)
```

**Always tap colored sources for colored pips first**, generic pips last —
tapping a dual land before a basic is a common micro-error that loses options
later in the turn.

### 5.2 Floating mana

If you tap mana and the phase ends with mana in pool, it empties (CR 106.4)
and (in casual rules; modern rules) you take 1 damage **only** if you used to
play under old mana-burn (no longer a rule, CR 106.4b). Floating mana into
the next phase requires intent: announce it (or in a digital interface, add
to mana pool intentionally before phase change).

### 5.3 Reserve mana for instants

When deciding how much mana to commit on your main phase, reserve enough
for:

- The cheapest **counterspell** in hand (if any).
- The cheapest **protection** spell for your most valuable permanent.
- Any **trigger you must pay** later in the turn (e.g., Phyrexian Mana
  upkeep cards, Tangle Wire upkeep, Sword of Feast and Famine you'll
  attack with).

Don't reserve more than ~2 mana past turn 5 unless you actually have an
instant in hand — telegraphed open mana reduces opponents' aggression
*against you* but also reduces the times you can deploy threats.

### 5.4 Commander tax discipline

Cast the commander only if **`benefit ≥ tax`**. Specifically:

```text
def should_cast_commander_from_command():
  base_cost = commander.mv + 2 * times_cast_before
  expected_value_this_game = E[mana of effect produced]
  if base_cost > 8 and not commander.is_primary_wincon:
    return False
  if commander_will_die_on_resolution():           # removal up
    return base_cost <= 4 and commander.has_etb_value
  return True
```

Voltron decks must accept the tax 1–2 times; combo decks should usually
**not** recast a non-combo commander after the second cast unless it's the
combo line itself.

---

## 6. Action Scoring (the central rubric)

For every legal action, compute a numeric score; pick the maximum. The score
combines short-term utility, long-term position, and threat-management
considerations.

```text
def score(action, state):
  return (
      W_TEMPO       * tempo_delta(action, state)
    + W_CARDADV     * card_advantage_delta(action, state)
    + W_LIFE        * life_delta(action, state)
    + W_BOARD       * board_position_delta(action, state)
    + W_COMBO       * combo_progress_delta(action, state)
    + W_THREAT_MGMT * threat_management_delta(action, state)
    + W_POLITICS    * politics_delta(action, state)
    + W_INFO        * information_value(action, state)
    - W_RISK        * downside_risk(action, state)
    - W_PAINT       * archenemy_paint_added(action, state)
  )
```

Suggested default weights for **mid-power Commander**:

```
W_TEMPO=1, W_CARDADV=1.5, W_LIFE=0.4, W_BOARD=1.2,
W_COMBO=2, W_THREAT_MGMT=1.5, W_POLITICS=0.5,
W_INFO=0.3, W_RISK=1.5, W_PAINT=1
```

For **cEDH** raise `W_COMBO` to 3 and `W_PAINT` to 0.3 (you're racing). For
**casual battlecruiser** raise `W_POLITICS` to 1.5 and `W_PAINT` to 1.5.

### 6.1 Definitions of each term

- **`tempo_delta`**: net mana value swing this turn cycle. Casting a 4-mana
  removal spell on a 6-mana threat is +2 tempo.
- **`card_advantage_delta`**: cards drawn − cards spent + opponents' cards
  forced out (counter, discard, sweep). 1 cantrip = 0; Rhystic Study
  resolved ≈ +3 over its lifetime.
- **`life_delta`**: damage dealt − damage prevented + life gained, weighted
  more highly under 15 life.
- **`board_position_delta`**: change in `attack_power_total +
  defensive_blockers` for me vs. each opponent.
- **`combo_progress_delta`**: distance to assembling a kill, counted as
  `1 / (cards_still_needed)`. Resolving a tutor that fetches a missing piece
  is huge.
- **`threat_management_delta`**: change in `max(opponent.archenemy_index)`.
  Removing the scariest threat reduces the future damage you'll absorb.
- **`politics_delta`**: change in expected goodwill from each opponent.
  Killing the player who hates you most ≈ +0; killing the leader for the
  table ≈ +1.
- **`information_value`**: cards revealed about opponents' hands or libraries.
- **`downside_risk`**: probability-weighted bad outcomes (counterspell,
  removal trade, dying to aggro).
- **`archenemy_paint_added`**: how scary you look after this action. A
  resolved Smothering Tithe ≈ +2 paint; a 1/1 token ≈ +0.

### 6.2 Prune obviously bad actions before scoring

Save compute by short-circuiting:

- Don't enumerate casting a 6+ MV spell if you can't reach 6 mana this turn.
- Don't target your own permanents with destruction unless that creates net
  value (sac-replace, save from worse fate, trigger ability).
- Don't activate `{T}: discard a card` engines when hand is already optimal.

---

## 7. Resolving In-Game Choices

The game asks the agent dozens of "make a choice" questions. Handle them with
domain-specific rules.

### 7.1 Mode selection (Charms, modal DFCs)

Pick the mode whose **execution** maximizes `score()` *given the choices*.
Default tie-breakers:

1. Affect the highest-archenemy-index opponent if removal.
2. Draw card if no removal target is high-value.
3. Save flexible modes for post-resolution if "choose two" allows it.

### 7.2 Targeting

When a spell/ability needs a target:

```text
def pick_target(spell, candidates):
  if spell.is_buff_or_protection: return [most_valuable_self_perm]
  if spell.is_removal: return [highest_archenemy_threat]
  if spell.is_burn: return [best_kill_target_or_face_for_clock]
  if spell.is_card_draw_to_player: return [me]
  if spell.is_pump_to_attacker: return [my_creature_with_best_unblocked_path]
```

Pick the target that **denies the biggest opponent gain** or **enables the
biggest agent gain**, not the target that "looks scariest in isolation."

### 7.3 Stacking trigger order (CR 603.3)

When multiple of *your* triggers go on the stack at once, *you* choose order.
Pick the order that gives the most flexibility:

- Triggers that **search** or **draw** first (so you have information for
  later triggers).
- **Damage triggers** before **state-based actions** that would kill the
  source.
- **ETB triggers that target** before **ETB triggers that don't** (target
  pool changes during resolution).

When opponents' triggers stack with yours simultaneously, the active
player's triggers go on the stack first, then non-active in turn order
(CR 603.3b). The agent should compute the resulting resolution order and
plan responses accordingly.

### 7.4 Replacement effect ordering (CR 616)

When multiple replacements apply to the same event and you control the
choice, order so that the **last-applied replacement is the most valuable**
(it sees the others' modifications). Common cases:

- ETB tapped + ETB with counters → resolve "with counters" last so counter
  doublers (Doubling Season) layer correctly.
- Damage prevention + lifelink → take prevention first to keep a buffer,
  unless you need the lifegain trigger.

### 7.5 Mulligan for opening hand: §3.3.

### 7.6 Surveil / Scry decisions

```text
def scry_or_surveil(top_cards):
  for card in top_cards:
    if card is land and you_already_have_4+_lands: send_to_bottom_or_gy
    elif card is dead_in_matchup: send_to_bottom_or_gy
    elif card is on_curve_for_next_turn: keep_on_top
    elif surveil and card is reanimator_target: send_to_graveyard
    elif card is wincon_when_you_dont_have_one_in_hand: keep_on_top
```

### 7.7 Tutoring

When activating a tutor, choose the card that **maximizes
`score(future_play(card))`** given your current state and the *next 2 turns*'
expected mana. Specifically:

1. If a single missing combo piece would let you win this turn or next →
   tutor it.
2. Otherwise, tutor the highest-impact card you can cast next turn given
   available mana.
3. If under pressure, tutor an answer (counterspell or wipe).

Don't tutor for a card you can't cast within 2 turns — by then the tutor
target may be obsolete.

### 7.8 Discarding to hand size

At cleanup (CR 514.1) discard down to 7. Discard:

1. Lands you don't need.
2. Duplicates by function (3rd ramp spell with 3 already on board).
3. Narrow answers no longer relevant.
4. **Never discard your wincon or your last counterspell** unless replaced.

### 7.9 Information-only choices (revealing cards, looking at hands)

Always pay the small cost to **gain information** unless it telegraphs
strategy or is symmetric (gives opponents the same info). Information
asymmetry is value.

---

## 8. Combat Decision Procedures

Combat is the most error-prone subsystem. Use these step-by-step procedures.

### 8.1 Beginning of combat (CR 507)

- Take any combat-only buff actions whose effect benefits attackers
  (e.g., First Strike auras).
- Activate "until end of turn" boosts that aren't combat-tricks reliant on
  surprise.
- Decide attack plan in §8.2 before Declare Attackers (CR 508).

### 8.2 Attack decision

```text
def declare_attackers(state):
  attackers = []
  for c in my_untapped_creatures_without_summoning_sickness_or_with_haste:
    if c.has(Defender) and not c.attacks_anyway: continue
    if c.must_attack_a_player_each_combat:        attackers.append(c); continue
    target = pick_target_player_or_pw(state, c)
    if target is None: continue
    if expected_outcome(c, target).agent_value > 0:
      attackers.append((c, target))
  return attackers

def pick_target_player_or_pw(state, c):
  # Prefer:
  # 1. A planeswalker about to ult (kill it before owner's next turn).
  # 2. The opponent with highest archenemy_index, if defended_clock
  #    favors me.
  # 3. The opponent to my left whose untapped lands include counter mana
  #    (deplete their resources before my next turn).
  # 4. The weakest defender if I want to chip damage and deal commander damage.
  # NEVER attack the opponent who is most likely to retaliate fatally.
```

### 8.3 Threat-of-attack vs. attacking

If attacking with **all** creatures leaves you exposed to a counter-attack
that could lethal you, hold attackers as blockers. The implicit threat of an
attack is often worth more than the actual swing, especially with
vigilance.

Rule of thumb: keep enough toughness on board to absorb the largest
single-turn output any opponent can produce. In landfall or token decks that
might be 30+ damage; in control mirrors it might be 0.

### 8.4 Block decision

Blocking is a *constrained optimization* (CR 509):

```text
def declare_blockers(attackers, my_creatures):
  # Step 1: enumerate all assignment maps.
  # Step 2: for each, simulate combat damage given:
  #   - first/double strike steps
  #   - deathtouch (any nonzero is lethal, CR 702.2)
  #   - trample (excess to player, CR 702.19)
  #   - lifelink (post-damage life gain)
  #   - protection/indestructible/menace
  # Step 3: score outcome:
  #   damage_to_me + creatures_lost*power_value
  #     - creatures_killed_attacking*power_value
  #     - commander_damage_taken*5  # commander damage is unbounded across game
  # Step 4: choose minimal-loss assignment.
```

Heuristics:

- **Always block to prevent commander damage** if a single hit pushes you
  past 21 from that commander (CR 903.10a).
- **Chump-block the largest attacker** with your smallest creature when the
  attacker has trample only if the trample damage ≤ alternative damage.
- **Multi-block** to kill an indestructible attacker by tapping it
  (deathtouch + multi-block kills via state-based even when not destroyable
  in some cases — but usually indestructibles survive; consider exiling
  instead).
- **Don't block** when the attacker is a deathtouch creature unless you can
  kill it in the first-strike step (no return damage) or you must block to
  survive.

### 8.5 Combat tricks timing (CR 510)

The key window for instant-speed pumps and removal is **between Declare
Blockers and Combat Damage Step**, after blockers are locked. Rules:

- A pump/removal cast before Declare Blockers may cause the opponent to
  block differently — only do this when you *want* to influence their block.
- Save protection until after Declare Blockers if you want to change the
  fight outcome but not the block decisions.

### 8.6 Commander damage accounting

Track per-opponent: `cmd_damage_received_from(commander_X, by_player_Y)`. At
21+, that player loses (CR 903.10a). When attacking:

- If a single hit puts a player past 21, **expect them to use their best
  removal**.
- Voltron pilots should split attacks across opponents until close to lethal
  on multiple, then alpha-strike.

---

## 9. Stack & Priority Management

### 9.1 When to hold priority

After casting a spell, you can either:

- **Pass priority** immediately → opponents may respond before any of your
  abilities trigger.
- **Hold priority** → cast another spell that responds to the one you just
  cast (rare in Commander).

When **opponents** cast a spell:

- **Counter on cast**: cheapest mana, opponent has no info on what you held.
- **Counter on resolution attempt** (e.g., Stifle on a trigger): only if the
  trigger is the actual problem.
- **Wait to see targets**: if the spell hasn't fully announced, you have no
  decision yet.

### 9.2 Counterspell decision

```text
def should_counter(spell, state):
  # Free counters (Force of Will, Fierce Guardianship, Pact of Negation)
  # can be used reactively at low marginal cost.
  threat_score = simulate_resolve(spell)
  cost_score   = cards_used + mana_used
  return threat_score >= cost_score + safety_margin
       and not other_opponent_will_counter_for_me
```

Don't counter what an opponent will counter. Don't counter the third-best
spell in a sequence — save for the actual win-attempt.

### 9.3 Removal timing

- **Spot removal**: cast at the latest beneficial moment. Usually opponent's
  end step (after they tapped out, before your draw step) so they can't
  redeploy with the same mana.
- **Mass removal (sweepers)**: cast on your turn with a follow-up plan, OR
  in response to an opponent's overcommitment (post-attack with their
  creatures tapped is ideal in some cases).
- **Sacrifice / edict effects**: most effective just *before* the opponent
  casts a buff or activates a tutor; punishes their tap-out.

### 9.4 Stack response to triggers

Triggered abilities go on the stack at the next priority opportunity (CR
603.3). The window between trigger and resolution is your chance to:

- Counter the trigger (Stifle, Disallow).
- Remove the source so that "last known information" applies and the trigger
  fizzles in cases where the trigger references the source's current state
  on the battlefield (CR 603.10 — most "When this dies" triggers still
  resolve with last-known info; check the ability text).

---

## 10. Multi-Turn Planning

### 10.1 Turn N+1, N+2 simulation

At the start of each of your turns, simulate the next 2 turns under
no-interference assumptions:

```text
plan = []
for t in [now, now+1, now+2]:
  drawable_mana = current_mana + expected_land_drops + expected_ramp
  best_play = pick_best_n_card_sequence(hand_estimate, drawable_mana)
  plan.append(best_play)
return plan
```

Use this to:

- Decide whether to **hold** a card or play it now.
- Decide which **lands** to play (color sequencing).
- Identify the **earliest combo turn** and budget protection for it.

### 10.2 Sequencing lands

Land order matters:

- Tap-lands first (they only enter tapped). Don't drop a tap-land on the
  turn you need it untapped.
- Fetches when you actually need fixing or graveyard fuel; otherwise
  delay so you can deck-thin late.
- Utility lands (Bojuka Bog, Wasteland equivalent) when their effect is
  needed.
- Basics for shock/dual interactions where life loss matters.

### 10.3 Holding action vs. deploying

Default rule: **deploy a threat each turn** if it reduces opponents' clocks
on net AND doesn't overcommit into a sweeper. Specifically:

- Don't go from 3 creatures to 6 in one turn before your finisher is
  available — that's a sweeper magnet.
- Don't sandbag a 6-mana threat on turn 8 if you have nothing in turns 9–10
  that's better.

### 10.4 The "don't be the threat" rule

If `archenemy_index(me) >= max(archenemy_index(opp) for opp in opponents) +
2`, you are the table's primary target. Adjust:

- Stop deploying engines visibly.
- Cast removal on others' threats so the threat baton passes.
- Trade a small political concession ("I'll *Path* their commander if you
  swing at the green deck") to reduce paint.

---

## 11. Politics

Politics is a real resource. Apply the rules from
[`COMMANDER-STRATEGY.md`](./COMMANDER-STRATEGY.md) §3.3 plus these agent-specific rules.

### 11.1 What the agent should say (when communication is allowed)

| Situation | Template |
|-----------|----------|
| Asked "Will you attack me?" | "If you don't [specific action], I'll go elsewhere this turn." |
| Persuading away from you | "If they kill me, [other player] wins next turn — see [evidence]." |
| Forming a temp alliance | "I'll trade my removal for your counter on the [combo player]'s next tutor." |
| Refusing | Decline politely, no need to justify. |
| Declaring a kingmaker concern | "I can't win — I will play whatever extends the game." |

Rules:

- **Make narrow, time-bounded promises.** "This turn" / "next turn" — never
  whole-game alliances.
- **Honor deals** unless breaking wins the game on the spot (one-shot ethic).
- **Use evidence**, not pleas. Cite cards, life totals, mana availability.
- **Don't reveal your own combo-proximity**. If asked, deflect with what's
  visible and answer questions about *opponents'* hands.

### 11.2 Threat assessment shared with the table

When the table asks "who's winning?", honestly point to the opponent with
the highest combo-proximity — even if that's politically expensive — *as
long as it isn't you*. If it is you, deflect ("Hard to say, [other player]
has a lot of mana up").

### 11.3 Don't kingmake

Apply the rule from [`COMMANDER-STRATEGY.md`](./COMMANDER-STRATEGY.md) §3.4: if you cannot win,
make the choice that *prolongs the game*. If no choice prolongs, prefer the
opponent with the lowest archenemy_index (the table's "fairest" winner).
Never decide based on who beat you most recently.

---

## 12. Win-Condition Execution

Once you can identify a turn that wins, switch from "long game" to "execute"
mode. The execution checklist:

1. **Mana check**: Can you cast every piece on the chain, with cost
   reductions applied?
2. **Stack check**: Are all opponents' counterspells accounted for? Count
   *open mana* and *known counters* and assume **at least one unknown
   counter** in casual; assume *two* in cEDH.
3. **Protection check**: Do you have enough Veil of Summer / Silence /
   Force of Will / Heroic Intervention to push through?
4. **Sequencing check**: Cantrips before commitment; cost reducers
   first (they're ETB-vulnerable); kill spell last.
5. **Trigger order check**: Any triggered ability that matters
   (combo-relevant ETBs) — make sure the order yields the kill.
6. **End-state check**: After the chain resolves, do *all* opponents
   actually lose? Or just one, leaving you 1v2 with no follow-up?
7. **Abort check**: If any of (1)–(6) fails, **don't go off**. Convert this
   turn into setup (cantrip, draw, hold) and try again next turn.

Common failure modes to design out:

- Casting your wincon **before** your protection is up (e.g., Thassa's
  Oracle without a counter for opponents' Silence-style spells in hand).
- Casting tutors so the table can see your wincon coming next turn —
  prefer **flash tutors** or end-of-turn tutors.
- Forgetting commander tax on the kill turn.
- Forgetting that **a spell with 0 legal targets is countered by the rules**
  (CR 608.2b) — the agent must ensure each combo target remains legal.

---

## 13. Endgame and Post-Game

### 13.1 Concession

Concede only when continuing has 0 chance to win AND no chance to extend
others' games meaningfully. In Commander, a player at 1 life with no board
can still draw out of it; don't concede prematurely.

### 13.2 Logging

After each game, record:

```text
GameLog {
  decklist_hash, archetype, seat, turns,
  mulligans: [reason, kept_count],
  major_decisions: [{turn, action, alternatives, outcome}],
  wincon_attempted, wincon_succeeded, killed_by_or_killed,
  threat_misreads: [{opponent, miss_type}],
  political_outcomes: [...],
  result: 'win' | 'loss' | 'draw' | 'kingmade',
}
```

Use logs to **tune weights** in §6 (`W_*` constants) and to update the
opponent reliability model for future games against the same players.

---

## 14. Common Edge Cases (Quick Recall)

| Situation | Correct behavior |
|-----------|------------------|
| Commander dies on the stack (replaced to command zone option, CR 903.9) | Choose command zone if you want to recast cheaply soon; choose graveyard if you have reanimation + want to dodge tax. |
| Commander exiled | Same choice; usually pick command zone, since reanimation from exile is rarer. |
| 21 commander damage from your *partner* commanders | Each partner counts separately (CR 903.10a). |
| Two legendary copies (legend rule) | You choose which to keep (CR 704.5j). Keep the higher-utility one. |
| Stack of 4 ETB triggers from a single ETB (yours + opponents') | AP order first (you, if active), then NAP in turn order. Build the dependency tree before assigning order. |
| Mana ability vs. spell timing | Mana abilities don't use the stack (CR 605.3) — opponents can't respond between activation and mana production. Use this to surprise-pay for spells. |
| "Can't be countered" spell | Don't waste a counter. Use removal post-resolution instead. |
| Hexproof target | Use sacrifice/edict, board wipes, or "destroy/exile all" effects. |
| Indestructible target | Use exile, sacrifice, −X/−X, "send to library/hand", or counter-it-on-cast. |
| Empty library imminent (yours) | Win immediately or stop self-mill; consider Thassa's Oracle / Laboratory Maniac if in deck. |
| Empty library (opponent) | Force them to draw on their turn — ping, force-draw, or just wait through their draw step. |
| Stax piece both sides hurt by | Only deploy if your deck dodges it materially better than theirs. |
| Storm count check before going off | Count *spells cast this turn by all players* (CR 702.40). Includes opponents' spells. |
| Cascade target choice | When you can cast it, do unless it disrupts your line; "may cast" means optional (CR 702.85b). |
| Replacement effects you control | Order to maximize your final state (CR 616.1). |
| Two of your triggers on stack | LIFO resolution; choose order accordingly. |

---

## 15. Pseudocode: Full Turn for the Agent

```text
def take_my_turn(state):
  begin_phase()                       # untap, upkeep, draw
  process_triggers(state)             # any stack response opportunities
  reassess_threat_model(state)        # §3.1
  plan = make_turn_plan(state)        # §10.1, choose best 2-turn sequence

  # Main phase 1
  while priority_held and there_is_a_useful_action:
    a = best_action(state)
    if a == PASS: break
    execute(a)
    state = query_full_game_state()

  # Combat
  if should_enter_combat(state):
    declare_attackers(state)          # §8.2
    handle_responses()
    declare_blockers(state)           # §8.4
    handle_responses()
    deal_combat_damage(state)
    handle_post_combat_triggers()

  # Main phase 2
  while priority_held and there_is_a_useful_action:
    a = best_action(state)
    if a == PASS: break
    execute(a)
    state = query_full_game_state()

  # End phase
  pass_priority()
  if I_must_discard(): discard_to_hand_size(state)   # §7.8
  remove_until_end_of_turn_effects()
  end_turn()

def take_opponent_turn(state):
  while game_continues:
    state = query_full_game_state()
    if priority_holder == me:
      a = best_action(state)
      execute(a)
    else:
      wait_for_priority_or_choice()
```

---

## 16. Quick-Reference Decision Cards

### 16.1 Should I cast this threat now?

```
YES if:
  - I gain at least +1 tempo or +1 board
  - I won't be the table's #1 paint after casting
  - The threat survives the most likely removal in opponents' hands
  - I have a follow-up if it dies

NO if:
  - I'm already archenemy and casting it makes it worse
  - The next turn is a sweeper turn for an opponent (read: 4+ mana on the
    end step of the wipe-color player)
  - The threat is win-relevant and I have no protection up
```

### 16.2 Should I attack X with creature C?

```
YES if:
  - I survive the counter-attack from each opponent
  - The damage progresses my plan (combat damage trigger, commander
    damage, racing)
  - X cannot remove C in response without losing tempo

NO if:
  - X has open mana for instant-speed removal of C and C is irreplaceable
  - The attack creates the impression of leadership (paint) without ROI
  - I need C as a blocker against another opponent
```

### 16.3 Should I use my counterspell on this spell?

```
YES if:
  - This spell wins the game on resolve
  - This spell removes my critical permanent
  - This spell's resolution shifts the threat order against me

NO if:
  - I have a stronger answer post-resolution
  - Another opponent will counter it for me (with high confidence)
  - The next 2 spells from any player are likely worse and I should save
    counters
```

### 16.4 Should I attempt my combo this turn?

```
YES if:
  - All §12 checklist items pass
  - I can survive failure (have a "Plan B" turn next round)
  - Probability(succeed) >= ~70% (cEDH tolerates 50–60% with stax of own)

NO otherwise — set up another turn.
```

---

## 17. What the Agent Must NEVER Do

A short list of always-wrong actions in Commander:

1. **Never tap out** if you have a counterspell in hand and an opponent has
   shown a clear win line.
2. **Never break the legend rule unintentionally** — if your commander is
   already on the battlefield, don't recast a copy without a plan.
3. **Never miss commander damage tracking** — opponents count separately.
4. **Never target your own indestructible/hexproof permanent with
   destruction** when you intended to remove an opponent's; verify target
   chosen.
5. **Never make a binding promise that costs you the game to keep.** Politics
   is best-effort, not contractual.
6. **Never play a board wipe with no follow-up** if any opponent is set up
   to recover faster than you.
7. **Never fail to read the new card** an opponent played — opponents may
   sleeve cards you don't know; query the card text from the interface.
8. **Never reveal hidden info** unintentionally (face-down cards, foretold
   cards, hand-stash effects).
9. **Never forget "may" choices on triggers** — "may" means you can decline,
   often the right call when an effect would hurt you.
10. **Never spend a tutor on a card you can't cast within 2 turns.**

---

## 18. Per-Phase Quick Reference

| Phase / Step | Agent's primary actions |
|--------------|--------------------------|
| Untap (CR 502) | None — turn-based action; verify all permanents that should untap did. |
| Upkeep (CR 503) | Resolve triggers (own first, choose order). Pay any "at the beginning of your upkeep" costs. Activate one-per-turn upkeep abilities (e.g., Bone Miser, Solemnity-style). Open priority window — opponents and you may cast instants. |
| Draw (CR 504) | Draw 1 (skip if first turn going first in some formats). Apply replacement draws. |
| Pre-combat Main (CR 505) | Play land (if not yet); deploy ramp; deploy threats; cast sorceries you must cast pre-combat (e.g., haste enablers, attack-dependent buffs); plan combat. |
| Beginning of Combat (CR 507) | Activate "beginning of combat" triggers; cast pump that affects attack declaration. |
| Declare Attackers (CR 508) | Choose attackers per §8.2; pay attack costs (Propaganda, etc.). |
| Declare Blockers (CR 509) | (As defending player only) per §8.4. |
| First-strike damage step (CR 510 if applicable) | Resolve first/double strike. Window for instants. |
| Combat damage step (CR 510) | Assign and deal damage. Window for instants between assignment and dealing — rarely useful. |
| End of Combat (CR 511) | Last chance for "during combat" triggers. |
| Post-combat Main (CR 505) | Deploy permanents you didn't want exposed pre-combat; cast extra-turn spells; reassess plan. |
| End Step (CR 513) | "At the beginning of the end step" triggers. **This is the prime window for opponents' spot removal/cantrips/tutors.** |
| Cleanup (CR 514) | Discard to 7. Damage clears. Until-end-of-turn effects end. **No priority unless triggers happen** (CR 514.3a). |

---

*End of agent playbook. Pair with the comprehensive rules and the card-text
parsing guide for full operational coverage.*
