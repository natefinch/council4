# Card-Support Planning Report

Capability-aware blockers for eligible paper cards that cannot yet be generated. Each distinct diagnostic summary and capability is counted at most once per card.

## Diagnostic reasons

A sole blocker is the card's only distinct diagnostic summary. The most common co-blocker excludes the reason in its own row.

| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | unsupported ordered effect sequence | 4,223 | 2,539 | 60.1% | unsupported optional effect |
| 2 | unsupported Oracle construct | 3,353 | 0 | 0.0% | unsupported static ability |
| 3 | unsupported static ability | 2,373 | 404 | 17.0% | unsupported Oracle construct |
| 4 | unsupported ability content | 1,303 | 147 | 11.3% | unsupported Oracle construct |
| 5 | unsupported triggered ability | 1,281 | 749 | 58.5% | unsupported Oracle construct |
| 6 | unsupported optional effect | 1,217 | 6 | 0.5% | unsupported ordered effect sequence |
| 7 | unsupported counter placement | 475 | 224 | 47.2% | unsupported Oracle construct |
| 8 | unsupported static declaration operation | 446 | 301 | 67.5% | unsupported static ability |
| 9 | unsupported activation cost | 442 | 151 | 34.2% | unsupported cost |
| 10 | unsupported damage spell | 426 | 324 | 76.1% | unsupported Oracle construct |
| 11 | unsupported enters-tapped replacement | 422 | 186 | 44.1% | unsupported Oracle construct |
| 12 | unsupported ability word | 406 | 145 | 35.7% | unsupported Oracle construct |
| 13 | unsupported activation condition | 382 | 250 | 65.4% | unsupported Oracle construct |
| 14 | unsupported static declaration group | 335 | 194 | 57.9% | unsupported Oracle construct |
| 15 | unsupported token creation | 328 | 186 | 56.7% | unsupported ordered effect sequence |
| 16 | unsupported return spell | 322 | 215 | 66.8% | unsupported Oracle construct |
| 17 | unsupported phase/step trigger phrase | 310 | 178 | 57.4% | unsupported ability content |
| 18 | unsupported static declaration condition | 307 | 202 | 65.8% | unsupported Oracle construct |
| 19 | unsupported search effect | 286 | 157 | 54.9% | unsupported optional effect |
| 20 | unsupported permanent zone-change trigger effect | 274 | 59 | 21.5% | unsupported Oracle construct |
| 21 | unsupported destroy spell | 271 | 196 | 72.3% | unsupported Oracle construct |
| 22 | unsupported power/toughness spell | 243 | 153 | 63.0% | unsupported static ability |
| 23 | unsupported exile spell | 239 | 121 | 50.6% | unsupported ordered effect sequence |
| 24 | unsupported cast effect | 213 | 92 | 43.2% | unsupported optional effect |
| 25 | unsupported triggered ability effect | 211 | 89 | 42.2% | unsupported Oracle construct |
| 26 | unsupported mixed keyword ability | 210 | 115 | 54.8% | unsupported static ability |
| 27 | unsupported activation ability word | 191 | 119 | 62.3% | unsupported Oracle construct |
| 28 | unsupported cost | 188 | 0 | 0.0% | unsupported activation cost |
| 29 | unsupported activation references | 158 | 113 | 71.5% | unsupported Oracle construct |
| 30 | unsupported temporary keyword spell | 157 | 119 | 75.8% | unsupported Oracle construct |
| 31 | unsupported permanent zone-change trigger | 155 | 92 | 59.4% | unsupported Oracle construct |
| 32 | unsupported phase/step trigger phrase effect | 151 | 70 | 46.4% | unsupported Oracle construct |
| 33 | unsupported gain-control spell | 149 | 97 | 65.1% | unsupported static ability |
| 34 | unsupported life spell | 144 | 101 | 70.1% | unsupported Oracle construct |
| 35 | unsupported sacrifice spell | 136 | 93 | 68.4% | unsupported optional effect |
| 36 | unsupported enters-with-counters replacement | 133 | 66 | 49.6% | unsupported ordered effect sequence |
| 37 | unsupported draw spell | 96 | 61 | 63.5% | unsupported Oracle construct |
| 38 | unsupported parameterized keyword | 96 | 45 | 46.9% | unsupported ordered effect sequence |
| 39 | unsupported counter spell | 84 | 65 | 77.4% | unsupported Oracle construct |
| 40 | unsupported loyalty ability | 80 | 8 | 10.0% | unsupported ordered effect sequence |
| 41 | unsupported optional replacement effect | 75 | 63 | 84.0% | unsupported Oracle construct |
| 42 | unsupported mana symbol | 75 | 50 | 66.7% | unsupported static ability |
| 43 | unsupported keyword or ability grant | 73 | 58 | 79.5% | unsupported Oracle construct |
| 44 | unsupported tap spell | 66 | 45 | 68.2% | unsupported Oracle construct |
| 45 | unsupported attach effect | 63 | 42 | 66.7% | unsupported static ability |
| 46 | unsupported type line | 61 | 60 | 98.4% | unsupported ordered effect sequence |
| 47 | unsupported shuffle effect | 60 | 46 | 76.7% | unsupported Oracle construct |
| 48 | unsupported mana effect | 58 | 32 | 55.2% | unsupported static ability |
| 49 | unsupported library placement | 55 | 37 | 67.3% | unsupported ordered effect sequence |
| 50 | unsupported Enchant ability | 52 | 28 | 53.8% | unsupported static ability |
| 51 | unsupported untap spell | 45 | 24 | 53.3% | unsupported static ability |
| 52 | unsupported activation timing | 42 | 34 | 81.0% | unsupported ordered effect sequence |
| 53 | unsupported static declaration duration | 40 | 32 | 80.0% | unsupported Oracle construct |
| 54 | unsupported static declaration shell | 40 | 30 | 75.0% | unsupported activation ability word |
| 55 | incomplete executable lowering | 39 | 33 | 84.6% | unsupported static ability |
| 56 | unsupported keyword or ability loss | 38 | 23 | 60.5% | unsupported Oracle construct |
| 57 | unsupported Equip ability | 37 | 12 | 32.4% | unsupported static ability |
| 58 | unsupported alternative effects | 31 | 21 | 67.7% | unsupported Oracle construct |
| 59 | unsupported emblem ability | 31 | 8 | 25.8% | unsupported ordered effect sequence |
| 60 | unsupported delayed effect | 29 | 23 | 79.3% | unsupported Oracle construct |
| 61 | unsupported can't-block effect | 27 | 21 | 77.8% | unsupported Oracle construct |
| 62 | unsupported discard spell | 25 | 17 | 68.0% | unsupported ability content |
| 63 | unsupported group power/toughness spell | 23 | 20 | 87.0% | unsupported Oracle construct |
| 64 | unsupported can't-be-blocked effect | 23 | 18 | 78.3% | unsupported Oracle construct |
| 65 | unsupported overload effect | 22 | 11 | 50.0% | unsupported ordered effect sequence |
| 66 | unsupported keyword ability | 22 | 6 | 27.3% | unsupported Oracle construct |
| 67 | unsupported card layout | 20 | 20 | 100.0% | - |
| 68 | unsupported manifest spell | 16 | 13 | 81.2% | unsupported Oracle construct |
| 69 | unsupported ability modes | 16 | 12 | 75.0% | unsupported Oracle construct |
| 70 | unsupported mill spell | 14 | 10 | 71.4% | unsupported activation ability word |
| 71 | unsupported copy effect | 14 | 8 | 57.1% | unsupported optional effect |
| 72 | unsupported fight spell | 14 | 5 | 35.7% | unsupported Oracle construct |
| 73 | unsupported multiple spell abilities | 13 | 13 | 100.0% | - |
| 74 | unsupported phase out spell | 13 | 5 | 38.5% | unsupported Oracle construct |
| 75 | unsupported unknown ability | 13 | 0 | 0.0% | unsupported Oracle construct |
| 76 | unsupported draw/discard trigger effect | 11 | 9 | 81.8% | unsupported Oracle construct |
| 77 | unsupported Flashback ability | 11 | 6 | 54.5% | unsupported ordered effect sequence |
| 78 | unsupported Protection ability | 10 | 7 | 70.0% | unsupported Oracle construct |
| 79 | unsupported alternative spell cost | 10 | 4 | 40.0% | unsupported ordered effect sequence |
| 80 | validation failed: invalid-ability-body | 9 | 9 | 100.0% | - |
| 81 | unsupported Crew ability | 9 | 6 | 66.7% | unsupported Oracle construct |
| 82 | unsupported connive effect | 9 | 5 | 55.6% | unsupported Oracle construct |
| 83 | unsupported double effect | 9 | 5 | 55.6% | unsupported ability content |
| 84 | unsupported prevent-damage effect | 8 | 6 | 75.0% | unsupported activation cost |
| 85 | unsupported Max speed ability | 7 | 7 | 100.0% | - |
| 86 | validation failed: oracle-without-abilities | 7 | 7 | 100.0% | - |
| 87 | unsupported divided damage spell | 7 | 6 | 85.7% | unsupported Oracle construct |
| 88 | unsupported Cumulative upkeep ability | 7 | 4 | 57.1% | unsupported life spell |
| 89 | unsupported investigate spell | 7 | 4 | 57.1% | unsupported cast effect |
| 90 | validation failed: invalid-selection | 6 | 6 | 100.0% | - |
| 91 | unsupported source-spell cost reduction | 6 | 5 | 83.3% | unsupported Oracle construct |
| 92 | unsupported scry spell | 6 | 4 | 66.7% | unsupported Oracle construct |
| 93 | unsupported lose-game effect | 6 | 3 | 50.0% | unsupported Oracle construct |
| 94 | unsupported gain player counter spell | 6 | 2 | 33.3% | unsupported optional effect |
| 95 | unsupported Read ahead ability | 6 | 0 | 0.0% | unsupported ordered effect sequence |
| 96 | unsupported regenerate spell | 5 | 5 | 100.0% | - |
| 97 | unsupported Persist ability | 5 | 4 | 80.0% | unsupported counter placement |
| 98 | unsupported delayed trigger | 5 | 3 | 60.0% | unsupported Oracle construct |
| 99 | unsupported draw-doubling replacement | 4 | 3 | 75.0% | unsupported activation references |
| 100 | unsupported damage replacement | 4 | 2 | 50.0% | unsupported enters-with-counters replacement |

