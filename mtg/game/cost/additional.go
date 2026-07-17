package cost

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AdditionalKind classifies a non-mana cost component.
type AdditionalKind int

// AdditionalDynamicAmount identifies a rules-derived amount for an additional
// cost whose value is not a fixed integer or the announced X. The rules engine
// resolves it against live game state while building the payment plan, so the
// payment vocabulary stays independent of the effect-resolution dynamic-amount
// machinery (which lives in package game and cannot be imported here).
type AdditionalDynamicAmount uint8

// PowerContributionKind identifies the characteristic used when a tap cost
// totals the selected permanents' power.
type PowerContributionKind uint8

// Power contribution kinds recognized by total-power tap costs.
const (
	PowerContributionEffective PowerContributionKind = iota
	PowerContributionCrew
)

// Additional dynamic amount kinds recognized by the payment planner.
const (
	AdditionalDynamicAmountNone AdditionalDynamicAmount = iota
	// AdditionalDynamicCommanderColorIdentityCount is the number of colors in
	// the paying player's commander's color identity (CR 903.4), backing
	// "Pay life equal to the number of colors in your commanders' color
	// identity" (War Room).
	AdditionalDynamicCommanderColorIdentityCount
	// AdditionalDynamicHandSize is the number of cards in the paying player's
	// hand, backing a "discard your hand" cost (Lion's Eye Diamond).
	AdditionalDynamicHandSize
	// AdditionalDynamicLifeGainedThisTurn is the amount of life the paying player
	// has gained so far this turn, backing a "pay X life, where X is the amount
	// of life you gained this turn" cost (Tivash, Gloom Summoner).
	AdditionalDynamicLifeGainedThisTurn
)

// SubtypeSet holds the one or two alternative subtypes supported by a card
// cost. Empty entries are ignored.
type SubtypeSet [2]types.Sub

// Additional cost kinds identify supported non-mana costs.
const (
	AdditionalUnknown AdditionalKind = iota
	AdditionalSacrifice
	AdditionalSacrificeSource
	AdditionalDiscard
	AdditionalPayLife
	AdditionalExile
	AdditionalReveal
	AdditionalTap
	AdditionalExileSource
	AdditionalUntap
	AdditionalRemoveCounter
	AdditionalReturnUnblockedAttacker
	AdditionalTapPermanents
	AdditionalEnergy
	AdditionalReturnToHand
	AdditionalExert
	AdditionalMill
	AdditionalPutCounter
	AdditionalCollectEvidence
	// AdditionalRemoveCounterAmong removes a total of Amount counters of
	// CounterKind spread across permanents the paying player controls that
	// match the cost's permanent constraint, as required by "remove N +1/+1
	// counters from among creatures you control." The payer chooses which
	// matching permanents to remove counters from.
	AdditionalRemoveCounterAmong
)

