# Card-Support Planning Report

Capability-aware blockers for eligible paper cards that cannot yet be generated. Each distinct diagnostic summary and capability is counted at most once per card.

## Diagnostic reasons

A sole blocker is the card's only distinct diagnostic summary. The most common co-blocker excludes the reason in its own row.

| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | unsupported ordered effect sequence | 4,062 | 2,277 | 56.1% | unsupported optional effect |
| 2 | unsupported Oracle construct | 2,747 | 0 | 0.0% | unsupported static ability |
| 3 | unsupported static ability | 1,984 | 378 | 19.1% | unsupported Oracle construct |
| 4 | unsupported triggered ability | 1,213 | 763 | 62.9% | unsupported Oracle construct |
| 5 | unsupported optional effect | 1,141 | 2 | 0.2% | unsupported ordered effect sequence |
| 6 | unsupported ability content | 1,021 | 81 | 7.9% | unsupported Oracle construct |
| 7 | unsupported reference subject | 591 | 0 | 0.0% | unsupported ordered effect sequence |
| 8 | unsupported counter placement | 486 | 198 | 40.7% | unsupported ordered effect sequence |
| 9 | unsupported static declaration operation | 450 | 315 | 70.0% | unsupported Oracle construct |
| 10 | unsupported damage spell | 433 | 284 | 65.6% | unsupported reference subject |
| 11 | unsupported activation cost | 422 | 165 | 39.1% | unsupported cost |
| 12 | unsupported enters-tapped replacement | 418 | 200 | 47.8% | unsupported Oracle construct |
| 13 | unsupported ability word | 363 | 136 | 37.5% | unsupported Oracle construct |
| 14 | unsupported static declaration group | 331 | 205 | 61.9% | unsupported Oracle construct |
| 15 | unsupported static declaration condition | 315 | 218 | 69.2% | unsupported static ability |
| 16 | unsupported return spell | 304 | 198 | 65.1% | unsupported Oracle construct |
| 17 | unsupported token creation | 304 | 174 | 57.2% | unsupported ordered effect sequence |
| 18 | unsupported phase/step trigger phrase | 301 | 204 | 67.8% | unsupported Oracle construct |
| 19 | unsupported search effect | 255 | 117 | 45.9% | unsupported optional effect |
| 20 | unsupported destroy spell | 242 | 169 | 69.8% | unsupported Oracle construct |
| 21 | unsupported exile spell | 225 | 110 | 48.9% | unsupported ordered effect sequence |
| 22 | unsupported power/toughness spell | 220 | 129 | 58.6% | unsupported static ability |
| 23 | unsupported permanent zone-change trigger effect | 204 | 46 | 22.5% | unsupported Oracle construct |
| 24 | unsupported cast effect | 199 | 77 | 38.7% | unsupported optional effect |
| 25 | unsupported activation ability word | 191 | 130 | 68.1% | unsupported Oracle construct |
| 26 | unsupported mixed keyword ability | 185 | 102 | 55.1% | unsupported static ability |
| 27 | unsupported activation condition | 182 | 109 | 59.9% | unsupported Oracle construct |
| 28 | unsupported cost | 168 | 0 | 0.0% | unsupported activation cost |
| 29 | unsupported permanent zone-change trigger | 157 | 102 | 65.0% | unsupported Oracle construct |
| 30 | unsupported activation references | 151 | 110 | 72.8% | unsupported Oracle construct |
| 31 | unsupported gain-control spell | 141 | 89 | 63.1% | unsupported static ability |
| 32 | unsupported temporary keyword spell | 139 | 108 | 77.7% | unsupported Oracle construct |
| 33 | unsupported triggered ability effect | 134 | 68 | 50.7% | unsupported Oracle construct |
| 34 | unsupported life spell | 126 | 91 | 72.2% | unsupported ordered effect sequence |
| 35 | unsupported enters-with-counters replacement | 117 | 59 | 50.4% | unsupported ordered effect sequence |
| 36 | unsupported sacrifice spell | 111 | 80 | 72.1% | unsupported ordered effect sequence |
| 37 | unsupported draw spell | 103 | 66 | 64.1% | unsupported Oracle construct |
| 38 | unsupported phase/step trigger phrase effect | 96 | 48 | 50.0% | unsupported Oracle construct |
| 39 | unsupported counter spell | 69 | 50 | 72.5% | unsupported Oracle construct |
| 40 | unsupported optional replacement effect | 64 | 52 | 81.2% | unsupported Oracle construct |
| 41 | unsupported type line | 61 | 60 | 98.4% | unsupported ordered effect sequence |
| 42 | unsupported mana effect | 59 | 34 | 57.6% | unsupported static declaration condition |
| 43 | unsupported keyword or ability grant | 58 | 45 | 77.6% | unsupported Oracle construct |
| 44 | unsupported mana symbol | 54 | 38 | 70.4% | unsupported static ability |
| 45 | unsupported library placement | 54 | 25 | 46.3% | unsupported reference subject |
| 46 | unsupported attach effect | 51 | 33 | 64.7% | unsupported static declaration operation |
| 47 | unsupported Enchant ability | 50 | 30 | 60.0% | unsupported ordered effect sequence |
| 48 | unsupported parameterized keyword | 50 | 24 | 48.0% | unsupported ordered effect sequence |
| 49 | unsupported activation timing | 47 | 39 | 83.0% | unsupported Oracle construct |
| 50 | unsupported shuffle effect | 44 | 30 | 68.2% | unsupported reference subject |
| 51 | unsupported static declaration duration | 40 | 32 | 80.0% | unsupported Oracle construct |
| 52 | unsupported untap spell | 40 | 21 | 52.5% | unsupported ordered effect sequence |
| 53 | unsupported tap spell | 39 | 31 | 79.5% | unsupported static ability |
| 54 | unsupported target count parameter | 39 | 28 | 71.8% | unsupported activation cost |
| 55 | unsupported keyword or ability loss | 38 | 23 | 60.5% | unsupported static ability |
| 56 | incomplete executable lowering | 37 | 31 | 83.8% | unsupported static ability |
| 57 | unsupported Equip ability | 36 | 13 | 36.1% | unsupported static ability |
| 58 | unsupported alternative effects | 29 | 21 | 72.4% | unsupported Oracle construct |
| 59 | unsupported discard spell | 27 | 17 | 63.0% | unsupported ordered effect sequence |
| 60 | unsupported emblem ability | 27 | 10 | 37.0% | unsupported ordered effect sequence |
| 61 | unsupported keyword ability | 24 | 6 | 25.0% | unsupported Oracle construct |
| 62 | unsupported group power/toughness spell | 22 | 19 | 86.4% | unsupported counter placement |
| 63 | unsupported card layout | 20 | 20 | 100.0% | - |
| 64 | unsupported can't-block effect | 20 | 14 | 70.0% | unsupported reference subject |
| 65 | unsupported can't-be-blocked effect | 19 | 16 | 84.2% | unsupported Oracle construct |
| 66 | unsupported ability modes | 18 | 13 | 72.2% | unsupported Oracle construct |
| 67 | unsupported overload effect | 17 | 7 | 41.2% | unsupported ordered effect sequence |
| 68 | unsupported loyalty ability | 17 | 1 | 5.9% | unsupported ordered effect sequence |
| 69 | unsupported alternative spell cost | 16 | 9 | 56.2% | unsupported damage spell |
| 70 | unsupported copy effect | 16 | 5 | 31.2% | unsupported ordered effect sequence |
| 71 | unsupported multiple spell abilities | 14 | 13 | 92.9% | unsupported Oracle construct |
| 72 | validation failed: invalid-ability-body | 13 | 13 | 100.0% | - |
| 73 | unsupported delayed object capture | 13 | 8 | 61.5% | unsupported ordered effect sequence |
| 74 | unsupported manifest spell | 13 | 8 | 61.5% | unsupported ordered effect sequence |
| 75 | unsupported fight spell | 13 | 6 | 46.2% | unsupported Oracle construct |
| 76 | unsupported unknown ability | 13 | 0 | 0.0% | unsupported Oracle construct |
| 77 | unsupported Flashback ability | 11 | 6 | 54.5% | unsupported ordered effect sequence |
| 78 | unsupported Protection ability | 10 | 7 | 70.0% | unsupported Oracle construct |
| 79 | unsupported mill spell | 9 | 7 | 77.8% | unsupported destroy spell |
| 80 | unsupported Crew ability | 9 | 6 | 66.7% | unsupported Oracle construct |
| 81 | unsupported delayed trigger | 9 | 6 | 66.7% | unsupported Oracle construct |
| 82 | unsupported double effect | 9 | 6 | 66.7% | unsupported activation ability word |
| 83 | unsupported enters-as-copy replacement | 8 | 8 | 100.0% | - |
| 84 | validation failed: oracle-without-abilities | 8 | 8 | 100.0% | - |
| 85 | unsupported draw/discard trigger effect | 8 | 7 | 87.5% | unsupported Oracle construct |
| 86 | unsupported investigate spell | 8 | 6 | 75.0% | unsupported mana symbol |
| 87 | unsupported connive effect | 8 | 5 | 62.5% | unsupported Oracle construct |
| 88 | unsupported prevent-damage effect | 8 | 5 | 62.5% | unsupported activation cost |
| 89 | unsupported goad spell | 8 | 3 | 37.5% | unsupported reference subject |
| 90 | unsupported Max speed triggered ability | 7 | 7 | 100.0% | - |
| 91 | validation failed: invalid-selection | 7 | 7 | 100.0% | - |
| 92 | unsupported divided damage spell | 7 | 6 | 85.7% | unsupported Oracle construct |
| 93 | unsupported Cumulative upkeep ability | 7 | 4 | 57.1% | unsupported life spell |
| 94 | unsupported damage prevention replacement | 7 | 4 | 57.1% | unsupported Oracle construct |
| 95 | unsupported transform effect | 7 | 3 | 42.9% | unsupported Oracle construct |
| 96 | unsupported gain player counter spell | 7 | 2 | 28.6% | unsupported optional effect |
| 97 | unsupported lose-game effect | 6 | 3 | 50.0% | unsupported Oracle construct |
| 98 | unsupported Read ahead ability | 6 | 0 | 0.0% | unsupported ordered effect sequence |
| 99 | unsupported Myriad ability | 5 | 4 | 80.0% | unsupported overload effect |
| 100 | unsupported Persist ability | 5 | 4 | 80.0% | unsupported counter placement |

