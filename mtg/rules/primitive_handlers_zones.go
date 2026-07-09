package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
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
			res.succeeded = r.engine.drawCards(r.game, playerID, res.amount, r.agents, r.log) || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = r.engine.drawCards(r.game, playerID, res.amount, r.agents, r.log)
	}
	return res
}

func handleReorderLibraryTop(r *effectResolver, prim game.ReorderLibraryTop) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	cards := peekLibrary(player, r.quantity(prim.Amount))
	res.amount = len(cards)
	if len(cards) == 0 {
		return res
	}
	options := make([]game.ChoiceOption, len(cards))
	defaultOrder := make([]int, len(cards))
	for i, cardID := range cards {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(r.game, cardID),
			Card:  cardChoiceInfo(r.game, cardID),
		}
		defaultOrder[i] = i
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceOrder,
		Player:           playerID,
		Prompt:           "Put the looked-at cards back in order, top card first",
		Options:          options,
		MinChoices:       len(cards),
		MaxChoices:       len(cards),
		DefaultSelection: defaultOrder,
	}, r.log)
	ordered := make([]id.ID, len(selected))
	for i, index := range selected {
		ordered[i] = cards[index]
	}
	reorderLibraryTop(player, ordered)
	res.succeeded = true
	return res
}

func handleLookAtLibraryTop(r *effectResolver, prim game.LookAtLibraryTop) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
	clearLinkedObjects(r.game, key)
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	cards := peekLibrary(player, 1)
	if len(cards) == 0 {
		return res
	}
	cardID := cards[0]
	rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
	r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:       game.ChoiceResolution,
		Player:     playerID,
		Prompt:     "Look at the top card of your library",
		Subject:    cardChoiceInfo(r.game, cardID),
		MinChoices: 0,
		MaxChoices: 0,
	}, r.log)
	res.amount = 1
	res.succeeded = true
	return res
}

func handleShuffleLibrary(r *effectResolver, prim game.ShuffleLibrary) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	player.Library.Shuffle(r.engine.rng)
	res.succeeded = true
	return res
}

// handleShuffleGraveyardIntoLibrary moves every card in the referenced player's
// graveyard into that player's library, then shuffles the library. Per the
// shuffle rules the library is shuffled even when no cards moved.
func handleShuffleGraveyardIntoLibrary(r *effectResolver, prim game.ShuffleGraveyardIntoLibrary) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	for _, cardID := range player.Graveyard.All() {
		moveCardBetweenZones(r.game, playerID, cardID, zone.Graveyard, zone.Library)
	}
	player.Library.Shuffle(r.engine.rng)
	res.succeeded = true
	return res
}

// handleLookAtHand resolves a "look at target player's hand" effect. Looking at
// a hand reveals hidden information to the source's controller but does not
// change game state, so the handler resolves the player and succeeds.
func handleLookAtHand(r *effectResolver, prim game.LookAtHand) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	if _, ok := playerByID(r.game, playerID); !ok {
		return res
	}
	res.succeeded = true
	return res
}

// handleChooseDiscardFromHand resolves the targeted hand-disruption family
// (Coercion / Duress / Thoughtseize / Inquisition of Kozilek): the resolving
// controller looks at the referenced player's hand, chooses one card matching
// the filter, and that player discards it. The hand is revealed even when no
// card matches, so the effect succeeds whenever the player resolves.
func handleChooseDiscardFromHand(r *effectResolver, prim game.ChooseDiscardFromHand) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	res.succeeded = true
	var candidates []id.ID
	for _, cardID := range player.Hand.All() {
		card, ok := r.game.GetCardInstance(cardID)
		if !ok {
			continue
		}
		if prim.ExcludeCreature && card.Def.HasType(types.Creature) {
			continue
		}
		if prim.ExcludeLand && card.Def.HasType(types.Land) {
			continue
		}
		if prim.MaxManaValue.Exists && card.Def.ManaValue() > prim.MaxManaValue.Val {
			continue
		}
		if !handCardMatchesSelection(r.game, card, prim.Selection, playerID) {
			continue
		}
		candidates = append(candidates, cardID)
	}
	if len(candidates) == 0 {
		return res
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, cardID := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(r.game, cardID),
			Card:  cardChoiceInfo(r.game, cardID),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           "Choose a card for that player to discard",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}, r.log)
	simultaneousID := r.game.IDGen.Next()
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		discardCardFromHandInBatch(r.game, playerID, candidates[idx], simultaneousID)
	}
	return res
}

func handleDiscard(r *effectResolver, prim game.Discard) effectResolved {
	if prim.EntireHand {
		return handleDiscardEntireHand(r, prim)
	}
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			if prim.AtRandom {
				res.succeeded = len(discardCardsAtRandomFromHand(r.game, playerID, res.amount)) > 0 || res.succeeded
			} else {
				res.succeeded = discardCards(r.game, playerID, res.amount) || res.succeeded
			}
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		var publishKey game.LinkedObjectKey
		if prim.PublishLinked != "" {
			publishKey = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
			clearLinkedObjects(r.game, publishKey)
		}
		if prim.AtRandom {
			res.succeeded = r.discardCardsAtRandom(playerID, res.amount, publishKey)
		} else {
			res.succeeded = r.discardCardsWithChoices(playerID, res.amount, publishKey)
		}
	}
	return res
}

// handleDiscardEntireHand resolves a "discard their hand" effect: each affected
// player discards every card in hand. res.amount carries the count discarded by
// a single player, or the greatest count across a player group.
func handleDiscardEntireHand(r *effectResolver, prim game.Discard) effectResolved {
	res := effectResolved{accepted: true}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			discarded := discardEntireHand(r.game, playerID)
			if discarded > res.amount {
				res.amount = discarded
			}
			res.succeeded = discarded > 0 || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.amount = discardEntireHand(r.game, playerID)
		res.succeeded = res.amount > 0
	}
	return res
}

func (r *effectResolver) discardCardsWithChoices(playerID game.PlayerID, amount int, publishKey game.LinkedObjectKey) bool {
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return false
	}
	candidates := player.Hand.All()
	amount = min(amount, len(candidates))
	if amount <= 0 {
		return false
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, cardID := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(r.game, cardID),
			Card:  cardChoiceInfo(r.game, cardID),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose cards to discard",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}, r.log)
	simultaneousID := r.game.IDGen.Next()
	discarded := false
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		if discardCardFromHandInBatch(r.game, playerID, candidates[idx], simultaneousID) {
			discarded = true
			if publishKey != (game.LinkedObjectKey{}) {
				rememberLinkedObject(r.game, publishKey, game.LinkedObjectRef{CardID: candidates[idx]})
			}
		}
	}
	return discarded
}

// discardCardsAtRandom discards up to amount cards chosen uniformly at random
// from the player's hand, as one simultaneous batch ("Discard a card at
// random."). It returns whether any card was discarded.
func (r *effectResolver) discardCardsAtRandom(playerID game.PlayerID, amount int, publishKey game.LinkedObjectKey) bool {
	discarded := discardCardsAtRandomFromHand(r.game, playerID, amount)
	for _, cardID := range discarded {
		if publishKey != (game.LinkedObjectKey{}) {
			rememberLinkedObject(r.game, publishKey, game.LinkedObjectRef{CardID: cardID})
		}
	}
	return len(discarded) > 0
}

func handleSearch(r *effectResolver, prim game.Search) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if !searchSpecSupported(prim.Spec) {
		return res
	}
	if prim.Spec.RevealOnly {
		return handleSearchRevealOnly(r, prim)
	}
	if prim.Spec.AlsoGraveyard {
		playerID, ok := r.resolvePlayer(prim.Player)
		if !ok {
			return res
		}
		res.succeeded = r.engine.searchLibraryAndGraveyard(r.game, r.obj, r.agents, r.log, playerID, prim.Spec)
		return res
	}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		// "Each player searches their library ..." — every member searches their
		// own library and any found permanent enters under that searcher's
		// control (no Controller rider applies to a group search).
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			succeeded, _ := r.engine.searchLibrary(r.game, r.obj, r.agents, r.log, playerID, playerID, prim.Spec, res.amount)
			res.succeeded = succeeded || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	controllerID := playerID
	if prim.Controller.Exists {
		// "under target player's control" routes the found permanent to a named
		// player; an unresolvable controller leaves it under the searcher's.
		if resolved, controllerOK := r.resolvePlayer(prim.Controller.Val); controllerOK {
			controllerID = resolved
		}
	}
	var key game.LinkedObjectKey
	if prim.PublishLinked != "" {
		key = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
	}
	if ok {
		var permanent *game.Permanent
		res.succeeded, permanent = r.engine.searchLibrary(r.game, r.obj, r.agents, r.log, playerID, controllerID, prim.Spec, res.amount)
		if prim.PublishLinked != "" && permanent != nil {
			rememberLinkedObject(r.game, key, permanentLinkedObjectRef(permanent))
		}
	}
	return res
}

func handleSearchRevealOnly(r *effectResolver, prim game.Search) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	var key game.LinkedObjectKey
	if prim.PublishLinked != "" {
		key = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
	}
	cardID, found := r.engine.searchLibraryRevealOnly(r.game, r.obj, r.agents, r.log, playerID, prim.Spec, res.amount)
	if !found {
		return res
	}
	res.succeeded = true
	res.amount = 1
	if prim.PublishLinked != "" {
		rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
	}
	return res
}

