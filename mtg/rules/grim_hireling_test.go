package rules

import (
	"testing"

	cardsg "github.com/natefinch/council4/mtg/cards/g"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// addTreasureTokenPermanent puts a Treasure artifact token onto the battlefield
// under the controller so the "Sacrifice X Treasures" additional cost has real
// permanents to sacrifice through the ordinary payment-planning path.
func addTreasureTokenPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      controller,
		Controller: controller,
		Token:      true,
		TokenDef: &game.CardDef{CardFace: game.CardFace{
			Name:     string(types.Treasure),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
		}},
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func countTreasureTokens(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == controller && permanentHasSubtype(g, permanent, types.Treasure) {
			count++
		}
	}
	return count
}

// TestGrimHirelingSacrificeXTreasuresScalesMinusXMinusX proves the variable
// "Sacrifice X Treasures" activation cost feeding an X-scaled effect: the player
// chooses how many Treasures to sacrifice (X), exactly that many are sacrificed
// as the additional cost, and the target creature gets -X/-X until end of turn.
// X is bounded by the Treasures available, and because Grim Hireling has no "one
// or more" wording, X=0 is a legal (do-nothing) activation.
func TestGrimHirelingSacrificeXTreasuresScalesMinusXMinusX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hireling := addCombatPermanent(g, game.Player1, cardsg.GrimHireling())
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	for range 3 {
		addTreasureTokenPermanent(g, game.Player1)
	}
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	setMainPhasePriority(g, game.Player1)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}

	// X cannot exceed the Treasures available (only 3 present).
	if containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(hireling.ObjectID, 0, targets, 4)) {
		t.Fatal("X=4 activation was legal with only 3 Treasures present")
	}
	// No "one or more" wording: X=0 is a legal (pointless) activation.
	if !containsAction(engine.legalActions(g, game.Player1), action.ActivateAbility(hireling.ObjectID, 0, targets, 0)) {
		t.Fatal("X=0 activation was not legal; Grim Hireling has no 'one or more' requirement")
	}

	act := action.ActivateAbility(hireling.ObjectID, 0, targets, 2)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("sacrificing two Treasures (X=2) was not a legal activation")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(activate Grim Hireling) = false, want true")
	}
	// Exactly two Treasures are sacrificed as the additional cost when the
	// ability goes on the stack, before it resolves.
	if got := countTreasureTokens(g, game.Player1); got != 1 {
		t.Fatalf("Treasures after paying cost = %d, want 1 (3 - 2 sacrificed)", got)
	}

	engine.resolveTopOfStack(g, nil)

	if got := effectivePower(g, target); got != 3 {
		t.Fatalf("target power after -2/-2 = %d, want 3 (5 - 2)", got)
	}
	toughness, ok := effectiveToughness(g, target)
	if !ok || toughness != 3 {
		t.Fatalf("target toughness after -2/-2 = %d (ok=%v), want 3 (5 - 2)", toughness, ok)
	}
}

// TestGrimHirelingSacrificeZeroTreasuresDoesNothing proves the X=0 boundary:
// activating with X=0 sacrifices no Treasures and applies -0/-0, leaving the
// target creature's power and toughness unchanged.
func TestGrimHirelingSacrificeZeroTreasuresDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hireling := addCombatPermanent(g, game.Player1, cardsg.GrimHireling())
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	addTreasureTokenPermanent(g, game.Player1)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	setMainPhasePriority(g, game.Player1)

	targets := []game.Target{game.PermanentTarget(target.ObjectID)}
	act := action.ActivateAbility(hireling.ObjectID, 0, targets, 0)
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(activate Grim Hireling with X=0) = false, want true")
	}
	if got := countTreasureTokens(g, game.Player1); got != 1 {
		t.Fatalf("Treasures after X=0 activation = %d, want 1 (none sacrificed)", got)
	}

	engine.resolveTopOfStack(g, nil)

	if got := effectivePower(g, target); got != 5 {
		t.Fatalf("target power after -0/-0 = %d, want 5 (unchanged)", got)
	}
	toughness, ok := effectiveToughness(g, target)
	if !ok || toughness != 5 {
		t.Fatalf("target toughness after -0/-0 = %d (ok=%v), want 5 (unchanged)", toughness, ok)
	}
}
