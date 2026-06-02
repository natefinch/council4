package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) resolveSpellEffects(g *game.Game, obj *game.StackObject, card *game.CardInstance, log *TurnLog) {
	e.resolveSpellEffectsWithChoices(g, obj, card, [game.NumPlayers]PlayerAgent{}, log)
}

func (e *Engine) resolveSpellEffectsWithChoices(g *game.Game, obj *game.StackObject, card *game.CardInstance, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if e.resolveCardImplementationSpell(g, obj, card, log) {
		return
	}
	spellDef := cardFaceOrDefault(card, obj.Face)
	ability, ok := firstSpellAbility(spellDef)
	if !ok {
		return
	}
	if len(ability.Modes) > 0 {
		for _, modeIndex := range obj.ChosenModes {
			if modeIndex < 0 || modeIndex >= len(ability.Modes) {
				continue
			}
			for i := range ability.Modes[modeIndex].Effects {
				e.resolveEffectWithChoices(g, obj, &ability.Modes[modeIndex].Effects[i], agents, log)
			}
		}
		return
	}
	for i := range ability.Effects {
		e.resolveEffectWithChoices(g, obj, &ability.Effects[i], agents, log)
	}
	if obj.KickerPaid {
		for i := range ability.KickerEffects {
			e.resolveEffectWithChoices(g, obj, &ability.KickerEffects[i], agents, log)
		}
	}
}

func spellHasKicker(card *game.CardDef) bool {
	ability, ok := firstSpellAbility(card)
	return ok && ability.KickerCost.Exists
}

func firstSpellAbility(card *game.CardDef) (*game.AbilityDef, bool) {
	for i := range card.Abilities {
		if card.Abilities[i].Kind == game.SpellAbility {
			return &card.Abilities[i], true
		}
	}
	return nil, false
}

func (e *Engine) resolveEffect(g *game.Game, obj *game.StackObject, effect *game.Effect, log *TurnLog) {
	newEffectResolver(e, g, obj, [game.NumPlayers]PlayerAgent{}, log).resolve(effect)
}

func (e *Engine) resolveEffectWithChoices(g *game.Game, obj *game.StackObject, effect *game.Effect, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	newEffectResolver(e, g, obj, agents, log).resolve(effect)
}

// effectResolver bundles the per-resolution context so the resolution body
// can be a method rather than a free function with five repeated parameters.
type effectResolver struct {
	engine *Engine
	game   *game.Game
	obj    *game.StackObject
	agents [game.NumPlayers]PlayerAgent
	log    *TurnLog
}

func newEffectResolver(e *Engine, g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) *effectResolver {
	return &effectResolver{engine: e, game: g, obj: obj, agents: agents, log: log}
}

// effectResolved captures the outcome of executing one effect: whether it was
// accepted by the player (for optional effects), whether it successfully
// applied, and the computed amount (used by linked "that much" follow-ups,
// CR 608.2c).
type effectResolved struct {
	accepted     bool
	succeeded    bool
	amount       int
	excessDamage int
}

// record writes the resolution state into the stack object so that follow-up
// "if you do" / "that much" effects see what actually happened
// (CR 608.2c; impossible actions CR 101.3).
func (res effectResolved) record(obj *game.StackObject, effect *game.Effect) {
	if res.accepted && res.succeeded {
		rememberEffectAmount(obj, effect, res.amount)
		rememberEffectExcessDamage(obj, effect, res.excessDamage)
	}
	rememberEffectResolutionResult(obj, effect, res.accepted, res.succeeded, res.amount)
}

// amount returns the computed effect amount, resolving any dynamic formula.
func (r *effectResolver) amount(effect *game.Effect) int {
	return effectAmount(r.game, r.obj, effect)
}

// permanent resolves the target permanent for this effect, using the effect's
// TargetIndex to look up the chosen target on the stack object.
func (r *effectResolver) permanent(effect *game.Effect) (*game.Permanent, bool) {
	return effectPermanent(r.game, r.obj, effect)
}

// player resolves the target player for this effect.
func (r *effectResolver) player(effect *game.Effect) (game.PlayerID, bool) {
	return effectPlayer(r.game, r.obj, effect)
}

func (r *effectResolver) effectRecipientOrPlayer(effect *game.Effect) (game.PlayerID, bool) {
	if effect.Recipient.Exists {
		return resolvePlayerReference(r.game, r.obj, effect.Recipient.Val)
	}
	return r.player(effect)
}

func (r *effectResolver) effectRecipientOrController(effect *game.Effect) (game.PlayerID, bool) {
	if effect.Recipient.Exists {
		return resolvePlayerReference(r.game, r.obj, effect.Recipient.Val)
	}
	return r.obj.Controller, true
}

// manaColor returns the mana color for an add-mana effect, respecting any
// resolution choice that overrides the effect's declared color.
func (r *effectResolver) manaColor(effect *game.Effect) mana.Color {
	return effectManaColor(r.obj, effect)
}

// resolve checks conditions and then executes the effect, recording the result
// for any linked follow-up effects.
func (r *effectResolver) resolve(effect *game.Effect) {
	if !effectConditionSatisfied(r.game, r.obj, effect.Condition) {
		return
	}
	if !cardConditionSatisfied(r.game, r.obj, effect.CardCondition) {
		return
	}
	if !effectResultConditionSatisfied(r.obj, effect.ResultCondition) {
		return
	}
	res := r.executeEffect(effect)
	res.record(r.obj, effect)
}