func handleReveal(r *effectResolver, prim game.Reveal) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if prim.Card.Kind != game.CardReferenceNone {
		cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
		if !ok || fromZone != zone.Library {
			return res
		}
		card, ok := r.game.GetCardInstance(cardID)
		if !ok {
			return res
		}
		emitCardRevealEvent(r.game, r.obj, card.Owner, cardID, fromZone)
		res.amount = 1
		res.succeeded = true
		return res
	}
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
	if len(prim.Sources) > 0 {
		refs := make([]game.CardReference, 0, len(prim.Sources))
		for _, source := range prim.Sources {
			ref, ok := source.CardRef()
			if !ok {
				return res
			}
			refs = append(refs, ref)
		}
		res.succeeded = r.putReferencedCardsOnBattlefieldValue(
			refs,
			recipient,
			prim.ContinuousEffects,
			battlefieldEntryOptions(prim),
		)
		return res
	}
	if card, ok := prim.Source.CardRef(); ok {
		var key game.LinkedObjectKey
		if prim.PublishLinked != "" {
			key = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
			clearLinkedObjects(r.game, key)
		}
		permanent, succeeded := r.putReferencedCardOnBattlefieldValue(card, recipient, prim.ContinuousEffects, battlefieldEntryOptions(prim))
		res.succeeded = succeeded
		if succeeded && prim.PublishLinked != "" {
			rememberLinkedObject(
				r.game,
				key,
				permanentLinkedObjectRef(permanent),
			)
		}
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
			res.succeeded = returnLinkedNonBattlefieldObjects(r.engine, r.game, r.obj, string(key), prim.LinkedReturnZonesOrExile(), controllerOverride, battlefieldEntryOptions(prim), r.agents, r.log)
		}
	}
	return res
}

func handleCreateToken(r *effectResolver, prim game.CreateToken) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if res.amount <= 0 {
		res.amount = 1
	}
	if prim.RecipientGroup.Kind != game.PlayerGroupReferenceNone {
		return r.createTokenForGroup(prim, res.amount)
	}
	var recipientRef game.PlayerReference
	if prim.Recipient.Exists {
		recipientRef = prim.Recipient.Val
	}
	recipient, ok := r.recipientController(recipientRef)
	if !ok {
		return res
	}
	if spec, ok := prim.Source.TokenCopy(); ok && spec.Source == game.TokenCopySourceEachInGroup {
		return r.createCopyTokensForEach(prim, spec, recipient)
	}
	if spec, ok := prim.Source.TokenCopy(); ok && spec.Source == game.TokenCopySourceChosenFromTriggerBatch {
		return r.createCopyTokenFromTriggerBatch(prim, spec, recipient)
	}
	token, ok := r.typedTokenDefinition(prim.Source)
	if !ok {
		return res
	}
	if prim.Power.Exists && prim.Toughness.Exists {
		sized := *token
		sized.Power = opt.Val(game.PT{Value: r.quantity(prim.Power.Val)})
		sized.Toughness = opt.Val(game.PT{Value: r.quantity(prim.Toughness.Val)})
		token = &sized
	}
	created, ok := createTokenPermanentsCollectingWithChoices(r.engine, r.game, recipient, token, res.amount, prim.EntryTapped, r.agents, r.log)
	if !ok {
		return res
	}
	if prim.EntryAttacking {
		declareCreatedTokensAttacking(r.engine, r.game, recipient, created, r.agents, r.log)
	}
	if prim.PublishLinked != "" {
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		// Drop any tokens a prior resolution published under this key before
		// remembering this resolution's tokens, so a later "it" reference binds
		// to the tokens created now rather than to a stale entry. The key is
		// source-and-link scoped, hence constant across repeated activations of
		// the same ability, and a sacrificed token lingers in last-known
		// information; without clearing, a repeatable "create a token ... it ...
		// at the beginning of the next end step" (Feldon of the Third Path)
		// would resolve the dead first token and leak the new one. This mirrors
		// the other single-binding publish sites (handleLookAtLibraryTop) that
		// clear before remembering.
		clearLinkedObjects(r.game, key)
		for _, permanent := range created {
			rememberLinkedObject(r.game, key, game.LinkedObjectRef{ObjectID: permanent.ObjectID, CardID: permanent.CardInstanceID})
		}
	}
	res.succeeded = res.amount > 0
	return res
}

// createTokenForGroup creates the token for every player in the primitive's
// recipient group ("Each player creates a 1/1 white Soldier creature token.",
// "Each opponent creates a Treasure token."). Members are resolved in APNAP
// order so the created tokens enter in the correct turn-based sequence, and each
// member receives the full token amount. The reported amount is the total number
// of tokens created across the group.
func (r *effectResolver) createTokenForGroup(prim game.CreateToken, amount int) effectResolved {
	res := effectResolved{accepted: true}
	token, ok := r.typedTokenDefinition(prim.Source)
	if !ok {
		return res
	}
	if prim.Power.Exists && prim.Toughness.Exists {
		sized := *token
		sized.Power = opt.Val(game.PT{Value: r.quantity(prim.Power.Val)})
		sized.Toughness = opt.Val(game.PT{Value: r.quantity(prim.Toughness.Val)})
		token = &sized
	}
	members := playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.RecipientGroup))
	for _, member := range members {
		created, ok := createTokenPermanentsCollectingWithChoices(r.engine, r.game, member, token, amount, prim.EntryTapped, r.agents, r.log)
		if !ok {
			continue
		}
		if prim.EntryAttacking {
			declareCreatedTokensAttacking(r.engine, r.game, member, created, r.agents, r.log)
		}
		res.amount += len(created)
		res.succeeded = true
	}
	return res
}

// createCopyTokensForEach creates one token copying each member of the spec's
// controlled battlefield group ("For each token you control, create a token
// that's a copy of that permanent." — Second Harvest). It snapshots the group
// and builds every copy definition before creating any token so the new tokens
// are not themselves copied (the copies are created simultaneously, CR 707).
func (r *effectResolver) createCopyTokensForEach(prim game.CreateToken, spec game.TokenCopySpec, recipient game.PlayerID) effectResolved {
	res := effectResolved{accepted: true}
	members := r.groupPermanents(*spec.Group)
	defs := make([]*game.CardDef, 0, len(members))
	for _, member := range members {
		source, ok := permanentCopyDef(r.game, member)
		if !ok {
			continue
		}
		def, ok := applyTokenCopyOverrides(source, spec)
		if !ok {
			continue
		}
		defs = append(defs, def)
	}
	for _, def := range defs {
		created, ok := createTokenPermanentsCollectingWithChoices(r.engine, r.game, recipient, def, 1, prim.EntryTapped, r.agents, r.log)
		if !ok {
			continue
		}
		if prim.EntryAttacking {
			declareCreatedTokensAttacking(r.engine, r.game, recipient, created, r.agents, r.log)
		}
		res.amount++
		res.succeeded = true
	}
	return res
}

// createCopyTokenFromTriggerBatch creates one token copying a controller-chosen
// member of the resolving ability's triggering event batch ("create a token
// that's a copy of one of them.", Twilight Diviner). The candidate set is the
// permanents that triggered the resolving ability and are still on the
// battlefield; the controller chooses one and the token starts as a copy of it,
// applying the spec's copy modifiers. No candidates yields no token.
func (r *effectResolver) createCopyTokenFromTriggerBatch(prim game.CreateToken, spec game.TokenCopySpec, recipient game.PlayerID) effectResolved {
	res := effectResolved{accepted: true}
	candidates := r.triggeringBatchPermanents()
	chosen := r.chooseTriggeringBatchMember(recipient, candidates)
	if chosen == nil {
		return res
	}
	source, ok := permanentCopyDef(r.game, chosen)
	if !ok {
		return res
	}
	def, ok := applyTokenCopyOverrides(source, spec)
	if !ok {
		return res
	}
	created, ok := createTokenPermanentsCollectingWithChoices(r.engine, r.game, recipient, def, 1, prim.EntryTapped, r.agents, r.log)
	if !ok {
		return res
	}
	if prim.EntryAttacking {
		declareCreatedTokensAttacking(r.engine, r.game, recipient, created, r.agents, r.log)
	}
	res.amount = 1
	res.succeeded = true
	return res
}

// triggeringBatchPermanents returns the battlefield permanents that triggered
// the resolving ability: the entering permanents of its triggering event batch
// (the primary event plus every event sharing its simultaneous batch) that still
// match the ability's own trigger pattern and remain on the battlefield. The
// list is deduplicated and order-stable in event order.
func (r *effectResolver) triggeringBatchPermanents() []*game.Permanent {
	obj := r.obj
	if obj == nil || !obj.HasTriggerEvent {
		return nil
	}
	pattern, ok := resolvingTriggerPattern(r.game, obj)
	if !ok {
		return nil
	}
	source, _ := permanentByObjectID(r.game, obj.SourceID)
	batchID := obj.TriggerEvent.SimultaneousID
	seen := make(map[id.ID]bool)
	var members []*game.Permanent
	consider := func(event game.Event) {
		if event.PermanentID == 0 || seen[event.PermanentID] {
			return
		}
		if !triggerMatchesEvent(r.game, source, pattern, event) {
			return
		}
		permanent, ok := permanentByObjectID(r.game, event.PermanentID)
		if !ok || !activeBattlefieldPermanent(permanent) {
			return
		}
		seen[event.PermanentID] = true
		members = append(members, permanent)
	}
	consider(obj.TriggerEvent)
	if batchID != 0 {
		for _, event := range r.game.Events {
			if event.SimultaneousID == batchID {
				consider(event)
			}
		}
	}
	return members
}

