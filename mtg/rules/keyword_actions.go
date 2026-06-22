package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func resolveFightTargets(g *game.Game, obj *game.StackObject, firstIndex, secondIndex int) {
	first, firstOK := effectPermanentTarget(g, obj, firstIndex)
	second, secondOK := effectPermanentTarget(g, obj, secondIndex)
	if !firstOK || !secondOK || first.ObjectID == second.ObjectID || !permanentHasType(g, first, types.Creature) || !permanentHasType(g, second, types.Creature) {
		return
	}
	resolveFightPermanents(g, first, second)
}

func resolveFightPermanents(g *game.Game, first, second *game.Permanent) {
	if first == nil || second == nil || first.ObjectID == second.ObjectID || !permanentHasType(g, first, types.Creature) || !permanentHasType(g, second, types.Creature) {
		return
	}
	simultaneousID := g.IDGen.Next()
	emitFightEvent(g, first, second, simultaneousID)
	emitFightEvent(g, second, first, simultaneousID)
	dealPermanentDamage(g, first.CardInstanceID, first.ObjectID, effectiveController(g, first), second, effectivePower(g, first), false)
	dealPermanentDamage(g, second.CardInstanceID, second.ObjectID, effectiveController(g, second), first, effectivePower(g, second), false)
}

func effectPermanentTarget(g *game.Game, obj *game.StackObject, targetIndex int) (*game.Permanent, bool) {
	if targetIndex < 0 || targetIndex >= len(obj.Targets) {
		return nil, false
	}
	target := obj.Targets[targetIndex]
	if target.Kind != game.TargetPermanent || target.PermanentID == 0 {
		return nil, false
	}
	return permanentByObjectID(g, target.PermanentID)
}

func emitFightEvent(g *game.Game, permanent, related *game.Permanent, simultaneousID id.ID) {
	emitEvent(g, game.Event{
		Kind:               game.EventFight,
		SourceID:           permanent.CardInstanceID,
		SourceObjectID:     permanent.ObjectID,
		Controller:         effectiveController(g, permanent),
		PermanentID:        permanent.ObjectID,
		RelatedPermanentID: related.ObjectID,
		SimultaneousID:     simultaneousID,
	})
}

func counterTargetStackObject(g *game.Game, obj *game.StackObject, targetIndex int, exileInstead bool) bool {
	stackObjectID, ok := effectStackObjectID(obj, targetIndex)
	if !ok {
		return false
	}
	target, ok := stackObjectByID(g, stackObjectID)
	if !ok {
		return false
	}
	if exileInstead {
		target.ExileOnResolution = true
	}
	if obj.TargetControllerLKI == nil {
		obj.TargetControllerLKI = make(map[int]game.PlayerID)
	}
	obj.TargetControllerLKI[targetIndex] = target.Controller
	if manaValue, known := stackObjectManaValue(g, target); known {
		if obj.TargetManaValueLKI == nil {
			obj.TargetManaValueLKI = make(map[int]int)
		}
		obj.TargetManaValueLKI[targetIndex] = manaValue
	}
	return counterStackObject(g, stackObjectID)
}

func stackObjectManaValue(g *game.Game, obj *game.StackObject) (int, bool) {
	if obj == nil || obj.Kind != game.StackSpell {
		return 0, false
	}
	if obj.FaceDown {
		return 0, true
	}
	if obj.SourceTokenDef != nil {
		face, ok := obj.SourceTokenDef.FaceDef(obj.Face)
		if !ok {
			return 0, false
		}
		return stackManaValue(face, obj.XValue), true
	}
	card, ok := g.GetCardInstance(obj.SourceID)
	if !ok && obj.SourceCardID != 0 {
		card, ok = g.GetCardInstance(obj.SourceCardID)
	}
	if !ok {
		return 0, false
	}
	return stackManaValue(cardFaceOrDefault(card, obj.Face), obj.XValue), true
}

func effectStackObjectID(obj *game.StackObject, targetIndex int) (id.ID, bool) {
	if targetIndex < 0 || targetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[targetIndex]
	if target.Kind != game.TargetStackObject || target.StackObjectID == 0 {
		return 0, false
	}
	return target.StackObjectID, true
}