// executeEffect runs the effect instruction and returns the outcome. It does
// not record the result; the caller (resolve) handles that so the deferred
// memory write is explicit rather than scattered through the switch.
// Each branch mutates res before returning so early returns keep the same
// state the old deferred recorder observed.
//
//nolint:maintidx // Effect dispatch is intentionally centralized around the rules effect enum.
func (r *effectResolver) executeEffect(effect *game.Effect) (res effectResolved) {
	res.accepted = true
	res.amount = r.amount(effect)
	if effect.Optional && !r.engine.chooseMay(r.game, r.agents, stackObjectController(r.obj), "Apply optional effect?", r.log) {
		res.accepted = false
		return res
	}
	if effect.Choice.Exists {
		if !r.engine.resolveResolutionChoice(r.game, r.obj, effect, r.agents, r.log) {
			return res
		}
		res.succeeded = true
		if effect.Type == game.EffectChoose {
			return res
		}
	}
	if effect.Payment.Exists {
		res.accepted, res.succeeded = r.engine.resolveResolutionPayment(r.game, r.obj, effect, r.agents, r.log)
		if !res.succeeded || effect.Type == game.EffectPay {
			return res
		}
	}
	if !IsEffectTypeExecuted(effect.Type) {
		logUnsupportedEffect(r.log, r.obj, effect)
		return res
	}
	if effect.Selector != game.EffectSelectorNone {
		res.succeeded = resolveMassPermanentEffect(r.game, r.obj, effect, res.amount)
		return res
	}
	if effect.PlayerSelector != game.PlayerSelectorNone {
		res.succeeded = resolveMassPlayerEffect(r.game, r.obj, effect, res.amount)
		return res
	}
	switch effect.Type {
	case game.EffectDraw:
		if res.amount <= 0 {
			return res
		}
		playerID, ok := r.player(effect)
		if !ok {
			return res
		}
		res.succeeded = r.engine.drawCards(r.game, playerID, res.amount, r.log)
	case game.EffectGainLife:
		if res.amount <= 0 {
			return res
		}
		playerID, ok := r.player(effect)
		if !ok {
			return res
		}
		res.succeeded = gainLife(r.game, playerID, res.amount) > 0
	case game.EffectLoseLife:
		if res.amount <= 0 {
			return res
		}
		playerID, ok := r.player(effect)
		if !ok {
			return res
		}
		res.succeeded = loseLife(r.game, playerID, res.amount) > 0
	case game.EffectAddMana:
		if res.amount <= 0 {
			res.amount = 1
		}
		player, ok := playerByID(r.game, r.obj.Controller)
		if !ok || player.Eliminated {
			return res
		}
		if stackObjectSourceIsSnow(r.game, r.obj) {
			player.ManaPool.AddSnow(r.manaColor(effect), res.amount)
		} else {
			player.ManaPool.Add(r.manaColor(effect), res.amount)
		}
		res.succeeded = true
	case game.EffectDamage:
		if res.amount <= 0 {
			return res
		}
		source, ok := resolveEffectDamageSource(r.game, r.obj, effect)
		if !ok {
			return res
		}
		if playerID, ok := r.player(effect); ok {
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, playerID, res.amount, false)
			if source.permanent != nil {
				applyLifelink(r.game, source.permanent, dealt)
			}
			res.amount = damageResultAmount(effect, dealt, 0)
			res.succeeded = dealt > 0
			return res
		}
		permanent, ok := r.permanent(effect)
		if !ok {
			return res
		}
		lethalRemaining := lethalDamageRemaining(r.game, permanent)
		if source.permanent != nil {
			lethalRemaining = lethalDamageRemainingFromSource(r.game, source.permanent, permanent)
		}
		dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
		applyDamageSourceKeywordEffects(r.game, source.permanent, permanent, dealt)
		res.excessDamage = max(0, dealt-lethalRemaining)
		res.amount = damageResultAmount(effect, dealt, res.excessDamage)
		res.succeeded = dealt > 0 && (effect.ResultAmount != game.EffectResultAmountExcessDamage || res.excessDamage > 0)
	case game.EffectDestroy:
		permanent, ok := r.permanent(effect)
		if !ok {
			return res
		}
		_, res.succeeded = destroyPermanent(r.game, permanent.ObjectID)
	case game.EffectCounter:
		res.succeeded = counterTargetStackObject(r.game, r.obj, effect)
	case game.EffectExile:
		permanent, ok := r.permanent(effect)
		if !ok {
			return res
		}
		linkedObjectRef := permanentLinkedObjectRef(permanent)
		res.succeeded = movePermanentToZone(r.game, permanent, game.ZoneExile)
		if effect.LinkID != "" {
			rememberLinkedObject(r.game, linkedObjectSourceKey(r.game, r.obj, effect.LinkID), linkedObjectRef)
		}
	case game.EffectBounce:
		permanent, ok := r.permanent(effect)
		if !ok {
			return res
		}
		res.succeeded = movePermanentToZone(r.game, permanent, game.ZoneHand)
	case game.EffectSacrifice:
		permanent, ok := r.permanent(effect)
		if !ok {
			permanent, ok = firstPermanentControlledBy(r.game, r.obj.Controller)
		}
		if !ok || effectiveController(r.game, permanent) != r.obj.Controller {
			return res
		}
		res.succeeded = movePermanentToZone(r.game, permanent, game.ZoneGraveyard)
	case game.EffectDiscard:
		playerID, ok := r.player(effect)
		if ok {
			res.succeeded = discardCards(r.game, playerID, res.amount)
		}
	case game.EffectTap:
		if permanent, ok := r.permanent(effect); ok {
			setPermanentTapped(r.game, permanent, true)
			res.succeeded = true
		}
	case game.EffectUntap:
		if permanent, ok := r.permanent(effect); ok {
			setPermanentTapped(r.game, permanent, false)
			res.succeeded = true
		}
	case game.EffectModifyPT:
		if permanent, ok := r.permanent(effect); ok && effect.UntilEndOfTurn {
			r.game.ContinuousEffects = append(r.game.ContinuousEffects, untilEndOfTurnContinuousEffect(r.game, r.obj, permanent, effect))
			res.succeeded = true
		}
	case game.EffectAddCounter:
		if permanent, ok := r.permanent(effect); ok && res.amount > 0 {
			permanent.Counters.Add(effect.CounterKind, res.amount)
			res.succeeded = true
		}
	case game.EffectRemoveCounter:
		if permanent, ok := r.permanent(effect); ok && res.amount > 0 {
			permanent.Counters.Remove(effect.CounterKind, res.amount)
			res.succeeded = true
		}
	case game.EffectMoveCounters:
		res.succeeded = moveCounters(r.game, r.obj, effect)
	case game.EffectApplyContinuous:
		permanent, _ := r.permanent(effect)
		res.succeeded = applyContinuousEffectTemplates(r.game, r.obj, permanent, effect)
	case game.EffectCreateToken:
		if res.amount <= 0 {
			res.amount = 1
		}
		if !effect.Token.Exists && !effect.TokenCopy.Exists {
			return res
		}
		recipient := r.obj.Controller
		if effect.Recipient.Exists {
			var ok bool
			recipient, ok = resolvePlayerReference(r.game, r.obj, effect.Recipient.Val)
			if !ok {
				return res
			}
		}
		for range res.amount {
			token, ok := r.tokenDefinition(effect)
			if !ok {
				return res
			}
			if _, ok := createTokenPermanent(r.game, recipient, token); !ok {
				return res
			}
		}
		res.succeeded = res.amount > 0
	case game.EffectInvestigate:
		if res.amount <= 0 {
			res.amount = 1
		}
		recipient := r.obj.Controller
		if effect.Recipient.Exists {
			var ok bool
			recipient, ok = resolvePlayerReference(r.game, r.obj, effect.Recipient.Val)
			if !ok {
				return res
			}
		}
		for range res.amount {
			if _, ok := createTokenPermanent(r.game, recipient, clueTokenDef()); !ok {
				return res
			}
		}
		res.succeeded = true
	case game.EffectCreateDelayedTrigger:
		res.succeeded = effect.DelayedTrigger.Exists && scheduleDelayedTrigger(r.game, r.obj, &effect.DelayedTrigger.Val)
	case game.EffectPutOnBattlefield:
		if effect.Card.Exists {
			res.succeeded = r.putReferencedCardOnBattlefield(effect)
		} else if effect.LinkID != "" {
			res.succeeded = r.putLinkedCardOnBattlefield(effect)
			if !res.succeeded {
				res.succeeded = returnLinkedExiledObjects(r.engine, r.game, r.obj, effect.LinkID, r.agents, r.log)
			}
		}
	case game.EffectPrevent:
		res.succeeded = createPreventionShield(r.game, r.obj, effect)
	case game.EffectRegenerate:
		if permanent, ok := r.permanent(effect); ok {
			permanent.RegenerationShields++
			res.succeeded = true
		}
	case game.EffectSkipStep:
		playerID, ok := r.player(effect)
		if ok {
			scheduleSkipStep(r.game, playerID, effect.Step)
			res.succeeded = true
		}
	case game.EffectTransform:
		if permanent, ok := r.permanent(effect); ok {
			res.succeeded = transformPermanent(r.game, permanent)
		}
	case game.EffectPhaseOut:
		if permanent, ok := r.permanent(effect); ok {
			permanent.PhasedOut = true
			removePermanentFromCombat(r.game, permanent.ObjectID)
			res.succeeded = true
		}
	case game.EffectCreateEmblem:
		r.game.Emblems = append(r.game.Emblems, game.Emblem{Owner: r.obj.Controller, Abilities: append([]game.AbilityDef(nil), effect.EmblemAbilities...)})
		res.succeeded = true
	case game.EffectMill:
		playerID, ok := r.player(effect)
		if ok {
			millCards(r.game, playerID, res.amount)
			res.succeeded = res.amount > 0
		}
	case game.EffectSearch:
		if !effect.Search.Exists || !searchSpecSupported(effect.Search.Val) {
			logUnsupportedEffect(r.log, r.obj, effect)
			return res
		}
		playerID, ok := r.player(effect)
		if ok {
			res.succeeded = r.engine.searchLibrary(r.game, r.obj, playerID, effect.Search.Val, res.amount)
		}
	case game.EffectReveal:
		playerID, ok := r.effectRecipientOrPlayer(effect)
		if ok {
			revealed := revealCardIDs(r.game, r.obj, playerID, game.ZoneLibrary, res.amount)
			if effect.LinkID != "" {
				for _, cardID := range revealed {
					rememberLinkedObject(r.game, linkedObjectSourceKey(r.game, r.obj, effect.LinkID), game.LinkedObjectRef{CardID: cardID})
				}
			}
			res.succeeded = len(revealed) > 0
		}
	case game.EffectScry:
		playerID, ok := r.player(effect)
		if ok {
			r.engine.scryCards(r.game, r.agents, r.log, playerID, res.amount)
			res.succeeded = res.amount > 0
		}
	case game.EffectSurveil:
		playerID, ok := r.player(effect)
		if ok {
			r.engine.surveilCards(r.game, r.agents, r.log, playerID, res.amount)
			res.succeeded = res.amount > 0
		}
	case game.EffectFight:
		resolveFight(r.game, r.obj, effect)
		res.succeeded = true
	case game.EffectReplace:
		res.succeeded = createReplacementEffect(r.game, r.obj, effect)
	case game.EffectChoose, game.EffectPay:
		res.succeeded = true
	case game.EffectApplyRule:
		res.succeeded = createRuleEffects(r.game, r.obj, effect)
	case game.EffectProliferate:
		res.succeeded = r.engine.resolveProliferate(r.game, r.obj, r.agents, r.log)
	case game.EffectGoad:
		if permanent, ok := r.permanent(effect); ok && permanentHasType(r.game, permanent, types.Creature) {
			goadPermanent(r.game, permanent, r.obj.Controller)
			res.succeeded = true
		}
	case game.EffectStartEngines:
		playerID, ok := r.player(effect)
		if ok {
			res.succeeded = startEngines(r.game, playerID)
		}
	case game.EffectSetClassLevel:
		permanent, ok := r.permanent(effect)
		if ok && res.amount > permanent.ClassLevel {
			permanent.ClassLevel = res.amount
			res.succeeded = true
		}
	case game.EffectMonstrosity:
		permanent, ok := r.permanent(effect)
		if ok && !permanent.Monstrous {
			if res.amount > 0 {
				permanent.Counters.Add(counter.PlusOnePlusOne, res.amount)
			}
			permanent.Monstrous = true
			res.succeeded = true
		}
	case game.EffectShufflePermanentIntoLibrary:
		permanent, ok := r.permanent(effect)
		if !ok {
			return res
		}

		owner := permanent.Owner
		if !movePermanentToZone(r.game, permanent, game.ZoneLibrary) {
			return res
		}
		if player, ok := playerByID(r.game, owner); ok {
			player.Library.Shuffle(r.engine.rng)
		}
		res.succeeded = true
	case game.EffectDiscover:
		res.succeeded = r.engine.resolveDiscover(r.game, r.obj, res.amount, r.agents, r.log)
	default:
	}
	return res
}

