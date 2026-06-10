package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func (r *effectResolver) quantity(q game.Quantity) int {
	if q.IsDynamic() {
		return dynamicAmountValue(r.game, r.obj, stackObjectController(r.obj), q.DynamicAmount().Val)
	}
	return q.Value()
}

func (r *effectResolver) resolveObject(object game.ObjectReference) (*game.Permanent, bool) {
	resolved, ok := resolveObjectReference(r.game, r.obj, object)
	return resolved.permanent, ok && resolved.permanent != nil
}

func (r *effectResolver) resolvePlayer(player game.PlayerReference) (game.PlayerID, bool) {
	return resolvePlayerReference(r.game, r.obj, player)
}

func (r *effectResolver) recipientController(recipient game.PlayerReference) (game.PlayerID, bool) {
	if recipient.Kind() != game.PlayerReferenceNone {
		return r.resolvePlayer(recipient)
	}
	return r.obj.Controller, true
}

func (r *effectResolver) groupPermanents(group game.GroupReference) []*game.Permanent {
	ids := newReferenceResolver(r.game, r.obj).groupMembers(group)
	permanents := make([]*game.Permanent, 0, len(ids))
	for _, permanentID := range ids {
		if permanent, ok := permanentByObjectID(r.game, permanentID); ok {
			permanents = append(permanents, permanent)
		}
	}
	return permanents
}

func (r *effectResolver) groupPermanentsWithSource(group game.GroupReference, source *game.Permanent) []*game.Permanent {
	ids := newReferenceResolverWithSource(r.game, r.obj, source).groupMembers(group)
	permanents := make([]*game.Permanent, 0, len(ids))
	for _, permanentID := range ids {
		if permanent, ok := permanentByObjectID(r.game, permanentID); ok {
			permanents = append(permanents, permanent)
		}
	}
	return permanents
}

func (r *effectResolver) playerGroupMembers(group game.PlayerGroupReference) []game.PlayerID {
	return newReferenceResolver(r.game, r.obj).playerGroup(group)
}

func (r *effectResolver) damageSource(source game.ObjectReference) (effectDamageSource, bool) {
	if source.Kind() == game.ObjectReferenceNone {
		sourceID, sourceObjectID := damageSourceIDs(r.game, r.obj)
		return effectDamageSource{
			sourceID:       sourceID,
			sourceObjectID: sourceObjectID,
			controller:     r.obj.Controller,
		}, true
	}
	resolved, ok := resolveObjectReference(r.game, r.obj, source)
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
			deathtouch:     slices.Contains(resolved.snapshot.Keywords, game.Deathtouch),
			lifelink:       slices.Contains(resolved.snapshot.Keywords, game.Lifelink),
		}, true
	}
	return effectDamageSource{
		sourceID:       resolved.permanent.CardInstanceID,
		sourceObjectID: resolved.permanent.ObjectID,
		controller:     effectiveController(r.game, resolved.permanent),
		permanent:      resolved.permanent,
		deathtouch:     hasKeyword(r.game, resolved.permanent, game.Deathtouch),
		lifelink:       hasKeyword(r.game, resolved.permanent, game.Lifelink),
	}, true
}

