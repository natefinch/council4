# Card-Support Planning Report

Capability-aware blockers for eligible paper cards that cannot yet be generated. Each distinct diagnostic summary and capability is counted at most once per card.

## Diagnostic reasons

A sole blocker is the card's only distinct diagnostic summary. The most common co-blocker excludes the reason in its own row.

| Rank | Reason | Affected cards | Sole blockers | Sole blocker % | Most common co-blocker |
| ---: | --- | ---: | ---: | ---: | --- |
| 1 | unsupported Oracle construct | 6,341 | 0 | 0.0% | unsupported static ability |
| 2 | unsupported static ability | 5,854 | 1,196 | 20.4% | unsupported Oracle construct |
| 3 | unsupported ordered effect sequence | 4,632 | 3,108 | 67.1% | unsupported ability content |
| 4 | unsupported triggered ability | 4,628 | 2,147 | 46.4% | unsupported Oracle construct |
| 5 | unsupported ability content | 3,900 | 993 | 25.5% | unsupported Oracle construct |
| 6 | unsupported activated ability | 1,601 | 719 | 44.9% | unsupported Oracle construct |
| 7 | unsupported enters-tapped replacement | 1,192 | 251 | 21.1% | unsupported Oracle construct |
| 8 | unsupported damage spell | 1,157 | 730 | 63.1% | unsupported Oracle construct |
| 9 | unsupported power/toughness spell | 1,100 | 629 | 57.2% | unsupported Oracle construct |
| 10 | unsupported mixed keyword ability | 1,065 | 458 | 43.0% | unsupported Oracle construct |
| 11 | unsupported counter placement | 1,015 | 434 | 42.8% | unsupported Oracle construct |
| 12 | unsupported enter trigger effect | 897 | 473 | 52.7% | unsupported Oracle construct |
| 13 | unsupported ability word | 755 | 194 | 25.7% | unsupported Oracle construct |
| 14 | unsupported phase/step trigger phrase | 639 | 308 | 48.2% | unsupported Oracle construct |
| 15 | unsupported triggered ability effect | 612 | 333 | 54.4% | unsupported Oracle construct |
| 16 | unsupported phase/step trigger phrase effect | 597 | 253 | 42.4% | unsupported Oracle construct |
| 17 | unsupported return spell | 592 | 346 | 58.4% | unsupported Oracle construct |
| 18 | unsupported destroy spell | 586 | 392 | 66.9% | unsupported Oracle construct |
| 19 | unsupported exile spell | 527 | 206 | 39.1% | unsupported ordered effect sequence |
| 20 | unsupported temporary keyword spell | 515 | 291 | 56.5% | unsupported Oracle construct |
| 21 | unsupported life spell | 406 | 249 | 61.3% | unsupported Oracle construct |
| 22 | unsupported search effect | 406 | 245 | 60.3% | unsupported ability content |
| 23 | unsupported mana ability | 376 | 191 | 50.8% | unsupported static ability |
| 24 | unsupported modal ability | 333 | 256 | 76.9% | unsupported Oracle construct |
| 25 | unsupported untap spell | 212 | 92 | 43.4% | unsupported static ability |
| 26 | unsupported Enchant ability | 212 | 27 | 12.7% | unsupported static ability |
| 27 | unsupported enter trigger | 209 | 121 | 57.9% | unsupported Oracle construct |
| 28 | unsupported enters-with-counters replacement | 193 | 35 | 18.1% | unsupported Oracle construct |
| 29 | unsupported draw spell | 189 | 93 | 49.2% | unsupported Oracle construct |
| 30 | unsupported unknown ability | 186 | 0 | 0.0% | unsupported Oracle construct |
| 31 | unsupported regenerate spell | 180 | 100 | 55.6% | unsupported static ability |
| 32 | unsupported tap spell | 170 | 85 | 50.0% | unsupported static ability |
| 33 | unsupported mana symbol | 158 | 77 | 48.7% | unsupported enters-tapped replacement |
| 34 | unsupported gain-control spell | 154 | 78 | 50.6% | unsupported static ability |
| 35 | unsupported discard spell | 134 | 74 | 55.2% | unsupported Oracle construct |
| 36 | unsupported keyword ability | 134 | 38 | 28.4% | unsupported triggered ability |
| 37 | unsupported sacrifice spell | 129 | 69 | 53.5% | unsupported Oracle construct |
| 38 | unsupported counter spell | 108 | 82 | 75.9% | unsupported ability content |
| 39 | unsupported dies trigger body | 108 | 49 | 45.4% | unsupported static ability |
| 40 | unsupported multiple spell abilities | 98 | 91 | 92.9% | unsupported ability content |
| 41 | unsupported dies trigger effect | 88 | 43 | 48.9% | unsupported Oracle construct |
| 42 | unsupported cost | 88 | 0 | 0.0% | unsupported activated ability |
| 43 | unsupported mill spell | 82 | 50 | 61.0% | unsupported ability content |
| 44 | unsupported loyalty ability | 82 | 0 | 0.0% | unsupported ordered effect sequence |
| 45 | unsupported parameterized keyword | 68 | 13 | 19.1% | unsupported triggered ability |
| 46 | unsupported Equip ability | 63 | 15 | 23.8% | unsupported static ability |
| 47 | unsupported type line | 61 | 59 | 96.7% | unsupported Oracle construct |
| 48 | unsupported group power/toughness spell | 37 | 25 | 67.6% | unsupported Oracle construct |
| 49 | unsupported damage replacement | 35 | 20 | 57.1% | unsupported static ability |
| 50 | unsupported reminder ability | 34 | 0 | 0.0% | unsupported Oracle construct |
| 51 | unsupported dies trigger | 33 | 13 | 39.4% | unsupported Oracle construct |
| 52 | unsupported manifest spell | 31 | 21 | 67.7% | unsupported activated ability |
| 53 | unsupported fight spell | 30 | 14 | 46.7% | unsupported ordered effect sequence |
| 54 | unsupported conditional enters-tapped replacement | 25 | 2 | 8.0% | unsupported ability content |
| 55 | incomplete executable lowering | 24 | 17 | 70.8% | unsupported Oracle construct |
| 56 | unsupported draw/discard trigger effect | 23 | 21 | 91.3% | unsupported Oracle construct |
| 57 | unsupported card layout | 20 | 20 | 100.0% | - |
| 58 | unsupported counter-placement replacement | 19 | 6 | 31.6% | unsupported Oracle construct |
| 59 | unsupported dies trigger phrase | 15 | 7 | 46.7% | unsupported static ability |
| 60 | unsupported delayed effect | 13 | 8 | 61.5% | unsupported Oracle construct |
| 61 | unsupported Protection ability | 10 | 3 | 30.0% | unsupported Oracle construct |
| 62 | unsupported explore spell | 9 | 5 | 55.6% | unsupported Oracle construct |
| 63 | unsupported Read ahead ability | 9 | 0 | 0.0% | unsupported ordered effect sequence |
| 64 | validation failed: oracle-without-abilities | 7 | 7 | 100.0% | - |
| 65 | unsupported investigate spell | 7 | 1 | 14.3% | unsupported triggered ability |
| 66 | unsupported static rule declaration | 7 | 0 | 0.0% | unsupported ability content |
| 67 | unsupported package letter | 6 | 6 | 100.0% | - |
| 68 | unsupported scry spell | 6 | 4 | 66.7% | unsupported Oracle construct |
| 69 | unsupported draw/discard trigger | 5 | 2 | 40.0% | unsupported ordered effect sequence |
| 70 | unsupported proliferate spell | 5 | 1 | 20.0% | unsupported Oracle construct |
| 71 | unsupported token-creation replacement | 4 | 2 | 50.0% | unsupported ability content |
| 72 | unsupported Mutate ability | 3 | 3 | 100.0% | - |
| 73 | unsupported self zone-destination replacement | 3 | 3 | 100.0% | - |
| 74 | unsupported Ninjutsu ability | 3 | 0 | 0.0% | unsupported triggered ability |
| 75 | unsupported surveil spell | 2 | 1 | 50.0% | unsupported Oracle construct |
| 76 | unsupported Cycling ability | 1 | 1 | 100.0% | - |
| 77 | unsupported hand Cycling grant | 1 | 0 | 0.0% | unsupported counter placement |

