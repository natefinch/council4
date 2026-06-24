package rules

import (
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// resolveChooseFromZone resolves a game.ChooseFromZone envelope: it gathers the
// candidates in env.SourceZone that match env.Filter (and the optional X-bound
// and linked-set restrictions), has env.Player choose a set respecting
// env.Quantity, env.Count, and env.Grouping, then moves each chosen card to its
// destination applying env.Riders.
//
// It is the single canonical resolver the historically separate zone-choice
// primitives (ReturnFromGraveyard, the typed library Search, ExileFromHand,
// PutFromHand, CastForFree, ...) will migrate onto. Each grouping drives the
// choice; the movement is shared. An unsupported envelope, an unresolvable
// player, or an empty candidate pool is a legal no-op (fail closed).
func (r *effectResolver) resolveChooseFromZone(env game.ChooseFromZone) effectResolved {
	res := effectResolved{accepted: true, amount: r.quantity(env.Quantity)}
	if !chooseFromZoneSupported(env) {
		return res
	}
	playerID, ok := r.resolvePlayer(env.Player)
	if !ok {
		return res
	}
	candidates := r.chooseFromZoneCandidates(env, playerID)
	if env.Riders.PublishLinked != "" {
		clearLinkedObjects(r.game, r.chooseFromZonePublishKey(env))
	}
	if len(candidates) == 0 {
		return res
	}
	switch env.Grouping {
	case game.ChooseSplitDestination:
		res.succeeded = r.chooseFromZoneSplit(env, playerID, candidates)
	default:
		chosen := r.chooseFromZoneChosen(env, playerID, candidates)
		res.amount = len(chosen)
		res.succeeded = r.chooseFromZoneMoveAll(env, playerID, chosen, env.Destination, env.Riders.EntersTapped)
	}
	return res
}

// chooseFromZoneSupported gates the envelope shapes the resolver implements,
// mirroring searchSpecSupported: an unsupported shape fails closed rather than
// silently mis-resolving. The split grouping needs both slots; the
// one-of-each-named-type grouping needs the named types; a face-down entry needs
// a real kind, a battlefield destination, and none of the entry riders it cannot
// compose with.
func chooseFromZoneSupported(env game.ChooseFromZone) bool {
	if env.SourceZone == zone.None {
		return false
	}
	if env.Grouping == game.ChooseSplitDestination && !env.SplitSecondary.Exists {
		return false
	}
	if env.Grouping == game.ChooseOneOfEachNamedType && len(env.Filter.RequiredTypesAny) == 0 {
		return false
	}
	if env.Riders.FaceDown {
		if env.Riders.FaceDownKind == game.FaceDownNone ||
			env.Destination.Zone != zone.Battlefield ||
			env.Riders.EntersTapped ||
			len(env.Riders.EntryCounters) > 0 {
			return false
		}
	}
	return true
}

// chooseFromZoneCandidates collects the cards in env.SourceZone that match the
// filter, narrowed by the optional FromLinked set and the X-bound mana-value
// rider. The candidate order follows the zone order so a fallback choice is
// deterministic. When env.AllOwners is set the pool spans every player's copy
// of the zone rather than only playerID's own.
func (r *effectResolver) chooseFromZoneCandidates(env game.ChooseFromZone, playerID game.PlayerID) []id.ID {
	var sourceCardIDs []id.ID
	if env.AllOwners {
		for _, player := range r.game.Players {
			zoneCards, ok := destinationZone(r.game, player.ID, env.SourceZone)
			if !ok {
				continue
			}
			sourceCardIDs = append(sourceCardIDs, zoneCards.All()...)
		}
	} else {
		source, ok := destinationZone(r.game, playerID, env.SourceZone)
		if !ok {
			return nil
		}
		sourceCardIDs = source.All()
	}
	var linkedFilter map[id.ID]bool
	if env.Riders.FromLinked != "" {
		refs := linkedObjects(r.game, r.chooseFromZoneLinkedKey(env.Riders.FromLinked))
		linkedFilter = make(map[id.ID]bool, len(refs))
		for _, ref := range refs {
			if ref.CardID != 0 {
				linkedFilter[ref.CardID] = true
			}
		}
	}
	maxManaValue := -1
	if env.Riders.MaxManaValueFromX {
		maxManaValue = r.obj.XValue
	}
	var candidates []id.ID
	for _, cardID := range sourceCardIDs {
		if linkedFilter != nil && !linkedFilter[cardID] {
			continue
		}
		card, cardOK := r.game.GetCardInstance(cardID)
		if !cardOK {
			continue
		}
		if maxManaValue >= 0 && card.Def.ManaValue() > maxManaValue {
			continue
		}
		if handCardMatchesSelection(r.game, card, env.Filter, playerID) {
			candidates = append(candidates, cardID)
		}
	}
	return candidates
}

// chooseFromZoneChosen runs the choice for every grouping except the split
// destination, returning the chosen card IDs in choice order.
func (r *effectResolver) chooseFromZoneChosen(env game.ChooseFromZone, playerID game.PlayerID, candidates []id.ID) []id.ID {
	switch env.Grouping {
	case game.ChooseSharedSubtype:
		return r.chooseFromZoneSharedSubtype(env, playerID, candidates)
	case game.ChooseOneOfEachNamedType:
		return r.chooseFromZoneOneOfEachNamedType(env, playerID, candidates)
	default:
		return r.chooseFromZoneAcrossSet(env, playerID, candidates)
	}
}

// chooseFromZoneAcrossSet runs the ordinary multi-card choice: the player picks
// between the minimum and maximum allowed cards from the whole matching pool.
func (r *effectResolver) chooseFromZoneAcrossSet(env game.ChooseFromZone, playerID game.PlayerID, candidates []id.ID) []id.ID {
	maxChoices, minChoices := r.chooseFromZoneBounds(env, len(candidates))
	if maxChoices <= 0 {
		return nil
	}
	options := chooseFromZoneOptions(r.game, candidates)
	defaultSelection := firstChoiceIndices(minChoices)
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:              game.ChoiceResolution,
		Player:            playerID,
		Prompt:            chooseFromZonePrompt(env),
		Options:           options,
		MinChoices:        minChoices,
		MaxChoices:        maxChoices,
		DefaultSelection:  defaultSelection,
		MaxTotalManaValue: env.Riders.MaxTotalManaValue,
	}, r.log)
	return chooseFromZoneResolveIndices(candidates, selected)
}

