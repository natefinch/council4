package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func addCountersToPermanent(g *game.Game, permanent *game.Permanent, kind counter.Kind, amount int) bool {
	if permanent == nil || amount <= 0 {
		return false
	}
	previous := permanent.Counters.Get(kind)
	permanent.Counters.Add(kind, amount)
	emitCounterAddedEvent(g, permanent, kind, previous, amount)
	return true
}

func emitCounterAddedEvent(g *game.Game, permanent *game.Permanent, kind counter.Kind, previous, amount int) {
	emitEvent(g, game.Event{
		Kind:                  game.EventCountersAdded,
		SourceID:              permanent.CardInstanceID,
		SourceObjectID:        permanent.ObjectID,
		Controller:            effectiveController(g, permanent),
		CardID:                permanent.CardInstanceID,
		PermanentID:           permanent.ObjectID,
		CounterKind:           kind,
		PreviousCounterAmount: previous,
		Amount:                amount,
	})
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
		if event.Kind == game.EventCountersAdded &&
			event.PermanentID == permanent.ObjectID &&
			event.CounterKind == counter.Lore &&
			event.PreviousCounterAmount < final &&
			event.PreviousCounterAmount+event.Amount >= final {
			return true
		}
	}
	return false
}
