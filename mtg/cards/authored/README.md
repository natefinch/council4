# Non-generated card implementations

Start with the [canonical authoring guide and executable scaffold](../../../docs/card-authoring.md)
and [card-authoring skill](../../../.agents/skills/card-authoring/SKILL.md).

This package owns authored Go definitions and versioned synchronous programs.
Add completed `cardset.Entry` values to `Cards`. The default registry includes
them explicitly; `genregistry` and the legacy compiler write only generated
letter packages and do not rewrite this directory.

Keep complete costs, targets, triggers and static/replacement declarations
inspectable. Use the [program context](../../cardprogram/README.md) for shared
Engine operations, not raw Game mutation. Preserve existing registrations; a
duplicate identity is an error unless a unique legacy replacement is explicitly
declared. Infrastructure and synthetic tests do not themselves promote a card.

Intuition and Realms Uncharted compose reveal-only `Search`, recorded `Ask`,
simultaneous `MoveCards`, and `ShuffleLibrary`. Intuition declares its targeted
opponent and explicitly requires three unrestricted finds (or all available
cards, CR 701.23d). Realms searches for up to four different-name lands and chooses
a living opponent during resolution, even if that opponent has hexproof or
shroud; revealing fewer than three puts all finds in the graveyard.
`mtg/rules/search_partition_spells_test.go` executes real casts, forced versus
fail-to-find choices, private disclosure, duplicate names, short/empty libraries,
all-illegal targets, retargeted copies, zone replacements, deferred replacement
shuffles, Opposition Agent, incarnation handles and clone isolation. Exact
Oracle identities and explicit scenario references are retained in each entry;
rules evidence is the pinned corpus and official CR effective 2026-09-25.

Twincast, Reverberate and Dualcaster Mage compose `Context.Copy` with inspectable
instant/sorcery targets; Dualcaster's Flash and entry trigger are ordinary
declarations. Cackling Counterpart and Fated Infatuation use `CreateToken` with
`TokenCopyOf`; flashback remains Engine-owned. Fated's turn check uses the
read-only `View().ActivePlayer()` query and shared `Scry`, which offers the
whole looked-at group and independent top/bottom ordering (CR 701.22).
`mtg/rules/authored_copy_spells_test.go` verifies casts, retargeting, modes/X,
copy-versus-cast events, departed sources/targets, copiable token values,
entry abilities/replacements, flashback exits, clones and scry timing/order.
Reiterate remains wholly unsupported: canonical [#170](https://github.com/natefinch/council4/issues/170)
owns buyback announcement/payment and its resolution-only hand replacement
(CR 702.27). A normal copy body alone is not complete support.

Blood Moon is an authored, program-free static declaration: battlefield
`Selection` excludes the Basic supertype and `LayerType` applies
`SetBasicLandTypes`. This shared CR 305.7 SET removes old land types and
intrinsic/copiable abilities, unlike Urborg's retained ADD operation. It preserves
other type families, supertypes, animation and external grants. Type-layer
source-erasure dependencies are discovered from declarations, never card names.
CR 613.6 means layer-6 ability loss does **not** stop a Blood Moon effect already
begun in layer 4; overwriting its ability in layer 1 or removing the source does.
Acceptance uses the Oracle corpus cached October 13, 2026 and the official
September 25, 2026 Comprehensive Rules. Copied Aura restrictions (#3227) remain
separate work.

Chains of Mephistopheles composes `Discard`, `Draw` and `MoveTopOfLibrary` inside
the existing individual draw replacement program. It branches on an actual
discard, including a redirected discard, not the size of the hand or the
intended graveyard destination. The Engine carries replacement lineage through
derived draws; multiple Chains, Abundance and dredge compete at each remaining
draw. Only the first actual draw of each own draw step is exempt.
`mtg/rules/chains_of_mephistopheles_test.go` verifies real casts, choices,
empty hands/libraries, extra draw steps, sibling lineage, source departure and
clones with the pinned Oracle corpus SHA-256 `6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4`.

Humility and Opalescence are program-free static declarations with one body
per multi-layer effect. The shared type and ability layers capture begun
recipient sets (CR 613.6), discover source-erasure/applicability dependencies
(613.8), and retain timestamp order for independent effects and dependency
loops. Base mana-value setters use the front mana cost of real transforming
permanents and full double-faced tokens, but not individual back-face copies
(CR 202.3b, 712.8e). Copied static abilities keep the copy effect's timestamp
without borrowing its blueprint's physical identity (613.7a).
Keyword counters participate in layer 6 with clone-safe placement timestamps
(613.7c), rather than unconditionally restoring lost abilities after layering.
`mtg/rules/humility_opalescence_test.go` verifies every published one-/two-
Opalescence timestamp order, physical-copy exclusions, Auras, mana abilities,
PT modifiers/counters, keyword-counter timestamp refresh, type dependencies,
replacement ability loss, phase-out/source departure and clones. Evidence is
the official CR effective 2026-09-25 and assigned Scryfall rulings retrieved
October 14, 2026; all three entries retain specific `Verification` references.
`mtg/rules/challenge_followup_test.go` executes Thalia's removed tax, Rhonas's
removed combat restrictions, later externally granted rule abilities, The Broken
Sky and its double-faced token copies, and old permanents newly copying
Opalescence. Printed haste and riot grants bypass sickness only while they
survive layer 6; control history remains independent for attacks, tap/untap
payments, automatic mana payment, control exchanges, turn starts and clones.
