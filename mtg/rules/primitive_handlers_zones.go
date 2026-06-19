package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func handleDraw(r *effectResolver, prim game.Draw) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		return res
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			res.succeeded = r.engine.drawCards(r.game, playerID, res.amount, r.log) || res.succeeded
		}
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
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			res.succeeded = discardCards(r.game, playerID, res.amount) || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = discardCards(r.game, playerID, res.amount)
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
		var permanent *game.Permanent
		res.succeeded, permanent = r.engine.searchLibrary(r.game, r.obj, r.agents, r.log, playerID, prim.Spec, res.amount)
		if prim.PublishLinked != "" && permanent != nil {
			rememberLinkedObject(
				r.game,
				linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)),
				permanentLinkedObjectRef(permanent),
			)
		}
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
		res.succeeded = r.putReferencedCardOnBattlefieldValue(card, recipient, prim.ContinuousEffects, battlefieldEntryOptions(prim))
		return res
	}
	if key, ok := prim.Source.LinkedKey(); ok {
		res.succeeded = r.putLinkedCardOnBattlefieldValue(key, recipient, battlefieldEntryOptions(prim))
		if !res.succeeded {
			var controllerOverride opt.V[game.PlayerID]
			if prim.Recipient.Exists {
				if controller, ok := r.recipientController(recipient); ok {
					controllerOverride = opt.Val(controller)
				}
			}
			res.succeeded = returnLinkedExiledObjects(r.engine, r.game, r.obj, string(key), controllerOverride, battlefieldEntryOptions(prim), r.agents, r.log)
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
	token, ok := r.typedTokenDefinition(prim.Source)
	if !ok {
		return res
	}
	created, ok := createTokenPermanentsCollectingWithChoices(r.engine, r.game, recipient, token, res.amount, prim.EntryTapped, r.agents, r.log)
	if !ok {
		return res
	}
	if prim.EntryAttacking {
		declareCreatedTokensAttacking(r.engine, r.game, recipient, created, r.agents, r.log)
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

func handleDiscoverCards(r *effectResolver, prim game.DiscoverCards) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	res.succeeded = r.engine.resolveDiscover(r.game, r.obj, res.amount, r.agents, r.log)
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
	if prim.ControlledChoice {
		for _, permanent := range r.chooseControlledBouncePermanents(prim) {
			res.succeeded = movePermanentToZone(r.game, permanent, zone.Hand) || res.succeeded
		}
		return res
	}
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

// chooseControlledBouncePermanents has the resolving controller choose
// prim.Amount permanents from prim.Group's candidate pool (the permanents they
// control matching the bounce's selection), for "Return a creature you control
// to its owner's hand." style bounces. When the candidate pool holds no more
// permanents than the requested amount, every candidate is chosen without a
// prompt.
func (r *effectResolver) chooseControlledBouncePermanents(prim game.Bounce) []*game.Permanent {
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		return nil
	}
	candidates := r.groupPermanents(prim.Group)
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) <= amount {
		return candidates
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, permanent := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(r.game, permanent), Card: permanentChoiceInfo(r.game, permanent)}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           "Choose a permanent to return to its owner's hand",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}
	selected := r.engine.chooseChoice(r.game, r.agents, request, r.log)
	chosen := make([]*game.Permanent, 0, len(selected))
	for _, idx := range selected {
		if idx >= 0 && idx < len(candidates) {
			chosen = append(chosen, candidates[idx])
		}
	}
	return chosen
}

func handleMoveCard(r *effectResolver, prim game.MoveCard) effectResolved {
	if prim.Player.Kind() != game.PlayerReferenceNone {
		return handleMoveCardZoneGroup(r, prim)
	}
	res := effectResolved{accepted: true}
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
	if !ok || fromZone != prim.FromZone {
		return res
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return res
	}
	res.succeeded = moveCardBetweenZonesWithPlacement(r.game, card.Owner, cardID, fromZone, prim.Destination, prim.DestinationBottom)
	return res
}