## Capability clusters

A fully unlockable card has every distinct diagnostic summary in one capability cluster. Constituent summaries list the diagnostics currently observed in that cluster.

| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |
| --- | ---: | ---: | --- |
| shared-ability-content | 7,927 | 4,571 | unsupported ability content; unsupported ability modes; unsupported counter placement; unsupported counter spell; unsupported damage spell; unsupported delayed effect; unsupported destroy spell; unsupported discard spell; unsupported draw spell; unsupported draw/discard trigger effect; unsupported exile spell; unsupported explore spell; unsupported fight spell; unsupported gain-control spell; unsupported group power/toughness spell; unsupported investigate spell; unsupported keyword or ability grant; unsupported keyword or ability loss; unsupported library placement; unsupported life spell; unsupported mana effect; unsupported mana symbol; unsupported manifest spell; unsupported mill spell; unsupported multiple spell abilities; unsupported ordered effect sequence; unsupported phase/step trigger phrase effect; unsupported power/toughness spell; unsupported proliferate spell; unsupported regenerate spell; unsupported return spell; unsupported scry spell; unsupported search effect; unsupported tap spell; unsupported temporary keyword spell; unsupported triggered ability effect; unsupported untap spell |
| static-declaration | 3,269 | 1,351 | unsupported Enchant ability; unsupported Protection ability; unsupported Read ahead ability; unsupported keyword ability; unsupported mixed keyword ability; unsupported parameterized keyword; unsupported static ability; unsupported static declaration condition; unsupported static declaration duration; unsupported static declaration group; unsupported static declaration operation; unsupported static declaration shell |
| other | 3,179 | 1,068 | incomplete executable lowering; unsupported Bestow ability; unsupported Bloodthirst ability; unsupported Channel ability; unsupported Class level ability; unsupported Crew ability; unsupported Cumulative upkeep ability; unsupported Dash ability; unsupported Embalm ability; unsupported Evoke ability; unsupported Flanking ability; unsupported Flashback ability; unsupported Foretell ability; unsupported Level up ability; unsupported Max speed ability; unsupported Max speed triggered ability; unsupported Myriad ability; unsupported Offspring ability; unsupported Persist ability; unsupported Plot ability; unsupported Spectacle ability; unsupported Tempting offer; unsupported Undying ability; unsupported Unearth ability; unsupported activation restriction; unsupported adapt spell; unsupported alternative effects; unsupported alternative spell cost; unsupported amass spell; unsupported attach effect; unsupported become-a-copy effect; unsupported bolster spell; unsupported can't-attack effect; unsupported can't-attack-or-block effect; unsupported can't-be-blocked effect; unsupported can't-block effect; unsupported can-attack-as-though-defender effect; unsupported card layout; unsupported cast effect; unsupported choose effect; unsupported color-change effect; unsupported connive effect; unsupported copy effect; unsupported damage prevention replacement; unsupported delayed object capture; unsupported delayed trigger; unsupported discard-then-draw spell; unsupported discover spell; unsupported divided damage spell; unsupported double counters spell; unsupported double effect; unsupported draw-doubling replacement; unsupported draw-from-empty-library win replacement; unsupported emblem ability; unsupported emblem effect; unsupported enters-as-copy replacement; unsupported entry-choice replacement; unsupported forced-attack effect; unsupported gain player counter spell; unsupported goad spell; unsupported graveyard-redirect replacement; unsupported hand-library amount; unsupported hand-library resolving cost; unsupported hand-library source owner; unsupported historical untap; unsupported impulse exile effect; unsupported incubate spell; unsupported keep-one-per-type sacrifice; unsupported life action duration; unsupported life payment; unsupported life-gain replacement; unsupported linked X spell cost; unsupported look-at-hand spell; unsupported lose-game effect; unsupported monstrosity spell; unsupported optional effect; unsupported optional replacement effect; unsupported overload effect; unsupported permanent choice; unsupported permanent union scope; unsupported permanent zone-change trigger; unsupported permanent zone-change trigger effect; unsupported phase out effect; unsupported phase out spell; unsupported polymorph effect; unsupported populate spell; unsupported prevent-damage effect; unsupported reference subject; unsupported retarget effect; unsupported ring tempts effect; unsupported sacrifice spell; unsupported set base power/toughness effect; unsupported shuffle effect; unsupported source-spell cost reduction; unsupported surveil spell; unsupported tap or untap spell; unsupported target count parameter; unsupported target-animation effect; unsupported token creation; unsupported transform effect; unsupported trigger source zone; unsupported turn-face-down effect; unsupported type line; unsupported win-game effect; validation failed: invalid-ability-body; validation failed: invalid-rule-effect; validation failed: invalid-selection; validation failed: oracle-without-abilities |
| trigger-pattern | 1,503 | 982 | unsupported draw/discard trigger; unsupported phase/step trigger phrase; unsupported triggered ability |
| activation | 1,050 | 666 | unsupported Cycling ability; unsupported Equip ability; unsupported Mutate ability; unsupported Ninjutsu ability; unsupported activation ability word; unsupported activation condition; unsupported activation cost; unsupported activation modes; unsupported activation references; unsupported activation timing; unsupported activation zone; unsupported cost; unsupported loyalty ability |
| replacement | 544 | 268 | unsupported conditional enters-tapped replacement; unsupported damage replacement; unsupported enters-tapped replacement; unsupported enters-with-counters replacement; unsupported self zone-destination replacement; unsupported token-creation replacement |
| recognition-fallback | 2,922 | 231 | unsupported Oracle construct; unsupported ability word; unsupported unknown ability |

## Unblock roadmap

Greedy set-cover priority: each step fixes the reason that — given the reasons already fixed in the steps above it — newly fully unblocks the most still-blocked cards. Cumulative is the running total of cards fully unblocked. Fan-out lowerers (ordered sequence, modal, optional) now report every independent blocker they carry, so these counts account for co-blockers rather than crediting a fix with cards that need other fixes too. A few remaining lowerers still short-circuit within an ability, so the counts stay a slight over-estimate.

