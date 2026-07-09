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
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			return milled
		}
		destinationCards.Add(cardID)
		if destination == zone.Graveyard {
			milled = append(milled, cardID)
		}
		emitZoneChangeEvent(g, game.Event{
			Player:         playerID,
			CardID:         cardID,
			FromZone:       zone.Library,
			ToZone:         destination,
			Amount:         1,
			SimultaneousID: batchID,
		})
	}
	return milled
}

// revealUntilCards reveals cards from the top of playerID's library one at a
// time until a revealed card matches until, then puts every card revealed this
// way (including the matching card) into destination. When the library empties
// before a match, every revealed card is still moved. destination must be
// zone.Graveyard or zone.Hand; a graveyard move honors the commander
// replacement (CR 903.9a).
func revealUntilCards(g *game.Game, playerID game.PlayerID, until game.Selection, destination zone.Type) {
	player, ok := playerByID(g, playerID)
	if !ok {
		return
	}
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			return
		}
		player.Library.Remove(cardID)
		matched := revealedCardMatches(g, playerID, cardID, until)
		dest := destination
		if destination == zone.Graveyard {
			dest = commanderReplacementDestination(g, cardID, zone.Graveyard)
		}
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); dest == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, dest)
		if !ok {
			return
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   dest,
			Amount:   1,
		})
		if matched {
			return
		}
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

func exileTopOfLibraryCards(g *game.Game, playerID game.PlayerID, amount int, counterKind opt.V[counter.Kind]) {
	player, ok := playerByID(g, playerID)
	if !ok || amount <= 0 {
		return
	}
	for range amount {
		cardID, ok := player.Library.Top()
		if !ok {
			return
		}
		player.Library.Remove(cardID)
		destination := commanderReplacementDestination(g, cardID, zone.Exile)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			return
		}
		destinationCards.Add(cardID)
		// Place the named exile counter only when the card actually landed in
		// exile: a CR 614/903.9 replacement or commander redirect can divert the
		// move to the command zone, and gating on the intended destination would
		// orphan a counter on a card that never reached exile.
		if counterKind.Exists && destination == zone.Exile {
			g.AddExileCounter(cardID, counterKind.Val, 1)
		}
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
	}
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
			destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
			zoneOwner := playerID
			if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
				zoneOwner = card.Owner
			}
			destinationCards, ok := destinationZone(g, zoneOwner, destination)
			if !ok {
				continue
			}
			destinationCards.Add(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Library,
				ToZone:   destination,
				Amount:   1,
			})
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
// when filter.selection is present, by the cards matching it) to take to the
// filter.destination (their hand by default, or the battlefield), and the
// remaining cards go to the destination identified by remainder (graveyard or
// the bottom of the library, in seen order). When filter.takeUpTo is set the
// controller may take fewer than take cards (down to none); when filter.reveal
// is set each taken card is revealed as it is taken; when filter.destination is
// the battlefield each taken card is put onto the battlefield under the player's
// control, tapped when filter.entersTapped is set.
func (e *Engine) digCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, obj *game.StackObject, playerID game.PlayerID, look, take int, remainder game.DigRemainder, filter digFilter) bool {
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
		taken = e.chooseDigCards(g, agents, log, playerID, eligible, minTake, take)
	}
	for _, cardID := range taken {
		if !player.Library.Remove(cardID) {
			continue
		}
		if filter.reveal {
			emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
		}
		if filter.destination == zone.Battlefield {
			card, cardOK := g.GetCardInstance(cardID)
			if !cardOK {
				continue
			}
			_, _ = createCardPermanentFaceWithOptions(e, g, card, playerID, zone.Library, game.FaceFront, nil, permanentCreationOptions{ForceTapped: filter.entersTapped}, agents, log)
			continue
		}
		player.Hand.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Hand,
			Amount:   1,
		})
	}
	for _, cardID := range seen {
		if slices.Contains(taken, cardID) {
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
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			continue
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
	}
	return true
}

// chooseDigCards asks the digging player which of the eligible cards (already
// filtered to those that may be taken) to put into their hand. minTake and
// maxTake bound the selection: an exact dig passes minTake == maxTake, while a
// typed "you may reveal up to N" dig passes minTake 0 so the player may decline.
// Agents that do not answer fall back to the deterministic first-take selection,
// preserving prior engine behavior.
func (e *Engine) chooseDigCards(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, eligible []id.ID, minTake, maxTake int) []id.ID {
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
		Prompt:           "Dig: choose cards to put into your hand.",
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
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			continue
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
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

// chooseSearchMatches asks the searching player which matching library cards to
// take. minChoices is zero for qualified or "up to" searches and one for an
// unrestricted exact-card search with a nonempty library.
func (e *Engine) chooseSearchMatches(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, candidates []id.ID, amount, minChoices int) []id.ID {
	if len(candidates) == 0 || amount <= 0 {
		return nil
	}
	selected := e.chooseChoice(g, agents, searchChoiceRequest(g, playerID, candidates, amount, minChoices), log)
	found := make([]id.ID, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(candidates) {
			found = append(found, candidates[index])
		}
	}
	return found
}

func searchChoiceRequest(g *game.Game, playerID game.PlayerID, candidates []id.ID, amount, minChoices int) game.ChoiceRequest {
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
		Player:           playerID,
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
func (e *Engine) chooseCorrelatedSearchMatches(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, candidates []id.ID, amount int) []id.ID {
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
		pick, ok := e.chooseCorrelatedSearchCard(g, agents, log, playerID, pool)
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
func (e *Engine) chooseDifferentNameSearchMatches(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, candidates []id.ID, amount int) []id.ID {
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
		pick, ok := e.chooseCorrelatedSearchCard(g, agents, log, playerID, pool)
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
func (e *Engine) chooseCorrelatedSearchCard(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, pool []id.ID) (id.ID, bool) {
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
		Player:           playerID,
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
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if card, ok := g.GetCardInstance(cardID); destination == zone.Command && ok {
			zoneOwner = card.Owner
		}
		destinationCards, ok := destinationZone(g, zoneOwner, destination)
		if !ok {
			return true
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
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
		destination := commanderReplacementDestination(g, cardID, zone.Graveyard)
		zoneOwner := playerID
		if destination == zone.Command {
			zoneOwner = card.Owner
		}
		destinationCards, zoneOK := destinationZone(g, zoneOwner, destination)
		if !zoneOK {
			continue
		}
		destinationCards.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   destination,
			Amount:   1,
		})
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