// chooseFromZoneSharedSubtype runs the staged "share a subtype" choice: each
// successive pick is offered only from cards that still share a subtype with
// every card already chosen, so an illegal correlated set cannot be assembled
// (CR 701.19).
func (r *effectResolver) chooseFromZoneSharedSubtype(env game.ChooseFromZone, playerID game.PlayerID, candidates []id.ID) []id.ID {
	maxChoices, _ := r.chooseFromZoneBounds(env, len(candidates))
	if maxChoices <= 0 {
		return nil
	}
	remaining := slices.Clone(candidates)
	var chosen []id.ID
	var common []types.Sub
	for len(chosen) < maxChoices {
		pool := make([]id.ID, 0, len(remaining))
		for _, cardID := range remaining {
			if len(chosen) == 0 || cardSharesAnySubtype(r.game, cardID, common) {
				pool = append(pool, cardID)
			}
		}
		if len(pool) == 0 {
			break
		}
		pick, ok := r.chooseFromZoneOptionalCard(playerID, pool, "Choose a card that shares a subtype")
		if !ok {
			break
		}
		chosen = append(chosen, pick)
		common = restrictSharedSubtypes(r.game, common, pick, len(chosen) == 1)
		remaining = removeFoundID(remaining, pick)
	}
	return chosen
}