## Capability clusters

A fully unlockable card has every distinct diagnostic summary in one capability cluster. Constituent summaries list the diagnostics currently observed in that cluster.

| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |
| --- | ---: | ---: | --- |
| shared-ability-content | 8,773 | 5,277 | unsupported ability content; unsupported ability modes; unsupported counter placement; unsupported counter spell; unsupported damage spell; unsupported delayed effect; unsupported destroy spell; unsupported discard spell; unsupported draw spell; unsupported draw/discard trigger effect; unsupported exile spell; unsupported explore spell; unsupported fight spell; unsupported gain-control spell; unsupported group power/toughness spell; unsupported investigate spell; unsupported keyword or ability grant; unsupported keyword or ability loss; unsupported library placement; unsupported life spell; unsupported mana effect; unsupported mana symbol; unsupported manifest spell; unsupported mill spell; unsupported multiple spell abilities; unsupported ordered effect sequence; unsupported phase/step trigger phrase effect; unsupported power/toughness spell; unsupported proliferate spell; unsupported regenerate spell; unsupported return spell; unsupported scry spell; unsupported search effect; unsupported tap spell; unsupported temporary keyword spell; unsupported triggered ability effect; unsupported untap spell |
| static-declaration | 3,734 | 1,401 | unsupported Enchant ability; unsupported Protection ability; unsupported Read ahead ability; unsupported keyword ability; unsupported mixed keyword ability; unsupported parameterized keyword; unsupported static ability; unsupported static declaration condition; unsupported static declaration duration; unsupported static declaration group; unsupported static declaration operation; unsupported static declaration shell |
| other | 2,901 | 1,061 | incomplete executable lowering; source generation failed; unsupported Bloodthirst ability; unsupported Channel ability; unsupported Class level ability; unsupported Crew ability; unsupported Cumulative upkeep ability; unsupported Embalm ability; unsupported Evoke ability; unsupported Flanking ability; unsupported Flashback ability; unsupported Hideaway play ability; unsupported Level up ability; unsupported Max speed ability; unsupported Persist ability; unsupported Spectacle ability; unsupported Undying ability; unsupported Unearth ability; unsupported adapt spell; unsupported alternative effects; unsupported alternative spell cost; unsupported amass spell; unsupported attach effect; unsupported become monarch effect; unsupported become-a-copy effect; unsupported can't-attack effect; unsupported can't-attack-or-block effect; unsupported can't-be-blocked effect; unsupported can't-block effect; unsupported can-attack-as-though-defender effect; unsupported card layout; unsupported cast effect; unsupported choose effect; unsupported color-change effect; unsupported connive effect; unsupported copy effect; unsupported delayed trigger; unsupported discard-then-draw spell; unsupported divided damage spell; unsupported double counters spell; unsupported double effect; unsupported draw-doubling replacement; unsupported draw-from-empty-library win replacement; unsupported emblem ability; unsupported emblem effect; unsupported enters-as-copy replacement; unsupported entry-choice replacement; unsupported forced-attack effect; unsupported gain player counter spell; unsupported graveyard-redirect replacement; unsupported landcycling ability; unsupported life-gain replacement; unsupported linked X spell cost; unsupported look-at-hand spell; unsupported lose-game effect; unsupported optional effect; unsupported optional replacement effect; unsupported overload effect; unsupported permanent zone-change trigger; unsupported permanent zone-change trigger effect; unsupported phase out effect; unsupported phase out spell; unsupported polymorph effect; unsupported prevent-damage effect; unsupported retarget effect; unsupported ring tempts effect; unsupported sacrifice spell; unsupported set base power/toughness effect; unsupported shuffle effect; unsupported source-spell cost reduction; unsupported surveil spell; unsupported tap or untap spell; unsupported target-animation effect; unsupported token creation; unsupported type line; unsupported win-game effect; validation failed: invalid-ability-body; validation failed: invalid-rule-effect; validation failed: invalid-selection; validation failed: oracle-without-abilities |
| trigger-pattern | 1,577 | 940 | unsupported draw/discard trigger; unsupported phase/step trigger phrase; unsupported triggered ability |
| activation | 1,331 | 800 | unsupported Cycling ability; unsupported Equip ability; unsupported Mutate ability; unsupported Ninjutsu ability; unsupported activation ability word; unsupported activation condition; unsupported activation cost; unsupported activation modes; unsupported activation references; unsupported activation timing; unsupported activation zone; unsupported cost; unsupported loyalty ability |
| replacement | 566 | 260 | unsupported conditional enters-tapped replacement; unsupported damage replacement; unsupported enters-tapped replacement; unsupported enters-with-counters replacement; unsupported self zone-destination replacement; unsupported token-creation replacement |
| recognition-fallback | 3,569 | 238 | unsupported Oracle construct; unsupported ability word; unsupported unknown ability |

## Unblock roadmap

Greedy set-cover priority: each step fixes the reason that — given the reasons already fixed in the steps above it — newly fully unblocks the most still-blocked cards. Cumulative is the running total of cards fully unblocked. Fan-out lowerers (ordered sequence, modal, optional) now report every independent blocker they carry, so these counts account for co-blockers rather than crediting a fix with cards that need other fixes too. A few remaining lowerers still short-circuit within an ability, so the counts stay a slight over-estimate.

