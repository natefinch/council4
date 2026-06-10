package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type conditionContext struct {
	controller             game.PlayerID
	source                 *game.Permanent
	event                  *game.Event
	obj                    *game.StackObject
	useBaseCharacteristics bool
}

func conditionSatisfied(g *game.Game, ctx conditionContext, condition opt.V[game.Condition]) bool {
	if !condition.Exists || condition.Val.Empty() {
		return true
	}
	cond := condition.Val
	if cond.ControllerLifeAtLeast < 0 ||
		cond.AnyPlayerLifeAtMost < 0 ||
		cond.OpponentCountAtLeast < 0 {
		return false
	}
	matches := true
	if cond.ControlsMatching.Exists {
		matches = matches && controllerControlsMatchingSelection(g, ctx, cond.ControlsMatching.Val)
	} else if !cond.ControllerControls.Empty() {
		matches = matches && controllerControlsMatchingSelection(g, ctx, controlSelectionFromFilter(cond.ControllerControls))
	}
	if cond.ControllerLifeAtLeast > 0 {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.Life >= cond.ControllerLifeAtLeast
	}
	if cond.AnyPlayerLifeAtMost > 0 {
		matches = matches && anyPlayerLifeAtMost(g, cond.AnyPlayerLifeAtMost)
	}
	if cond.OpponentCountAtLeast > 0 {
		matches = matches && len(aliveOpponents(g, ctx.controller)) >= cond.OpponentCountAtLeast
	}
	if cond.AnyOpponentControls.Exists {
		matches = matches && anyOpponentControlsMatchingSelection(g, ctx, cond.AnyOpponentControls.Val)
	}
	if cond.OpponentsControl.Exists {
		matches = matches && playersControlMatchingSelection(g, ctx, aliveOpponents(g, ctx.controller), cond.OpponentsControl.Val)
	}
	if cond.Object.Exists || len(cond.Types) > 0 {
		matches = matches && conditionObjectMatches(g, ctx, &cond)
	}
	if cond.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures {
		matches = matches && eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g, ctx)
	}
	if cond.SourceClassLevelAtLeast > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel >= cond.SourceClassLevelAtLeast
	}
	if cond.SourceClassLevelLessThan > 0 {
		matches = matches && ctx.source != nil && ctx.source.ClassLevel < cond.SourceClassLevelLessThan
	}
	if cond.SourceNotMonstrous {
		matches = matches && ctx.source != nil && !ctx.source.Monstrous
	}
	if cond.ControllerHasMaxSpeed {
		player, ok := playerByID(g, ctx.controller)
		matches = matches && ok && player.Speed >= 4
	}
	if cond.TargetEnteredThisTurn.Exists {
		matches = matches && conditionTargetEnteredThisTurn(g, ctx, cond.TargetEnteredThisTurn.Val)
	}
	if cond.CastFromZone.Exists {
		matches = matches && ctx.obj != nil && !ctx.obj.Copy && ctx.obj.SourceZone == cond.CastFromZone.Val
	}
	if cond.Negate {
		return !matches
	}
	return matches
}

func conditionObjectMatches(g *game.Game, ctx conditionContext, cond *game.Condition) bool {
	obj := ctx.obj
	if obj == nil && ctx.event != nil {
		obj = &game.StackObject{HasTriggerEvent: true, TriggerEvent: *ctx.event, Controller: ctx.controller}
	}
	ref := game.EventPermanentReference()
	if cond.Object.Exists {
		ref = cond.Object.Val
	}
	resolved, ok := resolveObjectReference(g, obj, ref)
	if !ok {
		return false
	}
	for _, cardType := range cond.Types {
		if !resolvedObjectHasType(g, &resolved, cardType) {
			return false
		}
	}
	return true
}

func resolvedObjectHasType(g *game.Game, resolved *resolvedObjectReference, cardType types.Card) bool {
	if resolved.permanent != nil {
		return permanentHasType(g, resolved.permanent, cardType)
	}
	return slices.Contains(resolved.snapshot.Types, cardType)
}

func controlSelectionFromFilter(filter game.PermanentFilter) game.SelectionCount {
	return game.SelectionCount{
		Selection:  filter.Selection(),
		MinCount:   filter.MinCount,
		TotalPower: filter.TotalPower,
	}
}

func controllerControlsMatchingSelection(g *game.Game, ctx conditionContext, control game.SelectionCount) bool {
	return playersControlMatchingSelection(g, ctx, []game.PlayerID{ctx.controller}, control)
}

