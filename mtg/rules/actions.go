package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
)

func (e *Engine) legalActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) {
		return []action.Action{action.Pass()}
	}

	actions := e.legalLandActions(g, playerID)
	actions = append(actions, action.Pass())
	return actions
}

func (e *Engine) legalLandActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canPlayAnyLand(g, playerID) {
		return nil
	}

	player := g.Players[playerID]
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		if _, ok := landCardInstance(g, player, cardID); ok {
			actions = append(actions, action.PlayLand(cardID))
		}
	}
	return actions
}

func (e *Engine) applyAction(g *game.Game, playerID game.PlayerID, act action.Action) bool {
	switch act.Kind {
	case action.ActionPass:
		return true
	case action.ActionPlayLand:
		return e.applyPlayLand(g, playerID, act.PlayLand.CardID)
	default:
		return false
	}
}

func (e *Engine) applyPlayLand(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	if !canPlayAnyLand(g, playerID) {
		return false
	}

	player := g.Players[playerID]
	card, ok := landCardInstance(g, player, cardID)
	if !ok {
		return false
	}
	if !player.Hand.Remove(cardID) {
		return false
	}

	objectID := g.IDGen.Next()
	g.Battlefield = append(g.Battlefield, &game.Permanent{
		ObjectID:       objectID,
		CardInstanceID: cardID,
		Owner:          card.Owner,
		Controller:     playerID,
		SummoningSick:  entersSummoningSick(card.Def),
		Timestamp:      int64(objectID), // Placeholder until timestamps need their own counter.
	})
	g.Turn.LandsPlayedThisTurn++
	return true
}

func canAct(g *game.Game, playerID game.PlayerID) bool {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return false
	}
	player := g.Players[playerID]
	return player != nil && !player.Eliminated && !g.TurnOrder.IsEliminated(playerID)
}

func canPlayAnyLand(g *game.Game, playerID game.PlayerID) bool {
	return canAct(g, playerID) &&
		playerID == g.Turn.ActivePlayer &&
		playerID == g.Turn.PriorityPlayer &&
		g.Turn.IsMainPhase() &&
		g.Turn.Step == game.StepNone &&
		g.Stack.IsEmpty() &&
		g.Turn.CanPlayLand()
}

func landCardInstance(g *game.Game, player *game.Player, cardID id.ID) (*game.CardInstance, bool) {
	if player == nil || !player.Hand.Contains(cardID) {
		return nil, false
	}
	card := g.GetCardInstance(cardID)
	if card == nil || card.Def == nil || !card.Def.HasType(game.TypeLand) {
		return nil, false
	}
	return card, true
}

func entersSummoningSick(card *game.CardDef) bool {
	return card != nil && !card.HasKeyword(game.Haste)
}
