package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PrimitiveKind identifies the variant of a Primitive.
type PrimitiveKind int

// PrimitiveKind values identify each supported primitive variant.
const (
	PrimitiveUnknown PrimitiveKind = iota
	PrimitiveDamage
	PrimitiveDraw
	PrimitiveDiscard
	PrimitiveDestroy
	PrimitiveAddMana
	PrimitiveAddCounter
	PrimitiveAddPlayerCounter
	PrimitiveMoveCounters
	PrimitiveApplyContinuous
	PrimitiveApplyRule
	PrimitiveModifyPT
	PrimitiveFight
	PrimitiveTap
	PrimitiveSearch
	PrimitiveReveal
	PrimitivePutOnBattlefield
	PrimitiveCreateToken
	PrimitiveShufflePermanentIntoLibrary
	PrimitiveStartEngines
	PrimitiveSetClassLevel
	PrimitiveMonstrosity
	PrimitiveDiscoverCards
	PrimitivePay
	PrimitiveChoose
	PrimitiveGainLife
	PrimitiveLoseLife
	PrimitiveExile
	PrimitiveBounce
	PrimitiveSacrifice
	PrimitiveUntap
	PrimitiveCounterObject
	PrimitiveMill
	PrimitiveScry
	PrimitiveSurveil
	PrimitiveInvestigate
	PrimitiveProliferate
	PrimitiveGoad
	PrimitiveRemoveCounter
	PrimitiveTransform
	PrimitivePhaseOut
	PrimitiveRegenerate
	PrimitiveSkipStep
	PrimitiveCreateEmblem
	PrimitiveCreateDelayedTrigger
	PrimitiveCreateReplacement
	PrimitivePreventDamage
	PrimitiveMoveCard
	PrimitiveGrantCastPermission
	PrimitiveExplore
	PrimitiveManifest
	PrimitiveSacrificePermanents
	PrimitiveSkipNextUntap
	PrimitiveDig
)

// primitiveKindCount is the number of supported primitive kinds.
const primitiveKindCount = int(PrimitiveDig) + 1

// PrimitiveKindCount exposes primitiveKindCount to packages that need fixed-size tables.
const PrimitiveKindCount = primitiveKindCount

// Primitive is a sealed data-only interface for a single effect building block.
// Only types in this package may implement it.
type Primitive interface {
	Kind() PrimitiveKind
	isPrimitive()
	instructionRefs() primitiveRefs
	validatePrimitive([]TargetSpec, bool) error
}

// primitiveRefs describes what keys a Primitive consumes and publishes
// (distinct from the Instruction envelope's PublishResult).
type primitiveRefs struct {
	consumesResults []ResultKey
	consumesChoices []ChoiceKey
	consumesLinked  []LinkedKey
	publishesChoice ChoiceKey
	publishesLinked LinkedKey
}

// Damage deals an amount of damage to a target.
type Damage struct {
	Amount           Quantity
	Recipient        DamageRecipient
	DamageSource     opt.V[ObjectReference]
	ResultAmountKind EffectResultAmountKind

	// Divided reports that the controller divides Amount as a fixed total among
	// the targets chosen for the recipient's target spec, allocating at least
	// one to each at resolution (CR 601.2d). It is valid only with an
	// any-target recipient that addresses a multi-target spec.
	Divided bool
}

// Draw draws cards for a referenced player, or for every player in a referenced
// group ("each player draws", "each opponent draws"). Exactly one of Player or
// PlayerGroup is set.
type Draw struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
}

// Discard causes a referenced player, or every player in a referenced group
// ("each player discards", "each opponent discards"), to discard cards. Exactly
// one of Player or PlayerGroup is set.
type Discard struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
}

// Destroy destroys one referenced permanent or every permanent in a referenced group.
type Destroy struct {
	Object ObjectReference
	Group  GroupReference
	// PreventRegeneration marks a destruction that can't be regenerated
	// ("Destroy target creature. It can't be regenerated."). Regeneration
	// shields cannot replace the destruction; indestructibility and shield
	// counters still apply.
	PreventRegeneration bool
}

