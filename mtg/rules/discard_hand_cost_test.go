package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// discardHandAbilityDef builds a permanent whose only activated ability costs
// "Discard your hand" (modeled as a dynamic hand-size discard) and gains life.
func discardHandAbilityDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Hand Sink",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind:          cost.AdditionalDiscard,
				Text:          "Discard your hand",
				Source:        zone.Hand,
				AmountDynamic: cost.AdditionalDynamicHandSize,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}}
}

func TestActivatedAbilityDiscardWholeHandCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, discardHandAbilityDef())
	for range 3 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spare Card", Types: []types.Card{types.Instant}}})
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("hand size before activation = %d, want 3", got)
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(discard-hand ability) = false, want true")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size after activation = %d, want 0", got)
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 3 {
		t.Fatalf("graveyard size after activation = %d, want 3", got)
	}
}

func TestActivatedAbilityDiscardWholeHandCostWithEmptyHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, discardHandAbilityDef())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(discard-hand ability with empty hand) = false, want true")
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 0 {
		t.Fatalf("graveyard size = %d, want 0 for an empty-hand discard", got)
	}
}