// chooseFromZoneOneOfEachNamedType chooses at most one card for each card type
// listed in the filter's RequiredTypesAny, modeling "a creature card and a land
// card". A card already chosen for an earlier type is not offered again.
func (r *effectResolver) chooseFromZoneOneOfEachNamedType(env game.ChooseFromZone, playerID game.PlayerID, candidates []id.ID) []id.ID {
	chosenSet := make(map[id.ID]bool)
	var chosen []id.ID
	for _, cardType := range env.Filter.RequiredTypesAny {
		pool := make([]id.ID, 0, len(candidates))
		for _, cardID := range candidates {
			if chosenSet[cardID] {
				continue
			}
			card, ok := r.game.GetCardInstance(cardID)
			if !ok || !card.Def.HasType(cardType) {
				continue
			}
			pool = append(pool, cardID)
		}
		if len(pool) == 0 {
			continue
		}
		pick, ok := r.chooseFromZoneOptionalCard(playerID, pool, "Choose a "+string(cardType)+" card")
		if !ok {
			continue
		}
		chosenSet[pick] = true
		chosen = append(chosen, pick)
	}
	return chosen
}

// chooseFromZoneSplit resolves the split-destination grouping: the player
// chooses up to two cards, then distributes them across the primary slot
// (Destination, Riders.EntersTapped) and the secondary slot (SplitSecondary).
func (r *effectResolver) chooseFromZoneSplit(env game.ChooseFromZone, playerID game.PlayerID, candidates []id.ID) bool {
	maxChoices, minChoices := r.chooseFromZoneBounds(env, len(candidates))
	maxChoices = min(maxChoices, 2)
	if maxChoices <= 0 {
		return false
	}
	options := chooseFromZoneOptions(r.game, candidates)
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           chooseFromZonePrompt(env),
		Options:          options,
		MinChoices:       min(minChoices, maxChoices),
		MaxChoices:       maxChoices,
		DefaultSelection: firstChoiceIndices(min(minChoices, maxChoices)),
	}, r.log)
	chosen := chooseFromZoneResolveIndices(candidates, selected)
	secondary := env.SplitSecondary.Val
	switch len(chosen) {
	case 0:
		return false
	case 1:
		dest := env.Destination
		tapped := env.Riders.EntersTapped
		if r.chooseFromZoneSplitSlot(playerID, env, secondary) == 1 {
			dest = secondary.Destination
			tapped = secondary.EntersTapped
		}
		return r.chooseFromZoneMoveOne(env, playerID, chosen[0], dest, tapped)
	default:
		primaryIndex := r.chooseFromZoneSplitPrimaryCard(playerID, chosen)
		succeeded := false
		for i, cardID := range chosen {
			dest := secondary.Destination
			tapped := secondary.EntersTapped
			if i == primaryIndex {
				dest = env.Destination
				tapped = env.Riders.EntersTapped
			}
			if r.chooseFromZoneMoveOne(env, playerID, cardID, dest, tapped) {
				succeeded = true
			}
		}
		return succeeded
	}
}

// chooseFromZoneBounds computes the maximum and minimum number of cards the
// player chooses from the matching pool, honoring the count mode, the
// fail-to-find policy, and the optional total-mana-value cap.
func (r *effectResolver) chooseFromZoneBounds(env game.ChooseFromZone, poolSize int) (maxChoices, minChoices int) {
	switch env.Count {
	case game.ChooseAnyNumber:
		return poolSize, 0
	case game.ChooseUpTo:
		return min(r.quantity(env.Quantity), poolSize), 0
	default:
		maxChoices = min(r.quantity(env.Quantity), poolSize)
		minChoices = maxChoices
		// A ChooseExactly choice is mandatory by default: the player must choose
		// the full bound when able. A total-mana-value cap or an explicit
		// may-fail-to-find policy relaxes the minimum to zero so the empty choice
		// becomes legal (the qualified-tutor and "up to total mana value" forms).
		if env.Riders.MaxTotalManaValue.Exists || env.Riders.FailToFindPolicy == game.SearchMayFailToFind {
			minChoices = 0
		}
		return maxChoices, minChoices
	}
}

