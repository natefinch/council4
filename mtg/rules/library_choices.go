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

// putLibraryCardIntoGraveyard moves cardID — already removed from playerID's
// library — toward playerID's graveyard, then emits the zone-change event. Every
// direct library-to-graveyard move (mill, surveil, dig remainder, manifest dread,
// explore, reveal-until/partition remainder) routes through it so they all honor
// the same replacement effects the central moveCardBetweenZonesAfterReplacement
// path applies (zones.go): graveyard-redirect replacements (CR 614; e.g. Dauthi
// Voidwalker's "If a card would be put into an opponent's graveyard from anywhere,
// instead exile it with a void counter on it.") and the commander replacement
// (CR 903.9a). When a redirect exiles the card it places the named exile counter
// once the card lands. simultaneousID batches simultaneous moves (a mill of
// several cards); pass 0 for a lone move. It returns the zone the card actually
// entered and whether the move succeeded.
func putLibraryCardIntoGraveyard(g *game.Game, playerID game.PlayerID, cardID, simultaneousID id.ID) (zone.Type, bool) {
	event := game.Event{
		Kind:           game.EventZoneChanged,
		Controller:     playerID,
		Player:         playerID,
		CardID:         cardID,
		FromZone:       zone.Library,
		ToZone:         zone.Graveyard,
		Amount:         1,
		SimultaneousID: simultaneousID,
	}
	replacement := replacementZoneChange(g, event)
	destination := commanderReplacementDestination(g, cardID, replacement.destination)
	zoneOwner := playerID
	if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
		zoneOwner = card.Owner
	}
	destinationCards, ok := destinationZone(g, zoneOwner, destination)
	if !ok {
		return destination, false
	}
	revealZoneReplacementSource(g, event, replacement.revealSource)
	destinationCards.Add(cardID)
	shuffleLibraryIfRequested(g, destinationCards, destination, replacement.shuffleIntoLibrary)
	placeRedirectExileCounter(g, zoneOwner, cardID, replacement)
	emitZoneChangeEvent(g, game.Event{
		Player:         playerID,
		CardID:         cardID,
		FromZone:       zone.Library,
		ToZone:         destination,
		Amount:         1,
		SimultaneousID: simultaneousID,
	})
	return destination, true
}

func millCards(g *game.Game, playerID game.PlayerID, amount int) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return nil
	}
	var milled []id.ID
	batchID := g.IDGen.Next()
	for range amount {
		cardID, ok := player.Library.Top()
		if !ok {
			return milled
		}
		player.Library.Remove(cardID)
		destination, ok := putLibraryCardIntoGraveyard(g, playerID, cardID, batchID)
		if !ok {
			return milled
		}
		if destination == zone.Graveyard {
			milled = append(milled, cardID)
		}
	}
	return milled
}

// revealUntilCards reveals cards from the top of playerID's library one at a
// time until a revealed card matches until, then puts every card revealed this
// way (including the matching card) into destination. When the library empties
// before a match, every revealed card is still moved. destination must be
// zone.Graveyard or zone.Hand; a graveyard move honors the commander
// replacement (CR 903.9a).
func revealUntilCards(g *game.Game, playerID game.PlayerID, prim game.RevealUntil) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return
	}
	if prim.MatchToDestinationRestRandomBottom {
		revealUntilMatchAndBottomRest(g, playerID, prim)
		return
	}
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			return
		}
		player.Library.Remove(cardID)
		matched := revealedCardMatches(g, playerID, cardID, prim.Until)
		if prim.Destination == zone.Graveyard {
			if _, ok := putLibraryCardIntoGraveyard(g, playerID, cardID, 0); !ok {
				return
			}
		} else {
			destinationCards, ok := destinationZone(g, playerID, prim.Destination)
			if !ok {
				return
			}
			destinationCards.Add(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   prim.Destination,
				Amount:   1,
			})
		}
		if matched {
			return
		}
	}
}

