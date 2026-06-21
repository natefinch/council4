package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CounterPlacement describes counters a permanent enters with.
type CounterPlacement struct {
	Kind   counter.Kind
	Amount int
	// AmountFromX places a number of counters equal to the value of X chosen for
	// the entering permanent's spell (CR 107.3, CR 614). Amount is ignored when
	// it is set; a permanent that entered without a cast X (a copy or a
	// put-onto-the-battlefield effect) enters with zero such counters. It backs
	// "This creature enters with X +1/+1 counters on it." (Walking Ballista,
	// Hangarback Walker, Endless One).
	AmountFromX bool
}

// ConditionalCounterPlacement places Amount counters of Kind on an entering
// permanent only when its (copied) card types include IfType. It backs the
// conditional copiable counter riders of enters-as-copy replacements, such as
// Spark Double's "it enters with an additional +1/+1 counter on it if it's a
// creature" and "it enters with an additional loyalty counter on it if it's a
// planeswalker".
type ConditionalCounterPlacement struct {
	Kind   counter.Kind
	Amount int
	IfType types.Card
}

// PreventionShield prevents an amount of future damage to a player or
// permanent. When All is set the shield has no fixed capacity and prevents
// every qualifying event for its duration; when CombatOnly is set it prevents
// only combat damage. SourcePermanentID, when non-zero, prevents damage dealt
// BY that permanent rather than damage dealt TO PermanentID/Player.
type PreventionShield struct {
	ID                id.ID
	Controller        PlayerID
	Player            PlayerID
	PermanentID       id.ID
	SourcePermanentID id.ID
	Amount            int
	All               bool
	CombatOnly        bool
	Duration          EffectDuration
	CreatedTurn       int
}

// ReplacementDecision records deterministic ordering for competing replacement
// or prevention effects.
type ReplacementDecision struct {
	Player       PlayerID
	Options      []string
	Selected     []int
	UsedFallback bool
}

// StateTriggerKey identifies one state-triggered ability latch (CR 603.8).
type StateTriggerKey struct {
	SourceObjectID id.ID
	SourceCardID   id.ID
	AbilityIndex   int
}

// ResolutionChoiceKind classifies a value chosen during spell or ability
// resolution (CR 608.2c, CR 609.3).
type ResolutionChoiceKind int

// Resolution choice kind values classify value-producing choices.
const (
	ResolutionChoiceNone ResolutionChoiceKind = iota
	ResolutionChoiceMana
	ResolutionChoiceCardType
	ResolutionChoicePlayer
	ResolutionChoiceCard
	// ResolutionChoiceSubtype chooses a subtype (such as a creature type) from
	// the subtypes defined for SubtypeOfType (CR 614.12, CR 205.3).
	ResolutionChoiceSubtype
	// ResolutionChoiceNumber chooses one integer in the inclusive MinNumber to
	// MaxNumber range.
	ResolutionChoiceNumber
)

// ResolutionChoiceColorSource identifies dynamic sources for color choice
// options.
type ResolutionChoiceColorSource int

// Resolution choice color source values identify dynamic color-choice sources.
const (
	ResolutionChoiceColorSourceStatic ResolutionChoiceColorSource = iota
	ResolutionChoiceColorSourceCommanderIdentity
	// ResolutionChoiceColorSourceFixedOrEntryChosen offers a fixed color (Colors)
	// together with the color chosen as the source permanent entered, read from
	// the stack object's seeded entry choice under EntryChoiceKey. It models the
	// composite "Add {C} or one mana of the chosen color." (the Gate/Thriving land
	// cycle).
	ResolutionChoiceColorSourceFixedOrEntryChosen
	// ResolutionChoiceColorSourceLandsProduce offers every color of mana that a
	// land matching PlayerRelation (relative to the choosing player) could
	// currently produce. It models "Add one mana of any color that a land you
	// control / an opponent controls could produce." (Reflecting Pool, Exotic
	// Orchard, Fellwar Stone). The candidate colors are recomputed from the
	// battlefield at resolution; an empty set leaves the ability unactivatable
	// (CR 605.1a).
	ResolutionChoiceColorSourceLandsProduce
	// ResolutionChoiceColorSourceLinkedExileColors offers each color of the card
	// linked to the source permanent under LinkID — the card imprinted by an
	// optional enter-the-battlefield exile from hand. It models "Add one mana of
	// any of the exiled card's colors." (Chrome Mox). The colors are recomputed
	// from the linked card at resolution (CR 106.6): a missing, declined, or
	// colorless imprint yields an empty set, leaving the ability unactivatable
	// (CR 605.1a, CR 202.2), while a multicolored imprint offers exactly its
	// colors. The link is scoped to the permanent's object identity so a
	// re-entered object has no imprint until it imprints again.
	ResolutionChoiceColorSourceLinkedExileColors
	// ResolutionChoiceColorSourceControlledPermanentColors offers every color
	// found among the colors of the permanents the choosing player controls that
	// match Selection. It models "Add one mana of any color among <permanents>
	// you control." (Mox Amber's "legendary creatures and planeswalkers you
	// control", Plaza of Heroes' "legendary permanents you control"). The
	// candidate colors are the union of the matching permanents' colors,
	// recomputed from the battlefield at resolution; a board with no matching
	// colored permanent yields an empty set and leaves the ability unactivatable
	// (CR 605.1a). Colorless ({C}) is never offered because a permanent's colors
	// are only the five colors (CR 105.2, CR 202.2).
	ResolutionChoiceColorSourceControlledPermanentColors
)

