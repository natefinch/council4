package game

import (
	"github.com/natefinch/council4/mtg/game/color"
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
	PrimitiveLookAtLibraryTop
	PrimitiveCastForFree
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
	PrimitiveDiscardThenDraw
	PrimitiveRevealUntil
	PrimitiveBecomeSaddled
	PrimitiveAddExtraPhases
	PrimitiveLookAtHand
	PrimitiveRollDie
	PrimitiveRemoveFromCombat
	PrimitiveChooseDiscardFromHand
	PrimitiveShuffleGraveyardIntoLibrary
	PrimitiveGroupSelfPowerDamage
	PrimitiveBecomeMonarch
	PrimitiveCantBecomeMonarch
	PrimitiveRingTempts
	PrimitiveVote
	PrimitiveExileEntireHand
	PrimitiveReturnExiledCardsToHand
	PrimitivePutLinkedExiledCardsInLibrary
	PrimitiveConditionalDestinationPlace
	PrimitiveExileForEachPlayer
	PrimitiveReturnLinkedExiledCardsToBattlefield
	PrimitiveDestroyForEachPlayer
	PrimitiveCreateTokenForEachDestroyed
	PrimitiveRemoveTargetsForToken
	PrimitiveAdapt
	PrimitiveConnive
	PrimitivePayRepeatedly
	PrimitiveExileForPlay
	// PrimitiveHideawayExile is the Hideaway N enters action (look at top N,
	// exile one face down linked to the source, rest to bottom in random order).
	PrimitiveHideawayExile
	// PrimitivePlayHideawayCard plays the source's hidden-away exiled card
	// without paying its mana cost, gated by the enclosing instruction condition.
	PrimitivePlayHideawayCard
	// PrimitiveChooseFromZone is the single canonical "player chooses cards from
	// a zone matching a filter, then those cards move to a destination" primitive
	// (game.ChooseFromZone). It supersedes the retired per-family wrapper
	// primitives ExileFromHand, ExileFromGraveyard, PutFromHand, and
	// ReturnFromGraveyard, which now lower to a ChooseFromZone envelope.
	PrimitiveChooseFromZone
	// PrimitivePileSplit reveals the top N cards of a player's library, has the
	// separating player split them into two piles, has the choosing player pick
	// one pile, and routes the kept and other piles to their destinations
	// (game.PileSplit). It models the "Fact or Fiction" family.
	PrimitivePileSplit
	// PrimitiveRevealTopPartition reveals the top N cards of a player's library,
	// puts every revealed card matching a typed filter into that player's hand,
	// and routes the rest to a remainder destination (game.RevealTopPartition).
	// It models the "Reveal the top N cards of your library. Put all <type>
	// cards revealed this way into your hand and the rest <remainder>." family.
	PrimitiveRevealTopPartition
	// PrimitiveChampionExile is the Champion keyword enters action (CR 702.71):
	// the controller exiles another permanent they control matching the keyword's
	// type under an exile-until-leaves link, sacrificing the source when no
	// eligible permanent exists (game.ChampionExile).
	PrimitiveChampionExile
	// PrimitiveExileLibraryUntilNonlandCast exiles cards from the top of a
	// player's library until a nonland card is exiled, then lets that player cast
	// the nonland card without paying its mana cost (game.ExileLibraryUntilNonlandCast).
	PrimitiveExileLibraryUntilNonlandCast
	// PrimitiveDiscardUnlessType discards a fixed number of cards unless the
	// player instead discards a single card of an exempt type (game.DiscardUnlessType).
	PrimitiveDiscardUnlessType
	// PrimitiveEachPlayerChooseDestroy has every player, in turn order starting
	// with the resolving controller, choose up to one permanent from a shared
	// controller-relative pool, then destroys every chosen permanent
	// simultaneously (game.EachPlayerChooseDestroy).
	PrimitiveEachPlayerChooseDestroy
	// PrimitivePlayerMayPayGenericOrRule offers a referenced player the option to
	// pay a generic mana amount and, when they decline or cannot pay, installs a
	// set of rule effects on that player's permanents for a duration
	// (game.PlayerMayPayGenericOrRule). It models the "that opponent may pay {X},
	// where X is the number of cards in their hand. If they don't, they can't
	// attack you this combat." punisher (Champions of Minas Tirith).
	PrimitivePlayerMayPayGenericOrRule
	// PrimitivePartitionExiledCostCards disposes of the cards exiled to pay the
	// resolving ability's cost: one player chooses one card, which goes to the
	// bottom (or top) of its owner's library, and every other exiled card
	// returns to the battlefield under the controller's control, optionally
	// tapped (game.PartitionExiledCostCards). It models "An opponent chooses one
	// of the exiled cards. You put that card on the bottom of your library and
	// return the other to the battlefield tapped." (Coin of Fate).
	PrimitivePartitionExiledCostCards
	// PrimitiveExileForEachOpponent walks each opponent of the resolving
	// controller and, for each, has Chooser exile up to one permanent that
	// opponent controls matching Selection, publishing each exiled permanent
	// under LinkedKey (game.ExileForEachOpponent). It models "for each opponent,
	// exile up to one target permanent that player controls ..." (King Solomon's
	// Frogs).
	PrimitiveExileForEachOpponent
	// PrimitiveDrawForEachExiled has each linked exiled permanent's last-known
	// controller draw one card, consuming the LinkedKey a sibling
	// ExileForEachOpponent published (game.DrawForEachExiled). It models "For
	// each permanent exiled this way, its controller draws a card." (King
	// Solomon's Frogs).
	PrimitiveDrawForEachExiled
	// PrimitiveCreateReflexiveTrigger puts a reflexive triggered ability
	// (CR 603.11) on the stack: "When you do, <effect>." following an optional
	// enabling action in the same resolution. Unlike CreateDelayedTrigger (which
	// is timed or event-based), the reflexive trigger is queued immediately when
	// the enabling action was performed and put on the stack the next time a
	// player would receive priority, with its targets chosen then — after the
	// enabling action has resolved (game.CreateReflexiveTrigger).
	PrimitiveCreateReflexiveTrigger

	// PrimitiveExilePermanentForPlay exiles a target permanent from the
	// battlefield and grants that card's owner permission to play it from exile
	// for as long as it remains exiled ("exile up to one other target tapped
	// creature or Vehicle. For as long as that card remains exiled, its owner may
	// play it.", Prowl, Stoic Strategist).
	PrimitiveExilePermanentForPlay

	// PrimitivePlayChosenExiledCard has the resolving controller choose one card
	// in exile that a scoped player owns and that bears a named exile counter,
	// then grants the controller permission to play the chosen card for a bounded
	// duration, optionally without paying its mana cost ("Choose an exiled card an
	// opponent owns with a void counter on it. You may play it this turn without
	// paying its mana cost.", Dauthi Voidwalker).
	PrimitivePlayChosenExiledCard

	// PrimitiveReturnExiledCardsWithCounter returns every card Player owns in
	// exile that bears a named marker counter to Player's hand ("Put all exiled
	// cards you own with intel counters on them into your hand.", Flamewar, Brash
	// Veteran). It is the return companion to the exile-with-named-counter
	// substrate (game.ReturnExiledCardsWithCounter).
	PrimitiveReturnExiledCardsWithCounter

	// PrimitiveBolster performs the bolster keyword action (game.Bolster).
	PrimitiveBolster

	// PrimitiveChooseDrawnPayLifeOrTop has Player choose ChooseCount cards in
	// their hand that were drawn this turn and, for each chosen card, pay LifeCost
	// life to keep it or put it on top of their library ("choose two cards in your
	// hand drawn this turn. For each of those cards, pay 4 life or put the card on
	// top of your library.", Sylvan Library).
	PrimitiveChooseDrawnPayLifeOrTop

	// PrimitiveExileTopEachLibraryCastFree exiles the top Amount cards of every
	// player's library into their owners' exile and then lets the resolving
	// controller cast any number of those just-exiled cards without paying their
	// mana costs ("exile the top card of each player's library, then you may cast
	// any number of spells from among those cards without paying their mana
	// costs.", Etali, Primal Storm). It is game.ExileTopEachLibraryCastFree.
	PrimitiveExileTopEachLibraryCastFree

	// PrimitiveExchangeLifeTotalWithSourceCharacteristic exchanges a player's
	// life total with the resolving source permanent's power or toughness.
	PrimitiveExchangeLifeTotalWithSourceCharacteristic

	// PrimitiveRecordEchoObligation records the resolving source permanent's
	// current controller as the player for whom its Echo obligation (CR 702.29)
	// has been resolved, so later upkeeps of that same controller do not
	// re-trigger the echo.
	PrimitiveRecordEchoObligation

	// PrimitiveGainCityBlessing is the spell form of ascend (CR 702.131a). As the
	// spell resolves, before its other instructions, its controller gets the
	// city's blessing if they control ten or more permanents and don't already
	// have it. The primitive carries no payload; it always acts on the resolving
	// object's controller.
	PrimitiveGainCityBlessing

	// PrimitiveCopyCard offers the resolving controller the chance to copy the
	// card exiled by this source under an object-scoped link (the imprinted card,
	// CR 707.12). It is the enabling half of the "You may copy the exiled card. If
	// you do, you may cast the copy without paying its mana cost." imprint idiom
	// (Isochron Scepter, Spellbinder): it succeeds only when a linked exiled card
	// still rests in exile, so a following PlayLinkedExiledCard cast is gated on a
	// copy actually being available. The copy itself is materialized and cast by
	// that following instruction; a copy never cast ceases to exist (CR 707.12a),
	// so this consent step performs no observable game action of its own.
	PrimitiveCopyCard

	// PrimitivePlayLinkedExiledCard casts the card exiled by this source under an
	// object-scoped link (the imprinted card) without paying its mana cost when
	// WithoutPayingManaCost is set, choosing the first legal targets and modes with
	// X treated as 0. When Copy is set it casts a copy of the linked exiled card
	// (Isochron Scepter, Spellbinder) rather than the card itself: the copy is a
	// spell that carries the exiled card's copiable values and ceases to exist when
	// it leaves the stack, leaving the original card in exile (CR 707.12). It is
	// the consequence half of the imprint copy/cast idiom, gated by a preceding
	// CopyCard.
	PrimitivePlayLinkedExiledCard

	// PrimitiveTapChosenGroup lets the resolving controller choose any number of
	// permanents from a group the enclosing ability restricts to untapped
	// permanents they control matching a subtype or other selection, and taps
	// each chosen permanent, publishing the number tapped for later scaled
	// effects (Myr Battlesphere's "you may tap X untapped Myr you control").
	PrimitiveTapChosenGroup
	// PrimitiveIterativeLibraryProcess exiles or reveals cards from the top of a
	// player's library one at a time, tracking the cards processed this way,
	// until a name-based stop predicate fires (Tainted Pact, Demonic
	// Consultation). It is the generic iterative library processor.
	PrimitiveIterativeLibraryProcess
	// PrimitiveManifestForEachLinked manifests or cloaks one card for each object
	// a prior instruction published under LinkedKey, using each linked object's
	// last-known controller as the manifesting player.
	PrimitiveManifestForEachLinked
	// PrimitiveIncubate performs the incubate keyword action (CR 701.55): the
	// recipient creates an Incubator token with Amount +1/+1 counters on it
	// (game.Incubate). Added last so existing kinds keep their wire values.
	PrimitiveIncubate
)