func revealUntilMatchAndBottomRest(g *game.Game, playerID game.PlayerID, prim game.RevealUntil) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return
	}
	var remainder []id.ID
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		player.Library.Remove(cardID)
		if revealedCardMatches(g, playerID, cardID, prim.Until) {
			player.Hand.Add(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   zone.Hand,
				Amount:   1,
			})
			break
		}
		remainder = append(remainder, cardID)
	}
	g.RNG.Shuffle(len(remainder), func(i, j int) {
		remainder[i], remainder[j] = remainder[j], remainder[i]
	})
	for _, cardID := range remainder {
		player.Library.AddToBottom(cardID)
	}
}

// revealedCardMatches reports whether the card revealed from playerID's library
// satisfies the reveal-until predicate. An empty predicate matches the first
// revealed card.
func revealedCardMatches(g *game.Game, playerID game.PlayerID, cardID id.ID, until game.Selection) bool {
	if until.Empty() {
		return true
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	return matchSelection(&selectionSubject{
		kind:   subjectCard,
		g:      g,
		card:   card,
		viewer: playerID,
	}, &until)
}

func exileTopOfLibraryCards(g *game.Game, playerID game.PlayerID, amount int, counterKind opt.V[counter.Kind], exiledBy game.PlayerID, faceDown bool) []id.ID {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return nil
	}
	exiled := make([]id.ID, 0, amount)
	for range amount {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		player.Library.Remove(cardID)
		destination := commanderReplacementDestination(g, cardID, zone.Exile)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			break
		}
		destinationCards.Add(cardID)
		if destination == zone.Exile {
			exiled = append(exiled, cardID)
		}
		// Place the named exile counter only when the card actually landed in
		// exile: a CR 614/903.9 replacement or commander redirect can divert the
		// move to the command zone, and gating on the intended destination would
		// orphan a counter on a card that never reached exile. Record the exiling
		// controller alongside the counter so a paired play/cast-from-exile
		// permission can filter to cards "exiled by an ability you controlled".
		if counterKind.Exists && destination == zone.Exile {
			g.AddExileCounterFromController(cardID, counterKind.Val, 1, exiledBy)
		}
		// A face-down exile hides the card's identity from every observer. The
		// zone records the face-down state (cleared automatically when the card
		// leaves the zone); the command-zone redirect never applies here because
		// only a card's owner's exile is a valid face-down destination.
		if faceDown && destination == zone.Exile {
			destinationCards.SetFaceDown(cardID, true)
		}
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
	}
	return exiled
}

func (e *Engine) scryCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, amount int) {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return
	}
	// TODO: replace sequential prompts with one partition+ordering choice.
	for _, cardID := range peekLibrary(player, amount) {
		request := libraryChoiceRequest(game.ChoiceScry, playerID, "Scry: choose where to put card.", []string{"top", "bottom"})
		request.Subject = cardChoiceInfo(g, cardID)
		selected := e.chooseChoice(g, agents, request, log)
		if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
			player.Library.AddToBottom(cardID)
		}
	}
	emitEvent(g, game.Event{
		Kind:                       game.EventScry,
		Controller:                 playerID,
		Player:                     playerID,
		Amount:                     amount,
		PlayerEventOrdinalThisTurn: nextPlayerEventOrdinalThisTurn(g, game.EventScry, playerID),
	})
}

func (e *Engine) surveilCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, amount int) {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return
	}
	// TODO: replace sequential prompts with one partition+ordering choice.
	for _, cardID := range peekLibrary(player, amount) {
		request := libraryChoiceRequest(game.ChoiceSurveil, playerID, "Surveil: choose where to put card.", []string{"top", "graveyard"})
		request.Subject = cardChoiceInfo(g, cardID)
		selected := e.chooseChoice(g, agents, request, log)
		if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
			putLibraryCardIntoGraveyard(g, playerID, cardID, 0)
		}
	}
	emitEvent(g, game.Event{
		Kind:                       game.EventSurveil,
		Controller:                 playerID,
		Player:                     playerID,
		Amount:                     amount,
		PlayerEventOrdinalThisTurn: nextPlayerEventOrdinalThisTurn(g, game.EventSurveil, playerID),
	})
}