func handleDamage(r *effectResolver, prim game.Damage) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	sourceRef := game.ObjectReference{}
	if prim.DamageSource.Exists {
		sourceRef = prim.DamageSource.Val
	}
	source, ok := r.damageSource(sourceRef)
	if !ok {
		return res
	}
	if object, ok := prim.Recipient.ObjectReference(); ok {
		return r.damageReferencedPermanent(res, source, prim.ResultAmountKind, object)
	}
	if player, ok := prim.Recipient.PlayerReference(); ok {
		return r.damageReferencedPlayer(res, source, prim.ResultAmountKind, player)
	}
	if player, ok := prim.Recipient.AnyTargetPlayerReference(); ok {
		if resolvedPlayer, playerOK := r.resolvePlayer(player); playerOK {
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, resolvedPlayer, res.amount, false)
			applyDamageSourceLifelink(r.game, source, dealt)
			res.amount = typedDamageResultAmount(prim.ResultAmountKind, dealt, 0)
			res.succeeded = dealt > 0
			return res
		}
	}
	if object, ok := prim.Recipient.AnyTargetObjectReference(); ok {
		return r.damageReferencedPermanent(res, source, prim.ResultAmountKind, object)
	}
	if group, ok := prim.Recipient.GroupReference(); ok {
		return r.damageSelectedPermanents(res, source, group)
	}
	if group, ok := prim.Recipient.PlayerGroupReference(); ok {
		for _, playerID := range r.playerGroupMembers(group) {
			dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, playerID, res.amount, false)
			applyDamageSourceLifelink(r.game, source, dealt)
			res.succeeded = dealt > 0 || res.succeeded
		}
	}
	return res
}

func (r *effectResolver) damageReferencedPlayer(res effectResolved, source effectDamageSource, resultKind game.EffectResultAmountKind, player game.PlayerReference) effectResolved {
	playerID, ok := r.resolvePlayer(player)
	if !ok {
		return res
	}
	dealt := dealPlayerDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, playerID, res.amount, false)
	applyDamageSourceLifelink(r.game, source, dealt)
	res.amount = typedDamageResultAmount(resultKind, dealt, 0)
	res.succeeded = dealt > 0
	return res
}

func (r *effectResolver) damageReferencedPermanent(res effectResolved, source effectDamageSource, resultKind game.EffectResultAmountKind, object game.ObjectReference) effectResolved {
	permanent, ok := r.resolveObject(object)
	if !ok {
		return res
	}
	lethalRemaining := lethalDamageRemaining(r.game, permanent)
	if source.deathtouch {
		lethalRemaining = 1
		if permanent.MarkedDeathtouchDamage {
			lethalRemaining = 0
		}
	} else if source.permanent != nil {
		lethalRemaining = lethalDamageRemainingFromSource(r.game, source.permanent, permanent)
	}
	dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
	applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
	res.excessDamage = max(0, dealt-lethalRemaining)
	res.amount = typedDamageResultAmount(resultKind, dealt, res.excessDamage)
	res.succeeded = dealt > 0 && (resultKind != game.EffectResultAmountExcessDamage || res.excessDamage > 0)
	return res
}

func (r *effectResolver) damageSelectedPermanents(res effectResolved, source effectDamageSource, group game.GroupReference) effectResolved {
	for _, permanent := range r.groupPermanentsWithSource(group, source.permanent) {
		dealt := dealPermanentDamage(r.game, source.sourceID, source.sourceObjectID, source.controller, permanent, res.amount, false)
		applyDamageSourceKeywordEffects(r.game, source, permanent, dealt)
		res.succeeded = dealt > 0 || res.succeeded
	}
	return res
}

func typedDamageResultAmount(kind game.EffectResultAmountKind, dealt, excess int) int {
	if kind == game.EffectResultAmountExcessDamage {
		return excess
	}
	return dealt
}

func handleDraw(r *effectResolver, prim game.Draw) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = r.engine.drawCards(r.game, playerID, res.amount, r.log)
	}
	return res
}

func handleDiscard(r *effectResolver, prim game.Discard) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = discardCards(r.game, playerID, res.amount)
	}
	return res
}

func handleDestroy(r *effectResolver, prim game.Destroy) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		permanents := r.groupPermanents(prim.Group)
		snapshots := make(map[id.ID]game.ObjectSnapshot, len(permanents))
		for _, permanent := range permanents {
			snapshots[permanent.ObjectID] = snapshotPermanent(r.game, permanent, zone.Battlefield)
		}
		for _, permanent := range permanents {
			_, destroyed := destroyPermanent(r.game, permanent.ObjectID)
			if _, remains := permanentByObjectID(r.game, permanent.ObjectID); !remains {
				snapshot := snapshots[permanent.ObjectID]
				rememberLastKnown(r.game, &snapshot)
			}
			res.succeeded = destroyed || res.succeeded
		}
		return res
	}
	permanent, ok := r.resolveObject(prim.Object)
	if ok {
		_, res.succeeded = destroyPermanent(r.game, permanent.ObjectID)
	}
	return res
}

