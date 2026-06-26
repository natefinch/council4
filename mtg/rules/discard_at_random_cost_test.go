package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// discardAtRandomAbilityDef builds a permanent whose only activated ability
// costs "Discard a card at random" and gains life, modeling the corpus pattern
// unlocked for issue #1983.
func discardAtRandomAbilityDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Random Sink",
		Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			AdditionalCosts: []cost.Additional{{
				Kind:   cost.AdditionalDiscard,
				Text:   "Discard a card at random",
				Source: zone.Hand,
				Amount: 1,
				Random: true,
			}},
			Content: game.Mode{
				Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
			}.Ability(),
		}},
	}}
}

func TestActivatedAbilityDiscardAtRandomCost(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, discardAtRandomAbilityDef())
	for range 3 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spare Card", Types: []types.Card{types.Instant}}})
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("hand size before activation = %d, want 3", got)
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(discard-at-random ability) = false, want true")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size after activation = %d, want 2", got)
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 1 {
		t.Fatalf("graveyard size after activation = %d, want 1", got)
	}
}

func TestActivatedAbilityDiscardAtRandomCostWithEmptyHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, discardAtRandomAbilityDef())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(discard-at-random with empty hand) = true, want false")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size after failed activation = %d, want 0", got)
	}
}