// digFilter carries the optional typed parameters of a Dig: an optional
// Selection restricting which looked-at cards may be taken, whether the take
// count is an upper bound (the controller may take fewer, including none),
// whether each taken card is revealed as it is taken, the zone the taken cards
// move to (the player's hand by default, or the battlefield), and whether
// battlefield-bound taken cards enter tapped. Its zero value reproduces the
// plain impulse dig: no filter, an exact take, no reveal, into hand.
type digFilter struct {
	selection    opt.V[game.Selection]
	takeUpTo     bool
	reveal       bool
	destination  zone.Type
	entersTapped bool
}

// digCards resolves a Dig effect: the player looks at the top look cards of
// their library, chooses take of them (bounded by the cards actually seen and,
// when filter.selection is present, by the cards matching it) to move to the
// filter.destination, and the remaining cards go to the destination identified
// by remainder (graveyard or the bottom of the library, in seen order). The
// taken cards go to the player's hand by default, to the battlefield when
// filter.destination is the battlefield (tapped when filter.entersTapped is
// set), or back onto the top of the library when filter.destination is the
// library ("put up to one of them on top of your library"). When
// filter.takeUpTo is set the controller may take fewer than take cards (down to
// none); when filter.reveal is set each taken card is revealed as it is taken.
func (e *Engine) digCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, obj *game.StackObject, playerID game.PlayerID, look, take int, remainder game.DigRemainder, filter digFilter, slots []digRouteSlot) bool {
	player, ok := playerByID(g, playerID)
	if !ok || look <= 0 {
		return false
	}
	seen := peekLibrary(player, look)
	if len(seen) == 0 {
		return false
	}
	eligible := seen
	if filter.selection.Exists {
		eligible = make([]id.ID, 0, len(seen))
		for _, cardID := range seen {
			card, cardOK := g.GetCardInstance(cardID)
			if cardOK && cardMatchesSelection(g, obj, card, filter.selection.Val) {
				eligible = append(eligible, cardID)
			}
		}
	}
	if take > len(eligible) {
		take = len(eligible)
	}
	minTake := take
	if filter.takeUpTo {
		minTake = 0
	}
	var taken []id.ID
	if take > 0 {
		taken = e.chooseDigCards(g, agents, log, playerID, eligible, minTake, take, filter.destination)
	}
	// A library-top destination returns the chosen cards to the top of the
	// library ("Put up to one of them on top of your library"). Place them in
	// reverse selection order so the first chosen card ends up on top.
	topOrdered := taken
	if filter.destination == zone.Library {
		topOrdered = make([]id.ID, len(taken))
		for i, cardID := range taken {
			topOrdered[len(taken)-1-i] = cardID
		}
	}
	for _, cardID := range topOrdered {
		if !player.Library.Remove(cardID) {
			continue
		}
		if filter.reveal {
			emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
		}
		switch filter.destination {
		case zone.Battlefield:
			card, cardOK := g.GetCardInstance(cardID)
			if !cardOK {
				continue
			}
			_, _ = createCardPermanentFaceWithOptions(e, g, card, playerID, zone.Library, game.FaceFront, nil, permanentCreationOptions{ForceTapped: filter.entersTapped}, agents, log)
		case zone.Library:
			// The card never leaves the library zone; it is only reordered onto
			// the top, so no cross-zone change event is emitted.
			player.Library.Add(cardID)
		default:
			player.Hand.Add(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   zone.Hand,
				Amount:   1,
			})
		}
	}
	routed := make(map[id.ID]bool, len(seen))
	for _, cardID := range taken {
		routed[cardID] = true
	}
	e.routeDigSlots(g, agents, log, obj, playerID, player, seen, routed, slots)
	for _, cardID := range seen {
		if routed[cardID] {
			continue
		}
		if !player.Library.Remove(cardID) {
			continue
		}
		if remainder == game.DigRemainderLibraryBottom {
			player.Library.AddToBottom(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   zone.Library,
				Amount:   1,
			})
			continue
		}
		putLibraryCardIntoGraveyard(g, playerID, cardID, 0)
	}
	return true
}

