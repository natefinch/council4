package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChoiceKind classifies an engine-mediated decision that is not a normal
// priority action.
type ChoiceKind int

// Choice kind values classify engine-mediated decisions.
const (
	ChoiceMay ChoiceKind = iota
	ChoiceTarget
	ChoiceOrder
	ChoicePayment
	ChoiceScry
	ChoiceSurveil
	ChoiceZoneSelection
	ChoiceSearch
	ChoiceModal
	ChoiceResolution
	ChoiceProliferate
	ChoicePlayer
	ChoiceExplore
	ChoiceManifest
	// ChoiceDamageAllocation asks the controller of a divided-damage effect to
	// split a fixed total among the chosen targets. Each option corresponds to
	// one chosen target; the returned selection lists option indices with
	// repetition, so the number of times an option appears equals the damage
	// allocated to that target. MinChoices and MaxChoices both equal the total
	// damage, and every target must receive at least one (CR 601.2d).
	ChoiceDamageAllocation
	// ChoiceCounterAllocation asks the controller of a distribute-counters effect
	// to split a fixed total of counters among the chosen targets. Each option
	// corresponds to one chosen target; the returned selection lists option
	// indices with repetition, so the number of times an option appears equals
	// the counters allocated to that target. MinChoices and MaxChoices both equal
	// the total counters, and every target must receive at least one. It mirrors
	// ChoiceDamageAllocation for the "Distribute N +1/+1 counters among ... target
	// creatures" placement.
	ChoiceCounterAllocation
	// ChoiceDig asks the controller of a Dig effect which of the cards revealed
	// from the top of their library to put into their hand. Each option
	// corresponds to one looked-at card; the returned selection lists the option
	// indices of the cards taken. MinChoices and MaxChoices both equal the number
	// of cards taken, bounded by the number of cards actually seen.
	ChoiceDig
	// ChoiceVote asks one player to cast their vote in a "Starting with you,
	// each player votes for <A> or <B>." voting interaction (CR 701.32). Each
	// option corresponds to one named choice; the returned selection is the
	// single option index the player votes for. MinChoices and MaxChoices both
	// equal one.
	ChoiceVote
	// ChoiceReplacement asks the affected object's controller or the affected
	// player which of several applicable replacement or prevention effects to
	// apply to an event (CR 616.1). Each option corresponds to one applicable
	// effect; the returned selection is the single option index chosen.
	// MinChoices and MaxChoices both equal one.
	ChoiceReplacement
	// ChoicePileSeparate asks the separating player of a PileSplit effect to
	// divide the revealed cards into two piles. Each option corresponds to one
	// revealed card; the returned selection is the indices of the cards placed
	// in the first pile (the rest form the second pile). MinChoices is zero and
	// MaxChoices equals the number of revealed cards, so either pile may be
	// empty.
	ChoicePileSeparate
	// ChoicePileChoose asks the choosing player of a PileSplit effect which of
	// the two piles is kept. There are exactly two options (the first and second
	// pile); the returned selection is the single option index of the kept pile.
	// MinChoices and MaxChoices both equal one.
	ChoicePileChoose
	// ChoiceManaCombination asks the recipient of an "add <N> mana in any
	// combination of <colors>" effect (Manamorphose, Goblin Clearcutter,
	// Cascading Cataracts) to split the produced mana freely among the offered
	// colors. Each option corresponds to one offered color; the returned
	// selection lists option indices with repetition, so the number of times an
	// option appears equals the mana produced of that color. MinChoices and
	// MaxChoices both equal the total amount. Unlike ChoiceDamageAllocation and
	// ChoiceCounterAllocation, a color may receive zero (RRR, RRG, RGG, and GGG
	// are all legal splits of three mana among {R} and {G}), so every unit is
	// distributed but no color is guaranteed a share.
	ChoiceManaCombination
	// ChoicePreventionAllocation selects a source for one-shot prevention or
	// assigns one point of finite prevention (CR 615.7-8).
	ChoicePreventionAllocation
	// ChoiceNumber returns one exact integer, not an index into Options.
	ChoiceNumber
	// ChoiceCombatDeclaration selects one extendable attack choice or block edge.
	ChoiceCombatDeclaration
	// ChoiceCombatDamageDivision asks a player to divide one creature's combat
	// damage among the creatures it is blocked by or blocking (CR 510.1c-d), as
	// when banding transfers that division to them (CR 702.22j-k). Each option
	// corresponds to one recipient creature; the returned selection lists option
	// indices with repetition, so the number of times an option appears equals
	// the damage assigned to that creature. MinChoices and MaxChoices both equal
	// the total damage, and a recipient may receive zero. CombatAttacker or
	// CombatBlocker identifies the creature whose damage is divided.
	ChoiceCombatDamageDivision
	// ChoiceAttackingBand asks the active player, after the attack targets are
	// fixed, whether to add another attacking creature to the band of the
	// creature identified by CombatAttacker (CR 508.1e, 702.22c-d). Option 0
	// finishes that band; each other option's Attack identifies one attacking
	// creature that can legally join it. A finished band of one creature is no
	// band, so the default declares no bands.
	ChoiceAttackingBand
)

