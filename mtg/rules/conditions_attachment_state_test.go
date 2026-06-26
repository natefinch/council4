package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestConditionSourceEquipped verifies the "as long as it's equipped"
// source-state predicate tracks whether an Equipment is attached to the source
// permanent, matching the MatchEquipped selection lowered from cards such as
// Training Drone.
func TestConditionSourceEquipped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	equipment := addEquipmentPermanent(g, game.Player1)
	ctx := conditionContext{controller: game.Player1, source: source}

	equipped := opt.Val(game.Condition{
		Object:        opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{MatchEquipped: true}),
	})

	if conditionSatisfied(g, ctx, equipped) {
		t.Fatal("equipped condition satisfied before any Equipment is attached")
	}

	if !attachPermanent(g, equipment, source) {
		t.Fatal("attachPermanent() = false, want true")
	}

	if !conditionSatisfied(g, ctx, equipped) {
		t.Fatal("equipped condition unsatisfied after Equipment is attached")
	}
}

// TestConditionSourceEnchanted verifies the "as long as it's enchanted"
// source-state predicate tracks whether an Aura is attached to the source
// permanent, matching the MatchEnchanted selection lowered from cards such as
// Fledgling Osprey. It also confirms an attached Equipment does not satisfy the
// enchanted predicate, nor an Aura the equipped predicate.
func TestConditionSourceEnchanted(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanent(g, game.Player1)
	aura := addAuraPermanent(g, game.Player1)
	equipment := addEquipmentPermanent(g, game.Player1)
	ctx := conditionContext{controller: game.Player1, source: source}

	enchanted := opt.Val(game.Condition{
		Object:        opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{MatchEnchanted: true}),
	})
	equipped := opt.Val(game.Condition{
		Object:        opt.Val(game.SourcePermanentReference()),
		ObjectMatches: opt.Val(game.Selection{MatchEquipped: true}),
	})

	if conditionSatisfied(g, ctx, enchanted) {
		t.Fatal("enchanted condition satisfied before any Aura is attached")
	}

	if !attachPermanent(g, equipment, source) {
		t.Fatal("attachPermanent() = false, want true")
	}
	if conditionSatisfied(g, ctx, enchanted) {
		t.Fatal("enchanted condition satisfied by an Equipment attachment")
	}

	if !attachPermanent(g, aura, source) {
		t.Fatal("attachPermanent() = false, want true")
	}
	if !conditionSatisfied(g, ctx, enchanted) {
		t.Fatal("enchanted condition unsatisfied after Aura is attached")
	}
	if !conditionSatisfied(g, ctx, equipped) {
		t.Fatal("equipped condition unsatisfied while both Aura and Equipment are attached")
	}
}