// digRouteSlot is one resolved ordered destination of a dig: count cards the
// player chooses from the still-unrouted looked-at cards move to destination
// (bottom of the library when bottom is set for a library destination), with an
// impulse play/cast grant applied to each exiled card when play is present. It
// is the runtime form of game.DigSlot with its Count already resolved.
type digRouteSlot struct {
	count       int
	destination zone.Type
	bottom      bool
	play        opt.V[game.ImpulsePlayGrant]
}

// routeDigSlots fans the looked-at cards into the ordered slots after the
// primary take. For each slot in printed order the digging player chooses from
// the cards not yet routed, taking as many as the remaining pool allows when the
// library is short ("as much as possible"), then those cards move to the slot's
// destination. It records every routed card in routed so the caller sends only
// the leftover looked-at cards to the remainder destination.
func (e *Engine) routeDigSlots(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, obj *game.StackObject, playerID game.PlayerID, player *game.Player, seen []id.ID, routed map[id.ID]bool, slots []digRouteSlot) {
	for _, slot := range slots {
		avail := make([]id.ID, 0, len(seen))
		for _, cardID := range seen {
			if !routed[cardID] {
				avail = append(avail, cardID)
			}
		}
		count := min(slot.count, len(avail))
		if count <= 0 {
			continue
		}
		for _, cardID := range e.chooseDigSlotCards(g, agents, log, playerID, avail, count, slot) {
			if routeDigSlotCard(g, obj, playerID, player, cardID, slot) {
				routed[cardID] = true
			}
		}
	}
}

// routeDigSlotCard moves cardID from playerID's library to slot's destination
// and, for an exile slot with a play grant, records the impulse play/cast
// permission over the exiled card. The exile slot moves the card through the
// replacement-aware batch mover and the graveyard slot removes it and places it
// through the replacement-aware graveyard mover, so a commander may be
// redirected to the command zone and graveyard-replacement effects apply,
// matching ImpulseExile and the dig graveyard remainder; the hand and library
// destinations move it directly like the primary take and the library-bottom
// remainder. It reports whether the card was routed.
func routeDigSlotCard(g *game.Game, obj *game.StackObject, playerID game.PlayerID, player *game.Player, cardID id.ID, slot digRouteSlot) bool {
	switch slot.destination {
	case zone.Exile:
		if !moveCardBetweenZonesInBatch(g, playerID, cardID, zone.Library, zone.Exile, false, 0) {
			return false
		}
		if slot.play.Exists {
			appendPlayFromExileGrant(g, obj, cardID, slot.play.Val)
		}
		return true
	case zone.Graveyard:
		if !player.Library.Remove(cardID) {
			return false
		}
		putLibraryCardIntoGraveyard(g, playerID, cardID, 0)
		return true
	case zone.Library:
		if !player.Library.Remove(cardID) {
			return false
		}
		if slot.bottom {
			player.Library.AddToBottom(cardID)
		} else {
			player.Library.Add(cardID)
		}
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Library,
			Amount:   1,
		})
		return true
	default:
		if !player.Library.Remove(cardID) {
			return false
		}
		player.Hand.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Hand,
			Amount:   1,
		})
		return true
	}
}

