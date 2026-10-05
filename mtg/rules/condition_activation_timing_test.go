package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func TestCompiledBodyConditionVersusActivationRestrictionTiming(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                   string
		restriction            bool
		initialHand            int
		addCardBeforeResolving bool
		wantLegal              bool
		wantHand               int
	}{
		{"body checks after cost", false, 1, false, true, 2},
		{"body checks at resolution", false, 1, true, true, 1},
		{"restriction checks before cost", true, 1, false, false, 1},
		{"restriction not checked again at resolution", true, 0, true, true, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := "If you have no cards in hand, draw a card, then draw a card."
			if tt.restriction {
				body = "Draw a card, then draw a card. Activate only if you have no cards in hand."
			}
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Activation Timing", Layout: "normal", TypeLine: "Artifact",
				OracleText: "Discard your hand: " + body,
			})
			ability := &def.ActivatedAbilities[0]
			if ability.ActivationCondition.Exists != tt.restriction {
				t.Fatal("typed body scope and activation-only scope were conflated")
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone
			source := addCombatPermanent(g, game.Player1, def)
			for range tt.initialHand {
				addCardToHand(g, game.Player1, evidenceCard("Initial", 1))
			}
			for range 3 {
				addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
			}
			if legal := canActivateGeneralAbility(g, game.Player1, source, ability, 0, nil, 0); legal != tt.wantLegal {
				t.Fatalf("activation legal=%v, want %v", legal, tt.wantLegal)
			}
			engine := NewEngine(nil)
			if applied := engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)); applied != tt.wantLegal {
				t.Fatalf("activation applied=%v, want %v", applied, tt.wantLegal)
			}
			if tt.wantLegal {
				if g.Stack.Size() != 1 || g.Players[game.Player1].Hand.Size() != 0 {
					t.Fatal("activation must pay the discard cost before putting the ability on the stack")
				}
				if tt.addCardBeforeResolving {
					addCardToHand(g, game.Player1, evidenceCard("Interposed", 1))
				}
				engine.resolveTopOfStack(g, &TurnLog{})
			} else if !g.Stack.IsEmpty() || g.Players[game.Player1].Graveyard.Size() != 0 {
				t.Fatal("illegal activation must neither pay costs nor use the stack")
			}
			if got := g.Players[game.Player1].Hand.Size(); got != tt.wantHand {
				t.Fatalf("hand=%d, want %d", got, tt.wantHand)
			}
		})
	}
}
