# Magic: The Gathering — Rules of Play and Optimal Strategy

## Executive Summary

Magic: The Gathering (MTG) is a two‑or‑more‑player trading card game published by Wizards of the Coast in which each player builds a deck representing a "planeswalker" wizard and tries to reduce opponents from a starting life total (20 in most formats, 40 in Commander) to 0. Each turn flows through five phases — **Beginning, Precombat Main, Combat, Postcombat Main, and Ending** — and spells/abilities resolve through a last‑in‑first‑out structure called **the stack**, gated by **priority**[^1][^2]. Cards come in eight types (Land, Creature, Instant, Sorcery, Enchantment, Artifact, Planeswalker, Battle), and the game's five colors — **W**hite, Bl**u**e, **B**lack, **R**ed, **G**reen — each have signature strengths and a defined "color pie" of weaknesses[^3][^7]. Optimal play is built on four pillars: a **disciplined mana curve**, **card advantage**, **tempo**, and **threat assessment**, applied through a coherent deck archetype (aggro, midrange, control, combo) and refined through **mulligans**, **sideboarding**, and bluffing of priority/open mana[^4][^5][^6].

This report has two focus areas as requested: **(1) How to actually play the game** (rules, turn structure, card types, combat, win conditions, formats), and **(2) How to play well** (deckbuilding, the color pie, tempo/card advantage, the stack, mulligans, sideboarding, threat assessment, and archetype‑specific guidance).

---

# Part 1 — How to Play Magic: The Gathering

## 1.1 Objective and Setup

- **Goal:** Reduce each opponent's life total from the starting value (20 in 1v1 constructed; 40 in multiplayer Commander) to **0**, or win by an alternate condition (e.g., your opponent is forced to draw from an empty library, takes 10+ "commander damage" from a single commander in Commander, or a card with a specific "you win the game" effect resolves)[^1][^8].
- **Setup:** Each player shuffles their deck (the **library**), draws an opening hand of **7 cards**, and may take **mulligans** (see §2.5). The non‑active player is decided randomly; the player going first **skips their first draw step** in most formats[^2][^5].

## 1.2 The Eight Card Types

| Card Type | When Cast | Stays on Battlefield? | Primary Role |
|-----------|-----------|----------------------|--------------|
| **Land** | Your main phase, once per turn (not "cast") | Yes | Produces **mana** when tapped |
| **Creature** | Your main phase (sorcery speed) | Yes | Attacks and blocks; has Power/Toughness |
| **Instant** | **Anytime you have priority** | No (goes to graveyard) | Reactive effects, removal, counters |
| **Sorcery** | Your main phase, empty stack | No | One‑shot powerful effects |
| **Enchantment** | Your main phase | Yes | Persistent effects (Auras attach to permanents) |
| **Artifact** | Your main phase | Yes | Usually colorless, versatile (Equipment, mana rocks, etc.) |
| **Planeswalker** | Your main phase | Yes (with loyalty counters) | Activate one loyalty ability per turn; can be attacked |
| **Battle** | Your main phase | Yes (until defeated) | Newer type; has defense counters, transforms when defeated |

Sources: [^3]. Battles were introduced in *March of the Machine* (2023) and most commonly appear as the **Siege** subtype, where the controller chooses an opponent to "protect" (defend) the battle[^3].

## 1.3 Mana and Casting Spells

