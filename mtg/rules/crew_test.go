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

func TestCrewUsesContributionModifiersButSaddleDoesNot(t *testing.T) {
	t.Run("crew contribution", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		vehicle := crewVehiclePermanent(g, game.Player1, 3)
		pilot := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:      "Pilot",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{{
				CrewPowerBonus: 2,
			}},
		}})

		if !engine.applyAction(g, game.Player1, action.ActivateAbility(vehicle.ObjectID, 0, nil, 0)) {
			t.Fatal("Crew 3 rejected a 1-power Pilot with +2 crew contribution")
		}
		if !pilot.Tapped {
			t.Fatal("Pilot was not tapped to crew")
		}
	})

	t.Run("saddle uses ordinary power", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		mount := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:               "Mount",
			Types:              []types.Card{types.Creature},
			Power:              opt.Val(game.PT{Value: 3}),
			Toughness:          opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{game.SaddleActivatedAbility(3)},
		}})
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:      "Pilot",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{{
				CrewPowerBonus: 2,
			}},
		}})

		if engine.applyAction(g, game.Player1, action.ActivateAbility(mount.ObjectID, 0, nil, 0)) {
			t.Fatal("Saddle 3 incorrectly used the Pilot's crew-only bonus")
		}
	})

	t.Run("opponent Pilot cannot contribute", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		vehicle := crewVehiclePermanent(g, game.Player1, 3)
		addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name:      "Opponent Pilot",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{{
				CrewPowerBonus: 2,
			}},
		}})
		if engine.applyAction(g, game.Player1, action.ActivateAbility(vehicle.ObjectID, 0, nil, 0)) {
			t.Fatal("Crew used a Pilot the Vehicle's controller does not control")
		}
	})
}

func TestCrewContributionComposesThroughAbilityLayer(t *testing.T) {
	newGame := func() (*game.Game, *Engine, *game.Permanent) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		pilot := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:      "Pilot",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				{CrewPowerBonus: 2},
				{CrewPowerBonus: 1},
			},
		}})
		return g, NewEngine(nil), pilot
	}

	t.Run("multiple modifiers add", func(t *testing.T) {
		g, engine, _ := newGame()
		vehicle := crewVehiclePermanent(g, game.Player1, 4)
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(vehicle.ObjectID, 0, nil, 0)) {
			t.Fatal("Crew 4 rejected cumulative +2 and +1 contribution modifiers")
		}
	})

	t.Run("ability removal clears modifiers", func(t *testing.T) {
		g, engine, pilot := newGame()
		vehicle := crewVehiclePermanent(g, game.Player1, 2)
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID:                 g.IDGen.Next(),
			AffectedObjectID:   pilot.ObjectID,
			Timestamp:          10,
			Layer:              game.LayerAbility,
			RemoveAllAbilities: true,
		})
		if engine.applyAction(g, game.Player1, action.ActivateAbility(vehicle.ObjectID, 0, nil, 0)) {
			t.Fatal("Crew 2 used contribution modifiers removed in the ability layer")
		}
	})

	t.Run("later ability grant contributes", func(t *testing.T) {
		g, engine, pilot := newGame()
		vehicle := crewVehiclePermanent(g, game.Player1, 4)
		granted := game.StaticAbility{CrewPowerBonus: 3}
		g.ContinuousEffects = append(g.ContinuousEffects,
			game.ContinuousEffect{
				ID:                 g.IDGen.Next(),
				AffectedObjectID:   pilot.ObjectID,
				Timestamp:          10,
				Layer:              game.LayerAbility,
				RemoveAllAbilities: true,
			},
			game.ContinuousEffect{
				ID:               g.IDGen.Next(),
				AffectedObjectID: pilot.ObjectID,
				Timestamp:        20,
				Layer:            game.LayerAbility,
				AddAbilities:     []game.Ability{&granted},
			},
		)
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(vehicle.ObjectID, 0, nil, 0)) {
			t.Fatal("Crew 4 ignored a later +3 contribution grant")
		}
	})
}