func anyOpponentControlsMatchingSelection(g *game.Game, ctx conditionContext, control game.SelectionCount) bool {
	for _, opponent := range aliveOpponents(g, ctx.controller) {
		if playersControlMatchingSelection(g, ctx, []game.PlayerID{opponent}, control) {
			return true
		}
	}
	return false
}

func playersControlMatchingSelection(g *game.Game, ctx conditionContext, controllers []game.PlayerID, control game.SelectionCount) bool {
	want := control.MinCount
	if want <= 0 {
		want = 1
	}
	allowed := make(map[game.PlayerID]bool, len(controllers))
	for _, controller := range controllers {
		allowed[controller] = true
	}
	count := 0
	totalPower := 0
	sel := control.Selection
	for _, permanent := range g.Battlefield {
		if ctx.useBaseCharacteristics {
			if !allowed[permanent.Controller] {
				continue
			}
		} else if !allowed[effectiveController(g, permanent)] {
			continue
		}
		var values permanentEffectiveValues
		if ctx.useBaseCharacteristics {
			values = basePermanentValues(g, permanent)
		} else {
			values = effectivePermanentValues(g, permanent)
		}
		subject := selectionSubject{
			kind:      subjectPermanent,
			g:         g,
			permanent: permanent,
			values:    &values,
			viewer:    ctx.controller,
			useBase:   ctx.useBaseCharacteristics,
		}
		if sel.Controller != game.ControllerAny {
			if ctx.useBaseCharacteristics {
				subject.controller = permanent.Controller
			} else {
				subject.controller = effectiveController(g, permanent)
			}
		}
		if ctx.source != nil {
			subject.sourceObjectID = ctx.source.ObjectID
		}
		if !matchSelection(&subject, &sel) {
			continue
		}
		count++
		if control.TotalPower.Exists {
			powerValues := &values
			if ctx.useBaseCharacteristics {
				effective := effectivePermanentValues(g, permanent)
				powerValues = &effective
			}
			if powerValues.powerOK {
				totalPower += powerValues.power
			}
		}
		if count >= want {
			if !control.TotalPower.Exists || control.TotalPower.Val.Matches(totalPower) {
				return true
			}
		}
	}
	if control.TotalPower.Exists {
		return count >= want && control.TotalPower.Val.Matches(totalPower)
	}
	return false
}

func anyPlayerLifeAtMost(g *game.Game, maximum int) bool {
	for playerID := range game.PlayerID(game.NumPlayers) {
		player, ok := playerByID(g, playerID)
		if ok && !player.Eliminated && player.Life <= maximum {
			return true
		}
	}
	return false
}

func conditionTargetEnteredThisTurn(g *game.Game, ctx conditionContext, targetIndex int) bool {
	if ctx.obj == nil {
		return false
	}
	permanent, ok := effectPermanentAt(g, ctx.obj, targetIndex)
	if !ok {
		return false
	}
	for _, event := range g.EventsThisTurn() {
		if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanent.ObjectID {
			return true
		}
	}
	return false
}

func eventPermanentNameUniqueAmongControlledAndGraveyardCreatures(g *game.Game, ctx conditionContext) bool {
	if ctx.event == nil || ctx.event.PermanentID == 0 {
		return false
	}
	resolved, ok := resolvePermanentOrLastKnown(g, ctx.event.PermanentID)
	if !ok {
		return false
	}
	name := resolvedObjectName(g, &resolved)
	if name == "" {
		return false
	}
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == ctx.event.PermanentID || effectiveController(g, permanent) != ctx.controller || !permanentHasType(g, permanent, types.Creature) {
			continue
		}
		if def, ok := permanentCardDef(g, permanent); ok && def.Name == name {
			return false
		}
	}
	player, ok := playerByID(g, ctx.controller)
	if !ok {
		return false
	}
	for _, cardID := range player.Graveyard.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if def.Name == name && def.HasType(types.Creature) {
			return false
		}
	}
	return true
}

func resolvedObjectName(g *game.Game, resolved *resolvedObjectReference) string {
	if resolved.permanent != nil {
		if resolved.permanent.Token {
			return permanentTokenName(resolved.permanent)
		}
		return permanentEffectiveName(g, resolved.permanent)
	}
	return resolved.snapshot.Name
}

func activationConditionSatisfied(g *game.Game, playerID game.PlayerID, permanent *game.Permanent, condition opt.V[game.Condition]) bool {
	return conditionSatisfied(g, conditionContext{
		controller: playerID,
		source:     permanent,
	}, condition)
}
