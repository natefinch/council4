package rules

import "github.com/natefinch/council4/mtg/game"

func resolveFight(g *game.Game, obj *game.StackObject, effect game.Effect) {
	if obj == nil || len(obj.Targets) < 2 {
		return
	}
	first := permanentByObjectID(g, obj.Targets[0].PermanentID)
	second := permanentByObjectID(g, obj.Targets[1].PermanentID)
	if first == nil || second == nil || first.ObjectID == second.ObjectID || !permanentHasType(g, first, game.TypeCreature) || !permanentHasType(g, second, game.TypeCreature) {
		return
	}
	dealPermanentDamage(g, first.CardInstanceID, first.ObjectID, effectiveController(g, first), second, effectivePower(g, first), false)
	dealPermanentDamage(g, second.CardInstanceID, second.ObjectID, effectiveController(g, second), first, effectivePower(g, second), false)
}