| Step | Fix this reason | Capability | Newly unblocked | Cumulative | Sample cards |
| ---: | --- | --- | ---: | ---: | --- |
| 1 | unsupported ordered effect sequence | shared-ability-content | 2,277 | 2,277 | Abdel Adrian, Gorion's Ward, Abigale, Eloquent First-Year, Abnormal Endurance, Abstergo Entertainment, Abstruse Interference |
| 2 | unsupported triggered ability | trigger-pattern | 795 | 3,072 | A Good Day to Pie, Aang and Katara, Aboleth Spawn, Abomination, Abomination, Irradiated Brute |
| 3 | unsupported optional effect | other | 708 | 3,780 | Absorb Identity, Abstruse Archaic, Academy Loremaster, Ace, Fearless Rebel, Act of Authority |
| 4 | unsupported static ability | static-declaration | 455 | 4,235 | Absorbing Man and Titania, Abyssal Persecutor, Aerial Modification, Ahn-Crop Champion, Ahn-Crop Crasher |
| 5 | unsupported Oracle construct | recognition-fallback | 985 | 5,220 | "Name Sticker" Goblin, Abigale, Poet Laureate // Heroic Stanza, Abominable Treefolk, Abomination of Llanowar, Abomination, World Ravager |
| 6 | unsupported ability content | shared-ability-content | 784 | 6,004 | A Little Chat, About Face, Absorbing Man, Abuna's Chant, Academic Ascent |
| 7 | unsupported static declaration operation | static-declaration | 360 | 6,364 | Aboshan's Desire, Abzan Runemark, Acidic Sliver, Acolyte of Bahamut, Adelbert Steiner |
| 8 | unsupported enters-tapped replacement | replacement | 339 | 6,703 | Aberrant Return, Aether Refinery, Aether Revolt, Alhammarret, High Arbiter, Ali from Cairo |
| 9 | unsupported damage spell | shared-ability-content | 314 | 7,017 | Acolyte's Reward, Advanced Reconstruction, Aether Flash, Ajani Unrelenting, Ajani's Aid |
| 10 | unsupported ability word | recognition-fallback | 310 | 7,327 | Abaddon the Despoiler, Aboroth, Aeon Chronicler, Alisaie Leveilleur, Alliance of Arms |
| 11 | unsupported counter placement | shared-ability-content | 302 | 7,629 | Academy Researchers, Acrobatic Cheerleader, Adder-Staff Boggart, Aetherborn Marauder, Aetherspouts |
| 12 | unsupported reference subject | other | 356 | 7,985 | Abstruse Appropriation, Abyssal Harvester, Academic Dispute, Acidic Soil, Aether Gust |
| 13 | unsupported static declaration condition | static-declaration | 281 | 8,266 | Aang, A Lot to Learn, Ace's Baseball Bat, Alirios, Enraptured, Angelic Voices, Animus of Predation |
| 14 | unsupported static declaration group | static-declaration | 278 | 8,544 | A Tale for the Ages, Adventurers' Guildhouse, Aetherflame Wall, Aminatou, Veil Piercer, Angel of Jubilation |
| 15 | unsupported phase/step trigger phrase | trigger-pattern | 271 | 8,815 | Afflicted Deserter // Werewolf Ransacker, Agent of Treachery, Air Nomad Student, Akuta, Born of Ash, Arachnus Web |
| 16 | unsupported return spell | shared-ability-content | 266 | 9,081 | Accursed Witch // Infectious Curse, Adarkar Valkyrie, Alchemist's Retrieval, Ancestral Statue, Aphetto Dredging |
| 17 | unsupported token creation | other | 257 | 9,338 | Aatchik, Emerald Radian, Abby, Merciless Soldier, Aerid Konstrari, Ajani Goldmane, Ajani Resolute |
| 18 | unsupported activation cost | activation | 246 | 9,584 | A Killer Among Us, Abandon Hope, Aether Tide, Alms, Altar of Bhaal // Bone Offering |
| 19 | unsupported search effect | shared-ability-content | 236 | 9,820 | Aang's Journey, Acquire, Aether Searcher, Agency Outfitter, Alpine Houndmaster |
| 20 | unsupported destroy spell | shared-ability-content | 222 | 10,042 | Abu Ja'far, Aether Storm, Age of Ultron, Ajani Vengeant, Alaborn Zealot |
| 21 | unsupported exile spell | shared-ability-content | 202 | 10,244 | Admonition Angel, Agrus Kos, Spirit of Justice, Aligned Hedron Network, All Hallow's Eve, Angel of Condemnation |
| 22 | unsupported power/toughness spell | shared-ability-content | 196 | 10,440 | Acquired Mutation, Aethertide Whale, Alistair, the Brigadier, All-Seeing Arbiter, Allied Assault |
| 23 | unsupported permanent zone-change trigger effect | other | 195 | 10,635 | "Lifetime" Pass Holder, Aberrant Mind Sorcerer, Aerie Auxiliary, Airbender Ascension, Anafenza, Unyielding Lineage |
| 24 | unsupported cast effect | other | 185 | 10,820 | Aberrant Manawurm, Abeyance, Abstract Performance, Academic Probation, Aisha of Sparks and Smoke |
| 25 | unsupported activation ability word | activation | 170 | 10,990 | Abomination, Terrifying Titan, Adorned Crocodile, Adric, Mathematical Genius, Aerial Doombot, Afterburner Expert |
| 26 | unsupported activation condition | activation | 171 | 11,161 | Aclazotz, Deepest Betrayal // Temple of the Dead, Alluring Siren, Altar of the Pantheon, Amulet of Quoz, Animal Attendant |
| 27 | unsupported mixed keyword ability | static-declaration | 162 | 11,323 | A Mysterious Creature, Animate Wall, Aragorn, Hornburg Hero, Arcades, the Strategist, Archetype of Aggression |
| 28 | unsupported cost | activation | 156 | 11,479 | Aang's Iceberg, Adagia, Windswept Bastion, Aetherflux Conduit, Aethersquall Ancient, Aethertorch Renegade |
| 29 | unsupported permanent zone-change trigger | other | 154 | 11,633 | Aang, Airbending Master, Aang, Swift Savior // Aang and La, Ocean's Fury, Aang, the Last Airbender, Acererak the Archlich, Alex Wilder, Runaway |
| 30 | unsupported activation references | activation | 149 | 11,782 | Aegis of Honor, Akiri, Fearless Voyager, Aladdin's Lamp, Allosaurus Shepherd, Aphelia, Viper Whisperer |

## Ordered effect sequence sub-categories

Breakdown of the `unsupported ordered effect sequence` reason by the specific blocker within the sequence. A `sub-effect` row names the single-effect lowering a clause needs before its sequence can compile; a `structural` row names a sequence-machinery limitation. Counts mirror the diagnostic-reasons table: affected cards include co-blocked cards, sole blockers do not.

