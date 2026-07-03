package agent

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

// Choice heuristic tuning. A player wants to keep lands coming while developing,
// and keep cards it can cast soon; cards far above its current mana are better
// off on the bottom (scry) or in the graveyard (surveil).
const (
	scryLandKeepBelow = 5
	scryCastableSlack = 1
)

const (
	placementKeepTop   = 0
	placementElsewhere = 1
)

// ChooseChoice implements rules.ChoiceAgent. It makes card-aware decisions for
// the choices that carry card information (scry, surveil, and card- or
// permanent-loss payments), and falls back to the BaselineStrategy for choices
// that do not expose enough information to decide well (mode selection, trigger
// ordering, and similar).
func (GenericStrategy) ChooseChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceScry, game.ChoiceSurveil:
		return placementChoice(obs, request)
	case game.ChoicePayment:
		if selection, ok := loseLeastValuable(request); ok {
			return selection
		}
	case game.ChoiceTarget:
		if selection, ok := threatAwareTargetChoice(obs, request); ok {
			return selection
		}
	case game.ChoicePlayer:
		if selection, ok := threatAwarePlayerChoice(obs, request); ok {
			return selection
		}
	default:
	}
	return BaselineStrategy{}.ChooseChoice(obs, request)
}

// placementChoice decides whether to keep the scryed/surveiled card on top
// (index 0) or move it to the bottom/graveyard (index 1).
func placementChoice(obs rules.PlayerObservation, request game.ChoiceRequest) []int {
	if len(request.Options) < 2 {
		return BaselineStrategy{}.ChooseChoice(obs, request)
	}
	if keepOnTop(obs, request.Subject) {
		return []int{placementKeepTop}
	}
	return []int{placementElsewhere}
}

// keepOnTop reports whether the agent should keep the subject card on top of its
// library. Unknown subjects are kept (the conservative default).
func keepOnTop(obs rules.PlayerObservation, subject opt.V[game.ChoiceCardInfo]) bool {
	if !subject.Exists {
		return true
	}
	info := subject.Val
	lands := controlledLandCount(obs)
	if slices.Contains(info.Types, types.Land) {
		// Keep lands coming until the mana base is developed; otherwise this is
		// an excess land and is better on the bottom.
		return lands < scryLandKeepBelow
	}
	// Keep a spell the agent can cast soon; bury one it cannot afford yet.
	return info.ManaValue <= lands+scryCastableSlack
}

// loseLeastValuable selects the cheapest cards or permanents to give up for a
// payment choice whose options all identify a card, so the agent keeps its more
// valuable cards. It reports ok=false when the choice is not a pure card/
// permanent loss (e.g. a hybrid mana-or-life payment) so the caller can fall
// back.
func loseLeastValuable(request game.ChoiceRequest) ([]int, bool) {
	if request.MinChoices < 1 || request.MinChoices > len(request.Options) {
		return nil, false
	}
	ordered := make([]game.ChoiceOption, 0, len(request.Options))
	for i := range request.Options {
		if !request.Options[i].Card.Exists {
			return nil, false
		}
		ordered = append(ordered, request.Options[i])
	}
	slices.SortStableFunc(ordered, func(a, b game.ChoiceOption) int {
		if a.Card.Val.ManaValue != b.Card.Val.ManaValue {
			return a.Card.Val.ManaValue - b.Card.Val.ManaValue
		}
		return a.Index - b.Index
	})
	selection := make([]int, 0, request.MinChoices)
	for i := 0; i < request.MinChoices; i++ {
		selection = append(selection, ordered[i].Index)
	}
	return selection, true
}

func controlledLandCount(obs rules.PlayerObservation) int {
	count := 0
	battlefield := obs.Battlefield()
	for i := range battlefield {
		if battlefield[i].Controller == obs.Player && slices.Contains(battlefield[i].Types, types.Land) {
			count++
		}
	}
	return count
}