func (r *effectResolver) putLinkedCardOnBattlefield(effect *game.Effect) bool {
	key := linkedObjectSourceKey(r.game, r.obj, effect.LinkID)
	refs := linkedObjects(r.game, key)
	if len(refs) == 0 {
		return false
	}
	controller, ok := r.effectRecipientOrController(effect)
	if !ok {
		return false
	}
	for _, ref := range refs {
		if ref.CardID == 0 {
			continue
		}
		card, ok := r.game.GetCardInstance(ref.CardID)
		if !ok || !cardMatchesCondition(card.Def, effect.CardCondition) {
			continue
		}
		owner, ok := playerByID(r.game, card.Owner)
		if !ok || !owner.Library.Remove(card.ID) {
			continue
		}
		if _, ok := createCardPermanentWithChoices(r.engine, r.game, card, controller, game.ZoneLibrary, r.agents, r.log); ok {
			clearLinkedObjects(r.game, key)
			return true
		}
		owner.Library.Add(card.ID)
	}
	return false
}

func (r *effectResolver) putReferencedCardOnBattlefield(effect *game.Effect) bool {
	ref := effect.Card.Val
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, ref)
	if !ok || fromZone == game.ZoneNone {
		return false
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return false
	}
	controller, ok := r.effectRecipientOrController(effect)
	if !ok {
		return false
	}
	if ref.Kind == game.CardReferenceEvent {
		if owner, ok := playerByID(r.game, card.Owner); ok {
			controller = owner.ID
		}
	}
	if !removeCardFromZone(r.game, card.Owner, cardID, fromZone) {
		return false
	}
	permanent, ok := createCardPermanentFaceWithContinuous(r.engine, r.game, card, controller, fromZone, game.FaceFront, effect.ContinuousEffects, r.agents, r.log)
	if !ok {
		zone, zoneOK := destinationZone(r.game, card.Owner, fromZone)
		if zoneOK {
			zone.Add(cardID)
		}
		return false
	}
	return permanent != nil
}

