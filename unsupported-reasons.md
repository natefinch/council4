# Card-Support Planning Report

Capability-aware blockers for eligible paper cards that cannot yet be generated. Each distinct diagnostic summary and capability is counted at most once per card.

## Diagnostic reasons

A sole blocker is the card's only distinct diagnostic summary. The most common co-blocker excludes the reason in its own row.

| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | unsupported Oracle construct | 6,341 | 0 | 0.0% | unsupported static ability |
| 2 | unsupported static ability | 5,830 | 1,184 | 20.3% | unsupported Oracle construct |
| 3 | unsupported ordered effect sequence | 4,649 | 3,120 | 67.1% | unsupported ability content |
| 4 | unsupported triggered ability | 4,624 | 2,143 | 46.3% | unsupported Oracle construct |
| 5 | unsupported ability content | 3,905 | 996 | 25.5% | unsupported Oracle construct |
| 6 | unsupported activated ability | 1,592 | 712 | 44.7% | unsupported Oracle construct |
| 7 | unsupported enters-tapped replacement | 1,233 | 282 | 22.9% | unsupported Oracle construct |
| 8 | unsupported damage spell | 1,160 | 734 | 63.3% | unsupported Oracle construct |
| 9 | unsupported power/toughness spell | 1,103 | 631 | 57.2% | unsupported Oracle construct |
| 10 | unsupported mixed keyword ability | 1,047 | 445 | 42.5% | unsupported Oracle construct |
| 11 | unsupported counter placement | 1,020 | 437 | 42.8% | unsupported Oracle construct |
| 12 | unsupported enter trigger effect | 898 | 474 | 52.8% | unsupported Oracle construct |
| 13 | unsupported ability word | 755 | 194 | 25.7% | unsupported Oracle construct |
| 14 | unsupported phase/step trigger phrase | 629 | 298 | 47.4% | unsupported Oracle construct |
| 15 | unsupported triggered ability effect | 612 | 333 | 54.4% | unsupported Oracle construct |
| 16 | unsupported phase/step trigger phrase effect | 603 | 259 | 43.0% | unsupported Oracle construct |
| 17 | unsupported return spell | 601 | 350 | 58.2% | unsupported Oracle construct |
| 18 | unsupported destroy spell | 586 | 392 | 66.9% | unsupported Oracle construct |
| 19 | unsupported exile spell | 534 | 206 | 38.6% | unsupported ordered effect sequence |
| 20 | unsupported temporary keyword spell | 515 | 294 | 57.1% | unsupported Oracle construct |
| 21 | unsupported life spell | 411 | 252 | 61.3% | unsupported Oracle construct |
| 22 | unsupported search effect | 407 | 247 | 60.7% | unsupported ability content |
| 23 | unsupported mana ability | 359 | 174 | 48.5% | unsupported static ability |
| 24 | unsupported modal ability | 333 | 256 | 76.9% | unsupported Oracle construct |
| 25 | unsupported untap spell | 212 | 92 | 43.4% | unsupported static ability |
| 26 | unsupported Enchant ability | 212 | 27 | 12.7% | unsupported static ability |
| 27 | unsupported enter trigger | 198 | 112 | 56.6% | unsupported Oracle construct |
| 28 | unsupported enters-with-counters replacement | 192 | 35 | 18.2% | unsupported Oracle construct |
| 29 | unsupported draw spell | 190 | 94 | 49.5% | unsupported Oracle construct |
| 30 | unsupported unknown ability | 186 | 0 | 0.0% | unsupported Oracle construct |
| 31 | unsupported regenerate spell | 181 | 104 | 57.5% | unsupported static ability |
| 32 | unsupported tap spell | 171 | 85 | 49.7% | unsupported static ability |
| 33 | unsupported mana symbol | 158 | 77 | 48.7% | unsupported enters-tapped replacement |
| 34 | unsupported gain-control spell | 154 | 78 | 50.6% | unsupported static ability |
| 35 | unsupported discard spell | 134 | 74 | 55.2% | unsupported Oracle construct |
| 36 | unsupported keyword ability | 134 | 38 | 28.4% | unsupported triggered ability |
| 37 | unsupported sacrifice spell | 132 | 71 | 53.8% | unsupported Oracle construct |
| 38 | unsupported counter spell | 108 | 82 | 75.9% | unsupported ability content |
| 39 | unsupported multiple spell abilities | 98 | 91 | 92.9% | unsupported ability content |
| 40 | unsupported dies trigger effect | 88 | 43 | 48.9% | unsupported Oracle construct |
| 41 | unsupported cost | 88 | 0 | 0.0% | unsupported activated ability |
| 42 | unsupported mill spell | 82 | 50 | 61.0% | unsupported ability content |
| 43 | unsupported loyalty ability | 82 | 0 | 0.0% | unsupported ordered effect sequence |
| 44 | unsupported parameterized keyword | 68 | 13 | 19.1% | unsupported triggered ability |
| 45 | unsupported Equip ability | 63 | 15 | 23.8% | unsupported static ability |
| 46 | unsupported type line | 61 | 59 | 96.7% | unsupported Oracle construct |
| 47 | unsupported dies trigger body | 54 | 27 | 50.0% | unsupported Oracle construct |
| 48 | unsupported group power/toughness spell | 37 | 25 | 67.6% | unsupported Oracle construct |
| 49 | unsupported reminder ability | 34 | 0 | 0.0% | unsupported Oracle construct |
| 50 | unsupported dies trigger | 32 | 12 | 37.5% | unsupported Oracle construct |
| 51 | unsupported manifest spell | 31 | 21 | 67.7% | unsupported activated ability |
| 52 | unsupported fight spell | 30 | 14 | 46.7% | unsupported ordered effect sequence |
| 53 | unsupported conditional enters-tapped replacement | 30 | 2 | 6.7% | unsupported ability content |
| 54 | incomplete executable lowering | 25 | 18 | 72.0% | unsupported Oracle construct |
| 55 | unsupported draw/discard trigger effect | 23 | 21 | 91.3% | unsupported Oracle construct |
| 56 | unsupported card layout | 20 | 20 | 100.0% | - |
| 57 | unsupported delayed effect | 15 | 9 | 60.0% | unsupported Oracle construct |
| 58 | unsupported dies trigger phrase | 15 | 7 | 46.7% | unsupported static ability |
| 59 | unsupported Protection ability | 10 | 3 | 30.0% | unsupported Oracle construct |
| 60 | unsupported explore spell | 9 | 5 | 55.6% | unsupported Oracle construct |
| 61 | unsupported Read ahead ability | 9 | 0 | 0.0% | unsupported ordered effect sequence |
| 62 | validation failed: oracle-without-abilities | 7 | 7 | 100.0% | - |
| 63 | unsupported investigate spell | 7 | 1 | 14.3% | unsupported triggered ability |
| 64 | unsupported static rule declaration | 7 | 0 | 0.0% | unsupported ability content |
| 65 | unsupported package letter | 6 | 6 | 100.0% | - |
| 66 | unsupported scry spell | 6 | 4 | 66.7% | unsupported Oracle construct |
| 67 | unsupported draw/discard trigger | 5 | 2 | 40.0% | unsupported ordered effect sequence |
| 68 | unsupported proliferate spell | 5 | 1 | 20.0% | unsupported Oracle construct |
| 69 | unsupported counter-placement replacement | 4 | 1 | 25.0% | unsupported life spell |
| 70 | unsupported Mutate ability | 3 | 3 | 100.0% | - |
| 71 | unsupported self zone-destination replacement | 3 | 3 | 100.0% | - |
| 72 | unsupported Ninjutsu ability | 3 | 0 | 0.0% | unsupported triggered ability |
| 73 | unsupported damage replacement | 2 | 2 | 100.0% | - |
| 74 | unsupported surveil spell | 2 | 1 | 50.0% | unsupported Oracle construct |
| 75 | unsupported Cycling ability | 1 | 1 | 100.0% | - |
| 76 | unsupported hand Cycling grant | 1 | 0 | 0.0% | unsupported counter placement |

