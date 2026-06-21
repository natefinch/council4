package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addLandPermanent(g *game.Game, controller game.PlayerID, name string, subtypes ...types.Sub) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Land},
		Subtypes: subtypes,
	}})
}

func addLandwalkAttacker(g *game.Game, controller game.PlayerID, body *game.StaticAbility) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:            "Landwalker",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{*body},
	}})
}

// TestForestwalkUnblockableWhenDefenderControlsForest proves CR 702.14c: a
// creature with forestwalk can't be blocked while the defending player controls
// a Forest, and can be blocked otherwise.
func TestForestwalkUnblockableWhenDefenderControlsForest(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addLandwalkAttacker(g, game.Player1, &game.ForestwalkStaticBody)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	// With no Forest under the defender's control, the blocker may block.
	if !canBlockAttacker(g, blocker, attacker) {
		t.Fatal("forestwalker could not be blocked while defender controls no Forest")
	}

	// Defender controls a Forest: the forestwalker is unblockable.
	addLandPermanent(g, game.Player2, "Forest", types.Forest)
	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("forestwalker could be blocked while defender controls a Forest")
	}
}

// TestLandwalkKeyedToDefendingPlayer confirms the rule keys off the defending
// player's lands, not the attacker's: a Forest the attacking player controls
// does not make its own forestwalker unblockable.
func TestLandwalkKeyedToDefendingPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addLandwalkAttacker(g, game.Player1, &game.ForestwalkStaticBody)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addLandPermanent(g, game.Player1, "Attacker Forest", types.Forest)

	if !canBlockAttacker(g, blocker, attacker) {
		t.Fatal("forestwalker became unblockable from a Forest the attacker controls")
	}
}

// TestLandwalkWrongSubtypeDoesNotEvade confirms a typed landwalk keys only off
// its own subtype: an Islandwalker is not evasive against a Forest.
func TestLandwalkWrongSubtypeDoesNotEvade(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addLandwalkAttacker(g, game.Player1, &game.IslandwalkStaticBody)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addLandPermanent(g, game.Player2, "Forest", types.Forest)

	if !canBlockAttacker(g, blocker, attacker) {
		t.Fatal("islandwalker became unblockable while defender controls only a Forest")
	}

	addLandPermanent(g, game.Player2, "Island", types.Island)
	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("islandwalker could be blocked while defender controls an Island")
	}
}

// TestGenericLandwalkAnyLandEvades confirms generic landwalk keys off any land.
func TestGenericLandwalkAnyLandEvades(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addLandwalkAttacker(g, game.Player1, &game.LandwalkStaticBody)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if !canBlockAttacker(g, blocker, attacker) {
		t.Fatal("generic landwalker unblockable while defender controls no land")
	}

	// Any land under the defender's control suffices for generic landwalk.
	addLandPermanent(g, game.Player2, "Wastes")
	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("generic landwalker could be blocked while defender controls a land")
	}
}

// TestLandwalkIgnoresPhasedOutLand confirms a phased-out matching land does not
// grant landwalk evasion: a phased-out permanent is treated as nonexistent
// (CR 702.26e), so the forestwalker remains blockable.
func TestLandwalkIgnoresPhasedOutLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addLandwalkAttacker(g, game.Player1, &game.ForestwalkStaticBody)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	forest := addLandPermanent(g, game.Player2, "Forest", types.Forest)
	forest.PhasedOut = true

	if !canBlockAttacker(g, blocker, attacker) {
		t.Fatal("forestwalker became unblockable from a phased-out Forest")
	}
}
