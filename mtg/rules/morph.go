package rules

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

var faceDownCastCost = cost.Mana{cost.O(3)}

func faceDownDisguiseWardBody() game.StaticAbility {
	return game.StaticAbility{
		Text:             "Ward {2}",
		KeywordAbilities: []game.KeywordAbility{game.WardKeyword{Cost: cost.Mana{cost.O(2)}}},
	}
}

func faceDownCostForCard(card *game.CardDef, kind game.FaceDownKind) (cost.Mana, bool) {
	costs := faceDownCostsForCard(card, kind)
	if len(costs) == 0 {
		return nil, false
	}
	return costs[0], true
}

func faceDownCostsForCard(card *game.CardDef, kind game.FaceDownKind) []cost.Mana {
	var costs []cost.Mana
	if kind == game.FaceDownManifest || kind == game.FaceDownCloak {
		if card.HasType(types.Creature) && card.ManaCost.Exists {
			costs = append(costs, card.ManaCost.Val)
		}
	}
	for i := range card.ActivatedAbilities {
		ability := &card.ActivatedAbilities[i]
		if kind == game.FaceDownMorph || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.ActivatedBodyMorphCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
		if kind == game.FaceDownDisguise || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.ActivatedBodyDisguiseCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
	}
	for i := range card.StaticAbilities {
		ability := &card.StaticAbilities[i]
		if kind == game.FaceDownMorph || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.StaticBodyMorphCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
		if kind == game.FaceDownDisguise || kind == game.FaceDownManifest || kind == game.FaceDownCloak || kind == game.FaceDownEffect {
			if manaCost, ok := game.StaticBodyDisguiseCost(ability); ok {
				costs = append(costs, manaCost)
			}
		}
	}
	return costs
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
	prefs := e.paymentPreferencesForCost(g, playerID, &faceDownCastCost, nil, 0, agents, log)
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
	pushSpellToStack(g, obj, game.Event{
		SourceID:      cast.CardID,
		StackObjectID: obj.ID,
		Controller:    playerID,
		CardID:        cast.CardID,
		Face:          cast.Face,
		CardTypes:     []types.Card{types.Creature},
		Colors:        nil, // Face-down spells are colorless (CR 708.2b).
		ManaValue:     opt.Val(0),
		// A face-down permanent is cast for the fixed {3} face-down cast cost
		// (CR 702.37e); record that spend for mana-spent-to-cast triggers. The
		// face-down cast path does not track per-unit provenance, so no creature
		// mana is attributed.
		ManaSpentToCast:              opt.Val(faceDownCastCost.ManaValue()),
		ManaFromCreaturesSpentToCast: opt.Val(0),
		FromZone:                     zone.Hand,
		ToZone:                       zone.Stack,
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
	return len(payableFaceDownCostsForPermanent(g, playerID, permanent)) > 0
}

func (e *Engine) applyTurnFaceUpWithChoices(g *game.Game, playerID game.PlayerID, permanentID id.ID, agents [game.NumPlayers]PlayerAgent, log *TurnLog) bool {
	if !e.canTurnFaceUp(g, playerID, permanentID) {
		return false
	}
	permanent, _ := permanentByObjectID(g, permanentID)
	manaCost, ok := e.chooseFaceDownCostOptions(
		g,
		playerID,
		payableFaceDownCostsForPermanent(g, playerID, permanent),
		agents,
		log,
	)
	if !ok {
		return false
	}
	prefs := e.paymentPreferencesForCost(g, playerID, &manaCost, nil, 0, agents, log)
	if !paymentOrch.payGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost, Prefs: prefs}) {
		return false
	}
	permanent.Face = permanent.FaceDownFace
	permanent.FaceDown = false
	permanent.FaceDownKind = game.FaceDownNone
	permanent.FaceDownCharacteristics = opt.V[game.FaceDownCharacteristics]{}
	emitFaceDownRevealEvent(g, permanent)
	emitMergedTurnFaceUpRevealEvents(g, permanent)
	emitEvent(g, game.Event{
		Kind:        game.EventPermanentTurnedFaceUp,
		Controller:  playerID,
		Player:      permanent.Owner,
		CardID:      permanent.CardInstanceID,
		Face:        permanent.Face,
		PermanentID: permanent.ObjectID,
	})
	return true
}

