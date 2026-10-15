# Production support inventory

Production registry: **17969 entries, 17969 loadable, 40 behavior-verified authored identities**. Frozen compiler: **18700/33016 eligible Oracle records accepted**. These are separate measurements; name-only legacy entries are not certified exact Oracle/face identities or behavior-verified games.

Pinned corpus SHA256 `6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4`; compiler evidence SHA256 `f8f1f1447e8b56484b40ab91f86bd5d3e427b71991a03b3926e64972c1b78601`; Comprehensive Rules 2026-09-25. Reproduce with `mage supportInventory` (no corpus compilation).

## Exact authored Oracle/front admissions

Source revisions bind authored Go (all handlers and dependency declarations), not just CardDefs. Shared-rules revisions bind non-test Go in `mtg/{rules,game,cardprogram}` and `opt`. `go generate ./mtg/cards/authored` updates them; a native test rejects stale bindings. Digests detect changes; they do not prove behavior.

| Card | Oracle identity | Compiler accepted | Loadable | Behavior evidence | Program versions |
| --- | --- | --- | --- | --- | --- |
| Abundance | `eccc9a54-2b56-4a02-9926-258d5b2e25fb/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{eccc9a54-2b56-4a02-9926-258d5b2e25fb/replacement/0 1}]` |
| An Offer You Can't Refuse | `234a734b-ba28-4f1b-9d01-3c3e7d516590/front` | true | true | mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/An_Offer_You_Can't_Refuse, TestProductionCountersDepartedTargetAndClone, TestProductionOfferTokensAndCounterReplacements, TestProductionOfferTreasuresPayAndReplace; TestProductionBatchLiveAndSimulator; WotC 2022-04-29; CR 2026-09-25 111.10a, 608.2b/h, 700.7, 701.6 | `[{234a734b-ba28-4f1b-9d01-3c3e7d516590/spell/0 1}]` |
| Bedevil | `bceecc64-96f1-4e7b-8904-0aef90377764/front` | true | true | mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Bedevil, TestProductionDestructionProtectionAndReplacements/Bedevil, TestProductionDestructionIdentityAndCopies/Bedevil; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.8 | `[{bceecc64-96f1-4e7b-8904-0aef90377764/spell/0 1}]` |
| Blood Moon | `94fac5fe-97d5-4c12-a80c-8efff9d853ae/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[]` |
| Cackling Counterpart | `9e2adca5-f39c-4a09-bcce-8238ebac2c4a/front` | true | true | mtg/rules/authored_copy_spells_test.go: TestAuthoredCreatureCopyTokens/Cackling_Counterpart; TestAuthoredCreatureCopyTargetLegality/Cackling_Counterpart; TestAuthoredCacklingCounterpartFlashback; TestAuthoredCreatureCopyRetargetedSpellAndClone/Cackling_Counterpart | `[{9e2adca5-f39c-4a09-bcce-8238ebac2c4a/spell/0 1}]` |
| Chains of Mephistopheles | `eae87919-6322-4bd2-ae9c-b1ce25d686da/front` | false | true | mtg/rules/chains_of_mephistopheles_test.go; CR 121, 614.5, 614.11, 616 (2026-09-25) | `[{eae87919-6322-4bd2-ae9c-b1ce25d686da/replacement/0 1}]` |
| Counterspell | `cc187110-1148-4090-bbb8-e205694a39f5/front` | true | true | mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Counterspell, TestProductionCountersDepartedTargetAndClone, TestProductionCountersRetargetedCopy; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.6 | `[{cc187110-1148-4090-bbb8-e205694a39f5/spell/0 1}]` |
| Disallow | `88b51e15-6630-4e14-a6b8-db0aa12e34ef/front` | true | true | mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Disallow, TestProductionCountersDepartedTargetAndClone, TestProductionDisallowSourceDepartureAndManaAbility, TestProductionDisallowDelayedTrigger; TestProductionBatchLiveAndSimulator; WotC 2017-02-09; CR 2026-09-25 113.7a, 603.7b, 605, 608.2b, 701.6 | `[{88b51e15-6630-4e14-a6b8-db0aa12e34ef/spell/0 1}]` |
| Disenchant | `a7e97fa9-4b72-4548-b854-5be5f18a6f1a/front` | true | true | mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Disenchant, TestProductionDestructionProtectionAndReplacements/Disenchant, TestProductionDestructionIdentityAndCopies/Disenchant; TestProductionBatchLiveAndSimulator; WotC ruling 2004-10-04; CR 2026-09-25 608.2b, 701.8 | `[{a7e97fa9-4b72-4548-b854-5be5f18a6f1a/spell/0 1}]` |
| Dualcaster Mage | `8eb7c0a5-6190-40de-b473-2d1daa3bbe28/front` | true | true | mtg/rules/authored_copy_spells_test.go: TestAuthoredDualcasterMageFlashAndSourceDeparture; TestAuthoredSpellCopiesChoicesAndCasting/Dualcaster_Mage; TestAuthoredSpellCopiesUntargeted/Dualcaster_Mage | `[{8eb7c0a5-6190-40de-b473-2d1daa3bbe28/enter/0 1}]` |
| Eladamri's Call | `4acb6612-54e8-428d-acb6-c7259a5ad6a8/front` | true | true | mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Eladamri's_Call, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; WotC 2004-10-04; CR 2026-09-25 701.23b/e | `[{4acb6612-54e8-428d-acb6-c7259a5ad6a8/spell/0 1}]` |
| Essence Scatter | `46665089-aa3d-44c3-964d-6638dfbb5782/front` | true | true | mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Essence_Scatter, TestProductionCountersDepartedTargetAndClone, TestProductionCountersRetargetedCopy; TestProductionBatchLiveAndSimulator; WotC 2026-03-20; CR 2026-09-25 608.2b, 701.6 | `[{46665089-aa3d-44c3-964d-6638dfbb5782/spell/0 1}]` |
| Fabricate | `422e1869-134f-463d-9fa1-86b66a998b3e/front` | true | true | mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Fabricate, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e | `[{422e1869-134f-463d-9fa1-86b66a998b3e/spell/0 1}]` |
| Fated Infatuation | `ae771e78-1d29-4b8c-8511-9da196a78dcf/front` | true | true | mtg/rules/authored_copy_spells_test.go: TestAuthoredCreatureCopyTokens/Fated_Infatuation; TestAuthoredCreatureCopyTargetLegality/Fated_Infatuation; TestAuthoredFatedInfatuationScry; TestAuthoredCreatureCopyRetargetedSpellAndClone/Fated_Infatuation | `[{ae771e78-1d29-4b8c-8511-9da196a78dcf/spell/0 1}]` |
| Generous Gift | `fae37e28-e137-4177-b973-fa8b4dd8f409/front` | true | true | mtg/rules/removal_token_spells_test.go: TestAuthoredGenerousGiftScenario; TestAuthoredRemovalTokensProtectedTargets/Generous_Gift; TestAuthoredRemovalTokensAllIllegalTargets/Generous_Gift | `[{fae37e28-e137-4177-b973-fa8b4dd8f409/spell/0 1}]` |
| Gifts Ungiven | `58aec411-167d-4709-8560-793eaaed62c5/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{58aec411-167d-4709-8560-793eaaed62c5/spell/0 1}]` |
| Gilded Drake | `7f06c098-6482-4bf3-a9a1-110d6d5b5703/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{7f06c098-6482-4bf3-a9a1-110d6d5b5703/enter/0 1}]` |
| Hive Mind | `f97e405d-4c24-4d4d-9e19-e349c073113c/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{f97e405d-4c24-4d4d-9e19-e349c073113c/cast/0 1}]` |
| Humility | `ed7bdb3e-5c51-4547-9266-76a791e0b2b0/front` | false | true | mtg/rules/humility_opalescence_test.go; mtg/rules/challenge_followup_test.go: TestHumilityRemovesThaliaTaxAndRhonasRestrictions; TestHumilityAllowsLaterExternallyGrantedRuleAbility; TestHumilityPrintedHastePreservesControlHistoryForAttackTapUntap; CR 302.6, 613.1/4/6–8/11 (2026-09-25) | `[]` |
| Idyllic Tutor | `57c9ff89-dd30-467b-bb3e-499eeea8cb94/front` | true | true | mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Idyllic_Tutor, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e | `[{57c9ff89-dd30-467b-bb3e-499eeea8cb94/spell/0 1}]` |
| Intuition | `3c9faba7-f2d3-4978-be94-020dc8003dc0/front` | false | true | mtg/rules/search_partition_spells_test.go: TestAuthoredIntuitionForcedSearchQuantity, TestAuthoredIntuitionCastAndTargetLegality, TestAuthoredSearchPartitionZoneReplacementsAndDeferredShuffle; CR 2026-09-25 701.23d, 608.2b; WotC ruling 2006-07-15 | `[{3c9faba7-f2d3-4978-be94-020dc8003dc0/spell/0 1}]` |
| Mindslaver | `a806f4ee-f48a-46c3-8772-5aaabebe4c7e/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{a806f4ee-f48a-46c3-8772-5aaabebe4c7e/activate/0 1}]` |
| Muldrotha, the Gravetide | `e4625704-1d52-44e4-804f-2f45644d76ac/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[]` |
| Naturalize | `bdb3ca68-ec1f-4e16-81cc-d23f8f52c728/front` | true | true | mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Naturalize, TestProductionDestructionProtectionAndReplacements/Naturalize, TestProductionDestructionIdentityAndCopies/Naturalize; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.8 | `[{bdb3ca68-ec1f-4e16-81cc-d23f8f52c728/spell/0 1}]` |
| Negate | `3407fe41-fdd3-4119-8f70-4bc4590a379f/front` | true | true | mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Negate, TestProductionCountersDepartedTargetAndClone, TestProductionCountersRetargetedCopy; TestProductionBatchLiveAndSimulator; WotC 2026-01-27; CR 2026-09-25 608.2b, 701.6 | `[{3407fe41-fdd3-4119-8f70-4bc4590a379f/spell/0 1}]` |
| Nim Deathmantle | `66d41377-626d-4ae6-ba86-17bf0c8b3362/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{66d41377-626d-4ae6-ba86-17bf0c8b3362/death/0 1}]` |
| Opalescence | `59489b46-9d02-4f3c-bcd0-884e7605e9a5/front` | false | true | mtg/rules/humility_opalescence_test.go; mtg/rules/challenge_followup_test.go: TestOpalescenceBrokenSkyManaValueAndDoubleFacedTokenCopies; TestCopiedOpalescenceUsesNewAbilityTimestampAndPhysicalExclusion; Scryfall rulings 2006-02-01, 2009-10-01; CR 202.3b, 613.7a, 707.8a, 712.8e (2026-09-25) | `[]` |
| Pongify | `05849bd6-8f38-4031-be2b-e2aa03beb8cc/front` | true | true | mtg/rules/removal_token_spells_test.go: TestAuthoredRemovalTokensCastAndTargetTypes/Pongify; TestAuthoredRemovalTokensProtectedTargets/Pongify; TestAuthoredRemovalTokensAllIllegalTargets/Pongify | `[{05849bd6-8f38-4031-be2b-e2aa03beb8cc/spell/0 1}]` |
| Rapid Hybridization | `06692cd9-ac2f-4a32-8fd1-043ba3c0fe71/front` | true | true | mtg/rules/removal_token_spells_test.go: TestAuthoredRemovalTokensCastAndTargetTypes/Rapid_Hybridization; TestAuthoredRemovalTokensProtectedTargets/Rapid_Hybridization; TestAuthoredRemovalTokensAllIllegalTargets/Rapid_Hybridization | `[{06692cd9-ac2f-4a32-8fd1-043ba3c0fe71/spell/0 1}]` |
| Realms Uncharted | `e21c8fc6-d4ef-42b6-b11e-d9c931da1387/front` | false | true | mtg/rules/search_partition_spells_test.go: TestAuthoredRealmsUnchartedDifferentNamesAndFewerFinds, TestAuthoredRealmsUnchartedLibraryExhaustionAndLivingOpponent, TestAuthoredSearchPartitionOppositionAgent; CR 2026-09-25 201.2b, 701.23b, 608.2d; WotC ruling 2010-06-15 | `[{e21c8fc6-d4ef-42b6-b11e-d9c931da1387/spell/0 1}]` |
| Resculpt | `814e321d-16fd-4f7b-a8d8-2b089be76f2c/front` | true | true | mtg/rules/removal_token_spells_test.go: TestAuthoredRemovalTokensCastAndTargetTypes/Resculpt; TestAuthoredRemovalTokensControllerSnapshot/Resculpt; TestAuthoredRemovalTokensZoneAndTokenReplacements/Resculpt; TestAuthoredRemovalTokensAllIllegalTargets/Resculpt | `[{814e321d-16fd-4f7b-a8d8-2b089be76f2c/spell/0 1}]` |
| Reverberate | `a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca/front` | true | true | mtg/rules/authored_copy_spells_test.go: TestAuthoredSpellCopiesChoicesAndCasting/Reverberate; TestAuthoredSpellCopiesIllegalTarget/Reverberate; TestAuthoredSpellCopiesUntargeted/Reverberate | `[{a1f55890-31c5-4ed4-a2cd-7a4a9f05f8ca/spell/0 1}]` |
| Skullbriar, the Walking Grave | `088ac883-6f44-4046-bf19-81c7fd952e55/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[]` |
| Steelshaper's Gift | `d9abda7e-6ca2-42ea-ab24-c542e57014f1/front` | true | true | mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Steelshaper's_Gift, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e | `[{d9abda7e-6ca2-42ea-ab24-c542e57014f1/spell/0 1}]` |
| Sylvan Scrying | `ee24bf27-484d-4e1c-998e-6a74e3d3f6c4/front` | true | true | mtg/rules/production_tutors_test.go: TestProductionTutorsSearchChoicesAndVisibility/Sylvan_Scrying, TestProductionTutorsCastTimingAndPayment, TestProductionTutorsReplacementAndSearchControl, TestProductionTutorsCopyAndClone; TestProductionBatchLiveAndSimulator; CR 2026-09-25 701.23b/e | `[{ee24bf27-484d-4e1c-998e-6a74e3d3f6c4/spell/0 1}]` |
| Terastodon | `17975a20-2c13-4325-be1e-0bcda5063a2d/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{17975a20-2c13-4325-be1e-0bcda5063a2d/enter/0 1}]` |
| Terminate | `6257c2fd-005f-41e3-8a72-af76df1eb134/front` | true | true | mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Terminate, TestProductionDestructionProtectionAndReplacements/Terminate, TestProductionDestructionIdentityAndCopies/Terminate; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.8, 701.19 | `[{6257c2fd-005f-41e3-8a72-af76df1eb134/spell/0 1}]` |
| Twincast | `8f878efc-850f-43d2-a6fe-5ea8d1dd5afb/front` | true | true | mtg/rules/authored_copy_spells_test.go: TestAuthoredSpellCopiesChoicesAndCasting/Twincast; TestAuthoredSpellCopiesIllegalTarget/Twincast; TestAuthoredSpellCopiesUntargeted/Twincast | `[{8f878efc-850f-43d2-a6fe-5ea8d1dd5afb/spell/0 1}]` |
| Vindicate | `63c1ac21-e3d8-40c2-8c09-3f31c52992ef/front` | true | true | mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Vindicate, TestProductionDestructionProtectionAndReplacements/Vindicate, TestProductionDestructionIdentityAndCopies/Vindicate; TestProductionBatchLiveAndSimulator; CR 2026-09-25 307, 608.2b, 701.8 | `[{63c1ac21-e3d8-40c2-8c09-3f31c52992ef/spell/0 1}]` |
| Worldgorger Dragon | `a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c/front` | false | true | mtg/rules named-card scenarios; mixed production acceptance (#3278) | `[{a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c/enter/0 1} {a628186d-b7d9-40a5-9ae2-fbc9d2a14c7c/leave/0 1}]` |

- Abundance: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- An Offer You Can't Refuse: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Bedevil: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Blood Moon: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Cackling Counterpart: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Chains of Mephistopheles: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Counterspell: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Disallow: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Disenchant: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Dualcaster Mage: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Eladamri's Call: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Essence Scatter: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Fabricate: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Fated Infatuation: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Generous Gift: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Gifts Ungiven: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Gilded Drake: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Hive Mind: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Humility: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Idyllic Tutor: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Intuition: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Mindslaver: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Muldrotha, the Gravetide: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Naturalize: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Negate: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Nim Deathmantle: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Opalescence: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Pongify: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Rapid Hybridization: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Realms Uncharted: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Resculpt: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Reverberate: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Skullbriar, the Walking Grave: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Steelshaper's Gift: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Sylvan Scrying: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Terastodon: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Terminate: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Twincast: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Vindicate: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

- Worldgorger Dragon: authored `sha256:b6c1d1ff753e7bb9ea9f5bdd918ccb246494ffad9ef29ab54e4cd35e0e54962e`; shared rules `sha256:1bdd9556783c70a003f0f3d387ca482d061a10f4c5ace17387bf375ad3c3b57a`.

## Fidelity and consumers

`cards.Registry.Inventory/SupportFor` is authoritative for installed entries; deck loading returns the same evidence. JSON inventory output joins corpus IDs without guessing legacy identity. Engine, Simulator, CLI and WASM share the installed catalog. A deliberately absent handler fails checked loading or produces `unsupported-execution`, never a rules draw.

Rules wins/draws, turn budgets, work budgets, unsupported and invalid execution are distinct. Action/choice fallbacks and invalid answers are counted separately. `GameResult.Exact()` rejects declared restrictions and policy fallbacks. Limits include X/multikicker 20 ([#3284](https://github.com/natefinch/council4/issues/3284)), representative combat declarations ([#3126](https://github.com/natefinch/council4/issues/3126)) and incomplete evaluation knowledge. Search action width, determinization count, lookahead and deep width are policy budgets, not rules. Unknown authored/effect valuation is explicit; search executes real programs, not metadata.

See [mixed production validation](docs/perf/mixed-production.md) for executed scenarios and measurements. CI runs code/scenario checks and registry preservation for authored edits, not frozen-corpus regeneration. Legacy compiler diagnostics remain in `unsupported-reasons.md` and `card-backlog.md`; they are not the automatic authoring roadmap.
