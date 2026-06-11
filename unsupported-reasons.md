# Card-Support Planning Report

Capability-aware blockers for eligible paper cards that cannot yet be generated. Each distinct diagnostic summary and capability is counted at most once per card.

## Diagnostic reasons

A sole blocker is the card's only distinct diagnostic summary. The most common co-blocker excludes the reason in its own row.

| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | unsupported Oracle construct | 6,341 | 0 | 0.0% | unsupported static ability |
| 2 | unsupported static ability | 5,854 | 1,175 | 20.1% | unsupported Oracle construct |
| 3 | unsupported activated ability | 5,607 | 2,972 | 53.0% | unsupported Oracle construct |
| 4 | unsupported triggered ability | 4,691 | 2,258 | 48.1% | unsupported Oracle construct |
| 5 | unsupported ordered effect sequence | 2,748 | 2,061 | 75.0% | unsupported spell ability |
| 6 | unsupported enter trigger | 2,744 | 1,269 | 46.2% | unsupported Oracle construct |
| 7 | unsupported spell ability | 1,862 | 407 | 21.9% | unsupported Oracle construct |
| 8 | unsupported enter trigger effect | 1,517 | 879 | 57.9% | unsupported static ability |
| 9 | unsupported phase/step trigger phrase | 1,243 | 568 | 45.7% | unsupported Oracle construct |
| 10 | unsupported enters-tapped replacement | 1,192 | 250 | 21.0% | unsupported Oracle construct |
| 11 | unsupported mixed keyword ability | 1,065 | 457 | 42.9% | unsupported Oracle construct |
| 12 | unsupported ability word | 755 | 194 | 25.7% | unsupported Oracle construct |
| 13 | unsupported phase/step trigger phrase effect | 634 | 266 | 42.0% | unsupported static ability |
| 14 | unsupported triggered ability effect | 586 | 381 | 65.0% | unsupported Oracle construct |
| 15 | unsupported damage spell | 383 | 276 | 72.1% | unsupported spell ability |
| 16 | unsupported mana ability | 376 | 190 | 50.5% | unsupported activated ability |
| 17 | unsupported dies trigger | 356 | 144 | 40.4% | unsupported Oracle construct |
| 18 | unsupported modal ability | 333 | 256 | 76.9% | unsupported Oracle construct |
| 19 | unsupported destroy spell | 298 | 214 | 71.8% | unsupported spell ability |
| 20 | unsupported loyalty ability | 298 | 161 | 54.0% | unsupported Oracle construct |
| 21 | unsupported dies trigger effect | 273 | 182 | 66.7% | unsupported Oracle construct |
| 22 | unsupported return spell | 219 | 131 | 59.8% | unsupported spell ability |
| 23 | unsupported Enchant ability | 212 | 24 | 11.3% | unsupported static ability |
| 24 | unsupported search effect | 210 | 139 | 66.2% | unsupported spell ability |
| 25 | unsupported Saga chapter ability | 202 | 141 | 69.8% | unsupported Oracle construct |
| 26 | unsupported enters-with-counters replacement | 193 | 35 | 18.1% | unsupported activated ability |
| 27 | unsupported unknown ability | 186 | 0 | 0.0% | unsupported Oracle construct |
| 28 | unsupported exile spell | 185 | 85 | 45.9% | unsupported spell ability |
| 29 | unsupported mana symbol | 158 | 76 | 48.1% | unsupported enters-tapped replacement |
| 30 | unsupported power/toughness spell | 143 | 98 | 68.5% | unsupported spell ability |
| 31 | unsupported counter placement | 138 | 89 | 64.5% | unsupported spell ability |
| 32 | unsupported keyword ability | 134 | 38 | 28.4% | unsupported triggered ability |
| 33 | unsupported life spell | 104 | 79 | 76.0% | unsupported spell ability |
| 34 | unsupported dies trigger body effect | 102 | 52 | 51.0% | unsupported static ability |
| 35 | unsupported multiple spell abilities | 98 | 91 | 92.9% | unsupported ability word |
| 36 | unsupported cost | 88 | 0 | 0.0% | unsupported activated ability |
| 37 | unsupported temporary keyword spell | 85 | 52 | 61.2% | unsupported spell ability |
| 38 | unsupported counter spell | 82 | 61 | 74.4% | unsupported spell ability |
| 39 | unsupported dies trigger body | 77 | 33 | 42.9% | unsupported mixed keyword ability |
| 40 | unsupported gain-control spell | 74 | 53 | 71.6% | unsupported spell ability |
| 41 | unsupported parameterized keyword | 68 | 13 | 19.1% | unsupported triggered ability |
| 42 | unsupported Equip ability | 63 | 15 | 23.8% | unsupported static ability |
| 43 | unsupported type line | 61 | 59 | 96.7% | unsupported Oracle construct |
| 44 | unsupported draw spell | 59 | 35 | 59.3% | unsupported spell ability |
| 45 | unsupported draw/discard trigger effect | 57 | 21 | 36.8% | unsupported activated ability |
| 46 | unsupported sacrifice spell | 56 | 37 | 66.1% | unsupported spell ability |
| 47 | unsupported dies trigger phrase | 46 | 23 | 50.0% | unsupported Oracle construct |
| 48 | unsupported discard spell | 36 | 19 | 52.8% | unsupported Oracle construct |
| 49 | unsupported tap spell | 35 | 22 | 62.9% | unsupported spell ability |
| 50 | unsupported damage replacement | 35 | 20 | 57.1% | unsupported static ability |
| 51 | unsupported reminder ability | 34 | 0 | 0.0% | unsupported Oracle construct |
| 52 | unsupported untap spell | 28 | 14 | 50.0% | unsupported spell ability |
| 53 | unsupported conditional enters-tapped replacement | 25 | 2 | 8.0% | unsupported activated ability |
| 54 | incomplete executable lowering | 24 | 17 | 70.8% | unsupported Oracle construct |
| 55 | unsupported card layout | 20 | 20 | 100.0% | - |
| 56 | unsupported draw/discard trigger | 19 | 14 | 73.7% | unsupported Oracle construct |
| 57 | unsupported counter-placement replacement | 19 | 6 | 31.6% | unsupported Oracle construct |
| 58 | unsupported mill spell | 16 | 8 | 50.0% | unsupported spell ability |
| 59 | unsupported cycling trigger effect | 11 | 10 | 90.9% | unsupported hand Cycling grant |
| 60 | unsupported fight spell | 10 | 9 | 90.0% | unsupported Oracle construct |
| 61 | unsupported Protection ability | 10 | 3 | 30.0% | unsupported Oracle construct |
| 62 | unsupported cycling trigger | 9 | 9 | 100.0% | - |
| 63 | unsupported Read ahead ability | 9 | 0 | 0.0% | unsupported Saga chapter ability |
| 64 | validation failed: oracle-without-abilities | 7 | 7 | 100.0% | - |
| 65 | unsupported static rule declaration | 7 | 0 | 0.0% | unsupported damage spell |
| 66 | unsupported manifest spell | 6 | 6 | 100.0% | - |
| 67 | unsupported package letter | 6 | 6 | 100.0% | - |
| 68 | unsupported group power/toughness spell | 5 | 5 | 100.0% | - |
| 69 | unsupported token-creation replacement | 4 | 2 | 50.0% | unsupported activated ability |
| 70 | unsupported Mutate ability | 3 | 3 | 100.0% | - |
| 71 | unsupported self zone-destination replacement | 3 | 3 | 100.0% | - |
| 72 | unsupported proliferate spell | 3 | 1 | 33.3% | unsupported Oracle construct |
| 73 | unsupported Ninjutsu ability | 3 | 0 | 0.0% | unsupported triggered ability |
| 74 | unsupported delayed effect | 3 | 0 | 0.0% | unsupported Oracle construct |
| 75 | unsupported investigate spell | 3 | 0 | 0.0% | unsupported mana symbol |
| 76 | unsupported Cycling ability | 1 | 1 | 100.0% | - |
| 77 | unsupported regenerate spell | 1 | 1 | 100.0% | - |
| 78 | unsupported scry spell | 1 | 1 | 100.0% | - |
| 79 | unsupported explore spell | 1 | 0 | 0.0% | unsupported Oracle construct |
| 80 | unsupported hand Cycling grant | 1 | 0 | 0.0% | unsupported cycling trigger effect |