| Step | Fix this reason | Capability | Newly unblocked | Cumulative | Sample cards |
| ---: | --- | --- | ---: | ---: | --- |
| 1 | unsupported ordered effect sequence | shared-ability-content | 2,539 | 2,539 | Abdel Adrian, Gorion's Ward, Aberrant Manawurm, Aberrant Researcher // Perfected Form, Abigale, Eloquent First-Year, Abnormal Endurance |
| 2 | unsupported optional effect | other | 808 | 3,347 | Absorb Identity, Abstract Performance, Abstruse Appropriation, Abstruse Archaic, Academy Loremaster |
| 3 | unsupported triggered ability | trigger-pattern | 789 | 4,136 | A Good Day to Pie, Aang and Katara, Aboleth Spawn, Abomination, Abomination, Irradiated Brute |
| 4 | unsupported static ability | static-declaration | 480 | 4,616 | Absorbing Man and Titania, Abyssal Persecutor, Aerial Modification, Ahn-Crop Champion, Ahn-Crop Crasher |
| 5 | unsupported Oracle construct | recognition-fallback | 1,227 | 5,843 | "Name Sticker" Goblin, Abominable Treefolk, Abomination of Llanowar, Abomination, World Ravager, Absolver Thrull |
| 6 | unsupported ability content | shared-ability-content | 1,012 | 6,855 | A Little Chat, About Face, Absorbing Man, Abuna's Chant, Academic Dispute |
| 7 | unsupported static declaration operation | static-declaration | 358 | 7,213 | Aboshan's Desire, Abzan Runemark, Acidic Sliver, Acolyte of Bahamut, Adelbert Steiner |
| 8 | unsupported damage spell | shared-ability-content | 360 | 7,573 | Acidic Soil, Acolyte's Reward, Advanced Reconstruction, Aether Flash, Ajani's Aid |
| 9 | unsupported enters-tapped replacement | replacement | 336 | 7,909 | Aberrant Return, Aether Revolt, Alhammarret, High Arbiter, Ali from Cairo, Alms Collector |
| 10 | unsupported ability word | recognition-fallback | 338 | 8,247 | Abaddon the Despoiler, Aboroth, Aeon Chronicler, Aerial Formation, Ajani's Presence |
| 11 | unsupported counter placement | shared-ability-content | 335 | 8,582 | Academy Researchers, Acrobatic Cheerleader, Adder-Staff Boggart, Aether Gust, Aether Vial |
| 12 | unsupported activation condition | activation | 337 | 8,919 | Aclazotz, Deepest Betrayal // Temple of the Dead, Aether Refinery, Agadeem Occultist, Agency Coroner, Akiri, Fearless Voyager |
| 13 | unsupported static declaration condition | static-declaration | 276 | 9,195 | Aang, A Lot to Learn, Ace's Baseball Bat, Angelic Voices, Animus of Predation, Arcades Sabboth |
| 14 | unsupported static declaration group | static-declaration | 275 | 9,470 | A Tale for the Ages, Adventurers' Guildhouse, Aetherflame Wall, Aminatou, Veil Piercer, Angel of Jubilation |
| 15 | unsupported phase/step trigger phrase | trigger-pattern | 275 | 9,745 | Afflicted Deserter // Werewolf Ransacker, Agent of Treachery, Air Nomad Student, Akuta, Born of Ash, Arachnus Web |
| 16 | unsupported return spell | shared-ability-content | 272 | 10,017 | Accursed Witch // Infectious Curse, Adarkar Valkyrie, Alchemist's Retrieval, Alesha, Who Laughs at Fate, Alexi, Zephyr Mage |
| 17 | unsupported token creation | other | 275 | 10,292 | A Killer Among Us, Aatchik, Emerald Radian, Abby, Merciless Soldier, Arachnogenesis, Ashiok, Nightmare Muse |
| 18 | unsupported search effect | shared-ability-content | 261 | 10,553 | Aang's Journey, Acquire, Aether Searcher, Agency Outfitter, Alpine Houndmaster |
| 19 | unsupported permanent zone-change trigger effect | other | 252 | 10,805 | "Lifetime" Pass Holder, Aang, Airbending Master, Aang, the Last Airbender, Aarakocra Sneak, Aberrant Mind Sorcerer |
| 20 | unsupported activation cost | activation | 251 | 11,056 | Abandon Hope, Aether Tide, Alms, Altar of Bhaal // Bone Offering, Anje, Maid of Dishonor |
| 21 | unsupported destroy spell | shared-ability-content | 247 | 11,303 | Aberrant, Abu Ja'far, Aether Storm, Age of Ultron, Ajani Vengeant |
| 22 | unsupported power/toughness spell | shared-ability-content | 219 | 11,522 | Acquired Mutation, Aegis of the Meek, Aethertide Whale, Alandra, Sky Dreamer, Alistair, the Brigadier |
| 23 | unsupported exile spell | shared-ability-content | 213 | 11,735 | Admonition Angel, Agrus Kos, Spirit of Justice, Aligned Hedron Network, All Hallow's Eve, Angel of Condemnation |
| 24 | unsupported triggered ability effect | shared-ability-content | 201 | 11,936 | Abigale, Poet Laureate // Heroic Stanza, Aetherstorm Roc, Aetherstream Leopard, Aisling Leprechaun, Alela, Cunning Conqueror |
| 25 | unsupported cast effect | other | 195 | 12,131 | Abeyance, Academic Probation, Aisha of Sparks and Smoke, Ajani's Response, Aleatory |
| 26 | unsupported mixed keyword ability | static-declaration | 181 | 12,312 | A Mysterious Creature, Animate Wall, Aragorn, Hornburg Hero, Arcades, the Strategist, Archetype of Aggression |
| 27 | unsupported activation ability word | activation | 168 | 12,480 | Abomination, Terrifying Titan, Adorned Crocodile, Adric, Mathematical Genius, Aerial Doombot, Afterburner Expert |
| 28 | unsupported cost | activation | 177 | 12,657 | Aang's Iceberg, Aang, Swift Savior // Aang and La, Ocean's Fury, Adagia, Windswept Bastion, Aetherflux Conduit, Aethersquall Ancient |
| 29 | unsupported activation references | activation | 154 | 12,811 | Aegis of Honor, Aladdin's Lamp, Alchemist's Refuge, Allosaurus Shepherd, Aphelia, Viper Whisperer |
| 30 | unsupported temporary keyword spell | shared-ability-content | 150 | 12,961 | Akroma's Blessing, Angelic Guardian, Aphotic Wisps, Apostle's Blessing, Arrester's Zeal |

## Ordered effect sequence sub-categories

Breakdown of the `unsupported ordered effect sequence` reason by the specific blocker within the sequence. A `sub-effect` row names the single-effect lowering a clause needs before its sequence can compile; a `structural` row names a sequence-machinery limitation. Counts mirror the diagnostic-reasons table: affected cards include co-blocked cards, sole blockers do not.