## Capability clusters

A fully unlockable card has every distinct diagnostic summary in one capability cluster. Constituent summaries list the diagnostics currently observed in that cluster.

| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |
| --- | ---: | ---: | --- |
| shared-ability-content | 16,595 | 11,238 | unsupported ability content; unsupported counter placement; unsupported counter spell; unsupported damage spell; unsupported delayed effect; unsupported destroy spell; unsupported dies trigger body; unsupported dies trigger effect; unsupported discard spell; unsupported draw spell; unsupported draw/discard trigger effect; unsupported enter trigger effect; unsupported exile spell; unsupported explore spell; unsupported fight spell; unsupported gain-control spell; unsupported group power/toughness spell; unsupported investigate spell; unsupported life spell; unsupported manifest spell; unsupported mill spell; unsupported modal ability; unsupported multiple spell abilities; unsupported ordered effect sequence; unsupported phase/step trigger phrase effect; unsupported power/toughness spell; unsupported proliferate spell; unsupported regenerate spell; unsupported return spell; unsupported scry spell; unsupported search effect; unsupported tap spell; unsupported temporary keyword spell; unsupported triggered ability effect; unsupported untap spell |
| trigger-pattern | 5,417 | 2,633 | unsupported dies trigger; unsupported dies trigger phrase; unsupported draw/discard trigger; unsupported enter trigger; unsupported phase/step trigger phrase; unsupported triggered ability |
| static-declaration | 6,971 | 1,795 | unsupported Enchant ability; unsupported Protection ability; unsupported Read ahead ability; unsupported hand Cycling grant; unsupported keyword ability; unsupported mixed keyword ability; unsupported parameterized keyword; unsupported static ability; unsupported static rule declaration |
| activation | 2,227 | 1,037 | unsupported Cycling ability; unsupported Equip ability; unsupported Mutate ability; unsupported Ninjutsu ability; unsupported activated ability; unsupported cost; unsupported loyalty ability; unsupported mana ability; unsupported mana symbol |
| replacement | 1,457 | 326 | unsupported conditional enters-tapped replacement; unsupported counter-placement replacement; unsupported damage replacement; unsupported enters-tapped replacement; unsupported enters-with-counters replacement; unsupported self zone-destination replacement |
| recognition-fallback | 6,739 | 268 | unsupported Oracle construct; unsupported ability word; unsupported reminder ability; unsupported unknown ability |
| other | 253 | 182 | incomplete executable lowering; unsupported card layout; unsupported package letter; unsupported sacrifice spell; unsupported surveil spell; unsupported type line; validation failed: oracle-without-abilities |
