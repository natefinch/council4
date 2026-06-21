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
	if spec, ok := prim.Source.TokenCopy(); ok && spec.Source == game.TokenCopySourceEachInGroup {
		return r.createCopyTokensForEach(prim, spec, recipient)
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
	if prim.PublishLinked != "" {
		key := linkedObjectSourceKey(r.game, r.obj, string(prim.PublishLinked))
		for _, permanent := range created {
			rememberLinkedObject(r.game, key, game.LinkedObjectRef{ObjectID: permanent.ObjectID, CardID: permanent.CardInstanceID})
		}
	}
	res.succeeded = res.amount > 0
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

// handleCastForFree has the resolving player cast one card matching the
// selection from prim.Zone without paying its mana cost. It offers only cards
// with a legal cast choice; the enclosing instruction's Optional flag already
// gathered "you may" consent, so here the player picks which eligible spell to
// cast, casting nothing when none qualify.
func handleCastForFree(r *effectResolver, prim game.CastForFree) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
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
	var cantSacrifice []game.PlayerID
	for _, playerID := range players {
		if prim.Fallback.Kind != game.SacrificeFallbackNone &&
			!playerControlsSelection(r.game, resolver, playerID, prim.Selection) {
			cantSacrifice = append(cantSacrifice, playerID)
			continue
		}
		chosen = append(chosen, r.engine.chooseSacrificePermanentsForPlayer(r.game, resolver, playerID, amount, prim.Selection, r.agents, r.log)...)
	}
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
		if resolver.permanentMatchesGroupSelection(&sel, nil, permanent) {
			return true
		}
	}
	return false
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
			r.discardCardsWithChoices(playerID, amount)
		case game.SacrificeFallbackLoseLife:
			loseLife(r.game, playerID, amount)
		default:
		}
	}
}

// playerHasCardsInHand reports whether playerID has at least one card in hand,
// i.e. can pay a discard-a-card alternative.
func playerHasCardsInHand(g *game.Game, playerID game.PlayerID) bool {
	player, ok := playerByID(g, playerID)
	return ok && player.Hand.Size() > 0
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
	if prim.AllowDiscard && playerHasCardsInHand(r.game, playerID) {
		options = append(options, game.ChoiceOption{Index: len(options), Label: "Discard a card"})
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
		if !r.discardCardsWithChoices(playerID, 1) {
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
	return effectResolved{accepted: true, succeeded: counterTargetStackObject(r.game, r.obj, prim.Object.TargetIndex(), prim.ExileInstead)}
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
	if prim.Object.Kind() != game.ObjectReferenceTargetStackObject {
		return effectResolved{accepted: true}
	}
	stackObjectID, ok := effectStackObjectID(r.obj, prim.Object.TargetIndex())
	if !ok {
		return effectResolved{accepted: true}
	}
	original, ok := stackObjectByID(r.game, stackObjectID)
	if !ok {
		return effectResolved{accepted: true}
	}
	copyObj := game.NewStackObjectCopy(original, r.game.IDGen.Next())
	r.game.Stack.Push(copyObj)
	if prim.MayChooseNewTargets {
		r.engine.retargetStackObjectChoice(r.game, r.obj.Controller, copyObj, r.agents, r.log)
	}
	return effectResolved{accepted: true, succeeded: true}
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
	result := targetChoicesForSpecs(g, target.Controller, specs.def, specs.sourceObjectID, specs.specs)
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
	if prim.Player.Kind() != game.PlayerReferenceNone {
		resolved, ok := r.resolvePlayer(prim.Player)
		if !ok {
			return res
		}
		playerID = resolved
	}
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
