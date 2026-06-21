package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// choiceCostSpell builds an instant whose only additional cost is the printed
// choice "sacrifice an artifact or discard a card", with both alternatives
// tagged into one choice group.
func choiceCostSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Demand Answers",
		ManaCost: opt.Val(cost.Mana{cost.R}),
		Types:    []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalSacrifice, Text: "sacrifice an artifact", Amount: 1, MatchPermanentType: true, PermanentType: types.Artifact, ChoiceGroup: 1},
			{Kind: cost.AdditionalDiscard, Text: "discard a card", Amount: 1, ChoiceGroup: 1},
		},
		SpellAbility: opt.Val(game.AbilityContent{})},
	}
}

func TestCastSpellChoiceCostPaysFirstPayableAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, choiceCostSpell())
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spare Card",
		Types: []types.Card{types.Sorcery}},
	})
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Clue Token",
		Types: []types.Card{types.Artifact}},
	})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell with payable choice cost was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with choice cost) = false, want true")
	}
	// The sacrifice alternative is the first payable choice, so it is taken.
	if _, ok := permanentByObjectID(g, artifact.ObjectID); ok {
		t.Fatal("artifact was not sacrificed for the first payable choice")
	}
	if !g.Players[game.Player1].Graveyard.Contains(artifact.CardInstanceID) {
		t.Fatal("sacrificed artifact was not put into graveyard")
	}
}

func TestCastSpellChoiceCostFallsBackToSecondAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, choiceCostSpell())
	spareID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spare Card",
		Types: []types.Card{types.Sorcery}},
	})
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	act := action.CastSpell(spellID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spell with only the discard alternative payable was not legal")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast with discard fallback) = false, want true")
	}
	// No artifact exists, so the discard alternative must be chosen instead.
	if !g.Players[game.Player1].Graveyard.Contains(spareID) {
		t.Fatal("spare card was not discarded for the fallback choice")
	}
	if g.Players[game.Player1].Hand.Contains(spareID) {
		t.Fatal("discarded card remained in hand")
	}
}