// Additional describes a typed non-mana cost printed on a spell, ability, or
// alternative cost. It is data only; mtg/rules chooses and pays it.
type Additional struct {
	Kind AdditionalKind

	// Text preserves the human-readable cost text for logs and diagnostics.
	Text string

	// Amount is the number of matching objects/cards or life points required.
	// Zero means one for object/card costs.
	Amount int

	// AmountFromX uses the announced X value as the required amount.
	AmountFromX bool

	// AmountAtLeastOne marks an AmountFromX cost whose announced count is the
	// player's "one or more" choice, so the payer must remove at least one, as
	// required by "Remove one or more +1/+1 counters from <this permanent>"
	// (Arcee, Sharpshooter). It restricts the activation's legal X values to
	// one or more; without it AmountFromX permits an announced X of zero.
	AmountAtLeastOne bool

	// AmountDynamic, when not AdditionalDynamicAmountNone, names a rules-derived
	// amount the payment planner resolves against live game state. It takes
	// precedence over Amount and AmountFromX for the cost's required count.
	AmountDynamic AdditionalDynamicAmount

	// MatchPermanentType constrains battlefield costs such as "sacrifice a
	// creature." When false, any permanent is allowed for permanent costs.
	MatchPermanentType bool
	PermanentType      types.Card

	// PermanentTypeAlt is an optional second permanent type accepted by a
	// battlefield cost printed as a two-type union, such as "sacrifice an
	// artifact or creature." It is honored only when MatchPermanentType is
	// true; an empty value constrains the cost to PermanentType alone.
	PermanentTypeAlt types.Card

	// ExcludePermanentType constrains a battlefield cost to permanents that are
	// not of the named card type, as required by "sacrifice ten nonland
	// permanents" (Bolas's Citadel). An empty value imposes no exclusion. It is
	// independent of MatchPermanentType.
	ExcludePermanentType types.Card

	// ExcludeSubtype constrains a battlefield cost to permanents that lack the
	// named subtype, as required by "return a non-Lair land you control to its
	// owner's hand" (the Lair cycle). An empty value imposes no exclusion.
	ExcludeSubtype types.Sub

	// MatchCardType constrains card costs such as "discard a creature card."
	// When false, any card in the relevant zone is allowed for card costs.
	MatchCardType bool
	CardType      types.Card

	// MatchHistoric constrains card costs (and battlefield permanent costs) to
	// historic objects, i.e. artifacts, legendaries, or Sagas (CR 702.61b), as
	// required by "exile any number of historic cards from your graveyard." It
	// is independent of MatchCardType.
	MatchHistoric bool

	// MatchCardColor constrains card costs to cards with the listed color, and
	// battlefield permanent costs (such as "sacrifice a black creature") to
	// permanents with the listed color.
	MatchCardColor bool
	CardColor      color.Color

	// SubtypesAny constrains card costs to cards with at least one listed
	// subtype. It is independent of MatchCardType and remains bounded so
	// Additional values are comparable.
	SubtypesAny SubtypeSet

	// Source identifies the zone cards are chosen from for card costs.
	// zone.None delegates to the rules-defined default for the cost kind.
	Source zone.Type

	// CounterKind identifies the counter removed from the source permanent by
	// an AdditionalRemoveCounter cost. It is ignored when AnyCounterKind is set.
	CounterKind counter.Kind

	// AnyCounterKind marks a generic counter-removal cost that removes counters
	// of any kind rather than a single named kind, as required by "remove N
	// counters from among <permanents> you control" (AdditionalRemoveCounterAmong)
	// and the bare "remove a counter from this permanent"
	// (AdditionalRemoveCounter). CounterKind is ignored when it is set; the
	// payment planner removes whatever counters the chosen permanents carry.
	AnyCounterKind bool

	// RequireTapped constrains battlefield costs to tapped permanents.
	RequireTapped bool

	// RequireUntapped constrains battlefield costs to untapped permanents, as
	// required by "return an untapped Plains you control to its owner's hand"
	// (the Karoo cycle). It parallels RequireTapped.
	RequireUntapped bool

	// RequireToken constrains a battlefield cost to token permanents (CR 111),
	// as required by "sacrifice an artifact token" or a bare "sacrifice a token."
	RequireToken bool
	// RequireNonToken constrains a battlefield cost to nontoken permanents.
	RequireNonToken bool

	// MatchArtifactEnchantmentOrToken constrains a battlefield permanent cost to
	// permanents that are an artifact, an enchantment, or a token of any type, as
	// required by the Bargain keyword's optional additional cost "sacrifice an
	// artifact, enchantment, or token" (CR 702.166a). Like MatchHistoric it is a
	// disjunctive named union expanded to a Selection by the payment planner, so
	// it stays a single comparable bool rather than a slice. It is independent of
	// MatchPermanentType and the token flags; an unset value imposes no union.
	MatchArtifactEnchantmentOrToken bool

	// RequireSupertype constrains battlefield costs to permanents with a
	// particular supertype, such as Snow.
	RequireSupertype types.Super

	// ExcludeSource constrains a battlefield cost to permanents other than the
	// paying ability's own source, as required by "another" (e.g. "Sacrifice
	// another creature").
	ExcludeSource bool

	// TotalPowerAtLeast, when positive, changes an AdditionalTapPermanents cost
	// from a fixed count to "tap any number of matching permanents with total
	// power N or more," as required by the Saddle keyword (CR 702.166). The
	// payer taps enough matching permanents to reach the threshold; Amount is
	// ignored when this is set.
	TotalPowerAtLeast int
	// PowerContribution selects the contribution rule used to total power.
	// Crew applies effects that let a creature crew Vehicles as though its power
	// were greater; the zero value uses ordinary effective power.
	PowerContribution PowerContributionKind

	// TotalManaValueAtLeast, when positive, changes an AdditionalExile cost from
	// a fixed count to "exile any number of matching cards with total mana value
	// N or greater" (The Capitoline Triad). The payer exiles enough matching
	// cards from the cost's Source zone for their mana values to total at least
	// N; Amount is ignored when this is set. It generalizes the collect-evidence
	// payment to arbitrary card filters such as MatchHistoric.
	TotalManaValueAtLeast int

	// ChoiceGroup tags this cost as one alternative within a numbered choice
	// group printed as "<cost> or <cost>" (e.g. "sacrifice an artifact or
	// discard a card"). Zero means a mandatory standalone cost. Costs sharing a
	// nonzero ChoiceGroup are alternatives; the payer pays exactly one member of
	// each group. It stays a scalar so Additional values remain comparable.
	ChoiceGroup uint8

	// Random, set only on an AdditionalDiscard cost, marks a "discard <count>
	// card(s) at random" cost (CR 701.9a). The payer discards randomly chosen
	// cards rather than cards of their choice, so the rules layer selects the
	// discarded cards uniformly at random instead of honoring a player choice.
	Random bool
}

