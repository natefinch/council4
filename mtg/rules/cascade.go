package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
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
		emitZoneChangeEvent(g, game.GameEvent{
			Player:   obj.Controller,
			CardID:   cardID,
			FromZone: game.ZoneLibrary,
			ToZone:   game.ZoneExile,
		})
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		def := cardFaceOrDefault(card, game.FaceFront)
		if !def.HasType(game.TypeLand) && def.ManaValue < spellDef.ManaValue {
			found = cardID
			break
		}
	}
	if found != 0 {
		e.castFreeSpellFromExile(g, obj.Controller, found, agents, log)
	}
	for _, cardID := range revealed {
		if player.Exile.Remove(cardID) {
			player.Library.AddToBottom(cardID)
			emitZoneChangeEvent(g, game.GameEvent{
				Player:   obj.Controller,
				CardID:   cardID,
				FromZone: game.ZoneExile,
				ToZone:   game.ZoneLibrary,
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
	if !player.Exile.Remove(cardID) {
		return false
	}
	obj := &game.StackObject{
		ID:          g.IDGen.Next(),
		Kind:        game.StackSpell,
		SourceID:    cardID,
		Face:        game.FaceFront,
		Controller:  playerID,
		Targets:     append([]game.Target(nil), targets...),
		ChosenModes: append([]int(nil), modes...),
	}
	stormCopies := stormCopyCount(g, spellDef)
	pushSpellToStack(g, obj, game.GameEvent{
		SourceID:      cardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cardID,
		CardTypes:     cardTypes(spellDef),
		FromZone:      game.ZoneExile,
		ToZone:        game.ZoneStack,
	})
	createStormCopies(g, obj, stormCopies)
	e.resolveCascadeForCast(g, obj, spellDef, agents, log)
	return true
}