func (r *effectResolver) tokenDefinition(effect *game.Effect) (*game.CardDef, bool) {
	if effect.TokenCopy.Exists {
		return buildTokenCopyDef(r.game, r.obj, effect.TokenCopy.Val)
	}
	if effect.Token.Exists {
		return effect.Token.Val, effect.Token.Val != nil
	}
	return nil, false
}

func resolveCardReference(g *game.Game, obj *game.StackObject, ref game.CardReference) (id.ID, game.ZoneType, bool) {
	switch ref.Kind {
	case game.CardReferenceSource:
		if obj == nil || obj.SourceCardID == 0 {
			return 0, game.ZoneNone, false
		}
		zone, ok := cardZone(g, obj.SourceCardID)
		return obj.SourceCardID, zone, ok
	case game.CardReferenceEvent:
		if obj == nil || !obj.HasTriggerEvent || obj.TriggerEvent.CardID == 0 {
			return 0, game.ZoneNone, false
		}
		zone, ok := cardZone(g, obj.TriggerEvent.CardID)
		return obj.TriggerEvent.CardID, zone, ok
	case game.CardReferenceLinked:
		for _, linked := range linkedObjects(g, linkedObjectSourceKey(g, obj, ref.LinkID)) {
			if linked.CardID == 0 {
				continue
			}
			if zone, ok := cardZone(g, linked.CardID); ok {
				return linked.CardID, zone, true
			}
		}
		return 0, game.ZoneNone, false
	default:
		return 0, game.ZoneNone, false
	}
}

func cardZone(g *game.Game, cardID id.ID) (game.ZoneType, bool) {
	for _, player := range g.Players {
		if player.Library.Contains(cardID) {
			return game.ZoneLibrary, true
		}
		if player.Hand.Contains(cardID) {
			return game.ZoneHand, true
		}
		if player.Graveyard.Contains(cardID) {
			return game.ZoneGraveyard, true
		}
		if player.Exile.Contains(cardID) {
			return game.ZoneExile, true
		}
		if player.CommandZone.Contains(cardID) {
			return game.ZoneCommand, true
		}
	}
	return game.ZoneNone, false
}

func buildTokenCopyDef(g *game.Game, obj *game.StackObject, spec game.TokenCopySpec) (*game.CardDef, bool) {
	var source *game.CardDef
	switch spec.Source {
	case game.TokenCopySourceSourceCard:
		cardID := stackObjectSourceID(obj)
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			return nil, false
		}
		source = cardFaceOrDefault(card, game.FaceFront)
	case game.TokenCopySourceObject:
		resolved, ok := resolveObjectReference(g, obj, spec.Object)
		if !ok {
			return nil, false
		}
		switch {
		case resolved.permanent != nil:
			var ok bool
			source, ok = permanentCardDef(g, resolved.permanent)
			if !ok {
				return nil, false
			}
		case resolved.snapshot.TokenDef != nil:
			source = resolved.snapshot.TokenDef
		case resolved.snapshot.CardID != 0:
			card, ok := g.GetCardInstance(resolved.snapshot.CardID)
			if !ok {
				return nil, false
			}
			source = cardFaceOrDefault(card, resolved.snapshot.Face)
		default:
		}
	default:
		return nil, false
	}
	token := copyCardDef(source)
	if spec.SetName != "" {
		token.Name = spec.SetName
	}
	if len(spec.SetColors) > 0 {
		token.Colors = append([]mana.Color(nil), spec.SetColors...)
	}
	if len(spec.SetTypes) > 0 {
		token.Types = append([]types.Card(nil), spec.SetTypes...)
	}
	if len(spec.SetSubtypes) > 0 {
		token.Subtypes = append([]types.Sub(nil), spec.SetSubtypes...)
	}
	if spec.SetPower.Exists {
		token.Power = spec.SetPower
		token.DynamicPower = opt.V[game.DynamicValue]{}
	}
	if spec.SetToughness.Exists {
		token.Toughness = spec.SetToughness
		token.DynamicToughness = opt.V[game.DynamicValue]{}
	}
	if spec.NoManaCost {
		token.ManaCost = opt.V[mana.Cost]{}
	}
	if spec.NoPrintedText {
		token.OracleText = ""
		token.Abilities = nil
	}
	return token, true
}

func copyCardDef(source *game.CardDef) *game.CardDef {
	copied := *source
	copied.Colors = append([]mana.Color(nil), source.Colors...)
	copied.ColorIdentity = source.ColorIdentity
	copied.Supertypes = append([]types.Super(nil), source.Supertypes...)
	copied.Types = append([]types.Card(nil), source.Types...)
	copied.Subtypes = append([]types.Sub(nil), source.Subtypes...)
	copied.EntersWithCounters = append([]game.CounterPlacement(nil), source.EntersWithCounters...)
	copied.Abilities = append([]game.AbilityDef(nil), source.Abilities...)
	if source.Back.Exists {
		copied.Back = opt.Val(copyCardFace(&source.Back.Val))
	}
	return &copied
}

func copyCardFace(source *game.CardFace) game.CardFace {
	copied := *source
	copied.Colors = append([]mana.Color(nil), source.Colors...)
	copied.Supertypes = append([]types.Super(nil), source.Supertypes...)
	copied.Types = append([]types.Card(nil), source.Types...)
	copied.Subtypes = append([]types.Sub(nil), source.Subtypes...)
	copied.EntersWithCounters = append([]game.CounterPlacement(nil), source.EntersWithCounters...)
	copied.Abilities = append([]game.AbilityDef(nil), source.Abilities...)
	return copied
}