func handleAddMana(r *effectResolver, prim game.AddMana) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		res.amount = 1
	}
	player, ok := playerByID(r.game, r.obj.Controller)
	if !ok || player.Eliminated {
		return res
	}
	manaColor := prim.ManaColor
	if choice, ok := linkedResolutionChoice(r.obj, string(prim.ChoiceFrom)); ok && choice.Kind == game.ResolutionChoiceMana {
		manaColor = choice.Color
	}
	if stackObjectSourceIsSnow(r.game, r.obj) {
		player.ManaPool.AddSnow(manaColor, res.amount)
	} else {
		player.ManaPool.Add(manaColor, res.amount)
	}
	res.succeeded = true
	return res
}

func handleAddCounter(r *effectResolver, prim game.AddCounter) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && res.amount > 0 {
		addCountersToPermanent(r.game, permanent, prim.CounterKind, res.amount)
		res.succeeded = true
	}
	return res
}

func handleMoveCounters(r *effectResolver, prim game.MoveCounters) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	destination, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	counters, source, ok := effectCounterSource(r.game, r.obj, prim.Source)
	if !ok || counters.IsEmpty() || source != nil && source.ObjectID == destination.ObjectID {
		return res
	}
	for kind, amount := range counters.All() {
		addCountersToPermanent(r.game, destination, kind, amount)
		if source != nil {
			source.Counters.Remove(kind, amount)
		}
	}
	res.succeeded = true
	return res
}

func handleApplyContinuous(r *effectResolver, prim game.ApplyContinuous) effectResolved {
	res := effectResolved{accepted: true}
	var permanent *game.Permanent
	if prim.Object.Exists {
		permanent, _ = r.resolveObject(prim.Object.Val)
	}
	res.succeeded = applyTypedContinuousEffects(r.game, r.obj, permanent, prim.ContinuousEffects, prim.Duration)
	return res
}

func handleApplyRule(r *effectResolver, prim game.ApplyRule) effectResolved {
	return effectResolved{
		accepted:  true,
		succeeded: createRuleEffectTemplates(r.game, r.obj, prim.Object, prim.RuleEffects, prim.Duration),
	}
}

func handleModifyPT(r *effectResolver, prim game.ModifyPT) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok || prim.Duration != game.DurationUntilEndOfTurn {
		return res
	}
	powerDelta := r.quantity(prim.PowerDelta)
	toughnessDelta := r.quantity(prim.ToughnessDelta)
	r.game.ContinuousEffects = append(r.game.ContinuousEffects, untilEndOfTurnPTContinuousEffect(r.game, r.obj, permanent, powerDelta, toughnessDelta))
	res.succeeded = true
	return res
}

func handleFight(r *effectResolver, prim game.Fight) effectResolved {
	first, firstOK := r.resolveObject(prim.Object)
	second, secondOK := r.resolveObject(prim.RelatedObject)
	if !firstOK || !secondOK || first.ObjectID == second.ObjectID ||
		!permanentHasType(r.game, first, types.Creature) || !permanentHasType(r.game, second, types.Creature) {
		return effectResolved{accepted: true}
	}
	resolveFightPermanents(r.game, first, second)
	return effectResolved{accepted: true, succeeded: true}
}

func handleTap(r *effectResolver, prim game.Tap) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		setPermanentTapped(r.game, permanent, true)
		res.succeeded = true
	}
	return res
}