func discardCards(g *game.Game, playerID game.PlayerID, amount int) bool {
	if amount <= 0 {
		amount = 1
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	discarded := false
	for range amount {
		cardID, ok := player.Hand.Top()
		if !ok {
			return discarded
		}
		if !discardCardFromHand(g, playerID, cardID) {
			return discarded
		}
		discarded = true
	}
	return discarded
}

// discardEntireHand discards every card in a player's hand as one simultaneous
// batch and returns the number of cards discarded.
func discardEntireHand(g *game.Game, playerID game.PlayerID) int {
	player, ok := playerByID(g, playerID)
	if !ok {
		return 0
	}
	cards := slices.Clone(player.Hand.All())
	if len(cards) == 0 {
		return 0
	}
	simultaneousID := g.IDGen.Next()
	discarded := 0
	for _, cardID := range cards {
		if discardCardFromHandInBatch(g, playerID, cardID, simultaneousID) {
			discarded++
		}
	}
	return discarded
}

func searchSpecSupported(spec game.SearchSpec) bool {
	primary := game.SearchDestination{
		Zone:         spec.Destination,
		Position:     spec.DestinationPosition,
		EntersTapped: spec.EntersTapped,
	}
	if spec.SourceZone != zone.Library || !searchDestinationSupported(primary) {
		return false
	}
	if spec.SplitDestination.Exists &&
		(!searchDestinationSupported(spec.SplitDestination.Val) ||
			spec.SplitDestination.Val.Zone == zone.Library) {
		return false
	}
	return true
}

func searchDestinationSupported(destination game.SearchDestination) bool {
	switch destination.Zone {
	case zone.Hand, zone.Graveyard:
		return destination.Position == game.SearchPositionUnspecified && !destination.EntersTapped
	case zone.Battlefield:
		return destination.Position == game.SearchPositionUnspecified
	case zone.Library:
		return destination.Position == game.SearchPositionTop && !destination.EntersTapped
	default:
		return false
	}
}

func (e *Engine) searchLibrary(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID, controllerID game.PlayerID, spec game.SearchSpec, amount int) (bool, *game.Permanent) {
	if amount <= 0 {
		amount = 1
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return false, nil
	}
	// A player searches their library when this runs regardless of whether a
	// matching card is found (CR 701.19a), so the search event fires once here
	// for "whenever a player searches their library" triggers.
	emitEvent(g, game.Event{
		Kind:       game.EventLibrarySearched,
		Controller: playerID,
		Player:     playerID,
	})
	if spec.MaxManaValueFromX {
		// "with mana value X or less" bounds the search by the spell's chosen X,
		// resolved from the resolving stack object as the search runs.
		spec.Filter.ManaValue = opt.Val(compare.Int{Op: compare.LessOrEqual, Value: obj.XValue})
		spec.MaxManaValueFromX = false
	}
	var candidates []id.ID
	for _, cardID := range player.Library.All() {
		if searchSpecMatches(g, obj, cardID, spec) {
			candidates = append(candidates, cardID)
		}
	}
	// The searching player chooses which matching cards to take. Qualified
	// searches may legally fail to find even when matches exist (CR 701.19e);
	// unrestricted exact-card searches must find one when the library is nonempty.
	// A correlated search ("that share a land type") chooses cards through a
	// staged dependent choice that only offers cards still able to share a subtype
	// with those already chosen, so an illegal combination can never be assembled.
	var found []id.ID
	if spec.SharedSubtype {
		found = e.chooseCorrelatedSearchMatches(g, agents, log, playerID, candidates, amount)
	} else {
		minChoices := 0
		if searchMustFindIfAvailable(spec, amount) {
			minChoices = 1
		}
		found = e.chooseSearchMatches(g, agents, log, playerID, candidates, amount, minChoices)
	}
	if spec.SplitDestination.Exists {
		return e.placeSplitSearch(g, obj, agents, log, playerID, controllerID, player, spec, found), nil
	}
	primary := game.SearchDestination{
		Zone:         spec.Destination,
		Position:     spec.DestinationPosition,
		EntersTapped: spec.EntersTapped,
	}
	if primary.Zone == zone.Library {
		if amount != 1 {
			return false, nil
		}
		if len(found) == 0 {
			player.Library.Shuffle(e.rng)
			return false, nil
		}
		cardID := found[0]
		if !player.Library.Remove(cardID) {
			return false, nil
		}
		if spec.Reveal {
			emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
		}
		player.Library.Shuffle(e.rng)
		_, placed := e.placeFoundCard(g, obj, playerID, controllerID, player, cardID, primary)
		return placed, nil
	}
	var foundPermanent *game.Permanent
	for _, cardID := range found {
		if !player.Library.Remove(cardID) {
			return len(found) > 0, foundPermanent
		}
		if spec.Reveal {
			emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
		}
		permanent, placed := e.placeFoundCard(g, obj, playerID, controllerID, player, cardID, primary)
		if !placed {
			return len(found) > 0, foundPermanent
		}
		if permanent != nil {
			foundPermanent = permanent
		}
	}
	player.Library.Shuffle(e.rng)
	return len(found) > 0, foundPermanent
}

func searchMustFindIfAvailable(spec game.SearchSpec, amount int) bool {
	switch spec.FailToFindPolicy {
	case game.SearchMustFindIfAvailable:
		return true
	case game.SearchMayFailToFind:
		return false
	default:
		return amount == 1 && spec.IsUnrestricted()
	}
}

// placeFoundCard moves a found library card into a single-card search
// destination slot, emitting the library-to-zone change event. The card must
// already be removed from the library. A found card put onto the battlefield
// enters under controllerID's control (which equals playerID unless a search
// names a different controller, e.g. "under target player's control"); other
// destinations always go to the searching player. It returns the created
// permanent for a battlefield destination and false if placement fails.
func (e *Engine) placeFoundCard(g *game.Game, obj *game.StackObject, playerID, controllerID game.PlayerID, player *game.Player, cardID id.ID, dest game.SearchDestination) (*game.Permanent, bool) {
	switch dest.Zone {
	case zone.Hand:
		player.Hand.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			SourceID:      stackObjectSourceID(obj),
			StackObjectID: stackObjectID(obj),
			Controller:    stackObjectController(obj),
			Player:        playerID,
			CardID:        cardID,
			FromZone:      zone.Library,
			ToZone:        zone.Hand,
			Amount:        1,
		})
		return nil, true
	case zone.Graveyard:
		player.Graveyard.Add(cardID)
		emitZoneChangeEvent(g, game.Event{
			SourceID:      stackObjectSourceID(obj),
			StackObjectID: stackObjectID(obj),
			Controller:    stackObjectController(obj),
			Player:        playerID,
			CardID:        cardID,
			FromZone:      zone.Library,
			ToZone:        zone.Graveyard,
			Amount:        1,
		})
		return nil, true
	case zone.Battlefield:
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			return nil, false
		}
		return createCardPermanentFaceWithOptions(e, g, card, controllerID, zone.Library, game.FaceFront, nil, permanentCreationOptions{ForceTapped: dest.EntersTapped}, [game.NumPlayers]PlayerAgent{}, nil)
	case zone.Library:
		if dest.Position != game.SearchPositionTop {
			return nil, false
		}
		player.Library.Add(cardID)
		return nil, true
	default:
		return nil, false
	}
}

