package game

import (
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
)

// ResolutionChoiceColorSource identifies dynamic sources for color choice
// options.
type ResolutionChoiceColorSource int

// Resolution choice color source values identify dynamic color-choice sources.
const (
	ResolutionChoiceColorSourceStatic ResolutionChoiceColorSource = iota
	ResolutionChoiceColorSourceCommanderIdentity
)

// ResolutionChoice describes a bounded value-producing choice made during
// resolution.
type ResolutionChoice struct {
	Kind ResolutionChoiceKind

	// Prompt overrides the default choice prompt.
	Prompt string

	// Player is the choosing player when UsePlayer is true; otherwise the stack
	// object's controller chooses.
	Player    PlayerID
	UsePlayer bool

	ColorSource    ResolutionChoiceColorSource
	Colors         []mana.Color
	CardTypes      []types.Card
	PlayerRelation PlayerRelation
	Zone           zone.Type
}

// ResolutionChoiceResult stores the selected value from a ResolutionChoice.
type ResolutionChoiceResult struct {
	Kind     ResolutionChoiceKind
	Color    mana.Color
	CardType types.Card
	Player   PlayerID
	CardID   id.ID
}

// ResolutionPayment describes an optional cost that may be paid during
// resolution (CR 608.2c, CR 117.12).
type ResolutionPayment struct {
	Prompt          string
	ManaCost        opt.V[cost.Mana]
	AdditionalCosts []cost.Additional
	XValue          int
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

	ReplaceToZone      zone.Type
	EntersTapped       bool
	EntersWithCounters []CounterPlacement
}
