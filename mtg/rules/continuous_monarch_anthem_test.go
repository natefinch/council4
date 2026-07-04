package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// monarchHexproofAnthem mirrors the lowered shape of Dawnglade Regent's "As long
// as you're the monarch, permanents you control have hexproof.": a StaticAbility
// whose Condition tests the controller's monarch designation and whose continuous
// effect grants Hexproof to the controller's other permanents.
func monarchHexproofAnthem(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Monarch Regent",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{ControllerIsMonarch: true}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{}),
				AddKeywords: []game.Keyword{game.Hexproof},
			}},
		}},
	}})
}

// TestMonarchHexproofAnthemGatedByDesignation confirms the "as long as you're the
// monarch" group anthem grants hexproof to the controller's permanents only while
// the controller holds the monarch designation, and stops when it is lost.
func TestMonarchHexproofAnthemGatedByDesignation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := monarchHexproofAnthem(g, game.Player1)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	// The controller is not the monarch: no hexproof.
	if hasKeyword(g, other, game.Hexproof) {
		t.Fatal("permanent gained hexproof while controller was not the monarch")
	}

	// Grant the monarch designation: the anthem turns on.
	g.Players[game.Player1].IsMonarch = true
	if !hasKeyword(g, other, game.Hexproof) {
		t.Fatal("permanent lacks hexproof while controller is the monarch")
	}
	if !hasKeyword(g, source, game.Hexproof) {
		t.Fatal("the source permanent should also gain hexproof from the anthem")
	}

	// Lose the monarch designation: the anthem turns off again.
	g.Players[game.Player1].IsMonarch = false
	if hasKeyword(g, other, game.Hexproof) {
		t.Fatal("permanent kept hexproof after the controller lost the monarch")
	}
}