// NumberRange describes a compact integer choice without enumerating its values.
type NumberRange struct {
	Minimum int
	Maximum opt.V[int]
}

// ChoiceCardInfo carries the public characteristics of a card or permanent that
// a ChoiceOption or ChoiceRequest refers to, so an agent can make a card-aware
// decision (which card to discard, where to scry a card) without looking the
// card up itself. Only public characteristics are included.
type ChoiceCardInfo struct {
	CardID    id.ID
	Name      string
	Types     []types.Card
	ManaValue int
	Colors    []color.Color
}

// ChoiceOption is one legal option in a ChoiceRequest.
type ChoiceOption struct {
	Index int
	Label string

	// Card carries the public characteristics of the card or permanent this
	// option selects, when the option represents one. It is unset for options
	// that are not a specific card (e.g. "top"/"bottom" or "Pay 2 life").
	Card opt.V[ChoiceCardInfo]

	// Targets are the game targets this option would select, when the option
	// represents a targeting decision (a triggered ability's target, a redirect,
	// or an opponent-to-target choice). It lets an agent resolve which permanent
	// or player each option affects and choose the best one — for example aiming
	// removal at the biggest threat — instead of taking the first option. It is
	// empty for non-targeting options.
	Targets []Target

	// SourcePermanent is the object ID of the battlefield permanent whose
	// permission this option uses, when options differ by source rather than by
	// what they select: which of two same-named permanents' once-each-turn
	// permissions a cast from exile or a land play spends. It is zero when the
	// option uses no permanent's permission.
	SourcePermanent id.ID

	// Attack and Block identify a proposed declaration, not a rules target.
	Attack opt.V[AttackDeclaration]
	Block  opt.V[BlockDeclaration]
}

// ScryDecision is public progress within one simultaneous scry, never card
// identities or the private order of cards kept on top or put on the bottom.
type ScryDecision struct {
	Player          PlayerID
	Accepted        bool
	PlacementChosen bool
	TopCount        int
	BottomCount     int
}

// ChoiceRequest describes a bounded decision the rules engine needs from a
// player while resolving rules procedures such as triggered abilities.
type ChoiceRequest struct {
	ID           int
	Kind         ChoiceKind
	Player       PlayerID
	DecisionKind DecisionKind
	Prompt       string
	Options      []ChoiceOption
	MinChoices   int
	MaxChoices   int

	// DefaultSelection is used by the rules engine when no agent supplies a
	// valid answer. It must contain option indices valid for this request.
	DefaultSelection []int

	// Provisional reports that DefaultSelection is the deciding player's own
	// earlier, provisional answer to this decision, such as the targets named
	// by a cast action before its Phyrexian payments were announced (CR
	// 601.2b-c). Strategies keep it unless they deliberately change it.
	Provisional bool

	// MaxTotalManaValue, when set, caps the combined mana value of the selected
	// options. A selection is valid only when the sum of its options' card mana
	// values does not exceed this cap. Options without card info contribute zero.
	MaxTotalManaValue opt.V[int]

	// Subject carries the public characteristics of the single card a decision
	// concerns, when there is one — for example the card being placed by a scry
	// or surveil prompt. It is unset for choices that are not about one specific
	// card.
	Subject opt.V[ChoiceCardInfo]

	// Private keeps option identities out of logs, not the deciding agent's request.
	Private bool
	// ScryDecisions exposes earlier public participation/placement decisions in
	// the same simultaneous scry (CR 101.4b), before library changes commit.
	ScryDecisions []ScryDecision
	// NumberRange is required for ChoiceNumber; Options must be empty.
	NumberRange opt.V[NumberRange]
	// CombatAttacker and CombatBlocker identify the current staged decision.
	// CombatBlocker is zero for an attack decision. For a combat damage
	// division, exactly one identifies the creature whose damage is divided.
	CombatAttacker id.ID
	CombatBlocker  id.ID
}

// ChoiceDecision records the selected option indices for a ChoiceRequest.
type ChoiceDecision struct {
	Request       ChoiceRequest
	DecisionMaker PlayerID
	Selected      []int
	UsedFallback  bool
}