// chooseFromZoneMoveAll moves every chosen card to dest, applying the riders, and
// reports whether at least one card moved.
func (r *effectResolver) chooseFromZoneMoveAll(env game.ChooseFromZone, playerID game.PlayerID, chosen []id.ID, dest game.ChooseDestination, tapped bool) bool {
	if dest.Zone == zone.Battlefield && !env.Riders.FaceDown {
		return r.chooseFromZoneEnterBattlefield(env, playerID, chosen, tapped)
	}
	succeeded := false
	for _, cardID := range chosen {
		if r.chooseFromZoneMoveOne(env, playerID, cardID, dest, tapped) {
			succeeded = true
		}
	}
	return succeeded
}

// chooseFromZoneMoveOne moves a single chosen card to dest, applying the riders.
func (r *effectResolver) chooseFromZoneMoveOne(env game.ChooseFromZone, playerID game.PlayerID, cardID id.ID, dest game.ChooseDestination, tapped bool) bool {
	if dest.Zone == zone.Battlefield {
		return r.chooseFromZoneEnterBattlefield(env, playerID, []id.ID{cardID}, tapped)
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return false
	}
	if env.Riders.Reveal {
		emitCardRevealEvent(r.game, r.obj, card.Owner, cardID, env.SourceZone)
	}
	if !moveCardBetweenZonesWithPlacement(r.game, card.Owner, cardID, env.SourceZone, dest.Zone, env.Riders.DestinationBottom) {
		return false
	}
	r.chooseFromZonePublish(env, cardID)
	return true
}

// chooseFromZoneEnterBattlefield puts the chosen cards onto the battlefield at
// once, applying the tapped, owner-control, entry-counter, face-down, and reveal
// riders. The cards enter as a simultaneous batch so their enter-the-battlefield
// events and replacements resolve together.
func (r *effectResolver) chooseFromZoneEnterBattlefield(env game.ChooseFromZone, playerID game.PlayerID, chosen []id.ID, tapped bool) bool {
	if env.Riders.FaceDown {
		succeeded := false
		for _, cardID := range chosen {
			card, ok := r.game.GetCardInstance(cardID)
			if !ok {
				continue
			}
			controller := playerID
			if env.Riders.UnderOwnerControl {
				controller = card.Owner
			}
			if env.Riders.Reveal {
				emitCardRevealEvent(r.game, r.obj, card.Owner, cardID, env.SourceZone)
			}
			if _, placed := createCardPermanentFaceDownWithChoices(r.engine, r.game, card, controller, env.SourceZone, game.FaceFront, env.Riders.FaceDownKind, false, r.agents, r.log); placed {
				succeeded = true
				r.chooseFromZonePublish(env, cardID)
			}
		}
		return succeeded
	}
	resolved := make([]resolvedBattlefieldCard, 0, len(chosen))
	for _, cardID := range chosen {
		card, ok := r.game.GetCardInstance(cardID)
		if !ok {
			continue
		}
		controller := playerID
		if env.Riders.UnderOwnerControl {
			controller = card.Owner
		}
		if env.Riders.Reveal {
			emitCardRevealEvent(r.game, r.obj, card.Owner, cardID, env.SourceZone)
		}
		resolved = append(resolved, resolvedBattlefieldCard{
			card:       card,
			fromZone:   env.SourceZone,
			controller: controller,
		})
		r.chooseFromZonePublish(env, cardID)
	}
	return r.putResolvedCardsOnBattlefieldValue(resolved, nil, permanentCreationOptions{
		ForceTapped: tapped,
		Counters:    env.Riders.EntryCounters,
	})
}

// chooseFromZonePublish remembers a moved card under the PublishLinked key when
// one is set.
func (r *effectResolver) chooseFromZonePublish(env game.ChooseFromZone, cardID id.ID) {
	if env.Riders.PublishLinked == "" {
		return
	}
	rememberLinkedObject(r.game, r.chooseFromZonePublishKey(env), game.LinkedObjectRef{CardID: cardID})
}