// chooseTriggeringBatchMember asks chooser to pick one of the triggering-batch
// candidates to copy. A single candidate is chosen automatically; an empty set
// yields nil.
func (r *effectResolver) chooseTriggeringBatchMember(chooser game.PlayerID, candidates []*game.Permanent) *game.Permanent {
	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	options := make([]game.ChoiceOption, 0, len(candidates))
	for i, candidate := range candidates {
		options = append(options, game.ChoiceOption{Index: i, Label: permanentEffectiveName(r.game, candidate)})
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           chooser,
		Prompt:           "Choose a creature to copy",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(candidates) {
		return candidates[0]
	}
	return candidates[selected[0]]
}

// resolvingTriggerPattern returns the trigger pattern of the triggered ability
// represented by obj, whether it is an inline-generated trigger or an ability
// addressable on the source definition by AbilityIndex.
func resolvingTriggerPattern(g *game.Game, obj *game.StackObject) (*game.TriggerPattern, bool) {
	if obj.InlineTrigger != nil {
		return &obj.InlineTrigger.Trigger.Pattern, true
	}
	def, ok := stackObjectSourceDef(g, obj)
	if !ok {
		return nil, false
	}
	body, ok := def.BodyAt(obj.AbilityIndex).(*game.TriggeredAbility)
	if !ok {
		return nil, false
	}
	return &body.Trigger.Pattern, true
}

func handleShufflePermanentIntoLibrary(r *effectResolver, prim game.ShufflePermanentIntoLibrary) effectResolved {
	res := effectResolved{accepted: true}
	resolved, ok := resolveObjectReference(r.game, r.obj, prim.Object)
	if !ok {
		return res
	}
	if resolved.permanent != nil {
		owner := resolved.permanent.Owner
		if !movePermanentToZone(r.game, resolved.permanent, zone.Library) {
			return res
		}
		if player, ok := playerByID(r.game, owner); ok {
			player.Library.Shuffle(r.engine.rng)
		}
		res.succeeded = true
		return res
	}
	// The permanent has left the battlefield: a dies / put-into-graveyard
	// trigger's "Shuffle it into its owner's library." resolves after the object
	// became a card in the graveyard, so its last-known snapshot names the card
	// to move. Shuffle that card from wherever it now is into its owner's
	// library. A token snapshot carries no card and ceases to exist, so it moves
	// nothing (CR 111.7).
	cardID := resolved.snapshot.CardID
	if cardID == 0 {
		return res
	}
	owner := resolved.snapshot.Owner
	current, ok := cardZone(r.game, cardID)
	if !ok || current != zone.Graveyard {
		// The card only remains this ability's object while it stays in the
		// graveyard it went to on death (CR 400.7). If it left in response to the
		// trigger — an opponent exiling it, the owner returning it to hand — it is
		// a new object the ability no longer tracks, so shuffle nothing rather
		// than dragging it out of its new zone. This lets the documented response
		// (exile it from the graveyard before it is shuffled) work.
		return res
	}
	if !moveCardBetweenZones(r.game, owner, cardID, current, zone.Library) {
		return res
	}
	if player, ok := playerByID(r.game, owner); ok {
		player.Library.Shuffle(r.engine.rng)
	}
	res.succeeded = true
	return res
}

func handleShuffleSpellIntoLibrary(r *effectResolver, _ game.ShuffleSpellIntoLibrary) effectResolved {
	res := effectResolved{accepted: true}
	if r.obj != nil {
		r.obj.ShuffleIntoLibraryOnResolution = true
		res.succeeded = true
	}
	return res
}

func handlePutPermanentOnLibrary(r *effectResolver, prim game.PutPermanentOnLibrary) effectResolved {
	res := effectResolved{accepted: true}
	permanent, ok := r.resolveObject(prim.Object)
	if !ok {
		return res
	}
	owner := permanent.Owner
	cardID := permanent.CardInstanceID
	token := permanent.Token
	if !movePermanentToZone(r.game, permanent, zone.Library) {
		return res
	}
	if prim.Bottom && !token {
		if player, ok := playerByID(r.game, owner); ok && player.Library.Remove(cardID) {
			player.Library.AddToBottom(cardID)
		}
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
	if prim.SourceSpell {
		if r.obj != nil {
			r.obj.ExileOnResolution = true
			res.succeeded = true
		}
		return res
	}
	targets := r.resolveObjectGroup(prim.Object, prim.Group)
	if !targets.single {
		// A group exile that carries a linked key (group blink) must remember
		// every exiled permanent under that key, capturing each link before the
		// move so a later linked return brings the whole group back together.
		var key game.LinkedObjectKey
		if prim.ExileLinkedKey != "" {
			key = linkedObjectSourceKey(r.game, r.obj, string(prim.ExileLinkedKey))
		}
		for _, permanent := range targets.permanents {
			linkedObjectRef := permanentLinkedObjectRef(permanent)
			if movePermanentToZone(r.game, permanent, zone.Exile) {
				res.succeeded = true
				if prim.ExileLinkedKey != "" {
					rememberLinkedObject(r.game, key, linkedObjectRef)
				}
			}
		}
		return res
	}
	if !targets.resolved {
		return res
	}
	permanent := targets.permanents[0]
	linkedObjectRef := permanentLinkedObjectRef(permanent)
	res.succeeded = movePermanentToZone(r.game, permanent, zone.Exile)
	if prim.ExileLinkedKey != "" {
		rememberLinkedObject(r.game, linkedObjectSourceKey(r.game, r.obj, string(prim.ExileLinkedKey)), linkedObjectRef)
	}
	return res
}

// handleChooseFromZone resolves the canonical choose-from-zone primitive: the
// resolving player chooses cards from prim.SourceZone matching prim.Filter, and
// each chosen card moves to prim.Destination with prim.Riders applied. It is the
// single runtime handler for the retired per-family wrappers (exile from hand /
// graveyard, put from hand, return from graveyard), all of which now lower to a
// game.ChooseFromZone envelope. res.amount preserves the requested quantity (the
// historical resolution reported the requested amount rather than the number
// chosen).
func handleChooseFromZone(r *effectResolver, prim game.ChooseFromZone) effectResolved {
	res := r.resolveChooseFromZone(prim)
	res.amount = r.quantity(prim.Quantity)
	return res
}

// handleCastForFree has the resolving player cast one card matching the
// selection from prim.Zone without paying its mana cost. It offers only cards
// with a legal cast choice; the enclosing instruction's Optional flag already
// gathered "you may" consent, so here the player picks which eligible spell to
// cast, casting nothing when none qualify.
//
// DO-NOT-COPY(zone-choice): the chosen card is cast (put on the stack) rather
// than moved to a destination zone, so it has no game.ChooseFromZone movement to
// reuse; prefer game.ChooseFromZone. (retire: #1396)
func handleCastForFree(r *effectResolver, prim game.CastForFree) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	if prim.Card.Kind != game.CardReferenceNone {
		cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
		if !ok || fromZone != prim.Zone {
			return res
		}
		if r.engine.castFreeTargetedSpell(r.game, playerID, cardID, prim.Zone, prim.ExileOnResolution, r.agents, r.log) {
			res.succeeded = true
		}
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	var candidates []id.ID
	for _, cardID := range castSourceZoneCards(player, prim.Zone) {
		card, cardOK := r.game.GetCardInstance(cardID)
		if !cardOK {
			continue
		}
		if !handCardMatchesSelection(r.game, card, prim.Selection, playerID) {
			continue
		}
		spellDef := cardFaceOrDefault(card, game.FaceFront)
		if _, _, legal := firstLegalSpellCastChoice(r.game, playerID, spellDef); !legal {
			continue
		}
		candidates = append(candidates, cardID)
	}
	if len(candidates) == 0 {
		return res
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, cardID := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(r.game, cardID),
			Card:  cardChoiceInfo(r.game, cardID),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose a spell to cast without paying its mana cost",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: firstChoiceIndices(1),
	}, r.log)
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		if r.engine.castFreeSpellFromZone(r.game, playerID, candidates[idx], prim.Zone, r.agents, r.log) {
			res.succeeded = true
		}
	}
	return res
}

// handleMassReturnFromGraveyard returns every matching graveyard card at once
// with no player choice, optionally scanning multiple players' graveyards
// (SourceGroup) and entering cards under their owners' control (ControlledByOwner).
//
// DO-NOT-COPY(zone-choice): the player makes no choice (all matching cards move)
// and the candidate pool spans several players' graveyards, neither of which the
// single-zone, choice-issuing game.ChooseFromZone envelope models; prefer
// game.ChooseFromZone. (retire: #1396)
func handleMassReturnFromGraveyard(r *effectResolver, prim game.MassReturnFromGraveyard) effectResolved {
	res := effectResolved{accepted: true}
	controllerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	var sources []game.PlayerID
	if prim.SourceGroup.Kind != game.PlayerGroupReferenceNone {
		sources = playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.SourceGroup))
	} else {
		sources = []game.PlayerID{controllerID}
	}
	type graveyardCandidate struct {
		cardID id.ID
		owner  game.PlayerID
	}
	var candidates []graveyardCandidate
	for _, owner := range sources {
		player, playerOK := playerByID(r.game, owner)
		if !playerOK {
			continue
		}
		for _, cardID := range player.Graveyard.All() {
			card, cardOK := r.game.GetCardInstance(cardID)
			if !cardOK {
				continue
			}
			if handCardMatchesSelection(r.game, card, prim.Selection, owner) {
				candidates = append(candidates, graveyardCandidate{cardID: cardID, owner: owner})
			}
		}
	}
	if prim.FromTriggerBatch {
		batch := triggerBatchCardIDs(r.game, r.obj)
		filtered := candidates[:0]
		for _, candidate := range candidates {
			if batch[candidate.cardID] {
				filtered = append(filtered, candidate)
			}
		}
		candidates = filtered
	}
	if len(candidates) == 0 {
		return res
	}
	if prim.Destination == zone.Battlefield {
		resolved := make([]resolvedBattlefieldCard, 0, len(candidates))
		for _, candidate := range candidates {
			card, cardOK := r.game.GetCardInstance(candidate.cardID)
			if !cardOK {
				continue
			}
			controller := controllerID
			if prim.ControlledByOwner {
				controller = candidate.owner
			}
			resolved = append(resolved, resolvedBattlefieldCard{
				card:       card,
				fromZone:   zone.Graveyard,
				controller: controller,
			})
		}
		res.succeeded = r.putResolvedCardsOnBattlefieldValue(resolved, nil, permanentCreationOptions{ForceTapped: prim.EntryTapped})
		return res
	}
	for _, candidate := range candidates {
		card, cardOK := r.game.GetCardInstance(candidate.cardID)
		if !cardOK {
			continue
		}
		if moveCardBetweenZonesWithPlacement(r.game, card.Owner, candidate.cardID, zone.Graveyard, prim.Destination, false) {
			res.succeeded = true
		}
	}
	return res
}

