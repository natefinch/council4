# Card-Support Planning Report

Capability-aware blockers for eligible paper cards that cannot yet be generated. Each distinct diagnostic summary and capability is counted at most once per card.

## Diagnostic reasons

A sole blocker is the card's only distinct diagnostic summary. The most common co-blocker excludes the reason in its own row.

| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | unsupported Oracle construct | 6,341 | 0 | 0.0% | unsupported static ability |
| 2 | unsupported static ability | 5,854 | 1,175 | 20.1% | unsupported Oracle construct |
| 3 | unsupported triggered ability | 4,691 | 2,258 | 48.1% | unsupported Oracle construct |
| 4 | unsupported ordered effect sequence | 4,444 | 3,004 | 67.6% | unsupported ability content |
| 5 | unsupported ability content | 3,714 | 902 | 24.3% | unsupported Oracle construct |
| 6 | unsupported enter trigger | 2,744 | 1,269 | 46.2% | unsupported Oracle construct |
| 7 | unsupported activated ability | 1,601 | 718 | 44.8% | unsupported Oracle construct |
| 8 | unsupported phase/step trigger phrase | 1,243 | 568 | 45.7% | unsupported Oracle construct |
| 9 | unsupported enters-tapped replacement | 1,192 | 250 | 21.0% | unsupported Oracle construct |
| 10 | unsupported mixed keyword ability | 1,065 | 457 | 42.9% | unsupported Oracle construct |
| 11 | unsupported damage spell | 1,040 | 660 | 63.5% | unsupported Oracle construct |
| 12 | unsupported power/toughness spell | 980 | 548 | 55.9% | unsupported Oracle construct |
| 13 | unsupported counter placement | 840 | 365 | 43.5% | unsupported Oracle construct |
| 14 | unsupported ability word | 755 | 194 | 25.7% | unsupported Oracle construct |
| 15 | unsupported destroy spell | 571 | 381 | 66.7% | unsupported Oracle construct |
| 16 | unsupported return spell | 571 | 334 | 58.5% | unsupported Oracle construct |
| 17 | unsupported exile spell | 496 | 202 | 40.7% | unsupported ordered effect sequence |
| 18 | unsupported temporary keyword spell | 493 | 275 | 55.8% | unsupported Oracle construct |
| 19 | unsupported search effect | 396 | 244 | 61.6% | unsupported ability content |
| 20 | unsupported mana ability | 376 | 190 | 50.5% | unsupported static ability |
| 21 | unsupported dies trigger | 356 | 144 | 40.4% | unsupported Oracle construct |
| 22 | unsupported life spell | 337 | 204 | 60.5% | unsupported Oracle construct |
| 23 | unsupported modal ability | 333 | 256 | 76.9% | unsupported Oracle construct |
| 24 | unsupported Enchant ability | 212 | 24 | 11.3% | unsupported static ability |
| 25 | unsupported enters-with-counters replacement | 193 | 35 | 18.1% | unsupported Oracle construct |
| 26 | unsupported unknown ability | 186 | 0 | 0.0% | unsupported Oracle construct |
| 27 | unsupported regenerate spell | 179 | 98 | 54.7% | unsupported static ability |
| 28 | unsupported untap spell | 175 | 81 | 46.3% | unsupported static ability |
| 29 | unsupported draw spell | 165 | 84 | 50.9% | unsupported Oracle construct |
| 30 | unsupported mana symbol | 158 | 76 | 48.1% | unsupported enters-tapped replacement |
| 31 | unsupported gain-control spell | 148 | 73 | 49.3% | unsupported static ability |
| 32 | unsupported keyword ability | 134 | 38 | 28.4% | unsupported triggered ability |
| 33 | unsupported tap spell | 123 | 80 | 65.0% | unsupported Oracle construct |
| 34 | unsupported discard spell | 122 | 69 | 56.6% | unsupported Oracle construct |
| 35 | unsupported sacrifice spell | 115 | 62 | 53.9% | unsupported ability content |
| 36 | unsupported counter spell | 107 | 81 | 75.7% | unsupported ability content |
| 37 | unsupported multiple spell abilities | 98 | 91 | 92.9% | unsupported ability content |
| 38 | unsupported cost | 88 | 0 | 0.0% | unsupported activated ability |
| 39 | unsupported loyalty ability | 82 | 0 | 0.0% | unsupported ordered effect sequence |
| 40 | unsupported dies trigger body | 77 | 33 | 42.9% | unsupported mixed keyword ability |
| 41 | unsupported mill spell | 74 | 43 | 58.1% | unsupported ability content |
| 42 | unsupported parameterized keyword | 68 | 13 | 19.1% | unsupported triggered ability |
| 43 | unsupported Equip ability | 63 | 15 | 23.8% | unsupported static ability |
| 44 | unsupported type line | 61 | 59 | 96.7% | unsupported Oracle construct |
| 45 | unsupported dies trigger phrase | 46 | 23 | 50.0% | unsupported Oracle construct |
| 46 | unsupported damage replacement | 35 | 20 | 57.1% | unsupported static ability |
| 47 | unsupported reminder ability | 34 | 0 | 0.0% | unsupported Oracle construct |
| 48 | unsupported group power/toughness spell | 33 | 21 | 63.6% | unsupported Oracle construct |
| 49 | unsupported manifest spell | 31 | 21 | 67.7% | unsupported activated ability |
| 50 | unsupported fight spell | 26 | 14 | 53.8% | unsupported ordered effect sequence |
| 51 | unsupported conditional enters-tapped replacement | 25 | 2 | 8.0% | unsupported ability content |
| 52 | incomplete executable lowering | 24 | 17 | 70.8% | unsupported Oracle construct |
| 53 | unsupported card layout | 20 | 20 | 100.0% | - |
| 54 | unsupported draw/discard trigger | 19 | 14 | 73.7% | unsupported Oracle construct |
| 55 | unsupported counter-placement replacement | 19 | 6 | 31.6% | unsupported Oracle construct |
| 56 | unsupported delayed effect | 13 | 8 | 61.5% | unsupported Oracle construct |
| 57 | unsupported Protection ability | 10 | 3 | 30.0% | unsupported Oracle construct |
| 58 | unsupported cycling trigger | 9 | 9 | 100.0% | - |
| 59 | unsupported explore spell | 9 | 5 | 55.6% | unsupported Oracle construct |
| 60 | unsupported Read ahead ability | 9 | 0 | 0.0% | unsupported ordered effect sequence |
| 61 | validation failed: oracle-without-abilities | 7 | 7 | 100.0% | - |
| 62 | unsupported static rule declaration | 7 | 0 | 0.0% | unsupported ability content |
| 63 | unsupported package letter | 6 | 6 | 100.0% | - |
| 64 | unsupported investigate spell | 6 | 1 | 16.7% | unsupported triggered ability |
| 65 | unsupported scry spell | 5 | 3 | 60.0% | unsupported Oracle construct |
| 66 | unsupported proliferate spell | 5 | 1 | 20.0% | unsupported Oracle construct |
| 67 | unsupported token-creation replacement | 4 | 2 | 50.0% | unsupported ability content |
| 68 | unsupported Mutate ability | 3 | 3 | 100.0% | - |
| 69 | unsupported self zone-destination replacement | 3 | 3 | 100.0% | - |
| 70 | unsupported Ninjutsu ability | 3 | 0 | 0.0% | unsupported triggered ability |
| 71 | unsupported surveil spell | 2 | 1 | 50.0% | unsupported Oracle construct |
| 72 | unsupported Cycling ability | 1 | 1 | 100.0% | - |
| 73 | unsupported hand Cycling grant | 1 | 0 | 0.0% | unsupported counter placement |

