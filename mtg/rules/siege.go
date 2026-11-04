package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// siegeDefeatAbility is the intrinsic triggered ability every Siege has
// (CR 310.12b). It is shared immutable data; its behavior is
// resolveSiegeDefeat, selected by StackObject.SiegeDefeat.
var siegeDefeatAbility = &game.TriggeredAbility{
	Text: "When the last defense counter is removed from this permanent, exile it, then you may cast it transformed without paying its mana cost.",
}

func counterRemovedEvent(permanent *game.Permanent, controller game.PlayerID, kind counter.Kind, previous, amount int, simultaneousID id.ID) game.Event {
	return game.Event{
		Kind:                  game.EventCountersRemoved,
		SourceID:              permanent.CardInstanceID,
		SourceObjectID:        permanent.ObjectID,
		Controller:            controller,
		CardID:                permanent.CardInstanceID,
		PermanentID:           permanent.ObjectID,
		CounterKind:           kind,
		PreviousCounterAmount: previous,
		Amount:                amount,
		SimultaneousID:        simultaneousID,
	}
}

// siegeDefeatTriggeredByEvent reports whether event removed the last defense
// counter from permanent. A permanent that never had a defense counter can't
// have its last one removed, and putting a 0-defense battle into a graveyard
// removes no counter (Siege rulings 2023-04-14).
func siegeDefeatTriggeredByEvent(permanent *game.Permanent, event game.Event) bool {
	return event.Kind == game.EventCountersRemoved &&
		event.PermanentID == permanent.ObjectID &&
		event.CounterKind == counter.Defense &&
		event.PreviousCounterAmount > 0 &&
		event.Amount >= event.PreviousCounterAmount
}

// resolveSiegeDefeat resolves a Siege's intrinsic ability: exile the source
// permanent, then its controller may cast the exiled card transformed without
// paying its mana cost (CR 310.12b, 712.11a). Only the card the source became
// when exiled is offered; a token ceases to exist, and a card without a
// transforming back face can't be cast transformed and stays in exile. A
// source that already left the battlefield, or that a replacement put
// elsewhere, is not followed.
func (e *Engine) resolveSiegeDefeat(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	source, ok := permanentByObjectID(g, obj.SourceID)
	if !ok || source.PhasedOut {
		return "resolved"
	}
	for _, result := range movePermanentsToZoneSimultaneouslyWithResults(g, []*game.Permanent{source}, zone.Exile) {
		if !result.moved || result.destination != zone.Exile {
			continue
		}
		for _, arrival := range result.arrivals {
			card, ok := g.GetCardInstance(arrival.CardID)
			if !ok || card.ZoneVersion != arrival.ZoneVersion {
				continue
			}
			e.chooseAndCastFreeSpell(g, g.Players[card.Owner], obj.Controller, card.ID, zone.Exile,
				game.Selection{}, opt.V[int]{}, freeCastSpell{transformed: true}, true, agents, log)
		}
	}
	return "resolved"
}
