package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// handleRevealPutOntoBattlefield resolves RevealPutOntoBattlefield: it reveals
// the top cards, lets the player put one matching card onto the battlefield with
// counters and a granted keyword, then shuffles. It backs Undercity's Throne of
// the Dead Three.
func handleRevealPutOntoBattlefield(r *effectResolver, prim game.RevealPutOntoBattlefield) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	look := r.quantity(prim.Look)
	if look <= 0 {
		return res
	}
	seen := peekLibrary(player, look)
	if len(seen) == 0 {
		return res
	}
	for _, cardID := range seen {
		emitCardRevealEvent(r.game, r.obj, playerID, cardID, zone.Library)
	}
	var eligible []id.ID
	for _, cardID := range seen {
		card, cardOK := r.game.GetCardInstance(cardID)
		if cardOK && cardMatchesSelection(r.game, r.obj, card, prim.Selection) {
			eligible = append(eligible, cardID)
		}
	}
	if len(eligible) > 0 {
		chosen := r.engine.chooseDigCards(r.game, r.agents, r.log, playerID, eligible, 1, 1, zone.Battlefield)
		if len(chosen) > 0 {
			r.putRevealedCardOntoBattlefield(playerID, player, chosen[0], prim)
			res.succeeded = true
		}
	}
	if prim.Shuffle {
		player.Library.Shuffle(r.engine.rng)
	}
	return res
}

// putRevealedCardOntoBattlefield removes cardID from the player's library, puts
// it onto the battlefield, and applies the counters and granted keyword.
func (r *effectResolver) putRevealedCardOntoBattlefield(playerID game.PlayerID, player *game.Player, cardID id.ID, prim game.RevealPutOntoBattlefield) {
	if !player.Library.Remove(cardID) {
		return
	}
	card, ok := r.game.GetCardInstance(cardID)
	if !ok {
		return
	}
	permanent, created := createCardPermanentFaceWithOptions(r.engine, r.game, card, playerID, zone.Library, game.FaceFront, nil, permanentCreationOptions{}, r.agents, r.log)
	if !created || permanent == nil {
		return
	}
	if counters := r.quantity(prim.Counters); counters > 0 {
		addCountersToPermanent(r.game, permanent, prim.CounterKind, counters)
	}
	if prim.GrantKeyword.Exists {
		applyTypedContinuousEffects(r.game, r.obj, permanent, []game.ContinuousEffect{{
			Layer:       game.LayerAbility,
			AddKeywords: []game.Keyword{prim.GrantKeyword.Val},
		}}, prim.KeywordDuration)
	}
}

// handleCastLinkedCardForFree resolves CastLinkedCardForFree: the controller may
// cast one card from the linked group that is still in their hand and castable,
// without paying its mana cost. It backs Mad Wizard's Lair.
func handleCastLinkedCardForFree(r *effectResolver, prim game.CastLinkedCardForFree) effectResolved {
	res := effectResolved{accepted: true}
	playerID, ok := r.resolvePlayer(prim.Player)
	if !ok {
		return res
	}
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	key := linkedObjectSourceKey(r.game, r.obj, string(prim.LinkID))
	var castable []id.ID
	for _, ref := range linkedObjects(r.game, key) {
		if ref.CardID == 0 || !player.Hand.Contains(ref.CardID) {
			continue
		}
		// Reveal each drawn card still in hand ("Draw three cards and reveal
		// them."), then offer the castable ones for the free cast.
		emitCardRevealEvent(r.game, r.obj, playerID, ref.CardID, zone.Hand)
		if castableHandSpell(r.game, playerID, ref.CardID) {
			castable = append(castable, ref.CardID)
		}
	}
	clearLinkedObjects(r.game, key)
	if len(castable) == 0 {
		return res
	}
	options := make([]game.ChoiceOption, len(castable))
	for i, cardID := range castable {
		options[i] = game.ChoiceOption{Index: i, Label: cardChoiceLabel(r.game, cardID), Card: cardChoiceInfo(r.game, cardID)}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:       game.ChoiceResolution,
		Player:     playerID,
		Prompt:     "Choose a card to cast without paying its mana cost, or none",
		Options:    options,
		MinChoices: 0,
		MaxChoices: 1,
	}, r.log)
	if len(selected) == 0 || selected[0] < 0 || selected[0] >= len(castable) {
		return res
	}
	res.succeeded = r.engine.castFreeTargetedSpell(r.game, playerID, castable[selected[0]], zone.Hand, false, r.agents, r.log)
	return res
}