| Category | Affected cards | Sole blockers |
| --- | ---: | ---: |
| structural — per-effect condition unrecognized | 728 | 416 |
| sub-effect — unsupported counter placement | 721 | 379 |
| sub-effect — unsupported ability content | 652 | 362 |
| sub-effect — unsupported exile spell | 393 | 172 |
| sub-effect — unsupported token creation | 205 | 158 |
| sub-effect — unsupported cast effect | 357 | 154 |
| sub-effect — unsupported damage spell | 214 | 144 |
| sub-effect — unsupported return spell | 207 | 141 |
| sub-effect — unsupported power/toughness spell | 195 | 131 |
| sub-effect — unsupported life spell | 179 | 121 |
| sub-effect — unsupported temporary keyword spell | 158 | 116 |
| sub-effect — unsupported draw spell | 177 | 101 |
| structural — per-effect condition spans multiple clauses | 137 | 97 |
| sub-effect — unsupported shuffle effect | 147 | 85 |
| sub-effect — unsupported discard spell | 118 | 79 |
| sub-effect — unsupported sacrifice spell | 97 | 63 |
| sub-effect — unsupported untap spell | 85 | 60 |
| sub-effect — unsupported library placement | 138 | 57 |
| sub-effect — unsupported destroy spell | 74 | 57 |
| sub-effect — unsupported keyword or ability loss | 68 | 56 |
| structural — per-effect condition kind not gateable | 64 | 53 |
| sub-effect — unsupported keyword or ability grant | 90 | 52 |
| sub-effect — unsupported delayed effect | 81 | 50 |
| sub-effect — unsupported manifest spell | 104 | 46 |
| sub-effect — unsupported tap spell | 74 | 38 |
| sub-effect — unsupported search effect | 58 | 33 |
| sub-effect — unsupported mana symbol | 35 | 23 |
| structural — multi-effect body not lowered as a sequence | 33 | 23 |
| structural — per-effect condition has no containing clause | 27 | 20 |
| structural — coin flip branch not lowered | 22 | 20 |
| sub-effect — unsupported attach effect | 34 | 16 |
| structural — unsupported resolving optionality | 87 | 15 |
| sub-effect — unsupported gain-control spell | 19 | 13 |
| sub-effect — unsupported can't-be-blocked effect | 15 | 13 |
| structural — instead replacement not gatable | 14 | 13 |
| sub-effect — unsupported gain player counter spell | 15 | 12 |
| sub-effect — unsupported group power/toughness spell | 14 | 12 |
| sub-effect — unsupported mill spell | 20 | 11 |
| sub-effect — unsupported can't-block effect | 14 | 11 |
| sub-effect — unsupported fight spell | 14 | 10 |
| structural — single effect requires ordered lowering | 11 | 10 |
| structural — per-effect condition predicate not gateable | 44 | 9 |
| structural — inherited target not remappable | 18 | 9 |
| structural — unconsumed targets/references/keywords | 11 | 9 |
| sub-effect — unsupported mana effect | 13 | 8 |
| sub-effect — unsupported investigate spell | 11 | 8 |
| sub-effect — unsupported double effect | 9 | 7 |
| structural — unsupported linked counter and token creation | 7 | 7 |
| structural — unsupported sacrifice-conditioned reanimation | 7 | 7 |
| mode 1: sub-effect — unsupported ability content | 9 | 6 |
| structural — delayed-target sacrifice not linkable | 7 | 6 |
| sub-effect — unsupported connive effect | 7 | 5 |
| mode 2: sub-effect — unsupported ability content | 6 | 5 |
| sub-effect — unsupported counter spell | 6 | 5 |
| sub-effect — unsupported regenerate spell | 6 | 5 |
| sub-effect — unsupported type-change effect | 6 | 5 |
| sub-effect — unsupported amass spell | 5 | 4 |
| sub-effect — unsupported copy effect | 77 | 3 |
| sub-effect — unsupported scry spell | 7 | 3 |
| sub-effect — unsupported divided damage spell | 3 | 3 |
| sub-effect — unsupported set base power/toughness effect | 3 | 3 |
| structural — non-exact legacy effect pair | 5 | 2 |
| sub-effect — unsupported emblem ability | 3 | 2 |
| sub-effect — unsupported explore spell | 3 | 2 |
| sub-effect — unsupported forced-attack effect | 3 | 2 |
| sub-effect — unsupported surveil spell | 3 | 2 |
| mode 1: sub-effect — unsupported discard spell | 2 | 2 |
| mode 2: sub-effect — unsupported power/toughness spell | 2 | 2 |
| mode 3: sub-effect — unsupported damage spell | 2 | 2 |
| structural — payoff amount references a non-target permanent | 2 | 2 |
| structural — vote arm not lowered | 2 | 2 |
| sub-effect — unsupported can't-attack effect | 2 | 2 |
| sub-effect — unsupported color-change effect | 2 | 2 |
| sub-effect — unsupported lose-game effect | 2 | 2 |
| sub-effect — unsupported phase out effect | 2 | 2 |
| sub-effect — unsupported proliferate spell | 2 | 2 |
| sub-effect — unsupported win-game effect | 2 | 2 |
| sub-effect — unsupported look-at-hand spell | 4 | 1 |
| mode 2: sub-effect — unsupported exile spell | 3 | 1 |
| sub-effect — unsupported double counters spell | 3 | 1 |
| mode 1: sub-effect — unsupported counter placement | 2 | 1 |
| mode 1: sub-effect — unsupported return spell | 2 | 1 |
| mode 1: sub-effect — unsupported sacrifice spell | 2 | 1 |
| mode 2: structural — inherited target not remappable | 2 | 1 |
| mode 2: sub-effect — unsupported library placement | 2 | 1 |
| mode 2: sub-effect — unsupported return spell | 2 | 1 |
| mode 2: sub-effect — unsupported temporary keyword spell | 2 | 1 |
| mode 1: structural — counter-spell target | 1 | 1 |
| mode 1: structural — inherited target not remappable | 1 | 1 |
| mode 1: structural — per-effect condition unrecognized: If it's your combat phase | 1 | 1 |
| mode 1: sub-effect — unsupported draw spell | 1 | 1 |
| mode 1: sub-effect — unsupported life spell | 1 | 1 |
| mode 1: sub-effect — unsupported mana symbol | 1 | 1 |
| mode 1: sub-effect — unsupported power/toughness spell | 1 | 1 |
| mode 1: sub-effect — unsupported temporary keyword spell | 1 | 1 |
| mode 2: structural — per-effect condition spans multiple clauses | 1 | 1 |
| mode 2: structural — per-effect condition unrecognized: If an opponent protects it | 1 | 1 |
| mode 2: structural — per-effect condition unrecognized: if it's an artifact creature card | 1 | 1 |
| mode 2: sub-effect — unsupported counter placement | 1 | 1 |
| mode 2: sub-effect — unsupported draw spell | 1 | 1 |
| mode 2: sub-effect — unsupported keyword or ability loss | 1 | 1 |
| mode 2: sub-effect — unsupported tap spell | 1 | 1 |
| mode 2: sub-effect — unsupported token creation | 1 | 1 |
| mode 3: sub-effect — unsupported exile spell | 1 | 1 |
| mode 3: sub-effect — unsupported manifest spell | 1 | 1 |
| mode 3: sub-effect — unsupported temporary keyword spell | 1 | 1 |
| mode 4: sub-effect — unsupported counter placement | 1 | 1 |
| mode 4: sub-effect — unsupported discard spell | 1 | 1 |
| mode 4: sub-effect — unsupported power/toughness spell | 1 | 1 |
| structural — counter-spell target | 1 | 1 |
| structural — instead negated gate not applicable | 1 | 1 |
| sub-effect — unsupported become monarch effect | 1 | 1 |
| sub-effect — unsupported can-attack-as-though-defender effect | 1 | 1 |
| sub-effect — unsupported double power/toughness spell | 1 | 1 |
| sub-effect — unsupported prevent-damage effect | 1 | 1 |
| sub-effect — unsupported retarget effect | 46 | 0 |
| sub-effect — unsupported remove from combat spell | 12 | 0 |
| mode 1: sub-effect — unsupported can't-be-blocked effect | 1 | 0 |
| mode 1: sub-effect — unsupported manifest spell | 1 | 0 |
| mode 1: sub-effect — unsupported token creation | 1 | 0 |
| mode 2: structural — single effect requires ordered lowering | 1 | 0 |
| mode 2: structural — unconsumed targets/references/keywords | 1 | 0 |
| mode 2: sub-effect — unsupported copy effect | 1 | 0 |
| mode 2: sub-effect — unsupported life spell | 1 | 0 |
| mode 2: sub-effect — unsupported scry spell | 1 | 0 |
| mode 3: sub-effect — unsupported ability content | 1 | 0 |
| structural — clause produced modal/shared/multi-mode content | 1 | 0 |
| structural — otherwise branch not gatable | 1 | 0 |
| sub-effect — unsupported adapt spell | 1 | 0 |
| sub-effect — unsupported delayed trigger | 1 | 0 |
| sub-effect — unsupported tap or untap spell | 1 | 0 |

## Unrecognized per-effect conditions (recognition backlog)

Distinct `if <condition>` wordings inside ordered sequences whose predicate the compiler does not yet recognize. Recognizing a wording unblocks ordered-sequence lowering for the listed cards. Rows are ranked by sole blockers (cards a single wording is the only blocker for) then affected cards.

