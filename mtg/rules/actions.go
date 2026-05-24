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
	actions = append(actions, e.legalCastActions(g, playerID)...)
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

func (e *Engine) legalCastActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}

	player := g.Players[playerID]
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card := g.GetCardInstance(cardID)
		if card == nil || card.Def == nil {
			continue
		}
		for _, targets := range targetChoicesForSpell(g, card.Def) {
			if e.canCastSpell(g, playerID, cardID, targets, 0, nil) {
				actions = append(actions, action.CastSpell(cardID, append([]game.Target(nil), targets...), 0, nil))
			}
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
	case action.ActionCastSpell:
		return e.applyCastSpell(g, playerID, act.CastSpell)
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

func (e *Engine) applyCastSpell(g *game.Game, playerID game.PlayerID, cast action.CastSpellAction) bool {
	if !e.canCastSpell(g, playerID, cast.CardID, cast.Targets, cast.XValue, cast.ChosenModes) {
		return false
	}
	player := g.Players[playerID]
	if !payCost(g, playerID, g.GetCardInstance(cast.CardID).Def.ManaCost) {
		return false
	}
	if !player.Hand.Remove(cast.CardID) {
		panic("cast spell disappeared from hand after validation")
	}
	g.Stack.Push(&game.StackObject{
		ID:          g.IDGen.Next(),
		Kind:        game.StackSpell,
		SourceID:    cast.CardID,
		Controller:  playerID,
		Targets:     append([]game.Target(nil), cast.Targets...),
		ChosenModes: append([]int(nil), cast.ChosenModes...),
		XValue:      cast.XValue,
	})
	return true
}

func (e *Engine) canCastSpell(g *game.Game, playerID game.PlayerID, cardID id.ID, targets []game.Target, xValue int, chosenModes []int) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	if len(chosenModes) != 0 || xValue != 0 {
		return false
	}
	player := g.Players[playerID]
	card := g.GetCardInstance(cardID)
	if card == nil || card.Def == nil || !player.Hand.Contains(cardID) {
		return false
	}
	if !isSupportedSpell(card.Def) || !targetsValidForSpell(g, card.Def, targets) {
		return false
	}
	if !canCastAtCurrentTiming(g, playerID, card.Def) {
		return false
	}
	return canPayCost(g, playerID, card.Def.ManaCost)
}

func canAct(g *game.Game, playerID game.PlayerID) bool {
	return isPlayerAlive(g, playerID)
}

func canPlayAnyLand(g *game.Game, playerID game.PlayerID) bool {
	return canAct(g, playerID) &&
		playerID == g.Turn.ActivePlayer &&
		playerID == g.Turn.PriorityPlayer &&
		isSorcerySpeed(g, playerID) &&
		g.Turn.CanPlayLand()
}

func canCastAtCurrentTiming(g *game.Game, playerID game.PlayerID, card *game.CardDef) bool {
	if card.HasType(game.TypeInstant) {
		return true
	}
	return isSorcerySpeed(g, playerID)
}

func isSorcerySpeed(g *game.Game, playerID game.PlayerID) bool {
	return playerID == g.Turn.ActivePlayer &&
		g.Turn.IsMainPhase() &&
		g.Turn.Step == game.StepNone &&
		g.Stack.IsEmpty()
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

func isSupportedSpell(card *game.CardDef) bool {
	return !card.HasType(game.TypeLand) &&
		(card.HasType(game.TypeCreature) ||
			card.HasType(game.TypeInstant) ||
			card.HasType(game.TypeSorcery))
}
