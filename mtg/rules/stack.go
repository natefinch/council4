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

// resolveTopOfStackWithChoices resolves the object on top of the stack.
// CR 608.1: each time all players pass in succession, the spell or ability on
// top of the stack resolves. The resolving object's last known information is
// snapshotted first so source-based checks (e.g. protection) still find the
// correct characteristics after it leaves the stack. The final destination is
// handled in the per-kind resolve paths: a permanent spell enters the
// battlefield (CR 608.3), while an instant or sorcery spell (and any ability) is
// put into the graveyard or removed from the stack (CR 608.2n).
func (e *Engine) resolveTopOfStackWithChoices(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	obj, ok := g.Stack.Pop()
	if !ok {
		return
	}
	// Snapshot the resolving spell's face characteristics into LKI before
	// effects run so that protection checks against the source can find the
	// correct face even after the object has been removed from the stack.
	if obj.Kind == game.StackSpell && obj.ID != 0 {
		snapshot := snapshotStackSpell(g, obj)
		rememberLastKnown(g, &snapshot)
	}
	startEntry := log.entryCount()
	eventsBefore := len(g.Events)
	result := e.resolveStackObjectWithChoices(g, obj, agents, log)
	releaseStateTriggerLatch(g, obj)
	if obj.Kind == game.StackSpell && spellResolved(result) {
		emitEvent(g, game.Event{
			Kind:          game.EventSpellResolved,
			SourceID:      obj.SourceID,
			StackObjectID: obj.ID,
			Controller:    obj.Controller,
			CardID:        obj.SourceID,
		})
	}
	recordEnteredBattlefield(g, log, eventsBefore)
	log.addResolve(ResolveLog{
		StackObjectID: obj.ID,
		SourceID:      obj.SourceID,
		Controller:    obj.Controller,
		Kind:          obj.Kind,
		Result:        result,
		SourceName:    stackObjectSourceName(g, obj),
		StartEntry:    startEntry,
	})
}

// recordEnteredBattlefield logs each permanent that entered the battlefield while
// the just-resolved stack object was resolving, scanning the events emitted since
// eventsBefore. These entries fall in the resolution's [StartEntry, resolve)
// range, so a report can show a fetched land or created token nested under the
// spell or ability that caused it. Priority-time entries (a land drop) are not
// scanned here, so they are not double-reported.
func recordEnteredBattlefield(g *game.Game, log *TurnLog, eventsBefore int) {
	if log == nil {
		return
	}
	for i := eventsBefore; i < len(g.Events); i++ {
		event := g.Events[i]
		if event.Kind != game.EventPermanentEnteredBattlefield {
			continue
		}
		log.addEnter(PermanentEnterLog{
			Permanent:  event.PermanentID,
			SourceID:   event.CardID,
			TokenName:  event.TokenName,
			Controller: event.Controller,
		})
	}
}

