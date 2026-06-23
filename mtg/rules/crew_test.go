package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// crewVehiclePermanent adds a Crew N Vehicle (printed 4/3 artifact) controlled
// by the player.
func crewVehiclePermanent(g *game.Game, controller game.PlayerID, n int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:               "Test Vehicle",
		Types:              []types.Card{types.Artifact},
		Subtypes:           []types.Sub{types.Vehicle},
		Power:              opt.Val(game.PT{Value: 4}),
		Toughness:          opt.Val(game.PT{Value: 3}),
		ActivatedAbilities: []game.ActivatedAbility{game.CrewActivatedAbility(n)},
	}})
}

func TestCrewTapsCreaturesAndAnimatesVehicleUntilEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vehicle := crewVehiclePermanent(g, game.Player1, 2)
	crewer := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	// Before crewing the Vehicle is not a creature and cannot attack.
	if permanentHasType(g, vehicle, types.Creature) {
		t.Fatal("Vehicle is a creature before being crewed")
	}

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(vehicle.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(crew) = false, want true")
	}
	engine.resolveTopOfStack(g, nil)

	if !crewer.Tapped {
		t.Fatal("crewing creature was not tapped")
	}
	if vehicle.Tapped {
		t.Fatal("Vehicle was tapped by its own crew ability")
	}
	if !permanentHasType(g, vehicle, types.Creature) {
		t.Fatal("crewed Vehicle is not a creature")
	}
	if got := effectivePower(g, vehicle); got != 4 {
		t.Fatalf("crewed Vehicle power = %d, want 4 (its printed power)", got)
	}

	// A crewed Vehicle is a legal attacker for its controller.
	eligible := eligibleAttackers(g, game.Player1)
	found := false
	for _, permanent := range eligible {
		if permanent == vehicle {
			found = true
		}
	}
	if !found {
		t.Fatal("crewed Vehicle is not an eligible attacker")
	}

	// The animation lasts only until end of turn.
	expireCleanupDurations(g)
	if permanentHasType(g, vehicle, types.Creature) {
		t.Fatal("Vehicle is still a creature after end of turn cleanup")
	}
}

func TestCrewCannotActivateWithoutEnoughPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vehicle := crewVehiclePermanent(g, game.Player1, 3)
	// Only one power available, the Crew 3 cost cannot be paid.
	addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.ActivateAbility(vehicle.ObjectID, 0, nil, 0)) {
		t.Fatal("applyAction(crew) = true, want false (insufficient total power)")
	}
	if permanentHasType(g, vehicle, types.Creature) {
		t.Fatal("Vehicle became a creature despite unpaid crew cost")
	}
}