// IsEffectTypeExecuted reports whether the generic rules resolver currently
// implements the given effect primitive.
func IsEffectTypeExecuted(effectType game.EffectType) bool {
	switch effectType {
	case game.EffectDraw,
		game.EffectGainLife,
		game.EffectLoseLife,
		game.EffectAddMana,
		game.EffectDamage,
		game.EffectDestroy,
		game.EffectCounter,
		game.EffectExile,
		game.EffectBounce,
		game.EffectSacrifice,
		game.EffectDiscard,
		game.EffectTap,
		game.EffectUntap,
		game.EffectModifyPT,
		game.EffectAddCounter,
		game.EffectRemoveCounter,
		game.EffectMoveCounters,
		game.EffectApplyContinuous,
		game.EffectCreateToken,
		game.EffectCreateDelayedTrigger,
		game.EffectPutOnBattlefield,
		game.EffectPrevent,
		game.EffectRegenerate,
		game.EffectSkipStep,
		game.EffectTransform,
		game.EffectPhaseOut,
		game.EffectCreateEmblem,
		game.EffectMill,
		game.EffectSearch,
		game.EffectReveal,
		game.EffectScry,
		game.EffectSurveil,
		game.EffectFight,
		game.EffectReplace,
		game.EffectChoose,
		game.EffectPay,
		game.EffectApplyRule,
		game.EffectProliferate,
		game.EffectGoad,
		game.EffectInvestigate,
		game.EffectShufflePermanentIntoLibrary,
		game.EffectDiscover,
		game.EffectStartEngines,
		game.EffectSetClassLevel,
		game.EffectMonstrosity:
		return true
	default:
		return false
	}
}

func (e *Engine) drawCards(g *game.Game, playerID game.PlayerID, amount int, log *TurnLog) bool {
	if amount <= 0 {
		return false
	}
	drew := false
	for range amount {
		cardID, ok := e.drawCard(g, playerID)
		drew = drew || ok
		log.addDraw(DrawLog{
			Player: playerID,
			CardID: cardID,
			Failed: !ok,
		})
	}
	return drew
}

func stackObjectSourceIsSnow(g *game.Game, obj *game.StackObject) bool {
	permanent, ok := permanentByObjectID(g, obj.SourceID)
	return ok && permanentIsSnow(g, permanent)
}

func permanentIsSnow(g *game.Game, permanent *game.Permanent) bool {
	return permanentHasSupertype(g, permanent, types.Snow)
}

func resolveMassPermanentEffect(g *game.Game, obj *game.StackObject, effect *game.Effect, amount int) bool {
	damageSource, sourceOK := resolveEffectDamageSource(g, obj, effect)
	if !sourceOK {
		return false
	}
	permanentIDs := selectedPermanentIDsForEffect(g, obj, effect, obj.Controller, damageSource.permanent)
	succeeded := false
	for _, permanentID := range permanentIDs {
		permanent, ok := permanentByObjectID(g, permanentID)
		if !ok {
			continue
		}
		switch effect.Type {
		case game.EffectDamage:
			if amount > 0 {
				dealt := dealPermanentDamage(g, damageSource.sourceID, damageSource.sourceObjectID, damageSource.controller, permanent, amount, false)
				applyDamageSourceKeywordEffects(g, damageSource.permanent, permanent, dealt)
				succeeded = dealt > 0 || succeeded
			}
		case game.EffectDestroy:
			_, ok := destroyPermanent(g, permanent.ObjectID)
			succeeded = succeeded || ok
		case game.EffectExile:
			succeeded = movePermanentToZone(g, permanent, game.ZoneExile) || succeeded
		case game.EffectBounce:
			succeeded = movePermanentToZone(g, permanent, game.ZoneHand) || succeeded
		case game.EffectTap:
			setPermanentTapped(g, permanent, true)
			succeeded = true
		case game.EffectUntap:
			setPermanentTapped(g, permanent, false)
			succeeded = true
		case game.EffectAddCounter:
			if amount > 0 {
				permanent.Counters.Add(effect.CounterKind, amount)
				succeeded = true
			}
		case game.EffectRemoveCounter:
			if amount > 0 {
				permanent.Counters.Remove(effect.CounterKind, amount)
				succeeded = true
			}
		case game.EffectApplyContinuous:
			succeeded = applyContinuousEffectTemplates(g, obj, permanent, effect) || succeeded
		default:
		}
	}
	return succeeded
}

func resolveMassPlayerEffect(g *game.Game, obj *game.StackObject, effect *game.Effect, amount int) bool {
	if effect.Type != game.EffectDamage || amount <= 0 {
		return false
	}
	damageSource, ok := resolveEffectDamageSource(g, obj, effect)
	if !ok {
		return false
	}
	succeeded := false
	for _, playerID := range selectedPlayerIDs(g, obj.Controller, effect.PlayerSelector) {
		dealt := dealPlayerDamage(g, damageSource.sourceID, damageSource.sourceObjectID, damageSource.controller, playerID, amount, false)
		if damageSource.permanent != nil {
			applyLifelink(g, damageSource.permanent, dealt)
		}
		succeeded = dealt > 0 || succeeded
	}
	return succeeded
}

func logUnsupportedEffect(log *TurnLog, obj *game.StackObject, effect *game.Effect) {
	log.addUnsupportedEffect(UnsupportedEffectLog{
		StackObjectID: stackObjectID(obj),
		SourceID:      stackObjectSourceID(obj),
		Controller:    stackObjectController(obj),
		EffectType:    effect.Type,
		Description:   effect.Description,
	})
}

func effectAmount(g *game.Game, obj *game.StackObject, effect *game.Effect) int {
	if !effect.DynamicAmount.Exists || effect.DynamicAmount.Val.Kind == game.DynamicAmountNone {
		return effect.Amount
	}
	return dynamicAmountValue(g, obj, stackObjectController(obj), effect.DynamicAmount.Val)
}

