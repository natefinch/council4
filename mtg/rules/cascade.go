package rules

import (
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func (e *Engine) resolveCascadeForCast(g *game.Game, obj *game.StackObject, spellDef *game.CardDef, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	if obj.Copy || spellDef == nil || !spellDef.HasKeyword(game.Cascade) {
		return
	}
	player, ok := playerByID(g, obj.Controller)
	if !ok {
		return
	}
	var revealed []id.ID
	var found id.ID
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		player.Library.Remove(cardID)
		player.Exile.Add(cardID)
		revealed = append(revealed, cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   obj.Controller,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Exile,
		})
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if !def.HasType(types.Land) && def.ManaValue() < spellDef.ManaValue() {
			found = cardID
			break
		}
	}
	if found != 0 {
		e.castFreeSpellFromExile(g, obj.Controller, found, agents, log)
	}
	bottomExiledCards(g, player, obj.Controller, revealed, e.rng)
}

func (e *Engine) resolveDiscover(g *game.Game, obj *game.StackObject, manaValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if manaValue < 0 {
		return false
	}
	player, ok := playerByID(g, obj.Controller)
	if !ok {
		return false
	}
	revealed, found := exileUntilDiscoverHit(g, player, obj.Controller, manaValue)
	if found == 0 {
		bottomExiledCards(g, player, obj.Controller, revealed, e.rng)
		return len(revealed) > 0
	}
	if e.chooseMay(g, agents, obj.Controller, "Cast discovered card without paying its mana cost?", log) && e.castFreeSpellFromExile(g, obj.Controller, found, agents, log) {
		bottomExiledCards(g, player, obj.Controller, revealed, e.rng)
		return true
	}
	if player.Exile.Remove(found) {
		player.Hand.Add(found)
		emitZoneChangeEvent(g, game.Event{
			Player:   obj.Controller,
			CardID:   found,
			FromZone: zone.Exile,
			ToZone:   zone.Hand,
		})
	}
	bottomExiledCards(g, player, obj.Controller, revealed, e.rng)
	return true
}

func exileUntilDiscoverHit(g *game.Game, player *game.Player, playerID game.PlayerID, manaValue int) (revealed []id.ID, found id.ID) {
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		player.Library.Remove(cardID)
		player.Exile.Add(cardID)
		revealed = append(revealed, cardID)
		emitZoneChangeEvent(g, game.Event{
			Player:   playerID,
			CardID:   cardID,
			FromZone: zone.Library,
			ToZone:   zone.Exile,
		})
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if !def.HasType(types.Land) && def.ManaValue() <= manaValue {
			found = cardID
			break
		}
	}
	return revealed, found
}

func bottomExiledCards(g *game.Game, player *game.Player, playerID game.PlayerID, cardIDs []id.ID, rng *rand.Rand) {
	cardIDs = append([]id.ID(nil), cardIDs...)
	if rng != nil && len(cardIDs) > 1 {
		rng.Shuffle(len(cardIDs), func(i, j int) {
			cardIDs[i], cardIDs[j] = cardIDs[j], cardIDs[i]
		})
	}
	for _, cardID := range cardIDs {
		if player.Exile.Remove(cardID) {
			player.Library.AddToBottom(cardID)
			emitZoneChangeEvent(g, game.Event{
				Player:   playerID,
				CardID:   cardID,
				FromZone: zone.Exile,
				ToZone:   zone.Library,
			})
		}
	}
}

func (e *Engine) castFreeSpellFromExile(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.castFreeSpellFromZone(g, playerID, cardID, zone.Exile, agents, log)
}

// castAnyNumberFromExileForFree lets controllerID cast any number of the given
// exiled cards without paying their mana costs, one at a time in an order the
// controller chooses, during the resolution of the ability that exiled them
// (Etali, Primal Storm). Each cast card is put on the stack under controllerID's
// control wherever it currently rests, so a card exiled from an opponent's
// library is still cast by the controller. Only cards still in exile that have a
// legal cast choice are offered, so lands and uncastable cards are skipped and
// left exiled. It stops when the controller casts nothing more or no castable
// card remains; removing the chosen card each iteration guarantees termination.
func (e *Engine) castAnyNumberFromExileForFree(g *game.Game, controllerID game.PlayerID, cards []id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) {
	remaining := append([]id.ID(nil), cards...)
	for len(remaining) > 0 {
		var candidates []id.ID
		for _, cardID := range remaining {
			if castableExiledSpell(g, controllerID, cardID) {
				candidates = append(candidates, cardID)
			}
		}
		if len(candidates) == 0 {
			return
		}
		options := make([]game.ChoiceOption, len(candidates))
		for i, cardID := range candidates {
			options[i] = game.ChoiceOption{
				Index: i,
				Label: cardChoiceLabel(g, cardID),
				Card:  cardChoiceInfo(g, cardID),
			}
		}
		selected := e.chooseChoice(g, agents, game.ChoiceRequest{
			Kind:       game.ChoiceResolution,
			Player:     controllerID,
			Prompt:     "Choose a spell to cast without paying its mana cost, or none to stop",
			Options:    options,
			MinChoices: 0,
			MaxChoices: 1,
		}, log)
		if len(selected) == 0 || selected[0] < 0 || selected[0] >= len(candidates) {
			return
		}
		chosen := candidates[selected[0]]
		if i := slices.Index(remaining, chosen); i >= 0 {
			remaining = slices.Delete(remaining, i, i+1)
		}
		e.castFreeTargetedSpell(g, controllerID, chosen, zone.Exile, false, agents, log)
	}
}

