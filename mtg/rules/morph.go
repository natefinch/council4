package rules

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

var faceDownCastCost = cost.Mana{cost.O(3)}
var faceDownDisguiseWardCost = cost.Mana{cost.O(2)}

func faceDownDisguiseWardAbility() game.AbilityDef {
	return game.AbilityDef{
		Kind:     game.StaticAbility,
		Text:     "Ward {2}",
		Keywords: []game.Keyword{game.Ward},
		WardCost: opt.Val(faceDownDisguiseWardCost),
	}
}

func faceDownCostForCard(card *game.CardDef, kind game.FaceDownKind) (cost.Mana, bool) {
	keyword := game.Morph
	if kind == game.FaceDownDisguise {
		keyword = game.Disguise
	}
	abilities := card.AbilityDefs()
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasKeyword(ability, keyword) {
			continue
		}
		switch kind {
		case game.FaceDownMorph:
			if ability.MorphCost.Exists {
				return ability.MorphCost.Val, true
			}
		case game.FaceDownDisguise:
			if ability.DisguiseCost.Exists {
				return ability.DisguiseCost.Val, true
			}
		default:
		}
	}
	return nil, false
}

func faceDownKindsForCard(card *game.CardDef) []game.FaceDownKind {
	var kinds []game.FaceDownKind
	if _, ok := faceDownCostForCard(card, game.FaceDownMorph); ok {
		kinds = append(kinds, game.FaceDownMorph)
	}
	if _, ok := faceDownCostForCard(card, game.FaceDownDisguise); ok {
		kinds = append(kinds, game.FaceDownDisguise)
	}
	return kinds
}

func (e *Engine) legalFaceDownCastActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || !isSorcerySpeed(g, playerID) || splitSecondOnStack(g) {
		return nil
	}
	player, ok := playerByID(g, playerID)
	if !ok {
		return nil
	}
	var actions []action.Action
	for _, cardID := range player.Hand.All() {
		card, ok := g.GetCardInstance(cardID)
		if !ok {
			continue
		}
		for _, face := range legalFaceDownFaces(card.Def) {
			spellDef := cardFaceOrDefault(card, face)
			for _, kind := range faceDownKindsForCard(spellDef) {
				if e.canCastFaceDown(g, playerID, cardID, face, kind) {
					actions = append(actions, actionBuild.castFaceDown(cardID, face, kind))
				}
			}
		}
	}
	return actions
}

func legalFaceDownFaces(card *game.CardDef) []game.FaceIndex {
	var faces []game.FaceIndex
	for _, face := range card.FaceIndexes() {
		if def, ok := card.FaceDef(face); ok && len(faceDownKindsForCard(def)) > 0 {
			faces = append(faces, face)
		}
	}
	return faces
}

func (*Engine) canCastFaceDown(g *game.Game, playerID game.PlayerID, cardID id.ID, face game.FaceIndex, kind game.FaceDownKind) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer || !isSorcerySpeed(g, playerID) || splitSecondOnStack(g) || kind == game.FaceDownNone {
		return false
	}
	player, ok := playerByID(g, playerID)
	if !ok || !player.Hand.Contains(cardID) {
		return false
	}
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		return false
	}
	spellDef, ok := cardFaceDef(card, face)
	if !ok {
		return false
	}
	if _, ok := faceDownCostForCard(spellDef, kind); !ok {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &faceDownCastCost})
}

func (e *Engine) applyCastFaceDownWithChoices(g *game.Game, playerID game.PlayerID, cast action.CastFaceDownAction, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.canCastFaceDown(g, playerID, cast.CardID, cast.Face, cast.FaceDownKind) {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, &faceDownCastCost, nil, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &faceDownCastCost, Prefs: prefs}) {
		return false
	}
	player := g.Players[playerID]
	if !player.Hand.Remove(cast.CardID) {
		panic("face-down cast card disappeared from hand after validation")
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     cast.CardID,
		Face:         cast.Face,
		Controller:   playerID,
		FaceDown:     true,
		FaceDownFace: cast.Face,
		FaceDownKind: cast.FaceDownKind,
	}
	pushSpellToStack(g, obj, game.GameEvent{
		SourceID:      cast.CardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cast.CardID,
		Face:          cast.Face,
		CardTypes:     []types.Card{types.Creature},
		FromZone:      game.ZoneHand,
		ToZone:        game.ZoneStack,
	})
	return true
}

func (e *Engine) legalTurnFaceUpActions(g *game.Game, playerID game.PlayerID) []action.Action {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return nil
	}
	var actions []action.Action
	for _, permanent := range g.Battlefield {
		if e.canTurnFaceUp(g, playerID, permanent.ObjectID) {
			actions = append(actions, actionBuild.turnFaceUp(permanent.ObjectID))
		}
	}
	return actions
}

func (*Engine) canTurnFaceUp(g *game.Game, playerID game.PlayerID, permanentID id.ID) bool {
	if !canAct(g, playerID) || playerID != g.Turn.PriorityPlayer {
		return false
	}
	permanent, ok := permanentByObjectID(g, permanentID)
	if !ok || permanent.PhasedOut || !permanent.FaceDown || effectiveController(g, permanent) != playerID || permanent.FaceDownKind == game.FaceDownNone {
		return false
	}
	card, ok := physicalPermanentDef(g, permanent)
	if !ok {
		return false
	}
	face, ok := card.FaceDef(permanent.FaceDownFace)
	if !ok {
		return false
	}
	manaCost, ok := faceDownCostForCard(face, permanent.FaceDownKind)
	if !ok {
		return false
	}
	return paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost})
}

func (e *Engine) applyTurnFaceUpWithChoices(g *game.Game, playerID game.PlayerID, permanentID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.canTurnFaceUp(g, playerID, permanentID) {
		return false
	}
	permanent, _ := permanentByObjectID(g, permanentID)
	card, _ := physicalPermanentDef(g, permanent)
	face, _ := card.FaceDef(permanent.FaceDownFace)
	manaCost, _ := faceDownCostForCard(face, permanent.FaceDownKind)
	prefs := e.paymentPreferencesForCost(g, playerID, &manaCost, nil, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost, Prefs: prefs}) {
		return false
	}
	kind := permanent.FaceDownKind
	permanent.Face = permanent.FaceDownFace
	permanent.FaceDown = false
	permanent.FaceDownKind = game.FaceDownNone
	if kind == game.FaceDownDisguise {
		permanent.Counters.Add(counter.Shield, 1)
	}
	emitFaceDownRevealEvent(g, permanent)
	emitEvent(g, game.GameEvent{
		Kind:        game.EventPermanentTurnedFaceUp,
		Controller:  playerID,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
		PermanentID: permanent.ObjectID,
	})
	return true
}

func emitFaceDownRevealEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, game.GameEvent{
		Kind:        game.EventCardRevealed,
		Controller:  effectiveController(g, permanent),
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.FaceDownFace,
		PermanentID: permanent.ObjectID,
		TokenName:   permanentTokenName(permanent),
		TokenDef:    permanent.TokenDef,
	})
}