// ResolutionChoice describes a bounded value-producing choice made during
// resolution.
type ResolutionChoice struct {
	Kind ResolutionChoiceKind

	// Prompt overrides the default choice prompt.
	Prompt string

	// Player is the choosing player when UsePlayer is true; otherwise the stack
	// object's controller chooses.
	Player          PlayerID
	UsePlayer       bool
	PlayerReference *PlayerReference

	ColorSource    ResolutionChoiceColorSource
	Colors         []mana.Color
	CardTypes      []types.Card
	PlayerRelation PlayerRelation
	Zone           zone.Type

	// Selection constrains which permanents a dynamic color source reads. It is
	// consulted by ResolutionChoiceColorSourceControlledPermanentColors, which
	// offers the union of colors of the choosing player's permanents matching it.
	Selection *Selection

	// SubtypeOfType names the card type whose defined subtypes are the candidates
	// for a ResolutionChoiceSubtype choice, as in "choose a creature type."
	// (types.Creature). It is consulted only by ResolutionChoiceSubtype.
	SubtypeOfType types.Card

	// EntryChoiceKey names the source permanent's entry-time choice that a
	// dynamic color source reads (CR 614.12). It is consulted by
	// ResolutionChoiceColorSourceFixedOrEntryChosen.
	EntryChoiceKey ChoiceKey

	// LinkID names the linked object the choice's color source reads. It is
	// consulted by ResolutionChoiceColorSourceLinkedExileColors, which offers the
	// colors of the card linked to the source permanent under this key (the
	// imprinted exiled card).
	LinkID string

	// IncludeColorless additionally offers colorless ({C}) for a
	// ResolutionChoiceColorSourceLandsProduce choice when a matching land could
	// produce it. It distinguishes the "any type" wording (Reflecting Pool) from
	// "any color" (Exotic Orchard), which offers only colored mana.
	IncludeColorless bool
	MinNumber        int
	MaxNumber        int
}

// ResolutionChoiceResult stores the selected value from a ResolutionChoice.
type ResolutionChoiceResult struct {
	Kind     ResolutionChoiceKind
	Color    mana.Color
	CardType types.Card
	Subtype  types.Sub
	Player   PlayerID
	CardID   id.ID
	Number   int
}

// ResolutionPayment describes an optional cost that may be paid during
// resolution (CR 608.2c, CR 117.12).
type ResolutionPayment struct {
	Prompt   string
	Payer    opt.V[PlayerReference]
	ManaCost opt.V[cost.Mana]
	// DynamicGenericManaCost is a generic mana amount evaluated as the payment
	// instruction resolves. Negative values are treated as zero.
	DynamicGenericManaCost opt.V[*DynamicAmount]
	// ManaCostMultiplier repeats the fixed ManaCost by an amount evaluated as
	// the payment instruction resolves. Negative values are treated as zero.
	ManaCostMultiplier opt.V[*DynamicAmount]
	AdditionalCosts    []cost.Additional
	XValue             int
}

