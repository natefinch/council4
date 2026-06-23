package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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
	PrimitiveTapOrUntap
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
	PrimitiveImpulseExile
	PrimitiveReorderLibraryTop
	PrimitiveShuffleLibrary
	PrimitiveExileFromHand
	PrimitiveLookAtLibraryTop
	PrimitivePutFromHand
	PrimitiveCastForFree
	PrimitiveReturnFromGraveyard
	PrimitivePlayerLosesGame
	PrimitiveAttach
	PrimitiveMoveCommander
	PrimitivePutPermanentOnLibrary
	PrimitiveChooseNewTargets
	PrimitiveGroupSourceDamage
	PrimitiveMassReturnFromGraveyard
	PrimitivePlayerWinsGame
	PrimitivePunisherEachLoseLife
	PrimitiveMassReanimationExchange
	PrimitiveRepeatProcess
	PrimitiveCopyStackObject
	PrimitiveBecomeCopy
	PrimitiveAmass
	PrimitiveRenown
	PrimitiveShuffleSpellIntoLibrary
	PrimitiveExileTopOfLibrary
	PrimitivePutHandOnLibraryThenDraw
	PrimitiveRevealUntil
	PrimitiveBecomeSaddled
	PrimitiveAddExtraPhases
	PrimitiveLookAtHand
	PrimitiveRollDie
	PrimitiveRemoveFromCombat
	PrimitiveChooseDiscardFromHand
	PrimitiveExileFromGraveyard
)

// primitiveKindCount is the number of supported primitive kinds.
const primitiveKindCount = int(PrimitiveExileFromGraveyard) + 1

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

// GroupSourceDamage has each permanent in a battlefield group deal an amount of
// damage to its own controller, or its owner when ToOwner is set. It models
// "Each creature deals 1 damage to its controller.": every group member is the
// damage source and the recipient is the player who controls (or owns) that
// member.
type GroupSourceDamage struct {
	Group   GroupReference
	Amount  Quantity
	ToOwner bool
}

// Draw draws cards for a referenced player, or for every player in a referenced
// group ("each player draws", "each opponent draws"). Exactly one of Player or
// PlayerGroup is set.
type Draw struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
}

// ReorderLibraryTop has a player look at up to Amount cards from the top of
// their library and put those exact cards back in a chosen top-first order.
type ReorderLibraryTop struct {
	Amount Quantity
	Player PlayerReference
}

// LookAtLibraryTop privately shows the top card of a player's library to that
// player and links the exact card for later instructions. It does not reveal the
// card or move it.
type LookAtLibraryTop struct {
	Player        PlayerReference
	PublishLinked LinkedKey
}

// ShuffleLibrary randomizes a referenced player's library.
type ShuffleLibrary struct {
	Player PlayerReference
}

// LookAtHand lets the source's controller privately look at a referenced
// player's hand. It conveys hidden information only and does not change game
// state (CR 701.x look effects).
type LookAtHand struct {
	Player PlayerReference
}

// ChooseDiscardFromHand makes the resolving spell's controller choose a card
// from a referenced player's revealed hand, which that player then discards
// (the Duress / Thoughtseize / Coercion targeted-discard family). The chooser
// is always the controller; Player names the discarding player. ExcludeCreature
// and ExcludeLand restrict the eligible cards ("noncreature card" /
// "nonland card"), and MaxManaValue, when set, bounds the chosen card's mana
// value ("with mana value N or less", Inquisition of Kozilek).
//
// Selection further restricts the eligible cards to those matching a typed card
// filter ("a creature card", "a land card", "a nonland card"). It is the
// general filter used by the controller's own filtered self-discard ("you may
// discard a creature card") and composes with the exclude flags above: a card
// must satisfy both. The zero Selection imposes no constraint.
type ChooseDiscardFromHand struct {
	Player          PlayerReference
	ExcludeCreature bool
	ExcludeLand     bool
	MaxManaValue    opt.V[int]
	Selection       Selection
}

