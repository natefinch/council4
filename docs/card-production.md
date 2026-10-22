# Bounded authored production queue (#3282)

**Historical production checkpoint.** Counts, reviews, benchmarks and validation
results below record completed work, not current completion requirements. The
[repository validation policy](../.github/copilot-instructions.md#verify-proportionally)
supersedes old full-suite/lint, fingerprint and report-refresh gates. For new work,
use the [authoring guide](card-authoring.md) and live queue, not hard-coded
continuation examples or superseded blocker statuses below.

## Continuous production (#3311)

**Publication update:** successful continuous-production cards and Engine
prerequisites now land on remote `main` through PRs as part of the run, after
integrating latest main, local validation, applicable review and all exact-head
CI including required `Race Tests`. The repository's
[Dynamic Workflow](dynamic-card-production.md) provides resumable execution,
serialized landing and measured card/Engine implementation, review and repair
stages. It does not start automatically.

The production loop now replenishes reviewed batches, resumes an owner's exact
claim, exposes pending Engine prerequisites, and explicitly reassesses blocked
cards after runtime evidence. See the canonical
[continuation commands](card-authoring.md#reviewed-continuation-commands).
The coordinator is the only queue writer; workers implement isolated slices
and integration is serialized. Exhausting saved batches is not completion of
an authorized remaining-cards-and-Engine assignment.

The first cycle adds **four genuinely new exact authored cards**, not legacy
migrations:

| Card | Oracle ID | Named Engine evidence |
| --- | --- | --- |
| Mundungu | `3fb4161d-a5b9-411d-942c-2fdb039e7bde` | `activated_utility_family_test.go`: spell-only targets, current-controller combined payment, failed payment, uncounterable/departed targets and clone |
| Sly Instigator | `8c2dd8dc-8623-4518-b3fd-dcb9f647fa98` | Same file: opponent targeting/revalidation, source departure and real ability copy, actual block legality and non-Player1 expiry |
| Half-Elf Monk | `00d11662-9d10-4915-b80c-c1ad54b4e71b` | Same file: vigilance attack, paid tap activation, sickness/timing and target departure |
| Myr Custodian | `43923f24-7e95-4934-9a5c-8d9a5e06ddea` | `myr_custodian_test.go` and independent `production_scry_independent_test.go`: paid default-registry entry, source departure, APNAP private staged choices, public decisions, controlled-player replay and clone |

Myr Custodian required two shared changes: #3312 sanitizes single/partition/order
scry logs without removing private information from the deciding agent; #3313
stages group-member scry participation and APNAP placement decisions before
simultaneous library commits/events. The complete official ruling exposed the
second requirement during implementation. A loop of immediately completed
single-player scries was not accepted as a shortcut. Public revealed ordering
remains public; this is not a general audit of every private-choice operation.

The card authors verified the pinned Oracle ARRAY
`6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4`
and official CR effective 2026-09-25,
`8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca`,
before design. Exact printings and their Scryfall rulings endpoints are retained
in the manifest. Complete responses were read on 2026-10-22: Mundungu and
Half-Elf Monk have empty lists; Sly Instigator has seven rulings dated
2021-04-16/2023-06-30; Myr Custodian has one dated 2023-02-04.
The verified-copy source is the supplied official `high-review-rules.txt`;
full source locations and the corresponding hashes are in the implementation
handoffs and Myr's source comment.

Rules 117.1b, 118.1-118.3c, 302.6, 601.2c, 602.2-602.2b, 608.2b,
611.2a-c and 701.6/701.15 govern actual activation, complete mana/life payment,
revalidation, countering, duration and goad. For Myr, 101.4/401.1-6/701.22a-d
require private look information, APNAP decisions before simultaneous actions,
and no scry event for zero versus a real positive scry of an empty library.
113.7a/603.3/707.10 preserve independent source-departed and copied triggers
with the actual ability controller.

The loop was also retuned from real delivery failures: ordinary queue commands
no longer require an unavailable historical compiler report; failed transitions
are atomic; saved checkpoints use a temporary file and guarded rename; `verify`
binds exact pinned Oracle metadata without fabricating source reads; and the
production fixture accepts legitimate whole-batch durable claims while rejecting
dangling/partial/multiowner claims. Four previously landed cards (Ghost Vacuum,
Marchesa, Tooth Collector and Necromantic Selection) were explicitly reconciled
with their existing scenarios and source metadata. They are **not** four more
new cards in this cycle.

Batch `07-activated-utility` contains the first three cards above. The saved
follow-on `08-scry-interaction` contains Eager Construct, Sea God's Revenge and
Devout Decree, with #3312/#3313 as explicit runtime prerequisites. Recording
checked prerequisite evidence makes it selectable; closed issue numbers alone
do not. Resume with `cardqueue -operation next -owner YOUR_SESSION_ID`, supplying
`-corpus` when the verified array is not at the command's default path.
Normal remaining-card production continues with that batch, not another model
comparison. Resource accounting distinguishes queue setup, coordination/review,
shared Engine/proving-card work, and card authoring; credits are not USD.

## Retained compiler installation checkpoint

The subsequently requested full compiler import checks in all 18,700 accepted
definitions (17,946 cards and 754 tokens), while preserving the authored
implementations and prior legacy sources. The saved queue is reconciled with
the expanded production inventory: `installed-compiled` is distinct from the
fifteen behavior-verified batch admissions below. Newly installed members of
the three saved continuation batches no longer need authored generation; `next`
skips them rather than duplicating their sources. The eight known whole-card
blocks remain unchanged, including the private tutors excluded from the default
playable registry. The historical population, batch measurements and evidence
below describe the completed migration checkpoint, not the current import.
Use `cardqueue -operation check` and `support-inventory.md` for current counts.

Current accounting: **15,903 installed-compiled**, 14,285 unclaimed, eight
blocked, fifteen batch-verified and 5,728 excluded, plus 2,805 retained baseline
associations. There are no active claims. The import also corrects 23 historical
name-only associations: a token sharing an ordinary card's name is not that card,
and an installed standalone spell does not implement a prepared creature with
that spell as its alternate face. Their former candidate keys remain in
`corrected_name_only_associations`; their actual work records are retained, not
dropped. These corrections do not remove any existing production definition.

The [saved manifest](../testdata/card-authoring/production.json) is the work
accounting source, not an Oracle compiler roadmap. Continue with the installed
[card-authoring skill](../.agents/skills/card-authoring/SKILL.md) and
[canonical guide](card-authoring.md), including its existing scaffold.

## Population and ordering

Baseline local main: `9a47dfd25ba4ef870cb0b780e617e4add667da22` (25 authored
identities: eleven proving cards plus fourteen frozen pilot admissions).
Oracle JSON SHA256:
`6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4`;
printing revisions and live rulings hashes/dates are recorded per selected
identity. Official CR effective **2026-09-25**, SHA256
`8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca`.

| Accounting | Count |
| --- | ---: |
| Pinned records, before canonical omission | 38,751 |
| Disowned records omitted under `cardgen.DisownedCard` (never listed) | 7 |
| Explicit policy exclusions | 5,728 |
| Eligible sanctioned-paper identities | 33,016 |
| Baseline loadable production **associations** | 2,828 |
| Initial mixed-production absence | 30,188 |
| Newly implemented/verified in these three batches | 15 |
| Remaining unclaimed | 30,165 |
| Whole blocked cards | 8 |
| Active claims after integration | 0 |

The complete remaining eligible population is retained exactly once, with an
Oracle ID **and printing revision**, even for non-Commander cards and
compiler-accepted but uninstalled cards. Exclusions retain the actual canonical
policy reason: memorabilia 2,696; no sanctioned-paper legality 1,683;
Alchemy 742; special format 538; digital-only 69. Banned/restricted paper cards
remain eligible. The seven omitted disowned records are not enumerated.

`baseline_production` retains the registry's exact authored keys and **candidate
name-only legacy keys**. A corpus association is not an invented exact legacy
registration or behavioral certification. Its population differs from entry
count because names/faces can associate with multiple corpus records.
Production moves from 2,804 to **2,819 loadable entries**, of which **40** are
exact authored admissions; legacy entries remain 2,779. The frozen compiler
18,700/33,016 is separate evidence, unchanged.

Priority uses actual occurrence in the four committed perf decks and the
recorded Moxfield mainboard/commanders; every input hash is saved. Bedevil and
An Offer each occur in one recorded deck. No occurrence is inferred from
EDHREC, prices, card titles or copy quantity. These deck completions lead their
established-capability batches. Remaining records use occurrence descending,
then deterministic family/layout and Oracle identity. `unreviewed/normal` and
other layout families mean **not yet semantically reviewed**, not unsupported
Engine capability diagnoses. Compiler diagnostics remain in `supportInventory`,
never selection/ranking.

## Completed batches

All printed abilities, costs and target unions are implemented. Fourteen ordinary
follow-ons needed no Engine changes; one narrow existing mana-payment adapter
defect needed a separate focused prerequisite commit (#3290). No parser/lowering
change or new Engine primitive was added.

| Batch/family | Card | Oracle ID |
| --- | --- | --- |
| 01 destruction | Bedevil | `bceecc64-96f1-4e7b-8904-0aef90377764` |
| 01 | Terminate | `6257c2fd-005f-41e3-8a72-af76df1eb134` |
| 01 | Vindicate | `63c1ac21-e3d8-40c2-8c09-3f31c52992ef` |
| 01 | Disenchant | `a7e97fa9-4b72-4548-b854-5be5f18a6f1a` |
| 01 | Naturalize | `bdb3ca68-ec1f-4e16-81cc-d23f8f52c728` |
| 02 stack interaction | An Offer You Can't Refuse | `234a734b-ba28-4f1b-9d01-3c3e7d516590` |
| 02 | Counterspell | `cc187110-1148-4090-bbb8-e205694a39f5` |
| 02 | Negate | `3407fe41-fdd3-4119-8f70-4bc4590a379f` |
| 02 | Essence Scatter | `46665089-aa3d-44c3-964d-6638dfbb5782` |
| 02 | Disallow | `88b51e15-6630-4e14-a6b8-db0aa12e34ef` |
| 03 library search | Eladamri's Call | `4acb6612-54e8-428d-acb6-c7259a5ad6a8` |
| 03 | Steelshaper's Gift | `d9abda7e-6ca2-42ea-ab24-c542e57014f1` |
| 03 | Fabricate | `422e1869-134f-463d-9fa1-86b66a998b3e` |
| 03 | Idyllic Tutor | `57c9ff89-dd30-467b-bb3e-499eeea8cb94` |
| 03 | Sylvan Scrying | `ee24bf27-484d-4e1c-998e-6a74e3d3f6c4` |

`production_destruction_test.go` executes exact paid casts, instant/sorcery
timing, type unions and changes, hexproof, blink/new identity, indestructible,
shield counters, regeneration, umbra armor, zone replacements, retargeted copies
and clones. Terminate prevents regeneration without bypassing other protections.

`production_counters_test.go` executes real spell costs/target counts and stack
kind/type filters, artifact-creature cases, departed/recast targets, retargeted
copies and clones, legal uncounterable targets, current controller versus owner,
zone/token/entry replacements, simultaneous Treasures and actual colored payment
by explicit tap/self-sacrifice mana activations, even after a Treasure loses
its artifact type (CR 700.7). Canonical #3290's separate focused prerequisite
fix also recognizes the exact unqualified self-sacrifice body for normal cast
and optional **resolving** payments; pre-activation alone would not suffice.
Recognition still rejects other cost constraints, quantities, targets, conditions
and extra effects. Runtime negatives preserve tapped/current-controller and
atomic payment restrictions; actual live/Simulator and An Offer-created
Treasures exercise both payment windows. Disallow counters actual activated/triggered abilities after
source departure and a one-shot delayed trigger without reviving it; actual mana
abilities resolve immediately and cannot be targeted.

`production_tutors_test.go` executes qualified fail-to-find, revelation and private
library visibility, empty/nonmatching libraries, timing/payment, actual
replacement-aware placements, Opposition Agent decision/beneficiary scopes,
copier-relative libraries and clone isolation. `production_batches_test.go`
casts all fifteen through **checked default production identity lookup**, actual
legal actions, live priority and Simulator Apply/ResolvePriority, interacting
with actual legacy cards and Humility. The pre-existing all-authored production
consumer also covers the new cards without dropping prior admissions.

One bounded independent **rules** review covered the nontrivial behaviors and
scenarios; it reported no substantive findings. No architecture/style review
restart or worker fleet was used. A narrowly focused follow-up independently
reviewed the later Treasure self-reference correction and reported no significant
issues; it did not redo the batch/architecture review. A focused review of the
small shared recognition prerequisite likewise reported no significant issues.

### Genuine blocked attempt, not omitted abilities

The third family initially claimed Demonic/Diabolic Tutor plus three qualified
tutors. Actual destination-replacement scenarios failed: ordinary private
`Execute(Search)` directly adds to hand/graveyard and bypasses replacements.
The revealed-handle API is not a legal workaround for a private find.
Both whole private tutors were initially left unregistered, linked to canonical
[#3289](https://github.com/natefinch/council4/issues/3289). The expected-correct
`TestProductionPrivateTutorPlacementBlocker` was retained as a skipped regression,
not a passing card or gameplay waiver.
The claimed attempt was released and the same bounded family substituted
Eladamri's Call/Steelshaper's Gift. Qualified tutors use existing
`Search` + `MoveCards` + `ShuffleLibrary`, so replacements actually execute.
The manifest retains both blocked identities and the selection change.
Process exception recorded for the parent review: the original family was
claimed before implementation, but replacement **unregistered drafts** were
edited during the released replan before their exact identities were reclaimed.
Reclaim preceded verification/registration and there was only one writer. This
does not change the verified count; future work must claim substitutes before
any draft edit, as the canonical guide requires.

The #3289 follow-up fixes ordinary private placement through the existing
replacement-aware zone movement, without using revealed handles. The retained
regression now executes, and complete authored Demonic Tutor and Diabolic Tutor
implementations supersede their unique compiled definitions. Named scenarios
cover mandatory/empty/private finds, printed payment and timing, actual redirected
arrivals and zone versions, search/turn control, Opposition Agent permissions,
copies, independent Simulator branches and mixed production execution. Search
choice logs retain replay indices but not private option identities. The original
third-batch selection history and frozen pilot evidence remain unchanged.

The original six frozen pilot blockers record that run's outcomes: #170 Reiterate,
#2161 Doomsday, #3285 Necromancy, #3286 Panglacial Wurm, #3287 Possibility Storm,
#3288 Warp World. The pilot's frozen pending fields and fourteen/six outcomes
were not rewritten; later exact admissions use current production evidence.
The #3288 follow-up now admits complete authored Warp World: pre-move ownership
counts include tokens and diverted commanders, revealed cards enter in two
cross-player simultaneous stages, Aura decisions use the pre-stage board in
APNAP order, and owners order the remaining revealed cards on the library bottom.
Shared entry replacements and event-time trigger snapshots preserve stage
boundaries without placing triggers on the stack during resolution.
The #3286 follow-up supplies Panglacial Wurm's inspectable library-only static
permission during an actual own-library search, before finding. Shared paid
announcement and mana abilities preserve the original actor/payer under search
or turn control, fix total costs before mana activations, and resume against live
library candidates. Failed proposals preserve order and reverse reversible mana
payments, while CR 733.1 forbids undoing an actual draw, library move, reveal or
shuffle. Named Engine/Simulator and JSON replay scenarios cover these boundaries.
The #3284 follow-up separates exact X/multikicker legality
from policy samples. A production X=21 scenario also proved that six existing
compiler-loaded definitions truncated counted targets at twenty. Their complete
authored replacements now execute exact target batches: Death Denied, Wildest
Dreams, Rats' Feast, Shattered Crypt, Return to the Ranks and Champion of Stray
Souls. This is repair of existing loaded cards, not another production batch.
Generated definitions and the pinned corpus are retained; public target handles,
replacement-aware moves and the existing simultaneous battlefield-source batch
are shared Engine operations. Current registry associations point to the exact
replacements. The #3126 follow-up separates complete typed/staged combat
declaration legality from bounded policy samples, including joint blocking
requirements and mixed defenders. Policy budgets and incomplete evaluation
knowledge remain explicit; budget stops never become rules draws. Worldgorger's
prior finite recurrence scopes are unchanged.

## Live backlog reconciliation

Planning counted 452 open issues; the prior snapshot counted 453. The initial
live refresh had **456**; the pre-completion refresh had **457**, including only
the newly observed private-search blocker. The manifest records each issue's
number, original title, labels, assignees, update time, body hash and conservative
handling, plus both snapshot hashes. Other issue bodies/labels/owners stayed
unchanged. Existing active claims, open work, linked research/tests and reset
classifications were preserved; no reparenting, mass-closing, owner takeover,
speculative per-card issues or readiness promotion occurred. Label routing is
explicitly **not a semantic review inferred from titles**. Canonical blocker
bodies/comments and the actual missing operations were reviewed separately.

Only #3282 is eligible for this completion action. #3269 remains open for the
parent's final aggregate review. Closing this child removes its In Progress
label; it does not authorize work on any other issue.

## Observed effort, checks and performance

Three sequential five-card integrations, fifteen verified admissions, two new
whole-card blocks and one precise shared blocker. No per-card elapsed
measurements were collected here or in the old pilot; there is no fabricated
authoring-time or throughput estimate. The same existing `cmd/cardauthor`
scaffold was exercised for each initial family and its qualified substitute.
Scratch scaffolds are unregistered and never count as delivered cards.

Final integrated gate results are recorded on #3282 and in local ignored
`.cardwork/production-validation/`: full tests then lint sequentially, focused
race tests, actual Node-backed WASM execution, registry preservation,
support inventory and saved-manifest validation. At that checkpoint source
bindings were generated after bound authored source changes, and registry
preservation required zero losses and zero waivers.
Those retired mechanisms did not certify behavior.

All gates **passed**. Full `go test ./...` passed at initial integration; a late
Treasure self-reference declaration correction and its precise payment-adapter
prerequisite received scoped tests, narrow rules reviews and final
full-test/lint validation. Initial lint reported thirty
style diagnostics, corrected with scoped follow-up tests and final lint
**0 issues**. Focused race and actual WASM executed the new named scenarios,
prior pilot/mixed consumers and queue tests. Registry-only preservation reported
3,370 fingerprint changes and 40 admissions since its older baseline, **0 losses
and 0 waivers**; fifteen admissions belong to this change. No corpus compilation
or support-count waiver was run. The private-search regression is the explicit
blocked prerequisite described above, not a claimed passing gate scenario.

Fixed seed `20261014`, PCG `(seed, seed^7)`, Apple M3 Max / macOS arm64 /
Go 1.26.4, `GOMAXPROCS=4`, one worker, `-benchtime=10x -count=3 -benchmem`.
Below are medians of three current samples. Completion is **stable priority**,
not a complete multiplayer game, rules draw or arbitrary-game throughput claim.

| Production + Simulator scenario | ns/op | B/op | allocs/op | decisions |
| --- | ---: | ---: | ---: | ---: |
| Bedevil | 488,850 | 418,528 | 2,704 | 8 |
| Terminate | 249,038 | 414,888 | 2,478 | 8 |
| Vindicate | 468,658 | 426,544 | 2,687 | 8 |
| Disenchant | 437,417 | 419,968 | 2,596 | 8 |
| Naturalize | 440,750 | 421,248 | 2,598 | 8 |
| An Offer You Can't Refuse | 275,888 | 613,304 | 2,099 | 8 |
| Counterspell | 132,833 | 347,912 | 1,414 | 8 |
| Negate | 144,238 | 359,176 | 1,464 | 8 |
| Essence Scatter | 133,800 | 349,696 | 1,430 | 8 |
| Disallow | 139,738 | 361,320 | 1,452 | 8 |
| Eladamri's Call | 188,158 | 687,280 | 1,395 | 9 |
| Steelshaper's Gift | 195,250 | 706,577 | 1,393 | 9 |
| Fabricate | 226,258 | 728,401 | 1,461 | 9 |
| Idyllic Tutor | 187,138 | 679,744 | 1,459 | 9 |
| Sylvan Scrying | 180,800 | 685,512 | 1,400 | 9 |

Current retained mixed fixtures measured 618,038 ns / 1,149,350 B / 4,383 allocs
(Hive Mind + Bolt), 1,681,529 ns / 1,316,609 B / 8,787 allocs (Terastodon +
Collar), and 1,697,021 ns / 3,291,081 B / 8,501 allocs (Worldgorger +
Beast Within). Their 27/14/24 decisions remain stable. These are current
measurements, **not** a same-input pre-batch baseline comparison or proof of
zero regression. The [earlier full-game evidence](perf/mixed-production.md)
remains: Generic ~4.83s versus ~4.40s baseline with substantial allocation
regressions. Full-game benchmarks were not remeasured or relabeled as 1,000/s.
The table measures the final card declarations after the self-reference correction
but before the final two-line payment-source recognition fix; no benchmark was
rerun after that fix because the user requested this stopping point. The
earlier, faster samples are retained in `perf-pre-self-reference.log`, not
substituted to improve the reported timing.

```sh
GOMAXPROCS=4 go test ./mtg/rules -run '^$' -bench '^Benchmark(ProductionBatches|MixedProduction)$' -benchtime=10x -count=3 -benchmem
go run ./cmd/cardqueue -operation check
```

## Historical next-batch examples — not implemented in this assignment

All members remain **unclaimed**, with exact printing revisions/rulings in the
manifest. Their shared prerequisites are closed; fresh work must verify relevant
live ownership before claiming, not promote old compiler-expansion issues.

| Batch | Card | Oracle ID |
| --- | --- | --- |
| 04 draw | Divination | `273b339c-964b-4a18-8eb5-ceb8abcdfd9e` |
| 04 | Concentrate | `a3d287ec-1a39-4f91-acf1-7c1d2c00ed89` |
| 04 | Harmonize | `7eff84f1-f772-497a-b350-bbc93d0230f7` |
| 04 | Night's Whisper | `7ffae8f8-3006-4969-a339-6d30678f87ea` |
| 04 | Sign in Blood | `c6207f6a-a624-4754-88f5-dbe700c841ff` |
| 05 qualified destruction | Doom Blade | `59e7f2ae-4535-4191-98be-3e65b6b2befa` |
| 05 | Go for the Throat | `2f092562-9e17-43cd-aeb8-d0567f99363e` |
| 05 | Smother | `acdf9fb0-96eb-4520-bd4b-b6480c7c9b29` |
| 05 | Hero's Downfall | `03df6a57-37c9-46d3-83b3-4a6240100714` |
| 05 | Shatter | `84d32e68-72b9-432d-a023-7bc6a26b8ad0` |
| 06 qualified counter | Cancel | `7d00fb28-ea6c-49a9-b4af-ffb38860a9a7` |
| 06 | Disdainful Stroke | `11e02134-7b1a-46a4-a89e-7539dd1efada` |
| 06 | Remove Soul | `b13c0f76-fbda-4911-9442-c3d7e97f1aac` |
| 06 | Annul | `d08e9784-75f7-4164-ac48-d06160f8c56b` |
| 06 | Envelop | `30062dd0-c049-4872-8011-8b4810a3fa26` |

04 prerequisites: #3272, #3275, #3283; reuse Draw/LoseLife and individual draw
replacement competition, including Sign in Blood's target player.
05: #3273, #3283; reuse Destroy with excluded colors/types, mana-value filters,
cannot-regenerate and ordinary union targets.
06: #3272, #3283; reuse CounterObject and stack spell type/mana-value predicates.
These are executable declarations of already-established operations, not promises
of completed behavior; each new card still needs actual rules scenarios.

**Next highest-leverage shared prerequisite:** canonical
[#170](https://github.com/natefinch/council4/issues/170), its existing focused
buyback handoff: inspectable optional announcement/payment plus paid,
successful-resolution-only hand return, preserving ordinary destinations when
countered. Reiterate is the already-reviewed representative. There are **39**
remaining records with the corpus Buyback keyword; this is an observed candidate
family size, **not** a claim that all 39 will be unlocked or lack other abilities.
Private search #3289 is implemented by the two-card follow-up described above;
it does not expand the saved production queue or the original five-card batch.

## Current continuation

The [canonical guide](card-authoring.md#saved-production-queue) owns queue
selection, claims, behavioral evidence and stopping. Do not resume a historical
batch by its ID here or repeat this checkpoint's retired validation checklist.
