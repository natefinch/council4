package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// foretellExileCost is the fixed generic cost paid to foretell a card
// (CR 702.143a): {2}, regardless of the card's foretell cost.
func foretellExileCost() cost.Mana {
	return cost.Mana{cost.O(2)}
}

// foretellCostForCard returns the Foretell cast cost printed on the card's front
// face (CR 702.143), or (nil, false) when the card has no Foretell keyword. This
// is the cost paid to cast the card from exile after foretelling it, not the
// fixed {2} paid to foretell it.
func foretellCostForCard(card *game.CardDef) (cost.Mana, bool) {
	return card.ForetellCost()
}

// canForetellCard reports whether the player may foretell the hand card cardID:
// foretelling is a special action usable any time the player has priority during
// their own turn (CR 702.143a, 116.2h), so it requires the player's turn,
// priority, the card in the player's hand, a Foretell keyword, and the fixed {2}
// foretell cost payable now or with payment-time mana abilities (CR 605.3a).
// Unlike Plot, foretelling is not restricted to sorcery speed.
func (e *Engine) canForetellCard(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	return e.activationFeasible(g, playerID, foretellCost(g, playerID, cardID))
}

// foretellCost is the cost of foretelling the hand card cardID, or the zero
// activationCost when that isn't legal.
func foretellCost(g *game.Game, playerID game.PlayerID, cardID id.ID) activationCost {
	if !canAct(g, playerID) || playerID != g.Turn.ActivePlayer || playerID != g.Turn.PriorityPlayer {
		return activationCost{}
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(cardID) {
		return activationCost{}
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return activationCost{}
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	if _, ok := foretellCostForCard(spellDef); !ok {
		return activationCost{}
	}
	exileCost := foretellExileCost()
	return specialActionCost(playerID, &exileCost, handCardPresent(playerID, cardID))
}

// applyForetellCard resolves a foretell special action: the player pays the fixed
// {2}, activating mana abilities while paying if they choose (CR 605.3a), and
// exiles the card from hand face down, recording the foretell turn so the card
// may be cast from exile for its foretell cost on a later turn (CR 702.143a).
func (e *Engine) applyForetellCard(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) actionOutcome {
	defer g.OpenLibraryTopWindow()()
	c := foretellCost(g, playerID, cardID)
	if !e.activationFeasible(g, playerID, c) {
		return actionFailed
	}
	card, _ := g.GetCardInstance(cardID)
	exileCost := foretellExileCost()
	if outcome := e.paySpecialActionCost(g, playerID, c, &exileCost, agents, log); outcome != actionApplied {
		return outcome
	}
	destination, moved := moveCardBetweenZonesWithEvent(g, game.Event{
		Controller: playerID, Player: card.Owner, CardID: cardID,
		FromZone: zone.Hand, ToZone: zone.Exile,
	}, zoneMoveOptions{faceDownExile: true})
	if !moved {
		return actionFailed
	}
	if destination == zone.Exile {
		if g.ForetoldCards == nil {
			g.ForetoldCards = make(map[id.ID]int)
		}
		g.ForetoldCards[cardID] = g.Turn.TurnNumber
	}
	return actionApplied
}

// legalForetellActions returns the foretell special actions available to the
// player: one for each hand card whose Foretell keyword may currently be
// activated.
func (e *Engine) legalForetellActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.ActivePlayer || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		if e.canForetellCard(g, playerID, cardID) {
			actions = append(actions, actionBuild.foretellCard(cardID))
		}
	}
	return actions
}

// cardIsForetoldInExile reports whether playerID may now cast cardID from exile
// for its foretell cost: it must be recorded as foretold, the current turn must
// be later than the turn it was foretold, and playerID must be the card's owner.
// Foretelling exiles a card from its owner's hand and only "that player" may
// cast it (CR 702.143a); a card that becomes foretold by an effect may be cast by
// its owner (CR 702.143d). Any other player's permission to cast the card is an
// independent route with its own cost.
func cardIsForetoldInExile(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	foretoldTurn, ok := g.ForetoldCards[cardID]
	return ok && g.Turn.TurnNumber > foretoldTurn && cardOwnedBy(g, playerID, cardID)
}