// chooseDigSlotCards asks the digging player which count of the avail cards to
// route to a dig slot's destination. It reuses the ChoiceDig pathway with the
// selection bounded to exactly count so a short library takes as many as remain.
// Agents that do not answer fall back to the deterministic first-count selection.
func (e *Engine) chooseDigSlotCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, avail []id.ID, count int, slot digRouteSlot) []id.ID {
	options := make([]game.ChoiceOption, len(avail))
	defaults := make([]int, 0, count)
	for i, cardID := range avail {
		options[i] = game.ChoiceOption{Index: i, Label: cardChoiceLabel(g, cardID), Card: cardChoiceInfo(g, cardID)}
		if i < count {
			defaults = append(defaults, i)
		}
	}
	selected := e.chooseChoice(g, agents, game.ChoiceRequest{
		Kind:             game.ChoiceDig,
		Player:           playerID,
		Prompt:           digSlotPrompt(slot),
		Options:          options,
		MinChoices:       count,
		MaxChoices:       count,
		DefaultSelection: defaults,
	}, log)
	chosen := make([]id.ID, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(avail) {
			chosen = append(chosen, avail[index])
		}
	}
	return chosen
}

// digSlotPrompt names the destination a dig slot's chosen cards move to so the
// choice prompt matches the printed effect.
func digSlotPrompt(slot digRouteSlot) string {
	switch slot.destination {
	case zone.Exile:
		return "Dig: choose cards to exile."
	case zone.Graveyard:
		return "Dig: choose cards to put into your graveyard."
	case zone.Library:
		if slot.bottom {
			return "Dig: choose cards to put on the bottom of your library."
		}
		return "Dig: choose cards to put on top of your library."
	default:
		return "Dig: choose cards to put into your hand."
	}
}

// chooseDigCards asks the digging player which of the eligible cards (already
// filtered to those that may be taken) to move to destination. minTake and
// maxTake bound the selection: an exact dig passes minTake == maxTake, while a
// typed "you may reveal up to N" dig passes minTake 0 so the player may decline.
// Agents that do not answer fall back to the deterministic first-take selection,
// preserving prior engine behavior.
func (e *Engine) chooseDigCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, eligible []id.ID, minTake, maxTake int, destination zone.Type) []id.ID {
	options := make([]game.ChoiceOption, 0, len(eligible))
	defaults := make([]int, 0, minTake)
	for i, cardID := range eligible {
		options = append(options, game.ChoiceOption{Index: i, Label: cardChoiceLabel(g, cardID), Card: cardChoiceInfo(g, cardID)})
		if i < minTake {
			defaults = append(defaults, i)
		}
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceDig,
		Player:           playerID,
		Prompt:           digChoicePrompt(destination),
		Options:          options,
		MinChoices:       minTake,
		MaxChoices:       maxTake,
		DefaultSelection: defaults,
	}
	selected := e.chooseChoice(g, agents, request, log)
	taken := make([]id.ID, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(eligible) {
			taken = append(taken, eligible[index])
		}
	}
	return taken
}

// digChoicePrompt names the destination the chosen dig cards move to so the
// choice prompt matches the printed effect.
func digChoicePrompt(destination zone.Type) string {
	switch destination {
	case zone.Battlefield:
		return "Dig: choose cards to put onto the battlefield."
	case zone.Library:
		return "Dig: choose cards to put on top of your library."
	default:
		return "Dig: choose cards to put into your hand."
	}
}

func (e *Engine) manifestTopCard(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, kind game.FaceDownKind) (*game.Permanent, bool) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil, false
	}
	cardID, ok := player.Library.Top()
	if !ok {
		return nil, false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok || !player.Library.Remove(cardID) {
		return nil, false
	}
	permanent, ok := createCardPermanentFaceDownWithChoices(e, g, card, playerID, zone.Library, game.FaceFront, kind, false, agents, log)
	return permanent, ok
}

func (e *Engine) manifestDread(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID) (*game.Permanent, bool) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil, false
	}
	cards := peekLibrary(player, 2)
	if len(cards) == 0 {
		return nil, false
	}
	chosenIndex := 0
	if len(cards) > 1 {
		selected := e.chooseChoice(g, agents, manifestDreadChoiceRequest(g, playerID, cards), log)
		if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(cards) {
			chosenIndex = selected[0]
		}
	}
	chosenID := cards[chosenIndex]
	chosen, ok := g.GetCardInstance(chosenID)
	if !ok || !player.Library.Remove(chosenID) {
		return nil, false
	}
	manifested, ok := createCardPermanentFaceDownWithChoices(e, g, chosen, playerID, zone.Library, game.FaceFront, game.FaceDownManifest, false, agents, log)
	if !ok {
		return nil, false
	}
	for _, cardID := range cards {
		if cardID == chosenID {
			continue
		}
		if !player.Library.Remove(cardID) {
			continue
		}
		putLibraryCardIntoGraveyard(g, playerID, cardID, 0)
	}
	return manifested, true
}

