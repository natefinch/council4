package rules

import (
	"fmt"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func initializeReadAhead(e *Engine, g *game.Game, permanent *game.Permanent, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if !permanentHasSubtype(g, permanent, types.Saga) || !hasKeyword(g, permanent, game.ReadAhead) {
		return
	}
	final := finalSagaChapter(g, permanent)
	if final <= 0 {
		return
	}
	options := make([]game.ChoiceOption, final)
	for chapter := 1; chapter <= final; chapter++ {
		options[chapter-1] = game.ChoiceOption{
			Index: chapter,
			Label: fmt.Sprintf("Chapter %d", chapter),
		}
	}
	chosen := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           effectiveController(g, permanent),
		Prompt:           "Choose a chapter for read ahead.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{1},
	}, log)[0]
	permanent.SagaEntryChapter = chosen
	permanent.Counters.Remove(counter.Lore, permanent.Counters.Get(counter.Lore))
	permanent.Counters.Add(counter.Lore, chosen)
}

func addCountersToPermanent(g *game.Game, permanent *game.Permanent, kind counter.Kind, amount int) bool {
	placementController := game.PlayerID(-1)
	if permanent != nil {
		placementController = effectiveController(g, permanent)
	}
	return addCountersToPermanentControlledBy(g, placementController, permanent, kind, amount)
}

func addCountersToPermanentControlledBy(g *game.Game, placementController game.PlayerID, permanent *game.Permanent, kind counter.Kind, amount int) bool {
	if permanent == nil || amount <= 0 {
		return false
	}
	amount = replacementPermanentCounterPlacementAmount(g, placementController, permanent, kind, amount)
	previous := permanent.Counters.Get(kind)
	permanent.Counters.Add(kind, amount)
	emitCounterAddedEvent(g, permanent, placementController, kind, previous, amount)
	return true
}

func emitCounterAddedEvent(g *game.Game, permanent *game.Permanent, controller game.PlayerID, kind counter.Kind, previous, amount int) {
	emitEvent(g, game.Event{
		Kind:                  game.EventCountersAdded,
		SourceID:              permanent.CardInstanceID,
		SourceObjectID:        permanent.ObjectID,
		Controller:            controller,
		CardID:                permanent.CardInstanceID,
		PermanentID:           permanent.ObjectID,
		CounterKind:           kind,
		PreviousCounterAmount: previous,
		Amount:                amount,
	})
}

func addCountersToPlayer(g *game.Game, player *game.Player, kind counter.Kind, amount int) bool {
	placementController := game.PlayerID(-1)
	if player != nil {
		placementController = player.ID
	}
	return addCountersToPlayerControlledBy(g, placementController, player, kind, amount)
}

func addCountersToPlayerControlledBy(g *game.Game, placementController game.PlayerID, player *game.Player, kind counter.Kind, amount int) bool {
	if player == nil || amount <= 0 {
		return false
	}
	amount = replacementPlayerCounterPlacementAmount(g, placementController, player.ID, kind, amount)
	previous := 0
	switch kind {
	case counter.Poison:
		previous = player.PoisonCounters
		player.PoisonCounters += amount
	case counter.Energy:
		previous = player.EnergyCounters
		player.EnergyCounters += amount
	case counter.Experience:
		previous = player.ExperienceCounters
		player.ExperienceCounters += amount
	default:
		return false
	}
	emitEvent(g, game.Event{
		Kind:                  game.EventCountersAdded,
		Controller:            placementController,
		Player:                player.ID,
		CounterKind:           kind,
		PreviousCounterAmount: previous,
		Amount:                amount,
	})
	return true
}

func advanceSagas(g *game.Game, controller game.PlayerID) {
	for _, permanent := range slices.Clone(g.Battlefield) {
		if effectiveController(g, permanent) != controller {
			continue
		}
		if !permanentHasSubtype(g, permanent, types.Saga) {
			continue
		}
		addCountersToPermanent(g, permanent, counter.Lore, 1)
	}
}

func finalSagaChapter(g *game.Game, permanent *game.Permanent) int {
	final := 0
	for _, body := range permanentEffectiveAbilities(g, permanent) {
		chapter, ok := body.(game.ChapterAbility)
		if !ok {
			continue
		}
		for _, number := range chapter.Chapters {
			if number > final {
				final = number
			}
		}
	}
	return final
}

func sagaAwaitingChapterAbility(g *game.Game, permanent *game.Permanent, final int) bool {
	for _, object := range g.Stack.Objects() {
		if object.Kind == game.StackTriggeredAbility &&
			object.SagaChapter &&
			object.SourceID == permanent.ObjectID {
			return true
		}
	}
	start := min(max(g.TriggerEventCursor, 0), len(g.Events))
	for _, event := range g.Events[start:] {
		if sagaChapterTriggeredByEvent(permanent, event, final) {
			return true
		}
	}
	return false
}

func sagaChapterTriggeredByEvent(permanent *game.Permanent, event game.Event, chapter int) bool {
	return event.Kind == game.EventCountersAdded &&
		event.PermanentID == permanent.ObjectID &&
		event.CounterKind == counter.Lore &&
		event.PreviousCounterAmount < chapter &&
		event.PreviousCounterAmount+event.Amount >= chapter &&
		(permanent.SagaEntryChapter == 0 || chapter >= permanent.SagaEntryChapter)
}