// triggerBatchCardIDs returns the card IDs that triggered the resolving
// one-or-more zone-change ability: the cards whose simultaneous zone-change
// events (the retained trigger event plus every event sharing its
// SimultaneousID) match the resolving trigger's pattern. Filtering by the
// pattern keeps "put them onto the battlefield" restricted to the cards the
// trigger actually fired for (e.g. only the land cards of "one or more land
// cards"), mirroring triggeringBatchPermanents for graveyard-bound cards.
func triggerBatchCardIDs(g *game.Game, obj *game.StackObject) map[id.ID]bool {
	if obj == nil || !obj.HasTriggerEvent {
		return nil
	}
	pattern, ok := resolvingTriggerPattern(g, obj)
	if !ok {
		return nil
	}
	source, _ := permanentByObjectID(g, obj.SourceID)
	ids := make(map[id.ID]bool)
	consider := func(event game.Event) {
		if event.CardID == 0 || ids[event.CardID] {
			return
		}
		if !triggerMatchesEvent(g, source, pattern, event) {
			return
		}
		ids[event.CardID] = true
	}
	consider(obj.TriggerEvent)
	if obj.TriggerEvent.SimultaneousID != 0 {
		for _, event := range g.Events {
			if event.SimultaneousID == obj.TriggerEvent.SimultaneousID {
				consider(event)
			}
		}
	}
	return ids
}

// handleMassReanimationExchange resolves "Each player exiles all <type> cards
// from their graveyard, then sacrifices all <type> they control, then puts all
// cards they exiled this way onto the battlefield." For every player it exiles
// the matching graveyard cards first, then sacrifices the matching battlefield
// permanents, then returns the just-exiled cards to the battlefield under their
// owners' control. Exiling before sacrificing keeps the freshly sacrificed
// permanents out of the returned set, realizing the "cards they exiled this way"
// back-reference.
func handleMassReanimationExchange(r *effectResolver, prim game.MassReanimationExchange) effectResolved {
	res := effectResolved{accepted: true}
	players := playersInAPNAPOrder(r.game, r.playerGroupMembers(game.AllPlayersReference()))
	resolver := newReferenceResolver(r.game, r.obj)
	type exiledCard struct {
		cardID id.ID
		owner  game.PlayerID
	}
	var exiled []exiledCard
	for _, owner := range players {
		player, ok := playerByID(r.game, owner)
		if !ok {
			continue
		}
		for _, cardID := range player.Graveyard.All() {
			card, cardOK := r.game.GetCardInstance(cardID)
			if !cardOK {
				continue
			}
			if handCardMatchesSelection(r.game, card, prim.Selection, owner) {
				exiled = append(exiled, exiledCard{cardID: cardID, owner: owner})
			}
		}
	}
	for _, candidate := range exiled {
		moveCardBetweenZonesWithPlacement(r.game, candidate.owner, candidate.cardID, zone.Graveyard, zone.Exile, false)
	}
	var sacrificed []*game.Permanent
	for _, permanent := range r.game.Battlefield {
		if permanentCantBeSacrificed(r.game, permanent) {
			continue
		}
		if resolver.permanentMatchesGroupSelection(&prim.Selection, nil, permanent) {
			sacrificed = append(sacrificed, permanent)
		}
	}
	if len(sacrificed) > 0 {
		sacrificePermanentsSimultaneously(r.game, sacrificed)
	}
	resolved := make([]resolvedBattlefieldCard, 0, len(exiled))
	for _, candidate := range exiled {
		card, cardOK := r.game.GetCardInstance(candidate.cardID)
		if !cardOK {
			continue
		}
		resolved = append(resolved, resolvedBattlefieldCard{
			card:       card,
			fromZone:   zone.Exile,
			controller: candidate.owner,
		})
	}
	res.succeeded = r.putResolvedCardsOnBattlefieldValue(resolved, nil, permanentCreationOptions{})
	return res
}

func handleBounce(r *effectResolver, prim game.Bounce) effectResolved {
	res := effectResolved{accepted: true}
	if prim.ControlledChoice {
		res.succeeded = movePermanentsToZoneSimultaneously(
			r.game,
			r.chooseControlledBouncePermanents(prim),
			zone.Hand,
		)
		return res
	}
	targets := r.resolveObjectGroup(prim.Object, prim.Group)
	if !targets.single {
		res.succeeded = movePermanentsToZoneSimultaneously(r.game, targets.permanents, zone.Hand)
		return res
	}
	if targets.resolved {
		res.succeeded = movePermanentToZone(r.game, targets.permanents[0], zone.Hand)
		return res
	}
	if resolved, ok := resolveObjectReference(r.game, r.obj, prim.Object); ok && resolved.stack != nil {
		res.succeeded = bounceStackSpellToHand(r.game, resolved.stack)
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
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		return handleMoveCardPlayerGroup(r, prim)
	}
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
	// Place the named exile counter only if the card actually landed in exile: a
	// CR 614/903.9 replacement or commander redirect can send an exile-bound move
	// elsewhere while still succeeding, and gating on the intended Destination
	// would orphan a counter on a card that never reached exile.
	if res.succeeded && prim.Counter.Exists && r.game.Players[card.Owner].Exile.Contains(cardID) {
		r.game.AddExileCounter(cardID, prim.Counter.Val, 1)
	}
	return res
}

// handleMoveCardPlayerGroup resolves the player-group zone form of MoveCard,
// moving every card in each affected player's source zone to the destination at
// once ("Exile all graveyards."). All moves across all players share one
// SimultaneousID so they emit as a single zone-change batch. Empty source zones
// are legal no-ops.
func handleMoveCardPlayerGroup(r *effectResolver, prim game.MoveCard) effectResolved {
	res := effectResolved{accepted: true}
	simultaneousID := r.game.IDGen.Next()
	for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
		from, ok := destinationZone(r.game, playerID, prim.FromZone)
		if !ok {
			continue
		}
		for _, cardID := range from.All() {
			card, ok := r.game.GetCardInstance(cardID)
			if !ok {
				continue
			}
			moved := moveCardBetweenZonesInBatch(r.game, card.Owner, cardID, prim.FromZone, prim.Destination, false, simultaneousID)
			res.succeeded = moved || res.succeeded
		}
	}
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
	if prim.Amount.IsDynamic() || prim.Amount.Value() != 0 {
		return handleMoveChosenHandCards(r, prim, playerID)
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

func handleMoveChosenHandCards(r *effectResolver, prim game.MoveCard, playerID game.PlayerID) effectResolved {
	res := effectResolved{accepted: true}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	candidates := player.Hand.All()
	amount := min(r.quantity(prim.Amount), len(candidates))
	res.amount = amount
	if amount <= 0 {
		return res
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, cardID := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(r.game, cardID),
			Card:  cardChoiceInfo(r.game, cardID),
		}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose cards to put on top of your library, top card first",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}, r.log)
	simultaneousID := r.game.IDGen.Next()
	for i := len(selected) - 1; i >= 0; i-- {
		idx := selected[i]
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		if moveCardBetweenZonesInBatch(
			r.game,
			playerID,
			candidates[idx],
			zone.Hand,
			zone.Library,
			false,
			simultaneousID,
		) {
			res.succeeded = true
		}
	}
	return res
}

// handleMoveCommander moves the resolved player's own commander cards from the
// command zone to the primitive's destination. It bypasses the commander-zone
// replacement (CR 903.9) by moving each card directly to the destination, since
// the effect explicitly relocates the commander. All moves share one
// SimultaneousID so they emit as a single zone-change batch.
func handleMoveCommander(r *effectResolver, prim game.MoveCommander) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	simultaneousID := r.game.IDGen.Next()
	for _, cardID := range player.CommandZone.All() {
		if !isCommanderCardID(r.game, cardID) {
			continue
		}
		card, ok := r.game.GetCardInstance(cardID)
		if !ok || card.Owner != playerID {
			continue
		}
		if moveCommanderToZone(r.game, playerID, cardID, prim.Destination, simultaneousID) {
			res.succeeded = true
		}
	}
	return res
}

