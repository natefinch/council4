# Commander (EDH) Strategy Guide

A focused strategy document for **Magic: The Gathering's Commander format** (also called **EDH** — *Elder Dragon Highlander*). Covers general multiplayer strategy and detailed playbooks for the most common deck archetypes.

> Companion docs in this repo:
> - [`magic-the-gathering-a-trading-card-game-rules-and-.md`](./magic-the-gathering-a-trading-card-game-rules-and-.md) — rules of play & general MTG strategy
> - [`MTG-GLOSSARY.md`](./MTG-GLOSSARY.md) — slang & jargon reference

---

## 1. What Makes Commander Different

Commander is a **100‑card singleton multiplayer** format with a few defining rules that change strategy completely[^1]:

- **100 cards, no duplicates** (except basic lands). One copy of every card maximum.
- **A legendary creature is your "Commander"** — starts in the **Command Zone**, can be cast from there at its mana cost + 2 generic per previous cast (the **commander tax**).
- **40 starting life** (vs. 20 in 1v1) — games are longer; small life swings matter less, but lifelink and big swings matter more.
- **Commander damage:** taking **21+ combat damage from the same commander** loses you the game (separate from your life total).
- **Color identity:** every card in the deck must use only colors found in the commander's mana cost or rules text.
- **Multiplayer (usually 4 players, free‑for‑all).** This is the single biggest strategic difference from 1v1 Magic.
- **Banlist** maintained by the Commander format panel (now Wizards of the Coast); separate from other formats[^1][^2].

**Why this changes strategy:**
- You face **3 sources of threats and damage**, but you only have **1 turn in 4** to act on your own.
- **Politics, deal‑making, and threat perception** matter as much as raw card power[^3].
- **Card advantage compounds 3x faster** — drawing 1 vs. opponents drawing 1 is fine; drawing 1 vs. *each* opponent drawing 1 means you lose the long game by default.
- **Symmetric effects** (group hug, mass draw) are far more powerful — and far more dangerous — because they affect 3 opponents.

---

## 2. The Universal Commander Deckbuilding Template

Most successful Commander decks hit roughly these slot targets out of 99 cards (plus the commander)[^2]:

| Category | Count | Purpose |
|----------|-------|---------|
| **Lands** | 36–38 | Mana base; lean to 38+ with high curve, 35–36 with heavy ramp/low curve |
| **Ramp** | 8–12 | Mana acceleration (rocks, dorks, land ramp). Higher with expensive commanders |
| **Card draw / advantage** | 8–12 | Sustained card flow; **repeatable engines preferred** in Commander |
| **Single‑target removal** | 6–8 | Spot answers to creatures, artifacts, enchantments, planeswalkers |
| **Board wipes** | 3–5 | Mass removal — essential for catching up vs. wide boards |
| **Win conditions / finishers** | 2–4 | Explicit ways to close (combo, big creatures, alt‑win cards) |
| **Synergy / theme cards** | ~25–30 | The actual deck identity (tokens, +1/+1, sacrifice, etc.) |
| **Utility / flex slots** | 5–10 | Tutors, protection, graveyard hate, meta tech |

**Rules of thumb:**
- **Mono‑color** decks need *more* draw/ramp than they think (no fixing pressure, but worse access to it). Aim 10+ of each.
- **3+ colors** needs heavy fixing: dual lands, fetches, signets, talismans. Budget at least 30+ lands that produce ≥2 colors.
- **Green decks** can run 2–4 fewer mana rocks because of green's land ramp (e.g., *Cultivate*, *Three Visits*, *Nature's Lore*).
- **Repeatable** beats one‑shot: *Phyrexian Arena* > *Divination*, *Smothering Tithe* > *Gilded Goose*, *Eternal Witness* > *Regrowth*.

---

## 3. General Multiplayer Strategy

### 3.1 The Archenemy Effect

In a 4‑player game, **the player who looks scariest gets attacked by the other three**. This is the single most important dynamic in Commander[^3]:

