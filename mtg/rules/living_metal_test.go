package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestLivingMetalGrantsCreatureTypeOnControllerTurn drives real continuous-layer
// evaluation to prove the Transformers "Living metal" static (game.LivingMetalStaticBody
// on the back-face Vehicle) makes the permanent a creature only while its
// controller is the active player, mirroring "During your turn, this Vehicle is
// also a creature."
func TestLivingMetalGrantsCreatureTypeOnControllerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardInstance(g, game.Player1, livingMetalVehicle())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceBack,
		Transformed:    true,
	}
	g.Battlefield = append(g.Battlefield, permanent)

	g.Turn.ActivePlayer = game.Player1
	if !permanentHasType(g, permanent, types.Creature) {
		t.Fatal("Living metal Vehicle is not a creature on its controller's turn, want creature")
	}

	g.Turn.ActivePlayer = game.Player2
	if permanentHasType(g, permanent, types.Creature) {
		t.Fatal("Living metal Vehicle is a creature on an opponent's turn, want non-creature")
	}
}

func livingMetalVehicle() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Robot Front",

		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Robot},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})}, Layout: game.LayoutTransform,

		Back: opt.Val(game.CardFace{
			Name:            "Vehicle Back",
			Types:           []types.Card{types.Artifact},
			Subtypes:        []types.Sub{types.Vehicle},
			Power:           opt.Val(game.PT{Value: 6}),
			Toughness:       opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{game.LivingMetalStaticBody},
		}),
	}
}
