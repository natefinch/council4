package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestPlayerControlledGroupCounterPlacementHonorsTargetPlayer resolves the
// "Put a +1/+1 counter on each creature target player controls." shape as a
// group AddCounter over a player-controlled group anchored to the spell's single
// player target. The counter must land on every creature the targeted player
// controls and skip creatures controlled by anyone else, so the resolver
// enumerates by the targeted player rather than the spell's controller.
func TestPlayerControlledGroupCounterPlacementHonorsTargetPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	targetCreature := addCombatPermanent(g, game.Player2, plainCreatureDef("Target Player Creature"))
	targetArtifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Target Player Artifact",
		Types: []types.Card{types.Artifact},
	}})
	controllerCreature := addCombatPermanent(g, game.Player1, plainCreatureDef("Caster Creature"))

	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Amount:      game.Fixed(1),
		CounterKind: counter.PlusOnePlusOne,
		Group: game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}),
	}, []game.Target{game.PlayerTarget(game.Player2)})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := targetCreature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("targeted player's creature got %d counters, want 1", got)
	}
	if got := targetArtifact.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("targeted player's non-creature got %d counters, want 0", got)
	}
	if got := controllerCreature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("caster's creature got %d counters, want 0", got)
	}
}

func plainCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
	}}
}