## Capability clusters

A fully unlockable card has every distinct diagnostic summary in one capability cluster. Constituent summaries list the diagnostics currently observed in that cluster.

| Capability | Affected cards | Fully unlockable cards | Constituent diagnostic summaries |
| --- | ---: | ---: | --- |
| shared-ability-content | 16,581 | 11,211 | unsupported ability content; unsupported counter placement; unsupported counter spell; unsupported damage spell; unsupported delayed effect; unsupported destroy spell; unsupported dies trigger body; unsupported dies trigger effect; unsupported discard spell; unsupported draw spell; unsupported draw/discard trigger effect; unsupported enter trigger effect; unsupported exile spell; unsupported explore spell; unsupported fight spell; unsupported gain-control spell; unsupported group power/toughness spell; unsupported investigate spell; unsupported life spell; unsupported manifest spell; unsupported mill spell; unsupported modal ability; unsupported multiple spell abilities; unsupported ordered effect sequence; unsupported phase/step trigger phrase effect; unsupported power/toughness spell; unsupported proliferate spell; unsupported regenerate spell; unsupported return spell; unsupported scry spell; unsupported search effect; unsupported tap spell; unsupported temporary keyword spell; unsupported triggered ability effect; unsupported untap spell |
| trigger-pattern | 5,443 | 2,657 | unsupported dies trigger; unsupported dies trigger phrase; unsupported draw/discard trigger; unsupported enter trigger; unsupported phase/step trigger phrase; unsupported triggered ability |
| static-declaration | 7,012 | 1,821 | unsupported Enchant ability; unsupported Protection ability; unsupported Read ahead ability; unsupported hand Cycling grant; unsupported keyword ability; unsupported mixed keyword ability; unsupported parameterized keyword; unsupported static ability; unsupported static rule declaration |
| activation | 2,252 | 1,062 | unsupported Cycling ability; unsupported Equip ability; unsupported Mutate ability; unsupported Ninjutsu ability; unsupported activated ability; unsupported cost; unsupported loyalty ability; unsupported mana ability; unsupported mana symbol |
| replacement | 1,458 | 324 | unsupported conditional enters-tapped replacement; unsupported counter-placement replacement; unsupported damage replacement; unsupported enters-tapped replacement; unsupported enters-with-counters replacement; unsupported self zone-destination replacement; unsupported token-creation replacement |
| recognition-fallback | 6,739 | 268 | unsupported Oracle construct; unsupported ability word; unsupported reminder ability; unsupported unknown ability |
| other | 249 | 179 | incomplete executable lowering; unsupported card layout; unsupported package letter; unsupported sacrifice spell; unsupported surveil spell; unsupported type line; validation failed: oracle-without-abilities |
