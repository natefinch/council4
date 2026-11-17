package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// ManaState is the engine-owned pure prediction and canonical activation seam.
// ProjectResources is read only; ActivatePlannedMana is used only by Apply.
type ManaState interface {
	State
	ProjectResources(game.ResourceProjection) State
	ManaActivationCandidates(game.PlayerID) []ManaActivationCandidate
	// ForecastMana evaluates a cost-paid candidate. False means its effect or
	// replacement/trigger context is opaque, never a predicted success.
	ForecastMana(game.PlayerID, ManaActivationCandidate) ([]ManaProduction, game.ResourceProjection, bool)
	ActivatePlannedMana(game.PlayerID, ManaActivation) bool
	CounterPlacement(game.PlayerID, id.ID, counter.Kind, int) (int, bool)
	PrepareManaEvent(game.Event) game.Event
	OrdinaryManaOnly(game.PlayerID) bool
	ManaSourceSnapshot(id.ID) game.ObjectSnapshot
	LifePaymentLoss(game.PlayerID, int) int
	RidersAfterInput(game.PlayerID, id.ID, map[mana.Unit]int) []game.ManaRiderInstance
	// ConsumePaymentRiders consumes riders on the pool mana a non-spell payment
	// spent. A turn-face-up payment is a special action, not an activation of
	// its source's ability, so it qualifies riders by its TurnFaceUpCost.
	ConsumePaymentRiders(game.PlayerID, *game.Permanent, TurnFaceUpCost, map[mana.Unit]int, map[mana.Unit]int)
	ManaAbilityLabel(id.ID, int) string
	UntapResources(id.ID) (game.PermanentResources, bool)
	ForecastCardCost(game.PlayerID, id.ID, zone.Type, zone.Type, bool) (game.ResourceProjection, bool)
}

// ManaProduction records exact units and their spend-linked semantics.
type ManaProduction struct {
	Unit   mana.Unit
	Amount int
	Rider  *game.ManaRiderInstance
}

// ManaActivationCandidate is a structurally modeled activation in the current
// resource view. Cost is paid before the forecast effects are projected.
type ManaActivationCandidate struct {
	Source       id.ID
	AbilityIndex int
	SourceZone   zone.Type
	Request      Request
	Choices      map[game.ChoiceKey]game.ResolutionChoiceResult
	Automatic    bool
	Reusable     bool
	Sacrifice    bool
	Body         *game.ManaAbility
	CardID       id.ID
}