// AdditionalChoiceOption is one payable branch of an AdditionalChoice. Exactly
// one branch of each choice is paid. Its Mana is additive: it is paid on top of
// the spell's printed cost rather than replacing it (unlike Alternative.ManaCost),
// so cost increases and reductions on the printed cost still apply to the branch.
type AdditionalChoiceOption struct {
	// Label names the branch for choice presentation and logs, e.g. "Pay 5
	// life" or "Pay {2}".
	Label string
	// Mana is the additive mana this branch pays in addition to the spell's
	// printed cost. It is nil when the branch pays no mana.
	Mana Mana
	// Costs are the non-mana additional costs this branch pays (pay life,
	// sacrifice, discard, ...). Members carry no ChoiceGroup; the branch itself
	// is the choice.
	Costs []Additional
}

// AdditionalChoice is a printed choice among alternative additional costs to
// cast a spell ("As an additional cost to cast this spell, pay 5 life or pay
// {2}." — Redirect Lightning). The caster pays exactly one Option in addition to
// the spell's printed cost and any mandatory additional costs. It differs from
// Alternative, whose mana replaces the printed cost: every Option's mana is
// additional, so a branch never discards the printed cost, and taxes and
// reductions on the printed cost apply to each branch.
type AdditionalChoice struct {
	Options []AdditionalChoiceOption
}

// Alternative describes an optional cost that replaces a spell or ability's
// normal mana cost when selected.
type Alternative struct {
	Label           string
	ManaCost        opt.V[Mana]
	AdditionalCosts []Additional
	Condition       AlternativeCondition
	// ConditionSubtype is the permanent subtype required by an
	// AlternativeConditionControlsPermanentSubtype condition, e.g. types.Swamp
	// for Snuff Out's "If you control a Swamp,". It is unused for every other
	// condition.
	ConditionSubtype types.Sub
	// ConditionCount is the threshold required by a count-based condition: the
	// attacking-creature threshold for AlternativeConditionCreaturesAttacking, the
	// on-battlefield permanent threshold for
	// AlternativeConditionPermanentsOnBattlefield, or the per-opponent
	// spells-cast-this-turn threshold for
	// AlternativeConditionOpponentCastSpellsThisTurn. ConditionExactly requires the
	// counted quantity to equal ConditionCount exactly ("If exactly one creature
	// is attacking,") rather than meet it as a minimum ("If N or more creatures
	// are attacking,"). Both are unused for every other condition.
	ConditionCount   int
	ConditionExactly bool
	// ConditionPermanentType is the permanent card type counted on the
	// battlefield by an AlternativeConditionPermanentsOnBattlefield condition,
	// e.g. types.Creature for Blasphemous Edict's "if there are thirteen or more
	// creatures on the battlefield." It is unused for every other condition.
	ConditionPermanentType types.Card
	// Mechanic identifies the rules mechanic this alternative grants, so the
	// rules layer decides Flashback/Escape/Evoke behavior from typed data
	// rather than the display Label. AlternativeMechanicNone leaves the
	// alternative's behavior fully described by its mana cost, additional
	// costs, and condition.
	Mechanic AlternativeMechanic
}

// AlternativeMechanic identifies the named rules mechanic an alternative cost
// grants. It lets the rules layer recognize graveyard-cast permissions and
// resolution riders (Flashback exile, Escape recast, Evoke sacrifice) from
// typed data instead of comparing the display Label.
type AlternativeMechanic uint8