| Unrecognized condition | Affected cards | Sole blockers |
| --- | ---: | ---: |
| If you win | 16 | 14 |
| If you can't | 14 | 9 |
| If it's a creature card | 13 | 9 |
| If five or more mana was spent to cast that spell | 9 | 8 |
| If this is the second time this ability has resolved this turn | 9 | 8 |
| If it's your turn | 8 | 8 |
| If it doesn't have suspend | 8 | 5 |
| if it has three or more +1/+1 counters on it | 5 | 5 |
| If it's a land card | 10 | 4 |
| If you lose the flip | 5 | 4 |
| If it's an artifact creature | 4 | 4 |
| If that spell is countered this way | 7 | 3 |
| If you cast a spell this way | 7 | 3 |
| If a card with the chosen name was milled this way | 3 | 3 |
| If a player is dealt damage this way | 3 | 3 |
| If it's a creature | 3 | 3 |
| If that creature has toxic | 3 | 3 |
| If the player can't | 3 | 3 |
| If you controlled that permanent | 3 | 3 |
| If you win the flip | 3 | 3 |
| If you search your library this way | 15 | 2 |
| If damage is prevented this way | 4 | 2 |
| If a land card was milled this way | 3 | 2 |
| If you have a full party | 3 | 2 |
| If{B}was spent to cast this spell | 3 | 2 |
| if able | 3 | 2 |
| If excess damage was dealt this way | 2 | 2 |
| If it entered under your control | 2 | 2 |
| If it was a creature card | 2 | 2 |
| If its mana value was 2 or less | 2 | 2 |
| If its mana value was 3 or less | 2 | 2 |
| If that creature dies this way | 2 | 2 |
| If that land was a snow land | 2 | 2 |
| If that land was nonbasic | 2 | 2 |
| If that spell has mana value 5 or greater | 2 | 2 |
| If that spell's mana value is 5 or greater | 2 | 2 |
| If they can't | 2 | 2 |
| If this spell was cast from anywhere other than your hand | 2 | 2 |
| If you controlled it | 2 | 2 |
| If you discard a creature card this way | 2 | 2 |
| If{U}was spent to cast this spell | 2 | 2 |
| if an opponent has more life than you | 2 | 2 |
| if its mana value is 2 or less | 2 | 2 |
| if that creature has a +1/+1 counter on it | 2 | 2 |
| if you control an artifact and an enchantment | 2 | 2 |
| if you have more life than an opponent | 2 | 2 |
| If a creature card was exiled this way | 4 | 1 |
| If they don't | 4 | 1 |
| If it's a permanent card | 3 | 1 |
| If that creature is attacking | 3 | 1 |
| If that creature is legendary | 3 | 1 |
| If a creature card is exiled this way | 2 | 1 |
| If a graveyard has twenty or more cards in it | 2 | 1 |
| If a land card is revealed this way | 2 | 1 |
| If a permanent's ability is countered this way | 2 | 1 |
| If it's a Zombie card | 2 | 1 |
| If it's an artifact | 2 | 1 |
| if it was attacking | 2 | 1 |
| If Falling Star doesn't turn completely over at least once during the flip | 1 | 1 |
| If Mishra | 1 | 1 |
| If X is greater than or equal to the number of cards in your library | 1 | 1 |
| If a Dinosaur is dealt damage this way | 1 | 1 |
| If a Food entered under your control this turn | 1 | 1 |
| If a Goblin is sacrificed this way | 1 | 1 |
| If a Hero enters this way | 1 | 1 |
| If a Pirate was exiled this way | 1 | 1 |
| If a Rat is attacking | 1 | 1 |
| If a card is put into exile this way | 1 | 1 |
| If a card named Stomping Slabs was revealed this way | 1 | 1 |
| If a card with the chosen name is revealed this way | 1 | 1 |
| If a creature card is put into a graveyard this way | 1 | 1 |
| If a creature is dealt damage this way | 1 | 1 |
| If a creature is put into exile this way | 1 | 1 |
| If a creature you controlled was destroyed this way | 1 | 1 |
| If a land card is discarded this way | 1 | 1 |
| If a land card is milled this way | 1 | 1 |
| If a land enters this way | 1 | 1 |
| If a nonland permanent left the battlefield this turn or a spell was warped this turn | 1 | 1 |
| If a nonred permanent is dealt damage this way | 1 | 1 |
| If a permanent was returned this way | 1 | 1 |
| If a permanent you controlled or a token was destroyed this way | 1 | 1 |
| If a white creature dies this way | 1 | 1 |
| If all cards revealed this way are creature cards | 1 | 1 |
| If an ability of an artifact | 1 | 1 |
| If an artifact or creature spell is countered this way | 1 | 1 |
| If an instant card or a card with flash is exiled this way | 1 | 1 |
| If an instant or sorcery card was milled this way | 1 | 1 |
| If an opponent controls that creature | 1 | 1 |
| If an opponent has more cards in hand than you | 1 | 1 |
| If another Desert was returned this way | 1 | 1 |
| If at least one Angel card is milled this way | 1 | 1 |
| If at least one creature card was exiled this way | 1 | 1 |
| If both coins come up heads | 1 | 1 |
| If damage from a black source is prevented this way | 1 | 1 |
| If damage from a creature source is prevented this way | 1 | 1 |
| If damage from a red source is prevented this way | 1 | 1 |
| If each opponent chose silence | 1 | 1 |
| If excess damage was dealt to a permanent this way | 1 | 1 |
| If excess damage was dealt to that creature this way | 1 | 1 |
| If four or more mana was spent to cast that spell | 1 | 1 |
| If it had mana value 3 or less | 1 | 1 |
| If it has a +1/+1 counter on it | 1 | 1 |
| If it has a counter on it | 1 | 1 |
| If it is | 1 | 1 |
| If it was a Gideon planeswalker | 1 | 1 |
| If it was a Jace planeswalker spell | 1 | 1 |
| If it was a Soldier | 1 | 1 |
| If it was a token | 1 | 1 |
| If it was an artifact | 1 | 1 |
| If it was attacking | 1 | 1 |
| If it was dealt damage this turn | 1 | 1 |
| If it was tapped | 1 | 1 |
| If it's a Forest card | 1 | 1 |
| If it's a creature or land card | 1 | 1 |
| If it's a land | 1 | 1 |
| If it's an artifact card | 1 | 1 |
| If it's an enchanted creature or enchantment creature | 1 | 1 |
| If it's an enchantment creature or legendary creature | 1 | 1 |
| If it's attacking a battle | 1 | 1 |
| If it's night | 1 | 1 |
| If it's not your turn | 1 | 1 |
| If it's paired with a creature | 1 | 1 |
| If it's renowned | 1 | 1 |
| If it's the first combat phase of your turn | 1 | 1 |
| If its controller has more than four cards in hand | 1 | 1 |
| If its mana value is 2 or less | 1 | 1 |
| If its mana value was 1 or less | 1 | 1 |
| If its mana value was 4 or less | 1 | 1 |
| If mana from a Treasure was spent to cast this spell | 1 | 1 |
| If no counters were removed this way | 1 | 1 |
| If no creature cards were revealed this way | 1 | 1 |
| If no creature got votes | 1 | 1 |
| If target creature has toughness 5 or greater | 1 | 1 |
| If that Equipment was attached to a creature | 1 | 1 |
| If that artifact had counters on it | 1 | 1 |
| If that card is a Goblin card | 1 | 1 |
| If that card is a land card | 1 | 1 |
| If that card's mana value is less than or equal to the number of experience counters you have | 1 | 1 |
| If that creature had a +1/+1 counter on it | 1 | 1 |
| If that creature had power 2 or less | 1 | 1 |
| If that creature has a +1/+1 counter on it | 1 | 1 |
| If that creature has flying | 1 | 1 |
| If that creature has infect | 1 | 1 |
| If that creature has two or more +1/+1 counters on it | 1 | 1 |
| If that creature is a Bird | 1 | 1 |
| If that creature is a Demon | 1 | 1 |
| If that creature is a Human | 1 | 1 |
| If that creature is a Zombie | 1 | 1 |
| If that creature is a token | 1 | 1 |
| If that creature is an Ally | 1 | 1 |
| If that creature is an Assassin | 1 | 1 |
| If that creature is an enchantment | 1 | 1 |
| If that creature is another Hero | 1 | 1 |
| If that creature is black or red | 1 | 1 |
| If that creature is white | 1 | 1 |
| If that creature is white or blue | 1 | 1 |
| If that creature was a Goblin or Orc | 1 | 1 |
| If that creature was a God | 1 | 1 |
| If that creature was a Human | 1 | 1 |
| If that creature was a Villain | 1 | 1 |
| If that creature was an Elemental | 1 | 1 |
| If that creature was blue or black | 1 | 1 |
| If that creature was cast for its warp cost | 1 | 1 |
| If that creature was green or white | 1 | 1 |
| If that creature wasn't dealt damage this turn | 1 | 1 |
| If that land is a Forest | 1 | 1 |
| If that land is a Mountain | 1 | 1 |
| If that land is a Swamp | 1 | 1 |
| If that land is an Island | 1 | 1 |
| If that land was legendary | 1 | 1 |
| If that library contains exactly the chosen number of cards with the chosen name | 1 | 1 |
| If that permanent had mana value 3 or less | 1 | 1 |
| If that permanent is a Spirit | 1 | 1 |
| If that permanent is black | 1 | 1 |
| If that permanent is green | 1 | 1 |
| If that permanent is red or green | 1 | 1 |
| If that permanent was a Liliana planeswalker | 1 | 1 |
| If that permanent was a Nissa planeswalker | 1 | 1 |
| If that permanent was blue or black | 1 | 1 |
| If that permanent's mana value was 3 or less | 1 | 1 |
| If that player can't | 1 | 1 |
| If that player is you | 1 | 1 |
| If that spell targets a commander you control | 1 | 1 |
| If that spell was all colors | 1 | 1 |
| If that spell's mana value was 3 or less | 1 | 1 |
| If that spell's power is 4 or greater | 1 | 1 |
| If the amount of mana spent to cast that spell was less than its mana value | 1 | 1 |
| If the card a player revealed has the name they chose | 1 | 1 |
| If the card's mana value is 1 or less | 1 | 1 |
| If the creature you control entered this turn | 1 | 1 |
| If the creature you control has trample | 1 | 1 |
| If the discarded card wasn't a land card | 1 | 1 |
| If the discovered card's mana value is less than 10 | 1 | 1 |
| If the exiled card is a God card | 1 | 1 |
| If the number is odd | 1 | 1 |
| If the player is your opponent and has four or more cards in hand | 1 | 1 |
| If the player mills at least one creature card this way | 1 | 1 |
| If the result is 3 or less | 1 | 1 |
| If the result is equal to or less than the number of Robots you control | 1 | 1 |
| If the result is greater than the damage dealt or the result is 12 | 1 | 1 |
| If the sacrificed artifact was legendary | 1 | 1 |
| If the sacrificed creature was legendary | 1 | 1 |
| If the sacrificed permanent was a Vehicle | 1 | 1 |
| If the sacrificed permanent was an artifact | 1 | 1 |
| If the spell is countered this way | 1 | 1 |
| If there are three or more depletion counters on this enchantment | 1 | 1 |
| If there is an instant card and a sorcery card in your graveyard | 1 | 1 |
| If they guessed right | 1 | 1 |
| If they lose the flip | 1 | 1 |
| If they lost life this turn | 1 | 1 |
| If this is the first time this ability has resolved this turn | 1 | 1 |
| If this spell was cast from your hand | 1 | 1 |
| If this spell was cast from your hand and you've cast another spell named Approach of the Second Sun this game | 1 | 1 |
| If those Auras would leave the battlefield | 1 | 1 |
| If time gets more votes | 1 | 1 |
| If two cards that share a card type are discarded this way | 1 | 1 |
| If two cards that share all their card types were milled this way | 1 | 1 |
| If two or more players are tied for fewest | 1 | 1 |
| If you are one of those players | 1 | 1 |
| If you can't sacrifice a creature | 1 | 1 |
| If you control a creature with a counter on it | 1 | 1 |
| If you control a modified creature | 1 | 1 |
| If you control an artifact and an enchantment | 1 | 1 |
| If you control more creatures of that type than each other player | 1 | 1 |
| If you control more creatures than each other player | 1 | 1 |
| If you control that creature | 1 | 1 |
| If you controlled a modified creature as you cast this spell | 1 | 1 |
| If you controlled that artifact | 1 | 1 |
| If you discarded a card this way | 1 | 1 |
| If you don't control a Faerie | 1 | 1 |
| If you don't control a Human | 1 | 1 |
| If you exiled a card this way | 1 | 1 |
| If you gain control of a creature this way | 1 | 1 |
| If you have less life than an opponent | 1 | 1 |
| If you lose a flip | 1 | 1 |
| If you paid one or more{E}this way | 1 | 1 |
| If you pay | 1 | 1 |
| If you put +1/+1 counters on five Dragons this way | 1 | 1 |
| If you put an artifact card into your hand this way | 1 | 1 |
| If you put fewer than two lands onto the battlefield this way | 1 | 1 |
| If you return a nonland card to your hand this way | 1 | 1 |
| If you sacrifice an Island this way | 1 | 1 |
| If you win all the flips | 1 | 1 |
| If you've cast another spell this turn | 1 | 1 |
| If you've cycled a card named Yidaro | 1 | 1 |
| If you've discarded a card this turn | 1 | 1 |
| If you've drawn two or more cards this turn | 1 | 1 |
| If{G}was spent to cast this spell | 1 | 1 |
| If{R}{G}was spent to cast this spell | 1 | 1 |
| If{S}was spent to cast this spell | 1 | 1 |
| If{W}was spent to cast this spell | 1 | 1 |
| if Jor Kadeen's power is 4 or greater | 1 | 1 |
| if a creature named Eight-and-a-Half-Tails is on the battlefield | 1 | 1 |
| if able without paying their mana costs | 1 | 1 |
| if an opponent has more cards in hand than you | 1 | 1 |
| if any player pays{2} | 1 | 1 |
| if it attacked or blocked since your last upkeep | 1 | 1 |
| if it had a death counter on it | 1 | 1 |
| if it has eight or more night counters on it | 1 | 1 |
| if it has five or more bloodstain counters on it | 1 | 1 |
| if it has three or more ritual counters on it | 1 | 1 |
| if it wasn't kicked | 1 | 1 |
| if it's attacking one of your opponents | 1 | 1 |
| if it's night | 1 | 1 |
| if its power is 16 or less | 1 | 1 |
| if its power is 2 | 1 | 1 |
| if its power is 4 or greater | 1 | 1 |
| if its power is exactly 20 | 1 | 1 |
| if its power is less than Shelinda's power | 1 | 1 |
| if that creature has three or more +1/+1 counters on it | 1 | 1 |
| if that creature is a Mutant | 1 | 1 |
| if that creature was destroyed this way | 1 | 1 |
| if that creature's power is 0 or less | 1 | 1 |
| if that creature's power is 2 or less | 1 | 1 |
| if that creature's power is greater than Yorvo's power | 1 | 1 |
| if that creature's toughness is 1 or greater | 1 | 1 |
| if that player has more cards in hand than you | 1 | 1 |
| if that target is white and/or blue | 1 | 1 |
| if the player hasn't played the card | 1 | 1 |
| if there are five basic land types among lands you control | 1 | 1 |
| if there are five or more hatchling counters on it | 1 | 1 |
| if there are three or more Lesson cards in your graveyard | 1 | 1 |
| if there are three or more bloodline counters on it | 1 | 1 |
| if there are three or more ribbon counters on this creature | 1 | 1 |
| if there is a colorless creature card in your graveyard | 1 | 1 |
| if they didn't attack you that turn | 1 | 1 |
| if they own three or more exiled cards with hit counters on them | 1 | 1 |
| if this creature has seven or more ember counters on it | 1 | 1 |
| if this enchantment has three or more invitation counters on it | 1 | 1 |
| if this spell was cast from anywhere other than your hand | 1 | 1 |
| if you control a creature named Bogbrew Witch | 1 | 1 |
| if you control a creature with a counter on it | 1 | 1 |
| if you control a creature with the greatest power among creatures on the battlefield | 1 | 1 |
| if you control a modified creature | 1 | 1 |
| if you control a permanent with an oil counter on it | 1 | 1 |
| if you control an outlaw | 1 | 1 |
| if you control eight or more artifacts with the same name as one another | 1 | 1 |
| if you discarded the card with the greatest mana value among those cards or tied for greatest | 1 | 1 |
| if you don't control a Snail | 1 | 1 |
| if you don't control a legendary creature | 1 | 1 |
| if you had no cards in hand at the beginning of this turn | 1 | 1 |
| if you have fewer than three cards in hand | 1 | 1 |
| if your life total is greater than your starting life total | 1 | 1 |
| if{B}was spent to cast this spell | 1 | 1 |
| if{W}was spent to cast this spell | 1 | 1 |
| If the gift was promised | 16 | 0 |
| If a player does | 10 | 0 |
| If this spell was bargained | 9 | 0 |
| If that spell would be put into a graveyard | 8 | 0 |
| If this spell was cast using teamwork | 7 | 0 |
| If this spell's additional cost was paid | 6 | 0 |
| If that spell would be put into your graveyard | 5 | 0 |
| If you didn't put a card into your hand this way | 5 | 0 |
| If it's a nonland card | 4 | 0 |
| If evidence was collected | 3 | 0 |
| If the gift wasn't promised | 3 | 0 |
| If this spell was cast from exile | 3 | 0 |
| If this spell was foretold | 3 | 0 |
| If you've completed a dungeon | 3 | 0 |
| If a Dragon was beheld | 2 | 0 |
| If a spell cast this way would be put into a graveyard | 2 | 0 |
| If an Equipment is put onto the battlefield this way | 2 | 0 |
| If it's your main phase | 2 | 0 |
| If no one does | 2 | 0 |
| If this spell's madness cost was paid | 2 | 0 |
| If you revealed a Dragon card or controlled a Dragon as you cast this spell | 2 | 0 |
| if it's a creature card | 2 | 0 |
| If X is 1 | 1 | 0 |
| If a Hero was beheld | 1 | 0 |
| If a creature card is revealed this way | 1 | 0 |
| If a creature card was discarded this way | 1 | 0 |
| If a creature card with mana value 6 or greater is revealed this way | 1 | 0 |
| If a creature dealt damage this way would die this turn | 1 | 0 |
| If a land was destroyed this way | 1 | 0 |
| If a planeswalker card is revealed this way | 1 | 0 |
| If a player does either | 1 | 0 |
| If a spell cast this way would be put into your graveyard | 1 | 0 |
| If an artifact card was exiled this way | 1 | 0 |
| If an enchantment was destroyed this way | 1 | 0 |
| If an instant or sorcery card is exiled this way | 1 | 0 |
| If an instant or sorcery card is revealed this way | 1 | 0 |
| If an instant or sorcery spell cast this way would be put into your graveyard | 1 | 0 |
| If another creature died this turn | 1 | 0 |
| If any of those cards shares a card type with that spell | 1 | 0 |
| If cards with at least six different mana values are revealed this way | 1 | 0 |
| If denial gets more votes | 1 | 0 |
| If fewer than two cards were discarded this way | 1 | 0 |
| If he's attacking | 1 | 0 |
| If her sneak cost was paid this turn | 1 | 0 |
| If it comes up heads | 1 | 0 |
| If it doesn't have rampage | 1 | 0 |
| If it entered from your library or was cast from your library | 1 | 0 |
| If it has mana value 3 or less | 1 | 0 |
| If it shares a card type with that permanent | 1 | 0 |
| If it was a land card | 1 | 0 |
| If it was dealt noncombat damage this turn | 1 | 0 |
| If it's a Goblin creature card | 1 | 0 |
| If it's a card of the chosen type | 1 | 0 |
| If it's a creature card that shares a creature type with a creature you control | 1 | 0 |
| If it's a creature card with mana value 3 or less | 1 | 0 |
| If it's a creature or planeswalker card | 1 | 0 |
| If it's a land or double-faced card | 1 | 0 |
| If it's an artifact or creature card | 1 | 0 |
| If it's an instant or sorcery card | 1 | 0 |
| If its mana cost contains{X} | 1 | 0 |
| If return gets more votes | 1 | 0 |
| If that artifact is put into a graveyard this way | 1 | 0 |
| If that card has mana value 2 or less | 1 | 0 |
| If that card is a Hero card | 1 | 0 |
| If that creature has three or more +1/+0 counters on it | 1 | 0 |
| If that creature is a Kraken | 1 | 0 |
| If that creature is equipped | 1 | 0 |
| If that creature is red | 1 | 0 |
| If that creature shares a color with the mana that land produced | 1 | 0 |
| If that creature was a Cleric | 1 | 0 |
| If that creature was a token | 1 | 0 |
| If that enchantment is an Aura | 1 | 0 |
| If that land is a Plains | 1 | 0 |
| If that permanent is a Chandra planeswalker | 1 | 0 |
| If that permanent is destroyed this way | 1 | 0 |
| If that player doesn't | 1 | 0 |
| If that token is a Squirrel | 1 | 0 |
| If the creature the opponent controls is dealt excess damage this way | 1 | 0 |
| If the first player does | 1 | 0 |
| If the modified creature was sacrificed | 1 | 0 |
| If the total mana value of the cards exiled this way is 13 or less | 1 | 0 |
| If there are four or more chorus counters on Malcolm | 1 | 0 |
| If there are no omen counters on this enchantment | 1 | 0 |
| If they guessed wrong | 1 | 0 |
| If the{2}{R}{R}cost was paid | 1 | 0 |
| If the{2}{U}cost was paid | 1 | 0 |
| If this creature's emerge cost was paid | 1 | 0 |
| If this creature's sneak cost was paid | 1 | 0 |
| If this is the fourth time this ability has resolved this turn | 1 | 0 |
| If this spell is the first spell you've cast this game | 1 | 0 |
| If this spell was kicked with its{2}{R}kicker | 1 | 0 |
| If this spell's freerunning cost was paid | 1 | 0 |
| If this spell's mayhem cost was paid | 1 | 0 |
| If this spell's prowl cost was paid | 1 | 0 |
| If this spell's surge cost was paid | 1 | 0 |
| If those creatures have total power 8 or greater | 1 | 0 |
| If two cards that share a card type were milled this way | 1 | 0 |
| If two creatures are put onto the battlefield this way | 1 | 0 |
| If two or more cards are exiled this way | 1 | 0 |
| If you control all three | 1 | 0 |
| If you control an outlaw | 1 | 0 |
| If you didn't draw cards this way | 1 | 0 |
| If you do or if you control another Dinosaur | 1 | 0 |
| If you draw one or more cards this way | 1 | 0 |
| If you exiled a land card this way | 1 | 0 |
| If you have one or fewer cards in hand | 1 | 0 |
| If you put a Cave onto the battlefield this way | 1 | 0 |
| If you put a Town card into your hand this way | 1 | 0 |
| If you put no cards into your hand this way | 1 | 0 |
| If you return four or more nontoken permanents you control this way | 1 | 0 |
| If you reveal a creature card this way | 1 | 0 |
| If you revealed a Dragon card or chose a Dragon as you cast this spell | 1 | 0 |
| If you rolled doubles | 1 | 0 |
| If you sacrifice a snow Forest this way | 1 | 0 |
| If you sacrificed a creature this way | 1 | 0 |
| If you sacrificed a permanent this way | 1 | 0 |
| If you won five flips this way | 1 | 0 |
| If you're not the monarch | 1 | 0 |
| If you've cast a spell named Peer Through Depths and a spell named Reach Through Mists this turn | 1 | 0 |
| If you've cast another white spell this turn | 1 | 0 |
| if Johan is untapped | 1 | 0 |
| if an opponent has 0 or less life | 1 | 0 |
| if any of those cards remain exiled | 1 | 0 |
| if it attacked or blocked this turn | 1 | 0 |
| if it attacked this turn | 1 | 0 |
| if it exploited that creature | 1 | 0 |
| if it has an odd number of counters on it | 1 | 0 |
| if it has five or more hunger counters on it | 1 | 0 |
| if it has seven or more phyresis counters on it | 1 | 0 |
| if it has ten or more point counters on it | 1 | 0 |
| if it has the same name as a permanent | 1 | 0 |
| if it has three or more doom counters on it | 1 | 0 |
| if it's a land card | 1 | 0 |
| if it's a permanent card with mana value 3 or less | 1 | 0 |
| if it's a spell with lesser mana value | 1 | 0 |
| if it's a spell with mana value less than or equal to Ryan's power | 1 | 0 |
| if it's an instant or sorcery card | 1 | 0 |
| if it's an instant or sorcery spell | 1 | 0 |
| if it's an instant spell with mana value 2 or less | 1 | 0 |
| if it's your turn | 1 | 0 |
| if its mana value is odd | 1 | 0 |
| if that creature has power 7 or greater | 1 | 0 |
| if that player has more cards in hand than each other player | 1 | 0 |
| if that spell's mana value is 8 or less | 1 | 0 |
| if that target is a creature or planeswalker | 1 | 0 |
| if the exiled creature was a Thrull | 1 | 0 |
| if the gift was promised | 1 | 0 |
| if the gift was promised and that creature isn't legendary | 1 | 0 |
| if the spell's mana value is less than or equal to the amount of life you gained this turn | 1 | 0 |
| if the spell's mana value is less than the number of Mountains you control | 1 | 0 |
| if there are eight or more different names among unlocked doors of Rooms you control | 1 | 0 |
| if there are four or more charge counters on it | 1 | 0 |
| if there are no echo counters on it | 1 | 0 |
| if there are seven or more time counters on it | 1 | 0 |
| if there are three or more cards exiled with Profane Procession | 1 | 0 |
| if there are three or more collection counters on it | 1 | 0 |
| if there are three or more dread counters on it | 1 | 0 |
| if there are three or more judgment counters on it | 1 | 0 |
| if this creature has five or more filibuster counters on it | 1 | 0 |
| if this spell was cast using teamwork | 1 | 0 |
| if this spell's additional cost was paid | 1 | 0 |
| if you control exactly thirteen permanents | 1 | 0 |
| if you control that creature | 1 | 0 |
| if you exiled four or more cards this way | 1 | 0 |
| if you have seven or more cards in your graveyard | 1 | 0 |