// moveCommanderToZone relocates one commander card from the command zone to
// destination without applying the commander-zone replacement, while still
// honoring other zone-change replacements.
func moveCommanderToZone(g *game.Game, playerID game.PlayerID, cardID id.ID, destination zone.Type, simultaneousID id.ID) bool {
	event := game.Event{
		Kind:       game.EventZoneChanged,
		Controller: playerID,
		Player:     playerID,
		CardID:     cardID,
		FromZone:   zone.Command,
		ToZone:     destination,
	}
	replacement := replacementZoneChange(g, event)
	return moveCardBetweenZonesAfterReplacement(g, playerID, cardID, zone.Command, replacement, event, false, simultaneousID)
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

func handleExileForPlay(r *effectResolver, prim game.ExileForPlay) effectResolved {
	res := effectResolved{accepted: true}
	cardID, ok := exileForPlayCardID(r, prim)
	if !ok {
		return res
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return res
	}
	simultaneousID := r.game.IDGen.Next()
	if !moveCardBetweenZonesInBatch(r.game, card.Owner, cardID, prim.FromZone, zone.Exile, false, simultaneousID) {
		return res
	}
	kind := game.RuleEffectPlayFromZone
	if prim.Cast {
		kind = game.RuleEffectCastFromZone
	}
	r.game.RuleEffects = append(r.game.RuleEffects, game.RuleEffect{
		ID:             r.game.IDGen.Next(),
		Kind:           kind,
		Controller:     r.obj.Controller,
		SourceCardID:   r.obj.SourceCardID,
		SourceObjectID: r.obj.SourceID,
		AffectedPlayer: game.PlayerYou,
		Duration:       prim.Duration,
		CreatedTurn:    r.game.Turn.TurnNumber,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
		ExpiresFor:     r.obj.Controller,
	})
	res.succeeded = true
	return res
}

// handleExilePermanentForPlay exiles the target permanent from the battlefield
// and grants that card's owner permission to play it from exile for as long as
// it remains exiled (Prowl, Stoic Strategist). The permission is a
// DurationPermanent RuleEffectPlayFromZone: persistsWhileCardExiled keeps it
// active while the card stays in exile and drops it once the card leaves,
// matching "for as long as that card remains exiled". AffectToOwner scopes the
// permission to the exiled card's owner, who may be an opponent of the resolving
// controller. When LinkedKey is set the exiled card is remembered under the
// source-keyed linked set so a paired "whenever a player plays a card exiled
// with this" trigger recognizes its provenance.
func handleExilePermanentForPlay(r *effectResolver, prim game.ExilePermanentForPlay) effectResolved {
	res := effectResolved{accepted: true}
	targets := r.resolveObjectGroup(prim.Object, game.GroupReference{})
	if !targets.resolved || len(targets.permanents) == 0 {
		return res
	}
	permanent := targets.permanents[0]
	cardID := permanent.CardInstanceID
	owner := permanent.Owner
	if !movePermanentToZone(r.game, permanent, zone.Exile) {
		return res
	}
	if cardID == 0 {
		res.succeeded = true
		return res
	}
	r.game.RuleEffects = append(r.game.RuleEffects, game.RuleEffect{
		ID:             r.game.IDGen.Next(),
		Kind:           game.RuleEffectPlayFromZone,
		Controller:     r.obj.Controller,
		SourceCardID:   r.obj.SourceCardID,
		SourceObjectID: r.obj.SourceID,
		AffectedPlayer: game.PlayerYou,
		AffectToOwner:  true,
		Duration:       game.DurationPermanent,
		CreatedTurn:    r.game.Turn.TurnNumber,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
		ExpiresFor:     owner,
	})
	if prim.LinkedKey != "" {
		rememberLinkedObject(r.game, linkedObjectSourceKey(r.game, r.obj, string(prim.LinkedKey)), game.LinkedObjectRef{CardID: cardID})
	}
	res.succeeded = true
	return res
}

// handlePlayChosenExiledCard has the resolving controller choose one card in
// exile owned by a player matching OwnerScope (evaluated relative to the
// controller) and, when Counter is set, bearing that named exile marker counter,
// then grants the controller a per-card RuleEffectPlayFromZone so it may play the
// chosen card for prim.Duration (Dauthi Voidwalker: "Choose an exiled card an
// opponent owns with a void counter on it. You may play it this turn without
// paying its mana cost."). The chosen card commonly rests in an opponent's exile
// bucket; foreignExileCastableCards surfaces it to the controller and
// castSourcePlayer keeps zone containment pointed at the owner. When
// WithoutPayingManaCost is set the granted permission casts the chosen card's
// spell for free; a played land has no mana cost regardless. With no eligible
// card the effect is a legal no-op.
func handlePlayChosenExiledCard(r *effectResolver, prim game.PlayChosenExiledCard) effectResolved {
	res := effectResolved{accepted: true}
	you, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	var candidates []id.ID
	for pid := game.PlayerID(0); int(pid) < game.NumPlayers; pid++ {
		if !playerRelationMatches(you, pid, prim.OwnerScope) {
			continue
		}
		owner, ok := playerByID(r.game, pid)
		if !ok {
			continue
		}
		for _, cardID := range owner.Exile.All() {
			if prim.Counter.Exists && !r.game.HasExileCounter(cardID, prim.Counter.Val) {
				continue
			}
			candidates = append(candidates, cardID)
		}
	}
	if len(candidates) == 0 {
		return res
	}
	chosen, ok := r.choosePlayChosenExiledCard(candidates)
	if !ok {
		return res
	}
	r.game.RuleEffects = append(r.game.RuleEffects, game.RuleEffect{
		ID:                    r.game.IDGen.Next(),
		Kind:                  game.RuleEffectPlayFromZone,
		Controller:            r.obj.Controller,
		SourceCardID:          r.obj.SourceCardID,
		SourceObjectID:        r.obj.SourceID,
		AffectedPlayer:        game.PlayerYou,
		Duration:              prim.Duration,
		CreatedTurn:           r.game.Turn.TurnNumber,
		CastFromZone:          prim.Zone,
		AffectedCardID:        chosen,
		WithoutPayingManaCost: prim.WithoutPayingManaCost,
		ExpiresFor:            r.obj.Controller,
	})
	res.succeeded = true
	return res
}

// choosePlayChosenExiledCard asks the resolving controller which eligible exiled
// card to play. The choice is mandatory once at least one card qualifies.
func (r *effectResolver) choosePlayChosenExiledCard(candidates []id.ID) (id.ID, bool) {
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           "Choose an exiled card to play",
		Options:          chooseFromZoneOptions(r.game, candidates),
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(candidates) {
		return candidates[selected[0]], true
	}
	return 0, false
}

// mode it reads prim.Card and confirms it rests in FromZone. In SelectFromBatch
// mode it gathers the triggering batch's cards still in FromZone ("one of them"
// over a "discard one or more cards" batch) and has the resolving controller
// choose one; with a single eligible card the choice is made automatically.
func exileForPlayCardID(r *effectResolver, prim game.ExileForPlay) (id.ID, bool) {
	if !prim.SelectFromBatch {
		cardID, fromZone, ok := resolveCardReference(r.game, r.obj, prim.Card)
		if !ok || fromZone != prim.FromZone {
			return 0, false
		}
		return cardID, true
	}
	pool := exileForPlayBatchCards(r.game, r.obj, prim.FromZone)
	switch len(pool) {
	case 0:
		return 0, false
	case 1:
		return pool[0], true
	default:
		return r.chooseExileForPlayBatchCard(pool)
	}
}

// exileForPlayBatchCards returns the cards from the resolving object's triggering
// batch event that currently rest in fromZone, in event order with duplicates
// removed. A "discard one or more cards" trigger coalesces its simultaneous batch
// into one trigger and retains the first matching event, so the batch is the set
// of events sharing the trigger event's SimultaneousID, Kind, and affected player
// (CR 603.3a). A trigger with no batch (SimultaneousID zero) yields the lone
// triggering card.
func exileForPlayBatchCards(g *game.Game, obj *game.StackObject, fromZone zone.Type) []id.ID {
	if obj == nil || !obj.HasTriggerEvent {
		return nil
	}
	trigger := obj.TriggerEvent
	var pool []id.ID
	seen := make(map[id.ID]bool)
	consider := func(cardID id.ID) {
		if cardID == 0 || seen[cardID] {
			return
		}
		if cardZoneType, ok := cardZone(g, cardID); !ok || cardZoneType != fromZone {
			return
		}
		seen[cardID] = true
		pool = append(pool, cardID)
	}
	if trigger.SimultaneousID == 0 {
		consider(trigger.CardID)
		return pool
	}
	for _, event := range g.Events {
		if event.SimultaneousID == trigger.SimultaneousID &&
			event.Kind == trigger.Kind &&
			event.Player == trigger.Player {
			consider(event.CardID)
		}
	}
	return pool
}

// chooseExileForPlayBatchCard asks the resolving controller which of the batch's
// eligible cards to exile. The caller has already accepted the optional "you may
// exile" offer, so the selection itself is mandatory.
func (r *effectResolver) chooseExileForPlayBatchCard(pool []id.ID) (id.ID, bool) {
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           r.obj.Controller,
		Prompt:           "Choose a card to exile",
		Options:          chooseFromZoneOptions(r.game, pool),
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(pool) {
		return pool[selected[0]], true
	}
	return 0, false
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
	if permanentCantBeSacrificed(r.game, permanent) {
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
	var cantSacrifice []game.PlayerID
	for _, playerID := range players {
		if prim.All {
			chosen = append(chosen, playerControlledMatchingSelection(r.game, resolver, playerID, prim.Selection)...)
			continue
		}
		if prim.AnyNumber {
			chosen = append(chosen, r.engine.chooseAnyNumberToSacrificeForPlayer(r.game, resolver, playerID, prim.Selection, r.agents, r.log)...)
			continue
		}
		if prim.Fallback.Kind != game.SacrificeFallbackNone &&
			!playerControlsSelection(r.game, resolver, playerID, prim.Selection) {
			cantSacrifice = append(cantSacrifice, playerID)
			continue
		}
		chosen = append(chosen, r.engine.chooseSacrificePermanentsForPlayer(r.game, resolver, playerID, amount, prim.Selection, r.agents, r.log)...)
	}
	if prim.PublishLinked != "" {
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
		for _, permanent := range chosen {
			rememberLinkedObject(r.game, key, permanentLinkedObjectRef(permanent))
		}
	}
	// Report the number of permanents sacrificed as the resolved amount so a
	// count-scaled follow-up published off this instruction ("add that much",
	// "draw that many", "create that many") reads it. The amount is recorded
	// only when this instruction carries a PublishResult key, so it is inert for
	// the many edicts that do not publish a count.
	res.amount = len(chosen)
	res.succeeded = sacrificePermanentsSimultaneously(r.game, chosen)
	r.applySacrificeFallback(prim.Fallback, cantSacrifice)
	return res
}

// playerControlsSelection reports whether playerID controls at least one
// permanent that satisfies sel, i.e. can satisfy a SacrificePermanents edict.
func playerControlsSelection(g *game.Game, resolver referenceResolver, playerID game.PlayerID, sel game.Selection) bool {
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID {
			continue
		}
		if permanentCantBeSacrificed(g, permanent) {
			continue
		}
		if resolver.permanentMatchesGroupSelection(&sel, nil, permanent) {
			return true
		}
	}
	return false
}

// playerControlledMatchingSelection returns every permanent playerID controls
// that satisfies sel, in battlefield order. It backs the "sacrifices all <group>
// they control" mass form (All Is Dust), where each player loses every matching
// permanent rather than a chosen amount.
func playerControlledMatchingSelection(g *game.Game, resolver referenceResolver, playerID game.PlayerID, sel game.Selection) []*game.Permanent {
	var matching []*game.Permanent
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) != playerID {
			continue
		}
		if permanentCantBeSacrificed(g, permanent) {
			continue
		}
		if resolver.permanentMatchesGroupSelection(&sel, nil, permanent) {
			matching = append(matching, permanent)
		}
	}
	return matching
}