- **Mana** is produced by tapping lands (and other mana sources). Each basic land taps for one mana of a specific color: **Plains→W, Island→U, Swamp→B, Mountain→R, Forest→G**[^7].
- A spell's **mana cost** (top right of the card) shows colored and/or generic mana required. Tap lands to pay the cost, then put the spell on the **stack**.
- You may play **at most one land per turn** from your hand[^1][^2].
- **Sorcery speed** = your own main phase with an empty stack. **Instant speed** = anytime you have priority (including during opponents' turns and in response to other spells)[^2][^3].

## 1.4 Turn Structure (Comprehensive Rules §500–§514)

Every turn proceeds through five phases, several broken into steps. The **active player** gets priority first in each step except Untap and (usually) Cleanup[^2].

```
┌─────────────────────────────────────────────────────────────────────┐
│                        ONE TURN                                     │
│                                                                     │
│  1. BEGINNING       2. PRECOMBAT     3. COMBAT       4. POSTCOMBAT  │
│  ─ Untap            MAIN PHASE       ─ Beginning     MAIN PHASE     │
│  ─ Upkeep          (sorcery speed)   ─ Declare Atk  (sorcery speed) │
│  ─ Draw                              ─ Declare Blk                  │
│                                      ─ Combat Dmg    5. ENDING      │
│                                      ─ End Combat    ─ End Step     │
│                                                      ─ Cleanup      │
└─────────────────────────────────────────────────────────────────────┘
```

### 1.4.1 Beginning Phase
1. **Untap step** — Untap all your tapped permanents. **No player gets priority**; spells cannot be cast[^2].
2. **Upkeep step** — "At the beginning of upkeep" triggers go on the stack. Players may cast instants/activate abilities.
3. **Draw step** — Active player draws **one card** (skipped on turn 1 for the player going first).

### 1.4.2 Precombat Main Phase
- Play a land (if you haven't already this turn). Cast sorceries, creatures, enchantments, artifacts, planeswalkers, battles. Activate abilities. Cast instants[^1][^2].

### 1.4.3 Combat Phase
1. **Beginning of Combat** — "Beginning of combat" triggers; instants/abilities can be played.
2. **Declare Attackers** — Active player chooses which untapped creatures (without summoning sickness) attack and which player/planeswalker/battle each is attacking. Tapping is the cost of attacking unless the creature has **vigilance**[^9].
3. **Declare Blockers** — Defending player(s) declare blockers. Multiple creatures may block one attacker; one creature can only block one attacker (unless it has special abilities)[^2].
4. **Combat Damage** — Damage is assigned and dealt simultaneously. **First strike** and **double strike** create an extra earlier damage step[^9].
5. **End of Combat** — "End of combat" triggers; last chance for combat tricks.

### 1.4.4 Postcombat Main Phase
- Same rules as the precombat main phase. A common pattern is to attack first with creatures, then deploy new threats post‑combat so opponents see less information before blocking[^4].

### 1.4.5 Ending Phase
1. **End step** — "At the beginning of the end step" triggers fire. Last window for instants before turn ends.
2. **Cleanup step** — Active player **discards down to maximum hand size (7)**. All damage is removed from creatures; "until end of turn" effects expire. Normally no priority is given unless a trigger occurs, in which case players get priority and another cleanup step follows[^2].

## 1.5 The Stack and Priority

- When a spell or activated/triggered ability is announced, it goes on **the stack**. The stack resolves **last‑in, first‑out**: the most recent thing resolves first[^4].
- After casting a spell or ability, **priority** passes. Both players must consecutively pass priority on an empty action for the top of the stack to resolve, or for the phase to advance[^4].
- This is why instants can interact with spells already on the stack — e.g., **Counterspell** responds to an opponent's sorcery before it resolves[^4].

## 1.6 Combat Keywords (Most Important)

| Keyword | Effect |
|---------|--------|
| **Flying** | Can only be blocked by creatures with flying or reach |
| **Reach** | Can block creatures with flying |
| **First strike** | Deals combat damage in a separate, earlier damage step |
| **Double strike** | Deals damage in both the first‑strike step and the regular step |
| **Trample** | Excess damage past blockers carries over to the defending player/planeswalker |
| **Deathtouch** | Any amount of damage dealt to a creature is lethal |
| **Lifelink** | Damage dealt by this creature also causes you to gain that much life |
| **Vigilance** | Doesn't tap to attack — can still block |
| **Hexproof** | Can't be targeted by opponents' spells/abilities |
| **Haste** | Ignores summoning sickness; can attack/tap the turn it enters |
| **Menace** | Can't be blocked except by two or more creatures |

Source: [^9].

## 1.7 Win/Lose Conditions

A player loses when[^1][^8]:
1. Their **life total reaches 0 or less**, OR
2. They are **required to draw a card from an empty library**, OR
3. They have **ten or more poison counters**, OR
4. A spell, ability, or rule (e.g., commander damage in EDH; specific cards like *Approach of the Second Sun*) explicitly states they lose / their opponent wins.

## 1.8 Common Formats

| Format | Deck Size | Card Pool | Notes |
|--------|-----------|-----------|-------|
| **Standard** | 60+ (15 sideboard) | Last ~2 years of sets | Rotates yearly; competitive entry point[^8] |
| **Pioneer** | 60+ (15 sideboard) | Sets from *Return to Ravnica* (2012) onward | Non‑rotating |
| **Modern** | 60+ (15 sideboard) | *8th Edition* (2003) onward | Larger pool, banlist[^8] |
| **Legacy / Vintage** | 60+ (15 sideboard) | Nearly all cards (banned/restricted lists) | Powerful eternal formats |
| **Commander (EDH)** | **Exactly 100**, singleton | Nearly all cards (separate banlist) | 1 legendary commander; multiplayer; **40 life**[^8] |
| **Limited (Draft / Sealed)** | 40+ | Only cards you open from packs | Tests deckbuilding under constraints[^8] |

---

# Part 2 — How to Play Optimally

Optimal Magic is the disciplined application of four interlocking concepts — **mana curve, card advantage, tempo, and threat assessment** — within a coherent deck archetype, supported by correct mulligan and sideboard decisions and good use of the stack/priority system.

## 2.1 Pick a Coherent Archetype

Every winning deck has a clear **game plan**. The four canonical archetypes:

| Archetype | Win Condition | Plays at | Cares Most About |
|-----------|---------------|----------|------------------|
| **Aggro** | Reduce life to 0 fast (turns 4–6) | Low curve, many 1–2 drops | Tempo, damage per mana |
| **Midrange** | Win attrition with efficient threats + answers | Curve out 2→3→4→5 | Card quality, flexibility |
| **Control** | Stall, then close with a few finishers | Counters, removal, sweepers | Card advantage |
| **Combo** | Assemble a card combination that wins instantly | Tutors, draw, protection | Consistency, redundancy |

The **archetype triangle** is loosely rock‑paper‑scissors: aggro beats control, control beats midrange/combo, midrange beats aggro, combo can beat anything if uninteracted with[^4][^6].

## 2.2 Deckbuilding Pillar #1 — Mana Curve

The **mana curve** is the distribution of mana costs in your deck. A smooth curve lets you spend all your mana every turn ("curving out"), which is one of the strongest tempo plays in Magic[^4].

**Typical 60‑card curves[^4]:**

- **Aggro:** ~20 lands, heavy at 1–2, almost nothing above 4.
- **Midrange:** ~23–24 lands, gentle hump at 2–4, a few 5+ finishers.
- **Control:** ~25–27 lands, light early interaction (1–2 mana), heavy 3–6 mana payoffs/sweepers.

**Limited (Draft/Sealed) rule of thumb** for a 40‑card deck[^4]:

| CMC | # of Cards |
|-----|-----------|
| 1 | 2–4 |
| 2 | 6–8 |
| 3 | 5–7 |
| 4 | 4–5 |
| 5+ | 3–4 |
| Lands | ~17 |

**Mana base color requirements:** Use a calculator or rule of thumb (e.g., "~14 sources of a color to reliably cast a CC spell on turn 2"). Frank Karsten's published manabase tables are the de facto standard reference for source counts[^4].

## 2.3 Deckbuilding Pillar #2 — Card Advantage

**Card advantage** = ending exchanges with more cards (and thus more options) than your opponent[^4]. Sources:

- **Pure draw spells** (e.g., *Divination*, *Sign in Blood*).
- **Two‑for‑ones**: one card that handles two of theirs (a sweeper like *Wrath of God*; a creature that comes with value like *Reflector Mage*).
- **Recurring engines**: planeswalkers that tick up, *Phyrexian Arena*, graveyard recursion.
- **Cantrips** (1‑mana "draw a card" spells like *Opt*, *Consider*) — small advantage but huge for **deck consistency**.

Control and midrange decks **must** generate more card advantage than they spend; aggro decks accept being card‑disadvantaged in exchange for tempo.

## 2.4 Deckbuilding Pillar #3 — Tempo

**Tempo** is mana efficiency over time — winning the race by making your opponent spend more mana than you do, or by deploying threats faster than they can answer[^4][^6].

Tempo plays:
- **Counter a 5‑mana bomb with a 2‑mana counterspell.** You traded 2 mana for 5 of theirs.
- **Bounce** a creature back to its owner's hand (e.g., *Unsummon*) — they re‑pay the mana cost and lose a turn.
- **Cheap, evasive threats** (1‑mana flyers, hasty creatures) that pressure life totals while opponents are still ramping or setting up.
- **Untapped lands** matter — playing a tapped land on turn 2 is a real tempo cost in fast formats.

A canonical tempo deck (e.g., **Izzet Delver** in eternal formats) uses cheap creatures + cheap interaction to force the opponent to play at a mana disadvantage every turn[^6].

## 2.5 Mulligans (London Mulligan, current)

Since 2019 Magic uses the **London Mulligan**[^5]:

1. Draw 7. If you keep, you keep all 7.
2. If you mulligan, **shuffle and draw 7 again** (not 6).
3. For each mulligan you took, **put that many cards from your hand on the bottom of your library** (you choose which) before the game starts.

**Optimal mulligan decisions[^5][^6]:**
- **Land count:** Keep hands with 2–5 lands for a typical 60‑card deck. 1‑land hands are usually keepable only with cheap spells + a cantrip; 6‑land hands are almost always mulligans.
- **Game plan viability:** Ask "Will this hand do anything in the first three turns?" If no, mulligan.
- **Combo decks** mulligan aggressively for key pieces — a 5‑card hand with the combo beats a 7‑card hand without it.
- **Synergy decks** need critical enablers (e.g., a graveyard deck without any enabler should mull).
- **Going second** lets you keep slightly more reactive hands due to the extra card.

## 2.6 The Stack, Priority, and Bluffing

Optimal play exploits priority and the stack[^4][^6]:

- **Hold priority** to chain abilities/spells before opponents can respond.
- **Respond at the right time:** kill a creature *in response to* its enters‑the‑battlefield ability triggering (it's still destroyed, but the ability already resolved or hasn't, depending on timing — know the difference between **cast triggers** and **ETB triggers**).
- **Leave mana up to "represent" an instant** (a counterspell, a removal spell, a combat trick). Even if your hand is empty, leaving 2U up vs. a Blue player is enough to make opponents play around *Counterspell*.
- **Bait counters** by casting a less important spell first to draw out interaction before the real threat.
- **End‑of‑turn timing:** activate sorcery‑speed‑equivalent draw effects (e.g., *Jace, the Mind Sculptor* +0) on the opponent's end step so you can use the new info and cards on your own turn.

## 2.7 Threat Assessment

The single most undervalued skill[^6]:

1. **Identify the biggest threat**, not just the most visible one. A 1/3 that draws a card every turn often beats a 5/5 vanilla creature in a long game.
2. **Don't trade your best removal for their worst threat.** Save *Path to Exile* for the *Sheoldred*, not the *Llanowar Elves*.
3. **Count damage and turns to lethal.** Always know how many turns you have until you die and how many until you can win — play to that math, not to "good plays."
4. **Read open mana.** If your opponent untaps with 2U up, assume *Counterspell*; if they have BB, assume removal. Plan around the worst case.
5. **Sequence to play around what's most likely**, not what's most feared. If they could have one of three answers, play around the most common one.

## 2.8 Sideboarding (Best‑of‑Three)

In tournament play (constructed), you have a **15‑card sideboard** you can swap into your main deck between games 2 and 3 of a match[^6][^8].

**Principles:**
- **Always swap one‑for‑one.** Decide what comes *out* before what goes *in*. Removing dead cards is often more valuable than adding new ones.
- **Don't dilute your plan.** Bringing in 8 cards just because you "could" usually weakens your deck's consistency.
- **Hate cards** (artifact destruction, graveyard hate, anti‑combo pieces) earn their slot only against decks where they're game‑changing.
- **Plan in advance.** Write a sideboard guide for each common matchup before the event so you don't burn time and mental energy at the table.
- **Be willing to "transform"** — some sideboards convert an aggro deck into a midrange deck post‑board to dodge the opponent's removal suite.

## 2.9 The Color Pie — Strategic Implications

Each color is mechanically defined by what it **can** and **cannot** do. Knowing this drives both deckbuilding (cover your weaknesses with a second color) and prediction (read your opponent's cards by color)[^7]:

| Color | Strengths | Weaknesses |
|-------|-----------|-----------|
| **White (W)** | "Go wide" tokens, efficient creature/enchantment removal, life gain, protection | Card draw, individually large creatures |
| **Blue (U)** | Card draw, counterspells, evasion, bounce, scry/filter | Permanent creature removal, aggressive bodies |
| **Black (B)** | Best creature kill, discard, recursion, drain | Enchantment removal, often pays life as cost |
| **Red (R)** | Direct damage ("burn"), haste, artifact/land destruction, speed | Card draw, late‑game sustain |
| **Green (G)** | Mana ramp, biggest creatures, artifact/enchantment removal, recursion | Stack interaction, evasive flyers, hard removal |

**Two‑color (guild) pairings** combine adjacent or opposed strengths — e.g., **Selesnya (GW)** "go wide and pump," **Izzet (UR)** "spells matter / tempo," **Golgari (BG)** "graveyard value." Three‑color decks (shards/wedges) trade consistency for power.

## 2.10 Archetype‑Specific Optimal Heuristics

**Aggro:**
- Mulligan into a curve — a one‑drop is usually worth keeping a 5‑card hand for.
- Attack every turn it's profitable. Damage is a clock; not attacking is a tempo loss.
- "Burn to the face" only when it kills (or is necessary to clear a blocker for lethal). Otherwise, point burn at creatures.

**Midrange:**
- Trade resources at parity, then close with one card that generates extra value.
- Almost always be the beatdown vs. control, the control vs. aggro. Identify your role each game[^6].

**Control:**
- Don't counter the first big spell — counter the **right** one. Your counterspell is a finite resource.
- Sweepers (e.g., *Wrath of God*, *Supreme Verdict*) are trade‑equity engines: always know when you can cast one without dying first.
- Win conditions should be **hard to interact with** (planeswalkers, manlands) so you can leave mana up for counters until the kill.

**Combo:**
- Goldfish your deck (play solitaire) until you know the average turn you "go off." That's your clock.
- Dedicate slots to **redundancy** (multiple cards that do the same job) over silver bullets.
- Run protection (counters, discard) proportional to how interactive your expected metagame is.

## 2.11 Practical Habits That Win Games

- **Track life totals and damage clocks every turn.** Not knowing you're dead next turn is the most common in‑game mistake.
- **Plan your turn before you untap.** When it's your opponent's end step, decide your full turn so you're not thinking on the clock.
- **Replay losses.** Every loss has a turn where a different decision changes the game — find it.
- **Play the matchup, not the deck.** The "right" play with the same hand differs against aggro vs. control.
- **Sleep, eat, and pace yourself in events.** Decision quality decays sharply when tired or rushed[^6].

---

## Confidence Assessment

- **High confidence:** All rules statements (turn structure, card types, stack/priority, combat keywords, mulligan procedure, format definitions) reflect the official Comprehensive Rules and well‑established Wizards of the Coast definitions. The Comprehensive Rules document itself is the canonical source[^2].
- **High confidence:** Strategic concepts (mana curve, card advantage, tempo, archetypes, color pie) are decades‑old, well‑established consensus among professional MTG players and content creators.
- **Medium confidence (advisory, not prescriptive):** Specific numeric guidelines (e.g., "20 lands for aggro," "14 sources for a CC spell") are widely cited rules of thumb but vary by deck and format. Frank Karsten's manabase research is the standard reference but evolves as new lands are printed.
- **Note on time sensitivity:** Banlists, the Standard rotation pool, and specific format metagames change frequently. Format rules summarized here reflect the situation as of 2024 (the most recent Comprehensive Rules referenced in research, dated 2024‑04‑12)[^2].

**Assumptions made (no clarification was sought, per task instructions):** Treated "how to play" as the rules of a 1v1 60‑card constructed game (the default), with extensions noted for Limited and Commander; treated "optimal play" as competitive‑constructed best practices applicable across formats rather than a deep dive into a specific format's metagame.

---

## Footnotes

[^1]: Wizards of the Coast, "How to Play Magic: The Gathering" beginner guide — objective (life total to 0), one‑land‑per‑turn rule, win conditions (life, deck‑out). See: <https://magic.wizards.com/en/how-to-play>.

[^2]: Wizards of the Coast, *Magic: The Gathering Comprehensive Rules*, §500–§514 ("Turn Structure"), revision dated 2024‑04‑12. Defines the five phases (Beginning, Precombat Main, Combat, Postcombat Main, Ending), their steps, and priority/timing rules. URL: <https://media.wizards.com/2024/downloads/MagicCompRules_20240412.txt>.

[^3]: Wizards of the Coast, "Card Types" reference; Comprehensive Rules §300–§310. Defines the eight card types (Land, Creature, Artifact, Enchantment, Planeswalker, Battle, Instant, Sorcery), their timing, and persistence. Battles introduced in *March of the Machine* (2023). URL: <https://magic.wizards.com/en/news/feature/card-types>.

[^4]: Channel Fireball / TCGplayer / Star City Games strategy primers (consensus references): mana curve construction, card advantage, tempo, the stack, priority. E.g., Reid Duke, "Level One" series at <https://magic.wizards.com/en/articles/columns/level-one>; Frank Karsten, "How Many Colored Mana Sources Do You Need to Consistently Cast Your Spells?" (TCGplayer).

[^5]: Wizards of the Coast announcement, "The London Mulligan Becomes Standard," 2019; current procedure documented in Comprehensive Rules §103.4. URL: <https://magic.wizards.com/en/articles/archive/news/london-mulligan-here-stay-2019-07-03>.

[^6]: Reid Duke, "Level One" curriculum, especially the lessons on Threat Assessment, Tempo, Sideboarding, and Roles ("Who's the Beatdown?" — Mike Flores's seminal article, *The Dojo*, 1999, reprinted at Star City Games: <https://articles.starcitygames.com/articles/whos-the-beatdown/>).

[^7]: Mark Rosewater, "Mechanical Color Pie" series (multi‑part), Wizards.com — definitive philosophy of WUBRG strengths and weaknesses. URL: <https://magic.wizards.com/en/news/making-magic/mechanical-color-pie-2021>.

[^8]: Wizards of the Coast, "Formats" overview — Standard, Pioneer, Modern, Legacy, Vintage, Commander, Limited (Draft/Sealed); Commander rules including 100‑card singleton, 40 life, commander damage. URL: <https://magic.wizards.com/en/formats>.

[^9]: Comprehensive Rules §702 ("Keyword Abilities") — definitions of Flying, First Strike, Double Strike, Trample, Deathtouch, Lifelink, Hexproof, Vigilance, Haste, Menace, Reach, etc. URL: <https://media.wizards.com/2024/downloads/MagicCompRules_20240412.txt> §702.