func handleSearch(r *effectResolver, prim game.Search) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if !searchSpecSupported(prim.Spec) {
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = r.engine.searchLibrary(r.game, r.obj, playerID, prim.Spec, res.amount)
	}
	return res
}

func handleReveal(r *effectResolver, prim game.Reveal) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerRef := prim.Player
	if prim.Recipient.Exists {
		playerRef = prim.Recipient.Val
	}
	playerID, ok := r.resolvePlayer(playerRef)
	if !ok {
		return res
	}
	revealed := revealCardIDs(r.game, r.obj, playerID, zone.Library, res.amount)
	if prim.PublishLinked != "" {
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		for _, cardID := range revealed {
			rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
		}
	}
	res.succeeded = len(revealed) > 0
	return res
}

func handlePutOnBattlefield(r *effectResolver, prim game.PutOnBattlefield) effectResolved {
	res := effectResolved{accepted: true}
	var recipient game.PlayerReference
	if prim.Recipient.Exists {
		recipient = prim.Recipient.Val
	}
	if card, ok := prim.Source.CardRef(); ok {
		res.succeeded = r.putReferencedCardOnBattlefieldValue(card, recipient, prim.ContinuousEffects)
		return res
	}
	if key, ok := prim.Source.LinkedKey(); ok {
		res.succeeded = r.putLinkedCardOnBattlefieldValue(key, recipient)
		if !res.succeeded {
			res.succeeded = returnLinkedExiledObjects(r.engine, r.game, r.obj, string(key), r.agents, r.log)
		}
	}
	return res
}

func handleCreateToken(r *effectResolver, prim game.CreateToken) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		res.amount = 1
	}
	var recipientRef game.PlayerReference
	if prim.Recipient.Exists {
		recipientRef = prim.Recipient.Val
	}
	recipient, ok := r.recipientController(recipientRef)
	if !ok {
		return res
	}
	for range res.amount {
		token, ok := r.typedTokenDefinition(prim.Source)
		if !ok {
			return res
		}
		if _, ok := createTokenPermanent(r.game, recipient, token); !ok {
			return res
		}
	}
	res.succeeded = res.amount > 0
	return res
}

func handleShufflePermanentIntoLibrary(r *effectResolver, prim game.ShufflePermanentIntoLibrary) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	owner := permanent.Owner
	if !movePermanentToZone(r.game, permanent, zone.Library) {
		return res
	}
	if player, ok := playerByID(r.game, owner); ok {
		player.Library.Shuffle(r.engine.rng)
	}
	res.succeeded = true
	return res
}

func handleStartEngines(r *effectResolver, prim game.StartEngines) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = startEngines(r.game, playerID)
	}
	return res
}

func handleSetClassLevel(r *effectResolver, prim game.SetClassLevel) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && res.amount > permanent.ClassLevel {
		permanent.ClassLevel = res.amount
		res.succeeded = true
	}
	return res
}

func handleMonstrosity(r *effectResolver, prim game.Monstrosity) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	permanent, ok := r.resolveObject(prim.Object)
	if ok && !permanent.Monstrous {
		if res.amount > 0 {
			permanent.Counters.Add(counter.PlusOnePlusOne, res.amount)
		}
		permanent.Monstrous = true
		res.succeeded = true
	}
	return res
}

func handleDiscoverCards(r *effectResolver, prim game.DiscoverCards) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	res.succeeded = r.engine.resolveDiscover(r.game, r.obj, res.amount, r.agents, r.log)
	return res
}

func handlePay(r *effectResolver, prim game.Pay) effectResolved {
	payment := prim.Payment
	if payment.Prompt == "" {
		payment.Prompt = prim.Prompt
	}
	accepted, succeeded := r.engine.resolveResolutionPaymentValue(r.game, r.obj, &payment, r.agents, r.log)
	return effectResolved{accepted: accepted, succeeded: succeeded}
}