// applySacrificeFallback applies a SacrificePermanents edict's who-can't rider
// to each player who controlled no eligible permanent ("Each player who can't
// discards a card.").
func (r *effectResolver) applySacrificeFallback(fallback game.SacrificeFallback, players []game.PlayerID) {
	if fallback.Kind == game.SacrificeFallbackNone || len(players) == 0 {
		return
	}
	amount := r.quantity(fallback.Amount)
	for _, playerID := range players {
		switch fallback.Kind {
		case game.SacrificeFallbackDiscard:
			r.discardCardsWithChoices(playerID, amount, game.LinkedObjectKey{})
		case game.SacrificeFallbackLoseLife:
			loseLife(r.game, playerID, amount)
		default:
		}
	}
}

// playerHandSize reports how many cards playerID has in hand, used to gate
// discard alternatives (a player must be able to discard the full required
// count to pick that option).
func playerHandSize(g *game.Game, playerID game.PlayerID) int {
	player, ok := playerByID(g, playerID)
	if !ok {
		return 0
	}
	return player.Hand.Size()
}

// handleRepeatProcess resolves a "Repeat the following process X times. <body>"
// loop (Torment of Hailfire): it evaluates the repeat count and re-resolves the
// body content that many times. The body is re-resolved from scratch on each
// iteration so any per-player or random choices it makes recur independently.
func handleRepeatProcess(r *effectResolver, prim game.RepeatProcess) effectResolved {
	res := effectResolved{accepted: true}
	times := r.quantity(prim.Times)
	for range times {
		r.engine.resolveAbilityContentWithChoices(r.game, r.obj, prim.Body, r.agents, r.log)
		res.succeeded = true
	}
	return res
}

// handlePunisherEachLoseLife resolves the "punisher" family ("Each opponent
// loses N life unless that player sacrifices a permanent or discards a card."):
// each affected player decides, in APNAP order, whether to take the life loss
// or pay one of the offered alternatives (CR 608). A player who can perform no
// offered alternative simply loses the life.
func handlePunisherEachLoseLife(r *effectResolver, prim game.PunisherEachLoseLife) effectResolved {
	res := effectResolved{accepted: true}
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		return res
	}
	resolver := newReferenceResolver(r.game, r.obj)
	for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
		r.applyPunisherForPlayer(prim, resolver, playerID, amount)
		res.succeeded = true
	}
	return res
}

// punisherAction identifies the choice an affected player makes when facing a
// punisher effect.
type punisherAction uint8

const (
	punisherLoseLife punisherAction = iota
	punisherSacrifice
	punisherDiscard
)

// applyPunisherForPlayer offers one affected player the punisher's alternatives
// and applies the action they pick. When no alternative is available the player
// loses the life.
func (r *effectResolver) applyPunisherForPlayer(prim game.PunisherEachLoseLife, resolver referenceResolver, playerID game.PlayerID, amount int) {
	actions := []punisherAction{punisherLoseLife}
	options := []game.ChoiceOption{{Index: 0, Label: "Lose life"}}
	if prim.AllowSacrifice && playerControlsSelection(r.game, resolver, playerID, prim.SacrificeSelection) {
		options = append(options, game.ChoiceOption{Index: len(options), Label: "Sacrifice a permanent"})
		actions = append(actions, punisherSacrifice)
	}
	discardCount := max(prim.DiscardCount, 1)
	if prim.AllowDiscard && playerHandSize(r.game, playerID) >= discardCount {
		label := "Discard a card"
		if discardCount > 1 {
			label = fmt.Sprintf("Discard %d cards", discardCount)
		}
		options = append(options, game.ChoiceOption{Index: len(options), Label: label})
		actions = append(actions, punisherDiscard)
	}
	action := punisherLoseLife
	if len(options) > 1 {
		selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
			Kind:             game.ChoiceResolution,
			Player:           playerID,
			Prompt:           "Choose how to respond",
			Options:          options,
			MinChoices:       1,
			MaxChoices:       1,
			DefaultSelection: []int{0},
		}, r.log)
		if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(actions) {
			action = actions[selected[0]]
		}
	}
	switch action {
	case punisherSacrifice:
		chosen := r.engine.chooseSacrificePermanentsForPlayer(r.game, resolver, playerID, 1, prim.SacrificeSelection, r.agents, r.log)
		if len(chosen) == 0 {
			loseLife(r.game, playerID, amount)
			return
		}
		sacrificePermanentsSimultaneously(r.game, chosen)
	case punisherDiscard:
		if !r.discardCardsWithChoices(playerID, discardCount, game.LinkedObjectKey{}) {
			loseLife(r.game, playerID, amount)
		}
	default:
		loseLife(r.game, playerID, amount)
	}
}

func handleCounterObject(r *effectResolver, prim game.CounterObject) effectResolved {
	if prim.Object.Kind() != game.ObjectReferenceTargetStackObject {
		return effectResolved{accepted: true}
	}
	return effectResolved{accepted: true, succeeded: counterTargetStackObject(r.game, r.obj, prim.Object.TargetIndex(), prim.ExileInstead, prim.Destination)}
}

func handleChooseNewTargets(r *effectResolver, prim game.ChooseNewTargets) effectResolved {
	if prim.Object.Kind() != game.ObjectReferenceTargetStackObject {
		return effectResolved{accepted: true}
	}
	return effectResolved{
		accepted:  true,
		succeeded: r.engine.retargetStackObject(r.game, r.obj, prim.Object.TargetIndex(), r.agents, r.log),
	}
}

// handleCopyStackObject puts a copy of the targeted activated or triggered
// ability onto the stack (CR 707). The copy is a new object that resolves
// independently of the original; when the effect allows it, the resolving
// controller may choose new targets for the copy (CR 707.10c).
func handleCopyStackObject(r *effectResolver, prim game.CopyStackObject) effectResolved {
	original, ok := copyStackObjectSource(r, prim.Object)
	if !ok {
		return effectResolved{accepted: true}
	}
	chooser := r.obj.Controller
	if prim.Chooser.Exists {
		resolved, ok := r.resolvePlayer(prim.Chooser.Val)
		if !ok {
			return effectResolved{accepted: true}
		}
		chooser = resolved
	}
	copyObj := game.NewStackObjectCopy(original, r.game.IDGen.Next())
	// The copier controls the copy (CR 707.10a): the "may copy this spell" player
	// becomes the copy's controller so the copy's own iterative offer chains off
	// the copier's new target rather than the original controller.
	copyObj.Controller = chooser
	r.game.Stack.Push(copyObj)
	if prim.MayChooseNewTargets {
		r.engine.retargetStackObjectChoice(r.game, chooser, copyObj, r.agents, r.log)
	}
	return effectResolved{accepted: true, succeeded: true}
}

// copyStackObjectSource resolves the stack object a CopyStackObject effect
// copies. It supports a chosen stack-object target ("Copy target spell."), the
// triggering spell of a spell-cast trigger ("Whenever you cast a spell ...,
// copy that spell.", Reflections of Littjara), and the resolving spell itself
// ("copy this spell", Sevinne's Reclamation). The resolving spell has already
// been popped from the stack when its effects run, so the resolving case copies
// the resolving object directly rather than looking it up by ID.
func copyStackObjectSource(r *effectResolver, ref game.ObjectReference) (*game.StackObject, bool) {
	if ref.Kind() == game.ObjectReferenceResolvingStackObject {
		return r.obj, true
	}
	stackObjectID, ok := copyStackObjectSourceID(r.obj, ref)
	if !ok {
		return nil, false
	}
	return stackObjectByID(r.game, stackObjectID)
}

