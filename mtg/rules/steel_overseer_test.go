package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestSteelOverseerCountersOnlyArtifactCreaturesYouControl proves the conjunctive
// "each artifact creature you control" group counter placement (Steel Overseer):
// a +1/+1 counter lands on each permanent that is BOTH an artifact and a creature
// you control, but never on a non-artifact creature, a non-creature artifact, or
// an opponent's artifact creature.
func TestSteelOverseerCountersOnlyArtifactCreaturesYouControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	artifactCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Myr",
		Types: []types.Card{types.Artifact, types.Creature},
	}})
	plainCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	plainArtifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	}})
	opponentArtifactCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Enemy Myr",
		Types: []types.Card{types.Artifact, types.Creature},
	}})

	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Artifact, types.Creature},
			Controller:    game.ControllerYou,
		}),
		Amount:      game.Fixed(1),
		CounterKind: counter.PlusOnePlusOne,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := artifactCreature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("artifact creature you control +1/+1 counters = %d, want 1", got)
	}
	if got := plainCreature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("non-artifact creature +1/+1 counters = %d, want 0", got)
	}
	if got := plainArtifact.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("non-creature artifact +1/+1 counters = %d, want 0", got)
	}
	if got := opponentArtifactCreature.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("opponent's artifact creature +1/+1 counters = %d, want 0", got)
	}
}
