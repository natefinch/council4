package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
)

func sacrificeDrawAbility() *game.ActivatedAbility {
	return &game.ActivatedAbility{
		AdditionalCosts: append(append([]cost.Additional(nil), cost.Tap...),
			cost.Additional{Kind: cost.AdditionalSacrifice, Amount: 1}),
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
}

func tapDrawAbility() *game.ActivatedAbility {
	return &game.ActivatedAbility{
		AdditionalCosts: cost.Tap,
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
}

func TestActivatedAbilityProfileFlagsResourceSpending(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	sacSource := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(sacrificeDrawAbility()))
	tapSource := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(tapDrawAbility()))
	obs := PlayerObservation{g: g, Player: game.Player1}

	sac, ok := obs.ActivatedAbilityProfile(action.ActivateAbility(sacSource.ObjectID, 0, nil, 0))
	if !ok || !sac.SpendsOwnResources {
		t.Fatalf("sacrifice ability profile = %+v ok=%v, want SpendsOwnResources", sac, ok)
	}

	plain, ok := obs.ActivatedAbilityProfile(action.ActivateAbility(tapSource.ObjectID, 0, nil, 0))
	if !ok || plain.SpendsOwnResources {
		t.Fatalf("tap ability profile = %+v ok=%v, want not SpendsOwnResources", plain, ok)
	}
}

func TestActivatedAbilityProfileRejectsNonAbilityAction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obs := PlayerObservation{g: g, Player: game.Player1}
	if _, ok := obs.ActivatedAbilityProfile(action.Pass()); ok {
		t.Fatal("ActivatedAbilityProfile(Pass) ok = true, want false")
	}
}