func manifestDreadChoiceRequest(g *game.Game, playerID game.PlayerID, cards []id.ID) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(cards))
	for i, cardID := range cards {
		label := "unknown card"
		if card, ok := g.GetCardInstance(cardID); ok {
			label = cardFaceOrDefault(card, game.FaceFront).Name
		}
		options = append(options, game.ChoiceOption{Index: i, Label: label})
	}
	return game.ChoiceRequest{
		Kind:             game.ChoiceManifest,
		Player:           playerID,
		Prompt:           "Manifest dread: choose a card to manifest.",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
}

// chooseSearchMatches asks the deciding player which matching library cards to
// take. The decider is the searching player for an ordinary search and the
// opponent controlling the search (Opposition Agent) when one applies; either way
// the candidate card names are offered as visible choice labels, so the decider
// sees the otherwise-hidden library. minChoices is zero for qualified or "up to"
// searches and one for an unrestricted exact-card search with a nonempty library,
// computed from the spec independently of who decides so fail-to-find stays
// correct.
func (e *Engine) chooseSearchMatches(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, decider game.PlayerID, candidates []id.ID, amount, minChoices int) []id.ID {
	if len(candidates) == 0 || amount <= 0 {
		return nil
	}
	selected := e.chooseChoice(g, agents, searchChoiceRequest(g, decider, candidates, amount, minChoices), log)
	found := make([]id.ID, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(candidates) {
			found = append(found, candidates[index])
		}
	}
	return found
}

func searchChoiceRequest(g *game.Game, decider game.PlayerID, candidates []id.ID, amount, minChoices int) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(candidates))
	for i, cardID := range candidates {
		label := "unknown card"
		if card, ok := g.GetCardInstance(cardID); ok {
			label = cardFaceOrDefault(card, game.FaceFront).Name
		}
		options = append(options, game.ChoiceOption{Index: i, Label: label})
	}
	maxChoices := min(amount, len(candidates))
	// Without an answering agent the engine falls back to DefaultSelection. Pick
	// the first maxChoices matches so nil/non-choice agents keep the prior
	// deterministic first-match behavior rather than failing to find.
	defaultSelection := make([]int, 0, maxChoices)
	for i := range maxChoices {
		defaultSelection = append(defaultSelection, i)
	}
	return game.ChoiceRequest{
		Kind:             game.ChoiceSearch,
		Player:           decider,
		Prompt:           "Search your library: choose matching cards to find.",
		Options:          options,
		MinChoices:       min(minChoices, maxChoices),
		MaxChoices:       maxChoices,
		DefaultSelection: defaultSelection,
	}
}

// chooseCorrelatedSearchMatches chooses up to amount matching library cards under
// a "share a land type" correlation: every chosen card must share at least one
// subtype with each other chosen card. It runs a staged dependent choice, one
// card at a time, only ever offering cards that still share a subtype with all
// cards already chosen, so an illegal combination cannot be assembled rather than
// being chosen and then silently dropped (CR 701.19). The player may stop early
// or fail to find entirely by choosing none at any stage. Agents that do not
// answer fall back to the deterministic first-compatible-card selection.
func (e *Engine) chooseCorrelatedSearchMatches(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, decider game.PlayerID, candidates []id.ID, amount int) []id.ID {
	if len(candidates) == 0 || amount <= 0 {
		return nil
	}
	remaining := slices.Clone(candidates)
	var found []id.ID
	var common []types.Sub // running subtype intersection; nil before the first pick
	for len(found) < amount {
		pool := make([]id.ID, 0, len(remaining))
		for _, cardID := range remaining {
			if len(found) == 0 || cardSharesAnySubtype(g, cardID, common) {
				pool = append(pool, cardID)
			}
		}
		if len(pool) == 0 {
			break
		}
		pick, ok := e.chooseCorrelatedSearchCard(g, agents, log, decider, pool)
		if !ok {
			break
		}
		found = append(found, pick)
		common = restrictSharedSubtypes(g, common, pick, len(found) == 1)
		remaining = removeFoundID(remaining, pick)
	}
	return found
}