| Category | Affected cards | Sole blockers |
| --- | ---: | ---: |
| sub-effect — unsupported counter placement | 643 | 314 |
| structural — per-effect condition unrecognized | 510 | 310 |
| sub-effect — unsupported ability content | 549 | 275 |
| sub-effect — unsupported token creation | 195 | 146 |
| sub-effect — unsupported cast effect | 336 | 138 |
| sub-effect — unsupported damage spell | 183 | 132 |
| sub-effect — unsupported power/toughness spell | 175 | 129 |
| sub-effect — unsupported exile spell | 316 | 123 |
| sub-effect — unsupported return spell | 181 | 121 |
| sub-effect — unsupported temporary keyword spell | 161 | 119 |
| structural — per-effect condition spans multiple clauses | 223 | 113 |
| sub-effect — unsupported life spell | 160 | 113 |
| sub-effect — unsupported draw spell | 168 | 100 |
| sub-effect — unsupported discard spell | 118 | 75 |
| sub-effect — unsupported sacrifice spell | 106 | 72 |
| sub-effect — unsupported untap spell | 75 | 54 |
| sub-effect — unsupported shuffle effect | 87 | 46 |
| sub-effect — unsupported destroy spell | 57 | 46 |
| sub-effect — unsupported keyword or ability loss | 55 | 45 |
| sub-effect — unsupported keyword or ability grant | 75 | 40 |
| sub-effect — unsupported library placement | 126 | 33 |
| sub-effect — unsupported tap spell | 64 | 31 |
| sub-effect — unsupported delayed object capture: fixed-phase subject producer has no modeled actual-result publication | 40 | 29 |
| structural — per-effect condition kind not gateable | 39 | 27 |
| structural — multi-effect body not lowered as a sequence | 36 | 27 |
| structural — actual-result condition has no unique earlier typed producer | 123 | 26 |
| search unit — unsupported contiguous search/filter/destination parameters | 72 | 21 |
| structural — coin flip branch not lowered | 22 | 20 |
| sub-effect — unsupported attach effect | 37 | 18 |
| structural — per-effect condition has no containing clause | 26 | 18 |
| sub-effect — unsupported gain-control spell | 21 | 15 |
| structural — unconsumed targets/references/keywords | 17 | 14 |
| sub-effect — unsupported manifest spell | 93 | 13 |
| structural — inherited target not remappable | 18 | 13 |
| sub-effect — unsupported group power/toughness spell | 15 | 13 |
| sub-effect — unsupported gain player counter spell | 16 | 12 |
| sub-effect — unsupported fight spell | 15 | 11 |
| sub-effect — unsupported can't-block effect | 14 | 11 |
| sub-effect — unsupported mill spell | 15 | 10 |
| sub-effect — unsupported mana symbol | 16 | 9 |
| structural — per-effect condition lowering failed | 12 | 9 |
| structural — instead replacement not gatable | 11 | 9 |
| sub-effect — unsupported double effect | 11 | 9 |
| mode 1: sub-effect — unsupported ability content | 12 | 8 |
| structural — single effect requires ordered lowering | 11 | 8 |
| structural — unsupported linked counter and token creation | 8 | 8 |
| sub-effect — unsupported mana effect | 12 | 7 |
| sub-effect — unsupported investigate spell | 10 | 7 |
| structural — unsupported sacrifice-conditioned reanimation | 7 | 7 |
| sub-effect — unsupported delayed effect | 7 | 7 |
| sub-effect — unsupported type-change effect | 10 | 6 |
| sub-effect — unsupported counter spell | 7 | 6 |
| sub-effect — unsupported regenerate spell | 7 | 6 |
| structural — per-effect condition predicate not gateable | 11 | 5 |
| sub-effect — unsupported goad spell | 8 | 5 |
| sub-effect — unsupported connive effect | 7 | 5 |
| mode 2: sub-effect — unsupported ability content | 6 | 5 |
| sub-effect — unsupported can't-be-blocked effect | 6 | 5 |
| sub-effect — unsupported delayed trigger | 6 | 5 |
| structural — actual-result actor has no compatible owned action | 25 | 4 |
| sub-effect — unsupported scry spell | 6 | 3 |
| sub-effect — unsupported amass spell | 4 | 3 |
| sub-effect — unsupported emblem ability | 4 | 3 |
| sub-effect — unsupported incubate spell | 4 | 3 |
| sub-effect — unsupported explore spell | 3 | 3 |
| sub-effect — unsupported set base power/toughness effect | 3 | 3 |
| timing — conditional divided damage requires an announced allocation receipt retained across branch target selection | 3 | 3 |
| sub-effect — unsupported copy effect | 59 | 2 |
| structural — condition subject producer not publishable | 6 | 2 |
| sub-effect — unsupported transform effect | 4 | 2 |
| mode 2: sub-effect — unsupported temporary keyword spell | 3 | 2 |
| structural — actual-card quantity requires one owned card-result publisher | 3 | 2 |
| sub-effect — unsupported double counters spell | 3 | 2 |
| sub-effect — unsupported forced-attack effect | 3 | 2 |
| mode 1: sub-effect — unsupported discard spell | 2 | 2 |
| mode 2: sub-effect — unsupported power/toughness spell | 2 | 2 |
| mode 3: sub-effect — unsupported temporary keyword spell | 2 | 2 |
| structural — counter payment condition has no unique typed owner | 2 | 2 |
| structural — vote arm not lowered | 2 | 2 |
| sub-effect — unsupported can't-attack effect | 2 | 2 |
| sub-effect — unsupported color-change effect | 2 | 2 |
| sub-effect — unsupported lose-game effect | 2 | 2 |
| sub-effect — unsupported permanent choice | 2 | 2 |
| sub-effect — unsupported proliferate spell | 2 | 2 |
| sub-effect — unsupported surveil spell | 2 | 2 |
| sub-effect — unsupported win-game effect | 2 | 2 |
| sub-effect — unsupported life payment | 13 | 1 |
| sub-effect — unsupported delayed object capture: fixed-phase body has no unique compatible typed subject | 7 | 1 |
| structural — actual-result actor ownership not modeled | 4 | 1 |
| sub-effect — unsupported delayed object capture: the captured subject's immediate body primitive or parameter is not modeled: unsupported return spell | 4 | 1 |
| sub-effect — unsupported look-at-hand spell | 4 | 1 |
| mode 1: sub-effect — unsupported return spell | 2 | 1 |
| structural — condition target not remappable | 2 | 1 |
| mode 1: structural — counter-spell target | 1 | 1 |
| mode 1: sub-effect — unsupported draw spell | 1 | 1 |
| mode 1: sub-effect — unsupported life spell | 1 | 1 |
| mode 1: sub-effect — unsupported mana symbol | 1 | 1 |
| mode 1: sub-effect — unsupported temporary keyword spell | 1 | 1 |
| mode 2: structural — per-effect condition lowering failed | 1 | 1 |
| mode 2: sub-effect — unsupported counter placement | 1 | 1 |
| mode 2: sub-effect — unsupported explore spell | 1 | 1 |
| mode 2: sub-effect — unsupported keyword or ability loss | 1 | 1 |
| mode 2: sub-effect — unsupported library placement | 1 | 1 |
| mode 2: sub-effect — unsupported tap spell | 1 | 1 |
| mode 2: sub-effect — unsupported token creation | 1 | 1 |
| mode 3: sub-effect — unsupported damage spell | 1 | 1 |
| mode 3: sub-effect — unsupported exile spell | 1 | 1 |
| mode 3: sub-effect — unsupported manifest spell | 1 | 1 |
| mode 4: sub-effect — unsupported counter placement | 1 | 1 |
| mode 4: sub-effect — unsupported discard spell | 1 | 1 |
| mode 4: sub-effect — unsupported power/toughness spell | 1 | 1 |
| structural — counter outcome has no unique owned target occurrence | 1 | 1 |
| structural — counter-spell target | 1 | 1 |
| structural — departure characteristic consumer primitive not modeled | 1 | 1 |
| structural — instead negated gate not applicable | 1 | 1 |
| structural — mass reanimation carries unsupported riders | 1 | 1 |
| sub-effect — unsupported bolster spell | 1 | 1 |
| sub-effect — unsupported can-attack-as-though-defender effect | 1 | 1 |
| sub-effect — unsupported delayed object capture: fixed-phase subject is not one battlefield target | 1 | 1 |
| sub-effect — unsupported delayed object capture: the captured subject's immediate body primitive or parameter is not modeled: unsupported captured card action | 1 | 1 |
| sub-effect — unsupported delayed object capture: the captured subject's immediate body primitive or parameter is not modeled: unsupported counter placement | 1 | 1 |
| sub-effect — unsupported delayed object capture: the captured subject's immediate body primitive or parameter is not modeled: unsupported exile spell | 1 | 1 |
| sub-effect — unsupported delayed object capture: the captured subject's immediate body primitive or parameter is not modeled: unsupported sacrifice spell | 1 | 1 |
| sub-effect — unsupported discover spell | 1 | 1 |
| sub-effect — unsupported divided damage spell | 1 | 1 |
| sub-effect — unsupported double power/toughness spell | 1 | 1 |
| sub-effect — unsupported phase out effect | 1 | 1 |
| sub-effect — unsupported prevent-damage effect | 1 | 1 |
| sub-effect — unsupported reference subject | 130 | 0 |
| structural — optional published-subject antecedent not modeled | 60 | 0 |
| structural — clause reference not localizable | 58 | 0 |
| structural — optional action group actor not modeled | 49 | 0 |
| sub-effect — unsupported retarget effect | 37 | 0 |
| sub-effect — unsupported delayed object capture: fixed-phase subject has no validated domain, occurrence, or lifetime | 19 | 0 |
| structural — optional action group contains delay or payment | 5 | 0 |
| sub-effect — unsupported look-at-library spell | 5 | 0 |
| structural — search group clauses, guards, or references lack exact typed ownership | 4 | 0 |
| mode 2: sub-effect — unsupported exile spell | 2 | 0 |
| structural — clause produced modal/shared/multi-mode content | 2 | 0 |
| structural — optional action has no typed group identity | 2 | 0 |
| mode 1: sub-effect — unsupported counter placement | 1 | 0 |
| mode 1: sub-effect — unsupported destroy spell | 1 | 0 |
| mode 1: sub-effect — unsupported group power/toughness spell | 1 | 0 |
| mode 1: sub-effect — unsupported sacrifice spell | 1 | 0 |
| mode 1: sub-effect — unsupported token creation | 1 | 0 |
| mode 2: structural — single effect requires ordered lowering | 1 | 0 |
| mode 2: structural — unconsumed targets/references/keywords | 1 | 0 |
| mode 2: sub-effect — unsupported damage spell | 1 | 0 |
| mode 2: sub-effect — unsupported draw spell | 1 | 0 |
| mode 2: sub-effect — unsupported keyword or ability grant | 1 | 0 |
| mode 2: sub-effect — unsupported return spell | 1 | 0 |
| mode 2: sub-effect — unsupported transform effect | 1 | 0 |
| mode 3: sub-effect — unsupported ability content | 1 | 0 |
| search unit — only hand or battlefield destinations compose | 1 | 0 |
| structural — actual-result producer is negated or delayed | 1 | 0 |
| structural — card predicate has no exact available incarnation alternatives | 1 | 0 |
| structural — non-exact legacy effect pair | 1 | 0 |
| structural — optional postfix Otherwise has no unique singleton action/condition | 1 | 0 |
| structural — search unit has a missing or duplicate original ClauseID | 1 | 0 |
| structural — whole optional action outcome not modeled | 1 | 0 |
| sub-effect — unsupported adapt spell | 1 | 0 |
| sub-effect — unsupported delayed object capture: future delayed-body condition or its subject is not modeled | 1 | 0 |
| sub-effect — unsupported polymorph effect | 1 | 0 |
| sub-effect — unsupported remove from combat spell | 1 | 0 |
| sub-effect — unsupported tap or untap spell | 1 | 0 |