func dynamicAmountValue(g *game.Game, obj *game.StackObject, controller game.PlayerID, dynamic game.DynamicAmount) int {
	amount := 0
	switch dynamic.Kind {
	case game.DynamicAmountConstant:
		amount = dynamic.Constant
	case game.DynamicAmountX:
		if obj != nil {
			amount = obj.XValue
		}
	case game.DynamicAmountTargetPower:
		if obj == nil {
			break
		}
		if permanent, ok := effectPermanent(g, obj, &game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			amount = effectivePower(g, permanent)
		}
	case game.DynamicAmountTargetToughness:
		if obj == nil {
			break
		}
		if permanent, ok := effectPermanent(g, obj, &game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			if toughness, ok := effectiveToughness(g, permanent); ok {
				amount = toughness
			}
		}
	case game.DynamicAmountTargetManaValue:
		if obj == nil {
			break
		}
		if permanent, ok := effectPermanent(g, obj, &game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			if def, ok := permanentCardDef(g, permanent); ok {
				amount = def.ManaValue()
			}
		}
	case game.DynamicAmountTargetCounters:
		if obj == nil {
			break
		}
		if permanent, ok := effectPermanent(g, obj, &game.Effect{TargetIndex: dynamic.TargetIndex}); ok {
			amount = permanent.Counters.Get(dynamic.CounterKind)
		}
	case game.DynamicAmountControllerLife:
		if player, ok := playerByID(g, controller); ok {
			amount = player.Life
		}
	case game.DynamicAmountControllerHandSize:
		if player, ok := playerByID(g, controller); ok {
			amount = player.Hand.Size()
		}
	case game.DynamicAmountControllerGraveyardSize:
		if player, ok := playerByID(g, controller); ok {
			amount = player.Graveyard.Size()
		}
	case game.DynamicAmountCountSelector:
		amount = len(selectedPermanentIDs(g, controller, nil, dynamic.Selector))
	case game.DynamicAmountPreviousEffectResult:
		if obj != nil && dynamic.LinkID != "" {
			amount = obj.ResolvedAmounts[dynamic.LinkID]
		}
	case game.DynamicAmountPreviousEffectExcessDamage:
		if obj != nil && dynamic.LinkID != "" {
			amount = obj.ResolvedExcessDamage[dynamic.LinkID]
		}
	case game.DynamicAmountOpponentCount:
		amount = len(aliveOpponents(g, controller))
	case game.DynamicAmountEventDamage:
		if obj != nil && obj.HasTriggerEvent {
			amount = obj.TriggerEvent.Amount
		}
	case game.DynamicAmountObjectPower:
		if obj == nil {
			break
		}
		if resolved, ok := resolveObjectReference(g, obj, dynamic.Object); ok {
			amount = resolvedObjectPower(g, &resolved)
		}
	default:
	}
	multiplier := dynamic.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return amount * multiplier
}

func resolvedObjectPower(g *game.Game, resolved *resolvedObjectReference) int {
	if resolved.permanent != nil {
		return effectivePower(g, resolved.permanent)
	}
	if resolved.snapshot.Power.Exists {
		return resolved.snapshot.Power.Val
	}
	return 0
}

func rememberEffectAmount(obj *game.StackObject, effect *game.Effect, amount int) {
	if effect.LinkID == "" {
		return
	}
	if obj.ResolvedAmounts == nil {
		obj.ResolvedAmounts = make(map[string]int)
	}
	obj.ResolvedAmounts[effect.LinkID] = amount
}

func rememberEffectExcessDamage(obj *game.StackObject, effect *game.Effect, excessDamage int) {
	if effect.LinkID == "" || excessDamage <= 0 {
		return
	}
	if obj.ResolvedExcessDamage == nil {
		obj.ResolvedExcessDamage = make(map[string]int)
	}
	obj.ResolvedExcessDamage[effect.LinkID] = excessDamage
}

func damageResultAmount(effect *game.Effect, dealt, excess int) int {
	if effect.ResultAmount == game.EffectResultAmountExcessDamage {
		return excess
	}
	return dealt
}

func moveCounters(g *game.Game, obj *game.StackObject, effect *game.Effect) bool {
	destination, ok := effectPermanent(g, obj, effect)
	if !ok {
		return false
	}
	counters, source, ok := effectCounterSource(g, obj, effect.CounterSource)
	if !ok || counters.IsEmpty() {
		return false
	}
	if source != nil && source.ObjectID == destination.ObjectID {
		return false
	}
	for kind, amount := range counters.All() {
		destination.Counters.Add(kind, amount)
	}
	if source == nil {
		return true
	}
	for kind, amount := range counters.All() {
		source.Counters.Remove(kind, amount)
	}
	return true
}

func effectCounterSource(g *game.Game, obj *game.StackObject, source game.CounterSourceSpec) (counter.Set, *game.Permanent, bool) {
	switch source.Kind {
	case game.CounterSourceTarget:
		permanent, ok := effectPermanent(g, obj, &game.Effect{TargetIndex: source.TargetIndex})
		if !ok {
			return counter.Set{}, nil, false
		}
		return cloneCounters(permanent.Counters), permanent, true
	case game.CounterSourceEventPermanent:
		if !obj.HasTriggerEvent || obj.TriggerEvent.PermanentID == 0 {
			return counter.Set{}, nil, false
		}
		// Zone-change triggers such as "put those counters on..." use the
		// triggering permanent's current state or its last-known information if it
		// has already left the battlefield (CR 603.10, CR 122).
		if permanent, ok := permanentByObjectID(g, obj.TriggerEvent.PermanentID); ok {
			return cloneCounters(permanent.Counters), permanent, true
		}
		if snapshot, ok := lastKnownObject(g, obj.TriggerEvent.PermanentID); ok {
			return cloneCounters(snapshot.Counters), nil, true
		}
	default:
	}
	return counter.Set{}, nil, false
}

func effectConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.EffectCondition]) bool {
	if !condition.Exists {
		return true
	}
	cond := condition.Val
	if cond.PermanentType.Exists {
		permanent, ok := effectPermanent(g, obj, &game.Effect{TargetIndex: cond.TargetIndex})
		if !ok {
			return false
		}
		matches := permanentHasType(g, permanent, cond.PermanentType.Val)
		if cond.Negate {
			matches = !matches
		}
		if !matches {
			return false
		}
	}
	if !conditionSatisfied(g, conditionContext{
		controller: stackObjectController(obj),
		obj:        obj,
	}, cond.Condition) {
		return false
	}
	return true
}

func cardConditionSatisfied(g *game.Game, obj *game.StackObject, condition opt.V[game.CardCondition]) bool {
	if !condition.Exists {
		return true
	}
	cond := condition.Val
	if cond.Card.Kind != game.CardReferenceLinked || cond.Card.LinkID == "" {
		return false
	}
	for _, ref := range linkedObjects(g, linkedObjectSourceKey(g, obj, cond.Card.LinkID)) {
		if ref.CardID == 0 {
			continue
		}
		card, ok := g.GetCardInstance(ref.CardID)
		if ok && cardMatchesCondition(card.Def, condition) {
			return true
		}
	}
	return false
}

func cardMatchesCondition(card *game.CardDef, condition opt.V[game.CardCondition]) bool {
	if !condition.Exists {
		return true
	}
	if card == nil {
		return false
	}
	cond := condition.Val
	if cond.RequirePermanentCard && !card.IsPermanent() {
		return false
	}
	face := card.DefaultFace()
	for _, cardType := range cond.Types {
		if !face.HasType(cardType) {
			return false
		}
	}
	for _, supertype := range cond.Supertypes {
		if !face.HasSupertype(supertype) {
			return false
		}
	}
	if len(cond.SubtypesAny) > 0 && !slices.ContainsFunc(cond.SubtypesAny, face.HasSubtype) {
		return false
	}
	return true
}