// chooseDifferentNameSearchMatches chooses up to amount matching library cards
// under a "with different names" correlation: no two chosen cards may share a
// name. It runs a staged dependent choice, one card at a time, only ever
// offering cards whose name no already-chosen card has, so a duplicate-name set
// cannot be assembled rather than being chosen and then silently dropped
// (CR 701.19). The player may stop early or fail to find entirely by choosing
// none at any stage. Agents that do not answer fall back to the deterministic
// first-available-card selection.
func (e *Engine) chooseDifferentNameSearchMatches(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, decider game.PlayerID, candidates []id.ID, amount int) []id.ID {
	if len(candidates) == 0 || amount <= 0 {
		return nil
	}
	remaining := slices.Clone(candidates)
	var found []id.ID
	chosenNames := map[string]bool{}
	for len(found) < amount {
		pool := make([]id.ID, 0, len(remaining))
		for _, cardID := range remaining {
			if !chosenNames[searchCardName(g, cardID)] {
				pool = append(pool, cardID)
			}
		}
		if len(pool) == 0 {
			break
		}
		pick, ok := e.chooseCorrelatedSearchCard(g, agents, log, decider, pool)
		if !ok {
			break
		}
		found = append(found, pick)
		chosenNames[searchCardName(g, pick)] = true
		remaining = removeFoundID(remaining, pick)
	}
	return found
}

// searchCardName returns a library card's front-face name, or "" when the card
// cannot be resolved so an unidentifiable card never blocks a distinct pick.
func searchCardName(g *game.Game, cardID id.ID) string {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return ""
	}
	return cardFaceOrDefault(card, game.FaceFront).Name
}

// chooseCorrelatedSearchCard offers one optional pick from the still-compatible
// pool of a correlated search. It returns the chosen card and true, or ok=false
// when the player declines (choosing none stops the search). Agents that do not
// answer default to the first pool card so nil agents find deterministically.
func (e *Engine) chooseCorrelatedSearchCard(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, decider game.PlayerID, pool []id.ID) (id.ID, bool) {
	options := make([]game.ChoiceOption, 0, len(pool))
	for i, cardID := range pool {
		label := "unknown card"
		if card, ok := g.GetCardInstance(cardID); ok {
			label = cardFaceOrDefault(card, game.FaceFront).Name
		}
		options = append(options, game.ChoiceOption{Index: i, Label: label})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceSearch,
		Player:           decider,
		Prompt:           "Search your library: choose matching cards to find.",
		Options:          options,
		MinChoices:       0,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(pool) {
		return pool[selected[0]], true
	}
	return 0, false
}

// cardSharesAnySubtype reports whether the library card has any subtype in subs.
// It returns false for an empty subtype set so the first card of a correlated
// search is never gated by it.
func cardSharesAnySubtype(g *game.Game, cardID id.ID, subs []types.Sub) bool {
	if len(subs) == 0 {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	return card.Def.HasAnySubtype(subs...)
}

// restrictSharedSubtypes narrows the running subtype intersection of a correlated
// search by the just-picked card's subtypes. The first pick seeds the
// intersection with the card's full subtype list; each later pick keeps only the
// subtypes the new card also has, honoring dual basics that carry more than one
// land subtype.
func restrictSharedSubtypes(g *game.Game, common []types.Sub, cardID id.ID, first bool) []types.Sub {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return common
	}
	subtypes := card.Def.Subtypes
	if first {
		return slices.Clone(subtypes)
	}
	kept := make([]types.Sub, 0, len(common))
	for _, sub := range common {
		if slices.Contains(subtypes, sub) {
			kept = append(kept, sub)
		}
	}
	return kept
}

// removeFoundID returns ids without the first occurrence of target.
func removeFoundID(ids []id.ID, target id.ID) []id.ID {
	out := make([]id.ID, 0, len(ids))
	for _, cardID := range ids {
		if cardID != target {
			out = append(out, cardID)
		}
	}
	return out
}

func (e *Engine) exploreCreature(
	g *game.Game,
	obj *game.StackObject,
	agents [game.NumPlayers]PlayerAgent,
	log *TurnLog,
	playerID game.PlayerID,
	creature *game.Permanent,
) bool {
	player, ok := playerByID(g, playerID)
	if !ok || creature == nil {
		return false
	}
	cardID, ok := player.Library.Top()
	if !ok {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
	if slices.Contains(cardFaceOrDefault(card, game.FaceFront).Types, types.Land) {
		player.Library.Remove(cardID)
		player.Hand.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Hand,
			Amount:   1,
		})
		return true
	}

	addCountersToPermanentControlledBy(g, playerID, creature, counter.PlusOnePlusOne, 1)
	selected := e.chooseChoice(g, agents, libraryChoiceRequest(game.ChoiceExplore, playerID, "Explore: choose where to put revealed nonland card.", []string{"top", "graveyard"}), log)
	if len(selected) == 1 && selected[0] == 1 && player.Library.Remove(cardID) {
		putLibraryCardIntoGraveyard(g, playerID, cardID, 0)
	}
	return true
}