- **Don't play your scariest threat first.** A turn‑3 *Smothering Tithe* paints a target on you for the rest of the game. A turn‑6 *Tithe* in a board state with bigger threats may go untouched.
- **Underplay your hand.** Resist deploying every card you draw. Holding interaction in hand is often more valuable than tapping out.
- **Never be the only one with a board.** If three players have empty boards, a single board wipe lands on you.
- **Watch removal counts.** Once two players have used their *Swords to Plowshares*, the table is short on answers — that's when you deploy your bombs.

### 3.2 Threat Assessment

You have *finite removal* and *finite turns*. Spending answers wrong loses games more often than drawing the wrong cards[^3][^4]:

1. **Identify game‑ending pieces first.** *Doubling Season*, *Aetherflux Reservoir*, *Grave Pact*, an active *Yawgmoth, Thran Physician* — these win games by themselves and must be answered.
2. **Distinguish "scary" from "actually winning."** A 9/9 vanilla creature looks dangerous. A 1/1 that draws 3 cards a turn is winning.
3. **Don't tunnel‑vision on the player attacking you.** The player swinging a 5/5 at your face is rarely the player about to win. Save your *Cyclonic Rift* for the combo player, not the aggro deck.
4. **Read open mana.** A blue player with 3 lands up at end of turn likely has a counter. Plan around the worst plausible card, not the worst card you can imagine.
5. **Count to lethal — both ways.** Always know how many turns until each opponent could kill you, and how many until you could kill each of them.

### 3.3 Politics & Deal‑Making

Politics is a real resource — use it deliberately[^3][^4]:

- **Make narrow, verifiable deals.** "I won't attack you next turn" is enforceable. "Let's be allies all game" is not.
- **Don't promise what costs you nothing.** "I won't board wipe" when you don't have a wipe is worthless.
- **Trade information.** Telling Player B that Player C has a tutored card in hand is free for you and changes the table's behavior.
- **Use removal as a favor.** "I'll *Path* their commander if you don't attack me" is one of the strongest moves in Commander.
- **Deflect with logic, not begging.** "If you kill me, the storm player wins next turn" is persuasive. "Please don't kill me" is not.
- **Honor most deals; break a deal *only* if it wins the game.** Reputation persists across games in your playgroup.

### 3.4 Avoiding Kingmaking

**Kingmaking** = handing a win to another player when you yourself can't win[^4].

- Always play to win. If you can't, play to extend the game, not to choose a winner.
- If you must affect outcomes, justify it with table‑level logic ("This player is one turn from winning, that one is two turns").
- Don't make decisions based on who beat you most recently — kingmaking out of spite is the worst look in EDH.

### 3.5 When to Attack

Attacking in Commander is more nuanced than in 1v1[^3][^4]:

- **Early chip damage** at one opponent often pays off — every point of damage they take is one less they can afford to take from someone else.
- **Pick the weakest defender, not the most threatening player**, unless you're trying to *be* the politician at the table. Beating up the player with no blockers makes you the bully; beating up the leader makes you the hero.
- **Don't attack just because you can.** A 2/2 swinging at a player with no board often makes you the next target without trading.
- **Attack into tapped opponents.** End‑of‑turn removal? Big tap‑out? That's your window.
- **Hold attackers as blockers** when the table is volatile — vigilance creatures are worth more than their stats suggest.

### 3.6 Mulligans in Commander

The standard **London Mulligan** applies (redraw 7, bottom one card per mulligan)[^4]. Most playgroups also offer a **free first mulligan** in casual Commander.

**Keep criteria:**
1. **2–4 lands.** 1‑land hands are usually mulligans even with a Sol Ring.
2. **Color access** for at least your first 2–3 turns and your commander.
3. **Some action by turn 3.** A hand of all 6+ mana spells with no ramp is a mulligan.
4. **A reason to win.** If your hand has no path to executing your gameplan, mull.

**Strong keep signals:** Sol Ring + 2–3 lands; turn‑1 ramp + 3 lands + a 4‑drop; for combo decks, a tutor + protection + lands.

**Mull aggressively** in combo decks (you need the combo) and in synergy decks missing their key enabler. **Mull conservatively** in midrange decks (most 7‑card hands have *something*).

---

## 4. Resource Management Specific to Commander

### 4.1 Card Draw Is Non‑Negotiable

