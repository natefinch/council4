package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
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
		emitEvent(g, game.Event{
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
	case "resolved", "battlefield", "graveyard", "adventure exile":
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
	if !defOK {
		return "missing source"
	}
	body := def.BodyAt(obj.AbilityIndex)
	if body == nil {
		return "missing source"
	}
	activatedBody, activatedOK := body.(game.ActivatedAbility)
	if activatedOK && obj.Ninjutsu && game.BodyHasKeyword(activatedBody, game.Ninjutsu) {
		player, ok := playerByID(g, obj.Controller)
		if !ok || !player.Hand.Contains(obj.SourceCardID) {
			return "resolved"
		}
		card, ok := g.GetCardInstance(obj.SourceCardID)
		if !ok || !player.Hand.Remove(obj.SourceCardID) {
			return "missing source"
		}
		ninja, ok := createCardPermanentFaceWithOptions(
			e,
			g,
			card,
			obj.Controller,
			zone.Hand,
			game.FaceFront,
			nil,
			permanentCreationOptions{ForceTapped: true},
			agents,
			log,
		)
		if !ok {
			player.Hand.Add(obj.SourceCardID)
			return "missing source"
		}
		if g.Combat != nil && ninjutsuAttackTargetValid(g, obj.NinjutsuAttackTarget) {
			g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
				Attacker: ninja.ObjectID,
				Target:   obj.NinjutsuAttackTarget,
			})
		}
		return "resolved"
	}
	if permanentOK && activatedOK && isEquipmentPermanent(g, permanent) && game.BodyHasKeyword(activatedBody, game.Equip) {
		sourceObjectID := obj.SourceID
		if !permanentOK {
			sourceObjectID = 0
		}
		if !bodyHasAnyLegalTargetsFromSourceObject(g, def, sourceObjectID, activatedBody, obj) {
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

	if activatedOK {
		if !bodyHasAnyLegalTargetsFromSourceObject(g, def, obj.SourceID, activatedBody, obj) {
			return "countered by rules"
		}
		if len(activatedBody.Content.Modes) > 0 {
			e.resolveAbilityContentWithChoices(g, obj, activatedBody.Content, agents, log)
		}
		return "resolved"
	}
	loyaltyBody, loyaltyOK := body.(game.LoyaltyAbility)
	if !loyaltyOK {
		return "missing source"
	}
	if !bodyHasAnyLegalTargetsFromSourceObject(g, def, obj.SourceID, loyaltyBody, obj) {
		return "countered by rules"
	}
	if len(loyaltyBody.Content.Modes) > 0 {
		e.resolveAbilityContentWithChoices(g, obj, loyaltyBody.Content, agents, log)
		return "resolved"
	}
	return "resolved"
}

func ninjutsuAttackTargetValid(g *game.Game, target game.AttackTarget) bool {
	if !isPlayerAlive(g, target.Player) {
		return false
	}
	if target.IsPlayerAttack() {
		return true
	}
	permanent, ok := attackTargetPermanent(g, target)
	return ok && !permanent.PhasedOut
}

func (e *Engine) resolveTriggeredAbility(g *game.Game, obj *game.StackObject, log *TurnLog) string {
	return e.resolveTriggeredAbilityWithChoices(g, obj, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveTriggeredAbilityWithChoices(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if obj.InlineTrigger != nil {
		source, _ := stackObjectSourceDef(g, obj)
		return e.resolveTriggeredAbilityBodyWithChoices(g, obj, source, obj.InlineTrigger, agents, log)
	}
	def, ok := stackObjectSourceDef(g, obj)
	if !ok {
		return "missing source"
	}
	body, ok := def.BodyAt(obj.AbilityIndex).(game.TriggeredAbility)
	if !ok {
		return "missing source"
	}
	return e.resolveTriggeredAbilityBodyWithChoices(g, obj, def, &body, agents, log)
}

func (e *Engine) resolveTriggeredAbilityBodyWithChoices(g *game.Game, obj *game.StackObject, source *game.CardDef, body *game.TriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if body == nil {
		return "missing source"
	}
	if _, ok := game.BodyWardCost(body); ok {
		return e.resolveWardTriggeredAbilityWithChoices(g, obj, body, agents, log)
	}
	if _, ok := game.BodyMadnessCost(body); ok {
		return e.resolveMadnessTriggeredAbilityWithChoices(g, obj, body, agents, log)
	}
	var event *game.Event
	if obj.HasTriggerEvent {
		event = &obj.TriggerEvent
	}
	sourcePermanent, _ := permanentByObjectID(g, obj.SourceID)
	if !triggerInterveningIf(g, sourcePermanent, obj.Controller, &body.Trigger, event) {
		return "intervening if false"
	}
	if !bodyHasAnyLegalTargetsFromSourceObject(g, source, obj.SourceID, *body, obj) {
		return "countered by rules"
	}
	if body.Optional && !e.chooseMay(g, agents, obj.Controller, "Apply optional triggered ability?", log) {
		return "declined"
	}
	e.resolveAbilityContentWithChoices(g, obj, body.Content, agents, log)
	return "resolved"
}

func (e *Engine) resolveWardTriggeredAbilityWithChoices(g *game.Game, obj *game.StackObject, ability *game.TriggeredAbility, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	targetObj, ok := stackObjectByID(g, obj.WardTargetStackObjectID)
	if !ok {
		return "resolved"
	}
	payer := targetObj.Controller
	wardCost, ok := game.BodyWardCost(ability)
	if !ok {
		return "resolved"
	}
	cost := &wardCost
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
	var spellDef *game.CardDef
	var ok bool
	if obj.SourceTokenDef != nil {
		spellDef, ok = obj.SourceTokenDef.FaceDef(obj.Face)
	} else {
		_, spellDef, ok = cardInstanceFaceDef(g, obj.SourceID, obj.Face)
	}
	if !ok {
		return true
	}
	for _, effect := range activeRuleEffects(g) {
		if effect.Kind != game.RuleEffectCantBeCountered {
			continue
		}
		if effect.AffectedObjectID != 0 && effect.AffectedObjectID != obj.ID {
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

func spellTypesMatch(card *game.CardDef, cardTypes []types.Card) bool {
	for _, cardType := range cardTypes {
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
	var card *game.CardInstance
	var spellDef *game.CardDef
	var ok bool
	if obj.SourceTokenDef != nil {
		spellDef, ok = obj.SourceTokenDef.FaceDef(obj.Face)
	} else {
		card, spellDef, ok = cardInstanceFaceDef(g, obj.SourceID, obj.Face)
	}
	if !ok {
		return "missing source"
	}
	if spellDef.IsPermanent() {
		return e.resolvePermanentSpellWithChoices(g, obj, card, spellDef, agents, log)
	}
	if spellDef.HasType(types.Instant) || spellDef.HasType(types.Sorcery) {
		if !spellHasAnyLegalTargets(g, spellDef, obj) {
			return counteredSpellResolution(g, obj, card)
		}
		e.resolveSpellEffectsWithChoices(g, obj, card, agents, log)
		if obj.Copy {
			return "resolved"
		}
		if isAdventureAlternateFaceSpell(g, obj) {
			if !moveAdventureSpellToExile(g, obj, card) {
				return "invalid owner"
			}
			return "adventure exile"
		}
		if !moveStackCardToGraveyard(g, obj, card) {
			return "invalid owner"
		}
		return "graveyard"
	}
	return "resolved"
}

func (e *Engine) resolvePermanentSpellWithChoices(g *game.Game, obj *game.StackObject, card *game.CardInstance, spellDef *game.CardDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if !spellHasAnyLegalTargets(g, spellDef, obj) {
		return counteredSpellResolution(g, obj, card)
	}
	if obj.Copy {
		return "resolved"
	}
	if obj.FaceDown {
		_, ok := createCardPermanentFaceDown(g, card, obj.Controller, zone.Stack, obj.FaceDownFace, obj.FaceDownKind, !obj.Copy)
		if !ok {
			return "invalid face-down"
		}
		return "battlefield"
	}
	permanent, ok := createCardPermanentFaceWithOptions(
		e,
		g,
		card,
		obj.Controller,
		zone.Stack,
		obj.Face,
		nil,
		permanentCreationOptions{KickerPaid: obj.KickerPaid, WasCast: !obj.Copy},
		agents,
		log,
	)
	if ok && obj.Suspend && permanentHasType(g, permanent, types.Creature) {
		permanent.SuspendHasteController = opt.Val(obj.Controller)
	}
	if ok && isAttachmentPermanent(g, permanent) && len(obj.Targets) > 0 {
		target, targetOK := effectPermanentAt(g, obj, 0)
		if !targetOK || !attachPermanent(g, permanent, target) {
			movePermanentToZone(g, permanent, zone.Graveyard)
			return "graveyard"
		}
	}
	return "battlefield"
}

func counteredSpellResolution(g *game.Game, obj *game.StackObject, card *game.CardInstance) string {
	if obj.Copy {
		return "countered by rules"
	}
	if !moveStackCardToGraveyard(g, obj, card) {
		return "invalid owner"
	}
	return "countered by rules"
}

func isAdventureAlternateFaceSpell(g *game.Game, obj *game.StackObject) bool {
	if obj.Face != game.FaceAlternate {
		return false
	}
	card, ok := g.GetCardInstance(obj.SourceID)
	if !ok {
		return false
	}
	return card.Def.Layout == game.LayoutAdventure
}

func moveAdventureSpellToExile(g *game.Game, obj *game.StackObject, card *game.CardInstance) bool {
	if _, ok := playerByID(g, card.Owner); !ok {
		return false
	}
	destination := replacementZoneChangeDestination(g, game.Event{
		Kind:          game.EventZoneChanged,
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FromZone:      zone.Stack,
		ToZone:        zone.Exile,
	})
	destination = commanderReplacementDestination(g, card.ID, destination)
	destinationCards, ok := destinationZone(g, card.Owner, destination)
	if !ok {
		return false
	}
	destinationCards.Add(card.ID)
	if destination == zone.Exile {
		if g.AdventureCards == nil {
			g.AdventureCards = make(map[id.ID]bool)
		}
		g.AdventureCards[card.ID] = true
	}
	event := game.Event{
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FromZone:      zone.Stack,
		ToZone:        destination,
	}
	emitZoneChangeEvent(g, event)
	return true
}

func moveStackCardToGraveyard(g *game.Game, obj *game.StackObject, card *game.CardInstance) bool {
	if _, ok := playerByID(g, card.Owner); !ok {
		return false
	}
	intendedDestination := zone.Graveyard
	if obj != nil && obj.Flashback {
		// Flashback replaces any move from the stack to anywhere else with exile
		// after the spell was cast from a graveyard (CR 702.34a, CR 702.34c).
		intendedDestination = zone.Exile
	}
	destination := replacementZoneChangeDestination(g, game.Event{
		Kind:          game.EventZoneChanged,
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FromZone:      zone.Stack,
		ToZone:        intendedDestination,
	})
	destination = commanderReplacementDestination(g, card.ID, destination)
	destinationCards, ok := destinationZone(g, card.Owner, destination)
	if !ok {
		return false
	}
	destinationCards.Add(card.ID)
	event := game.Event{
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FromZone:      zone.Stack,
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