// Discard causes a referenced player, or every player in a referenced group
// ("each player discards", "each opponent discards"), to discard cards. A
// single referenced player chooses exactly Amount distinct cards when available,
// or every available card when fewer remain. Exactly one of Player or
// PlayerGroup is set.
//
// EntireHand marks a "discard their hand" effect ("Each player discards their
// hand", "Discard your hand"): the affected player discards every card in hand
// and Amount is ignored.
//
// AtRandom marks an "at random" discard ("Discard a card at random."): the
// discarded cards are chosen at random rather than by the player.
type Discard struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
	EntireHand  bool
	AtRandom    bool
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
	// instruction with exact spend-linked semantics. It models triggered riders
	// such as Path of Ancestry and restricted spell effects such as Cavern of
	// Souls while preserving the producing mana ability (CR 605).
	SpendRider opt.V[ManaSpendRider]
	// Player, when present, overrides the recipient of the produced mana. It
	// models triggered abilities that add mana to a referenced object's
	// controller ("its controller adds an additional {G}", Wild Growth) rather
	// than the ability's controller. When absent, mana goes to the controller.
	Player opt.V[PlayerReference]
	// EachControlledColor, when non-nil, makes this instruction produce Amount
	// mana of EACH color among the permanents the recipient controls matching
	// the Selection, rather than a single color ("For each color among
	// permanents you control, add one mana of that color", Bloom Tender). The
	// colors are recomputed at resolution as the union of the matching
	// permanents' colors; an empty set produces no mana (CR 202.2, 605).
	EachControlledColor *Selection
}

// AddCounter places counters on a referenced permanent.
type AddCounter struct {
	Amount      Quantity
	Object      ObjectReference // single permanent; zero if Group is set
	Group       GroupReference  // every permanent in a group; zero if Object is set
	CounterKind counter.Kind
	// AllKinds doubles every kind of counter already on Object: the runtime adds,
	// for each counter kind present, that many more, ignoring Amount and
	// CounterKind. It backs "double the number of each kind of counter on
	// <permanent>" (Vorel of the Hull Clade) and is set only with a single
	// Object, never a Group.
	AllKinds bool
}

// AddPlayerCounter places counters on a referenced player or group of players.
// Exactly one of Player or PlayerGroup must be set.
type AddPlayerCounter struct {
	Amount      Quantity
	Player      PlayerReference
	PlayerGroup PlayerGroupReference
	CounterKind counter.Kind
}

