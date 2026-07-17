package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// handleConditionalDestinationPlace resolves a ConditionalDestinationPlace: it
// reads the referenced card from its source zone, evaluates the combined gate,
// and atomically routes the card to the battlefield or to its fallback zone. The
// battlefield put is offered only when the gate holds and the controller accepts;
// any other outcome (gate fails, controller declines, or the put cannot complete)
// falls through to the else placement, which the controller may itself decline
// when the move is optional, leaving the card where it was.
func handleConditionalDestinationPlace(r *effectResolver, prim game.ConditionalDestinationPlace) effectResolved {
	res := effectResolved{accepted: true}
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
	if !ok || fromZone != prim.FromZone {
		return res
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return res
	}
	gateHolds := cardConditionPredicateSatisfied(r.game, r.obj, card, prim.CardCondition) &&
		effectConditionSatisfied(r.game, r.obj, prim.Condition)
	if gateHolds && (prim.ThenMandatory ||
		r.engine.chooseMay(r.game, r.agents, r.obj.Controller, conditionalDestinationThenPrompt(prim), r.log)) {
		if prim.Then == zone.None {
			options := permanentCreationOptions{ForceTapped: prim.EntryTapped}
			if _, placed := r.putReferencedCardOnBattlefieldValue(
				prim.Card,
				game.PlayerReference{},
				nil,
				options,
				prim.Card.Kind == game.CardReferenceEvent,
			); placed {
				res.succeeded = true
				return res
			}
		} else {
			if prim.ThenReveal {
				emitCardRevealEvent(r.game, r.obj, card.Owner, cardID, fromZone)
			}
			if moveCardBetweenZonesWithPlacement(r.game, card.Owner, cardID, fromZone, prim.Then, false) {
				res.succeeded = true
				return res
			}
		}
	}
	if prim.Else == zone.None {
		return res
	}
	if prim.ElseOptional && !r.engine.chooseMay(r.game, r.agents, r.obj.Controller, conditionalDestinationElsePrompt(prim), r.log) {
		return res
	}
	res.succeeded = moveCardBetweenZonesWithPlacement(r.game, card.Owner, cardID, fromZone, prim.Else, prim.ElseBottom)
	return res
}

func conditionalDestinationThenPrompt(prim game.ConditionalDestinationPlace) string {
	switch prim.Then {
	case zone.None:
		return "Put the card onto the battlefield?"
	case zone.Hand:
		if prim.ThenReveal {
			return "Reveal the card and put it into your hand?"
		}
		return "Put the card into your hand?"
	default:
		return "Put the card into the chosen zone?"
	}
}

func conditionalDestinationElsePrompt(prim game.ConditionalDestinationPlace) string {
	if prim.ElseBottom {
		return "Put the card on the bottom of your library?"
	}
	switch prim.Else {
	case zone.Graveyard:
		return "Put the card into your graveyard?"
	default:
		return "Put the card into your hand?"
	}
}
