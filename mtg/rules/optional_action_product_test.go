package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestOptionalGroupRiderUsesActualEnteredObject(t *testing.T) {
	for _, tt := range []struct {
		name, first string
		wantLife    int
	}{
		{"ordinary group", "gain 1 life", 42},
		{"ineffective first", "discard a card", 40},
		{"expanded first", "add {R}{G}", 40},
	} {
		t.Run(tt.name, func(t *testing.T) {
			content := compiledScopedResultContent(t, "You may "+tt.first+
				" and return target creature card from your graveyard to the battlefield. It gains haste until end of turn.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Artifact}}})
			cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
			agent := &scopedMayAgent{accept: []bool{true, false, true}}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player1] = agent
			obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1,
				Targets: []game.Target{currentCardTarget(t, g, cardID)}}
			engine := NewEngine(nil)
			engine.resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			entered, ok := reanimatedPermanent(g, cardID)
			if !ok || !hasKeyword(g, entered, game.Haste) || agent.next != 1 {
				t.Fatal("mandatory rider did not address the actual entered object")
			}
			effectCount := len(g.ContinuousEffects)
			engine.resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			if len(g.ContinuousEffects) != effectCount || agent.next != 2 {
				t.Fatal("declined group reused a prior entered-object publication")
			}
			engine.resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
			if len(g.ContinuousEffects) != effectCount || agent.next != 3 || g.Players[game.Player1].Life != tt.wantLife {
				t.Fatal("ineffective return reused a prior entered-object publication")
			}
		})
	}
}

func TestOptionalGroupFalseConditionInvalidatesEnteredObject(t *testing.T) {
	content := compiledScopedResultContent(t,
		"If you have no cards in hand, you may gain 1 life and return target creature card from your graveyard to the battlefield. It gains haste until end of turn.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Artifact}}})
	cardID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Creature}}})
	agent := &scopedMayAgent{accept: []bool{true}}
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player1] = agent
	obj := &game.StackObject{SourceID: source.ObjectID, Controller: game.Player1,
		Targets: []game.Target{currentCardTarget(t, g, cardID)}}
	engine := NewEngine(nil)
	engine.resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
	entered, ok := reanimatedPermanent(g, cardID)
	if !ok || !hasKeyword(g, entered, game.Haste) {
		t.Fatal("conditioned group's rider did not address the entered object")
	}
	effectCount := len(g.ContinuousEffects)
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	engine.resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
	if len(g.ContinuousEffects) != effectCount || agent.next != 1 || g.Players[game.Player1].Life != 41 {
		t.Fatal("false group condition reused an earlier publication or prompted")
	}
}
