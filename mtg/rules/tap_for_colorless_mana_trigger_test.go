package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func tapForColorlessManaTriggerSource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Forsaken Monument",
		Types: []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                    game.EventPermanentTapped,
					Controller:               game.TriggerControllerYou,
					RequireTappedForMana:     true,
					RequireProducedManaColor: mana.C,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.C},
			}}}.Ability(),
		}},
	}}
}

func colorlessManaLand(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:          "Wastes",
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.C)},
	}})
}

// TestTapForColorlessManaTriggerAddsAdditionalC proves the "Whenever you tap a
// permanent for {C}, add an additional {C}" trigger (Forsaken Monument): tapping
// a permanent that produces colorless mana yields the land's {C} plus the
// trigger's additional {C}.
func TestTapForColorlessManaTriggerAddsAdditionalC(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, tapForColorlessManaTriggerSource())
	land := colorlessManaLand(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(land.ObjectID, 0, nil, 0)) {
		t.Fatal("activating colorless mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 2 {
		t.Fatalf("colorless mana = %d, want 2 (1 from land + 1 from trigger)", got)
	}
}

// TestTapForColorlessManaTriggerIgnoresColoredTaps proves the produced-mana
// color filter: tapping a permanent for colored mana does not fire the "tap a
// permanent for {C}" trigger, so no additional {C} is added.
func TestTapForColorlessManaTriggerIgnoresColoredTaps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, tapForColorlessManaTriggerSource())
	forest := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Forest",
		Types:         []types.Card{types.Land},
		Subtypes:      []types.Sub{types.Forest},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.G)},
	}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(forest.ObjectID, 0, nil, 0)) {
		t.Fatal("activating forest mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 0 {
		t.Fatalf("colorless mana = %d, want 0 (a colored tap must not fire the {C} trigger)", got)
	}
}