// AddMana adds mana to the controller's pool.
type AddMana struct {
	Amount Quantity
	// ManaColor is the color of mana produced.
	ManaColor mana.Color
	// ChoiceFrom links a prior Choose{Choice: ResolutionChoiceMana} result
	// to determine the mana color dynamically.
	ChoiceFrom ChoiceKey
	// EntryChoiceFrom reads the mana color from a choice made as the source
	// permanent entered the battlefield (its Permanent.EntryChoices), such as
	// "{T}: Add one mana of the chosen color." Unlike ChoiceFrom, the choice is
	// not published within this instruction sequence; the rules engine seeds it
	// from the source permanent before resolving the ability.
	EntryChoiceFrom ChoiceKey
	// SpendRider, when present, tags each unit of mana produced by this
	// instruction with a one-shot delayed triggered ability that fires when
	// that specific mana is later spent on a qualifying spell (CR 106.12,
	// 603.2c). It models "When that mana is spent to cast ..." riders such as
	// Path of Ancestry. Producing the mana remains a mana ability (CR 605); the
	// rider itself uses the stack when it fires.
	SpendRider opt.V[ManaSpendRider]
}

// AddCounter places counters on a referenced permanent.
type AddCounter struct {
	Amount      Quantity
	Object      ObjectReference // single permanent; zero if Group is set
	Group       GroupReference  // every permanent in a group; zero if Object is set
	CounterKind counter.Kind
}

// AddPlayerCounter places counters on a referenced player.
type AddPlayerCounter struct {
	Amount      Quantity
	Player      PlayerReference
	CounterKind counter.Kind
}

// MoveCounters moves counters from a source to a target permanent.
type MoveCounters struct {
	Amount      Quantity
	Object      ObjectReference
	CounterKind counter.Kind
	Source      CounterSourceSpec
}

// ApplyContinuous applies continuous effects to a target (or globally).
// PublishLinked remembers the affected permanent for a later linked effect, such
// as a delayed "sacrifice it" trigger that must resolve the earlier target.
type ApplyContinuous struct {
	Object            opt.V[ObjectReference]
	ContinuousEffects []ContinuousEffect
	Duration          EffectDuration
	PublishLinked     LinkedKey
}

// ApplyRule creates rule effects for a target (or globally).
type ApplyRule struct {
	Object      opt.V[ObjectReference]
	RuleEffects []RuleEffect
	Duration    EffectDuration
}

// ModifyPT modifies a permanent's power and/or toughness.
type ModifyPT struct {
	Object         ObjectReference
	PowerDelta     Quantity
	ToughnessDelta Quantity
	Duration       EffectDuration
	PublishLinked  LinkedKey
}

// Fight makes two permanents fight each other.
type Fight struct {
	Object        ObjectReference
	RelatedObject ObjectReference
}

// Tap taps one referenced permanent or every permanent in a referenced group
// ("Tap all creatures your opponents control."). Exactly one of Object or Group
// is set.
type Tap struct {
	Object ObjectReference
	Group  GroupReference
}

// Search searches a player's library for cards matching spec. PublishLinked may
// retain the permanent created by an exact singular battlefield search.
type Search struct {
	Player        PlayerReference
	Spec          SearchSpec
	Amount        Quantity
	PublishLinked LinkedKey
}

// Reveal reveals cards from a player's zone and optionally links them.
type Reveal struct {
	Amount        Quantity
	Player        PlayerReference
	Recipient     opt.V[PlayerReference]
	PublishLinked LinkedKey
}

// PutOnBattlefield puts a card or linked object onto the battlefield.
// PublishLinked retains the fresh permanent created by a successful move.
type PutOnBattlefield struct {
	Source            BattlefieldSource
	Recipient         opt.V[PlayerReference]
	ContinuousEffects []ContinuousEffect
	EntryTapped       bool
	EntryCounters     []CounterPlacement
	PublishLinked     LinkedKey
}