func handleChoose(r *effectResolver, prim game.Choose) effectResolved {
	succeeded := r.engine.resolveResolutionChoiceValue(r.game, r.obj, &prim.Choice, string(prim.PublishChoice), r.agents, r.log)
	return effectResolved{accepted: true, succeeded: succeeded}
}

func handleGainLife(r *effectResolver, prim game.GainLife) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = gainLife(r.game, playerID, res.amount) > 0
	}
	return res
}

func handleLoseLife(r *effectResolver, prim game.LoseLife) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = loseLife(r.game, playerID, res.amount) > 0
	}
	return res
}

func handleExile(r *effectResolver, prim game.Exile) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			res.succeeded = movePermanentToZone(r.game, permanent, zone.Exile) || res.succeeded
		}
		return res
	}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	linkedObjectRef := permanentLinkedObjectRef(permanent)
	res.succeeded = movePermanentToZone(r.game, permanent, zone.Exile)
	if prim.ExileLinkedKey != "" {
		rememberLinkedObject(r.game, linkedObjectSourceKey(r.game, r.obj, string(prim.ExileLinkedKey)), linkedObjectRef)
	}
	return res
}

func handleBounce(r *effectResolver, prim game.Bounce) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			res.succeeded = movePermanentToZone(r.game, permanent, zone.Hand) || res.succeeded
		}
		return res
	}
	permanent, ok := r.resolveObject(prim.Object)
	if ok {
		res.succeeded = movePermanentToZone(r.game, permanent, zone.Hand)
	}
	return res
}

func handleMoveCard(r *effectResolver, prim game.MoveCard) effectResolved {
	res := effectResolved{accepted: true}
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
	if !ok || fromZone != prim.FromZone {
		return res
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return res
	}
	res.succeeded = moveCardBetweenZones(r.game, card.Owner, cardID, fromZone, prim.Destination)
	return res
}

func handleGrantCastPermission(r *effectResolver, prim game.GrantCastPermission) effectResolved {
	res := effectResolved{accepted: true}
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
	if !ok || fromZone != prim.FromZone {
		return res
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return res
	}
	if _, ok := cardFaceDef(card, prim.Face); !ok {
		return res
	}
	r.game.RuleEffects = append(r.game.RuleEffects, game.RuleEffect{
		ID:             r.game.IDGen.Next(),
		Kind:           game.RuleEffectCastFromZone,
		Controller:     r.obj.Controller,
		SourceCardID:   r.obj.SourceCardID,
		SourceObjectID: r.obj.SourceID,
		AffectedPlayer: game.PlayerYou,
		Duration:       prim.Duration,
		CreatedTurn:    r.game.Turn.TurnNumber,
		CastFromZone:   prim.FromZone,
		AffectedCardID: cardID,
		CastFace:       opt.Val(prim.Face),
		ExpiresFor:     r.obj.Controller,
	})
	res.succeeded = true
	return res
}

func handleSacrifice(r *effectResolver, prim game.Sacrifice) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok {
		permanent, ok = firstPermanentControlledBy(r.game, r.obj.Controller)
	}
	if !ok || effectiveController(r.game, permanent) != r.obj.Controller {
		return res
	}
	res.succeeded = movePermanentToZone(r.game, permanent, zone.Graveyard)
	return res
}

func handleUntap(r *effectResolver, prim game.Untap) effectResolved {
	res := effectResolved{accepted: true}
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			setPermanentTapped(r.game, permanent, false)
			res.succeeded = true
		}
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		setPermanentTapped(r.game, permanent, false)
		res.succeeded = true
	}
	return res
}

func handleCounterObject(r *effectResolver, prim game.CounterObject) effectResolved {
	if prim.Object.Kind() != game.ObjectReferenceTargetStackObject {
		return effectResolved{accepted: true}
	}
	return effectResolved{accepted: true, succeeded: counterTargetStackObject(r.game, r.obj, prim.Object.TargetIndex())}
}

