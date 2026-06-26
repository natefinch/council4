package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

type priorityCastRecord struct {
	player    game.PlayerID
	stackSize int
}

// priorityCastRecorder records which player receives priority and the current
// stack size on each priority grant, and casts a specific spell the first time
// the designated caster is asked to act.
type priorityCastRecorder struct {
	g       *game.Game
	spellID id.ID
	caster  game.PlayerID
	cast    bool
	records []priorityCastRecord
}

func (r *priorityCastRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	r.records = append(r.records, priorityCastRecord{player: obs.Player, stackSize: r.g.Stack.Size()})
	if obs.Player == r.caster && !r.cast {
		for _, act := range legal {
			if payload, ok := act.CastSpellPayload(); ok && payload.CardID == r.spellID {
				r.cast = true
				return act
			}
		}
	}
	return action.Pass()
}

// TestNonactivePlayerKeepsPriorityAfterTheirSpellTriggers covers CR 603.3b: when a
// nonactive player casts a spell (keeping priority under CR 117.3c) and that cast
// triggers an ability, the trigger goes on the stack and that same nonactive
// player receives priority next — not the active player.
func TestNonactivePlayerKeepsPriorityAfterTheirSpellTriggers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// Player2 (nonactive) controls a "whenever you cast a spell, draw a card"
	// permanent and holds a castable instant.
	addTriggeredPermanent(g, game.Player2, &game.TriggerPattern{
		Event:      game.EventSpellCast,
		Controller: game.TriggerControllerYou,
	}, []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spellID := addCardToHand(g, game.Player2, greenInstant())
	addBasicLandPermanent(g, game.Player2, types.Forest)

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	recorder := &priorityCastRecorder{g: g, spellID: spellID, caster: game.Player2}
	agents := [game.NumPlayers]PlayerAgent{recorder, recorder, recorder, recorder}

	engine.runPriorityLoop(g, agents, &TurnLog{})

	if !recorder.cast {
		t.Fatal("nonactive player never cast the triggering spell")
	}
	// The first priority grant after the spell and its trigger are both on the
	// stack (stack size 2) must go to the caster (CR 603.3b), not the active
	// player.
	found := false
	for _, record := range recorder.records {
		if record.stackSize >= 2 {
			if record.player != game.Player2 {
				t.Fatalf("priority after the cast trigger went to Player%d, want the caster Player2", record.player+1)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("never observed the cast spell and its trigger on the stack together")
	}
}
