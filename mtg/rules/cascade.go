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
	// Only completed casts call this helper. A card copy can be cast (CR 707.12);
	// ordinary stack/storm copying never enters a cast path (CR 707.10).
	if spellDef == nil || !spellDef.HasKeyword(game.Cascade) {
		return
	}
	player, ok := playerByID(g, obj.Controller)
	if !ok {
		return
	}
	var revealed []game.ZoneCardSnapshot
	var found id.ID
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		epoch := player.Library.ReorderVersion()
		destination, moved := moveCardBetweenZonesWithEvent(g, game.Event{
			Controller: obj.Controller,
			Player:     obj.Controller,
			CardID:     cardID,
			FromZone:   zone.Library,
			ToZone:     zone.Exile,
		}, zoneMoveOptions{})
		if !moved {
			break
		}
		if destination == zone.Exile {
			card, _ := g.GetCardInstance(cardID)
			revealed = append(revealed, game.ZoneCardSnapshot{CardID: cardID, ZoneVersion: card.ZoneVersion})
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if destination == zone.Exile && !def.HasType(types.Land) && def.ManaValue() < spellDef.ManaValue() {
			found = cardID
			break
		}
		requireLibraryAdvance(player, cardID, epoch)
	}
	if found != 0 {
		e.castFreeSpellFromExileWithMaxManaValue(g, obj.Controller, found, spellDef.ManaValue()-1, agents, log)
	}
	bottomExiledCards(g, player, revealed, e.rng)
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
		bottomExiledCards(g, player, revealed, e.rng)
		return len(revealed) > 0
	}
	if e.chooseMay(g, agents, obj.Controller, "Cast discovered card without paying its mana cost?", log) &&
		e.castFreeSpellFromExileWithMaxManaValue(g, obj.Controller, found, manaValue, agents, log) {
		bottomExiledCards(g, player, revealed, e.rng)
		return true
	}
	moveCardBetweenZonesInBatch(g, obj.Controller, found, zone.Exile, zone.Hand, false, 0)
	bottomExiledCards(g, player, revealed, e.rng)
	return true
}

func exileUntilDiscoverHit(g *game.Game, player *game.Player, playerID game.PlayerID, manaValue int) (revealed []game.ZoneCardSnapshot, found id.ID) {
	for {
		cardID, ok := player.Library.Top()
		if !ok {
			break
		}
		epoch := player.Library.ReorderVersion()
		destination, moved := moveCardBetweenZonesWithEvent(g, game.Event{
			Controller: playerID,
			Player:     playerID,
			CardID:     cardID,
			FromZone:   zone.Library,
			ToZone:     zone.Exile,
		}, zoneMoveOptions{})
		if !moved {
			break
		}
		if destination == zone.Exile {
			card, _ := g.GetCardInstance(cardID)
			revealed = append(revealed, game.ZoneCardSnapshot{CardID: cardID, ZoneVersion: card.ZoneVersion})
		}
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if destination == zone.Exile && !def.HasType(types.Land) && def.ManaValue() <= manaValue {
			found = cardID
			break
		}
		requireLibraryAdvance(player, cardID, epoch)
	}
	return revealed, found
}

// bottomExiledCards puts the still-exiled cards on the bottom of player's library
// in a random order as one simultaneous move ("put all ... on the bottom"): every
// card shares one batch id, so a one-or-more zone-change trigger sees one event
// (CR 603.2c). Each card's zone-change replacements and commander choice are
// determined before any card moves (CR 400.6, 614, 903.9b). A snapshot whose
// card has left that exile incarnation is skipped.
func bottomExiledCards(g *game.Game, player *game.Player, cards []game.ZoneCardSnapshot, rng *rand.Rand) bool {
	var moves []simultaneousCardMove
	for _, snapshot := range randomlyOrderedSnapshots(cards, rng) {
		if player.Exile.Contains(snapshot.CardID) {
			moves = append(moves, simultaneousCardMove{card: snapshot, fromZone: zone.Exile})
		}
	}
	if len(moves) == 0 {
		return false
	}
	return moveCardsBetweenZonesSimultaneously(g, moves, zone.Library, true, g.IDGen.Next())
}

// randomlyOrderedSnapshots returns cards in a random order drawn from rng (or
// unchanged without one), leaving the caller's slice intact.
func randomlyOrderedSnapshots(cards []game.ZoneCardSnapshot, rng *rand.Rand) []game.ZoneCardSnapshot {
	cards = slices.Clone(cards)
	if rng != nil && len(cards) > 1 {
		rng.Shuffle(len(cards), func(i, j int) {
			cards[i], cards[j] = cards[j], cards[i]
		})
	}
	return cards
}

func (e *Engine) castFreeSpellFromExile(g *game.Game, playerID game.PlayerID, cardID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.castFreeSpellFromZone(g, playerID, cardID, zone.Exile, agents, log)
}

func (e *Engine) castFreeSpellFromExileWithMaxManaValue(g *game.Game, playerID game.PlayerID, cardID id.ID, maxManaValue int, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	return e.chooseAndCastFreeSpellFromSource(
		g,
		player,
		playerID,
		cardID,
		zone.Exile,
		game.Selection{},
		opt.Val(maxManaValue),
		false,
		agents,
		log,
	)
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
			if e.castableExiledSpell(g, controllerID, cardID) {
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
// any number of spells from among those cards" applies to the exiled pool.
func (e *Engine) castableExiledSpell(g *game.Game, controllerID game.PlayerID, cardID id.ID) bool {
	if _, ok := playerHoldingCastSource(g, cardID, zone.Exile); !ok {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	return len(e.freeCastOptionsForCard(g, controllerID, card, zone.Exile, game.Selection{}, opt.V[int]{})) > 0
}

// castFreeSpellFromZone casts cardID from fromZone for playerID without paying
// its mana cost, letting the player choose its legal face, modes, targets, and
// optional additional costs. It returns false (casting nothing) when the card is
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
// controllerID without paying its mana cost, letting the player choose its legal
// face, modes, targets, and optional additional costs before pushing it to the
// stack under controllerID's control. The source player and controller differ
// only for a targeted cast from another player's zone. exileOnResolution
// redirects the resolved spell to exile in place of its owner's graveyard.
func (e *Engine) castFreeSpellFromSource(g *game.Game, sourcePlayer *game.Player, controllerID game.PlayerID, cardID id.ID, fromZone zone.Type, exileOnResolution bool, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	return e.chooseAndCastFreeSpellFromSource(
		g,
		sourcePlayer,
		controllerID,
		cardID,
		fromZone,
		game.Selection{},
		opt.V[int]{},
		exileOnResolution,
		agents,
		log,
	)
}
