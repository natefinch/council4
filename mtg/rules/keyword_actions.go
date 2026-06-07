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
	emitFightEvent(g, first, second)
	emitFightEvent(g, second, first)
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

func emitFightEvent(g *game.Game, permanent, related *game.Permanent) {
	emitEvent(g, game.GameEvent{
		Kind:               game.EventFight,
		SourceID:           permanent.CardInstanceID,
		SourceObjectID:     permanent.ObjectID,
		Controller:         effectiveController(g, permanent),
		PermanentID:        permanent.ObjectID,
		RelatedPermanentID: related.ObjectID,
	})
}

func counterTargetStackObject(g *game.Game, obj *game.StackObject, targetIndex int) bool {
	stackObjectID, ok := effectStackObjectID(obj, targetIndex)
	return ok && counterStackObject(g, stackObjectID)
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
	return spec.SourceZone == zone.Library && (spec.Destination == zone.Hand || spec.Destination == zone.Battlefield)
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
			emitCardRevealEvent(g, obj, playerID, cardID, zone.Library)
		}
		switch spec.Destination {
		case zone.Hand:
			player.Hand.Add(cardID)
			emitZoneChangeEvent(g, game.GameEvent{
				SourceID:      stackObjectSourceID(obj),
				StackObjectID: stackObjectID(obj),
				Controller:    stackObjectController(obj),
				Player:        playerID,
				CardID:        cardID,
				FromZone:      zone.Library,
				ToZone:        zone.Hand,
				Amount:        1,
			})
		case zone.Battlefield:
			card, ok := g.GetCardInstance(cardID)
			if !ok {
				return len(found) > 0
			}
			if _, ok := createCardPermanentFaceWithOptions(e, g, card, playerID, zone.Library, game.FaceFront, nil, permanentCreationOptions{ForceTapped: spec.EntersTapped}, [game.NumPlayers]PlayerAgent{}, nil); !ok {
				return len(found) > 0
			}
		default:
		}
	}
	player.Library.Shuffle(e.rng)
	return len(found) > 0
}

func searchSpecMatches(g *game.Game, cardID id.ID, spec game.SearchSpec) bool {
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	if spec.CardType.Exists && !card.Def.HasType(spec.CardType.Val) {
		return false
	}
	if spec.Supertype.Exists && !card.Def.HasSupertype(spec.Supertype.Val) {
		return false
	}
	if len(spec.SubtypesAny) > 0 && !card.Def.HasAnySubtype(spec.SubtypesAny...) {
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
	emitEvent(g, game.GameEvent{
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
	drawContent := game.PlainAbilityContent{Sequence: []game.Instruction{
		{Primitive: game.Draw{Amount: game.Fixed(1), TargetIndex: game.TargetIndexController}},
	}}
	return &game.CardDef{CardFace: game.CardFace{Name: "Clue Token",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Clue},
		ActivatedAbilities: []game.ActivatedAbilityBody{{
			Text:            "{2}, Sacrifice this artifact: Draw a card.",
			ManaCost:        opt.Val(two),
			AdditionalCosts: additionalCosts,
			Content:         drawContent,
		}}},
	}
}