func peekLibrary(player *game.Player, amount int) []id.ID {
	if amount <= 0 {
		return nil
	}
	cards := player.Library.All()
	if amount > len(cards) {
		amount = len(cards)
	}
	return append([]id.ID(nil), cards[:amount]...)
}

// revealTopPartition reveals the top amount cards of playerID's library, puts
// every revealed card matching selection into that player's hand, and routes the
// rest to remainder (the player's graveyard or the bottom of their library). It
// backs the RevealTopPartition primitive: every revealed card is turned face up
// publicly and the matching cards are taken without a choice, so the partition
// is fully deterministic.
func revealTopPartition(g *game.Game, obj *game.StackObject, playerID game.PlayerID, amount int, selection game.Selection, remainder game.DigRemainder) bool {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return false
	}
	seen := peekLibrary(player, amount)
	if len(seen) == 0 {
		return false
	}
	for _, cardID := range seen {
		emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
	}
	for _, cardID := range seen {
		card, cardOK := g.GetCardInstance(cardID)
		if !cardOK || !player.Library.Remove(cardID) {
			continue
		}
		if cardMatchesSelection(g, obj, card, selection) {
			player.Hand.Add(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   zone.Hand,
				Amount:   1,
			})
			continue
		}
		if remainder == game.DigRemainderLibraryBottom {
			player.Library.AddToBottom(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   zone.Library,
				Amount:   1,
			})
			continue
		}
		putLibraryCardIntoGraveyard(g, playerID, cardID, 0)
	}
	return true
}

func reorderLibraryTop(player *game.Player, cards []id.ID) {
	if len(cards) == 0 {
		return
	}

	for _, cardID := range cards {
		player.Library.Remove(cardID)
	}
	for i := len(cards) - 1; i >= 0; i-- {
		player.Library.Add(cards[i])
	}
}

func libraryChoiceRequest(kind game.ChoiceKind, playerID game.PlayerID, prompt string, labels []string) game.ChoiceRequest {
	options := make([]game.ChoiceOption, 0, len(labels))
	for i, label := range labels {
		options = append(options, game.ChoiceOption{Index: i, Label: label})
	}
	return game.ChoiceRequest{
		Kind:             kind,
		Player:           playerID,
		Prompt:           prompt,
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
}
