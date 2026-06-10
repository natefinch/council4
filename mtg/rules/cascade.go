package rules

import (
	"math/rand/v2"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
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
	player, ok := playerByID(g, playerID)
	if !ok || !player.Exile.Contains(cardID) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef := cardFaceOrDefault(card, game.FaceFront)
	modes, targets, ok := firstLegalSpellCastChoice(g, playerID, spellDef)
	if !ok {
		return false
	}
	targetCounts, ok := spellTargetCounts(g, playerID, spellDef, modes, targets)
	if !ok {
		panic("validated cascade spell targets could not be segmented")
	}
	if !player.Exile.Remove(cardID) {
		return false
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     cardID,
		Face:         game.FaceFront,
		Controller:   playerID,
		Targets:      append([]game.Target(nil), targets...),
		TargetCounts: targetCounts,
		ChosenModes:  append([]int(nil), modes...),
	}
	stormCopies := stormCopyCount(g, spellDef)
	pushSpellToStack(g, obj, game.Event{
		SourceID:      cardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cardID,
		CardTypes:     cardTypes(spellDef),
		Colors:        spellColors(spellDef),
		FromZone:      zone.Exile,
		ToZone:        zone.Stack,
	})
	createStormCopies(g, obj, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}