// ReplacementEffect is a runtime replacement effect that changes a future event
// before it happens (CR 614). This first generic slice covers zone destination
// changes and enters-the-battlefield modifiers; specialized replacement paths
// such as commander replacement and regeneration remain rules-owned.
type ReplacementEffect struct {
	ID             id.ID
	Controller     PlayerID
	SourceObjectID id.ID
	SourceCardID   id.ID
	Description    string

	Duration    EffectDuration
	CreatedTurn int

	MatchEvent EventKind

	ControllerFilter TriggerControllerFilter

	MatchFromZone bool
	FromZone      zone.Type
	MatchToZone   bool
	ToZone        zone.Type

	// Condition gates this replacement against the in-flight event.
	Condition opt.V[Condition]

	ReplaceToZone                 zone.Type
	ShuffleIntoLibrary            bool
	RevealSource                  bool
	TokenMultiplier               int
	CounterMultiplier             int
	CounterAddend                 int
	MatchCounterKind              bool
	CounterKindFilter             counter.Kind
	CounterRecipientTypes         []types.Card
	CounterRecipientAnyPermanent  bool
	CounterUseRecipientController bool
	DamageMultiplier              int
	DamageAddend                  int
	DamageSourceColors            []color.Color
	DamageExcludeSource           bool

	// LifeGainMultiplier multiplies a single "you would gain life" event by the
	// replacement's controller before the life is gained (CR 614), backing "If
	// you would gain life, you gain twice that much life instead." (Boon
	// Reflection, Rhox Faithmender, Alhammarret's Archive). LifeGainAddend then
	// adds a fixed amount, backing "you gain that much life plus N instead."
	// (Angel of Vitality, Heron of Hope). A multiplier of zero or one with a zero
	// addend leaves life gain unchanged. Both apply only when the gaining player
	// is the replacement's controller.
	LifeGainMultiplier int
	LifeGainAddend     int

	EntersTapped       bool
	EntersWithCounters []CounterPlacement

	// EntryColorChoice marks an enters-the-battlefield replacement that prompts
	// the controller to choose a color as the permanent enters (CR 614.12), such
	// as "As this artifact enters, choose a color." The chosen color is stored on
	// the permanent under EntryColorChoiceKey for later abilities to read.
	EntryColorChoice bool

	// EntryColorChoiceExclude is a single forbidden color removed from the
	// entry-time color prompt, as in "As this land enters, choose a color other
	// than white." (the Gate/Thriving land cycle). It is empty when the choice is
	// unconstrained. It is only meaningful when EntryColorChoice is true.
	EntryColorChoiceExclude mana.Color

	// EntryTypeChoice marks an enters-the-battlefield replacement that prompts the
	// controller to choose a creature type as the permanent enters (CR 614.12),
	// as in "As this creature enters, choose a creature type." The chosen subtype
	// is stored on the permanent under EntryTypeChoiceKey for later abilities to
	// read.
	EntryTypeChoice bool

	// EntersTappedOthers marks a continuous static enters-tapped replacement that
	// taps a group of OTHER permanents as they enter (Authority of the Consuls),
	// as opposed to the self form printed on the entering permanent. It is
	// registered into Game.ReplacementEffects while its source is on the
	// battlefield and matched against every entering permanent that satisfies
	// ControllerFilter and EntersTappedTypes.
	EntersTappedOthers bool

	// EntersTappedTypes restricts an EntersTappedOthers replacement to entering
	// permanents that have any of these card types. It is empty when every
	// entering permanent is tapped ("Permanents ... enter tapped.").
	EntersTappedTypes []types.Card

	// CreateOneOfEachTokens replaces the creation of a token whose name matches
	// one of these definitions with the creation of one of each listed token
	// (Academy Manufactor: "If you would create a Clue, Food, or Treasure token,
	// instead create one of each."). It is empty for every other replacement.
	CreateOneOfEachTokens []*CardDef

	// DrawFromEmptyLibraryWins replaces an attempt to draw a card from an empty
	// library by the controller with that controller winning the game (CR 104.2,
	// CR 614). It backs "If you would draw a card while your library has no cards
	// in it, you win the game instead." (Laboratory Maniac, Jace, Wielder of
	// Mysteries). It is registered while its source is on the battlefield and
	// consulted when the controller would otherwise lose to a failed draw.
	DrawFromEmptyLibraryWins bool

	// EntersAsCopy marks a self enters-the-battlefield replacement that has the
	// permanent enter as a copy of another permanent chosen as it enters (CR 706,
	// CR 614), as in "You may have this creature enter the battlefield as a copy
	// of any creature on the battlefield." (Clone). The controller chooses one
	// permanent matching EntersAsCopySelection; the entering permanent's copiable
	// values are overlaid with the chosen permanent's via a layer-1 continuous
	// effect that lasts as long as the copy is on the battlefield.
	EntersAsCopy bool

	// EntersAsCopyOptional marks the "You may have ..." form of an EntersAsCopy
	// replacement: the controller is first asked whether to copy at all and the
	// permanent enters as itself if they decline. It is false for the mandatory
	// "this creature enters as a copy of ..." form.
	EntersAsCopyOptional bool

	// EntersAsCopySelection restricts which permanents may be copied by an
	// EntersAsCopy replacement ("any creature on the battlefield", "any nonland
	// permanent on the battlefield", "a creature you control"). It is nil for
	// every other replacement and only consulted when EntersAsCopy is true.
	EntersAsCopySelection *Selection

	// EntersAsCopyNotLegendary applies the "except it isn't legendary" copiable
	// rider (CR 706.9c) by dropping the legendary supertype from the copied
	// values. It is only meaningful when EntersAsCopy is true.
	EntersAsCopyNotLegendary bool

	// EntersAsCopyAddTypes applies the "except it's an <type> in addition to its
	// other types" copiable rider (Phyrexian Metamorph) by adding these card
	// types to the copied values. It is empty for every other replacement and
	// only consulted when EntersAsCopy is true.
	EntersAsCopyAddTypes []types.Card

	// EntersAsCopyConditionalCounters applies the conditional copiable counter
	// riders of an enters-as-copy replacement, placing additional counters on the
	// copy based on the copied card's types (Spark Double: "+1/+1 counter if it's
	// a creature" and "loyalty counter if it's a planeswalker"). It is empty for
	// every other replacement and only consulted when EntersAsCopy is true.
	EntersAsCopyConditionalCounters []ConditionalCounterPlacement

	// EntersAsCopyUntilEndOfTurn scopes an enters-as-copy replacement's copy
	// effect to end of turn rather than as long as the permanent stays on the
	// battlefield (Cursed Mirror's "become a copy ... until end of turn"). It is
	// only consulted when EntersAsCopy is true.
	EntersAsCopyUntilEndOfTurn bool

	// EntersAsCopyAddKeywords lists keywords granted to the copy by the "except it
	// has <keyword>" rider of an enters-as-copy replacement (Cursed Mirror's
	// haste). It is empty for every other replacement and only consulted when
	// EntersAsCopy is true.
	EntersAsCopyAddKeywords []Keyword

	// DrawCardMultiplier replaces a single "draw a card" event by the controller
	// with drawing this many cards instead (CR 614). It backs the draw-doubling
	// replacement "If you would draw a card, draw two cards instead." A value of
	// zero or one leaves draws unchanged. It is registered while its source is on
	// the battlefield.
	DrawCardMultiplier int

	// DrawCardExceptFirstInDrawStep exempts the controller's first draw in each
	// of their own draw steps from DrawCardMultiplier ("If you would draw a card
	// except the first one you draw in each of your draw steps, draw two cards
	// instead.", Teferi's Ageless Insight). It is only meaningful when
	// DrawCardMultiplier is greater than one.
	DrawCardExceptFirstInDrawStep bool

	// ContinuousZoneRedirect marks a continuous static replacement that redirects
	// a card (or permanent) headed for a graveyard to a different zone (CR 614),
	// as on "If a card would be put into a graveyard from anywhere, exile it
	// instead." (Leyline of the Void, Samurai of the Pale Curtain). It is
	// registered into Game.ReplacementEffects while its source is on the
	// battlefield and matched against every card the event moves, using
	// RedirectOwnerFilter and RedirectTypeFilter. The self form printed on a
	// single card ("If THIS would be put into a graveyard, ...") is not marked
	// and is handled by the static self-zone replacement path instead.
	ContinuousZoneRedirect bool

	// RedirectOwnerFilter restricts a ContinuousZoneRedirect replacement by the
	// owner of the moving card relative to the replacement's controller: You
	// watches the controller's own graveyard, Opponent an opponent's, and Any
	// every player's. It is only meaningful when ContinuousZoneRedirect is true.
	RedirectOwnerFilter TriggerControllerFilter

	// RedirectTypeFilter restricts a ContinuousZoneRedirect replacement to moving
	// cards that have any of these card types ("an instant or sorcery card").
	// It is empty when every card is redirected. It is only meaningful when
	// ContinuousZoneRedirect is true.
	RedirectTypeFilter []types.Card
}

// EntryTypeChoiceKey is the ChoiceKey under which an entry-time creature-type
// choice is stored on a Permanent's EntryChoices map. Abilities referencing "the
// chosen type" read the result from this key.
const EntryTypeChoiceKey = ChoiceKey("oracle-entry-type")

// EntryColorChoiceKey is the ChoiceKey under which an entry-time color choice is
// stored on a Permanent's EntryChoices map. Mana abilities that add "one mana of
// the chosen color" read the result from this key.
const EntryColorChoiceKey = ChoiceKey("oracle-entry-color")