## Modeled-capability envelope gaps (parameter backlog)

Families the compiler recognizes but lowers only within an exact envelope. Each row is one supported-envelope blocker ranked by sole blockers (cards it is the only blocker for); growing the envelope to cover the example wordings unblocks the listed cards. This is the effect-family analogue of the unrecognized-condition backlog above.

| Capability | Supported envelope (blocker) | Affected cards | Sole blockers | Example wordings |
| --- | --- | ---: | ---: | --- |
| unsupported counter placement | the executable source backend supports exact recognized counter placement on one valid target | 457 | 213 | put the cards exiled with it into their owner's hand.; remove an intervention counter from this enchantment.; put a name sticker on a nonland permanent you own. |
| unsupported return spell | the executable source backend supports only exact return of one target permanent to its owner's hand | 306 | 203 | return this card from your graveyard to the battlefield attached to that creature.; Return target creature or Vehicle an opponent controls to its owner's hand.; • Return target nonland permanent to its owner's hand. |
| unsupported destroy spell | the executable source backend supports only exact destruction of one target permanent | 253 | 187 | Destroy all non-Wall creatures blocking enchanted creature.; Destroy target creature with power less than this creature's power.; Destroy each creature whose coin comes up tails. |
| unsupported enters-tapped replacement | the executable source backend supports only exact unconditional self enters-tapped replacements | 422 | 186 | If a source you control would deal noncombat damage to a creature an opponent controls, pu…; If a player would begin an extra turn, that player skips that turn instead.; If a creature you control would explore, instead it explores, then it explores again. |
| unsupported token creation | the executable source backend supports only a single fixed-power/toughness creature token with one subtype and at most one color | 314 | 176 | Create a Wicked Role token attached to up to one target creature you control. (If you cont…; • Create a 1/1 colorless Eldrazi Scion creature token with "Sacrifice this token: Add {C…; • Create a 3/3 black Dalek artifact creature token with menace. |
| unsupported damage spell | the executable source backend supports only exact supported damage amounts to one target | 214 | 161 | Choose target creature. Whenever that creature is dealt damage this turn, it deals that mu…; Armed Response deals damage to target attacking creature equal to the number of Equipment…; Close Encounter deals damage equal to the power of the chosen creature or card to target c… |
| unsupported phase/step trigger phrase | the executable source backend does not support this intervening-if condition | 268 | 158 | At the beginning of your upkeep, if you control no Thopters other than this creature, retu…; At the beginning of each end step, if an opponent discarded a card this turn, you draw a c…; At the beginning of combat on your turn, if two or more players have lost the game, gain c… |
| unsupported temporary keyword spell | the executable source backend supports only exact non-parameterized keyword grants to one target creature or permanent until end of turn | 147 | 114 | unless you pay {R}, whenever this creature blocks or becomes blocked by a creature this co…; If you do, it gains flying until end of turn.; Until end of turn, target creature you control gains indestructible and "Whenever this cre… |
| unsupported power/toughness spell | the executable source backend supports only exact supported target-creature power/toughness changes until end of turn | 178 | 108 | • Target creature gets +3/-3 until end of turn.; you get {TK}.; Until your next turn, up to one target creature gets -3/-0. |
| unsupported exile spell | the executable source backend supports only exact exile of one target permanent | 210 | 102 | Exile all multicolored permanents.; exile them from your graveyard.; exile the top five cards of your library face down. |
| unsupported permanent zone-change trigger | the executable source backend does not support this semantic permanent zone-change trigger condition | 151 | 90 | Whenever one or more creatures you control enter, if one or more of them entered from a gr…; When this creature enters, if it was kicked twice, it deals 2 damage to any target.; Whenever another creature you control enters, if it has greater power or toughness than Hu… |
| unsupported search effect | the executable source backend supports only exact unconditional library-search sequences | 153 | 84 | you may search your library for a Food card, reveal it, put it into your hand, then shuffl…; each player searches their library for up to two basic land cards, puts them onto the batt…; you may search your library for up to two artifact, creature, and/or enchantment cards wit… |
| unsupported life spell | the executable source backend supports only exact fixed life changes | 87 | 66 | Each opponent with no cards in hand loses 10 life.; r leaves the battlefield, you lose 3 life if; you and that player each gain that much life. |
| unsupported damage spell | the executable source backend supports only exact fixed or X group damage amounts | 86 | 65 | The next time a source of your choice of the chosen color would deal damage to you this tu…; creature spell. Whenever you cast a noncreature spell, if at least four mana was spent to…; this creature deals X damage to that player, where X is the number of cards in their hand… |
| unsupported counter spell | the executable source backend supports only exact counter of one target spell | 79 | 63 | counter all abilities your opponents control.; Counter target spell with mana value X. (For example, if that spell's mana cost is {3}{U}{…; counter that spell or ability. |
| unsupported draw spell | the executable source backend supports only exact fixed card draw | 95 | 60 | Draw a card for each tapped creature target opponent controls.; Target player skips their next draw step.; draw cards equal to its power. |
| unsupported permanent zone-change trigger effect | the executable source backend does not support this permanent zone-change trigger body | 274 | 59 | When this enchantment enters, suspect target creature an opponent controls. As long as thi…; Whenever a nontoken creature you control enters, you may pay {2}. If you do, create a toke…; When this Equipment enters, you may pay {3}{U}. If you do, create a 4/4 blue Giant Wizard… |
| unsupported gain-control spell | the executable source backend supports only exact gain-control sequences targeting one permanent | 79 | 54 | Gain control of target creature until end of turn. Untap that creature. It gains haste unt…; Gain control of target creature. Untap that creature. It gains haste until end of turn. Sa…; gain control of target creature an opponent controls until end of turn. Untap that creatur… |
| unsupported phase/step trigger phrase effect | the executable source backend does not support this phase/step trigger body | 130 | 53 | At the beginning of your upkeep, you may flip a coin. If you win the flip, put a fuse coun…; At the beginning of your end step, if a creature died this turn, venture into the dungeon.…; At the beginning of your upkeep, if you control a white or blue permanent, this enchantmen… |
| unsupported triggered ability effect | the executable source backend supports only recognized semantic self triggers with supported effects | 103 | 50 | When this creature attacks, up to one target creature defending player controls blocks it…; Whenever Ma Chao attacks alone, it can't be blocked this combat.; Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it a… |
| unsupported enters-with-counters replacement | the executable source backend supports only exact self enters-with-counters replacements | 83 | 44 | If this creature was kicked with its {1}{U} kicker, it enters with two +1/+1 counters on i…; If this creature was kicked with its {B} kicker, it enters with a +1/+1 counter on it and…; This creature enters with two +1/+1 counters on it if it wasn't cast or no mana was spent… |
| unsupported gain-control spell | the executable source backend supports only exact gain-control of one target permanent | 69 | 43 | Gain control of target creature. Change the text of that creature by replacing all instanc…; Gain control of all lands target player controls.; Gain control of target creature with power less than or equal to the number of treasure co… |
| unsupported tap spell | the executable source backend supports only exact tap of one target permanent | 63 | 43 | tap all creatures defending player controls.; Tap X target nonland permanents.; tap target creature an opponent controls. |
| unsupported shuffle effect | the executable source backend supports only a source-spell shuffle into its owner's library or a controller graveyard shuffle into library | 56 | 43 | Choose target artifact or enchantment. Its owner shuffles it into their library.; shuffle it into its owner's library.; Target player shuffles up to four target cards from their graveyard into their library. |
| unsupported attach effect | the executable source backend supports only "attach it/that &lt;Equipment&gt; to target &lt;permanent&gt; you control" attaching the entering or source permanent | 63 | 42 | Attach target Aura attached to a creature to another creature.; attach this Equipment to that creature.; Attach this Equipment to target creature you control. |
| unsupported damage spell | the executable source backend supports only exact fixed group damage amounts | 48 | 37 | Flame Sweep deals 2 damage to each creature except for creatures you control with flying.; This creature deals X damage to each creature and each player. Spend only black mana on X.; it deals 7 damage to a creature an opponent controls chosen at random. |
| unsupported life spell | the executable source backend supports only exact supported life changes | 54 | 36 | defending player loses 1 life for each card in their graveyard.; Each opponent loses life equal to the number of cards in their graveyard.; that player gains X life, where X is the number of cards in all graveyards with the same n… |
| unsupported search effect | the executable source backend supports only searches of your library or a single target player's library ending with "then shuffle" | 55 | 35 | For each player, choose friend or foe. Each friend searches their library for a land card,…; Search target player's library for X cards, where X is the number of cards in your hand, a…; • Search your library for a basic land card, put it onto the battlefield tapped, then sh… |
| unsupported library placement | the executable source backend supports only exact target graveyard-to-library placement | 53 | 35 | Put up to three target cards from an opponent's graveyard on top of their library in any o…; target opponent puts a card from their hand on top of their library.; put it on the bottom of its owner's library. |
| unsupported triggered ability | the executable source backend supports only recognized semantic self triggers with supported effects | 42 | 35 | Whenever Whiplash attacks, if he's equipped, each opponent loses X life and you gain X lif…; Pack tactics — Whenever this creature attacks, if you attacked with creatures with total…; Whenever Ian the Reckless attacks, if it's modified, you may have it deal damage equal to… |
| unsupported Enchant ability | the executable source backend supports only exact Enchant with a supported target kind | 52 | 28 | Gain control of target Aura that's attached to a permanent. Attach it to another permanent…; At the beginning of combat on your turn, create a green Aura enchantment token named Settl…; Enchant nonland permanent |
| unsupported power/toughness spell | the executable source backend supports only exact until-end-of-turn power/toughness changes to the source or referenced permanent | 39 | 28 | This creature gets +1/+1 until end of turn. Any player may activate this ability.; This creature gets -1/-1 until end of turn. Any player may activate this ability.; Until end of turn, this creature gets +1/+1 for each permanent card in your graveyard. |
| unsupported damage spell | the executable source backend does not support this group recipient | 34 | 28 | Meteor Blast deals 4 damage to each of X targets.; Zo-Zu deals 2 damage to that land's controller.; Firestorm deals X damage to each of X targets. |
| unsupported delayed effect | the executable source backend supports only exact non-target delayed one-shot effects | 28 | 23 | Destroy target blocking creature at end of combat.; put a -1/-1 counter on it at end of combat.; return that card to the battlefield under your control at the beginning of the next end st… |
| unsupported untap spell | the executable source backend supports only exact untap of one target permanent | 42 | 22 | After the second main phase this turn, there's an additional combat phase followed by an a…; untap each enchanted permanent you control.; Untap up to one target creature and up to one target land. |
| unsupported triggered ability | the executable source backend does not support this semantic spell-cast trigger condition | 36 | 21 | Whenever a player casts a spell, if it's not their turn, that player draws a card.; Whenever you cast a spell, if that spell was kicked, put a +1/+1 counter on Hallar, then H…; Whenever a player casts a spell, if no colored mana was spent to cast it, counter that spe… |
| unsupported can't-block effect | the executable source backend supports only exact "&lt;targets&gt; can't block this turn." | 26 | 20 | target creature defending player controls with power less than this creature's power can't…; Target creature can't block this turn and becomes a Coward in addition to its other types…; It can't block this turn. |
| unsupported card layout | the source generator does not support Scryfall layout "flip" | 20 | 20 | - |
| unsupported triggered ability effect | the executable source backend does not support this life or damage trigger body | 54 | 19 | Whenever one or more creatures you control deal combat damage to a player, this creature b…; Whenever a creature you control deals combat damage to a player, goad each creature that p…; Whenever one or more creatures you control deal combat damage to a player, you may return… |
| unsupported triggered ability | the executable source backend does not support this semantic trigger condition | 29 | 19 | Whenever you activate an ability that isn't a mana ability, if life was paid to activate i…; Whenever a player attacks you, if that player has another opponent who isn't being attacke…; Whenever a Wolf you control attacks, if Tolsimir attacked this combat, target creature an… |
