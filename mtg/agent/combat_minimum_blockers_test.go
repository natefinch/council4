package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/rules"
)

// TestBlocksThreeOrMoreAttackerOnlyWithEnoughCreatures plays the default agent
// through the real Engine against "This creature can't be blocked except by
// three or more creatures." (CR 509.1b). Facing lethal damage, it chump-blocks
// with all three creatures; with only two it cannot block at all.
func TestBlocksThreeOrMoreAttackerOnlyWithEnoughCreatures(t *testing.T) {
	for _, tc := range []struct {
		blockers int
		want     int
	}{{blockers: 3, want: 3}, {blockers: 2, want: 0}} {
		e := searchTestEngine()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player2, game.Player1
		g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareBlockers
		g.Players[game.Player1].Life = 3
		def := creatureCardDef("Pathrazer of Ulamog", 9, 9)
		def.StaticAbilities = []game.StaticAbility{{
			Text: "This creature can't be blocked except by three or more creatures.",
			RuleEffects: []game.RuleEffect{{
				Kind:            game.RuleEffectCantBeBlockedExceptByNOrMore,
				AffectedSource:  true,
				MinimumBlockers: 3,
			}},
		}}
		attacker := addObservedPermanent(g, game.Player2, def)
		declareAttack(g, attacker, game.Player1)
		for range tc.blockers {
			addObservedPermanent(g, game.Player1, creatureCardDef("Eldrazi Spawn", 1, 1))
		}

		sim := e.Simulator()
		player := Agent{Strategy: GenericStrategy{}}
		legal := sim.LegalActions(g, game.Player1)
		chosen := player.ChooseAction(rules.NewObservation(g, game.Player1), legal)
		var policies [game.NumPlayers]rules.PlayerAgent
		for i := range policies {
			policies[i] = player
		}
		result, ok := sim.Apply(g, game.Player1, chosen, policies)
		if !ok {
			t.Fatalf("%d blockers: engine rejected the agent's declaration %v", tc.blockers, chosen)
		}
		got := 0
		for _, block := range result.Combat.Blockers {
			if block.Blocking == attacker.ObjectID {
				got++
			}
		}
		if got != tc.want {
			t.Fatalf("%d blockers: agent blocked with %d creatures, want %d", tc.blockers, got, tc.want)
		}
	}
}
