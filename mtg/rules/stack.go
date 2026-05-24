package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func (e *Engine) resolveTopOfStack(g *game.Game, log *TurnLog) {
	e.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveTopOfStackWithChoices(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if g == nil {
		return
	}
	obj := g.Stack.Pop()
	if obj == nil {
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
	permanent := permanentByObjectID(g, obj.SourceID)
	def := stackObjectSourceDef(g, obj)
	if def == nil && permanent != nil {
		def = permanentCardDef(g, permanent)
	}
	if def == nil || obj.AbilityIndex < 0 || obj.AbilityIndex >= len(def.Abilities) {
		return "missing source"
	}
	ability := &def.Abilities[obj.AbilityIndex]
	if isEquipmentPermanent(g, permanent) && abilityHasKeyword(ability, game.Equip) {
		if !abilityHasAnyLegalTargetsFromSourceObject(g, def, obj.SourceID, ability, obj.Controller, obj.Targets) {
			return "countered by rules"
		}
		if len(obj.Targets) != 1 || obj.Targets[0].Kind != game.TargetPermanent {
			return "countered by rules"
		}
		target := permanentByObjectID(g, obj.Targets[0].PermanentID)
		if !attachPermanent(g, permanent, target) {
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
	def := stackObjectSourceDef(g, obj)
	if def == nil || obj.AbilityIndex < 0 || obj.AbilityIndex >= len(def.Abilities) {
		return "missing source"
	}
	ability := &def.Abilities[obj.AbilityIndex]
	return e.resolveTriggeredAbilityDefWithChoices(g, obj, def, ability, agents, log)
}

func (e *Engine) resolveTriggeredAbilityDefWithChoices(g *game.Game, obj *game.StackObject, source *game.CardDef, ability *game.AbilityDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if ability.Kind != game.TriggeredAbility {
		return "missing source"
	}
	var event *game.GameEvent
	if obj.HasTriggerEvent {
		event = &obj.TriggerEvent
	}
	if ability.Trigger != nil && !triggerInterveningIf(g, obj.Controller, ability.Trigger, event) {
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

func (e *Engine) chooseMay(g *game.Game, agents [game.NumPlayers]PlayerAgent, player game.PlayerID, prompt string, log *TurnLog) bool {
	selected := e.chooseChoice(g, agents, mayChoiceRequest(player, prompt), log)
	return len(selected) == 1 && selected[0] == 1
}

func stackObjectSourceDef(g *game.Game, obj *game.StackObject) *game.CardDef {
	if obj == nil {
		return nil
	}
	if obj.SourceCardID != 0 {
		card := g.GetCardInstance(obj.SourceCardID)
		if card == nil {
			return nil
		}
		return card.Def
	}
	return obj.SourceTokenDef
}

func (e *Engine) resolveSpell(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	return e.resolveSpellWithChoices(g, obj, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveSpellWithChoices(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	card := g.GetCardInstance(obj.SourceID)
	if card == nil || card.Def == nil {
		return "missing source"
	}
	if card.Def.IsPermanent() {
		if !spellHasAnyLegalTargets(g, card.Def, obj.Controller, obj.ChosenModes, obj.Targets) {
			if !moveStackCardToGraveyard(g, obj, card) {
				return "invalid owner"
			}
			return "countered by rules"
		}
		permanent := createCardPermanent(g, card, obj.Controller, game.ZoneStack)
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
		if !spellHasAnyLegalTargets(g, card.Def, obj.Controller, obj.ChosenModes, obj.Targets) {
			if !moveStackCardToGraveyard(g, obj, card) {
				return "invalid owner"
			}
			return "countered by rules"
		}
		e.resolveSpellEffectsWithChoices(g, obj, card, agents, log)
		if !moveStackCardToGraveyard(g, obj, card) {
			return "invalid owner"
		}
		return "graveyard"
	}
	return "resolved"
}

func moveStackCardToGraveyard(g *game.Game, obj *game.StackObject, card *game.CardInstance) bool {
	if card == nil {
		return false
	}
	owner := playerByID(g, card.Owner)
	if owner == nil {
		return false
	}
	destination := commanderReplacementDestination(g, card.ID, game.ZoneGraveyard)
	zone := destinationZone(g, card.Owner, destination)
	if zone == nil {
		return false
	}
	zone.Add(card.ID)
	event := game.GameEvent{
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		FromZone:      game.ZoneStack,
		ToZone:        destination,
	}
	emitZoneChangeEvent(g, event)
	return true
}

func stackObjectID(obj *game.StackObject) id.ID {
	if obj == nil {
		return 0
	}
	return obj.ID
}

func stackObjectSourceID(obj *game.StackObject) id.ID {
	if obj == nil {
		return 0
	}
	if obj.SourceCardID != 0 {
		return obj.SourceCardID
	}
	return obj.SourceID
}

func stackObjectController(obj *game.StackObject) game.PlayerID {
	if obj == nil {
		return 0
	}
	return obj.Controller
}

func playerByID(g *game.Game, playerID game.PlayerID) *game.Player {
	if g == nil || playerID < 0 || int(playerID) >= len(g.Players) {
		return nil
	}
	return g.Players[playerID]
}
