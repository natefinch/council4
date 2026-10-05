package rules

import "github.com/natefinch/council4/mtg/game"

func conditionTargetCardAvailable(g *game.Game, obj *game.StackObject, cond *game.Condition) bool {
	if !cond.Object.Exists || cond.Object.Val.Kind() != game.ObjectReferenceTargetCard {
		return false
	}
	slot := remapTargetSlot(g, obj, cond.Object.Val.TargetIndex())
	if slot < 0 || slot >= len(obj.Targets) {
		return false
	}
	target := obj.Targets[slot]
	if target.Kind != game.TargetCard || !target.CardZoneVersionSet {
		return false
	}
	card, ok := g.GetCardInstance(target.CardID)
	if !ok || card.Def == nil {
		return false
	}
	if cond.TargetCardResultKey == "" {
		return card.ZoneVersion == target.CardZoneVersion
	}
	key := string(cond.TargetCardResultKey)
	result, published := obj.ResolutionResults[key]
	snapshots := obj.ResolutionResultObjects[key]
	if !published || !result.Succeeded || len(snapshots) != 1 {
		return false
	}
	snapshot := snapshots[0]
	if snapshot.ObjectID != 0 || snapshot.CardID != target.CardID ||
		!snapshot.TargetCardZoneVersion.Exists ||
		snapshot.TargetCardZoneVersion.Val != target.CardZoneVersion || len(snapshot.ZoneCards) != 1 {
		return false
	}
	reached := snapshot.ZoneCards[0]
	return reached.CardID == card.ID && reached.ZoneVersion == card.ZoneVersion
}