// placeSplitSearch resolves a split-destination library search (Cultivate,
// Kodama's Reach). It reveals the found cards, then distributes them across the
// two single-card slots: the primary slot is (spec.Destination,
// spec.EntersTapped) and the secondary slot is spec.SplitDestination. With two
// cards found the searching player assigns one card to each slot; with one card
// found the searching player chooses which slot it fills (CR 701.19). It always
// shuffles afterward and returns whether any card was found.
func (e *Engine) placeSplitSearch(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID, controllerID game.PlayerID, player *game.Player, spec game.SearchSpec, found []id.ID) bool {
	primary := game.SearchDestination{Zone: spec.Destination, EntersTapped: spec.EntersTapped}
	secondary := spec.SplitDestination.Val
	if spec.Reveal {
		for _, cardID := range found {
			emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
		}
	}
	switch len(found) {
	case 0:
		player.Library.Shuffle(e.rng)
		return false
	case 1:
		dest := primary
		if e.chooseSplitSearchSlot(g, agents, log, playerID, primary, secondary) == 1 {
			dest = secondary
		}
		if player.Library.Remove(found[0]) {
			_, _ = e.placeFoundCard(g, obj, playerID, controllerID, player, found[0], dest)
		}
	default:
		primaryCard := found[e.chooseSplitSearchPrimaryCard(g, agents, log, playerID, primary, found)]
		for _, cardID := range found {
			dest := secondary
			if cardID == primaryCard {
				dest = primary
			}
			if player.Library.Remove(cardID) {
				_, _ = e.placeFoundCard(g, obj, playerID, controllerID, player, cardID, dest)
			}
		}
	}
	player.Library.Shuffle(e.rng)
	return len(found) > 0
}

