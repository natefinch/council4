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
		if !spellHasAnyLegalTargets(g, card.Def, obj.Controller, obj.Targets) {
			owner := playerByID(g, card.Owner)
			if owner == nil {
				return "invalid owner"
			}
			owner.Graveyard.Add(card.ID)
			return "countered by rules"
		}
		permanent := createCardPermanent(g, card, obj.Controller)
		if permanent != nil && isAttachmentPermanent(g, permanent) && len(obj.Targets) > 0 {
			target := effectPermanent(g, obj, game.Effect{TargetIndex: 0})
			if !attachPermanent(g, permanent, target) {
				movePermanentToZone(g, permanent, game.ZoneGraveyard)
				return "graveyard"
			}
		}
		return "battlefield"
	}
	if card.Def.HasType(game.TypeInstant) || card.Def.HasType(game.TypeSorcery) {
		if !spellHasAnyLegalTargets(g, card.Def, obj.Controller, obj.Targets) {
			owner := playerByID(g, card.Owner)
			if owner == nil {
				return "invalid owner"
			}
			owner.Graveyard.Add(card.ID)
			return "countered by rules"
		}
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