## Capability clusters

A fully unlockable card has every distinct diagnostic summary in one capability cluster. Constituent summaries list the diagnostics currently observed in that cluster.

| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |
| --- | ---: | ---: | --- |
| shared-ability-content | 9,288 | 6,700 | unsupported counter placement; unsupported counter spell; unsupported cycling trigger effect; unsupported damage spell; unsupported delayed effect; unsupported destroy spell; unsupported dies trigger body; unsupported dies trigger body effect; unsupported dies trigger effect; unsupported discard spell; unsupported draw spell; unsupported draw/discard trigger effect; unsupported enter trigger effect; unsupported exile spell; unsupported explore spell; unsupported fight spell; unsupported gain-control spell; unsupported group power/toughness spell; unsupported investigate spell; unsupported life spell; unsupported manifest spell; unsupported mill spell; unsupported modal ability; unsupported multiple spell abilities; unsupported ordered effect sequence; unsupported phase/step trigger phrase effect; unsupported power/toughness spell; unsupported proliferate spell; unsupported regenerate spell; unsupported return spell; unsupported scry spell; unsupported search effect; unsupported spell ability; unsupported tap spell; unsupported temporary keyword spell; unsupported triggered ability effect; unsupported untap spell |
| trigger-pattern | 8,554 | 4,581 | unsupported cycling trigger; unsupported dies trigger; unsupported dies trigger phrase; unsupported draw/discard trigger; unsupported enter trigger; unsupported phase/step trigger phrase; unsupported triggered ability |
| activation | 6,393 | 3,537 | unsupported Cycling ability; unsupported Equip ability; unsupported Mutate ability; unsupported Ninjutsu ability; unsupported activated ability; unsupported cost; unsupported loyalty ability; unsupported mana ability; unsupported mana symbol |
| static-declaration | 7,194 | 1,941 | unsupported Enchant ability; unsupported Protection ability; unsupported Read ahead ability; unsupported Saga chapter ability; unsupported hand Cycling grant; unsupported keyword ability; unsupported mixed keyword ability; unsupported parameterized keyword; unsupported static ability; unsupported static rule declaration |
| replacement | 1,458 | 323 | unsupported conditional enters-tapped replacement; unsupported counter-placement replacement; unsupported damage replacement; unsupported enters-tapped replacement; unsupported enters-with-counters replacement; unsupported self zone-destination replacement; unsupported token-creation replacement |
| recognition-fallback | 6,739 | 268 | unsupported Oracle construct; unsupported ability word; unsupported reminder ability; unsupported unknown ability |
| other | 174 | 146 | incomplete executable lowering; unsupported card layout; unsupported package letter; unsupported sacrifice spell; unsupported type line; validation failed: oracle-without-abilities |