// castableExiledSpell reports whether cardID still rests in some player's exile
// and has a legal free-cast choice for controllerID, the filter Etali's "cast
// any number of spells from among those cards" applies to the exiled pool. Lands
// and uncastable cards fail isSupportedSpell inside firstLegalSpellCastChoice.
func castableExiledSpell(g *game.Game, controllerID game.PlayerID, cardID id.ID) bool {
	if _, ok := playerHoldingCastSource(g, cardID, zone.Exile); !ok {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	_, _, legal := firstLegalSpellCastChoice(g, controllerID, cardFaceOrDefault(card, game.FaceFront))
	return legal
}

// castFreeSpellFromZone casts cardID from fromZone for playerID without paying
// its mana cost, choosing the first legal modes/targets and pushing the spell to
// the stack as a free cast. It returns false (casting nothing) when the card is
// no longer in fromZone or has no legal cast choice.
func (e *Engine) castFreeSpellFromZone(g *game.Game, playerID game.PlayerID, cardID id.ID, fromZone zone.Type, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	return e.castFreeSpellFromSource(g, player, playerID, cardID, fromZone, false, agents, log)
}

// castFreeTargetedSpell casts a specific targeted card from fromZone for
// controllerID without paying its mana cost, locating whichever player's zone
// currently holds the card so a spell that targets a card in an opponent's
// graveyard (Memory Plunder) casts it under the controller's control. When
// exileOnResolution is set the cast spell moves to exile instead of its owner's
// graveyard after it resolves, modeling the "exile it instead" rider. It returns
// false when the card no longer rests in a player's fromZone or has no legal cast
// choice.
func (e *Engine) castFreeTargetedSpell(g *game.Game, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	source, ok := playerHoldingCastSource(g, cardID, fromZone)
	if !ok {
		return false
	}
	return e.castFreeSpellFromSource(g, source, controllerID, cardID, fromZone, exileOnResolution, agents, log)
}

// playerHoldingCastSource returns the player whose fromZone currently contains
// cardID, scanning every player in turn order. It backs targeted free casts,
// whose card may rest in any player's zone (an opponent's graveyard).
func playerHoldingCastSource(g *game.Game, cardID id.ID, fromZone zone.Type) (*game.Player, bool) {
	for i := range g.Players {
		player := g.Players[i]
		if castSourceContains(player, cardID, fromZone) {
			return player, true
		}
	}
	return nil, false
}

// castFreeSpellFromSource casts cardID out of sourcePlayer's fromZone for
// controllerID without paying its mana cost, choosing the first legal
// modes/targets and pushing the spell to the stack as a free cast under
// controllerID's control. The source player and controller differ only for a
// targeted cast from another player's zone. exileOnResolution redirects the
// resolved spell to exile in place of its owner's graveyard.
func (e *Engine) castFreeSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if sourcePlayer == nil || !castSourceContains(sourcePlayer, cardID, fromZone) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	if !cardCastRestrictionsSatisfied(g, controllerID, spellDef) {
		return false
	}
	modes, targets, ok := firstLegalSpellCastChoice(g, controllerID, spellDef)
	if !ok {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, controllerID, spellDef, modes, targets, game.CastBranch{})
	if !ok {
		panic("validated free-cast spell targets could not be segmented")
	}
	if !removeCastSourceCard(g, sourcePlayer, cardID, fromZone) {
		return false
	}
	obj := &game.StackObject{
		ID:                g.IDGen.Next(),
		Kind:              game.StackSpell,
		SourceID:          cardID,
		Face:              game.FaceFront,
		Controller:        controllerID,
		Targets:           append([]game.Target(nil), targets...),
		TargetCounts:      targetCounts,
		ChosenModes:       append([]int(nil), modes...),
		ExileOnResolution: exileOnResolution,
	}
	stormCopies := stormCopyCount(g, spellDef)
	pushSpellToStack(g, obj, game.Event{
		SourceID:                     cardID,
		StackObjectID:                obj.ID,
		Controller:                   controllerID,
		CardID:                       cardID,
		CardTypes:                    cardTypes(spellDef),
		CardSupertypes:               cardSupertypes(spellDef),
		CardSubtypes:                 cardSubtypes(spellDef),
		Colors:                       spellColors(spellDef),
		ManaValue:                    opt.Val(stackManaValue(spellDef, 0)),
		ManaSpentToCast:              opt.Val(0),
		ManaFromCreaturesSpentToCast: opt.Val(0),
		FromZone:                     fromZone,
		ToZone:                       zone.Stack,
	})
	createStormCopies(g, obj, spellDef, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}