// castableHandSpell reports whether cardID rests in playerID's hand and has a
// legal free-cast choice, so lands and uncastable cards are skipped.
func castableHandSpell(g *game.Game, playerID game.PlayerID, cardID id.ID) bool {
	if _, ok := playerHoldingCastSource(g, cardID, zone.Hand); !ok {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	_, _, legal := firstLegalSpellCastChoice(g, playerID, cardFaceOrDefault(card, game.FaceFront))
	return legal
}

// handleRollDiceCreateTokens resolves RollDiceCreateTokens: it rolls the dice,
// sums the results, and creates that many tokens for the controller. It backs
// Baldur's Gate Wilderness's Reithwin Tollhouse.
func handleRollDiceCreateTokens(r *effectResolver, prim game.RollDiceCreateTokens) effectResolved {
	res := effectResolved{accepted: true}
	playerID := r.obj.Controller
	token, ok := r.typedTokenDefinition(prim.Source)
	if !ok || prim.Dice <= 0 || prim.Sides <= 0 {
		return res
	}
	total := 0
	for range prim.Dice {
		total += r.engine.rng.IntN(prim.Sides) + 1
	}
	res.amount = total
	if total <= 0 {
		return res
	}
	_, created := createTokenPermanentsCollectingWithChoices(r.engine, r.game, playerID, token, total, false, r.agents, r.log)
	res.succeeded = created
	return res
}

// handleRevealToHandDrainManaValue resolves RevealToHandDrainManaValue: it
// reveals the top cards, puts them into the controller's hand, then makes each
// opponent lose life equal to those cards' total mana value. It backs Baldur's
// Gate Wilderness's Ansur's Sanctum.
func handleRevealToHandDrainManaValue(r *effectResolver, prim game.RevealToHandDrainManaValue) effectResolved {
	res := effectResolved{accepted: true}
	playerID := r.obj.Controller
	player, ok := playerByID(r.game, playerID)
	if !ok {
		return res
	}
	amount := r.quantity(prim.Amount)
	if amount <= 0 {
		return res
	}
	seen := peekLibrary(player, amount)
	totalManaValue := 0
	for _, cardID := range seen {
		if !player.Library.Remove(cardID) {
			continue
		}
		if card, cardOK := r.game.GetCardInstance(cardID); cardOK {
			emitCardRevealEvent(r.game, r.obj, playerID, cardID, zone.Library)
			totalManaValue += cardFaceOrDefault(card, game.FaceFront).ManaValue()
		}
		player.Hand.Add(cardID)
		emitZoneChangeEvent(r.game, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Hand,
			Amount:   1,
		})
	}
	res.succeeded = len(seen) > 0
	if totalManaValue <= 0 {
		return res
	}
	for _, opponentID := range aliveOpponents(r.game, playerID) {
		loseLife(r.game, opponentID, totalManaValue)
	}
	return res
}

// handleGoadForEachOpponent resolves GoadForEachOpponent: for each opponent, the
// controller chooses up to one creature that opponent controls and goads it. It
// backs Baldur's Gate Wilderness's Grymforge. It is modeled as a resolution-time
// choice per opponent rather than acquiring one target per opponent.
func handleGoadForEachOpponent(r *effectResolver, _ game.GoadForEachOpponent) effectResolved {
	res := effectResolved{accepted: true}
	controller := r.obj.Controller
	for _, opponentID := range aliveOpponents(r.game, controller) {
		var creatures []*game.Permanent
		for _, permanent := range r.game.Battlefield {
			if effectiveController(r.game, permanent) == opponentID && permanentHasType(r.game, permanent, types.Creature) {
				creatures = append(creatures, permanent)
			}
		}
		if chosen := r.chooseUpToOnePermanent(controller, creatures, "Choose a creature to goad"); chosen != nil {
			goadPermanent(r.game, chosen, controller, false)
			res.succeeded = true
		}
	}
	return res
}

// chooseUpToOnePermanent asks playerID to choose up to one of the given
// permanents, returning nil when they choose none or none are available.
func (r *effectResolver) chooseUpToOnePermanent(playerID game.PlayerID, permanents []*game.Permanent, prompt string) *game.Permanent {
	if len(permanents) == 0 {
		return nil
	}
	options := make([]game.ChoiceOption, len(permanents))
	for i, permanent := range permanents {
		options[i] = game.ChoiceOption{Index: i, Label: permanentChoiceLabel(r.game, permanent), Card: permanentChoiceInfo(r.game, permanent)}
	}
	selected := r.engine.chooseChoice(r.game, r.agents, game.ChoiceRequest{
		Kind:       game.ChoiceResolution,
		Player:     playerID,
		Prompt:     prompt,
		Options:    options,
		MinChoices: 0,
		MaxChoices: 1,
	}, r.log)
	if len(selected) == 0 || selected[0] < 0 || selected[0] >= len(permanents) {
		return nil
	}
	return permanents[selected[0]]
}

// handleCreateCommanderCopyToken resolves CreateCommanderCopyToken: it creates a
// token that is a copy of one of the controller's commanders, except it is not
// legendary. It backs Baldur's Gate Wilderness's Circus of the Last Days.
func handleCreateCommanderCopyToken(r *effectResolver, _ game.CreateCommanderCopyToken) effectResolved {
	res := effectResolved{accepted: true}
	controller := r.obj.Controller
	player, ok := playerByID(r.game, controller)
	if !ok || player.CommanderInstanceID == 0 {
		return res
	}
	commander, ok := r.game.GetCardInstance(player.CommanderInstanceID)
	if !ok || commander.Def == nil {
		return res
	}
	tokenDef := nonLegendaryTokenCopy(commander.Def)
	_, created := createTokenPermanentsCollectingWithChoices(r.engine, r.game, controller, tokenDef, 1, false, r.agents, r.log)
	res.succeeded = created
	return res
}

// nonLegendaryTokenCopy returns a copy of the commander's card definition with
// the legendary supertype removed on both faces, for creating a non-legendary
// token copy.
func nonLegendaryTokenCopy(def *game.CardDef) *game.CardDef {
	clone := *def
	clone.CardFace = stripLegendaryFace(def.CardFace)
	if def.Back.Exists {
		clone.Back = opt.Val(stripLegendaryFace(def.Back.Val))
	}
	return &clone
}

// stripLegendaryFace returns a face copy with the legendary supertype removed.
func stripLegendaryFace(face game.CardFace) game.CardFace {
	supertypes := make([]types.Super, 0, len(face.Supertypes))
	for _, super := range face.Supertypes {
		if super != types.Legendary {
			supertypes = append(supertypes, super)
		}
	}
	face.Supertypes = supertypes
	return face
}