func payableFaceDownCosts(g *game.Game, playerID game.PlayerID, face *game.CardDef, kind game.FaceDownKind) []cost.Mana {
	var costs []cost.Mana
	for _, manaCost := range faceDownCostsForCard(face, kind) {
		if paymentOrch.canPayGenericCost(g, payment.GenericRequest{PlayerID: playerID, Cost: &manaCost}) {
			costs = append(costs, manaCost)
		}
	}
	return costs
}

func payableFaceDownCostsForPermanent(g *game.Game, playerID game.PlayerID, permanent *game.Permanent) []cost.Mana {
	var costs []cost.Mana
	for _, face := range faceDownTurnUpFaces(g, permanent) {
		for _, manaCost := range payableFaceDownCosts(g, playerID, face, permanent.FaceDownKind) {
			duplicate := false
			for _, prior := range costs {
				if prior.String() == manaCost.String() {
					duplicate = true
					break
				}
			}
			if !duplicate {
				costs = append(costs, manaCost)
			}
		}
	}
	return costs
}

func faceDownTurnUpFaces(g *game.Game, permanent *game.Permanent) []*game.CardDef {
	faceUp := *permanent
	faceUp.FaceDown = false
	faceUp.FaceDownKind = game.FaceDownNone
	faceUp.FaceDownCharacteristics = opt.V[game.FaceDownCharacteristics]{}
	def, ok := permanentCopyDef(g, &faceUp)
	if !ok {
		return nil
	}
	return []*game.CardDef{def}
}

func (e *Engine) chooseFaceDownCost(g *game.Game, playerID game.PlayerID, face *game.CardDef, kind game.FaceDownKind, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (cost.Mana, bool) {
	return e.chooseFaceDownCostOptions(g, playerID, payableFaceDownCosts(g, playerID, face, kind), agents, log)
}

func (e *Engine) chooseFaceDownCostOptions(g *game.Game, playerID game.PlayerID, costs []cost.Mana, agents [game.NumPlayers]PlayerAgent, log *TurnLog) (cost.Mana, bool) {
	if len(costs) == 0 {
		return nil, false
	}
	if len(costs) == 1 {
		return costs[0], true
	}
	options := make([]game.ChoiceOption, 0, len(costs))
	for i, manaCost := range costs {
		options = append(options, game.ChoiceOption{
			Index: i,
			Label: fmt.Sprintf("Pay %s", manaCost),
		})
	}
	request := game.ChoiceRequest{
		Kind:             game.ChoicePayment,
		Player:           playerID,
		Prompt:           "Choose turn face-up cost",
		Options:          options,
		MinChoices:       1,
		MaxChoices:       1,
		DefaultSelection: []int{0},
	}
	selected := e.chooseChoice(g, agents, request, log)
	if len(selected) == 1 && selected[0] >= 0 && selected[0] < len(costs) {
		return costs[selected[0]], true
	}
	return costs[0], true
}

func emitMergedTurnFaceUpRevealEvents(g *game.Game, permanent *game.Permanent) {
	for _, component := range permanent.MergedCards {
		if component.FaceDown {
			continue
		}
		emitEvent(g, game.Event{
			Kind:       game.EventCardRevealed,
			Controller: effectiveController(g, permanent),
			Player:     component.Owner,
			CardID:     component.CardInstanceID,
			Face:       component.Face,
			TokenName:  permanentTokenDefName(component.TokenDef),
			TokenDef:   component.TokenDef,
		})
	}
}

func emitFaceDownRevealEvent(g *game.Game, permanent *game.Permanent) {
	emitEvent(g, game.Event{
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