// CreateToken creates one or more tokens. EntryTapped makes every created token
// enter the battlefield tapped, matching "Create a tapped ... token." wording.
// EntryAttacking puts every created token onto the battlefield already attacking
// (CR 508.4), matching "... token that's tapped and attacking." wording; it has
// effect only while the token's controller is the attacking player in an active
// combat and is otherwise ignored, leaving the token to enter normally.
type CreateToken struct {
	Amount         Quantity
	Source         TokenSource
	Recipient      opt.V[PlayerReference]
	EntryTapped    bool
	EntryAttacking bool
}

// ShufflePermanentIntoLibrary shuffles the referenced permanent into its owner's library.
type ShufflePermanentIntoLibrary struct {
	Object ObjectReference
}

// StartEngines starts engine effects for a player.
type StartEngines struct {
	Player PlayerReference
}

// SetClassLevel sets the class level of a referenced Class permanent.
type SetClassLevel struct {
	Object ObjectReference
	Amount Quantity
}

// Monstrosity makes a referenced creature monstrous.
type Monstrosity struct {
	Object ObjectReference
	Amount Quantity
}

// DiscoverCards performs a discover for N.
type DiscoverCards struct {
	Amount Quantity
}

// Pay prompts the controller to pay an optional cost during resolution.
// The instruction's Optional field controls whether declining is allowed.
// Results are published via the Instruction.PublishResult for downstream ResultGate checks.
type Pay struct {
	Payment ResolutionPayment
	Prompt  string
}

// Choose makes a resolution-time choice and publishes it via PublishChoice.
type Choose struct {
	Choice        ResolutionChoice
	PublishChoice ChoiceKey
}

// GainLife causes a referenced player or group of players to gain life.
// Exactly one of Player or PlayerGroup must be set.
type GainLife struct {
	Amount      Quantity
	Player      PlayerReference
	PlayerGroup PlayerGroupReference
}

// LoseLife causes a referenced player or group of players to lose life.
// Exactly one of Player or PlayerGroup must be set.
type LoseLife struct {
	Amount      Quantity
	Player      PlayerReference
	PlayerGroup PlayerGroupReference
}

// Exile exiles one referenced permanent or every permanent in a referenced group.
// ExileLinkedKey remembers the exiled object for later "exile it, then return it" patterns.
type Exile struct {
	Object         ObjectReference
	Group          GroupReference
	ExileLinkedKey LinkedKey
}

// Bounce returns one referenced permanent or every permanent in a referenced
// group to hand. When ControlledChoice is set, the resolving controller chooses
// Amount permanents from among the permanents matched by Group (its candidate
// pool, e.g. "permanents you control") and returns each to its owner's hand
// ("Return a creature you control to its owner's hand.").
type Bounce struct {
	Object ObjectReference
	Group  GroupReference

	// ControlledChoice has the resolving controller choose Amount permanents from
	// among those matched by Group. Object must be unset and Group set when it is
	// true; otherwise the whole Group (or single Object) is bounced.
	ControlledChoice bool
	Amount           Quantity
}

// MoveCard moves cards between two non-battlefield zones. It has two forms,
// distinguished by which reference is set (exactly one must be):
//
//   - Single-card form: Card references one card; that card moves from FromZone
//     to Destination ("Exile target card from a graveyard.").
//   - Player-zone group form: Player references a player; every card currently in
//     that player's FromZone moves to Destination at once, preserving ownership
//     ("Exile target player's graveyard."). An empty source zone is a legal
//     no-op.
type MoveCard struct {
	Card CardReference
	// Player selects the player whose entire FromZone is moved. It is set only
	// for the player-zone group form; Card must be unset when Player is set.
	Player            PlayerReference
	FromZone          zone.Type
	Destination       zone.Type
	DestinationBottom bool
}

// GrantCastPermission allows a referenced card to be cast from a specific zone
// using a specific face for a bounded duration.
type GrantCastPermission struct {
	Card     CardReference
	FromZone zone.Type
	Face     FaceIndex
	Duration EffectDuration
}

// Sacrifice sacrifices the referenced permanent. When no object is set, the
// controller's first permanent is used.
type Sacrifice struct {
	Object ObjectReference
}

