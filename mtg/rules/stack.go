package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	payment "github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) resolveTopOfStack(g *game.Game, log *TurnLog) {
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveTopOfStackWithChoices(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	obj, ok := g.Stack.Pop()
	if !ok {
		return
	}
	result := e.resolveStackObjectWithChoices(g, obj, agents, log)
	if obj.Kind == game.StackSpell && spellResolved(result) {
		emitEvent(g, game.GameEvent{
			Kind:          game.EventSpellResolved,
			SourceID:      obj.SourceID,
			StackObjectID: obj.ID,
			Controller:    obj.Controller,
			CardID:        obj.SourceID,
		})
	}
	log.addResolve(ResolveLog{
		StackObjectID: obj.ID,
		SourceID:      obj.SourceID,
		Controller:    obj.Controller,
		Kind:          obj.Kind,
		Result:        result,
	})
}

func (e *Engine) resolveStackObject(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	return e.resolveStackObjectWithChoices(g, obj, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveStackObjectWithChoices(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	switch obj.Kind {
	case game.StackSpell:
		return e.resolveSpellWithChoices(g, obj, agents, log)
	case game.StackActivatedAbility:
		return e.resolveActivatedAbilityWithChoices(g, obj, agents, log)
	case game.StackTriggeredAbility:
		return e.resolveTriggeredAbilityWithChoices(g, obj, agents, log)
	default:
		return "resolved"
	}
}

func spellResolved(result string) bool {
	switch result {
	case "resolved", "battlefield", "graveyard":
		return true
	default:
		return false
	}
}

func (e *Engine) resolveActivatedAbility(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	return e.resolveActivatedAbilityWithChoices(g, obj, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveActivatedAbilityWithChoices(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	permanent, permanentOK := permanentByObjectID(g, obj.SourceID)
	def, defOK := stackObjectSourceDef(g, obj)
	if !defOK && permanentOK {
		if physicalDef, ok := physicalPermanentDef(g, permanent); ok {
			def, defOK = physicalDef.FaceDef(obj.Face)
		}
	}
	if !defOK || obj.AbilityIndex < 0 || obj.AbilityIndex >= len(def.Abilities) {
		return "missing source"
	}
	ability := &def.Abilities[obj.AbilityIndex]
	if permanentOK && isEquipmentPermanent(g, permanent) && abilityHasKeyword(ability, game.Equip) {
		sourceObjectID := obj.SourceID
		if !permanentOK {
			sourceObjectID = 0
		}
		if !abilityHasAnyLegalTargetsFromSourceObject(g, def, sourceObjectID, ability, obj.Controller, obj.Targets) {
			return "countered by rules"
		}
		if len(obj.Targets) != 1 || obj.Targets[0].Kind != game.TargetPermanent {
			return "countered by rules"
		}
		target, ok := permanentByObjectID(g, obj.Targets[0].PermanentID)
		if !ok || !attachPermanent(g, permanent, target) {
			return "countered by rules"
		}
		return "resolved"
	}
	if !abilityHasAnyLegalTargetsFromSourceObject(g, def, obj.SourceID, ability, obj.Controller, obj.Targets) {
		return "countered by rules"
	}
	for _, effect := range ability.Effects {
		e.resolveEffectWithChoices(g, obj, effect, agents, log)
	}
	return "resolved"
}

func (e *Engine) resolveTriggeredAbility(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	return e.resolveTriggeredAbilityWithChoices(g, obj, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveTriggeredAbilityWithChoices(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if obj.InlineAbility != nil {
		return e.resolveTriggeredAbilityDefWithChoices(g, obj, nil, obj.InlineAbility, agents, log)
	}
	def, ok := stackObjectSourceDef(g, obj)
	if !ok || obj.AbilityIndex < 0 || obj.AbilityIndex >= len(def.Abilities) {
		return "missing source"
	}
	ability := &def.Abilities[obj.AbilityIndex]
	return e.resolveTriggeredAbilityDefWithChoices(g, obj, def, ability, agents, log)
}

func (e *Engine) resolveTriggeredAbilityDefWithChoices(g *game.Game, obj *game.StackObject, source *game.CardDef, ability *game.AbilityDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if ability.Kind != game.TriggeredAbility {
		return "missing source"
	}
	if abilityHasKeyword(ability, game.Ward) && ability.WardCost.Exists {
		return e.resolveWardTriggeredAbilityWithChoices(g, obj, ability, agents, log)
	}
	if abilityHasKeyword(ability, game.Madness) && ability.MadnessCost.Exists {
		return e.resolveMadnessTriggeredAbilityWithChoices(g, obj, ability, agents, log)
	}
	var event *game.GameEvent
	if obj.HasTriggerEvent {
		event = &obj.TriggerEvent
	}
	if ability.Trigger.Exists && !triggerInterveningIf(g, obj.Controller, &ability.Trigger.Val, event) {
		return "intervening if false"
	}
	if !abilityHasAnyLegalTargetsFromSourceObject(g, source, obj.SourceID, ability, obj.Controller, obj.Targets) {
		return "countered by rules"
	}
	if ability.Optional && !e.chooseMay(g, agents, obj.Controller, "Apply optional triggered ability?", log) {
		return "declined"
	}
	for _, effect := range ability.Effects {
		e.resolveEffectWithChoices(g, obj, effect, agents, log)
	}
	return "resolved"
}

func (e *Engine) resolveWardTriggeredAbilityWithChoices(g *game.Game, obj *game.StackObject, ability *game.AbilityDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	targetObj, ok := stackObjectByID(g, obj.WardTargetStackObjectID)
	if !ok {
		return "resolved"
	}
	payer := targetObj.Controller
	cost := manaCostPtr(ability.WardCost)
	if paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: payer, Cost: cost}) && e.chooseMay(g, agents, payer, "Pay ward cost?", log) {
		prefs := e.paymentPreferencesForCost(g, payer, cost, nil, agents, log)
		if paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: payer, Cost: cost, Prefs: prefs}) {
			return "resolved"
		}
	}
	counterStackObject(g, obj.WardTargetStackObjectID)
	return "resolved"
}

func (e *Engine) chooseMay(g *game.Game, agents [game.NumPlayers]PlayerAgent, player game.PlayerID, prompt string, log *TurnLog) bool {
	selected := e.chooseChoice(g, agents, mayChoiceRequest(player, prompt), log)
	return len(selected) == 1 && selected[0] == 1
}

func stackObjectSourceDef(g *game.Game, obj *game.StackObject) (*game.CardDef, bool) {
	if obj.SourceCardID != 0 {
		card, ok := g.GetCardInstance(obj.SourceCardID)
		if !ok {
			return nil, false
		}
		return card.Def.FaceDef(obj.Face)
	}
	if obj.SourceTokenDef == nil {
		return nil, false
	}
	return obj.SourceTokenDef.FaceDef(obj.Face)
}

func stackObjectByID(g *game.Game, objectID id.ID) (*game.StackObject, bool) {
	for _, obj := range g.Stack.Objects() {
		if obj.ID == objectID {
			return obj, true
		}
	}
	return nil, false
}

func counterStackObject(g *game.Game, objectID id.ID) bool {
	if obj, ok := stackObjectByID(g, objectID); ok && obj.Kind == game.StackSpell && !stackSpellCanBeCountered(g, obj) {
		return false
	}
	obj, ok := g.Stack.RemoveByID(objectID)
	if !ok {
		return false
	}
	if obj.Kind != game.StackSpell {
		return true
	}
	if obj.Copy {
		return true
	}
	card, ok := g.GetCardInstance(obj.SourceID)
	if !ok {
		return false
	}
	return moveStackCardToGraveyard(g, obj, card)
}

func stackSpellCanBeCountered(g *game.Game, obj *game.StackObject) bool {
	_, spellDef, ok := cardInstanceFaceDef(g, obj.SourceID, obj.Face)
	if !ok {
		return true
	}
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectCantBeCountered {
			continue
		}
		if !controllerRelationMatches(effect.Controller, obj.Controller, effect.AffectedController) {
			continue
		}
		if !spellTypesMatch(spellDef, effect.SpellTypes) {
			continue
		}
		return false
	}
	return true
}

func spellTypesMatch(card *game.CardDef, types []game.CardType) bool {
	for _, cardType := range types {
		if !card.HasType(cardType) {
			return false
		}
	}
	return true
}

func (e *Engine) resolveSpell(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	return e.resolveSpellWithChoices(g, obj, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveSpellWithChoices(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	card, spellDef, ok := cardInstanceFaceDef(g, obj.SourceID, obj.Face)
	if !ok {
		return "missing source"
	}
	if spellDef.IsPermanent() {
		if !spellHasAnyLegalTargets(g, spellDef, obj.Controller, obj.ChosenModes, obj.Targets) {
			if obj.Copy {
				return "countered by rules"
			}
			if !moveStackCardToGraveyard(g, obj, card) {
				return "invalid owner"
			}
			return "countered by rules"
		}
		if obj.Copy {
			return "resolved"
		}
		if obj.FaceDown {
			_, ok := createCardPermanentFaceDown(g, card, obj.Controller, game.ZoneStack, obj.FaceDownFace, obj.FaceDownKind)
			if !ok {
				return "invalid face-down"
			}
			return "battlefield"
		}
		permanent, ok := createCardPermanentFaceWithChoices(e, g, card, obj.Controller, game.ZoneStack, obj.Face, agents, log)
		if ok && obj.Suspend && permanentHasType(g, permanent, game.TypeCreature) {
			permanent.SuspendHasteController = opt.Val(obj.Controller)
		}
		if ok && isAttachmentPermanent(g, permanent) && len(obj.Targets) > 0 {
			target, targetOK := effectPermanent(g, obj, game.Effect{TargetIndex: 0})
			if !targetOK || !attachPermanent(g, permanent, target) {
				movePermanentToZone(g, permanent, game.ZoneGraveyard)
				return "graveyard"
			}
		}
		return "battlefield"
	}
	if spellDef.HasType(game.TypeInstant) || spellDef.HasType(game.TypeSorcery) {
		if !spellHasAnyLegalTargets(g, spellDef, obj.Controller, obj.ChosenModes, obj.Targets) {
			if obj.Copy {
				return "countered by rules"
			}
			if !moveStackCardToGraveyard(g, obj, card) {
				return "invalid owner"
			}
			return "countered by rules"
		}
		e.resolveSpellEffectsWithChoices(g, obj, card, agents, log)
		if obj.Copy {
			return "resolved"
		}
		if !moveStackCardToGraveyard(g, obj, card) {
			return "invalid owner"
		}
		return "graveyard"
	}
	return "resolved"
}

func moveStackCardToGraveyard(g *game.Game, obj *game.StackObject, card *game.CardInstance) bool {
	if _, ok := playerByID(g, card.Owner); !ok {
		return false
	}
	intendedDestination := game.ZoneGraveyard
	if obj != nil && obj.Flashback {
		// Flashback replaces any move from the stack to anywhere else with exile
		// after the spell was cast from a graveyard (CR 702.34a, CR 702.34c).
		intendedDestination = game.ZoneExile
	}
	destination := replacementZoneChangeDestination(g, game.GameEvent{
		Kind:          game.EventZoneChanged,
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FromZone:      game.ZoneStack,
		ToZone:        intendedDestination,
	})
	destination = commanderReplacementDestination(g, card.ID, destination)
	zone, ok := destinationZone(g, card.Owner, destination)
	if !ok {
		return false
	}
	zone.Add(card.ID)
	event := game.GameEvent{
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FromZone:      game.ZoneStack,
		ToZone:        destination,
	}
	emitZoneChangeEvent(g, event)
	return true
}

func stackObjectID(obj *game.StackObject) id.ID {
	return obj.ID
}

func stackObjectSourceID(obj *game.StackObject) id.ID {
	if obj.SourceCardID != 0 {
		return obj.SourceCardID
	}
	return obj.SourceID
}

func stackObjectController(obj *game.StackObject) game.PlayerID {
	return obj.Controller
}

func playerByID(g *game.Game, playerID game.PlayerID) (*game.Player, bool) {
	if playerID < 0 || int(playerID) >= len(g.Players) {
		return nil, false
	}
	return g.Players[playerID], true
}