## Unrecognized per-effect conditions (recognition backlog)

Distinct `if <condition>` wordings inside ordered sequences whose predicate the compiler does not yet recognize. Recognizing a wording unblocks ordered-sequence lowering for the listed cards. Rows are ranked by sole blockers (cards a single wording is the only blocker for) then affected cards.

| Unrecognized condition | Affected cards | Sole blockers |
| --- | ---: | ---: |
| If you win | 17 | 15 |
| If you can't | 14 | 10 |
| if it has three or more +1/+1 counters on it | 5 | 5 |
| If it's a creature card | 4 | 4 |
| If the player can't | 4 | 4 |
| If you have a full party | 5 | 3 |
| If you win the flip | 4 | 3 |
| If a card with the chosen name was milled this way | 3 | 3 |
| If a player is dealt damage this way | 3 | 3 |
| If that land was nonbasic | 3 | 3 |
| If you controlled that permanent | 3 | 3 |
| If it's a land card | 5 | 2 |
| If excess damage was dealt this way | 3 | 2 |
| If this spell was foretold | 3 | 2 |
| If{B}was spent to cast this spell | 3 | 2 |
| if able | 3 | 2 |
| If it entered under your control | 2 | 2 |
| If mana from a Treasure was spent to activate this ability | 2 | 2 |
| If they can't | 2 | 2 |
| If this spell was cast from anywhere other than your hand | 2 | 2 |
| If you controlled it | 2 | 2 |
| If{U}was spent to cast this spell | 2 | 2 |
| if a nonland permanent left the battlefield this turn or a spell was warped this turn | 2 | 2 |
| if an opponent has more life than you | 2 | 2 |
| if it was attacking | 2 | 2 |
| if that creature has a +1/+1 counter on it | 2 | 2 |
| if you control an artifact and an enchantment | 2 | 2 |
| if you have more life than an opponent | 2 | 2 |
| If they don't | 5 | 1 |
| If damage is prevented this way | 3 | 1 |
| If it's a nonland card | 3 | 1 |
| If that creature is attacking | 3 | 1 |
| If that creature is legendary | 3 | 1 |
| If that player doesn't | 3 | 1 |
| If a permanent's ability is countered this way | 2 | 1 |
| If that card is a land card | 2 | 1 |
| If Falling Star doesn't turn completely over at least once during the flip | 1 | 1 |
| If Mishra | 1 | 1 |
| If a Dinosaur is dealt damage this way | 1 | 1 |
| If a Dragon is dealt damage this way | 1 | 1 |
| If a Rat is attacking | 1 | 1 |
| If a Werewolf is dealt damage this way | 1 | 1 |
| If a card named Stomping Slabs was revealed this way | 1 | 1 |
| If a creature is dealt damage this way | 1 | 1 |
| If a creature you controlled was destroyed this way | 1 | 1 |
| If a noncreature | 1 | 1 |
| If a nonland permanent left the battlefield this turn or a spell was warped this turn | 1 | 1 |
| If a nonred permanent is dealt damage this way | 1 | 1 |
| If a permanent was returned this way | 1 | 1 |
| If a permanent you controlled or a token was destroyed this way | 1 | 1 |
| If a white creature dies this way | 1 | 1 |
| If all cards revealed this way are creature cards | 1 | 1 |
| If an ability of an artifact | 1 | 1 |
| If an artifact card is revealed this way | 1 | 1 |
| If an instant card or a card with flash is exiled this way | 1 | 1 |
| If an opponent controls that creature | 1 | 1 |
| If an opponent has more cards in hand than you | 1 | 1 |
| If an outlaw was sacrificed this way | 1 | 1 |
| If another Desert was returned this way | 1 | 1 |
| If at least one Angel card is milled this way | 1 | 1 |
| If at least one creature card was exiled this way | 1 | 1 |
| If both coins come up heads | 1 | 1 |
| If cards with five or more different mana values are exiled with Azor's Gateway | 1 | 1 |
| If damage from a black source is prevented this way | 1 | 1 |
| If damage from a red source is prevented this way | 1 | 1 |
| If exactly three colors of mana were spent to activate this ability | 1 | 1 |
| If excess damage was dealt to a creature this way | 1 | 1 |
| If excess damage was dealt to a permanent this way | 1 | 1 |
| If excess damage was dealt to that creature this way | 1 | 1 |
| If fewer than two cards were discarded this way | 1 | 1 |
| If he's attacking | 1 | 1 |
| If it had mana value 3 or less | 1 | 1 |
| If it has a +1/+1 counter on it | 1 | 1 |
| If it has a counter on it | 1 | 1 |
| If it has the same name as a card in your graveyard | 1 | 1 |
| If it is | 1 | 1 |
| If it isn't a creature card | 1 | 1 |
| If it was a Jace planeswalker spell | 1 | 1 |
| If it was attacking | 1 | 1 |
| If it was dealt damage this turn | 1 | 1 |
| If it was tapped | 1 | 1 |
| If it wasn't attacking | 1 | 1 |
| If it's a creature card with power less than or equal to Grenzo's power | 1 | 1 |
| If it's an enchanted creature or enchantment creature | 1 | 1 |
| If it's an enchantment creature or legendary creature | 1 | 1 |
| If it's attacking a battle | 1 | 1 |
| If it's night | 1 | 1 |
| If it's not a Vehicle | 1 | 1 |
| If it's not your turn | 1 | 1 |
| If it's paired with a creature | 1 | 1 |
| If it's renowned | 1 | 1 |
| If it's the first combat phase of your turn | 1 | 1 |
| If its mana value is less than or equal to the number of lands you control | 1 | 1 |
| If mana from a Treasure was spent to cast this spell | 1 | 1 |
| If no counters were removed this way | 1 | 1 |
| If no creature cards were revealed this way | 1 | 1 |
| If no creature got votes | 1 | 1 |
| If seven or more creatures died this turn | 1 | 1 |
| If that Equipment was attached to a creature | 1 | 1 |
| If that artifact had counters on it | 1 | 1 |
| If that card is a Goblin card | 1 | 1 |
| If that card is a creature card of the chosen type | 1 | 1 |
| If that card's mana value is less than or equal to the number of experience counters you have | 1 | 1 |
| If that creature had a +1/+1 counter on it | 1 | 1 |
| If that creature had power 2 or less | 1 | 1 |
| If that creature has a +1/+1 counter on it | 1 | 1 |
| If that creature has two or more +1/+1 counters on it | 1 | 1 |
| If that creature is a Demon | 1 | 1 |
| If that creature is another Hero | 1 | 1 |
| If that creature is black or red | 1 | 1 |
| If that creature is white | 1 | 1 |
| If that creature is white or blue | 1 | 1 |
| If that creature was a God | 1 | 1 |
| If that creature was blue or black | 1 | 1 |
| If that creature was cast for its warp cost | 1 | 1 |
| If that creature was green or white | 1 | 1 |
| If that creature wasn't dealt damage this turn | 1 | 1 |
| If that land was legendary | 1 | 1 |
| If that permanent is black | 1 | 1 |
| If that permanent is green | 1 | 1 |
| If that permanent is red or green | 1 | 1 |
| If that permanent was blue or black | 1 | 1 |
| If that permanent's mana value was 3 or less | 1 | 1 |
| If that player can't | 1 | 1 |
| If that player is you | 1 | 1 |
| If that spell targets a commander you control | 1 | 1 |
| If that spell was all colors | 1 | 1 |
| If that spell's mana value is 10 or greater | 1 | 1 |
| If that spell's mana value is 5 or greater | 1 | 1 |
| If the card's mana value is 1 or less | 1 | 1 |
| If the creature deals damage to a creature this turn | 1 | 1 |
| If the creature you control entered this turn | 1 | 1 |
| If the discovered card's mana value is less than 10 | 1 | 1 |
| If the number is odd | 1 | 1 |
| If the result is 3 or less | 1 | 1 |
| If the result is equal to or less than the number of Robots you control | 1 | 1 |
| If the result is greater than the damage dealt or the result is 12 | 1 | 1 |
| If the revealed land card was a Mountain | 1 | 1 |
| If the sacrificed creature was suspected | 1 | 1 |
| If the spell is countered this way | 1 | 1 |
| If there is an instant card and a sorcery card in your graveyard | 1 | 1 |
| If they lost life this turn | 1 | 1 |
| If the{2}{U}cost was paid | 1 | 1 |
| If this permanent is an Aura | 1 | 1 |
| If this spell was cast from exile | 1 | 1 |
| If this spell was cast from your hand | 1 | 1 |
| If time gets more votes | 1 | 1 |
| If two cards that share a card type are discarded this way | 1 | 1 |
| If two cards that share all their card types were milled this way | 1 | 1 |
| If two or more players are tied for fewest | 1 | 1 |
| If you are one of those players | 1 | 1 |
| If you control a black permanent and a red permanent | 1 | 1 |
| If you control a blue permanent and a black permanent | 1 | 1 |
| If you control a creature with a counter on it | 1 | 1 |
| If you control a green permanent and a white permanent | 1 | 1 |
| If you control a modified creature | 1 | 1 |
| If you control a red permanent and a green permanent | 1 | 1 |
| If you control a white permanent and a blue permanent | 1 | 1 |
| If you control an artifact and an enchantment | 1 | 1 |
| If you control creatures named Mine Worker and Tower Worker | 1 | 1 |
| If you control creatures named Power Plant Worker and Tower Worker | 1 | 1 |
| If you control more creatures of that type than each other player | 1 | 1 |
| If you control more creatures than each other player | 1 | 1 |
| If you control neither creature | 1 | 1 |
| If you control that creature | 1 | 1 |
| If you controlled that artifact | 1 | 1 |
| If you controlled that creature | 1 | 1 |
| If you don't control a Faerie | 1 | 1 |
| If you don't control a Glimmer creature | 1 | 1 |
| If you don't control a Human | 1 | 1 |
| If you have less life than an opponent | 1 | 1 |
| If you lose the flip | 1 | 1 |
| If you paid one or more{E}this way | 1 | 1 |
| If you pay | 1 | 1 |
| If you rolled 6 exactly seven times | 1 | 1 |
| If you win all the flips | 1 | 1 |
| If you've cast another spell this turn | 1 | 1 |
| If you've cycled a card named Yidaro | 1 | 1 |
| If you've discarded a card this turn | 1 | 1 |
| If you've drawn two or more cards this turn | 1 | 1 |
| If you've scried or surveilled this turn | 1 | 1 |
| If{G}was spent to cast this spell | 1 | 1 |
| If{R}{G}was spent to cast this spell | 1 | 1 |
| If{W}was spent to cast this spell | 1 | 1 |
| if Jor Kadeen's power is 4 or greater | 1 | 1 |
| if a creature named Eight-and-a-Half-Tails is on the battlefield | 1 | 1 |
| if able without paying their mana costs | 1 | 1 |
| if an opponent has more cards in hand than you | 1 | 1 |
| if any of those cards remain exiled | 1 | 1 |
| if any player pays{2} | 1 | 1 |
| if at least one creature card was exiled this way | 1 | 1 |
| if it attacked or blocked since your last upkeep | 1 | 1 |
| if it doesn't have a counter of that kind on it | 1 | 1 |
| if it had a death counter on it | 1 | 1 |
| if it has an egg counter on it | 1 | 1 |
| if it has eight or more night counters on it | 1 | 1 |
| if it has five or more bloodstain counters on it | 1 | 1 |
| if it has three or more ritual counters on it | 1 | 1 |
| if it shares a color with that permanent | 1 | 1 |
| if it wasn't kicked | 1 | 1 |
| if it's attacking one of your opponents | 1 | 1 |
| if it's night | 1 | 1 |
| if its controller has three or more poison counters | 1 | 1 |
| if its mana value is 4 or less instead | 1 | 1 |
| if its mana value is 5 or less instead | 1 | 1 |
| if its mana value is less than or equal to the number of basic land types among lands you control | 1 | 1 |
| if its power is 2 | 1 | 1 |
| if its power is exactly 20 | 1 | 1 |
| if its power is less than Shelinda's power | 1 | 1 |
| if that creature has three or more +1/+1 counters on it | 1 | 1 |
| if that creature is a Mutant, Ninja, or Turtle | 1 | 1 |
| if that creature's power is greater than Yorvo's power | 1 | 1 |
| if that player has more cards in hand than you | 1 | 1 |
| if that target is white and/or blue | 1 | 1 |
| if the gift was promised and that creature isn't legendary | 1 | 1 |
| if the player hasn't played the card | 1 | 1 |
| if there are five or more hatchling counters on it | 1 | 1 |
| if there are three or more Lesson cards in your graveyard | 1 | 1 |
| if there are three or more basic land types among lands you control | 1 | 1 |
| if there are three or more bloodline counters on it | 1 | 1 |
| if there are three or more landmark counters on it | 1 | 1 |
| if there is a colorless creature card in your graveyard | 1 | 1 |
| if there is an Elf card in your graveyard | 1 | 1 |
| if they didn't attack you that turn | 1 | 1 |
| if this enchantment has three or more invitation counters on it | 1 | 1 |
| if this spell was cast from anywhere other than your hand | 1 | 1 |
| if you control a creature named Bogbrew Witch | 1 | 1 |
| if you control a creature with a counter on it | 1 | 1 |
| if you control a creature with the greatest power among creatures on the battlefield | 1 | 1 |
| if you control a modified creature | 1 | 1 |
| if you control a permanent with an oil counter on it | 1 | 1 |
| if you control an outlaw | 1 | 1 |
| if you control artifacts named Crown of Empires and Scepter of Empires | 1 | 1 |
| if you control artifacts named Crown of Empires and Throne of Empires | 1 | 1 |
| if you control artifacts named Scepter of Empires and Throne of Empires | 1 | 1 |
| if you control eight or more artifacts with the same name as one another | 1 | 1 |
| if you control seven or more lands and/or Treefolk | 1 | 1 |
| if you discarded the card with the greatest mana value among those cards or tied for greatest | 1 | 1 |
| if you don't control a Snail | 1 | 1 |
| if you don't control a legendary creature | 1 | 1 |
| if you had no cards in hand at the beginning of this turn | 1 | 1 |
| if you have fewer than three cards in hand | 1 | 1 |
| if you've cast another black spell this turn | 1 | 1 |
| if you've cast another instant or sorcery spell this turn | 1 | 1 |
| if your life total is greater than your starting life total | 1 | 1 |
| if your life total is less than or equal to half your starting life total | 1 | 1 |
| if{B}was spent to cast this spell | 1 | 1 |
| if{W}was spent to cast this spell | 1 | 1 |
| If a player does | 10 | 0 |
| If this spell was cast using teamwork | 7 | 0 |
| If this spell's additional cost was paid | 5 | 0 |
| If a Dragon was beheld | 3 | 0 |
| If a creature card is revealed this way | 3 | 0 |
| If a land card is revealed this way | 2 | 0 |
| If an opponent protects it | 2 | 0 |
| If evidence was collected | 2 | 0 |
| If it's your main phase | 2 | 0 |
| If this spell's madness cost was paid | 2 | 0 |
| if it's an instant or sorcery spell | 2 | 0 |
| If X is 1 | 1 | 0 |
| If a Hero was beheld | 1 | 0 |
| If a creature card with mana value 6 or greater is revealed this way | 1 | 0 |
| If a permanent spell is countered this way | 1 | 0 |
| If a planeswalker card is revealed this way | 1 | 0 |
| If a player does either | 1 | 0 |
| If an artifact or creature spell is countered this way | 1 | 0 |
| If an instant or sorcery card is revealed this way | 1 | 0 |
| If another creature died this turn | 1 | 0 |
| If any of those cards shares a card type with that spell | 1 | 0 |
| If at least one Zombie card is milled this way | 1 | 0 |
| If both targets are still legal as this ability resolves | 1 | 0 |
| If cards with at least six different mana values are revealed this way | 1 | 0 |
| If damage from a creature source is prevented this way | 1 | 0 |
| If each opponent chose silence | 1 | 0 |
| If excess damage would be dealt to that creature this way | 1 | 0 |
| If her sneak cost was paid this turn | 1 | 0 |
| If it comes up heads | 1 | 0 |
| If it entered from your library or was cast from your library | 1 | 0 |
| If it shares a card type with that permanent | 1 | 0 |
| If it was a Gideon planeswalker | 1 | 0 |
| If it was dealt noncombat damage this turn | 1 | 0 |
| If it wasn't an Aura | 1 | 0 |
| If it's a Saga | 1 | 0 |
| If it's a card of the chosen type | 1 | 0 |
| If it's a creature card that shares a creature type with a creature you control | 1 | 0 |
| If it's a land card or a creature card with mana value less than or equal to the number of loyalty counters on Nissa | 1 | 0 |
| If it's a land or double-faced card | 1 | 0 |
| If it's a permanent card | 1 | 0 |
| If it's an Equipment | 1 | 0 |
| If it's an artifact | 1 | 0 |
| If it's an instant or sorcery card | 1 | 0 |
| If it's the third time | 1 | 0 |
| If its mana cost contains{X} | 1 | 0 |
| If no one does | 1 | 0 |
| If that card is a Hero card | 1 | 0 |
| If that card is snow | 1 | 0 |
| If that card was all colors | 1 | 0 |
| If that creature has three or more +1/+0 counters on it | 1 | 0 |
| If that creature is a Kraken, Leviathan, Octopus, or Serpent | 1 | 0 |
| If that creature is equipped | 1 | 0 |
| If that creature is red | 1 | 0 |
| If that creature shares a color with the mana that land produced | 1 | 0 |
| If that creature was a token | 1 | 0 |
| If that enchantment is an Aura | 1 | 0 |
| If that library contains exactly the chosen number of cards with the chosen name | 1 | 0 |
| If that permanent is destroyed this way | 1 | 0 |
| If that permanent or card has a time counter on it | 1 | 0 |
| If that token is a Squirrel | 1 | 0 |
| If the copy is a permanent spell | 1 | 0 |
| If the creature had power 4 or greater | 1 | 0 |
| If the creature the opponent controls is dealt excess damage this way | 1 | 0 |
| If the creature would leave the battlefield | 1 | 0 |
| If the first player does | 1 | 0 |
| If the modified creature was sacrificed | 1 | 0 |
| If the total mana value of the cards exiled this way is 13 or less | 1 | 0 |
| If there are four or more chorus counters on Malcolm | 1 | 0 |
| If there are no omen counters on this enchantment | 1 | 0 |
| If they guessed right | 1 | 0 |
| If they guessed wrong | 1 | 0 |
| If they match | 1 | 0 |
| If the{2}{R}{R}cost was paid | 1 | 0 |
| If this creature is a Detective | 1 | 0 |
| If this creature is a Phyrexian | 1 | 0 |
| If this creature is an Angel | 1 | 0 |
| If this creature's emerge cost was paid | 1 | 0 |
| If this creature's sneak cost was paid | 1 | 0 |
| If this spell was kicked with its{2}{R}kicker | 1 | 0 |
| If this spell's additional cost was paid and the revealed card is the chosen type | 1 | 0 |
| If this spell's freerunning cost was paid | 1 | 0 |
| If this spell's mayhem cost was paid | 1 | 0 |
| If this spell's prowl cost was paid | 1 | 0 |
| If this spell's surge cost was paid | 1 | 0 |
| If those Auras would leave the battlefield | 1 | 0 |
| If those permanents are exchanged this way | 1 | 0 |
| If two cards that share a card type were milled this way | 1 | 0 |
| If you control all three | 1 | 0 |
| If you control an outlaw | 1 | 0 |
| If you do and your opponent guessed wrong | 1 | 0 |
| If you do or if you control another Dinosaur | 1 | 0 |
| If you have an enduring story | 1 | 0 |
| If you have one or fewer cards in hand | 1 | 0 |
| If you lost life this way | 1 | 0 |
| If you won five flips this way | 1 | 0 |
| If you've committed a crime this turn | 1 | 0 |
| If you've drawn three or more cards this turn | 1 | 0 |
| If{S}was spent to cast this spell | 1 | 0 |
| if an opponent has 0 or less life | 1 | 0 |
| if it attacked or blocked this turn | 1 | 0 |
| if it exploited that creature | 1 | 0 |
| if it has an odd number of counters on it | 1 | 0 |
| if it has five or more hunger counters on it | 1 | 0 |
| if it has five or more point counters on it | 1 | 0 |
| if it has seven or more phyresis counters on it | 1 | 0 |
| if it has ten or more point counters on it | 1 | 0 |
| if it has the same name as a permanent | 1 | 0 |
| if it has three or more doom counters on it | 1 | 0 |
| if it's a creature card | 1 | 0 |
| if it's a creature spell | 1 | 0 |
| if it's a spell with lesser mana value | 1 | 0 |
| if it's a spell with mana value less than or equal to Ryan's power | 1 | 0 |
| if it's an instant or sorcery card | 1 | 0 |
| if it's an instant spell with mana value 2 or less | 1 | 0 |
| if its mana value is odd | 1 | 0 |
| if that player has more cards in hand than each other player | 1 | 0 |
| if that spell's mana value is 8 or less | 1 | 0 |
| if that target is a creature or planeswalker | 1 | 0 |
| if the exiled card's mana value is 4 or greater | 1 | 0 |
| if the exiled creature was a Thrull | 1 | 0 |
| if the spell's mana value is less than the number of Mountains you control | 1 | 0 |
| if there are eight or more different names among unlocked doors of Rooms you control | 1 | 0 |
| if there are five basic land types among lands you control | 1 | 0 |
| if there are no echo counters on it | 1 | 0 |
| if there are three or more bore counters on it | 1 | 0 |
| if there are three or more cards exiled with Profane Procession | 1 | 0 |
| if there are three or more dread counters on it | 1 | 0 |
| if there are three or more judgment counters on it | 1 | 0 |
| if there are three or more soul counters on it | 1 | 0 |
| if they own three or more exiled cards with hit counters on them | 1 | 0 |
| if this creature has five or more filibuster counters on it | 1 | 0 |
| if this creature has seven or more ember counters on it | 1 | 0 |
| if this spell was cast using teamwork | 1 | 0 |
| if this spell's additional cost was paid | 1 | 0 |
| if you control a blue permanent other than Ral | 1 | 0 |
| if you control a red permanent other than Ajani | 1 | 0 |
| if you control exactly thirteen permanents | 1 | 0 |
| if you control that creature | 1 | 0 |
| if you have seven or more cards in your graveyard | 1 | 0 |
| if you revealed it this way | 1 | 0 |

