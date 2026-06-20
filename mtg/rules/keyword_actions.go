package rules

import (
	"github.com/natefinch/council4/mtg/game"
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

func counterTargetStackObject(g *game.Game, obj *game.StackObject, targetIndex int) bool {
	stackObjectID, ok := effectStackObjectID(obj, targetIndex)
	if !ok {
		return false
	}
	target, ok := stackObjectByID(g, stackObjectID)
	if !ok {
		return false
	}
	if obj.TargetControllerLKI == nil {
		obj.TargetControllerLKI = make(map[int]game.PlayerID)
	}
	obj.TargetControllerLKI[targetIndex] = target.Controller
	return counterStackObject(g, stackObjectID)
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

func searchSpecSupported(spec game.SearchSpec) bool {
	if spec.SourceZone != zone.Library || !searchDestinationZoneSupported(spec.Destination) {
		return false
	}
	if spec.SplitDestination.Exists && !searchDestinationZoneSupported(spec.SplitDestination.Val.Zone) {
		return false
	}
	return true
}

// searchDestinationZoneSupported reports whether a library search may send a
// found card to the zone. The runtime models hand and battlefield destinations.
func searchDestinationZoneSupported(z zone.Type) bool {
	return z == zone.Hand || z == zone.Battlefield
}

func (e *Engine) searchLibrary(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, spec game.SearchSpec, amount int) (bool, *game.Permanent) {
	if amount <= 0 {
		amount = 1
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return false, nil
	}
	var candidates []id.ID
	for _, cardID := range player.Library.All() {
		if searchSpecMatches(g, cardID, spec) {
			candidates = append(candidates, cardID)
		}
	}
	// The searching player chooses which matching cards to take and may legally
	// fail to find even when matches exist (CR 701.19e). A correlated search
	// ("that share a land type") chooses cards through a staged dependent choice
	// that only offers cards still able to share a subtype with those already
	// chosen, so an illegal combination can never be assembled.
	var found []id.ID
	if spec.SharedSubtype {
		found = e.chooseCorrelatedSearchMatches(g, agents, log, playerID, candidates, amount)
	} else {
		found = e.chooseSearchMatches(g, agents, log, playerID, candidates, amount)
	}
	if spec.SplitDestination.Exists {
		return e.placeSplitSearch(g, obj, agents, log, playerID, player, spec, found), nil
	}
	primary := game.SearchDestination{Zone: spec.Destination, EntersTapped: spec.EntersTapped}
	var foundPermanent *game.Permanent
	for _, cardID := range found {
		if !player.Library.Remove(cardID) {
			return len(found) > 0, foundPermanent
		}
		if spec.Reveal {
			emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
		}
		permanent, placed := e.placeFoundCard(g, obj, playerID, player, cardID, primary)
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

// placeFoundCard moves a found library card into a single-card search
// destination slot, emitting the library-to-zone change event. The card must
// already be removed from the library. It returns the created permanent for a
// battlefield destination and false if placement fails.
func (e *Engine) placeFoundCard(g *game.Game, obj *game.StackObject, playerID game.PlayerID, player *game.Player, cardID id.ID, dest game.SearchDestination) (*game.Permanent, bool) {
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
	case zone.Battlefield:
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			return nil, false
		}
		return createCardPermanentFaceWithOptions(e, g, card, playerID, zone.Library, game.FaceFront, nil, permanentCreationOptions{ForceTapped: dest.EntersTapped}, [game.NumPlayers]PlayerAgent{}, nil)
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
func (e *Engine) placeSplitSearch(g *game.Game, obj *game.StackObject, agents [game.NumPlayers]PlayerAgent, log *TurnLog, playerID game.PlayerID, player *game.Player, spec game.SearchSpec, found []id.ID) bool {
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
			_, _ = e.placeFoundCard(g, obj, playerID, player, found[0], dest)
		}
	default:
		primaryCard := found[e.chooseSplitSearchPrimaryCard(g, agents, log, playerID, primary, found)]
		for _, cardID := range found {
			dest := secondary
			if cardID == primaryCard {
				dest = primary
			}
			if player.Library.Remove(cardID) {
				_, _ = e.placeFoundCard(g, obj, playerID, player, cardID, dest)
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

func searchSpecMatches(g *game.Game, cardID id.ID, spec game.SearchSpec) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	if spec.CardType.Exists && !card.Def.HasType(spec.CardType.Val) {
		return false
	}
	if spec.Permanent && !card.Def.IsPermanent() {
		return false
	}
	if spec.Supertype.Exists && !card.Def.HasSupertype(spec.Supertype.Val) {
		return false
	}
	if len(spec.SubtypesAny) > 0 && !card.Def.HasAnySubtype(spec.SubtypesAny...) {
		return false
	}
	if spec.MaxManaValue.Exists && card.Def.ManaValue() > spec.MaxManaValue.Val {
		return false
	}
	return true
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