In a 4‑player game, you naturally fall **3 cards behind per turn cycle** versus the table. To not lose to attrition:

- **Aim for 2–3 *repeatable* draw engines** in every deck (e.g., *Rhystic Study*, *Phyrexian Arena*, *Esper Sentinel*, *Beast Whisperer*, *Skullclamp*).
- **One‑shot draw spells** (*Harmonize*, *Concentrate*) are filler — fine for redundancy, not your engine.
- **Tutors** count as virtual card draw because they convert a card into the *right* card.

### 4.2 Ramp Curve

Optimal ramp deployment[^2]:

| Turn | Goal |
|------|------|
| 1 | Sol Ring / Arcane Signet / 1‑mana dork |
| 2 | Land + 2‑mana ramp (signet, Rampant Growth, Farseek) |
| 3 | Land + commander (if 4–5 CMC) or 3‑mana ramp |
| 4 | Real game plan begins |

Hitting a turn‑3 commander or turn‑3 high‑impact play wins more Commander games than any single card choice.

### 4.3 Board Wipe Theory

Board wipes are uniquely strong in Commander because they're **3‑for‑1‑for‑1‑for‑1**[^4]. Best practices:

- **Time wipes when you're most ahead post‑wipe.** A wipe that resets you and Player A while Player B and C still have boards is a bad wipe.
- **One‑sided or asymmetric wipes** are best (*Cyclonic Rift*, *Toxic Deluge* sized for them, *Farewell* for graveyard decks).
- **Wipe right after an opponent overcommits** to a board, not before.
- **Have a follow‑up.** A wipe with no plan to use the empty board is a tempo loss.

### 4.4 Commander Tax Discipline

- Recasting your commander 3 times = +6 mana over the original cost. Plan around it.
- **Voltron and Aristocrats** decks especially must accept the commander going to graveyard once or twice.
- **Use "send to hand" or "send to library"** removals to dodge tax (*Ephemerate*, *Eerie Interlude*, sacrifice‑then‑recast lines like *Phyrexian Reclamation*).

---

## 5. Common Deck Archetypes — Strategy Playbooks

Each archetype below includes: **core idea, sample commanders, key cards, win condition(s), how to pilot it, and weaknesses to play around.**

### 5.1 Voltron

> **One creature, infinitely buffed, hits opponents for 21 commander damage.**