func handleMill(r *effectResolver, prim game.Mill) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		millCards(r.game, playerID, res.amount)
		res.succeeded = res.amount > 0
	}
	return res
}

func handleScry(r *effectResolver, prim game.Scry) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		r.engine.scryCards(r.game, r.agents, r.log, playerID, res.amount)
		res.succeeded = res.amount > 0
	}
	return res
}

func handleSurveil(r *effectResolver, prim game.Surveil) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		r.engine.surveilCards(r.game, r.agents, r.log, playerID, res.amount)
		res.succeeded = res.amount > 0
	}
	return res
}

func handleInvestigate(r *effectResolver, prim game.Investigate) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		res.amount = 1
	}
	var recipientRef game.PlayerReference
	if prim.Recipient.Exists {
		recipientRef = prim.Recipient.Val
	}
	recipient, ok := r.recipientController(recipientRef)
	if !ok {
		return res
	}
	for range res.amount {
		if _, ok := createTokenPermanent(r.game, recipient, clueTokenDef()); !ok {
			return res
		}
	}
	res.succeeded = true
	return res
}

func handleProliferate(r *effectResolver, _ game.Proliferate) effectResolved {
	return effectResolved{accepted: true, succeeded: r.engine.resolveProliferate(r.game, r.obj, r.agents, r.log)}
}

func handleGoad(r *effectResolver, prim game.Goad) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok && permanentHasType(r.game, permanent, types.Creature) {
		goadPermanent(r.game, permanent, r.obj.Controller)
		res.succeeded = true
	}
	return res
}

func handleRemoveCounter(r *effectResolver, prim game.RemoveCounter) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	if prim.Group.Valid() {
		for _, permanent := range r.groupPermanents(prim.Group) {
			permanent.Counters.Remove(prim.CounterKind, res.amount)
			res.succeeded = true
		}
		return res
	}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		permanent.Counters.Remove(prim.CounterKind, res.amount)
		res.succeeded = true
	}
	return res
}

func handleTransform(r *effectResolver, prim game.Transform) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		res.succeeded = transformPermanent(r.game, permanent)
	}
	return res
}

func handlePhaseOut(r *effectResolver, prim game.PhaseOut) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		permanent.PhasedOut = true
		removePermanentFromCombat(r.game, permanent.ObjectID)
		res.succeeded = true
	}
	return res
}

func handleRegenerate(r *effectResolver, prim game.Regenerate) effectResolved {
	res := effectResolved{accepted: true}
	if permanent, ok := r.resolveObject(prim.Object); ok {
		permanent.RegenerationShields++
		res.succeeded = true
	}
	return res
}

func handleSkipStep(r *effectResolver, prim game.SkipStep) effectResolved {
	res := effectResolved{accepted: true}
	if playerID, ok := r.resolvePlayer(prim.Player); ok {
		scheduleSkipStep(r.game, playerID, prim.Step)
		res.succeeded = true
	}
	return res
}

func handleCreateEmblem(r *effectResolver, prim game.CreateEmblem) effectResolved {
	r.game.Emblems = append(r.game.Emblems, game.Emblem{
		Owner:     r.obj.Controller,
		Abilities: append([]game.Ability(nil), prim.EmblemAbilities...),
	})
	return effectResolved{accepted: true, succeeded: true}
}

func handleCreateDelayedTrigger(r *effectResolver, prim game.CreateDelayedTrigger) effectResolved {
	return effectResolved{accepted: true, succeeded: scheduleDelayedTrigger(r.game, r.obj, &prim.Trigger)}
}

func handleCreateReplacement(r *effectResolver, prim game.CreateReplacement) effectResolved {
	replacement := *prim.Replacement
	replacement.ID = r.game.IDGen.Next()
	replacement.Controller = r.obj.Controller
	replacement.SourceCardID, replacement.SourceObjectID = damageSourceIDs(r.game, r.obj)
	replacement.CreatedTurn = r.game.Turn.TurnNumber
	if prim.Duration != game.DurationPermanent {
		replacement.Duration = prim.Duration
	}
	r.game.ReplacementEffects = append(r.game.ReplacementEffects, replacement)
	return effectResolved{accepted: true, succeeded: true}
}