// copyStackObjectSourceID resolves the stack object id a CopyStackObject effect
// copies for references read from the stack ("Copy target spell." and the
// triggering spell of a spell-cast trigger).
func copyStackObjectSourceID(obj *game.StackObject, ref game.ObjectReference) (id.ID, bool) {
	switch ref.Kind() {
	case game.ObjectReferenceTargetStackObject:
		return effectStackObjectID(obj, ref.TargetIndex())
	case game.ObjectReferenceEventStackObject:
		if obj.HasTriggerEvent && obj.TriggerEvent.StackObjectID != 0 {
			return obj.TriggerEvent.StackObjectID, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// retargetStackObject re-chooses the targets of the spell or ability referenced
// by obj's target slot (CR 115.7). The new targets must be legal for the
// referenced object's own target specs, but the resolving controller of the
// retarget effect (obj.Controller) makes the selection.
func (e *Engine) retargetStackObject(g *game.Game, obj *game.StackObject, targetIndex int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	stackObjectID, ok := effectStackObjectID(obj, targetIndex)
	if !ok {
		return false
	}
	target, ok := stackObjectByID(g, stackObjectID)
	if !ok {
		return false
	}
	return e.retargetStackObjectChoice(g, obj.Controller, target, agents, log)
}

// retargetStackObjectChoice re-chooses the targets of target (CR 115.7). The
// new targets must be legal for target's own target specs, but the resolving
// controller chooser makes the selection. It is shared by the ChooseNewTargets
// retarget effect and by CopyStackObject's "you may choose new targets for the
// copy" rider, which retargets the freshly created copy.
func (e *Engine) retargetStackObjectChoice(g *game.Game, chooser game.PlayerID, target *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	specs, ok := stackObjectTargetSpecs(g, target)
	if !ok || len(specs.specs) == 0 {
		return false
	}
	// Re-enumerating a triggered ability's targets must reuse its captured
	// triggering event so event-relative target predicates (e.g. "target
	// creature that player controls", Garland, Royal Kidnapper) resolve the
	// same event player as the original enumeration and resolution recheck.
	retargetEvent := game.Event{}
	if target.HasTriggerEvent {
		retargetEvent = target.TriggerEvent
	}
	result := targetChoicesForSpecs(g, target.Controller, specs.def, specs.sourceObjectID, retargetEvent, specs.specs)
	if result.kind != targetLegalChoicesFound || len(result.choices) == 0 {
		return false
	}
	selected := e.chooseChoice(g, agents, targetChoiceRequest(chooser, "Choose new targets", result.choices), log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(result.choices) {
		return false
	}
	bound, ok := bindCardTargetZoneVersions(g, result.choices[selected[0]])
	if !ok {
		return false
	}
	target.Targets = bound
	target.TargetCounts = result.targetCounts[selected[0]]
	return true
}

// stackObjectSpecs bundles the target specs a stack object chose against with
// the source definition and source object ID that target legality is evaluated
// relative to.
type stackObjectSpecs struct {
	def            *game.CardDef
	sourceObjectID id.ID
	specs          []game.TargetSpec
}

// stackObjectTargetSpecs returns the target specs the referenced stack object
// chose against. It mirrors the per-kind body lookup used during normal
// resolution so a retarget reuses the same legality rules.
func stackObjectTargetSpecs(g *game.Game, obj *game.StackObject) (stackObjectSpecs, bool) {
	def, defOK := stackObjectSourceDef(g, obj)
	switch obj.Kind {
	case game.StackSpell:
		spellDef, spellOK := def, defOK
		if card, ok := g.GetCardInstance(obj.SourceID); ok {
			spellDef, spellOK = card.Def.FaceDef(obj.Face)
		}
		if !spellOK {
			return stackObjectSpecs{}, false
		}
		return stackObjectSpecs{def: spellDef, specs: spellTargetSpecs(spellDef, obj.ChosenModes)}, true
	case game.StackActivatedAbility:
		if !defOK {
			if permanent, permanentOK := permanentByObjectID(g, obj.SourceID); permanentOK {
				if physicalDef, physicalOK := physicalPermanentDef(g, permanent); physicalOK {
					def, defOK = physicalDef.FaceDef(obj.Face)
				}
			}
		}
		if !defOK {
			return stackObjectSpecs{}, false
		}
		body := stackObjectActivatedBody(def, obj)
		if body == nil {
			return stackObjectSpecs{}, false
		}
		return stackObjectSpecs{def: def, sourceObjectID: obj.SourceID, specs: bodyTargetSpecs(body, obj.ChosenModes)}, true
	case game.StackTriggeredAbility:
		var body game.Ability
		if obj.InlineTrigger != nil {
			body = obj.InlineTrigger
		} else if defOK {
			body = def.BodyAt(obj.AbilityIndex)
		}
		if body == nil {
			return stackObjectSpecs{}, false
		}
		return stackObjectSpecs{def: def, sourceObjectID: obj.SourceID, specs: bodyTargetSpecs(body, obj.ChosenModes)}, true
	default:
		return stackObjectSpecs{}, false
	}
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
		milled := millCards(r.game, playerID, res.amount)
		if prim.PublishLinked != "" {
			key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
			clearLinkedObjects(r.game, key)
			for _, cardID := range milled {
				rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
			}
		}
		res.succeeded = res.amount > 0
	}
	return res
}

func handleExileTopOfLibrary(r *effectResolver, prim game.ExileTopOfLibrary) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			exileTopOfLibraryCards(r.game, playerID, res.amount, prim.Counter, r.obj.Controller)
		}
		res.succeeded = res.amount > 0
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		exileTopOfLibraryCards(r.game, playerID, res.amount, prim.Counter, r.obj.Controller)
		res.succeeded = res.amount > 0
	}
	return res
}

// handlePutHandOnLibraryThenDraw has the resolving player put any number of
// cards from their hand on one end of their library, then draw a number of
// cards equal to the number put plus prim.DrawOffset.
// handleRevealUntil reveals cards from the top of one player's library, or each
// player's library in a referenced group, until a card matching prim.Until is
// revealed, then puts those cards into prim.Destination.
func handleRevealUntil(r *effectResolver, prim game.RevealUntil) effectResolved {
	res := effectResolved{accepted: true}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			revealUntilCards(r.game, playerID, prim.Until, prim.Destination)
		}
		res.succeeded = true
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		revealUntilCards(r.game, playerID, prim.Until, prim.Destination)
		res.succeeded = true
	}
	return res
}

func handlePutHandOnLibraryThenDraw(r *effectResolver, prim game.PutHandOnLibraryThenDraw) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	candidates := append([]id.ID(nil), player.Hand.All()...)
	put := 0
	if len(candidates) > 0 {
		options := make([]game.ChoiceOption, len(candidates))
		for i, cardID := range candidates {
			options[i] = game.ChoiceOption{
				Index: i,
				Label: cardChoiceLabel(r.game, cardID),
				Card:  cardChoiceInfo(r.game, cardID),
			}
		}
		selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
			Kind:       game.ChoiceResolution,
			Player:     playerID,
			Prompt:     "Choose any number of cards to put on your library",
			Options:    options,
			MinChoices: 0,
			MaxChoices: len(candidates),
		}, r.log)
		for _, idx := range selected {
			if idx < 0 || idx >= len(candidates) {
				continue
			}
			if moveCardBetweenZonesWithPlacement(r.game, playerID, candidates[idx], zone.Hand, zone.Library, prim.Bottom) {
				put++
			}
		}
	}
	res.amount = put + prim.DrawOffset
	if res.amount > 0 {
		res.succeeded = r.engine.drawCards(r.game, playerID, res.amount, r.agents, r.log)
	}
	return res
}

func handleDiscardThenDraw(r *effectResolver, prim game.DiscardThenDraw) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	candidates := append([]id.ID(nil), player.Hand.All()...)
	maxChoices := len(candidates)
	if prim.Max > 0 && prim.Max < maxChoices {
		maxChoices = prim.Max
	}
	discarded := 0
	if maxChoices > 0 {
		options := make([]game.ChoiceOption, len(candidates))
		for i, cardID := range candidates {
			options[i] = game.ChoiceOption{
				Index: i,
				Label: cardChoiceLabel(r.game, cardID),
				Card:  cardChoiceInfo(r.game, cardID),
			}
		}
		selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
			Kind:       game.ChoiceResolution,
			Player:     playerID,
			Prompt:     "Choose any number of cards to discard",
			Options:    options,
			MinChoices: 0,
			MaxChoices: maxChoices,
		}, r.log)
		simultaneousID := r.game.IDGen.Next()
		for _, idx := range selected {
			if idx < 0 || idx >= len(candidates) {
				continue
			}
			if discardCardFromHandInBatch(r.game, playerID, candidates[idx], simultaneousID) {
				discarded++
			}
		}
	}
	res.amount = discarded + prim.DrawOffset
	if res.amount > 0 {
		res.succeeded = r.engine.drawCards(r.game, playerID, res.amount, r.agents, r.log)
	}
	return res
}

// handleDiscardUnlessType resolves "discard N cards unless you discard a <type>
// card." (Thirst for Knowledge family): the player discards prim.Amount cards
// unless they instead discard a single exempt-type card. When the player holds
// an exempt-type card they choose which branch to take; otherwise they discard
// the full count. The exempt branch counts as full payment, so res.amount
// reflects the cards actually discarded.
func handleDiscardUnlessType(r *effectResolver, prim game.DiscardUnlessType) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	selection := game.Selection{RequiredTypesAny: prim.ExemptTypes}
	hasExempt := false
	for _, cardID := range player.Hand.All() {
		if card, found := r.game.GetCardInstance(cardID); found && handCardMatchesSelection(r.game, card, selection, playerID) {
			hasExempt = true
			break
		}
	}
	if hasExempt {
		exemptOptions := []game.ChoiceOption{
			{Index: 0, Label: "Discard one matching card"},
			{Index: 1, Label: "Discard cards"},
		}
		choice := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
			Kind:             game.ChoiceResolution,
			Player:           playerID,
			Prompt:           "Discard a matching card, or discard cards instead",
			Options:          exemptOptions,
			MinChoices:       1,
			MaxChoices:       1,
			DefaultSelection: []int{0},
		}, r.log)
		if len(choice) == 1 && choice[0] == 0 {
			res.succeeded = r.discardSingleMatching(playerID, selection)
			res.amount = 1
			return res
		}
	}
	res.succeeded = r.discardCardsWithChoices(playerID, prim.Amount, game.LinkedObjectKey{})
	res.amount = min(prim.Amount, len(player.Hand.All()))
	return res
}

// discardSingleMatching discards one chooser-selected card matching selection
// from the player's hand, returning whether a card was discarded.
func (r *effectResolver) discardSingleMatching(playerID game.PlayerID, selection game.Selection) bool {
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return false
	}
	var candidates []id.ID
	for _, cardID := range player.Hand.All() {
		if card, found := r.game.GetCardInstance(cardID); found && handCardMatchesSelection(r.game, card, selection, playerID) {
			candidates = append(candidates, cardID)
		}
	}
	if len(candidates) == 0 {
		return false
	}
	options := make([]game.ChoiceOption, len(candidates))
	for i, cardID := range candidates {
		options[i] = game.ChoiceOption{Index: i, Label: cardChoiceLabel(r.game, cardID), Card: cardChoiceInfo(r.game, cardID)}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose a card to discard",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) != 1 || selected[0] < 0 || selected[0] >= len(candidates) {
		return false
	}
	simultaneousID := r.game.IDGen.Next()
	return discardCardFromHandInBatch(r.game, playerID, candidates[selected[0]], simultaneousID)
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
		res.succeeded = r.engine.digCards(r.game, r.agents, r.log, r.obj, playerID, look, r.quantity(prim.Take), prim.Remainder, digFilter{
			selection:    prim.Filter,
			takeUpTo:     prim.TakeUpTo,
			reveal:       prim.Reveal,
			destination:  prim.Destination,
			entersTapped: prim.EntersTapped,
		})
	}
	return res
}

func handleRevealTopPartition(r *effectResolver, prim game.RevealTopPartition) effectResolved {
	amount := r.quantity(prim.Amount)
	res := effectResolved{accepted: true, amount: amount}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		res.succeeded = revealTopPartition(r.game, r.obj, playerID, amount, prim.Selection, prim.Remainder)
	}
	return res
}