// chooseSplitSearchSlot asks the searching player which slot the lone found card
// fills when a split-destination search finds only one card. It returns 0 for
// the primary slot and 1 for the secondary slot, defaulting to the primary slot
// for agents that do not answer.
func (e *Engine) chooseSplitSearchSlot(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, primary, secondary game.SearchDestination) int {
	request := libraryChoiceRequest(
		game.ChoiceSearch,
		playerID,
		"Split search: choose where to put the found card.",
		[]string{searchDestinationLabel(primary), searchDestinationLabel(secondary)},
	)
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) == 1 && selected[0] == 1 {
		return 1
	}
	return 0
}

// chooseSplitSearchPrimaryCard asks the searching player which of the two found
// cards enters the primary slot; the other card fills the secondary slot. The
// prompt names the primary destination so it stays accurate for hand-first
// wordings as well as the usual battlefield-first ones. It returns the index
// into found, defaulting to the first card for agents that do not answer.
func (e *Engine) chooseSplitSearchPrimaryCard(g *game.Game, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, primary game.SearchDestination, found []id.ID) int {
	options := make([]game.ChoiceOption, 0, len(found))
	for i, cardID := range found {
		label := "unknown card"
		if card, ok := g.GetCardInstance(cardID); ok {
			label = cardFaceOrDefault(card, game.FaceFront).Name
		}
		options = append(options, game.ChoiceOption{Index: i, Label: label})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoiceSearch,
		Player:           playerID,
		Prompt:           "Split search: choose which card goes to " + searchDestinationLabel(primary) + ".",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(found) {
		return selected[0]
	}
	return 0
}

// searchDestinationLabel renders a split-search slot for a choice prompt.
func searchDestinationLabel(dest game.SearchDestination) string {
	switch dest.Zone {
	case zone.Battlefield:
		if dest.EntersTapped {
			return "battlefield tapped"
		}
		return "battlefield"
	case zone.Hand:
		return "hand"
	default:
		return "unknown zone"
	}
}

func searchSpecMatches(g *game.Game, obj *game.StackObject, cardID id.ID, spec game.SearchSpec) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	if spec.Name != "" && card.Def.Name != spec.Name {
		return false
	}
	return cardMatchesSelection(g, obj, card, spec.Filter)
}

func revealCards(g *game.Game, obj *game.StackObject, playerID game.PlayerID, zoneType zone.Type, amount int) bool {
	return len(revealCardIDs(g, obj, playerID, zoneType, amount)) > 0
}

func revealCardIDs(g *game.Game, obj *game.StackObject, playerID game.PlayerID, zoneType zone.Type, amount int) []id.ID {
	if amount <= 0 {
		amount = 1
	}
	player, ok := playerByID(g, playerID)
	if !ok || zoneType != zone.Library {
		return nil
	}
	var revealed []id.ID
	for i, cardID := range player.Library.All() {
		if i >= amount {
			break
		}
		emitCardRevealEvent(g, obj, playerID, cardID, zoneType)
		revealed = append(revealed, cardID)
	}
	return revealed
}

func emitCardRevealEvent(g *game.Game, obj *game.StackObject, playerID game.PlayerID, cardID id.ID, zoneType zone.Type) {
	emitEvent(g, game.Event{
		Kind:          game.EventCardRevealed,
		SourceID:      stackObjectSourceID(obj),
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        playerID,
		CardID:        cardID,
		FromZone:      zoneType,
		Amount:        1,
	})
}

func clueTokenDef() *game.CardDef {
	two := cost.Mana{cost.O(2)}
	additionalCosts := []cost.Additional{{
		Kind:               cost.AdditionalSacrificeSource,
		Text:               "Sacrifice this artifact",
		Amount:             1,
		MatchPermanentType: true,
		PermanentType:      types.Artifact,
	}}
	drawContent := game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		},
	}.Ability()

	return &game.CardDef{CardFace: game.CardFace{Name: "Clue Token",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Clue},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:            "{2}, Sacrifice this artifact: Draw a card.",
			ManaCost:        opt.Val(two),
			AdditionalCosts: additionalCosts,
			Content:         drawContent,
		}}},
	}
}