## Modeled-capability envelope gaps (parameter backlog)

Families the compiler recognizes but lowers only within an exact envelope. Each row is one supported-envelope blocker ranked by sole blockers (cards it is the only blocker for); growing the envelope to cover the example wordings unblocks the listed cards. This is the effect-family analogue of the unrecognized-condition backlog above.

| Capability | Supported envelope (blocker) | Affected cards | Sole blockers | Example wordings |
| --- | --- | ---: | ---: | --- |
| unsupported enters-tapped replacement | the executable source backend supports only exact unconditional self enters-tapped replacements | 418 | 200 | If a source you control would deal noncombat damage to a creature an opponent controls, pu…; If a creature you control would explore, instead it explores, then it explores again.; If this land would enter, sacrifice two untapped lands instead. If you do, put this land o… |
| unsupported phase/step trigger phrase | the executable source backend does not support this intervening-if condition | 274 | 192 | At the beginning of your upkeep, if you control no Thopters other than this creature, retu…; At the beginning of each end step, if an opponent discarded a card this turn, you draw a c…; At the beginning of combat on your turn, if two or more players have lost the game, gain c… |
| unsupported return spell | the executable source backend supports only exact return of one target permanent to its owner's hand | 291 | 188 | return this card from your graveyard to the battlefield attached to that creature.; • Return target nonland permanent to its owner's hand.; Return enchanted creature and this Aura to their owners' hands. |
| unsupported counter placement | the executable source backend supports exact recognized counter placement on one valid target | 469 | 187 | put the cards exiled with it into their owner's hand.; If this creature is a Spirit, put five +1/+1 counters on it.; remove an intervention counter from this enchantment. |
| unsupported token creation | the executable source backend supports only a single fixed-power/toughness creature token with one subtype and at most one color | 293 | 166 | Create a Wicked Role token attached to up to one target creature you control. (If you cont…; • Create a 1/1 colorless Eldrazi Scion creature token with "Sacrifice this token: Add {C…; • Create a 3/3 black Dalek artifact creature token with menace. |
| unsupported destroy spell | the executable source backend supports only exact destruction of one target permanent | 226 | 160 | Destroy all non-Wall creatures blocking enchanted creature.; Destroy target creature with power less than this creature's power.; Destroy each creature whose coin comes up tails. |
| unsupported damage spell | the executable source backend supports only exact supported damage amounts to one target | 199 | 127 | Choose target creature. Whenever that creature is dealt damage this turn, it deals that mu…; Armed Response deals damage to target attacking creature equal to the number of Equipment…; Close Encounter deals damage equal to the power of the chosen creature or card to target c… |
| unsupported temporary keyword spell | the executable source backend supports only exact non-parameterized keyword grants to one target creature or permanent until end of turn | 133 | 104 | unless you pay {R}, whenever this creature blocks or becomes blocked by a creature this co…; If you do, it gains flying until end of turn.; Until end of turn, target creature you control gains indestructible and "Whenever this cre… |
| unsupported permanent zone-change trigger | the executable source backend does not support this semantic permanent zone-change trigger condition | 153 | 100 | Whenever one or more creatures you control enter, if one or more of them entered from a gr…; When Uldaros Theorix enters, if you cast him, exile up to one target nonland card of each…; When this creature enters, if it was kicked twice, it deals 2 damage to any target. |
| unsupported exile spell | the executable source backend supports only exact exile of one target permanent | 198 | 91 | Exile all multicolored permanents.; exile them from your graveyard.; Target opponent exiles a creature they control and their graveyard. |
| unsupported power/toughness spell | the executable source backend supports only exact supported target-creature power/toughness changes until end of turn | 162 | 91 | • Target creature gets +3/-3 until end of turn.; you get {TK}.; Until your next turn, up to one target creature gets -3/-0. |
| unsupported search effect | the executable source backend supports only exact unconditional library-search sequences | 137 | 75 | you may search your library for a Food card, reveal it, put it into your hand, then shuffl…; each player searches their library for up to two basic land cards, puts them onto the batt…; you may search your library for up to two artifact, creature, and/or enchantment cards wit… |
| unsupported draw spell | the executable source backend supports only exact fixed card draw | 99 | 63 | Draw a card for each tapped creature target opponent controls.; Target player skips their next draw step.; that creature's controller draws a card. |
| unsupported life spell | the executable source backend supports only exact fixed life changes | 82 | 59 | Each opponent with no cards in hand loses 10 life.; r leaves the battlefield, you lose 3 life if; you and that player each gain that much life. |
| unsupported damage spell | the executable source backend supports only exact fixed or X group damage amounts | 88 | 53 | The next time a source of your choice of the chosen color would deal damage to you this tu…; creature spell. Whenever you cast a noncreature spell, if at least four mana was spent to…; this creature deals X damage to that player, where X is the number of cards in their hand… |
| unsupported gain-control spell | the executable source backend supports only exact gain-control sequences targeting one permanent | 81 | 52 | Gain control of target creature until end of turn. Untap that creature. It gains haste unt…; Gain control of target creature. Untap that creature. It gains haste until end of turn. Sa…; Gain control of target creature until end of turn. Untap it. It gains haste until end of t… |
| unsupported counter spell | the executable source backend supports only exact counter of one target spell | 64 | 47 | counter all abilities your opponents control.; Counter target spell with mana value X. (For example, if that spell's mana cost is {3}{U}{…; Counter target activated or triggered ability. If a permanent's ability is countered this… |
| unsupported permanent zone-change trigger effect | the executable source backend does not support this permanent zone-change trigger body | 204 | 46 | When this creature enters, empower Jace 3. (Put three loyalty counters on a Jace token you…; When this enchantment enters, suspect target creature an opponent controls. As long as thi…; Whenever a nontoken creature you control enters, you may pay {2}. If you do, create a toke… |
| unsupported triggered ability effect | the executable source backend supports only recognized semantic self triggers with supported effects | 79 | 43 | When this creature attacks, up to one target creature defending player controls blocks it…; Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it a…; Whenever this creature attacks, you may pay {E}{E}. If you do, it gains first strike until… |
| unsupported damage spell | the executable source backend supports only exact fixed group damage amounts | 54 | 43 | Flame Sweep deals 2 damage to each creature except for creatures you control with flying.; This creature deals X damage to each creature and each player. Spend only black mana on X.; it deals 7 damage to a creature an opponent controls chosen at random. |
| unsupported triggered ability | the executable source backend supports only recognized semantic self triggers with supported effects | 44 | 37 | Whenever Whiplash attacks, if he's equipped, each opponent loses X life and you gain X lif…; Pack tactics — Whenever this creature attacks, if you attacked with creatures with total…; Whenever this creature attacks for the first time each turn, if it's attacking the player… |
| unsupported enters-with-counters replacement | the executable source backend supports only exact self enters-with-counters replacements | 69 | 35 | If this creature was kicked with its {1}{U} kicker, it enters with two +1/+1 counters on i…; If this creature was kicked with its {B} kicker, it enters with a +1/+1 counter on it and…; This creature enters with two +1/+1 counters on it if it wasn't cast or no mana was spent… |
| unsupported attach effect | the executable source backend supports only "attach it/that/this &lt;Equipment&gt; to target &lt;permanent&gt; you control" attaching the entering or source permanent | 51 | 33 | Attach target Aura attached to a creature to another creature.; attach this Equipment to that creature.; attach this Aura to that creature. |
| unsupported Enchant ability | the executable source backend supports only exact Enchant with a supported target kind | 50 | 30 | Gain control of target Aura that's attached to a permanent. Attach it to another permanent…; At the beginning of combat on your turn, create a green Aura enchantment token named Settl…; Enchant nonland permanent |
| unsupported life spell | the executable source backend supports only exact supported life changes | 41 | 30 | defending player loses 1 life for each card in their graveyard.; Each opponent loses life equal to the number of cards in their graveyard.; that player gains X life, where X is the number of cards in all graveyards with the same n… |
| unsupported phase/step trigger phrase effect | the executable source backend does not support this phase/step trigger body | 75 | 29 | At the beginning of your upkeep, you may flip a coin. If you win the flip, put a fuse coun…; your upkeep, choose left or right. Each player may at; At the beginning of combat on your turn, you may pay {E}. If you do, another target creatu… |
| unsupported gain-control spell | the executable source backend supports only exact gain-control of one target permanent | 52 | 29 | Gain control of target creature. Change the text of that creature by replacing all instanc…; Gain control of target creature with power less than or equal to the number of treasure co…; Gain control of up to two target creatures with total mana value 6 or less for as long as… |
| unsupported tap spell | the executable source backend supports only exact tap of one target permanent | 37 | 29 | tap all creatures defending player controls.; tap target creature an opponent controls.; Tap all creatures target player controls. |
| unsupported shuffle effect | the executable source backend supports only a source-spell shuffle into its owner's library, a controller graveyard shuffle into library, or a target player shuffling their graveyard into their library | 41 | 28 | Choose target artifact or enchantment. Its owner shuffles it into their library.; Target player shuffles up to four target cards from their graveyard into their library.; Shuffle your library. |
| unsupported library placement | the executable source backend supports only exact target graveyard-to-library placement | 53 | 24 | Put up to three target cards from an opponent's graveyard on top of their library in any o…; put it on the bottom of its owner's library.; put target card from an opponent's graveyard on the bottom of their library. |
| unsupported triggered ability | the executable source backend does not support this semantic trigger condition | 37 | 24 | Whenever you activate an ability that isn't a mana ability, if life was paid to activate i…; Whenever a player attacks you, if that player has another opponent who isn't being attacke…; Whenever you attack with two or more creatures, this creature becomes prepared. (While it'… |
| unsupported power/toughness spell | the executable source backend supports only exact until-end-of-turn power/toughness changes to the source or referenced permanent | 34 | 23 | This creature gets +1/+1 until end of turn. Any player may activate this ability.; This creature gets -1/-1 until end of turn. Any player may activate this ability.; Until end of turn, this creature gets +1/+1 for each permanent card in your graveyard. |
| unsupported damage spell | the executable source backend does not support this group recipient | 32 | 23 | Davriel deals 2 damage to them.; this creature deals 3 damage to that spell or ability's controller.; it deals that much damage to each other opponent. |
| unsupported triggered ability | the executable source backend does not support this semantic spell-cast trigger condition | 38 | 22 | Whenever a player casts a spell, if it's not their turn, that player draws a card.; Whenever you cast a spell, if that spell was kicked, put a +1/+1 counter on Hallar, then H…; Whenever a player casts a spell, if no colored mana was spent to cast it, counter that spe… |
| unsupported untap spell | the executable source backend supports only exact untap of one target permanent | 39 | 20 | untap each enchanted permanent you control.; Untap up to one target creature and up to one target land.; untap target land you control. |
| unsupported card layout | the source generator does not support Scryfall layout "flip" | 20 | 20 | - |
| unsupported damage spell | the executable source backend supports only exact fixed, X, or source-power damage to that player | 27 | 18 | this Aura deals 2 damage to that player unless that creature attacked this turn.; this enchantment deals damage to that player equal to the number of artifacts they control…; This enchantment deals damage equal to that card's mana value to that player. |
| unsupported group power/toughness spell | the executable source backend supports exact fixed group power/toughness changes and linked all-creatures -X/-X until end of turn | 20 | 18 | All creatures get -X/-X until end of turn, where X is the greatest mana value of a command…; All creatures get +2/+2 for as long as this artifact remains tapped.; Attacking creatures get +X/+0 until end of turn, where X is this creature's power. |
| unsupported discard spell | the executable source backend supports only exact fixed discard by one player | 25 | 17 | • Each opponent discards a card.; • Target opponent discards a card.; Target player discards a number of cards equal to the sacrificed creature's power. |
| unsupported mana effect | the executable source backend supports only exact unconditional add-mana content | 27 | 16 | that player adds one mana of any type that land produced.; If you do, this creature assigns no combat damage this turn and at the beginning of your n…; add one mana of any type that artifact token produced. |
