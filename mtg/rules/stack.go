package rules

import "github.com/natefinch/council4/mtg/game"

func (e *Engine) resolveTopOfStack(g *game.Game, log *TurnLog) {
	if g == nil {
		return
	}
	obj := g.Stack.Pop()
	if obj == nil {
		return
	}
	result := e.resolveStackObject(g, obj, log)
	if log == nil {
		return
	}
	log.Resolves = append(log.Resolves, ResolveLog{
		StackObjectID: obj.ID,
		SourceID:      obj.SourceID,
		Controller:    obj.Controller,
		Kind:          obj.Kind,
		Result:        result,
	})
}

func (e *Engine) resolveStackObject(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	switch obj.Kind {
	case game.StackSpell:
		return e.resolveSpell(g, obj, log)
	default:
		return "resolved"
	}
}

func (e *Engine) resolveSpell(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	card := g.GetCardInstance(obj.SourceID)
	if card == nil || card.Def == nil {
		return "missing source"
	}
	if card.Def.IsPermanent() {
		objectID := g.IDGen.Next()
		g.Battlefield = append(g.Battlefield, &game.Permanent{
			ObjectID:       objectID,
			CardInstanceID: obj.SourceID,
			Owner:          card.Owner,
			Controller:     obj.Controller,
			SummoningSick:  entersSummoningSick(card.Def),
			Timestamp:      int64(objectID),
		})
		return "battlefield"
	}
	if card.Def.HasType(game.TypeInstant) || card.Def.HasType(game.TypeSorcery) {
		e.resolveSpellEffects(g, obj, card, log)
		owner := playerByID(g, card.Owner)
		if owner == nil {
			return "invalid owner"
		}
		owner.Graveyard.Add(card.ID)
		return "graveyard"
	}
	return "resolved"
}

func playerByID(g *game.Game, playerID game.PlayerID) *game.Player {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return nil
	}
	return g.Players[playerID]
}