func effectResultConditionSatisfied(obj *game.StackObject, condition opt.V[game.EffectResultCondition]) bool {
	if !condition.Exists || condition.Val.LinkID == "" {
		return true
	}
	if obj == nil || obj.ResolutionResults == nil {
		return false
	}
	cond := condition.Val
	result, ok := obj.ResolutionResults[cond.LinkID]
	if !ok {
		return false
	}
	if cond.Accepted != game.TriAny && (cond.Accepted == game.TriTrue) != result.Accepted {
		return false
	}
	if cond.Succeeded != game.TriAny && (cond.Succeeded == game.TriTrue) != result.Succeeded {
		return false
	}
	return true
}

func rememberEffectResolutionResult(obj *game.StackObject, effect *game.Effect, accepted, succeeded bool, amount int) {
	if obj == nil || effect.LinkID == "" {
		return
	}
	if obj.ResolutionResults == nil {
		obj.ResolutionResults = make(map[string]game.EffectResolutionResult)
	}
	obj.ResolutionResults[effect.LinkID] = game.EffectResolutionResult{
		Accepted:  accepted,
		Succeeded: succeeded,
		Amount:    amount,
	}
}

func applyContinuousEffectTemplates(g *game.Game, obj *game.StackObject, permanent *game.Permanent, effect *game.Effect) bool {
	if len(effect.ContinuousEffects) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	timestamp := game.Timestamp(g.IDGen.Next())
	applied := false
	for i := range effect.ContinuousEffects {
		template := &effect.ContinuousEffects[i]
		// Runtime continuous effects are applied by the layer system; animation
		// effects such as "becomes a 0/0 creature" use type and P/T layers
		// (CR 611, CR 613).
		runtimeEffect := *template
		runtimeEffect.ID = g.IDGen.Next()
		runtimeEffect.SourceCardID = sourceID
		runtimeEffect.SourceObjectID = sourceObjectID
		runtimeEffect.Controller = obj.Controller
		runtimeEffect.Timestamp = timestamp
		runtimeEffect.CreatedTurn = g.Turn.TurnNumber
		if effect.UntilEndOfTurn {
			runtimeEffect.Duration = game.DurationUntilEndOfTurn
		} else if effect.Duration != game.DurationPermanent {
			runtimeEffect.Duration = effect.Duration
		}
		if runtimeEffect.Duration == game.DurationUntilYourNextTurn && runtimeEffect.ExpiresFor == game.Player1 {
			runtimeEffect.ExpiresFor = obj.Controller
		}
		if runtimeEffect.AffectedObjectID == 0 && runtimeEffect.Selector == game.EffectSelectorNone {
			if permanent == nil {
				continue
			}
			runtimeEffect.AffectedObjectID = permanent.ObjectID
		}
		g.ContinuousEffects = append(g.ContinuousEffects, runtimeEffect)
		applied = true
	}
	return applied
}

func damageSourceIDs(g *game.Game, obj *game.StackObject) (sourceID, sourceObjectID id.ID) {
	switch obj.Kind {
	case game.StackActivatedAbility, game.StackTriggeredAbility:
		if obj.SourceCardID != 0 {
			if permanent, ok := permanentByObjectID(g, obj.SourceID); ok && permanent.CardInstanceID == obj.SourceCardID {
				return obj.SourceCardID, obj.SourceID
			}
			return obj.SourceCardID, 0
		}
		permanent, ok := permanentByObjectID(g, obj.SourceID)
		if !ok {
			return 0, obj.SourceID
		}
		return permanent.CardInstanceID, permanent.ObjectID
	default:
		return obj.SourceID, 0
	}
}

type effectDamageSource struct {
	sourceID       id.ID
	sourceObjectID id.ID
	controller     game.PlayerID
	permanent      *game.Permanent
}

func resolveEffectDamageSource(g *game.Game, obj *game.StackObject, effect *game.Effect) (effectDamageSource, bool) {
	if !effect.DamageSource.Exists {
		sourceID, sourceObjectID := damageSourceIDs(g, obj)
		return effectDamageSource{
			sourceID:       sourceID,
			sourceObjectID: sourceObjectID,
			controller:     obj.Controller,
		}, true
	}
	resolved, ok := resolveObjectReference(g, obj, effect.DamageSource.Val)
	if !ok {
		return effectDamageSource{}, false
	}
	if resolved.permanent == nil {
		if resolved.snapshot.ObjectID == 0 {
			return effectDamageSource{}, false
		}
		return effectDamageSource{
			sourceID:       resolved.snapshot.CardID,
			sourceObjectID: resolved.snapshot.ObjectID,
			controller:     resolved.snapshot.Controller,
		}, true
	}
	return effectDamageSource{
		sourceID:       resolved.permanent.CardInstanceID,
		sourceObjectID: resolved.permanent.ObjectID,
		controller:     effectiveController(g, resolved.permanent),
		permanent:      resolved.permanent,
	}, true
}

func applyDamageSourceKeywordEffects(g *game.Game, source, damaged *game.Permanent, damage int) {
	if source == nil || damage <= 0 {
		return
	}
	if hasKeyword(g, source, game.Deathtouch) {
		damaged.MarkedDeathtouchDamage = true
	}
	applyLifelink(g, source, damage)
}

func selectedPermanentIDs(g *game.Game, controller game.PlayerID, source *game.Permanent, selector game.EffectSelector) []id.ID {
	permanentIDs := make([]id.ID, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		if !permanentMatchesSelectorForSource(g, source, controller, permanent, selector) {
			continue
		}
		permanentIDs = append(permanentIDs, permanent.ObjectID)
	}
	return permanentIDs
}

func selectedPermanentIDsForEffect(g *game.Game, obj *game.StackObject, effect *game.Effect, controller game.PlayerID, source *game.Permanent) []id.ID {
	if effect.Selector == game.EffectSelectorOtherCreaturesDefendingPlayerControls {
		return selectedOtherCreaturesDefendingPlayerControls(g, obj)
	}
	if effect.Selector != game.EffectSelectorAllCreaturesExceptTarget {
		return selectedPermanentIDs(g, controller, source, effect.Selector)
	}
	excluded, _ := targetPermanentObjectID(obj, effect.TargetIndex)
	permanentIDs := make([]id.ID, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == excluded || !permanentHasType(g, permanent, types.Creature) {
			continue
		}
		permanentIDs = append(permanentIDs, permanent.ObjectID)
	}
	return permanentIDs
}

