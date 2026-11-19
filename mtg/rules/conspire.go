package rules

import "github.com/natefinch/council4/mtg/game"

func conspireCastTriggers(g *game.Game, source *game.Permanent, event game.Event) []pendingTriggeredAbility {
	spell, ok := stackObjectByID(g, event.StackObjectID)
	if !ok || spell.Kind != game.StackSpell || spell.ConspirePaid == 0 {
		return nil
	}
	pending := make([]pendingTriggeredAbility, 0, spell.ConspirePaid)
	for range spell.ConspirePaid {
		pending = append(pending, pendingTriggeredAbility{
			controller: event.Controller, sourceID: source.ObjectID,
			sourceCardID: source.CardInstanceID, sourceToken: source.TokenDef, face: event.Face,
			inline: &game.TriggeredAbility{Text: "Conspire", KeywordAbilities: game.SimpleKeywords(game.Conspire)},
			event:  event, hasEvent: true, ordinaryTrigger: true,
		})
	}
	return pending
}

func (e *Engine) resolveConspireTriggeredAbility(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if !obj.HasTriggerEvent || obj.TriggerEvent.Kind != game.EventSpellCast {
		return "missing source"
	}
	spell, ok := triggeringSpellOrLKI(g, obj)
	if !ok || spell.ConspirePaid == 0 {
		return "missing source"
	}
	r := newEffectResolver(e, g, obj, agents, log)
	copyStackObjectForPlayer(r, game.CopyStackObject{MayChooseNewTargets: true}, spell, obj.Controller)
	return "resolved"
}