// MoveCounters moves counters from a source to a target permanent. When AllKinds
// is set every counter on the source moves regardless of kind and Amount is
// ignored ("Move all counters from this permanent onto target creature.");
// otherwise only Amount counters of CounterKind move ("Move a +1/+1 counter from
// this creature onto target creature."). When Distribute is set the counters are
// not moved to a single Object but distributed by the controller among the
// permanents of Group, one counter at a time, until the controller stops or the
// source runs out ("move any number of +1/+1 counters from this creature onto
// other creatures.").
type MoveCounters struct {
	Amount      Quantity
	Object      ObjectReference
	CounterKind counter.Kind
	Source      CounterSourceSpec
	AllKinds    bool
	// Group is the destination group of a distributed move ("move any number of
	// counters ... onto other creatures"); the controller distributes the
	// source's counters among its members. It is nil for the single-target forms.
	// It is held by pointer so the embedded GroupReference does not inflate the
	// heavily value-passed MoveCounters past the by-value size budget.
	Group      *GroupReference
	Distribute bool
	// ChooseKind moves one counter of a single kind the controller chooses among
	// the kinds present on the source ("Move a counter from target permanent you
	// control onto a second target permanent."). Amount counters of the chosen
	// kind move; CounterKind and AllKinds are ignored. It is false for a
	// named-kind move and the kind-agnostic AllKinds move.
	ChooseKind bool
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

// TapOrUntap lets the controller choose to tap or untap the referenced
// permanent ("Tap or untap target creature."). The choice is made when the
// instruction resolves.
type TapOrUntap struct {
	Object ObjectReference
}

// Search searches a player's library for cards matching spec. PublishLinked may
// retain the permanent created by an exact singular battlefield search. When
// Controller is set, a found card put onto the battlefield enters under that
// player's control instead of the searching player's ("put it onto the
// battlefield ... under target player's control", Yavimaya Dryad).
// Search has one referenced player, or every player in a referenced group
// ("each player searches their library"), search a library. Exactly one of
// Player or PlayerGroup is set. When PlayerGroup is set every member searches
// their own library and any found permanent enters under that searcher's
// control, so Controller must be unset.
type Search struct {
	Player        PlayerReference
	PlayerGroup   PlayerGroupReference
	Spec          SearchSpec
	Amount        Quantity
	Controller    opt.V[PlayerReference]
	PublishLinked LinkedKey
}

// Reveal reveals cards from a player's zone and optionally links them.
type Reveal struct {
	Amount        Quantity
	Player        PlayerReference
	Recipient     opt.V[PlayerReference]
	PublishLinked LinkedKey
	Card          CardReference
}

// PutOnBattlefield puts a card or linked object onto the battlefield. Sources
// moves multiple referenced cards simultaneously; exactly one of Source or
// Sources must be set.
// PublishLinked retains the fresh permanent created by a successful move.
type PutOnBattlefield struct {
	Source            BattlefieldSource
	Sources           []BattlefieldSource
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

	// Power and Toughness, when set, override the source definition's printed
	// power and toughness with a dynamic amount evaluated once at creation
	// ("create an X/X ... token, where X is the amount of life you gained this
	// turn."). Both are set together; the handler clones the token definition and
	// fixes its power and toughness to the resolved value so the token keeps that
	// size for its lifetime. They are unset for tokens with a printed
	// power/toughness.
	Power     opt.V[Quantity]
	Toughness opt.V[Quantity]

	// PublishLinked, when set, remembers each created token as an object-scoped
	// linked object so a later instruction can reference it ("create a 0/0 black
	// Phyrexian Germ creature token, then attach this Equipment to it.", Living
	// weapon). It is unused when empty.
	PublishLinked LinkedKey
}

// ShufflePermanentIntoLibrary shuffles the referenced permanent into its owner's library.
type ShufflePermanentIntoLibrary struct {
	Object ObjectReference
}

// ShuffleSpellIntoLibrary shuffles the resolving source spell into its owner's
// library instead of putting it into the graveyard. It backs the "Shuffle this
// card into its owner's library." resolution tail (Green Sun's Zenith, the
// Beacon cycle, Blue Sun's Zenith). Like Exile with SourceSpell, it has no
// referent of its own: it always acts on the spell currently resolving.
type ShuffleSpellIntoLibrary struct{}

// PutPermanentOnLibrary moves the referenced permanent from the battlefield to
// the top of its owner's library, or to the bottom when Bottom is set. It backs
// "put this [permanent] on top of its owner's library" (Sensei's Divining Top)
// and the corresponding bottom wording, without shuffling.
type PutPermanentOnLibrary struct {
	Object ObjectReference
	Bottom bool
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

// Amass performs the amass keyword action (CR 701.44): the controller puts
// Amount +1/+1 counters on an Army they control, first creating a 0/0 black
// Army creature token of Subtype if they control none.
type Amass struct {
	Amount  Quantity
	Subtype types.Sub
}

// Renown performs the renown keyword action (CR 702.111): if the referenced
// permanent is not already renowned, the controller puts Amount +1/+1 counters
// on it and it becomes renowned. A renowned permanent is left unchanged, so the
// effect applies at most once.
type Renown struct {
	Object ObjectReference
	Amount Quantity
}

// BecomeSaddled performs the Saddle keyword action (CR 702.166): the referenced
// Mount becomes saddled until end of turn. The saddled state is cleared during
// cleanup. The effect is idempotent; saddling an already-saddled Mount leaves it
// unchanged.
type BecomeSaddled struct {
	Object ObjectReference
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

// PlayerLosesGame causes a referenced player to lose the game (CR 104.3a). The
// player is marked to lose; state-based actions remove them the next time they
// are checked.
type PlayerLosesGame struct {
	Player PlayerReference
}

// PlayerWinsGame causes a referenced player to win the game (CR 104.2a). A
// player winning a two-or-more-player game means every other player loses, so
// each other still-active player is marked to lose; state-based actions remove
// them the next time they are checked, leaving the referenced player as the
// last one standing.
type PlayerWinsGame struct {
	Player PlayerReference
}

// Exile exiles one referenced permanent, every permanent in a referenced group,
// or the resolving source spell.
// ExileLinkedKey remembers the exiled object for later "exile it, then return it" patterns.
type Exile struct {
	Object         ObjectReference
	Group          GroupReference
	SourceSpell    bool
	ExileLinkedKey LinkedKey
}

// ExileFromHand has Player choose Amount cards from their hand that match
// Selection and exiles them, modelling "exile a ... card from your hand." The
// enclosing Instruction's Optional flag expresses the "you may" wrapper. When
// PublishLinked is set, each exiled card is remembered as an object-scoped
// linked object on the source permanent (imprint) so a later ability can read
// it; the link follows the permanent's object identity, so a re-entered object
// starts without an imprint. Fewer matching cards than Amount exiles all of
// them; no matching card exiles nothing.
type ExileFromHand struct {
	Player        PlayerReference
	Selection     Selection
	Amount        Quantity
	PublishLinked LinkedKey
}

// ExileFromGraveyard has Player choose up to Amount cards from their own
// graveyard that match Selection and exiles each, modeling the non-target
// graveyard wording "(you may) exile a <filter> card from your graveyard"
// (Masked Vandal, the Imoen cycle, Aphemia, ...). The targeted form ("exile
// target ... card from your graveyard") lowers to a card target instead; this
// primitive covers the choose-at-resolution form where the exiled card is
// selected rather than targeted. The enclosing Instruction's Optional flag
// expresses the "you may" wrapper, so the engine gathers consent before this
// runs; here the player chooses which matching card to exile, if any. Fewer
// matching cards than Amount exiles all of them; no matching card exiles
// nothing.
type ExileFromGraveyard struct {
	Player    PlayerReference
	Selection Selection
	Amount    Quantity
}

// PutFromHand has Player choose up to Amount cards from their hand that match
// Selection and puts each onto the battlefield under that player's control,
// modeling "put a land card from your hand onto the battlefield" and similar
// cheat-into-play / ramp effects. The enclosing Instruction's Optional flag
// expresses a "you may" wrapper, so the engine gathers consent before this runs;
// here the player chooses which matching card to put, if any. EntersTapped makes
// each card enter the battlefield tapped. Fewer matching cards than Amount puts
// all of them; no matching card puts nothing.
type PutFromHand struct {
	Player       PlayerReference
	Selection    Selection
	Amount       Quantity
	EntersTapped bool
}

// CastForFree has Player cast one card matching Selection from Zone without
// paying its mana cost, modeling "(You may) cast a spell [with mana value N or
// less] from your hand without paying its mana cost." and similar free-cast
// effects. The enclosing Instruction's Optional flag expresses a "you may"
// wrapper, so the engine gathers consent before this runs; here the player
// chooses which eligible card to cast, if any. No eligible card casts nothing.
type CastForFree struct {
	Player    PlayerReference
	Selection Selection
	Zone      zone.Type
}

// ReturnFromGraveyard has Player choose up to Amount cards from their graveyard
// that match Selection and returns each to their hand, modeling the non-target
// graveyard recursion wording "Return a <filter> card from your graveyard to
// your hand" (Takenuma's "creature or planeswalker card", Grapple with the
// Past, ...). The targeted form ("Return target creature card ...") lowers to a
// card target instead; this primitive covers the choose-at-resolution form
// where the returned card is selected rather than targeted. Fewer matching
// cards than Amount returns all of them; no matching card returns nothing.
//
// Destination selects where the chosen cards go: zone.None or zone.Hand returns
// each to its owner's hand, while zone.Battlefield reanimates each onto the
// battlefield under Player's control ("... to the battlefield", Tayam), tapped
// when EntryTapped is set.
//
// MaxTotalManaValue, when set, caps the combined mana value of the chosen cards
// ("Return up to two creature cards with total mana value 4 or less from your
// graveyard to the battlefield" — Lively Dirge). Player may choose any subset of
// matching cards whose total mana value does not exceed the cap, up to Amount
// cards; an empty choice is always legal, so the cap also makes the choice
// optional ("up to").
type ReturnFromGraveyard struct {
	Player            PlayerReference
	Selection         Selection
	Amount            Quantity
	Destination       zone.Type
	EntryTapped       bool
	MaxTotalManaValue opt.V[int]

	// AnyNumber models the "put any number of <filter> cards from among them
	// onto the battlefield" wording: the resolving player chooses any subset of
	// the matching candidate pool, from none up to all of them, rather than a
	// fixed count. Amount is ignored (and must be zero) when it is set, since the
	// upper bound is the whole matching pool. It pairs naturally with FromLinked
	// to put any number of a specific earlier-produced set (such as milled
	// cards) onto the battlefield.
	AnyNumber bool

	// FromLinked, when set, restricts the candidate pool to the cards remembered
	// under this key by a prior instruction (such as a Mill that published the
	// cards it milled). Only graveyard cards whose identity was linked this way
	// are eligible, modeling "put a card from among those cards into your hand"
	// where "those cards" denotes a specific earlier-produced set rather than the
	// whole graveyard. When empty, the whole graveyard is scanned as usual.
	FromLinked LinkedKey
}

// MassReturnFromGraveyard returns every card in Player's graveyard matching
// Selection to Destination at once, modeling the non-target mass recursion
// wording "Return all <filter> cards from your graveyard to the battlefield"
// (Brilliant Restoration) or "... to your hand". Unlike ReturnFromGraveyard,
// the resolving player makes no choice: all matching cards move. Destination is
// either zone.Hand (each card returns to its owner's hand) or zone.Battlefield
// (each card enters under Player's control, tapped when EntryTapped is set). An
// empty or fully unmatched graveyard is a legal no-op.
//
// SourceGroup widens the scanned graveyards beyond Player's own: when its Kind
// is not None, every matching card in each member player's graveyard moves at
// once ("... from all graveyards", Rise of the Dark Realms, Open the Vaults).
// When ControlledByOwner is set, each card entering the battlefield does so
// under its own owner's control rather than Player's ("... under their owners'
// control").
type MassReturnFromGraveyard struct {
	Player            PlayerReference
	Selection         Selection
	Destination       zone.Type
	EntryTapped       bool
	SourceGroup       PlayerGroupReference
	ControlledByOwner bool
}

// MassReanimationExchange resolves the symmetric mass-reanimation exchange "Each
// player exiles all <type> cards from their graveyard, then sacrifices all
// <type> they control, then puts all cards they exiled this way onto the
// battlefield." (Living Death, Living End, Scrap Mastery). For every player at
// once it (1) exiles each graveyard card matching Selection, (2) sacrifices each
// battlefield permanent matching Selection, then (3) returns the cards exiled in
// step 1 to the battlefield under their owners' control. Exiling before
// sacrificing keeps the freshly sacrificed permanents out of the returned set,
// realizing the "cards they exiled this way" back-reference without tracking it
// across separate effects. Selection carries only the card-type filter (creature
// or artifact); it never narrows by controller.
type MassReanimationExchange struct {
	Selection Selection
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
//   - Player-zone group form: Player references a player. With zero Amount, every
//     card currently in that player's FromZone moves to Destination at once
//     ("Exile target player's graveyard."). With positive Amount, that player
//     chooses up to Amount cards from their hand and orders them on top of their
//     library; the first selected card becomes the top card. An empty source zone
//     is a legal no-op.
type MoveCard struct {
	Card CardReference
	// Player selects the player whose entire FromZone is moved. It is set only
	// for the player-zone group form; Card must be unset when Player is set.
	Player PlayerReference
	// PlayerGroup selects every player whose entire FromZone is moved at once
	// ("Exile all graveyards."). It is set only for the player-group zone form;
	// Card and Player must be unset when PlayerGroup is set.
	PlayerGroup       PlayerGroupReference
	Amount            Quantity
	FromZone          zone.Type
	Destination       zone.Type
	DestinationBottom bool
}

// MoveCommander moves Player's commander(s) from the command zone to
// Destination, modeling "Put your commander into your hand from the command
// zone." (Command Beacon, Road of Return, Netherborn Altar). Only the player's
// own commander cards currently in their command zone move; other command-zone
// objects are left in place. The commander-replacement effect (CR 903.9) does
// not redirect the move, because the effect explicitly relocates the commander.
type MoveCommander struct {
	Player      PlayerReference
	Destination zone.Type
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
	// Fallback is applied to each affected player who controls no permanent
	// matching Selection, i.e. who can't satisfy the edict ("Each player who
	// can't discards a card."). SacrificeFallbackNone leaves no rider.
	Fallback SacrificeFallback
	// PublishLinked, when set, records the permanents sacrificed by this edict as
	// linked objects under the given key so a later instruction can read them
	// through last-known information once they have left the battlefield. It
	// backs an optional resolving sacrifice whose follow-up effect is scaled by
	// the sacrificed permanent ("you may sacrifice another creature. If you do,
	// you gain X life and draw X cards, where X is that creature's power." —
	// Disciple of Freyalise). Empty when no downstream effect reads the
	// sacrificed permanent.
	PublishLinked LinkedKey
}

// SacrificeFallbackKind identifies the per-player rider applied to players who
// can't satisfy a SacrificePermanents edict.
type SacrificeFallbackKind uint8

const (
	// SacrificeFallbackNone marks an edict with no who-can't rider.
	SacrificeFallbackNone SacrificeFallbackKind = iota
	// SacrificeFallbackDiscard makes each player who can't sacrifice discard
	// Amount cards ("Each player who can't discards a card.").
	SacrificeFallbackDiscard
	// SacrificeFallbackLoseLife makes each player who can't sacrifice lose
	// Amount life ("Each player who can't loses 2 life.").
	SacrificeFallbackLoseLife
)

// SacrificeFallback is the per-player rider applied to each player who can't
// satisfy a SacrificePermanents edict.
type SacrificeFallback struct {
	Kind   SacrificeFallbackKind
	Amount Quantity
}

// PunisherEachLoseLife makes each player in PlayerGroup lose Amount life unless
// that player chooses, as the effect resolves, to pay an offered alternative
// instead: sacrifice a permanent matching SacrificeSelection (when
// AllowSacrifice is set) or discard a card (when AllowDiscard is set). This is
// the "punisher" family ("Each opponent loses N life unless that player
// sacrifices a nonland permanent of their choice or discards a card." — Torment
// of Hailfire, Hag of Ceaseless Torment). Each affected player decides
// independently in APNAP order; a player who can perform no offered alternative
// simply loses the life.
type PunisherEachLoseLife struct {
	PlayerGroup        PlayerGroupReference
	Amount             Quantity
	AllowSacrifice     bool
	SacrificeSelection Selection
	AllowDiscard       bool
}

// RepeatProcess resolves Body a number of times equal to Times ("Repeat the
// following process X times. <body>" — Torment of Hailfire), where X is the
// resolving spell's chosen value of {X}. Body is a non-modal AbilityContent
// holding the repeated sub-effect; it is re-resolved from scratch on each
// iteration so any per-player or random choices recur independently. A Times of
// zero or fewer resolves Body no times.
type RepeatProcess struct {
	Times Quantity
	Body  AbilityContent
}

// Untap untaps one referenced permanent or permanents in a referenced group.
// ChooseUpTo has the resolving controller choose up to Amount distinct
// permanents from Group instead of untapping the whole group.
type Untap struct {
	Object ObjectReference
	Group  GroupReference

	ChooseUpTo bool
	Amount     Quantity
}

// SkipNextUntap marks the referenced permanent so it doesn't untap during its
// controller's next untap step (the "doesn't untap during its controller's next
// untap step" clause that follows a tap effect). The permanent stays tapped
// through one of its controller's untap steps and then untaps normally.
type SkipNextUntap struct {
	Object ObjectReference
}

// RemoveFromCombat removes the referenced creature from combat ("Remove target
// attacking creature you control from combat." — Reconnaissance). The permanent
// stops being an attacker or blocker: it deals and is dealt no further combat
// damage and its attack/block declarations are discarded. Object references the
// creature to remove.
type RemoveFromCombat struct {
	Object ObjectReference
}

// CounterObject counters a referenced spell or ability on the stack. When
// ExileInstead is set, a countered spell is exiled instead of being put into
// its owner's graveyard (CR 614-style replacement, e.g. Force of Negation).
type CounterObject struct {
	Object       ObjectReference
	ExileInstead bool
}

// ChooseNewTargets re-chooses the targets of a referenced spell or ability on
// the stack ("You may choose new targets for target spell or ability."). The
// resolving controller selects a new legal target for each of the referenced
// object's target specs; the choice is bounded by that object's own targeting
// restrictions (CR 115.7). The enclosing Instruction's Optional flag expresses
// the "you may" wrapper.
type ChooseNewTargets struct {
	Object ObjectReference
}

// CopyStackObject copies a targeted activated or triggered ability on the stack
// ("Copy target triggered ability you control."). The copy is put on the stack
// (CR 707.10) and resolves independently; it is not a card. When
// MayChooseNewTargets is set, the resolving controller may re-choose the copy's
// targets, bounded by the copied ability's own targeting restrictions (CR
// 707.12). Object references the targeted ability to copy.
type CopyStackObject struct {
	Object              ObjectReference
	MayChooseNewTargets bool
}

// Mill puts cards from the top of a referenced player's library into their
// graveyard, or does so for every player in a referenced group ("each player
// mills", "each opponent mills"). Exactly one of Player or PlayerGroup is set.
// Mill moves the top Amount cards of a referenced player's library to that
// player's graveyard.
type Mill struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set

	// PublishLinked, when set, remembers every card milled this way as a
	// card-scoped linked object on the source permanent so a later instruction
	// can act on exactly those cards ("mill three cards. ... put a card from
	// among those cards into your hand"). It is meaningful only for the single
	// Player form; the group form publishes nothing.
	PublishLinked LinkedKey
}

// ExileTopOfLibrary moves the top Amount cards of a referenced player's library
// to exile.
type ExileTopOfLibrary struct {
	Amount      Quantity
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
}

// RevealUntil reveals cards from the top of a referenced player's library one at
// a time until a revealed card matches Until, then puts every card revealed this
// way (including the matching card) into Destination. It models the closed
// "reveals cards from the top of their library until they reveal a <type> card,
// then puts those cards into their <zone>" family. Exactly one of Player or
// PlayerGroup is set; the group form runs the reveal for every member in APNAP
// order. Destination is the zone the revealed cards move to (graveyard or hand);
// other zones are not modeled and fail closed upstream. An empty Until matches
// the first card revealed.
type RevealUntil struct {
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
	Until       Selection            // first revealed card matching this stops the reveal
	Destination zone.Type            // graveyard or hand
}

// PutHandOnLibraryThenDraw has Player put any number of cards from their hand on
// one end of their library, then draw a number of cards equal to the number put
// plus DrawOffset. Bottom selects the library end the hand cards move to (bottom
// when true, top when false). It models "put any number of cards from your hand
// on the bottom of your library, then draw that many cards[ plus N]." The
// player-chosen count and the "that many" back-reference are not expressible
// through separate instructions, so the whole sequence resolves as one
// primitive.
type PutHandOnLibraryThenDraw struct {
	Player     PlayerReference
	Bottom     bool
	DrawOffset int
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

// ImpulseExile exiles cards from the top of a player's library and lets the
// resolving controller play those cards for a bounded duration.
type ImpulseExile struct {
	Player   PlayerReference
	Amount   Quantity
	Duration EffectDuration
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

// Manifest puts cards from a player's library onto the battlefield face down.
// Player identifies the manifesting player; the zero value (PlayerReferenceNone)
// manifests for the resolving ability's controller.
type Manifest struct {
	Dread  bool
	Player PlayerReference
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
	// ChooseKind removes a counter of a kind the resolving controller chooses
	// from among the kinds present on the object, modeling the kind-unspecified
	// "remove a counter from <permanent>" wording (Ferropede). When false the
	// fixed CounterKind is removed. It is ignored for Group removals.
	ChooseKind bool
}

// Transform transforms the referenced permanent.
type Transform struct {
	Object ObjectReference
}

// PhaseOut phases out one referenced permanent or every permanent in a
// referenced group.
type PhaseOut struct {
	Object ObjectReference
	Group  GroupReference
}

// Regenerate sets up a regeneration shield on the referenced permanent.
type Regenerate struct {
	Object ObjectReference
}

// BecomeCopy makes the source permanent become a copy of the referenced target
// permanent (CR 706), as for an activated/resolving copy ability ("This land
// becomes a copy of target land, except it has this ability.", Thespian's Stage;
// "... until end of turn.", Mirage Mirror). Object references the copied target.
// Card instead references a copied card in a non-battlefield zone, such as a
// permanent card in a graveyard ("... becomes a copy of target permanent card in
// your graveyard ...", Shifting Woodland); exactly one of Object or Card is set.
// UntilEndOfTurn limits the copy to end of turn; otherwise it lasts for as long
// as the source remains on the battlefield. RetainsThisAbility keeps the source's
// own become-a-copy ability so it can copy again, and AddKeywords applies any
// "except it has <keyword>" copiable riders.
type BecomeCopy struct {
	Object             ObjectReference
	Card               CardReference
	UntilEndOfTurn     bool
	RetainsThisAbility bool
	AddKeywords        []Keyword
}

// Attach attaches an Aura or Equipment to a permanent without paying an Equip
// cost, as for an enters-the-battlefield "attach it to target creature" trigger.
// Attachment references the moving attachment (typically the source permanent)
// and Target references the permanent it attaches to.
type Attach struct {
	Attachment ObjectReference
	Target     ObjectReference
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
// When Object references a permanent, the created replacement is bound to that
// resolved permanent (its AffectedObjectID), so it matches only events about
// that one object ("If it would leave the battlefield, exile it instead." on a
// just-reanimated creature). When Object is absent the replacement matches by
// its own filters alone.
type CreateReplacement struct {
	Replacement *ReplacementEffect
	Duration    EffectDuration
	Object      ObjectReference
}

// PreventDamage creates a damage-prevention shield for exactly one referenced
// player or permanent. When All is set the shield prevents every qualifying
// damage event (no fixed Amount); when CombatOnly is set it prevents only
// combat damage. By default the shield prevents damage dealt TO the referenced
// object; when BySource is set it prevents damage dealt BY that object instead.
// When Global is set the shield has no recipient or source object and prevents
// every qualifying damage event regardless of who would deal or receive it
// ("Prevent all combat damage that would be dealt this turn."); Global is
// mutually exclusive with Object, Player, and BySource.
type PreventDamage struct {
	Amount     Quantity
	Object     ObjectReference
	Player     PlayerReference
	All        bool
	CombatOnly bool
	BySource   bool
	Global     bool
}

// AddExtraPhases inserts additional phases into the current turn (CR 505.5,
// 506.2). It models "After this main phase, there is an additional combat
// phase[ followed by an additional main phase]." (Aggravated Assault, Aurelia
// the Warleader, World at War, Combat Celebrant). Combat queues an extra combat
// phase; Main queues an extra main phase after it. The runtime appends the
// queued phases to TurnState.ExtraPhases, which the turn loop drains in order.
type AddExtraPhases struct {
	Combat bool
	Main   bool
}

// RollDie rolls a single fair die with Sides faces and publishes the rolled
// value (1..Sides) as the instruction's resolved amount (CR 706). It backs
// "roll a d20" and similar dice mechanics; a later instruction consumes the
// result via a DynamicAmountPreviousEffectResult amount keyed to this
// instruction's PublishResult ("...equal to the result").
type RollDie struct {
	Sides int
}