// Supported alternative-cost mechanics.
const (
	// AlternativeMechanicNone marks an ordinary alternative cost (pitch,
	// discard, Spectacle, commander-free, Overload) whose behavior is fully
	// described by its costs and condition.
	AlternativeMechanicNone AlternativeMechanic = iota
	// AlternativeMechanicFlashback marks the Flashback graveyard cast
	// (CR 702.34) and the Jump-start cast (CR 702.134), which both grant the
	// graveyard Flashback permission and exile the spell on resolution.
	AlternativeMechanicFlashback
	// AlternativeMechanicEscape marks the Escape graveyard cast (CR 702.139),
	// which grants the graveyard Escape permission and does not exile the spell.
	AlternativeMechanicEscape
	// AlternativeMechanicEvoke marks the Evoke cast (CR 702.74), which lets the
	// resulting permanent be sacrificed by its evoke-sacrifice trigger.
	AlternativeMechanicEvoke
	// AlternativeMechanicMoreThanMeetsTheEye marks the Transformers "More Than
	// Meets the Eye" cast (CR 712): the front face is cast for the alternative
	// cost, and the resulting permanent enters the battlefield converted, as its
	// back face.
	AlternativeMechanicMoreThanMeetsTheEye
	// AlternativeMechanicDash marks the Dash cast (CR 702.109): the creature
	// spell is cast for its dash cost, the resulting permanent gains haste, and
	// it is returned to its owner's hand at the beginning of the next end step.
	AlternativeMechanicDash
)

// AlternativeCondition identifies a condition that must be true to select an
// alternative cost.
type AlternativeCondition uint8

// Supported alternative-cost conditions.
const (
	AlternativeConditionNone AlternativeCondition = iota
	AlternativeConditionControlsCommander
	// AlternativeConditionNotYourTurn requires that it is not the casting
	// player's turn, backing the Force of Negation pitch family.
	AlternativeConditionNotYourTurn
	// AlternativeConditionOpponentLostLifeThisTurn requires that an opponent of
	// the casting player has lost life so far this turn, backing the Spectacle
	// keyword (CR 702.107).
	AlternativeConditionOpponentLostLifeThisTurn
	// AlternativeConditionYourTurn requires that it is the casting player's turn,
	// backing free spells gated by "If it's your turn," (Mine Collapse).
	AlternativeConditionYourTurn
	// AlternativeConditionControlsPermanentSubtype requires that the casting
	// player controls a permanent with the alternative's ConditionSubtype,
	// backing free spells gated by "If you control a Swamp," (Snuff Out).
	AlternativeConditionControlsPermanentSubtype
	// AlternativeConditionOpponentGainedLifeThisTurn requires that an opponent of
	// the casting player has gained life so far this turn, backing the mana-only
	// alternative cost gated by "If an opponent gained life this turn," (Needlebite
	// Trap). It is the life-gain mirror of AlternativeConditionOpponentLostLifeThisTurn.
	AlternativeConditionOpponentGainedLifeThisTurn
	// AlternativeConditionCreaturesAttacking requires that the number of
	// creatures currently attacking meets the alternative's ConditionCount,
	// backing mana-only alternative costs gated by "If N or more creatures are
	// attacking," (Lethargy Trap, Arrow Volley Trap) or, when ConditionExactly is
	// set, "If exactly one creature is attacking," (Pitfall Trap).
	AlternativeConditionCreaturesAttacking
	// AlternativeConditionPermanentsOnBattlefield requires that the number of
	// permanents of the alternative's ConditionPermanentType currently on the
	// battlefield across all players meets the alternative's ConditionCount,
	// backing the mana-only alternative cost gated by "if there are thirteen or
	// more creatures on the battlefield" (Blasphemous Edict). When ConditionExactly
	// is set the count must equal ConditionCount exactly; otherwise it must meet it
	// as a minimum. The count reads current effective battlefield characteristics,
	// so tokens and animated/type-changed permanents are included.
	AlternativeConditionPermanentsOnBattlefield
	// AlternativeConditionOpponentCastSpellsThisTurn requires that a single
	// opponent of the casting player has cast at least ConditionCount spells so
	// far this turn, backing the mana-only alternative cost gated by "If an
	// opponent cast three or more spells this turn," (Mindbreak Trap). The
	// threshold is met per opponent, never by summing casts across opponents: the
	// condition holds only when some one opponent's own current-turn cast count
	// reaches ConditionCount. Only cast spells count (CR 601) — spell copies,
	// activated/triggered abilities, and played lands do not — while a spell that
	// was later countered still counts because it was cast. It is unused for every
	// other condition; ConditionCount carries the threshold.
	AlternativeConditionOpponentCastSpellsThisTurn
)
