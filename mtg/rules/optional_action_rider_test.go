package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestOptionalGroupPreservesIndependentSameSubjectRider(t *testing.T) {
	for _, prefix := range []string{"", "Tap target artifact. "} {
		t.Run(prefix, func(t *testing.T) {
			content := compiledScopedResultContent(t, prefix+
				"You may untap target creature and gain control of it until end of turn. It gains haste until end of turn.")
			for _, accept := range []bool{true, false} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Artifact}}})
				target := addCombatCreaturePermanent(g, game.Player2, game.KeywordNone)
				other := addCombatCreaturePermanent(g, game.Player3, game.KeywordNone)
				target.Tapped = true
				agent := &scopedMayAgent{accept: []bool{accept}}
				agents := [game.NumPlayers]PlayerAgent{}
				agents[game.Player1] = agent
				targets := []game.Target{game.PermanentTarget(target.ObjectID)}
				if prefix != "" {
					targets = append([]game.Target{game.PermanentTarget(artifact.ObjectID)}, targets...)
				}
				obj := &game.StackObject{Controller: game.Player1, Targets: targets}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, agents, &TurnLog{})
				wantController := game.Player2
				if accept {
					wantController = game.Player1
				}
				if target.Tapped == accept || effectiveController(g, target) != wantController ||
					!hasKeyword(g, target, game.Haste) || agent.next != 1 {
					t.Fatalf("accept=%v tapped=%v controller=%v haste=%v choices=%d, want %v/%v/true/1",
						accept, target.Tapped, effectiveController(g, target), hasKeyword(g, target, game.Haste), agent.next,
						!accept, wantController)
				}
				if artifact.Tapped != (prefix != "") || hasKeyword(g, artifact, game.Haste) ||
					hasKeyword(g, other, game.Haste) || effectiveController(g, other) != game.Player3 {
					t.Fatal("shared subject escaped its nonzero target slot")
				}
			}
		})
	}
}