## Capability clusters

A fully unlockable card has every distinct diagnostic summary in one capability cluster. Constituent summaries list the diagnostics currently observed in that cluster.

| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |
| --- | ---: | ---: | --- |
| shared-ability-content | 13,612 | 9,278 | unsupported ability content; unsupported counter placement; unsupported counter spell; unsupported damage spell; unsupported delayed effect; unsupported destroy spell; unsupported dies trigger body; unsupported discard spell; unsupported draw spell; unsupported exile spell; unsupported explore spell; unsupported fight spell; unsupported gain-control spell; unsupported group power/toughness spell; unsupported investigate spell; unsupported life spell; unsupported manifest spell; unsupported mill spell; unsupported modal ability; unsupported multiple spell abilities; unsupported ordered effect sequence; unsupported power/toughness spell; unsupported proliferate spell; unsupported regenerate spell; unsupported return spell; unsupported scry spell; unsupported search effect; unsupported tap spell; unsupported temporary keyword spell; unsupported untap spell |
| trigger-pattern | 8,554 | 4,581 | unsupported cycling trigger; unsupported dies trigger; unsupported dies trigger phrase; unsupported draw/discard trigger; unsupported enter trigger; unsupported phase/step trigger phrase; unsupported triggered ability |
| static-declaration | 7,012 | 1,791 | unsupported Enchant ability; unsupported Protection ability; unsupported Read ahead ability; unsupported hand Cycling grant; unsupported keyword ability; unsupported mixed keyword ability; unsupported parameterized keyword; unsupported static ability; unsupported static rule declaration |
| activation | 2,252 | 1,059 | unsupported Cycling ability; unsupported Equip ability; unsupported Mutate ability; unsupported Ninjutsu ability; unsupported activated ability; unsupported cost; unsupported loyalty ability; unsupported mana ability; unsupported mana symbol |
| replacement | 1,458 | 323 | unsupported conditional enters-tapped replacement; unsupported counter-placement replacement; unsupported damage replacement; unsupported enters-tapped replacement; unsupported enters-with-counters replacement; unsupported self zone-destination replacement; unsupported token-creation replacement |
| recognition-fallback | 6,739 | 268 | unsupported Oracle construct; unsupported ability word; unsupported reminder ability; unsupported unknown ability |
| other | 235 | 172 | incomplete executable lowering; unsupported card layout; unsupported package letter; unsupported sacrifice spell; unsupported surveil spell; unsupported type line; validation failed: oracle-without-abilities |