- **Sample commanders:** Uril, the Miststalker; Sigarda, Host of Herons; Rafiq of the Many; Light‑Paws, Emperor's Voice; Halvar, God of Battle.
- **Key cards:** Swiftfoot Boots, Lightning Greaves, Shielded by Faith, Sigarda's Aid, Sword of Feast and Famine, Sword of Fire and Ice, Ethereal Armor, Colossus Hammer + Sigarda's Aid, Eldrazi Conscription.
- **Win condition:** 21 commander damage from a single hit (with double strike or trample).
- **How to play:**
  - Resolve commander **before** loading equipment — protect it with hexproof/indestructible/Greaves.
  - Spread damage among opponents to avoid making one player desperate; dump commander damage in 2–3 chunks rather than telegraphing one target.
  - Save protection spells (*Heroic Intervention*, *Teferi's Protection*) for board wipes — losing your suited‑up commander often loses the game.
- **Weaknesses:** Mass exile (*Farewell*), edict effects (*Liliana's Triumph*) that ignore hexproof, *Cyclonic Rift*, having only one threat. Mitigate with **redundancy** (multiple equip‑and‑swing creatures), **recursion** (*Hammer of Nazahn*, *Stoneforge Mystic*), and **protection density**.

### 5.2 Aristocrats

> **Sacrifice your own creatures repeatedly for value (drain, draw, recursion).**

- **Sample commanders:** Teysa Karlov; Korvold, Fae‑Cursed King; Meren of Clan Nel Toth; Liesa, Forgotten Archangel; Yahenni, Undying Partisan.
- **Key cards:** Blood Artist, Zulaport Cutthroat, Cruel Celebrant, Bastion of Remembrance (drain payoffs); Ashnod's Altar, Phyrexian Altar, Viscera Seer (sacrifice outlets); Grave Pact, Dictate of Erebos, Butcher of Malakir (death triggers); Reassembling Skeleton, Bloodghast (recursion); Pitiless Plunderer (mana from death).
- **Win condition:** Drain three opponents to 0 with stacked death triggers; or assemble an infinite (e.g., *Persist creature + Ashnod's Altar + Blood Artist*).
- **How to play:**
  - Build out the **engine pieces** (drain + sac outlet) before filling the board with fodder — fodder without payoffs is wasted board presence.
  - **Sacrifice in response to removal.** Always exhaust value triggers before letting a creature die "naturally."
  - **Grave Pact** effects make you nearly board‑wipe‑proof — ride them.
- **Weaknesses:** Graveyard hate (*Rest in Peace*, *Bojuka Bog*, *Soul‑Guide Lantern*), exile effects, "you can't gain life" effects, mass artifact/enchantment removal. Maintain **2–3 graveyard recursion lines** outside the GY (*Reanimate*‑type effects with shuffle protection).

### 5.3 Stax / Prison

> **Resource denial — tax or restrict opponents' actions while you grind out value.**

- **Sample commanders:** Grand Arbiter Augustin IV; Derevi, Empyrial Tactician; Winota, Joiner of Forces; Drannith Magistrate (in 99); Tergrid, God of Fright.
- **Key cards:** Smokestack, Tangle Wire, Winter Orb, Static Orb, Sphere of Resistance, Thalia, Guardian of Thraben, Drannith Magistrate, Thalia, Heretic Cathar, Trinisphere, Rule of Law, Stony Silence.
- **Win condition:** Lock the table out, then close with a slow but inevitable threat (a *Karn, the Great Creator* + *Mycosynth Lattice*, or a single big creature).
- **How to play:**
  - **Asymmetric pieces are everything.** A lock that hurts you equally is a tempo loss; lean on cards where your deck dodges the constraint (mana dorks under *Winter Orb*, low‑curve under *Sphere of Resistance*).
  - **Deploy locks as opponents tap out**, not on your own turn 3 when answers are abundant.
  - **Have a kill plan.** Stax that doesn't close gets ground out by topdecks across 3 opponents.
- **Weaknesses:** Mass artifact/enchantment removal (*Vandalblast*, *Bane of Progress*, *Farewell*), social pressure (people target stax players first), running out of pieces in long games. Bring **redundant lock pieces** and a **fast clock** so the lock doesn't have to last forever.

### 5.4 Tokens / Go‑Wide

> **Generate many small creatures, then pump and swing.**

- **Sample commanders:** Rhys the Redeemed; Adrix and Nev, Twincasters; Ghave, Guru of Spores; Krenko, Mob Boss; Talrand, Sky Summoner; Chatterfang, Squirrel General.
- **Key cards:** Doubling Season, Parallel Lives, Anointed Procession, Divine Visitation, Cathars' Crusade, Beastmaster Ascension, Craterhoof Behemoth, Overrun, Coat of Arms, Intangible Virtue, Glorious Anthem.
- **Win condition:** A single overrun‑style alpha strike (Craterhoof, Overwhelming Stampede, Akroma's Will) for the whole table.
- **How to play:**
  - **Don't go infinite‑wide on turn 4.** A board of 8 tokens turn 4 invites a board wipe before you have a finisher. Build to ~20 power on the turn you want to swing for lethal.
  - **Save your finisher.** Holding *Craterhoof* is the difference between winning and getting wiped.
  - **Anthem stacking is exponential** — *Doubling Season* + a token doubler = 4x tokens; layered effects scale fast.
- **Weaknesses:** Sweepers, *Pyroclasm*‑style 2‑damage wipes, exile sweepers (*Farewell*). Run **instant‑speed token producers** (*Secure the Wastes*, *White Sun's Twilight*) and **board‑wipe protection** (*Heroic Intervention*, *Eerie Interlude*, *Rootborn Defenses*).

### 5.5 Spellslinger / Storm

> **Cast a flood of instants and sorceries; payoffs scale with spell count.**

- **Sample commanders:** Niv‑Mizzet, Parun; Mizzix of the Izmagnus; Kykar, Wind's Fury; Krark, the Thumbless / Sakashima; Veyran, Voice of Duality; Krenko's Buddy in 99: Birgi, God of Storytelling.
- **Key cards:** Young Pyromancer, Talrand, Sky Summoner, Murmuring Mystic (token payoffs); Thousand‑Year Storm, Bonus Round (multiplier effects); Isochron Scepter, Mizzix's Mastery; Aetherflux Reservoir; Goblin Electromancer / Birgi (cost reduction); Past in Flames, Mizzix's Mastery (recursion).
- **Win conditions:** *Aetherflux Reservoir* into 50‑damage shots; storm count + *Grapeshot* / *Brain Freeze*; lethal *Thousand‑Year Storm* copies.
- **How to play:**
  - **Set up before going off.** A turn dedicated to *Mystic Remora* / *Rhystic Study* pays for itself in fuel.
  - **Cost reducers are the linchpin.** A resolved *Goblin Electromancer* often determines whether the combo turn works.
  - **Sequence cantrips first** to dig deeper before committing to the kill.
- **Weaknesses:** Counterspells, graveyard hate (kills recursion lines), creature‑hate sweepers (kill *Pyromancer*/*Talrand*). Run **redundant cost reducers** and **stack protection** (*Veil of Summer*, *Pact of Negation*, *Defense Grid*).

### 5.6 Reanimator

> **Dump huge creatures into the graveyard; revive them for cheap.**

- **Sample commanders:** Meren of Clan Nel Toth; Chainer, Dementia Master; The Scarab God; Muldrotha, the Gravetide; Sefris of the Hidden Ways; Sheoldred, the Apocalypse (in 99).
- **Key cards:** Entomb, Buried Alive, Jarad's Orders, Shred Memory (selection); Reanimate, Animate Dead, Necromancy, Exhume, Victimize, Dread Return (revival); Sheoldred, Whispering One, Razaketh, Jin‑Gitaxias, Atraxa, Grand Unifier, Archon of Cruelty (targets).
- **Win condition:** Cheat a 7+ mana creature into play turn 3–4; ride its value to a slow grind, or chain reanimations for combo finishes.
- **How to play:**
  - **The dream is "*Entomb* on EOT, *Reanimate* on your turn"** — set up the kill before the opposition can disrupt.
  - **Pick the right target for the moment.** *Jin‑Gitaxias* if there's no removal up; *Archon* if you need value; *Razaketh* to find your win.
  - **Have multiple lines from different zones.** *Reanimate* (graveyard), *Sneak Attack* (hand), *Birthing Pod* (battlefield) ensures graveyard hate doesn't end the game.
- **Weaknesses:** *Rest in Peace*, *Leyline of the Void*, exile effects, counterspells timed on the reanimate. Run **graveyard recursion that exiles** (*Eternal Witness* + *Reanimate*) and **hand‑based cheats** (*Sneak Attack*, *Through the Breach*) as redundancy.

### 5.7 Group Hug

> **Give everyone resources, then leverage the political goodwill (or pivot to a wincon).**

- **Sample commanders:** Phelddagrif; Kynaios and Tiro of Meletis; Kenrith, the Returned King; Selvala, Explorer Returned; Zedruu the Greathearted (variant).
- **Key cards:** Howling Mine, Rites of Flourishing, Veteran Explorer, Tempting Wurm, Mana Flare, Heartbeat of Spring, Dictate of Kruphix, Collective Voyage, Hunted Horror cycle, Skyshroud Claim.
- **Win conditions:** **You must have a finishing plan.** Common finishers: *Approach of the Second Sun*, *Maze's End*, *Mechanized Production*, *Laboratory Maniac* + symmetric draw, *Finale of Devastation* with a creature toolbox, or politics‑powered alliances that knock down threats one at a time.
- **How to play:**
  - **Goodwill is currency.** Use it to deflect early aggression, then convert to a win when you're ready.
  - **Time your pivot carefully.** Stop "hugging" the moment you reveal your finisher — the table will turn on you.
  - **Don't help the strongest opponent.** Skip giving cards or mana to whoever is closest to winning.
- **Weaknesses:** Combo decks abuse your engine to kill faster than you can pivot. Your wincon is fragile and often a single removal spell. Run **alt‑wincons** (multiple paths) and **light interaction** to save the day when symmetrical draw fuels someone else's combo.

### 5.8 Landfall / Ramp

> **Drop multiple lands per turn; trigger payoffs that scale with land drops.**

- **Sample commanders:** Omnath, Locus of Creation; Omnath, Locus of Rage; Lord Windgrace; Aesi, Tyrant of Gyre Strait; Tatyova, Benthic Druid.
- **Key cards:** Exploration, Burgeoning, Azusa, Lost but Seeking, Oracle of Mul Daya, Wayward Swordtooth, Dryad of the Ilysian Grove (extra land drops); Crucible of Worlds, Ramunap Excavator, Titania, Protector of Argoth (recursion); Avenger of Zendikar, Rampaging Baloths, Felidar Retreat, Scute Swarm (payoffs); Scapeshift, Splendid Reclamation, World Shaper (mass landfall).
- **Win condition:** Avenger‑of‑Zendikar‑into‑Craterhoof; mass landfall via *Scapeshift*; commander value (Omnath triggers).
- **How to play:**
  - **Extra land drops > mana rocks.** A turn‑1 *Exploration* outpaces *Sol Ring* in this archetype.
  - **Fetch lands are double‑triggers** (the fetch + the basic). Maximize them.
  - **Save *Scapeshift* / *Splendid Reclamation*** for after multiple lands have hit the graveyard — the combo turn is exponential.
- **Weaknesses:** Land destruction (*Armageddon*, *Strip Mine* loops), nonbasic hate (*Blood Moon*, *Back to Basics*). Lean on **basic land density**, run **land recursion**, and respect that you'll be archenemy whenever Omnath is out.

### 5.9 +1/+1 Counters

> **Build a single board of creatures with stacking counters; multiply via doublers and proliferate.**

- **Sample commanders:** Atraxa, Praetors' Voice; Ezuri, Claw of Progress; Marchesa, the Black Rose; Hamza, Guardian of Arashin; Pir, Imaginative Rascal & Toothy, Imaginary Friend.
- **Key cards:** Hardened Scales, Doubling Season, Branching Evolution, Conjurer's Closet (doublers); Inexorable Tide, Contagion Engine, Karn's Bastion, Evolution Sage (proliferate); Walking Ballista, Hangarback Walker, Forgotten Ancient (creatures); Cathars' Crusade (anthem); Simic Ascendancy (alt‑win).
- **Win conditions:** Combat damage from oversized creatures; *Walking Ballista* shoot‑outs; *Simic Ascendancy* alt‑win at 20 counters.
- **How to play:**
  - **Doubling effects are exponential.** *Hardened Scales* + *Branching Evolution* + a single counter creature wins quickly.
  - **Proliferate is a removal target.** Protect your enablers; they rarely come back.
  - **Mind the wipe risk** — like tokens, one *Wrath* often ends the game.
- **Weaknesses:** *Solemnity*, removal of doublers, board wipes that ignore counters. Run **board‑wipe protection** and **multiple paths** (alt‑win + combat).

### 5.10 Combo

> **Assemble a specific card interaction that wins (often instantly).**

- **Sample commanders:** Urza, Lord High Artificer; Kinnan, Bonder Prodigy; Thrasios + Tymna; Niv‑Mizzet, Parun (draw‑damage); Tasigur / Kess for spell loops; Kenrith for Heliod combos.
- **Common combos:**
  - **Dramatic Reversal + Isochron Scepter** with 3+ mana from rocks → infinite mana.
  - **Heliod, Sun‑Crowned + Walking Ballista** at 2+ counters → infinite damage.
  - **Thassa's Oracle + Demonic Consultation / Tainted Pact** → exile library, win on ETB.
  - **Kiki‑Jiki + Pestermite/Zealous Conscripts/Felidar Guardian** → infinite hasty tokens.
  - **Ad Nauseam + Angel's Grace** → draw whole deck.
  - **Food Chain + Eternal Scourge / Misthollow Griffin** → infinite creature mana.
  - **Hermit Druid → flip whole library + Dread Return + Necrotic Ooze** lines.
- **Key support:** Tutors (*Demonic*, *Vampiric*, *Mystical*, *Enlightened*, *Worldly*); fast mana (*Sol Ring*, *Mana Crypt*, *Mana Vault*); protection (*Force of Will*, *Fierce Guardianship*, *Pact of Negation*, *Veil of Summer*).
- **Win condition:** The combo itself.
- **How to play:**
  - **Combo lines on the stack must be protected.** Always count opponents' open mana before going for the kill; assume the worst counterspell.
  - **Goldfish your deck** to learn average combo turn — your clock relative to the table determines whether you're a turn‑5 cEDH deck or a turn‑8 midrange‑combo deck.
  - **Redundancy > a single perfect line.** Run multiple win conditions, multiple tutors, multiple protection pieces.
- **Weaknesses:** Counterspells, *Stony Silence* (vs. artifact combos), graveyard hate (vs. recursion combos), social pressure. Match your **protection density to the table's interaction density**.

### 5.11 Control / Draw‑Go

> **Counter, remove, recur — outlast the table and grind to a single win.**

- **Sample commanders:** Talrand, Sky Summoner; Baral, Chief of Compliance; Kess, Dissident Mage; Brago, King Eternal; The Scarab God; Mairsil, the Pretender.
- **Key cards:** Counterspell, Mana Drain, Force of Negation, Swan Song (counters); Cyclonic Rift, Toxic Deluge, Damnation, Farewell (sweepers); Rhystic Study, Mystic Remora, Phyrexian Arena (draw); Snapcaster Mage, Archaeomancer, Eternal Witness, Mission Briefing (recursion).
- **Win condition:** Slow finish via commander beats, planeswalker ult, or a single combo.
- **How to play:**
  - **You can't counter everything in a 4‑player game.** Pick your battles — counter the *winning* play, not the first 4‑drop.
  - **Cyclonic Rift overload is a finisher**, not a defense. Use it to clear the path to your kill turn.
  - **Stax pieces are control's best friend** — *Drannith Magistrate*, *Notion Thief*, *Narset, Parter of Veils* shut whole strategies down.
- **Weaknesses:** 3 opponents = 3x the spells to counter; pure control rarely closes alone. Bring **a real wincon** and **proactive plays** so you don't simply die to 100 cards' worth of opposing topdecks.

---

## 6. Format‑Specific Heuristics

### 6.1 Power Levels & Brackets

The Commander format moved (in 2024–25) from informal **1–10 power levels** to an official **5‑bracket system** maintained by Wizards of the Coast[^1][^2]:

| Bracket | Vibe | Allowed |
|---------|------|---------|
| **1 — Exhibition** | Theme decks, no infinite combos, no MLD | Pre‑con caliber |
| **2 — Core** | Casual, upgraded pre‑cons | No infinite combos before turn ~9, no fast mana, no MLD |
| **3 — Upgraded** | Tuned synergy decks | Some combos OK, no game‑changers turn‑4 |
| **4 — Optimized** | Best version of a strategy | Game‑changers (*Mana Crypt*, *Smothering Tithe*, *Cyclonic Rift*, *Rhystic Study*, etc.) allowed |
| **5 — cEDH** | Competitive | Anything legal; meta‑defined |

A pre‑game **Rule 0** conversation should align bracket, banlist house rules, and "what your deck does" before play[^3].

### 6.2 Game‑Changers (notable always‑powerful cards)

The Commander panel maintains a "Game‑Changers" list — cards strong enough to define power level. Includes (non‑exhaustive): *Sol Ring* (legal but watched), *Mana Crypt*, *Smothering Tithe*, *Rhystic Study*, *Mystic Remora*, *Cyclonic Rift*, *Vampiric Tutor*, *Demonic Tutor*, *Drannith Magistrate*, *Jeweled Lotus*, *Gaea's Cradle*[^1][^2]. Brackets 1–3 typically keep these out; Bracket 4+ welcomes them.

### 6.3 The "Salt" List (Etiquette)

Some cards are technically legal but socially fraught — overusing them poisons playgroups[^3]:

- **Mass land destruction** (*Armageddon*, *Jokulhaups*, *Ruination*) — usually warned against unless your bracket allows it.
- **Stax in casual pods** — fine in dedicated pods, often unwelcome in random.
- **Infect / poison** in casual.
- **Extra turns spam** (*Time Warp* loops).
- **Unbounded discard / mass theft** (*Tergrid*, *Coalition Relic* + *Mindslaver*).

Match these to the table's expectations during Rule 0.

---

## 7. Practical Habits That Win Commander Games

1. **Track all four life totals every turn**, not just yours.
2. **Plan your turn during the previous opponent's end step.** With 3 opponents between turns, you have time — use it.
3. **Acknowledge triggers immediately** (especially missed *Smothering Tithe* triggers — you can't go back).
4. **Lead with politics, not threats.** A new player attacked turn 1 will remember it for 5 games.
5. **Hold up mana even when you have nothing.** "I have something" is sometimes worth more than actually having something.
6. **Resolve your commander when it's safe, not when you have the mana.** A turn‑3 commander that gets countered or removed is a 3 × commander‑tax disaster.
7. **Win when you can win.** Slow‑playing a winning position because "the game is fun" is how you lose to topdecks.
8. **Be a good loser.** Reputation across games matters in Commander more than any single win.

---

## 8. Confidence Assessment

- **High confidence:** Rules of the format (100‑card singleton, 40 life, color identity, commander tax, 21 commander damage), the standard deckbuilding template, and archetype‑level core cards/strategies — all reflect long‑standing community consensus and Wizards' official Commander rules[^1].
- **High confidence:** Multiplayer dynamics (archenemy effect, threat assessment, politics) — extensively documented across competitive and casual Commander coverage for over a decade[^3][^4].
- **Medium confidence:** Specific deckbuilding numbers (36–38 lands, 8–12 ramp, 8–12 draw) are widely accepted heuristics from EDHREC and prominent content creators, but vary by deck. The specific bands here represent typical recommendations rather than rules.
- **Medium confidence:** Bracket system specifics — Wizards officially launched the system in late 2024/2025; categorizations and game‑changers list evolve over time, so the snapshot here is current as of research date.
- **Lower confidence:** Card‑specific recommendations (e.g., "*Heliod + Walking Ballista*") may be banned or shifted by future banlist updates; verify legality against the current Commander banlist before deck construction.
- **Assumption (no clarification was sought, per task instructions):** The user wants this saved as a *companion document in the repo* alongside the existing rules and glossary files. Created `COMMANDER-STRATEGY.md` in the repo root.

---

## Footnotes

[^1]: Wizards of the Coast — "Commander Format Rules" and "Magic: The Gathering Comprehensive Rules" §903 (Commander). Defines 100‑card singleton, color identity, command zone, commander tax (CR §903.8), 21 commander damage (CR §903.10), 40 starting life (CR §903.7). URLs: <https://magic.wizards.com/en/formats/commander> ; <https://media.wizards.com/2024/downloads/MagicCompRules_20240412.txt>.

[^2]: EDHREC and TCGplayer / Channel Fireball deckbuilding articles — consensus deckbuilding templates (lands, ramp, draw, removal, wipes, win conditions). EDHREC publishes a "Recommended Deckbuilding Template" used widely as a baseline. URL: <https://articles.edhrec.com/the-100-card-deck-template/>.

[^3]: Sheldon Menery, "Threat Assessment" series and "The Commander Player's Committee Philosophy" articles, *Star City Games* and `magic.wizards.com`; The Command Zone podcast — political theory and threat assessment in EDH. Representative URL: <https://magic.wizards.com/en/news/making-magic/threat-assessment-2014-12-29>.

[^4]: Reid Duke, "Multiplayer Magic" articles; Mike Flores, "Who's the Beatdown?" (1999) for role assignment as adapted to Commander; The Command Zone "Game Knowledge" episodes — multiplayer politics, kingmaking, archenemy dynamics. Representative URL: <https://articles.starcitygames.com/articles/whos-the-beatdown/>.
