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
}

// PreventionShield prevents an amount of future damage to a player or
// permanent.
type PreventionShield struct {
	ID          id.ID
	Controller  PlayerID
	Player      PlayerID
	PermanentID id.ID
	Amount      int
	Duration    EffectDuration
	CreatedTurn int
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
	CounterUseRecipientController bool
	DamageMultiplier              int
	DamageAddend                  int
	DamageSourceColors            []color.Color
	DamageExcludeSource           bool
	EntersTapped                  bool
	EntersWithCounters            []CounterPlacement

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
}

// EntryTypeChoiceKey is the ChoiceKey under which an entry-time creature-type
// choice is stored on a Permanent's EntryChoices map. Abilities referencing "the
// chosen type" read the result from this key.
const EntryTypeChoiceKey = ChoiceKey("oracle-entry-type")

// EntryColorChoiceKey is the ChoiceKey under which an entry-time color choice is
// stored on a Permanent's EntryChoices map. Mana abilities that add "one mana of
// the chosen color" read the result from this key.
const EntryColorChoiceKey = ChoiceKey("oracle-entry-color")