func handleImpulseExile(r *effectResolver, prim game.ImpulseExile) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := resolvePlayerReference(r.game, r.obj, prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	amount := min(r.quantity(prim.Amount), player.Library.Size())
	res.amount = amount
	if amount <= 0 {
		return res
	}
	simultaneousID := r.game.IDGen.Next()
	var linkKey game.LinkedObjectKey
	if prim.PublishLinked != "" {
		linkKey = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
	}
	for range amount {
		cardID, ok := player.Library.Top()
		if !ok || !moveCardBetweenZonesInBatch(r.game, playerID, cardID, zone.Library, zone.Exile, false, simultaneousID) {
			continue
		}
		kind := game.RuleEffectPlayFromZone
		if prim.Cast {
			kind = game.RuleEffectCastFromZone
		}
		r.game.RuleEffects = append(r.game.RuleEffects, game.RuleEffect{
			ID:             r.game.IDGen.Next(),
			Kind:           kind,
			Controller:     r.obj.Controller,
			SourceCardID:   r.obj.SourceCardID,
			SourceObjectID: r.obj.SourceID,
			AffectedPlayer: game.PlayerYou,
			Duration:       prim.Duration,
			CreatedTurn:    r.game.Turn.TurnNumber,
			CastFromZone:   zone.Exile,
			AffectedCardID: cardID,
			ExpiresFor:     r.obj.Controller,
			SpendAnyMana:   prim.SpendAnyMana,
		})
		if prim.PublishLinked != "" {
			rememberLinkedObject(r.game, linkKey, game.LinkedObjectRef{CardID: cardID})
		}
		res.succeeded = true
	}
	return res
}

func handleExileLibraryUntilNonlandCast(r *effectResolver, prim game.ExileLibraryUntilNonlandCast) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := resolvePlayerReference(r.game, r.obj, prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	var found id.ID
	for {
		cardID, topOK := player.Library.Top()
		if !topOK {
			break
		}
		player.Library.Remove(cardID)
		player.Exile.Add(cardID)
		emitZoneChangeEvent(r.game, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Exile,
		})
		res.succeeded = true
		card, cardOK := r.game.GetCardInstance(cardID)
		if !cardOK {
			continue
		}
		if !cardFaceOrDefault(card, game.FaceFront).HasType(types.Land) {
			found = cardID
			break
		}
	}
	if found != 0 &&
		r.engine.chooseMay(r.game, r.agents, playerID, "Cast that card without paying its mana cost?", r.log) {
		r.engine.castFreeSpellFromExile(r.game, playerID, found, r.agents, r.log)
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
	if prim.Player.Kind() != game.PlayerReferenceNone {
		resolved, ok := r.resolvePlayer(prim.Player)
		if !ok {
			return res
		}
		playerID = resolved
	}
	if prim.PublishLinked != "" {
		clearLinkedObjects(r.game, linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)))
	}
	var manifested *game.Permanent
	var ok bool
	if prim.Dread {
		manifested, ok = r.engine.manifestDread(r.game, r.agents, r.log, playerID)
	} else {
		kind := game.FaceDownManifest
		if prim.Cloak {
			kind = game.FaceDownCloak
		}
		manifested, ok = r.engine.manifestTopCard(r.game, r.agents, r.log, playerID, kind)
	}
	res.succeeded = ok
	if ok && prim.PublishLinked != "" && manifested != nil {
		rememberLinkedObject(
			r.game,
			linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked)),
			permanentLinkedObjectRef(manifested),
		)
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
		ForceTapped:       prim.EntryTapped,
		EntersTransformed: prim.EntryTransformed,
		Counters:          prim.EntryCounters,
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
		if !ok || !cardConditionPredicateSatisfied(r.game, r.obj, card, cardCondition) {
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

func (r *effectResolver) putReferencedCardOnBattlefieldValue(ref game.CardReference, recipientRef game.PlayerReference, continuousEffects []game.ContinuousEffect, options permanentCreationOptions) (*game.Permanent, bool) {
	cardID, fromZone, ok := resolveCardReference(r.game, r.obj, ref)
	if !ok || fromZone == zone.None {
		return nil, false
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return nil, false
	}
	controller, ok := r.recipientController(recipientRef)
	if !ok {
		return nil, false
	}
	if ref.Kind == game.CardReferenceEvent {
		if owner, ok := playerByID(r.game, card.Owner); ok {
			controller = owner.ID
		}
	}
	return r.putResolvedCardOnBattlefieldValue(card, fromZone, controller, continuousEffects, options)
}

type resolvedBattlefieldCard struct {
	card       *game.CardInstance
	fromZone   zone.Type
	controller game.PlayerID
}

type preparedResolvedBattlefieldCard struct {
	resolved    resolvedBattlefieldCard
	event       game.Event
	replacement zoneChangeReplacementResult
}

func (r *effectResolver) putReferencedCardsOnBattlefieldValue(
	refs []game.CardReference,
	recipientRef game.PlayerReference,
	continuousEffects []game.ContinuousEffect,
	options permanentCreationOptions,
) bool {
	resolved := make([]resolvedBattlefieldCard, 0, len(refs))
	for _, ref := range refs {
		cardID, fromZone, ok := resolveCardReference(r.game, r.obj, ref)
		if !ok || fromZone == zone.None {
			continue
		}
		card, ok := r.game.GetCardInstance(cardID)
		if !ok {
			continue
		}
		controller, ok := r.recipientController(recipientRef)
		if !ok {
			continue
		}
		if ref.Kind == game.CardReferenceEvent {
			if owner, ok := playerByID(r.game, card.Owner); ok {
				controller = owner.ID
			}
		}
		resolved = append(resolved, resolvedBattlefieldCard{
			card:       card,
			fromZone:   fromZone,
			controller: controller,
		})
	}
	return r.putResolvedCardsOnBattlefieldValue(resolved, continuousEffects, options)
}

// putResolvedCardsOnBattlefieldValue moves each already-resolved card to the
// battlefield at once, sharing a simultaneous-entry ID when more than one card
// enters so their enter-the-battlefield events and replacements resolve as a
// single batch. It returns true when at least one card entered.
func (r *effectResolver) putResolvedCardsOnBattlefieldValue(
	resolved []resolvedBattlefieldCard,
	continuousEffects []game.ContinuousEffect,
	options permanentCreationOptions,
) bool {
	return len(r.putResolvedCardsOnBattlefieldCollecting(resolved, continuousEffects, options)) > 0
}

// putResolvedCardsOnBattlefieldCollecting is the collecting form of
// putResolvedCardsOnBattlefieldValue: it returns the permanents that entered so
// a caller can act on them (declaring them attacking for a "put ... onto the
// battlefield tapped and attacking" effect).
func (r *effectResolver) putResolvedCardsOnBattlefieldCollecting(
	resolved []resolvedBattlefieldCard,
	continuousEffects []game.ContinuousEffect,
	options permanentCreationOptions,
) []*game.Permanent {
	if len(resolved) > 1 {
		options.SimultaneousID = r.game.IDGen.Next()
	}
	moves := make([]preparedResolvedBattlefieldCard, 0, len(resolved))
	for _, card := range resolved {
		event := game.Event{
			Kind:           game.EventZoneChanged,
			Controller:     card.controller,
			Player:         card.card.Owner,
			CardID:         card.card.ID,
			FromZone:       card.fromZone,
			ToZone:         zone.Battlefield,
			SimultaneousID: options.SimultaneousID,
		}
		replacement := replacementZoneChange(r.game, event)
		replacement.destination = commanderReplacementDestination(
			r.game,
			card.card.ID,
			replacement.destination,
		)
		moves = append(moves, preparedResolvedBattlefieldCard{
			resolved:    card,
			event:       event,
			replacement: replacement,
		})
	}
	entries := make([]preparedCardPermanentEntry, 0, len(moves))
	for i := range moves {
		move := &moves[i]
		card := move.resolved
		if move.replacement.destination != zone.Battlefield {
			moveCardBetweenZonesAfterReplacement(
				r.game,
				card.card.Owner,
				card.card.ID,
				card.fromZone,
				move.replacement,
				move.event,
				false,
				options.SimultaneousID,
			)
			continue
		}
		if !removeCardFromZone(r.game, card.card.Owner, card.card.ID, card.fromZone) {
			continue
		}
		entry, ok := prepareCardPermanentFaceForSimultaneousEntry(
			r.engine,
			r.game,
			card.card,
			card.controller,
			card.fromZone,
			game.FaceFront,
			continuousEffects,
			options,
			r.agents,
			r.log,
		)
		if !ok {
			if cards, zoneOK := destinationZone(r.game, card.card.Owner, card.fromZone); zoneOK {
				cards.Add(card.card.ID)
			}
			continue
		}
		entries = append(entries, entry)
	}
	commitSimultaneousCardPermanentEntries(r.game, entries)
	permanents := make([]*game.Permanent, 0, len(entries))
	for i := range entries {
		permanents = append(permanents, entries[i].permanent)
	}
	return permanents
}

func (r *effectResolver) putResolvedCardOnBattlefieldValue(
	card *game.CardInstance,
	fromZone zone.Type,
	controller game.PlayerID,
	continuousEffects []game.ContinuousEffect,
	options permanentCreationOptions,
) (*game.Permanent, bool) {
	event := game.Event{
		Kind:           game.EventZoneChanged,
		Controller:     controller,
		Player:         card.Owner,
		CardID:         card.ID,
		FromZone:       fromZone,
		ToZone:         zone.Battlefield,
		SimultaneousID: options.SimultaneousID,
	}
	replacement := replacementZoneChange(r.game, event)
	replacement.destination = commanderReplacementDestination(r.game, card.ID, replacement.destination)
	if replacement.destination != zone.Battlefield {
		moveCardBetweenZonesAfterReplacement(
			r.game,
			card.Owner,
			card.ID,
			fromZone,
			replacement,
			event,
			false,
			options.SimultaneousID,
		)
		return nil, false
	}
	if !removeCardFromZone(r.game, card.Owner, card.ID, fromZone) {
		return nil, false
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
			destinationCards.Add(card.ID)
		}
		return nil, false
	}
	return permanent, permanent != nil
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
