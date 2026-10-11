package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestCompiledReferencedPlayerActionResult(t *testing.T) {
	for _, verb := range []string{"discards", "mills"} {
		for _, tc := range []struct {
			name                      string
			empty, optional, declined bool
			skipped, prefixed         bool
		}{
			{name: "success"},
			{name: "empty", empty: true},
			{name: "optional accepted", optional: true},
			{name: "optional declined", optional: true, declined: true},
			{name: "optional empty", optional: true, empty: true},
			{name: "skipped", optional: true, skipped: true},
			{name: "nonzero player and card targets", prefixed: true},
		} {
			t.Run(verb+"/"+tc.name, func(t *testing.T) {
				action := "That player " + verb
				if tc.optional {
					action = "That player may " + map[string]string{"discards": "discard", "mills": "mill"}[verb]
				}
				if tc.skipped {
					action = "If you control a creature, " + action
				}
				text := "Target player gains 1 life. " + action + " a card. You gain 1 life. If that player does, you gain 2 life."
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				targets := []game.Target{game.PlayerTarget(game.Player2)}
				counts := []int{1}
				if tc.prefixed {
					text = "Target player gains 3 life. Tap target creature. " + text
					permanent := addCombatCreaturePermanent(g, game.Player4)
					targets = []game.Target{game.PlayerTarget(game.Player4), game.PermanentTarget(permanent.ObjectID), game.PlayerTarget(game.Player2)}
					counts = []int{1, 1, 1}
				}
				def := compileCounterTaxCard(t, "Referenced Actor", text, "Instant")
				actionIndex := 1
				if tc.prefixed {
					actionIndex = 3
				}
				key := string(def.SpellAbility.Val.Modes[0].Sequence[actionIndex].PublishResult)
				if key == "" {
					t.Fatal("actual action lacks a publication")
				}
				if !tc.empty {
					if verb == "discards" {
						addCardToHand(g, game.Player2, vanillaCreature("Discard", 1, 1))
					} else {
						addCardToLibrary(g, game.Player2, vanillaCreature("Mill", 1, 1))
					}
				}
				source := addCardToHand(g, game.Player3, def)
				g.Players[game.Player3].Hand.Remove(source)
				obj := &game.StackObject{
					ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source, Controller: game.Player1,
					Targets: targets, TargetCounts: counts,
					ResolutionResults: map[string]game.InstructionResolutionResult{
						key: {Succeeded: true, Accepted: true},
					},
				}
				g.Stack.Push(obj)
				agents := [game.NumPlayers]PlayerAgent{}
				agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{0}}}
				if tc.declined {
					agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
				}
				log := TurnLog{}
				engine := NewEngine(nil)
				engine.resolveTopOfStackWithChoices(g, agents, &log)
				success := !tc.empty && !tc.declined && !tc.skipped
				wantLife, wantMoved := 41, 0
				if success {
					wantLife, wantMoved = 43, 1
				}
				if g.Players[game.Player1].Life != wantLife || g.Players[game.Player2].Life != 41 ||
					g.Players[game.Player3].Life != 40 {
					t.Fatalf("actor result borrowed Gain, decider or source owner: life=%d/%d/%d",
						g.Players[game.Player1].Life, g.Players[game.Player2].Life, g.Players[game.Player3].Life)
				}
				if g.Players[game.Player2].Graveyard.Size() != wantMoved {
					t.Fatal("action did not affect the exact target player")
				}
				if tc.skipped && len(log.Choices) != 0 {
					t.Fatal("skipped action offered a choice")
				}
				if len(obj.ResolutionResults) != 1 || !obj.ResolutionResults[key].Succeeded {
					t.Fatal("local result did not restore the enclosing frame")
				}
				if success {
					before := g.Players[game.Player1].Life
					engine.resolveAbilityContentWithChoices(g, obj, def.SpellAbility.Val, agents, &TurnLog{})
					if g.Players[game.Player1].Life != before+1 {
						t.Fatal("repeated empty action reused the preceding successful receipt")
					}
					if len(obj.ResolutionResults) != 1 || !obj.ResolutionResults[key].Succeeded {
						t.Fatal("repeat did not restore the enclosing frame")
					}
				}
			})
		}
	}
}