// SacrificePermanents causes the referenced player (or every player in a group)
// to choose and sacrifice the required number of eligible permanents during resolution.
type SacrificePermanents struct {
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
	Amount      Quantity             // number of permanents to sacrifice
	Selection   Selection            // eligible permanent filter; zero = any permanent
}

// Untap untaps one referenced permanent or every permanent in a referenced group.
type Untap struct {
	Object ObjectReference
	Group  GroupReference
}

// SkipNextUntap marks the referenced permanent so it doesn't untap during its
// controller's next untap step (the "doesn't untap during its controller's next
// untap step" clause that follows a tap effect). The permanent stays tapped
// through one of its controller's untap steps and then untaps normally.
type SkipNextUntap struct {
	Object ObjectReference
}

// CounterObject counters a referenced spell or ability on the stack.
type CounterObject struct {
	Object ObjectReference
}

// Mill puts cards from the top of a referenced player's library into their
// graveyard, or does so for every player in a referenced group ("each player
// mills", "each opponent mills"). Exactly one of Player or PlayerGroup is set.
type Mill struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
}

// Scry looks at and reorders the top cards of a referenced player's library.
type Scry struct {
	Amount Quantity
	Player PlayerReference
}

// Surveil looks at the top cards of a referenced player's library, putting any into the
// graveyard.
type Surveil struct {
	Amount Quantity
	Player PlayerReference
}

// DigRemainder identifies where the unchosen cards of a Dig effect are placed.
type DigRemainder uint8

// Dig remainder destinations.
const (
	// DigRemainderGraveyard puts the unchosen cards into the player's graveyard.
	DigRemainderGraveyard DigRemainder = iota
	// DigRemainderLibraryBottom puts the unchosen cards on the bottom of the
	// player's library.
	DigRemainderLibraryBottom
)

// Dig looks at the top Look cards of a referenced player's library, lets that
// player put Take of those cards into their hand, and puts the remaining cards
// into the destination identified by Remainder. It models the impulse form that
// looks at the top N cards, puts some into your hand, and sends the rest to your
// graveyard or the bottom of your library.
type Dig struct {
	Player    PlayerReference
	Look      Quantity
	Take      Quantity
	Remainder DigRemainder
}

// Investigate creates Clue tokens for the recipient (controller by default).
type Investigate struct {
	Amount    Quantity
	Recipient opt.V[PlayerReference]
}

// Proliferate lets the controller add a counter of an existing kind to each
// chosen permanent or player.
type Proliferate struct {
	Amount Quantity
}

// Explore resolves the explore keyword action for a referenced creature.
type Explore struct {
	Creature ObjectReference
}

// Manifest puts cards from the controller's library onto the battlefield face down.
type Manifest struct {
	Dread bool
}

// Goad goads the referenced creature.
type Goad struct {
	Object ObjectReference
}

// RemoveCounter removes counters from one referenced permanent or every permanent in a referenced group.
type RemoveCounter struct {
	Amount      Quantity
	Object      ObjectReference
	Group       GroupReference
	CounterKind counter.Kind
}

// Transform transforms the referenced permanent.
type Transform struct {
	Object ObjectReference
}

// PhaseOut phases out the referenced permanent.
type PhaseOut struct {
	Object ObjectReference
}

// Regenerate sets up a regeneration shield on the referenced permanent.
type Regenerate struct {
	Object ObjectReference
}

// SkipStep schedules a referenced player to skip a step.
type SkipStep struct {
	Player PlayerReference
	Step   Step
}

// CreateEmblem creates an emblem owned by the controller with the given abilities.
type CreateEmblem struct {
	EmblemAbilities []Ability
}

// CreateDelayedTrigger schedules a delayed triggered ability.
type CreateDelayedTrigger struct {
	Trigger DelayedTriggerDef
}

// CreateReplacement creates a replacement effect that applies to a future event.
type CreateReplacement struct {
	Replacement *ReplacementEffect
	Duration    EffectDuration
}

// PreventDamage creates a damage-prevention shield for exactly one referenced
// player or permanent.
type PreventDamage struct {
	Amount Quantity
	Object ObjectReference
	Player PlayerReference
}