func handlePreventDamage(r *effectResolver, prim game.PreventDamage) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	res.succeeded = createPreventionShield(r.game, r.obj, res.amount, prim.Object, prim.Player, game.DurationUntilEndOfTurn)
	return res
}

func applyTypedContinuousEffects(g *game.Game, obj *game.StackObject, permanent *game.Permanent, templates []game.ContinuousEffect, duration game.EffectDuration) bool {
	if len(templates) == 0 {
		return false
	}
	sourceID, sourceObjectID := damageSourceIDs(g, obj)
	timestamp := game.Timestamp(g.IDGen.Next())
	applied := false
	for i := range templates {
		runtimeEffect := templates[i]
		runtimeEffect.ID = g.IDGen.Next()
		runtimeEffect.SourceCardID = sourceID
		runtimeEffect.SourceObjectID = sourceObjectID
		runtimeEffect.Controller = obj.Controller
		runtimeEffect.Timestamp = timestamp
		runtimeEffect.CreatedTurn = g.Turn.TurnNumber
		if duration != game.DurationPermanent {
			runtimeEffect.Duration = duration
		}
		if runtimeEffect.Duration == game.DurationUntilYourNextTurn && runtimeEffect.ExpiresFor == game.Player1 {
			runtimeEffect.ExpiresFor = obj.Controller
		}
		if runtimeEffect.AffectedObjectID == 0 && !runtimeEffect.Group.Valid() {
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

func (r *effectResolver) putLinkedCardOnBattlefieldValue(linkedKey game.LinkedKey, recipientRef game.PlayerReference) bool {
	key := linkedObjectSourceKey(r.game, r.obj, string(linkedKey))
	refs := linkedObjects(r.game, key)
	if len(refs) == 0 {
		return false
	}
	controller, ok := r.recipientController(recipientRef)
	if !ok {
		return false
	}
	cardCondition := r.currentInstruction.CardCondition
	for _, ref := range refs {
		if ref.CardID == 0 {
			continue
		}
		card, ok := r.game.GetCardInstance(ref.CardID)
		if !ok || !cardMatchesCondition(card.Def, cardCondition) {
			continue
		}
		owner, ok := playerByID(r.game, card.Owner)
		if !ok || !owner.Library.Remove(card.ID) {
			continue
		}
		if _, ok := createCardPermanentWithChoices(r.engine, r.game, card, controller, zone.Library, r.agents, r.log); ok {
			clearLinkedObjects(r.game, key)
			return true
		}
		owner.Library.Add(card.ID)
	}
	return false
}

func (r *effectResolver) putReferencedCardOnBattlefieldValue(ref game.CardReference, recipientRef game.PlayerReference, continuousEffects []game.ContinuousEffect) bool {
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, ref)
	if !ok || fromZone == zone.None {
		return false
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return false
	}
	controller, ok := r.recipientController(recipientRef)
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
	permanent, ok := createCardPermanentFaceWithContinuous(
		r.engine,
		r.game,
		card,
		controller,
		fromZone,
		game.FaceFront,
		continuousEffects,
		r.agents,
		r.log,
	)
	if !ok {
		destinationCards, zoneOK := destinationZone(r.game, card.Owner, fromZone)
		if zoneOK {
			destinationCards.Add(cardID)
		}
		return false
	}
	return permanent != nil
}

func (r *effectResolver) typedTokenDefinition(source game.TokenSource) (*game.CardDef, bool) {
	if spec, ok := source.TokenCopy(); ok {
		return buildTokenCopyDef(r.game, r.obj, spec)
	}
	if def, ok := source.TokenDefRef(); ok {
		return def, def != nil
	}
	return nil, false
}