// chooseFromZonePublishKey derives the key under which PublishLinked remembers
// chosen cards. PublishObjectScoped selects the source permanent's object
// identity (imprint scoping, Chrome Mox); otherwise the card-identity source key
// is used. FromLinked always uses the source key (chooseFromZoneLinkedKey).
func (r *effectResolver) chooseFromZonePublishKey(env game.ChooseFromZone) game.LinkedObjectKey {
	if env.Riders.PublishObjectScoped {
		return linkedObjectByObjectKey(r.game, r.obj, string(env.Riders.PublishLinked))
	}
	return r.chooseFromZoneLinkedKey(env.Riders.PublishLinked)
}

func (r *effectResolver) chooseFromZoneLinkedKey(key game.LinkedKey) game.LinkedObjectKey {
	return linkedObjectSourceKey(r.game, r.obj, string(key))
}

// chooseFromZoneOptionalCard offers one optional pick from pool, returning the
// chosen card and true, or ok=false when the player declines. A non-answering
// agent defaults to the first pool card.
func (r *effectResolver) chooseFromZoneOptionalCard(playerID game.PlayerID, pool []id.ID, prompt string) (id.ID, bool) {
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           prompt,
		Options:          chooseFromZoneOptions(r.game, pool),
		MinChoices:       0,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(pool) {
		return pool[selected[0]], true
	}
	return 0, false
}

// chooseFromZoneSplitSlot asks which slot the lone chosen card fills, returning
// 0 for the primary slot and 1 for the secondary slot.
func (r *effectResolver) chooseFromZoneSplitSlot(playerID game.PlayerID, env game.ChooseFromZone, secondary game.ChooseSplitSlot) int {
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose where to put the chosen card",
		Options:          chooseFromZoneSlotOptions(env.Destination, env.Riders.EntersTapped, secondary),
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) == 1 && selected[0] == 1 {
		return 1
	}
	return 0
}

// chooseFromZoneSplitPrimaryCard asks which of the two chosen cards enters the
// primary slot, returning its index into chosen; the other card fills the
// secondary slot.
func (r *effectResolver) chooseFromZoneSplitPrimaryCard(playerID game.PlayerID, chosen []id.ID) int {
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:             game.ChoiceResolution,
		Player:           playerID,
		Prompt:           "Choose which card goes to the primary destination",
		Options:          chooseFromZoneOptions(r.game, chosen),
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}, r.log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(chosen) {
		return selected[0]
	}
	return 0
}

// chooseFromZonePrompt returns the message for a ChooseFromZone's primary card
// choice: the envelope's explicit Prompt when set, otherwise the generic
// "Choose cards" default.
func chooseFromZonePrompt(env game.ChooseFromZone) string {
	if env.Prompt != "" {
		return env.Prompt
	}
	return "Choose cards"
}

func chooseFromZoneOptions(g *game.Game, candidates []id.ID) []game.ChoiceOption {
	options := make([]game.ChoiceOption, len(candidates))
	for i, cardID := range candidates {
		options[i] = game.ChoiceOption{
			Index: i,
			Label: cardChoiceLabel(g, cardID),
			Card:  cardChoiceInfo(g, cardID),
		}
	}
	return options
}

func chooseFromZoneSlotOptions(primary game.ChooseDestination, primaryTapped bool, secondary game.ChooseSplitSlot) []game.ChoiceOption {
	return []game.ChoiceOption{
		{Index: 0, Label: chooseFromZoneSlotLabel(primary, primaryTapped)},
		{Index: 1, Label: chooseFromZoneSlotLabel(secondary.Destination, secondary.EntersTapped)},
	}
}

func chooseFromZoneSlotLabel(dest game.ChooseDestination, tapped bool) string {
	if dest.Zone == zone.Battlefield && tapped {
		return zone.Battlefield.String() + " tapped"
	}
	return dest.Zone.String()
}

func chooseFromZoneResolveIndices(candidates []id.ID, selected []int) []id.ID {
	chosen := make([]id.ID, 0, len(selected))
	for _, index := range selected {
		if index >= 0 && index < len(candidates) {
			chosen = append(chosen, candidates[index])
		}
	}
	return chosen
}
