package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestFixedPhaseOptionalChoiceOccursAtPrintedTime(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		text   string
		future bool
	}{
		{"Tap target creature. You may exile it at the beginning of the next end step.", false},
		{"Tap target creature. At the beginning of the next end step, you may exile it.", true},
	} {
		t.Run(test.text, func(t *testing.T) {
			t.Parallel()
			sequence := compiledCaptureSequence(t, test.text)
			trigger := sequence[1].Primitive.(game.CreateDelayedTrigger).Trigger
			if sequence[1].Optional == test.future || trigger.Optional != test.future {
				t.Fatal("current and delayed optional choices were conflated")
			}
			for _, acceptNow := range []bool{false, true} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				subject := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
				obj := &game.StackObject{
					Controller: game.Player1,
					Targets:    []game.Target{{Kind: game.TargetPermanent, PermanentID: subject.ObjectID}}, TargetCounts: []int{1},
				}
				engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: acceptNow}}, &TurnLog{})
				wantScheduled := 0
				if acceptNow || test.future {
					wantScheduled = 1
				}
				if len(g.DelayedTriggers) != wantScheduled {
					t.Fatal("optional choice was made at the wrong scheduling time")
				}
				engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: !acceptNow}})
				_, live := permanentByObjectID(g, subject.ObjectID)
				wantExiled := acceptNow
				if test.future {
					wantExiled = !acceptNow
				}
				if live == wantExiled {
					t.Fatal("optional delayed action was decided at the wrong resolving time")
				}
			}
		})
	}
}

func TestFixedPhaseOptionalProducerResultGatesScheduling(t *testing.T) {
	t.Parallel()
	sequence := compiledCaptureSequence(t, "You may create a 1/1 green Insect creature token. If you do, sacrifice it at the beginning of the next end step.")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{Controller: game.Player1, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID}
	for _, accept := range []bool{true, false} {
		engine.resolveInstructionSequence(g, obj, sequence, [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: accept}}, &TurnLog{})
	}
	if len(g.DelayedTriggers) != 1 || len(g.DelayedTriggers[0].CapturedObjectIDs) != 1 {
		t.Fatal("current optional-result gate scheduled a missing/stale producer")
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			t.Fatal("actual optional product escaped its scheduled body")
		}
	}
}