// handleMoveCardZoneGroup resolves the player-zone group form of MoveCard,
// moving every card currently in the chosen player's source zone to the
// destination at once ("Exile target player's graveyard."). All cards share one
// SimultaneousID so the moves emit as a single zone-change batch. An empty
// source zone is a legal no-op.
func handleMoveCardZoneGroup(r *effectResolver, prim game.MoveCard) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	from, ok := destinationZone(r.game, playerID, prim.FromZone)
	if !ok {
		return res
	}
	cardIDs := from.All()
	if len(cardIDs) == 0 {
		return res
	}
	simultaneousID := r.game.IDGen.Next()
	for _, cardID := range cardIDs {
		card, ok := r.game.GetCardInstance(cardID)
		if !ok {
			continue
		}
		moved := moveCardBetweenZonesInBatch(r.game, card.Owner, cardID, prim.FromZone, prim.Destination, false, simultaneousID)
		res.succeeded = moved || res.succeeded
	}
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
	if !ok && prim.Object.Kind() == game.ObjectReferenceNone {
		permanent, ok = firstPermanentControlledBy(r.game, r.obj.Controller)
	}
	if !ok || effectiveController(r.game, permanent) != r.obj.Controller {
		return res
	}
	res.succeeded = sacrificePermanent(r.game, permanent)
	return res
}

func handleSacrificePermanents(r *effectResolver, prim game.SacrificePermanents) effectResolved {
	res := effectResolved{accepted: true}
	amount := r.quantity(prim.Amount)
	var players []game.PlayerID
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		players = playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup))
	} else if playerID, ok := r.resolvePlayer(prim.Player); ok {
		players = []game.PlayerID{playerID}
	}
	resolver := newReferenceResolver(r.game, r.obj)
	var chosen []*game.Permanent
	for _, playerID := range players {
		chosen = append(chosen, r.engine.chooseSacrificePermanentsForPlayer(r.game, resolver, playerID, amount, prim.Selection, r.agents, r.log)...)
	}
	res.succeeded = sacrificePermanentsSimultaneously(r.game, chosen)
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
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			millCards(r.game, playerID, res.amount)
		}
		res.succeeded = res.amount > 0
		return res
	}
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

func handleDig(r *effectResolver, prim game.Dig) effectResolved {
	look := r.quantity(prim.Look)
	res := effectResolved{accepted: true, amount: look}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = r.engine.digCards(r.game, r.agents, r.log, playerID, look, r.quantity(prim.Take), prim.Remainder)
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
	if !createTokenPermanentsWithChoices(r.engine, r.game, recipient, clueTokenDef(), res.amount, false, r.agents, r.log) {
		return res
	}
	res.succeeded = true
	return res
}

func handleManifest(r *effectResolver, prim game.Manifest) effectResolved {
	res := effectResolved{accepted: true}
	playerID := stackObjectController(r.obj)
	if prim.Dread {
		res.succeeded = r.engine.manifestDread(r.game, r.agents, r.log, playerID)
		return res
	}
	if r.engine.manifestTopCard(r.game, r.agents, r.log, playerID) {
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

func battlefieldEntryOptions(prim game.PutOnBattlefield) permanentCreationOptions {
	return permanentCreationOptions{
		ForceTapped: prim.EntryTapped,
		Counters:    prim.EntryCounters,
	}
}

func (r *effectResolver) putLinkedCardOnBattlefieldValue(linkedKey game.LinkedKey, recipientRef game.PlayerReference, options permanentCreationOptions) bool {
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
		if ref.ObjectID != 0 || ref.CardID == 0 {
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
		if _, ok := createCardPermanentFaceWithOptions(r.engine, r.game, card, controller, zone.Library, game.FaceFront, nil, options, r.agents, r.log); ok {
			clearLinkedObjects(r.game, key)
			return true
		}
		owner.Library.Add(card.ID)
	}
	return false
}

func (r *effectResolver) putReferencedCardOnBattlefieldValue(ref game.CardReference, recipientRef game.PlayerReference, continuousEffects []game.ContinuousEffect, options permanentCreationOptions) bool {
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
	permanent, ok := createCardPermanentFaceWithOptions(
		r.engine,
		r.game,
		card,
		controller,
		fromZone,
		game.FaceFront,
		continuousEffects,
		options,
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