func selectedOtherCreaturesDefendingPlayerControls(g *game.Game, obj *game.StackObject) []id.ID {
	if obj == nil || !obj.HasTriggerEvent || obj.TriggerEvent.PermanentID == 0 {
		return []id.ID{}
	}
	resolved, ok := resolvePermanentOrLastKnown(g, obj.TriggerEvent.PermanentID)
	if !ok {
		return []id.ID{}
	}
	defendingPlayer, ok := resolved.controller(g)
	if !ok {
		return []id.ID{}
	}
	permanentIDs := make([]id.ID, 0, len(g.Battlefield))
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == obj.TriggerEvent.PermanentID {
			continue
		}
		if effectiveController(g, permanent) != defendingPlayer || !permanentHasType(g, permanent, types.Creature) {
			continue
		}
		permanentIDs = append(permanentIDs, permanent.ObjectID)
	}
	return permanentIDs
}

func selectedPlayerIDs(g *game.Game, controller game.PlayerID, selector game.PlayerSelector) []game.PlayerID {
	switch selector {
	case game.PlayerSelectorOpponents:
		return aliveOpponents(g, controller)
	default:
		return nil
	}
}

func permanentMatchesSelector(g *game.Game, permanent *game.Permanent, selector game.EffectSelector) bool {
	return permanentMatchesSelectorForSource(g, nil, 0, permanent, selector)
}

func permanentMatchesSelectorForSource(g *game.Game, source *game.Permanent, controller game.PlayerID, permanent *game.Permanent, selector game.EffectSelector) bool {
	switch selector {
	case game.EffectSelectorAllCreatures:
		return permanentHasType(g, permanent, types.Creature)
	case game.EffectSelectorAllArtifacts:
		return permanentHasType(g, permanent, types.Artifact)
	case game.EffectSelectorAllEnchantments:
		return permanentHasType(g, permanent, types.Enchantment)
	case game.EffectSelectorAllNonlandPermanents:
		return !permanentHasType(g, permanent, types.Land)
	case game.EffectSelectorAllPermanents:
		return true
	case game.EffectSelectorCreaturesYouControl:
		return effectiveController(g, permanent) == controller && permanentHasType(g, permanent, types.Creature)
	case game.EffectSelectorOtherCreaturesYouControl:
		return source != nil && permanent.ObjectID != source.ObjectID && effectiveController(g, permanent) == controller && permanentHasType(g, permanent, types.Creature)
	case game.EffectSelectorEquippedCreature:
		return source != nil && source.AttachedTo.Exists && permanent.ObjectID == source.AttachedTo.Val
	default:
		return false
	}
}

func effectPlayer(g *game.Game, obj *game.StackObject, effect *game.Effect) (game.PlayerID, bool) {
	if choice, ok := linkedResolutionChoice(obj, effect.ChoiceLinkID); ok && choice.Kind == game.ResolutionChoicePlayer {
		if !isPlayerAlive(g, choice.Player) {
			return 0, false
		}
		return choice.Player, true
	}
	if effect.TargetIndex == game.TargetIndexController {
		if !isPlayerAlive(g, obj.Controller) {
			return 0, false
		}
		return obj.Controller, true
	}
	if effect.TargetIndex < 0 || effect.TargetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[effect.TargetIndex]
	if target.Kind != game.TargetPlayer {
		return 0, false
	}
	if !isPlayerAlive(g, target.PlayerID) {
		return 0, false
	}
	return target.PlayerID, true
}

func effectManaColor(obj *game.StackObject, effect *game.Effect) mana.Color {
	if choice, ok := linkedResolutionChoice(obj, effect.ChoiceLinkID); ok && choice.Kind == game.ResolutionChoiceColor {
		return choice.Color
	}
	return effect.ManaColor
}

func effectPermanent(g *game.Game, obj *game.StackObject, effect *game.Effect) (*game.Permanent, bool) {
	if effect.Object.Exists {
		resolved, ok := resolveObjectReference(g, obj, effect.Object.Val)
		return resolved.permanent, ok && resolved.permanent != nil
	}
	if effect.TargetIndex == game.TargetIndexSourcePermanent {
		return sourcePermanent(g, obj)
	}
	if effect.TargetIndex < 0 || effect.TargetIndex >= len(obj.Targets) {
		return nil, false
	}
	target := obj.Targets[effect.TargetIndex]
	if target.Kind != game.TargetPermanent {
		return nil, false
	}
	return permanentByObjectID(g, target.PermanentID)
}

func sourcePermanent(g *game.Game, obj *game.StackObject) (*game.Permanent, bool) {
	return permanentByObjectID(g, obj.SourceID)
}

func firstPermanentControlledBy(g *game.Game, controller game.PlayerID) (*game.Permanent, bool) {
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) == controller {
			return permanent, true
		}
	}
	return nil, false
}

func permanentLinkedObjectRef(permanent *game.Permanent) game.LinkedObjectRef {
	if permanent.CardInstanceID == 0 {
		return game.LinkedObjectRef{}
	}
	return game.LinkedObjectRef{ObjectID: permanent.ObjectID, CardID: permanent.CardInstanceID}
}

func returnLinkedExiledObjects(e *Engine, g *game.Game, obj *game.StackObject, linkID string, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	key := linkedObjectSourceKey(g, obj, linkID)
	returned := false
	for _, ref := range linkedObjects(g, key) {
		if snapshot, ok := lastKnownObject(g, ref.ObjectID); !ok || snapshot.CardID != ref.CardID {
			continue
		}
		card, ok := g.GetCardInstance(ref.CardID)
		if !ok {
			continue
		}
		owner, ok := playerByID(g, card.Owner)
		if !ok || !owner.Exile.Remove(ref.CardID) {
			continue
		}
		if _, ok := createCardPermanentWithChoices(e, g, card, obj.Controller, game.ZoneExile, agents, log); ok {
			returned = true
		}
	}
	clearLinkedObjects(g, key)
	return returned
}

func createTokenPermanent(g *game.Game, controller game.PlayerID, token *game.CardDef) (*game.Permanent, bool) {
	if token == nil {
		return nil, false
	}
	objectID := g.IDGen.Next()
	permanent := &game.Permanent{
		ObjectID:      objectID,
		Owner:         controller,
		Controller:    controller,
		SummoningSick: entersSummoningSick(token),
		Token:         true,
		TokenDef:      token,
	}
	initializePermanentCounters(permanent, token)
	g.Battlefield = append(g.Battlefield, permanent)
	event := game.GameEvent{
		Controller:  controller,
		Player:      controller,
		PermanentID: objectID,
		TokenName:   token.Name,
		TokenDef:    token,
		FromZone:    game.ZoneNone,
		ToZone:      game.ZoneBattlefield,
	}
	emitZoneChangeEvent(g, event)
	event.Kind = game.EventPermanentEnteredBattlefield
	emitEvent(g, event)
	return permanent, true
}