// stackObjectSourceName resolves the display name of a stack object's source —
// a token definition, the source card instance for an activated or triggered
// ability, or the card instance a spell was cast from.
func stackObjectSourceName(g *game.Game, obj *game.StackObject) string {
	if obj.SourceTokenDef != nil {
		return obj.SourceTokenDef.Name
	}
	if instance, ok := g.CardInstances[obj.SourceCardID]; ok && instance.Def != nil {
		return instance.Def.Name
	}
	if instance, ok := g.CardInstances[obj.SourceID]; ok && instance.Def != nil {
		return instance.Def.Name
	}
	return ""
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
	case "resolved", "battlefield", "graveyard", "exile", "adventure exile", "mutated":
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
	body := stackObjectActivatedBody(def, obj)
	if body == nil {
		return "missing source"
	}
	activatedBody, activatedOK := body.(*game.ActivatedAbility)
	if activatedOK && obj.Ninjutsu && game.BodyHasKeyword(activatedBody, game.Ninjutsu) {
		player, ok := playerByID(g, obj.Controller)
		if !ok || !player.Hand.Contains(obj.SourceCardID) {
			return "resolved"
		}

		card, ok := g.GetCardInstance(obj.SourceCardID)
		if !ok || card.ZoneVersion != obj.SourceZoneVersion || !player.Hand.Remove(obj.SourceCardID) {
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
	if permanentOK && activatedOK && isEquipmentPermanent(g, permanent) && bodyAttachesLikeEquip(activatedBody) {
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
	loyaltyBody, loyaltyOK := body.(*game.LoyaltyAbility)
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

func stackObjectActivatedBody(def *game.CardDef, obj *game.StackObject) game.Ability {
	if obj.InlineActivated != nil {
		return obj.InlineActivated
	}
	if obj.InlineLoyalty != nil {
		return obj.InlineLoyalty
	}
	return def.BodyAt(obj.AbilityIndex)
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
	body, ok := def.BodyAt(obj.AbilityIndex).(*game.TriggeredAbility)
	if !ok {
		return "missing source"
	}
	return e.resolveTriggeredAbilityBodyWithChoices(g, obj, def, body, agents, log)
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
	if !bodyHasAnyLegalTargetsFromSourceObject(g, source, obj.SourceID, body, obj) {
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
	wardKeyword, ok := game.BodyWardKeyword(ability)
	if !ok {
		return "resolved"
	}
	wardCost := wardKeyword.Cost
	cost := &wardCost
	request := payment.GenericRequest{PlayerID: payer, Cost: cost, AdditionalCosts: wardKeyword.AdditionalCosts}
	if paymentOrch.canPayGenericCost(g, request) && e.chooseMay(g, agents, payer, "Pay ward cost?", log) {
		prefs := e.paymentPreferencesForCost(g, payer, cost, wardKeyword.AdditionalCosts, 0, agents, log)
		request.Prefs = prefs
		if paymentOrch.payGenericCost(g, request) {
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

// counterStackObject counters the spell or ability with the given stack object
// ID. CR 701.6a: to counter a spell or ability is to cancel it, removing it from
// the stack; it doesn't resolve and none of its effects occur, and a countered
// spell is put into its owner's graveyard. A spell with a "can't be countered"
// effect is left on the stack instead. A countered ability or a spell copy
// simply ceases to exist, with no card to move.
func counterStackObject(g *game.Game, objectID id.ID) bool {
	obj, ok := stackObjectByID(g, objectID)
	if !ok {
		return false
	}
	switch obj.Kind {
	case game.StackSpell:
		if !stackSpellCanBeCountered(g, obj) {
			return false
		}
	case game.StackActivatedAbility, game.StackTriggeredAbility:
	default:
		return false
	}
	obj, ok = g.Stack.RemoveByID(objectID)
	if !ok {
		return false
	}
	if obj.Kind != game.StackSpell {
		releaseStateTriggerLatch(g, obj)
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

// bounceStackSpellToHand removes a spell from the stack and returns its card to
// its owner's hand ("Return target spell to its owner's hand."). Only spells
// have a card to return; spell copies cease to exist with no card moved, the
// same way a countered copy simply disappears.
func bounceStackSpellToHand(g *game.Game, obj *game.StackObject) bool {
	if obj.Kind != game.StackSpell {
		return false
	}
	removed, ok := g.Stack.RemoveByID(obj.ID)
	if !ok {
		return false
	}
	if removed.Copy {
		return true
	}
	card, ok := g.GetCardInstance(removed.SourceID)
	if !ok {
		return false
	}
	return moveStackCardToZone(g, removed, card, zone.Hand, false)
}

func releaseStateTriggerLatch(g *game.Game, obj *game.StackObject) {
	if obj.Kind != game.StackTriggeredAbility ||
		obj.InlineTrigger == nil ||
		!obj.InlineTrigger.Trigger.State.Exists {
		return
	}
	deleteStateTriggerLatch(g, obj.SourceID, obj.SourceCardID, obj.AbilityIndex)
}

func deleteStateTriggerLatch(g *game.Game, sourceObjectID, sourceCardID id.ID, abilityIndex int) {
	delete(g.StateTriggerLatches, game.StateTriggerKey{
		SourceObjectID: sourceObjectID,
		SourceCardID:   sourceCardID,
		AbilityIndex:   abilityIndex,
	})
}

// stackSpellCanBeCountered reports whether a spell on the stack may be
// countered (CR 701.6). A spell with a "can't be countered" effect applying to
// it isn't countered. Both the object-scoped effects captured when the spell was
// cast and the currently active rule effects are consulted.
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
	effects := activeRuleEffects(g)
	if spellCantBeCounteredByEffects(obj, spellDef, obj.RuleEffects) {
		return false
	}
	if spellCantBeCounteredByEffects(obj, spellDef, effects) {
		return false
	}
	return true
}

func spellCantBeCounteredByEffects(obj *game.StackObject, spellDef *game.CardDef, effects []game.RuleEffect) bool {
	for i := range effects {
		effect := &effects[i]
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
		return true
	}
	return false
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
		return e.resolveInstantOrSorcerySpell(g, obj, card, spellDef, agents, log)
	}
	return "resolved"
}

// resolveInstantOrSorcerySpell resolves a non-permanent spell (CR 608).
// CR 608.2b: if all of a spell's targets are illegal as it resolves, the spell
// doesn't resolve; it is removed from the stack and put into its owner's
// graveyard (counteredSpellResolution). Otherwise its instructions are followed.
// CR 608.2n: as the final part of an instant or sorcery spell's resolution it is
// put into its owner's graveyard, unless an effect (exile, shuffle, adventure,
// rebound) sends it elsewhere first.
func (e *Engine) resolveInstantOrSorcerySpell(
	g *game.Game,
	obj *game.StackObject,
	card *game.CardInstance,
	spellDef *game.CardDef,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
) string {
	if !spellHasAnyLegalTargets(g, spellDef, obj) {
		return counteredSpellResolution(g, obj, card)
	}
	e.resolveSpellEffectsWithChoices(g, obj, card, agents, log)
	if obj.Copy {
		return "resolved"
	}
	if obj.ExileOnResolution {
		if !moveStackCardToGraveyard(g, obj, card) {
			return "invalid owner"
		}
		return "exile"
	}
	if obj.ShuffleIntoLibraryOnResolution {
		if !moveStackCardToOwnersLibrary(g, obj, card) {
			return "invalid owner"
		}
		return "library"
	}
	if isAdventureAlternateFaceSpell(g, obj) {
		if !moveAdventureSpellToExile(g, obj, card) {
			return "invalid owner"
		}
		return "adventure exile"
	}
	if obj.SourceZone == zone.Hand && cardHasRebound(spellDef) {
		if !e.reboundExileResolvingSpell(g, obj, card) {
			return "invalid owner"
		}
		return "rebound exile"
	}
	if !moveStackCardToGraveyard(g, obj, card) {
		return "invalid owner"
	}
	return "graveyard"
}

func (e *Engine) resolvePermanentSpellWithChoices(g *game.Game, obj *game.StackObject, card *game.CardInstance, spellDef *game.CardDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if obj.Mutate {
		return e.resolveMutateSpell(g, obj, card, spellDef, agents, log)
	}
	if !spellHasAnyLegalTargets(g, spellDef, obj) {
		return counteredSpellResolution(g, obj, card)
	}
	if obj.Copy {
		return "resolved"
	}
	if obj.FaceDown {
		_, ok := createCardPermanentFaceDownWithChoices(e, g, card, obj.Controller, zone.Stack, obj.FaceDownFace, obj.FaceDownKind, !obj.Copy, agents, log)
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
		permanentCreationOptions{
			KickerPaid:              obj.KickerPaid,
			KickCount:               obj.KickerCount,
			Evoked:                  obj.Evoked,
			EntersTransformed:       obj.Converted,
			WasCast:                 !obj.Copy,
			CastController:          obj.Controller,
			HasCastController:       !obj.Copy,
			CastFromZone:            obj.SourceZone,
			XValue:                  obj.XValue,
			ColorsOfManaSpentToCast: obj.ColorsOfManaSpentToCast,
			ManaSpentByColorToCast:  obj.ManaSpentByColorToCast,
		},
		agents,
		log,
	)
	if ok && obj.Suspend && permanentHasType(g, permanent, types.Creature) {
		permanent.SuspendHasteController = opt.Val(obj.Controller)
	}
	if ok && len(obj.GainsKeywordsUntilEndOfTurn) > 0 && permanentHasType(g, permanent, types.Creature) {
		applyTypedContinuousEffects(g, obj, permanent, []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			AddKeywords: obj.GainsKeywordsUntilEndOfTurn,
		}}, game.DurationUntilEndOfTurn)
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

func (e *Engine) resolveMutateSpell(g *game.Game, obj *game.StackObject, card *game.CardInstance, spellDef *game.CardDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) string {
	if card == nil && !obj.Copy {
		return "missing source"
	}
	owner := obj.Controller
	if card != nil && !obj.Copy {
		owner = card.Owner
	}
	if !mutateTargetLegal(g, obj.Controller, owner, spellDef, obj.MutateTargetID) {
		if obj.Copy {
			if _, ok := createTokenPermanent(g, obj.Controller, copyCardDef(spellDef)); !ok {
				return "invalid copy"
			}
			return "battlefield"
		}
		_, ok := createCardPermanentFaceWithOptions(
			e,
			g,
			card,
			obj.Controller,
			zone.Stack,
			obj.Face,
			nil,
			permanentCreationOptions{
				WasCast:           true,
				CastController:    obj.Controller,
				HasCastController: true,
				CastFromZone:      obj.SourceZone,
			},
			agents,
			log,
		)
		if !ok {
			return "invalid owner"
		}
		return "battlefield"
	}
	target, ok := permanentByObjectID(g, obj.MutateTargetID)
	if !ok {
		return "missing target"
	}
	onTop := mutateSpellOnTop(e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:   game.ChoiceResolution,
		Player: obj.Controller,
		Prompt: "Put the mutating creature spell over or under the target creature.",
		Options: []game.ChoiceOption{
			{Index: 0, Label: "Over"},
			{Index: 1, Label: "Under"},
		},
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, log))
	if onTop {
		shiftPermanentAbilityUseIndexes(g, target.ObjectID, spellDef.AbilityCount())
		lower := game.MergedCard{
			CardInstanceID: target.CardInstanceID,
			Face:           target.Face,
			FaceDown:       target.FaceDown,
			FaceDownFace:   target.FaceDownFace,
			FaceDownKind:   target.FaceDownKind,
			TokenDef:       target.TokenDef,
			Owner:          target.Owner,
		}
		target.MergedCards = append([]game.MergedCard{lower}, target.MergedCards...)
		target.Face = obj.Face
		target.Owner = owner
		if obj.Copy {
			target.CardInstanceID = 0
			target.Token = true
			target.TokenDef = copyCardDef(spellDef)
		} else {
			target.CardInstanceID = card.ID
			target.Token = false
			target.TokenDef = nil
		}
		target.FaceDown = false
		target.FaceDownFace = game.FaceFront
		target.FaceDownKind = game.FaceDownNone
		target.Flipped = false
		target.Transformed = false
	} else {
		lower := game.MergedCard{Face: obj.Face, Owner: owner}
		if obj.Copy {
			lower.TokenDef = copyCardDef(spellDef)
		} else {
			lower.CardInstanceID = card.ID
		}
		target.MergedCards = append(target.MergedCards, lower)
	}
	zoneEvent := game.Event{
		Controller:  effectiveController(g, target),
		Player:      owner,
		Face:        obj.Face,
		PermanentID: target.ObjectID,
		FromZone:    zone.Stack,
		ToZone:      zone.Battlefield,
	}
	if obj.Copy {
		zoneEvent.TokenDef = copyCardDef(spellDef)
		zoneEvent.TokenName = spellDef.Name
	} else {
		zoneEvent.CardID = card.ID
	}
	emitZoneChangeEvent(g, zoneEvent)
	emitEvent(g, game.Event{
		Kind:           game.EventPermanentMutated,
		SourceID:       zoneEvent.CardID,
		SourceObjectID: target.ObjectID,
		StackObjectID:  obj.ID,
		Controller:     effectiveController(g, target),
		CardID:         zoneEvent.CardID,
		Face:           obj.Face,
		PermanentID:    target.ObjectID,
		TokenName:      zoneEvent.TokenName,
		TokenDef:       zoneEvent.TokenDef,
	})
	return "mutated"
}

func mutateSpellOnTop(selection []int) bool {
	return len(selection) == 1 && selection[0] == 0
}

func shiftPermanentAbilityUseIndexes(g *game.Game, sourceID id.ID, offset int) {
	if offset <= 0 {
		return
	}
	activated := make(map[game.ActivatedAbilityUse]bool, len(g.ActivatedAbilitiesThisTurn))
	for use, used := range g.ActivatedAbilitiesThisTurn {
		if use.SourceID == sourceID && use.AbilityIndex >= 0 {
			use.AbilityIndex += offset
		}
		activated[use] = used
	}
	g.ActivatedAbilitiesThisTurn = activated
	triggered := make(map[game.TriggeredAbilityUse]int, len(g.TriggeredAbilitiesThisTurn))
	for use, count := range g.TriggeredAbilitiesThisTurn {
		if use.SourceID == sourceID && use.AbilityIndex >= 0 {
			use.AbilityIndex += offset
		}
		triggered[use] = count
	}
	g.TriggeredAbilitiesThisTurn = triggered
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
	event := game.Event{
		Kind:          game.EventZoneChanged,
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FaceDown:      obj != nil && obj.FaceDown,
		FromZone:      zone.Stack,
		ToZone:        zone.Exile,
	}
	replacement := replacementZoneChange(g, event)
	destination := replacement.destination
	destination = commanderReplacementDestination(g, card.ID, destination)
	destinationCards, ok := destinationZone(g, card.Owner, destination)
	if !ok {
		return false
	}
	revealZoneReplacementSource(g, event, replacement.revealSource)
	destinationCards.Add(card.ID)
	shuffleLibraryIfRequested(g, destinationCards, destination, replacement.shuffleIntoLibrary)
	if destination == zone.Exile {
		if g.AdventureCards == nil {
			g.AdventureCards = make(map[id.ID]bool)
		}
		g.AdventureCards[card.ID] = true
	}
	event = game.Event{
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
	return moveStackCardToZone(g, obj, card, zone.Graveyard, false)
}

// moveStackCardToOwnersLibrary moves a resolving spell's card from the stack
// into its owner's library and shuffles, backing the "shuffle this card into
// its owner's library" resolution tail.
func moveStackCardToOwnersLibrary(g *game.Game, obj *game.StackObject, card *game.CardInstance) bool {
	return moveStackCardToZone(g, obj, card, zone.Library, true)
}

// moveStackCardToZone moves a spell's card from the stack to intendedDestination
// (its owner's graveyard when countered, or hand when bounced), honoring the
// flashback/exile-on-resolution replacement that diverts the card to exile
// instead (CR 702.34c) and the commander zone-change replacement. When
// forceShuffle is set and the card lands in a library, that library is shuffled
// even without a replacement requesting it, backing the "shuffle this card into
// its owner's library" resolution tail.
func moveStackCardToZone(g *game.Game, obj *game.StackObject, card *game.CardInstance, intendedDestination zone.Type, forceShuffle bool) bool {
	if _, ok := playerByID(g, card.Owner); !ok {
		return false
	}
	if obj != nil && (obj.Flashback || obj.ExileOnResolution) {
		// Flashback replaces any move from the stack to anywhere else with exile
		// after the spell was cast from a graveyard (CR 702.34a, CR 702.34c).
		intendedDestination = zone.Exile
	} else if obj != nil {
		// A counter-and-redirect replacement diverts the countered spell's card
		// to a non-graveyard zone other than exile (Memory Lapse, Remand). Top of
		// library is honored by Add below, which adds to the top of a zone.
		switch obj.CounteredDestination {
		case game.CounteredSpellLibraryTop:
			intendedDestination = zone.Library
		case game.CounteredSpellHand:
			intendedDestination = zone.Hand
		default:
			// CounteredSpellGraveyard keeps the caller's intended destination.
		}
	}
	event := game.Event{
		Kind:          game.EventZoneChanged,
		SourceID:      card.ID,
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        card.Owner,
		CardID:        card.ID,
		Face:          stackObjectFace(obj),
		FaceDown:      obj != nil && obj.FaceDown,
		FromZone:      zone.Stack,
		ToZone:        intendedDestination,
	}
	replacement := replacementZoneChange(g, event)
	destination := replacement.destination
	destination = commanderReplacementDestination(g, card.ID, destination)
	destinationCards, ok := destinationZone(g, card.Owner, destination)
	if !ok {
		return false
	}
	revealZoneReplacementSource(g, event, replacement.revealSource)
	destinationCards.Add(card.ID)
	shuffleLibraryIfRequested(g, destinationCards, destination, replacement.shuffleIntoLibrary || forceShuffle)
	event = game.Event{
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
