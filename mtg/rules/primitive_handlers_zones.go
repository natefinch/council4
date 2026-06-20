package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
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

func handleDiscard(r *effectResolver, prim game.Discard) effectResolved {
	if prim.EntireHand {
		return handleDiscardEntireHand(r, prim)
	}
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if prim.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
		for _, playerID := range playersInAPNAPOrder(r.game, r.playerGroupMembers(prim.PlayerGroup)) {
			res.succeeded = discardCards(r.game, playerID, res.amount) || res.succeeded
		}
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	if ok {
		if prim.AtRandom {
			res.succeeded = r.discardCardsAtRandom(playerID, res.amount)
		} else {
			res.succeeded = r.discardCardsWithChoices(playerID, res.amount)
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

func (r *effectResolver) discardCardsWithChoices(playerID game.PlayerID, amount int) bool {
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
		discarded = discardCardFromHandInBatch(r.game, playerID, candidates[idx], simultaneousID) || discarded
	}
	return discarded
}

// discardCardsAtRandom discards up to amount cards chosen uniformly at random
// from the player's hand, as one simultaneous batch ("Discard a card at
// random."). It returns whether any card was discarded.
func (r *effectResolver) discardCardsAtRandom(playerID game.PlayerID, amount int) bool {
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return false
	}
	candidates := player.Hand.All()
	amount = min(amount, len(candidates))
	if amount <= 0 {
		return false
	}
	order := make([]int, len(candidates))
	for i := range order {
		order[i] = i
	}
	r.engine.rng.Shuffle(len(order), func(i, j int) {
		order[i], order[j] = order[j], order[i]
	})
	simultaneousID := r.game.IDGen.Next()
	discarded := false
	for _, idx := range order[:amount] {
		discarded = discardCardFromHandInBatch(r.game, playerID, candidates[idx], simultaneousID) || discarded
	}
	return discarded
}

func handleSearch(r *effectResolver, prim game.Search) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	if !searchSpecSupported(prim.Spec) {
		return res
	}
	playerID, ok := r.resolvePlayer(prim.Player)
	var key game.LinkedObjectKey
	if prim.PublishLinked != "" {
		key = linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
	}
	if ok {
		var permanent *game.Permanent
		res.succeeded, permanent = r.engine.searchLibrary(r.game, r.obj, r.agents, r.log, playerID, prim.Spec, res.amount)
		if prim.PublishLinked != "" && permanent != nil {
			rememberLinkedObject(r.game, key, permanentLinkedObjectRef(permanent))
		}
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
	if prim.SourceSpell {
		if r.obj != nil {
			r.obj.ExileOnResolution = true
			res.succeeded = true
		}
		return res
	}
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

// handleExileFromHand exiles up to prim.Amount cards a player chooses from hand
// that match prim.Selection, used for "you may exile a nonartifact, nonland card
// from your hand" (Chrome Mox's imprint). The whole instruction is optional, so
// the engine has already gathered the player's consent before this runs; here the
// player chooses which matching card to exile, if any. When prim.PublishLinked is
// set, the exiled card is linked to the source permanent by its object identity,
// so the imprint follows that specific object and a re-entered object (new object
// ID) finds no prior link. With no matching card, nothing is exiled and no link is
// recorded, leaving any reader (the imprint mana ability) with an empty color set.
func handleExileFromHand(r *effectResolver, prim game.ExileFromHand) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	publish := prim.PublishLinked != ""
	var key game.LinkedObjectKey
	if publish {
		key = linkedObjectByObjectKey(r.game, r.obj, string(prim.PublishLinked))
		clearLinkedObjects(r.game, key)
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	var candidates []id.ID
	for _, cardID := range player.Hand.All() {
		card, cardOK := r.game.GetCardInstance(cardID)
		if !cardOK {
			continue
		}
		if handCardMatchesSelection(r.game, card, prim.Selection, playerID) {
			candidates = append(candidates, cardID)
		}
	}
	amount := min(res.amount, len(candidates))
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
		Prompt:           "Choose a card to exile",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}, r.log)
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		cardID := candidates[idx]
		if moveCardBetweenZones(r.game, playerID, cardID, zone.Hand, zone.Exile) {
			res.succeeded = true
			if publish {
				rememberLinkedObject(r.game, key, game.LinkedObjectRef{CardID: cardID})
			}
		}
	}
	return res
}

// handlePutFromHand puts up to prim.Amount cards a player chooses from hand that
// match prim.Selection onto the battlefield under that player's control, used for
// ramp / cheat-into-play wording such as "put a land card from your hand onto the
// battlefield". A "you may" wrapper is expressed by the enclosing instruction's
// Optional flag, so the engine has already gathered consent before this runs;
// here the player chooses which matching card to put. With no matching card,
// nothing is put.
func handlePutFromHand(r *effectResolver, prim game.PutFromHand) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	var candidates []id.ID
	for _, cardID := range player.Hand.All() {
		card, cardOK := r.game.GetCardInstance(cardID)
		if !cardOK {
			continue
		}
		if handCardMatchesSelection(r.game, card, prim.Selection, playerID) {
			candidates = append(candidates, cardID)
		}
	}
	amount := min(res.amount, len(candidates))
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
		Prompt:           "Choose a card to put onto the battlefield",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}, r.log)
	creationOptions := permanentCreationOptions{ForceTapped: prim.EntersTapped}
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		card, cardOK := r.game.GetCardInstance(candidates[idx])
		if !cardOK {
			continue
		}
		if _, putOK := r.putResolvedCardOnBattlefieldValue(card, zone.Hand, playerID, nil, creationOptions); putOK {
			res.succeeded = true
		}
	}
	return res
}

func handleReturnFromGraveyard(r *effectResolver, prim game.ReturnFromGraveyard) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(prim.Amount)}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	var candidates []id.ID
	for _, cardID := range player.Graveyard.All() {
		card, cardOK := r.game.GetCardInstance(cardID)
		if !cardOK {
			continue
		}
		if handCardMatchesSelection(r.game, card, prim.Selection, playerID) {
			candidates = append(candidates, cardID)
		}
	}
	amount := min(res.amount, len(candidates))
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
		Prompt:           "Choose a card to return to your hand",
		Options:          options,
		MinChoices:       amount,
		MaxChoices:       amount,
		DefaultSelection: firstChoiceIndices(amount),
	}, r.log)
	for _, idx := range selected {
		if idx < 0 || idx >= len(candidates) {
			continue
		}
		card, cardOK := r.game.GetCardInstance(candidates[idx])
		if !cardOK {
			continue
		}
		if moveCardBetweenZonesWithPlacement(r.game, card.Owner, candidates[idx], zone.Graveyard, zone.Hand, false) {
			res.succeeded = true
		}
	}
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
	if prim.Group.Valid() {
		res.succeeded = movePermanentsToZoneSimultaneously(
			r.game,
			r.groupPermanents(prim.Group),
			zone.Hand,
		)
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
	for range amount {
		cardID, ok := player.Library.Top()
		if !ok || !moveCardBetweenZonesInBatch(r.game, playerID, cardID, zone.Library, zone.Exile, false, simultaneousID) {
			continue
		}
		r.game.RuleEffects = append(r.game.RuleEffects, game.RuleEffect{
			ID:             r.game.IDGen.Next(),
			Kind:           game.RuleEffectPlayFromZone,
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
		if !ok || !cardMatchesCondition(card.Def, cardCondition, r.obj) {
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
	return len(entries) > 0
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
