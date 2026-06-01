package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

func resolveFight(g *game.Game, obj *game.StackObject, effect game.Effect) {
	if obj == nil || len(obj.Targets) < 2 {
		return
	}
	first, firstOK := permanentByObjectID(g, obj.Targets[0].PermanentID)
	second, secondOK := permanentByObjectID(g, obj.Targets[1].PermanentID)
	if !firstOK || !secondOK || first.ObjectID == second.ObjectID || !permanentHasType(g, first, game.TypeCreature) || !permanentHasType(g, second, game.TypeCreature) {
		return
	}
	dealPermanentDamage(g, first.CardInstanceID, first.ObjectID, effectiveController(g, first), second, effectivePower(g, first), false)
	dealPermanentDamage(g, second.CardInstanceID, second.ObjectID, effectiveController(g, second), first, effectivePower(g, second), false)
}

func counterTargetStackObject(g *game.Game, obj *game.StackObject, effect game.Effect) bool {
	stackObjectID, ok := effectStackObjectID(obj, effect)
	return ok && counterStackObject(g, stackObjectID)
}

func effectStackObjectID(obj *game.StackObject, effect game.Effect) (id.ID, bool) {
	if effect.TargetIndex < 0 || effect.TargetIndex >= len(obj.Targets) {
		return 0, false
	}
	target := obj.Targets[effect.TargetIndex]
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
	return spec.SourceZone == game.ZoneLibrary && spec.Destination == game.ZoneHand
}

func (e *Engine) searchLibrary(g *game.Game, obj *game.StackObject, playerID game.PlayerID, spec game.SearchSpec, amount int) bool {
	if amount <= 0 {
		amount = 1
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return false
	}
	var found []id.ID
	for _, cardID := range player.Library.All() {
		if searchSpecMatches(g, cardID, spec) {
			found = append(found, cardID)
			if len(found) == amount {
				break
			}
		}
	}
	for _, cardID := range found {
		if !player.Library.Remove(cardID) {
			return len(found) > 0
		}
		if spec.Reveal {
			emitCardRevealEvent(g, obj, playerID, cardID, game.ZoneLibrary)
		}
		player.Hand.Add(cardID)
		emitZoneChangeEvent(g, game.GameEvent{
			SourceID:      stackObjectSourceID(obj),
			StackObjectID: stackObjectID(obj),
			Controller:    stackObjectController(obj),
			Player:        playerID,
			CardID:        cardID,
			FromZone:      game.ZoneLibrary,
			ToZone:        game.ZoneHand,
			Amount:        1,
		})
	}
	if spec.Shuffle {
		player.Library.Shuffle(e.rng)
	}
	return len(found) > 0
}

func searchSpecMatches(g *game.Game, cardID id.ID, spec game.SearchSpec) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	if spec.MatchCardType && !card.Def.HasType(spec.CardType) {
		return false
	}
	if spec.MatchSupertype && !card.Def.HasSupertype(spec.Supertype) {
		return false
	}
	return true
}

func revealCards(g *game.Game, obj *game.StackObject, playerID game.PlayerID, zone game.ZoneType, amount int) bool {
	return len(revealCardIDs(g, obj, playerID, zone, amount)) > 0
}

func revealCardIDs(g *game.Game, obj *game.StackObject, playerID game.PlayerID, zone game.ZoneType, amount int) []id.ID {
	if amount <= 0 {
		amount = 1
	}
	player, ok := playerByID(g, playerID)
	if !ok || zone != game.ZoneLibrary {
		return nil
	}
	var revealed []id.ID
	for i, cardID := range player.Library.All() {
		if i >= amount {
			break
		}
		emitCardRevealEvent(g, obj, playerID, cardID, zone)
		revealed = append(revealed, cardID)
	}
	return revealed
}

func emitCardRevealEvent(g *game.Game, obj *game.StackObject, playerID game.PlayerID, cardID id.ID, zone game.ZoneType) {
	emitEvent(g, game.GameEvent{
		Kind:          game.EventCardRevealed,
		SourceID:      stackObjectSourceID(obj),
		StackObjectID: stackObjectID(obj),
		Controller:    stackObjectController(obj),
		Player:        playerID,
		CardID:        cardID,
		FromZone:      zone,
		Amount:        1,
	})
}

func clueTokenDef() *game.CardDef {
	two := mana.Cost{mana.GenericMana(2)}
	return &game.CardDef{
		Name:     "Clue Token",
		Types:    []game.CardType{game.TypeArtifact},
		Subtypes: []string{game.ArtifactSubtypeClue},
		Abilities: []game.AbilityDef{{
			Kind:     game.ActivatedAbility,
			Text:     "{2}, Sacrifice this artifact: Draw a card.",
			ManaCost: opt.Val(two),
			AdditionalCosts: []game.AdditionalCost{{
				Kind:               game.AdditionalCostSacrificeSource,
				Text:               "Sacrifice this artifact",
				Amount:             1,
				MatchPermanentType: true,
				PermanentType:      game.TypeArtifact,
			}},
			Effects: []game.Effect{{Type: game.EffectDraw, Amount: 1, TargetIndex: -1}},
		}},
	}
}