// primitiveKindCount is the number of supported primitive kinds.
const primitiveKindCount = int(PrimitiveIncubate) + 1

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

	// ExcessRecipient redirects the portion of Amount beyond what is lethal to a
	// single permanent recipient onto this player instead, modeling "Excess
	// damage is dealt to that creature's controller instead." (Pigment Storm,
	// Flame Spill). When set, the permanent recipient is dealt only its lethal
	// damage and the remainder is dealt to this player as a single damage event.
	// It is valid only with a single-permanent (object or any-target) recipient
	// and a player ExcessRecipient. The zero value leaves damage undivided.
	ExcessRecipient DamageRecipient

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

// GroupSelfPowerDamage has each permanent in a battlefield group deal damage to
// itself equal to its own power, evaluated per member ("Each creature deals
// damage to itself equal to its power.", Wave of Reckoning; "Each tapped
// creature deals damage to itself equal to its power.", The Akroan War chapter
// III). Every group member is both the damage source and the recipient, and the
// amount is that member's power computed individually rather than a single
// group-wide value.
type GroupSelfPowerDamage struct {
	Group GroupReference
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

// ShuffleGraveyardIntoLibrary moves every card in a referenced player's
// graveyard into that player's library and then shuffles it ("shuffle your
// graveyard into your library", The Mending of Dominaria). The library is
// shuffled even when the graveyard is empty (CR 701.x shuffle).
//
// IncludeHand additionally moves every card in the player's hand into the
// library before shuffling ("shuffle your hand and graveyard into your
// library", Midnight Clock). It is false for the graveyard-only shuffle, so the
// zero value preserves the original single-zone behavior.
type ShuffleGraveyardIntoLibrary struct {
	Player      PlayerReference
	PlayerGroup PlayerGroupReference
	IncludeHand bool
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
//
// PublishLinked, when set, remembers each discarded card under this key so a
// later instruction can read it ("Discard a card, then ... deals damage equal
// to that card's mana value ..."). It is meaningful only for a single-player,
// non-entire-hand discard.
type Discard struct {
	Amount        Quantity
	Player        PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup   PlayerGroupReference // opponents or all players; zero if Player is set
	EntireHand    bool
	AtRandom      bool
	PublishLinked LinkedKey
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
	// CombinationColors, when non-empty, makes this instruction produce Amount
	// mana distributed freely by the recipient among these colors ("add three
	// mana in any combination of {R} and/or {G}", Goblin Clearcutter; "add two
	// mana in any combination of colors", Manamorphose). Each of the Amount mana
	// is independently one of these colors, chosen by the recipient at
	// resolution (CR 106.1b), so a color may receive any share from zero up to
	// the whole amount. It holds two or more distinct basic colors and is
	// mutually exclusive with ManaColor, ChoiceFrom, EntryChoiceFrom,
	// EachControlledColor, and SpendRider. Amount may be fixed or dynamic.
	CombinationColors []mana.Color
	// PersistUntilEndOfTurn, when set, makes the mana produced by this
	// instruction not empty as steps and phases end for the rest of the turn
	// (the CR 500.4 exception used by "Until end of turn, you don't lose this
	// mana as steps and phases end", Grand Warlord Radha). The mana is added to
	// the recipient's pool as persistent mana (Pool.AddPersistent); it is
	// released at end-of-turn cleanup so it empties normally thereafter.
	PersistUntilEndOfTurn bool
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
	// ChooseOne makes the resolving controller choose exactly one permanent from
	// Group to receive the counters, rather than every member ("put a vigilance
	// counter on a creature you control", Ajani Fells the Godsire chapter II). It
	// is set only with a Group and an empty Object; when no group member exists
	// the effect does nothing.
	ChooseOne bool
	// KindChoices, when non-empty, lets the resolving controller choose one
	// counter kind from this list to place, ignoring CounterKind ("Put a +1/+1
	// counter or a loyalty counter on it.", Elspeth Conquers Death chapter III).
	// It holds two or more distinct, permanent-placeable kinds and is set only
	// with a single Object, never a Group or AllKinds.
	KindChoices []counter.Kind
	// Distribute makes the resolving controller split Amount counters among the
	// permanents chosen for a target spec, each receiving at least one ("Distribute
	// three +1/+1 counters among one, two, or three target creatures"). It is the
	// counter analog of Damage.Divided: Object addresses the target spec through an
	// AllTargetPermanents reference, and it is set only with that Object, never a
	// Group, AllKinds, ChooseOne, or KindChoices.
	Distribute bool
	// DoubleKind doubles the CounterKind already on each permanent in Group: every
	// member receives as many more counters of CounterKind as it currently has,
	// ignoring Amount. It backs "double the number of +1/+1 counters on each
	// creature you control" (Bristly Bill, Spine Sower). It is set only with a
	// Group, never an Object, AllKinds, ChooseOne, KindChoices, or Distribute. The
	// single-object and one-target forms use a dynamic ObjectCounters Amount.
	DoubleKind bool
	// PublishLinked, when set, remembers the single permanent the counters were
	// placed on under this key so a later linked effect (such as a delayed
	// attacker-declared trigger that binds to that creature) can resolve it. It is
	// set only with a single Object, never a Group.
	PublishLinked LinkedKey
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
//
// When ChooseFrom is set, the resolving controller instead chooses up to
// ChooseUpTo distinct permanents from that group at resolution and the
// continuous effects are applied to each chosen permanent ("up to that many
// target lands you control become 3/3 creatures ...", Primal Adversary). The
// ChooseUpTo amount may be dynamic, so a payment-count or other resolution
// number can bound the selection. Object and ChooseFrom are mutually exclusive.
type ApplyContinuous struct {
	Object            opt.V[ObjectReference]
	ContinuousEffects []ContinuousEffect
	Duration          EffectDuration
	PublishLinked     LinkedKey

	ChooseFrom GroupReference
	ChooseUpTo Quantity
	Prompt     string
}

// ApplyRule creates rule effects for a target (or globally).
type ApplyRule struct {
	Object      opt.V[ObjectReference]
	RuleEffects []RuleEffect
	Duration    EffectDuration
}

// PlayerMayPayGenericOrRule offers Player the option to pay a generic mana
// amount. When Player declines or cannot pay, it installs RuleEffects on that
// player's permanents for Duration. It models the "that opponent may pay {X},
// where X is the number of cards in their hand. If they don't, they can't
// attack you this combat." punisher (Champions of Minas Tirith), where Amount
// is the payer's hand size and RuleEffects prohibit that player's creatures
// from attacking the resolving controller.
type PlayerMayPayGenericOrRule struct {
	Player      PlayerReference
	Amount      Quantity
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
	// EntryTransformed makes a transforming double-faced card enter the
	// battlefield converted, as its back face, backing "return it to the
	// battlefield converted" (CR 712). It is honored only for a transforming
	// double-faced card entering from its front face and is ignored otherwise.
	EntryTransformed bool
	EntryCounters    []CounterPlacement
	PublishLinked    LinkedKey
	// LinkedReturnZones is the ordered set of non-battlefield zones a
	// LinkedBattlefieldSource return may pull its linked card from. nil means
	// exile-only, the default for exile-until and blink returns (Palace Jailer,
	// Oblivion Ring), which must return the card only while it remains the same
	// object in exile and do nothing once it has left. A sacrifice-then-return
	// effect that put the card into the graveyard sets {zone.Graveyard} so it
	// returns that card from the graveyard.
	LinkedReturnZones []zone.Type
}

// LinkedReturnZonesOrExile returns the ordered non-battlefield zones a
// LinkedBattlefieldSource return may pull its linked card from, defaulting to
// exile-only when unset.
func (p PutOnBattlefield) LinkedReturnZonesOrExile() []zone.Type {
	if len(p.LinkedReturnZones) == 0 {
		return []zone.Type{zone.Exile}
	}
	return p.LinkedReturnZones
}

// CreateToken creates one or more tokens. EntryTapped makes every created token
// enter the battlefield tapped, matching "Create a tapped ... token." wording.
// EntryAttacking puts every created token onto the battlefield already attacking
// (CR 508.4), matching "... token that's tapped and attacking." wording; it has
// effect only while the token's controller is the attacking player in an active
// combat and is otherwise ignored, leaving the token to enter normally.
type CreateToken struct {
	Amount    Quantity
	Source    TokenSource
	Recipient opt.V[PlayerReference]
	// RecipientGroup, when set, creates the token for each player in the group
	// rather than for a single recipient ("Each player creates a 1/1 white
	// Soldier creature token.", "Each opponent creates a Treasure token."). The
	// handler resolves the group and creates the token amount for every member in
	// APNAP order. It is mutually exclusive with Recipient and unset for the
	// single-recipient forms.
	RecipientGroup PlayerGroupReference
	EntryTapped    bool
	EntryAttacking bool

	// AttackEachOtherOpponent, when set, creates one token for each opponent of
	// the ability's controller other than the defending player of the attack
	// that triggered the ability, offering the controller a separate "you may"
	// for each such opponent and putting each accepted token onto the
	// battlefield tapped and attacking that opponent (CR 508.4). It backs the
	// myriad keyword (CR 702.116): "for each opponent other than the defending
	// player, you may create a token that's a copy of this creature that's
	// tapped and attacking that player." The recipient of every token is the
	// controller regardless of which opponent it attacks; Recipient and
	// RecipientGroup are ignored. It is unset for ordinary token creation.
	AttackEachOtherOpponent bool

	// AttackSameAsSource, when set, puts each created token onto the battlefield
	// attacking the same player or planeswalker the ability's source creature is
	// attacking in the current combat (CR 508.4, CR 702.169b). It backs the
	// mobilize keyword's "create N tapped and attacking 1/1 red Warrior creature
	// tokens": the tokens join the source's attack rather than being declared
	// against a freely chosen defender. It reads the source's live attack
	// declaration, falling back to the defending player recorded on the trigger
	// event if the source has left combat. Unlike EntryAttacking it never prompts
	// for a defender. It is unset for ordinary token creation and is mutually
	// exclusive with EntryAttacking and AttackEachOtherOpponent.
	AttackSameAsSource bool

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

// PutLinkedExiledCardsInLibrary moves every card a sibling clause exiled under
// LinkedKey from exile to its owner's library, to the bottom when Bottom is set.
// It backs the linked disposal "The owner of each card exiled with <this
// permanent> puts that card on the bottom of their library." (Trial of a Time
// Lord), consuming the link the paired exile-until-leaves clause published so
// the runtime clears it and the synthesized leaves trigger returns nothing.
type PutLinkedExiledCardsInLibrary struct {
	LinkedKey LinkedKey
	Bottom    bool
}

// PartitionExiledCostCards disposes of the cards exiled to pay the resolving
// ability's activation cost by having one player choose a single card ("that
// card") and routing it and the remaining cards to two destinations. When
// ChooserOpponent is set, the next opponent of the ability's controller in turn
// order chooses; otherwise the controller chooses. The chosen card goes to the
// bottom of its owner's library when ChosenToLibraryBottom is set (top
// otherwise); every other exiled card returns to the battlefield under the
// controller's control, tapped when OtherEntersTapped is set. It backs "An
// opponent chooses one of the exiled cards. You put that card on the bottom of
// your library and return the other to the battlefield tapped." (Coin of Fate),
// reading the resolving object's cost-exiled card IDs. Only cards still in exile
// are considered, so a card that already moved is skipped.
type PartitionExiledCostCards struct {
	ChooserOpponent       bool
	ChosenToLibraryBottom bool
	OtherEntersTapped     bool
}

// StartEngines starts engine effects for a player.
type StartEngines struct {
	Player PlayerReference
}

// BecomeMonarch makes the referenced player the monarch (CR 720). At most one
// player is the monarch at a time, so the runtime clears any prior monarch when
// it applies this primitive. It backs "you become the monarch" and "target
// player becomes the monarch".
type BecomeMonarch struct {
	Player PlayerReference
}

// CantBecomeMonarch blocks the referenced player from becoming the monarch for
// the rest of the turn ("You can't become the monarch this turn.", Jared
// Carthalion). The runtime sets a per-turn flag the monarch-designation code
// honors; it is cleared as the next turn begins.
type CantBecomeMonarch struct {
	Player PlayerReference
}

// GainCityBlessing is the spell form of ascend (CR 702.131a): as the spell
// resolves, before its other instructions, its controller gets the city's
// blessing if they control ten or more permanents and don't already have it.
// The city's blessing is player-level persistent state that is never removed.
// It always acts on the resolving object's controller and carries no payload.
type GainCityBlessing struct{}

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

// Incubate performs the incubate keyword action (CR 701.55): the recipient
// creates an Incubator token with Amount +1/+1 counters on it. An Incubator
// token is a colorless artifact with "{2}: Transform this artifact." whose back
// face is a 0/0 colorless Phyrexian artifact creature (CR 701.55a-c); the
// +1/+1 counters carry through the transform, so the creature side has power
// and toughness equal to Amount. Recipient names the creating player when set
// (an exiled permanent's controller for "its controller incubates X"); it
// defaults to the resolving object's controller. Incubate 0 still creates the
// token with no counters (CR 701.55a places no minimum).
type Incubate struct {
	Amount    Quantity
	Recipient opt.V[PlayerReference]
}

// Renown performs the renown keyword action (CR 702.111): if the referenced
// permanent is not already renowned, the controller puts Amount +1/+1 counters
// on it and it becomes renowned. A renowned permanent is left unchanged, so the
// effect applies at most once.
type Renown struct {
	Object ObjectReference
	Amount Quantity
}

// Adapt performs the Adapt keyword action (CR 701.43): if the referenced
// creature has no +1/+1 counters on it, the controller puts Amount +1/+1
// counters on it. A creature that already has a +1/+1 counter is left
// unchanged, so the effect applies only while the creature is uncountered.
type Adapt struct {
	Object ObjectReference
	Amount Quantity
}

// Bolster performs the bolster keyword action (CR 701.37): the controller
// chooses a creature with the least toughness among creatures they control,
// then puts Amount +1/+1 counters on that creature. If several creatures are
// tied for the least toughness, the controller chooses one of them; if the
// controller controls no creatures, nothing happens. When PublishLinked is set,
// the chosen creature is remembered under that key so a later linked effect
// (such as "the chosen creature gains trample" or a delayed trigger watching
// that creature deal combat damage) can resolve it.
type Bolster struct {
	Amount        Quantity
	PublishLinked LinkedKey
}

// Connive performs the connive keyword action (CR 702.154): the controller of
// the conniving permanent draws Amount cards, then discards Amount cards, and a
// +1/+1 counter is placed on Object for each nonland card discarded this way.
// Player draws and discards (the conniving permanent's controller) and Object is
// the conniving permanent that receives the counters.
type Connive struct {
	Object ObjectReference
	Player PlayerReference
	Amount Quantity
}

// BecomeSaddled performs the Saddle keyword action (CR 702.166): the referenced
// Mount becomes saddled until end of turn. The saddled state is cleared during
// cleanup. The effect is idempotent; saddling an already-saddled Mount leaves it
// unchanged.
type BecomeSaddled struct {
	Object ObjectReference
}

// RecordEchoObligation records the resolving source permanent's current
// controller as the player for whom its Echo obligation (CR 702.29) has been
// resolved. The echo triggered ability runs it each time it resolves so a later
// upkeep of the same controller does not re-trigger the pay-or-sacrifice, while
// a new controller (whose identity no longer matches the recorded one) still
// triggers echo at their next upkeep.
type RecordEchoObligation struct {
	Object ObjectReference
}

// Pay prompts the controller to pay an optional cost during resolution.
// The instruction's Optional field controls whether declining is allowed.
// Results are published via the Instruction.PublishResult for downstream ResultGate checks.
type Pay struct {
	Payment ResolutionPayment
	Prompt  string
}

// PayRepeatedly prompts the controller to pay an optional cost any number of
// times during resolution and records how many times it was paid ("you may pay
// {1}{G} any number of times.", the Adversary cycle; "you may pay {2} any number
// of times.", Squad; "you may pay {1}{G} any number of times.", Taste of
// Paradise). The controller is offered Payment repeatedly; each accepted and
// successful payment increases the recorded count by one, and the loop stops the
// first time the controller declines or can no longer pay. The final count is
// published under PublishCount as a ResolutionChoiceNumber result so a later
// instruction reads it through DynamicAmountChosenNumber ("put that many +1/+1
// counters on this creature", "create that many tokens"). A count of zero is
// published when the controller never pays, which lets a gated reflexive payoff
// resolve to nothing. The loop is bounded by an internal cap so a free or
// fully-affordable cost cannot iterate without limit.
//
// MaxCount, when set, bounds the number of payments to a rules-derived amount
// evaluated as the instruction resolves (negative values are treated as zero).
// It backs "pay {X}, where X is less than or equal to <triggering amount>"
// (Well of Lost Dreams' "pay {X}, where X is less than or equal to the amount of
// life you gained"): the per-unit cost is offered up to that many times so the
// published count is the chosen X, never exceeding the triggering quantity. When
// MaxCount is absent the loop uses only the internal cap.
type PayRepeatedly struct {
	Payment      ResolutionPayment
	PublishCount ResultKey
	Prompt       string
	MaxCount     opt.V[*DynamicAmount]
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

// SourcePowerToughness identifies which source characteristic participates in
// a numerical exchange.
type SourcePowerToughness int

// Source power/toughness choices.
const (
	SourcePowerToughnessNone SourcePowerToughness = iota
	SourcePower
	SourceToughness
)

// ExchangeLifeTotalWithSourceCharacteristic exchanges Player's life total with
// the resolving source permanent's current power or toughness.
type ExchangeLifeTotalWithSourceCharacteristic struct {
	Player         PlayerReference
	Characteristic SourcePowerToughness
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

// ExileEntireHand exiles every card in Player's hand at once with no choice,
// modeling the involuntary whole-hand wording "exile all cards from your hand."
// (Wormfang Behemoth). Each exiled card is remembered under LinkedKey, keyed by
// the source permanent's card identity, so a paired ReturnExiledCardsToHand on
// the same face returns exactly that set when the source leaves. LinkedKey must
// be set; the exiled cards are otherwise unrecoverable.
type ExileEntireHand struct {
	Player    PlayerReference
	LinkedKey LinkedKey
}

// ReturnExiledCardsToHand returns the cards an earlier ExileEntireHand exiled
// under LinkedKey to their owners' hands, modeling "return the exiled cards to
// their owner's hand." (Wormfang Behemoth). It consumes the source-keyed linked
// set the paired exile published and clears it after returning; cards no longer
// in exile are skipped. LinkedKey must be set.
type ReturnExiledCardsToHand struct {
	LinkedKey LinkedKey
}

// ReturnExiledCardsWithCounter moves every card that Player owns in the exile
// zone bearing at least one Counter-kind named marker counter to Player's hand
// ("Put all exiled cards you own with intel counters on them into your hand.",
// Flamewar, Brash Veteran). It is the return companion to the exile-with-named-
// counter substrate: the counters recorded in Game.ExileCounters when the cards
// were exiled select exactly which cards return, so any card that uses a named
// marker counter (croak, intel, void, collection, ...) benefits without the
// primitive naming a specific counter. Cards without the counter, and cards
// owned by other players, are unaffected; an empty result is a legal no-op. The
// counters are cleared as the cards leave exile.
type ReturnExiledCardsWithCounter struct {
	Player  PlayerReference
	Counter counter.Kind
}

// ExileForEachPlayer walks every player in the game and, for each, has Chooser
// pick up to one permanent that player controls matching Selection and exiles
// it, remembering each chosen permanent under LinkedKey keyed by the source. It
// models both exile-until-leaves Saga chapters (Vault 13: Dweller's Journey,
// Battle at the Helvault) and plain distributive removal with a linked payoff
// (Unexplained Absence). Each player's permanents are an independent candidate
// pool, so the effect exiles at most one per player. Selection's ExcludeSource
// models the "other" qualifier. LinkedKey must be set so a paired return or
// payoff can consume the exiled set.
type ExileForEachPlayer struct {
	Chooser   PlayerReference
	Selection Selection
	LinkedKey LinkedKey
}

// ChampionExile is the Champion keyword enters-the-battlefield action (CR
// 702.71): the source's controller exiles another permanent they control
// matching Selection, remembering it under LinkedKey (an exile-until-leaves
// link) so the paired return brings it back when the source leaves. When the
// controller controls no other matching permanent, they instead sacrifice the
// source so nothing is championed. Selection's ExcludeSource models the
// "another" qualifier. LinkedKey must be set; the exiled card is otherwise
// unrecoverable.
type ChampionExile struct {
	Selection Selection
	LinkedKey LinkedKey
}

// ReturnLinkedExiledCardsToBattlefield returns up to Amount cards a sibling
// exile-until-leaves clause exiled under LinkedKey to the battlefield under
// their owners' control; Chooser picks which cards return when more than Amount
// remain in exile. When RestToLibraryBottom is set, every remaining linked card
// moves to the bottom of its owner's library. It models the partial Saga payoff
// "Return N cards exiled with this Saga to the battlefield under their owners'
// control and put the rest on the bottom of their owners' libraries." (Vault 13:
// Dweller's Journey). It consumes and clears the link after resolving, so the
// synthesized leaves-the-battlefield safety-net return finds nothing left.
// LinkedKey must be set.
type ReturnLinkedExiledCardsToBattlefield struct {
	Chooser             PlayerReference
	LinkedKey           LinkedKey
	Amount              Quantity
	RestToLibraryBottom bool
}

// DestroyForEachPlayer walks every player in the game and, for each, has Chooser
// pick up to one permanent that player controls matching Selection and destroys
// it, remembering each destroyed permanent under LinkedKey keyed by the source
// permanent. It models the distributive Saga chapter "For each player, destroy
// up to one target creature that player controls." (The Curse of Fenric, chapter
// I). Each player's permanents are an independent candidate pool, so the chapter
// destroys at most one per player. The destroyed permanents are linked so a
// paired CreateTokenForEachDestroyed clause creates one token for each, under
// that permanent's last-known controller. LinkedKey must be set; the destroyed
// permanents are otherwise unrecoverable for the token payoff.
type DestroyForEachPlayer struct {
	Chooser   PlayerReference
	Selection Selection
	LinkedKey LinkedKey
}

// EachPlayerChooseDestroy has every player, in turn order starting with the
// resolving controller, choose up to one permanent from a single shared
// candidate pool — the battlefield permanents matching Selection evaluated
// relative to the ability's controller — and then destroys every chosen
// permanent simultaneously. It models "Starting with you, each player may
// choose an artifact or enchantment you don't control. Destroy each permanent
// chosen this way." (Druid of Purification): each player is their own chooser,
// the pool is controller-relative and identical for every chooser (so
// Selection's Controller resolves against the source, letting "you don't
// control" offer the same permanents to all), and a permanent chosen by more
// than one player is destroyed once. Optional models the "may" so a player may
// decline even when the pool is non-empty; when false every player with a
// non-empty pool must choose one. PreventRegeneration carries a "can't be
// regenerated" rider onto the simultaneous destroy.
type EachPlayerChooseDestroy struct {
	Selection           Selection
	Optional            bool
	PreventRegeneration bool
}

// CreateTokenForEachDestroyed creates one token defined by Source for each
// permanent a sibling DestroyForEachPlayer recorded under LinkedKey, giving each
// token to that destroyed permanent's last-known controller. It models the per-
// controller Saga payoff "For each creature destroyed this way, its controller
// creates a <token>." (The Curse of Fenric, chapter I). It consumes and clears
// the link after resolving. LinkedKey must be set and Source must be valid.
type CreateTokenForEachDestroyed struct {
	Source    TokenSource
	LinkedKey LinkedKey
}

// ExileForEachOpponent walks each opponent of the resolving controller and, for
// each, has Chooser pick up to one permanent that opponent controls matching
// Selection and exiles it permanently, remembering each exiled permanent under
// LinkedKey keyed by the source permanent. It models the distributive enters
// trigger "for each opponent, exile up to one target permanent that player
// controls with mana value 3 or greater." (King Solomon's Frogs). Each
// opponent's permanents are an independent candidate pool, so the trigger exiles
// at most one per opponent. Unlike ExileForEachPlayer this is a plain exile with
// no return link; the linked set is recorded only so a paired DrawForEachExiled
// payoff can iterate the exiled permanents and read each one's last-known
// controller. LinkedKey must be set; the exiled permanents are otherwise
// unrecoverable for the draw payoff.
type ExileForEachOpponent struct {
	Chooser   PlayerReference
	Selection Selection
	LinkedKey LinkedKey
}

// DrawForEachExiled has each permanent a sibling ExileForEachOpponent recorded
// under LinkedKey draw one card for that permanent's last-known controller. It
// models the per-controller payoff "For each permanent exiled this way, its
// controller draws a card." (King Solomon's Frogs). It consumes and clears the
// link after resolving. LinkedKey must be set.
type DrawForEachExiled struct {
	LinkedKey LinkedKey
}

// ManifestForEachLinked manifests or cloaks one card for each permanent a prior
// linked removal recorded under LinkedKey, using that permanent's last-known
// controller as the acting player. It consumes and clears the linked set.
type ManifestForEachLinked struct {
	Dread     bool
	Cloak     bool
	LinkedKey LinkedKey
}

// RemoveTargetsForToken destroys (or, when Exile is set, exiles) every permanent
// chosen for the spell's single variable-count target spec, remembering each
// removed permanent under LinkedKey keyed by the source so a paired
// CreateTokenForEachDestroyed clause mints one token for each under that
// permanent's last-known controller. It models the variable-target removal-token
// family "Destroy any number of target creatures. For each creature destroyed
// this way, its controller creates a <token>." (Descent of the Dragons) and
// "Exile X target creatures. For each creature exiled this way, its controller
// creates a <token>." (Curse of the Swine). All chosen targets leave together as
// one simultaneous event; PreventRegeneration applies to the destroy form.
// LinkedKey must be set; the removed permanents are otherwise unrecoverable for
// the token payoff.
type RemoveTargetsForToken struct {
	Exile               bool
	PreventRegeneration bool
	LinkedKey           LinkedKey
}

// CastForFree has Player cast one card from Zone without paying its mana cost,
// modeling "(You may) cast a spell [with mana value N or less] from your hand
// without paying its mana cost." and similar free-cast effects. The enclosing
// Instruction's Optional flag expresses a "you may" wrapper, so the engine
// gathers consent before this runs.
//
// When Card is unset (CardReferenceNone), the resolving player chooses which
// eligible card matching Selection to cast from their own Zone; no eligible card
// casts nothing. When Card is set, that one referenced card is cast instead —
// the spell-or-ability already targeted it (Memory Plunder targets an instant or
// sorcery card in an opponent's graveyard), so Selection is ignored and the card
// is cast from whichever player's Zone currently holds it.
//
// ExileOnResolution sets the cast spell to move to exile instead of its owner's
// graveyard after it resolves or is countered, modeling the recurring rider "If
// that spell would be put into your graveyard, exile it instead." (Torrential
// Gearhulk).
type CastForFree struct {
	Player            PlayerReference
	Selection         Selection
	Zone              zone.Type
	Card              CardReference
	ExileOnResolution bool
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
//
// FromTriggerBatch restricts the moved cards to those that triggered the
// enclosing one-or-more zone-change ability, modeling "Whenever one or more
// <filter> cards are put into your graveyard ..., put them onto the
// battlefield" (Hedge Shredder). "Them" denotes exactly the coalesced batch
// of triggering cards rather than the whole graveyard, so only cards still in
// the graveyard whose IDs appear in that batch move.
type MassReturnFromGraveyard struct {
	Player            PlayerReference
	Selection         Selection
	Destination       zone.Type
	EntryTapped       bool
	SourceGroup       PlayerGroupReference
	ControlledByOwner bool
	FromTriggerBatch  bool
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
	// Counter names a named marker counter placed on the moved card when
	// Destination is the exile zone ("exile it with a croak counter on it.",
	// Grolnok, the Omnivore). The counter is recorded in Game.ExileCounters so
	// the source's paired play/cast/return ability can later select "cards ... in
	// exile with <name> counters on them". It is unset for every move that places
	// no counter, and meaningful only for the single-card exile form.
	Counter opt.V[counter.Kind]
	// PublishLinked remembers a successfully exiled single card under a
	// source-scoped linked key.
	PublishLinked LinkedKey
	// PublishLinkedObjectScoped keys PublishLinked by the source permanent's
	// current object identity, so a re-entered source starts with a fresh pool.
	PublishLinkedObjectScoped bool
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

// ExileForPlay exiles a referenced card from a specific zone and grants the
// resolving controller permission to play (or, when Cast is set, cast) it for a
// bounded duration. The move and the permission grant happen atomically: the
// card identity is captured before the move so the permission binds by identity
// rather than through the pre-exile event reference, which the move would
// otherwise invalidate by advancing the card's zone version.
//
// When SelectFromBatch is set the exiled card is not read from Card; instead the
// resolving controller chooses one card from the triggering batch event still in
// FromZone ("you may exile one of them from your graveyard" over a "discard one
// or more cards" batch). Card is ignored in that mode.
type ExileForPlay struct {
	Card            CardReference
	FromZone        zone.Type
	Duration        EffectDuration
	Cast            bool
	SelectFromBatch bool
}

// ExilePermanentForPlay exiles a target permanent from the battlefield and
// grants that card's owner permission to play it from exile for as long as it
// remains exiled ("exile up to one other target tapped creature or Vehicle. For
// as long as that card remains exiled, its owner may play it.", Prowl, Stoic
// Strategist). The owner scope is what distinguishes it from ExileForPlay, whose
// permission binds to the resolving controller: here the exiled card's owner —
// who may be an opponent — gains the permission. Each exiled card is remembered
// under LinkedKey, keyed by the source permanent, so a paired "whenever a player
// plays a card exiled with this" trigger can recognize that provenance. Object
// selects up to one target permanent, so an unresolved or absent target exiles
// nothing.
type ExilePermanentForPlay struct {
	Object    ObjectReference
	LinkedKey LinkedKey
}

// PlayChosenExiledCard has Player choose one card resting in Zone that is owned
// by a player matching OwnerScope (evaluated relative to Player) and, when
// Counter is set, bears at least one Counter-kind exile marker counter, then
// grants Player permission to play that chosen card for Duration. When
// WithoutPayingManaCost is set the chosen card's spell is cast without paying its
// mana cost (a played land has no mana cost regardless). It models the
// resolution-time "Choose an exiled card an opponent owns with a <kind> counter
// on it. You may play it this turn without paying its mana cost." activated
// ability (Dauthi Voidwalker).
//
// The choice is mandatory when at least one eligible card exists; the granted
// permission is optional to use. With no eligible card the effect is a legal
// no-op. The chosen card commonly rests in an opponent's exile bucket, and the
// granted per-card RuleEffectPlayFromZone authorizes Player through the ordinary
// cross-player exile play/cast machinery.
type PlayChosenExiledCard struct {
	Player                PlayerReference
	Zone                  zone.Type
	OwnerScope            PlayerRelation
	Counter               opt.V[counter.Kind]
	Duration              EffectDuration
	WithoutPayingManaCost bool
}

// CopyCard offers Player the chance to copy the card exiled by the resolving
// source under the object-scoped link named LinkID (the imprinted card). It is
// the enabling half of the imprint copy/cast idiom ("You may copy the exiled
// card. If you do, you may cast the copy without paying its mana cost." —
// Isochron Scepter, Spellbinder): resolution succeeds only when a card linked to
// this source under LinkID still rests in exile, so the following optional
// PlayLinkedExiledCard cast is gated (via prior-instruction acceptance) on a
// copy actually being available. Because a copy that is never cast ceases to
// exist (CR 707.12a), the consent step itself performs no observable action; the
// paired PlayLinkedExiledCard materializes and casts the copy.
type CopyCard struct {
	Player PlayerReference
	// LinkID is the object-scoped link key under which the source published its
	// imprinted card (for example, the imprint link established by the ETB
	// exile-from-hand choice).
	LinkID string
}

// PlayLinkedExiledCard casts the card exiled by the resolving source under the
// object-scoped link named LinkID (the imprinted card). With Copy set it casts a
// copy of that card (CR 707.12) rather than the card itself: the linked card
// stays in exile and a spell carrying its copiable values is put on the stack,
// ceasing to exist when it leaves the stack. WithoutPayingManaCost casts it for
// free. The cast chooses the first legal targets and modes with any X treated as
// 0; when no legal way to cast exists the effect is a legal no-op. It is the
// consequence half of the imprint copy/cast idiom, paired with a preceding
// CopyCard whose acceptance gates it.
type PlayLinkedExiledCard struct {
	Player PlayerReference
	// LinkID is the object-scoped link key under which the source published its
	// imprinted card.
	LinkID string
	// Copy casts a copy of the linked exiled card rather than the card itself,
	// leaving the original in exile.
	Copy bool
	// WithoutPayingManaCost casts the copy (or card) without paying its mana cost.
	WithoutPayingManaCost bool
}

// Sacrifice sacrifices one referenced permanent or every permanent in a
// referenced group. When neither is set, the controller's first permanent is
// used.
type Sacrifice struct {
	Object ObjectReference
	Group  GroupReference
	// ByItsController, when set, makes the referenced object's current
	// controller sacrifice it rather than requiring the ability's controller to
	// control it ("that creature's controller sacrifices it" — Animate Dead's
	// leaves-the-battlefield trigger). Without it a plain Sacrifice only affects
	// an object the ability controller still controls, so a reanimated creature
	// whose control had changed would incorrectly survive. It applies to the
	// single-object form only (Group unset).
	ByItsController bool
}

// SacrificePermanents causes the referenced player (or every player in a group)
// to choose and sacrifice the required number of eligible permanents during resolution.
type SacrificePermanents struct {
	Player      PlayerReference      // single player; zero if PlayerGroup is set
	PlayerGroup PlayerGroupReference // opponents or all players; zero if Player is set
	Amount      Quantity             // number of permanents to sacrifice
	Selection   Selection            // eligible permanent filter; zero = any permanent
	// All, when set, sacrifices every permanent each affected player controls
	// that matches Selection rather than a chosen Amount ("Each player sacrifices
	// all permanents they control that are one or more colors." — All Is Dust).
	// Amount is ignored when All is set, and no per-player choice is offered.
	All bool
	// AnyNumber, when set, lets each affected player choose any number of
	// permanents matching Selection to sacrifice, from none up to all eligible,
	// rather than a fixed Amount ("Sacrifice any number of lands, then add that
	// much {C}." — Mana Seism). Amount is ignored when AnyNumber is set. The
	// number actually sacrificed is reported as the instruction's resolved
	// amount, so a later count-scaled effect published off this instruction
	// ("add that much", "draw that many", "create that many") reads it through
	// DynamicAmountPreviousEffectResult. It is mutually exclusive with All.
	AnyNumber bool
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
	// PublishObjectBinding, when set, records each PublishLinked object by its
	// ObjectID even for a token (CardInstanceID == 0), the way
	// permanentObjectBindingRef binds it, rather than dropping tokens the way the
	// default permanentLinkedObjectRef does. Set it only when the downstream
	// reader resolves the sacrificed permanent by ObjectID through last-known
	// information (Braids, Arisen Nightmare reads the sacrificed permanent's card
	// types so each opponent's shared-card-type offer works when a token such as a
	// Treasure is sacrificed). Leave it unset when the downstream reader needs the
	// card instance itself, e.g. to return the sacrificed card from a zone
	// (Heart-Shaped Herb returns it from the graveyard by CardID), because a token
	// has no card instance to return. It is inert without PublishLinked.
	PublishObjectBinding bool
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
	// DiscardCount is the number of cards the discard alternative requires. Zero
	// means the common "discard a card" form (one card); it is only set above 1
	// for effects that demand more ("... unless they discard two cards." — Court
	// of Ambition's monarch escalation), so the ubiquitous one-card form stays
	// serialized identically.
	DiscardCount int
	// ControllerDrawEach, when set, draws one card for the effect's controller
	// for each affected player who takes the life loss rather than paying the
	// offered alternative ("For each opponent who doesn't, that player loses 2
	// life and you draw a card." — Braids, Arisen Nightmare). It couples the
	// controller's reward to the punisher's per-player outcome, so a player who
	// pays the alternative yields no draw while a player who takes the loss (by
	// choice or because they can't pay) yields one. Zero when the punisher grants
	// the controller no draw.
	ControllerDrawEach bool
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
//
// Chooser names the player who makes the ChooseUpTo selection when it is not the
// resolving controller. The zero value (PlayerReferenceNone) leaves the choice
// with the ability's controller; a non-empty reference (an event player on an
// "each player's upkeep" trigger, where the upkeep player chooses which of their
// own permanents to untap) redirects the prompt to that player. It is only
// consulted for the ChooseUpTo form.
type Untap struct {
	Object ObjectReference
	Group  GroupReference

	ChooseUpTo bool
	Amount     Quantity
	Chooser    PlayerReference
}

// SkipNextUntap marks the referenced permanent so it doesn't untap during its
// controller's next untap step (the "doesn't untap during its controller's next
// untap step" clause that follows a tap effect). The permanent stays tapped
// through one of its controller's untap steps and then untaps normally. A group
// form marks every permanent in Group instead (the mass "Lands you control don't
// untap during your next untap step." clause). Exactly one of Object or Group is
// set.
type SkipNextUntap struct {
	Object ObjectReference
	Group  GroupReference
}

// RemoveFromCombat removes the referenced creature from combat ("Remove target
// attacking creature you control from combat." — Reconnaissance). The permanent
// stops being an attacker or blocker: it deals and is dealt no further combat
// damage and its attack/block declarations are discarded. Object references the
// creature to remove.
type RemoveFromCombat struct {
	Object ObjectReference
}

// CounteredSpellDestination identifies the zone a countered spell's card is put
// into instead of its owner's graveyard, backing the CR 614-style replacement
// rider "If that spell is countered this way, put it [on top of its owner's
// library | into its owner's hand] instead of into that player's graveyard."
// (Memory Lapse, Lapse of Certainty, Remand). The zero value leaves a countered
// spell in its owner's graveyard. The exile destination is carried separately by
// CounterObject.ExileInstead, which predates this enum.
type CounteredSpellDestination uint8

const (
	// CounteredSpellGraveyard puts a countered spell into its owner's graveyard,
	// the default destination (CR 701.5g).
	CounteredSpellGraveyard CounteredSpellDestination = iota
	// CounteredSpellLibraryTop puts a countered spell on top of its owner's
	// library (Memory Lapse, Lapse of Certainty).
	CounteredSpellLibraryTop
	// CounteredSpellHand puts a countered spell into its owner's hand (Remand).
	CounteredSpellHand
)

// CounterObject counters a referenced spell or ability on the stack. When
// ExileInstead is set, a countered spell is exiled instead of being put into
// its owner's graveyard (CR 614-style replacement, e.g. Force of Negation).
// Destination redirects a countered spell to a non-graveyard zone other than
// exile; it is mutually exclusive with ExileInstead.
type CounterObject struct {
	Object       ObjectReference
	ExileInstead bool
	Destination  CounteredSpellDestination
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
//
// Chooser, when set, is the player who creates and controls the copy and chooses
// its new targets, rather than the resolving controller. It models the copy-chain
// family, where "that player or that permanent's controller ... may copy this
// spell and may choose a new target for that copy" (Chain Lightning, Chain Stasis,
// String of Disappearances): the affected target's controller controls the copy
// so its own iterative copy offer chains off the copy's new target. The zero
// value leaves the resolving controller as the copier, preserving prior behavior.
type CopyStackObject struct {
	Object              ObjectReference
	MayChooseNewTargets bool
	Chooser             opt.V[PlayerReference]
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
	// PublishLinked remembers each card that actually reaches exile under this
	// source-keyed set for a later instruction to select "from among them".
	PublishLinked LinkedKey
	// Counter names a named marker counter placed on each card exiled this way
	// once it reaches exile ("exile the top card of each player's library with a
	// collection counter on it.", Evelyn, the Covetous). The counter is recorded
	// in Game.ExileCounters, mirroring MoveCard.Counter, so the source's paired
	// play/cast-from-exile ability can later select "a card ... in exile with a
	// <name> counter on it". It is unset for every exile that places no counter.
	Counter opt.V[counter.Kind]
	// FaceDown exiles each card face down ("exile that many cards from the top of
	// your library face down.", Flamewar, Streetwise Operative). A face-down card
	// in exile hides its identity from every observer (CR 713); the zone records
	// the face-down state and clears it when the card leaves exile. It is false
	// for the ordinary face-up exile.
	FaceDown bool
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
	// MatchToDestinationRestRandomBottom moves only the matching card to
	// Destination and returns the other revealed cards to the library bottom in
	// random order.
	MatchToDestinationRestRandomBottom bool
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

// DiscardThenDraw has Player discard a player-chosen number of cards, then draw a
// number of cards equal to the number discarded plus DrawOffset. Max bounds the
// number that may be discarded: a positive Max means "up to Max cards" while a
// zero Max means "any number of cards". It models "discard {up to N|any number
// of} cards, then draw that many cards[ plus K]." The player-chosen count and the
// "that many" back-reference are not expressible through separate instructions,
// so the whole sequence resolves as one primitive.
type DiscardThenDraw struct {
	Player     PlayerReference
	Max        int
	DrawOffset int
}

// DiscardUnlessType has Player discard Amount cards unless they instead discard
// a single card of one of ExemptTypes. It models the "discard N cards unless you
// discard a <type> card." rider of the Thirst for Knowledge family: the
// controller may discard one exempt-type card to satisfy the effect, or discard
// Amount cards otherwise. ExemptTypes lists the card types whose disjunction
// waives the full discard. The player-chosen branch is not expressible through
// separate instructions, so the whole choice resolves as one primitive.
type DiscardUnlessType struct {
	Player      PlayerReference
	Amount      int
	ExemptTypes []types.Card
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
	// Filter, when present, restricts which of the looked-at cards the player may
	// put into their hand to those matching the Selection. It models the typed
	// optional dig "look at the top N cards. You may reveal a [filter] card from
	// among them and put it into your hand. Put the rest <remainder>." The zero
	// value imposes no filter, allowing any of the looked-at cards to be taken.
	Filter opt.V[Selection]
	// TakeUpTo makes Take an upper bound rather than an exact count: the player
	// puts from zero up to Take of the eligible cards into their hand, modeling
	// the optional "you may reveal" dig where taking nothing is allowed. The zero
	// value keeps Take an exact count.
	TakeUpTo bool
	// Reveal reports that each card put into the player's hand is revealed as it
	// is taken, modeling the "you may reveal ... and put it into your hand"
	// wording. The zero value puts the taken cards into hand without revealing
	// them.
	Reveal bool
	// Destination is the zone the taken cards move to. The zero value
	// (zone.None) puts them into the player's hand, the reveal-to-hand dig.
	// zone.Battlefield instead puts each taken card onto the battlefield under
	// the player's control, modeling "look at the top N cards of your library.
	// You may put a [filter] card from among them onto the battlefield. Put the
	// rest <remainder>." (Web of Life and Destiny, Elvish Rejuvenator).
	// zone.Library returns each taken card to the top of the player's library
	// without leaving the library zone, modeling "look at the top N cards of your
	// library. Put up to one of them on top of your library and the rest
	// <remainder>." (Thassa's Oracle). When more than one card is taken to the
	// library top, the first chosen card ends up on top. No other destination is
	// supported.
	Destination zone.Type
	// EntersTapped makes each card put onto the battlefield by a battlefield
	// Destination enter tapped, modeling the "... onto the battlefield tapped"
	// wording (Elvish Rejuvenator, Freestrider Lookout). It is meaningful only
	// when Destination is zone.Battlefield; the zero value enters untapped.
	EntersTapped bool
}

// PileSplit reveals the top Amount cards of the referenced player's library,
// has the separating player divide the revealed cards into two piles, has the
// choosing player pick one pile to keep, then moves the kept pile to Kept and
// the other pile to Other (both zones belonging to Player). It models the "Fact
// or Fiction" family: "Reveal the top N cards of your library. An opponent
// separates those cards into two piles. Put one pile into your hand and the
// other into your graveyard." SeparatorOpponent and ChooserOpponent select
// whether an opponent (rather than the controller) separates the piles and
// chooses which pile is kept; the unchosen actor role is the controller.
type PileSplit struct {
	Player            PlayerReference
	Amount            Quantity
	SeparatorOpponent bool
	ChooserOpponent   bool
	Kept              zone.Type
	Other             zone.Type
}

// RevealTopPartition reveals the top Amount cards of a referenced player's
// library, puts every revealed card matching Selection into that player's hand,
// and puts the rest into the Remainder destination (that player's graveyard or
// the bottom of their library). It models the closed "Reveal the top N cards of
// your library. Put all <type> cards revealed this way into your hand and the
// rest <remainder>." family (Borborygmos Enraged, Sift Through Sands, the Goblin
// Matron / tribal "reveal and gather" cards). Unlike Dig, every revealed card is
// turned face up publicly and the matching cards are taken without a choice, so
// the partition is fully deterministic. Remainder is DigRemainderGraveyard or
// DigRemainderLibraryBottom; the "in any order" and "in a random order" library-
// bottom riders share one placement.
type RevealTopPartition struct {
	Player    PlayerReference
	Amount    Quantity
	Selection Selection
	Remainder DigRemainder
}

// ImpulseExile exiles cards from the top of a player's library and lets the
// resolving controller play those cards for a bounded duration. Player is
// usually the resolving controller ("exile the top card of your library"), but
// may resolve to a target opponent so the controller plays the top card of an
// opponent's library ("exile the top card of target opponent's library ... You
// may play that card for as long as it remains exiled", Court of Locthwain).
type ImpulseExile struct {
	Player   PlayerReference
	Amount   Quantity
	Duration EffectDuration
	// SpendAnyMana, when set, lets the controller spend mana of any type to cast
	// the exiled cards ("mana of any type can be spent to cast it.", Court of
	// Locthwain). It carries onto the RuleEffectPlayFromZone permission.
	SpendAnyMana bool
	// Cast, when set, grants permission to *cast* the exiled cards ("you may cast
	// that card", Grenzo, Havoc Raiser) rather than to *play* them. A cast-only
	// grant lets the controller cast an exiled card as a spell but not play an
	// exiled land, so the handler emits a RuleEffectCastFromZone permission
	// instead of the play-permitting RuleEffectPlayFromZone. It is false for the
	// ordinary "you may play" impulse.
	Cast bool
	// PublishLinked, when set, remembers each exiled card under this source-keyed
	// linked set so a later ability can act on "cards exiled with this ..." (Court
	// of Locthwain's monarch free-cast reads the accumulated pool). It is empty
	// when the exiled cards need not be tracked.
	PublishLinked LinkedKey
}

// ExileLibraryUntilNonlandCast exiles cards from the top of Player's library one
// at a time until a nonland card is exiled (or the library empties), then lets
// Player cast that nonland card without paying its mana cost. The other cards
// exiled this way stay in exile. It models the single-effect family "Exile cards
// from the top of your library until you exile a nonland card. You may cast that
// card without paying its mana cost." The whole sequence is one primitive
// because the dig depth is the first nonland card and the free cast targets that
// same card, neither of which is expressible across separate instructions.
type ExileLibraryUntilNonlandCast struct {
	Player PlayerReference
}

// IterativeLibraryStop selects the name-based predicate that terminates an
// IterativeLibraryProcess loop.
type IterativeLibraryStop uint8

const (
	// IterativeLibraryStopChosenName stops when a processed card matches a card
	// name the player chose at the start of the process (Demonic Consultation).
	// The matching card is put into the recipient's hand; every other card
	// processed before it stays exiled. Reaching an empty library without a
	// match leaves the whole library exiled.
	IterativeLibraryStopChosenName IterativeLibraryStop = iota
	// IterativeLibraryStopDuplicateName stops when a processed card shares its
	// name with another card already processed this way (Tainted Pact). The
	// duplicate stays exiled and the process ends.
	IterativeLibraryStopDuplicateName
	iterativeLibraryStopCount
)

// IterativeLibraryProcess exiles or reveals cards from the top of a player's
// library one at a time, remembering every card processed during this single
// resolution, until a name-based stop predicate fires. It is the generic
// iterative library processor shared by Tainted Pact and Demonic Consultation.
//
// The processed-name history is scoped to one execution of this primitive, so
// independent copies of the same spell never share history and no shuffle
// occurs. When the library empties before the stop predicate fires the process
// simply ends with every processed card left exiled.
//
//   - ChooseName: before processing, the player names a card. The chosen name
//     feeds the IterativeLibraryStopChosenName predicate.
//   - PreExile: cards exiled from the top before the loop begins, without being
//     revealed or offered to hand (Demonic Consultation's "top six cards").
//   - Reveal: each processed card is revealed as public information before it is
//     routed (Demonic Consultation). When false, cards are exiled directly
//     (Tainted Pact) without a reveal event.
//   - OptionalTake: after a non-duplicate card is processed, the player may put
//     it into hand to end the process (Tainted Pact's "you may put that card
//     into your hand"). When declined the process continues.
//   - AllowAbsentName: the naming step offers an extra "a card name not in this
//     library" option that maps to a sentinel the chosen-name predicate never
//     matches, so the player can deliberately name an absent card and exile the
//     entire remaining library (Demonic Consultation's defining line). It is
//     only meaningful with the chosen-name stop, where the actual named card is
//     irrelevant once matching fails, and it keeps the naming step reachable
//     even when the library is empty.
//   - Stop: which name-based predicate terminates the loop.
type IterativeLibraryProcess struct {
	Player          PlayerReference
	Stop            IterativeLibraryStop
	PreExile        Quantity
	ChooseName      bool
	Reveal          bool
	OptionalTake    bool
	AllowAbsentName bool
}

// ExileTopEachLibraryCastFree exiles the top Amount cards of every player's
// library into their owners' exile as one simultaneous batch, then lets the
// resolving controller cast any number of those just-exiled cards without paying
// their mana costs, casting each under the controller's control regardless of
// which player's library it came from. Cards the controller declines to cast
// stay exiled. It models the attack-trigger family "exile the top card of each
// player's library, then you may cast any number of spells from among those
// cards without paying their mana costs." (Etali, Primal Storm). Amount is the
// per-library exile count, one for the printed card.
type ExileTopEachLibraryCastFree struct {
	Amount Quantity
}

// HideawayExile implements the Hideaway N enters-the-battlefield action (CR
// 702.75a): the resolving controller looks at the top Amount cards of their
// library, exiles one of them face down linked to the source permanent, and
// puts the rest on the bottom of their library in a random order. The exiled
// card is played later by the source permanent's Hideaway activated ability
// through PlayHideawayCard, which reads the same source-scoped link.
type HideawayExile struct {
	Amount Quantity
}

// PlayHideawayCard implements the "you may play the exiled card without paying
// its mana cost" half of the Hideaway mechanic (CR 702.75c). The resolving
// controller may play the card the source permanent exiled face down with its
// HideawayExile action, casting it as a spell or putting it onto the
// battlefield as a land without paying its mana cost. The enclosing
// instruction's Condition gates the play on the printed Hideaway condition and
// its Optional flag carries the "may".
type PlayHideawayCard struct{}

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
	Dread bool
	// Cloak, when set, cloaks the card instead of manifesting it (CR 701.56):
	// the face-down 2/2 additionally has ward {2}. The turn-face-up rule is
	// unchanged from manifest (mana cost if the hidden card is a creature).
	Cloak  bool
	Player PlayerReference
	// PublishLinked, when set, remembers the manifested permanent as an
	// object-scoped linked object so a later instruction can reference it ("put
	// three +1/+1 counters on that creature", Weight Room). It is empty when no
	// later instruction references the manifested creature.
	PublishLinked LinkedKey
}

// Goad goads the referenced creature, or every creature in the referenced group.
type Goad struct {
	Object ObjectReference
	Group  GroupReference
	// RestOfGame goads each affected creature permanently rather than until the
	// goading player's next turn, backing "goaded for the rest of the game" (Life
	// of the Party). The goad persists until the creature leaves the battlefield
	// (CR 701.38, the card's ruling), so it is not cleared by the goading
	// player's turn-based expiry. It is false for the ordinary turn-limited goad
	// keyword action.
	RestOfGame bool
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
	// AllKinds removes every counter of every kind from the object, modeling the
	// kind-agnostic mass form "remove all counters from <permanent>" (Vampire
	// Hexmage). When set, Amount, CounterKind, and ChooseKind are ignored.
	AllKinds bool
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

// Regenerate sets up a regeneration shield on one referenced permanent
// ("Regenerate target creature.") or on every permanent in a referenced group
// ("Regenerate each creature you control."). Exactly one of Object or Group is
// set.
type Regenerate struct {
	Object ObjectReference
	Group  GroupReference
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

// CreateReflexiveTrigger puts a reflexive triggered ability on the stack
// (CR 603.11). It backs "When you do, <effect>." following an optional enabling
// action in the same resolution. The enclosing instruction is gated on the
// enabling action's published result, so the primitive runs (queues the
// reflexive trigger) only when the enabling action was performed. The trigger is
// put on the stack the next time a player would receive priority, with its
// targets chosen then — after the enabling action has resolved. This differs
// from CreateDelayedTrigger, which waits for a future timing or game event.
type CreateReflexiveTrigger struct {
	Trigger ReflexiveTriggerDef
}

// ReflexiveTriggerDef is the card-definition-side data for a reflexive triggered
// ability. Its Content is put on the stack as an ordinary triggered ability with
// its own targets, chosen when the trigger is put on the stack.
type ReflexiveTriggerDef struct {
	Content AbilityContent
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
//
// AnyTarget names a single shield recipient through an any-target target slot
// that may be chosen as either a player or a permanent ("Prevent the next N
// damage that would be dealt to any target this turn."). It resolves to whichever
// the controller chose: a player recipient sets the shield's player, a permanent
// recipient sets its permanent. AnyTarget is mutually exclusive with Object,
// Player, Global, and BySource.
//
// OneShot marks a shield that prevents one qualifying damage event and then
// expires, modeling the "The next time a source would deal damage ... this turn,
// prevent that damage." replacement (Circle of Protection, Rune of Protection).
// It is combined with All so the single event is prevented in full. SourceColors,
// when non-empty, restricts the shield to damage from a source of one of the
// listed colors ("a white source of your choice"); an empty slice matches a
// source of any color.
//
// RedirectPreventedToSourceController makes a one-shot shield deal the amount it
// prevents to the prevented source's controller, with the shield's own source as
// the damage source ("If damage is prevented this way, Deflecting Palm deals that
// much damage to that source's controller."). It is only valid alongside OneShot
// and All (the whole prevented event is redirected).
type PreventDamage struct {
	Amount                              Quantity
	Object                              ObjectReference
	Player                              PlayerReference
	AnyTarget                           DamageRecipient
	SourceColors                        []color.Color
	All                                 bool
	CombatOnly                          bool
	BySource                            bool
	Global                              bool
	OneShot                             bool
	RedirectPreventedToSourceController bool
}

// AddExtraPhases inserts additional phases into the current turn (CR 505.5,
// 506.2). It models "After this main phase, there is an additional combat
// phase[ followed by an additional main phase]." (Aggravated Assault, Aurelia
// the Warleader, World at War, Combat Celebrant) and "there is an additional
// beginning phase after this phase." (Sphinx of the Second Sun, Temple of
// Atropos, Cyclonus, Cybertronian Fighter). Combat queues an extra combat
// phase; Main queues an extra main phase after it; Beginning queues an extra
// beginning phase (untap, upkeep, and draw steps). The runtime appends the
// queued phases to TurnState.ExtraPhases, which the turn loop drains in order.
type AddExtraPhases struct {
	Combat    bool
	Main      bool
	Beginning bool
}

// RollDie rolls a single fair die with Sides faces and publishes the rolled
// value (1..Sides) as the instruction's resolved amount (CR 706). It backs
// "roll a d20" and similar dice mechanics; a later instruction consumes the
// result via a DynamicAmountPreviousEffectResult amount keyed to this
// instruction's PublishResult ("...equal to the result").
type RollDie struct {
	Sides int
}
