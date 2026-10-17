package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCardExecutionCopiedEntryAndEventTimeTriggersTogether(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	for _, suppressed := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := permanentCopyEngine(t)
		addReplacementPermanent(t, g, game.Player1, registry.Lookup("Doubling Season"))
		addReplacementPermanent(t, g, game.Player1, registry.Lookup("Triskelion"))
		witness := addCombatPermanent(g, game.Player1, registry.Lookup("Soul Warden"))
		doubler := addCombatPermanent(g, game.Player1, registry.Lookup("Panharmonicon"))
		var suppressor *game.Permanent
		if suppressed {
			suppressor = addCombatPermanent(g, game.Player2, registry.Lookup("Elesh Norn, Mother of Machines"))
		}
		addImplementationSpellToStack(g, game.Player1, registry.Lookup("Clone"), nil)
		original, _ := g.Stack.Peek()
		copyPermanentSpell(t, e, g, original, game.Player1)
		if !counterStackObject(g, original.ID) {
			t.Fatal("could not remove the original Clone")
		}
		g.TriggerEventCursor = len(g.Events)
		pending := g.Clone()
		policies := func() [game.NumPlayers]PlayerAgent {
			agents := simPassPolicies()
			agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{1}, {0}}}
			return agents
		}
		start := len(g.Events)
		e.resolveTopOfStackWithChoices(g, policies(), &TurnLog{})
		copied := g.Battlefield[len(g.Battlefield)-1]
		if !copied.Token || permanentEffectiveName(g, copied) != "Triskelion" ||
			copied.Counters.Get(counter.PlusOnePlusOne) != 6 || copied.EnteredFromCast {
			t.Fatal("spell-copy entry skipped copied replacements or fabricated a cast")
		}
		arrivals := 0
		for _, event := range g.Events[start:] {
			if event.Kind == game.EventTokenCreated || event.Kind == game.EventSpellCast {
				t.Fatal("spell-copy entry went through creation/cast semantics")
			}
			if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == copied.ObjectID {
				arrivals++
			}
		}
		if arrivals != 1 {
			t.Fatalf("copied entry arrived %d times, want once", arrivals)
		}
		movePermanentToZone(g, witness, zone.Library)
		movePermanentToZone(g, doubler, zone.Graveyard)
		if suppressor != nil {
			movePermanentToZone(g, suppressor, zone.Graveyard)
		}
		want := 2
		if suppressed {
			want = 0
		}
		e.putTriggeredAbilitiesOnStack(g)
		if g.Stack.Size() != want {
			t.Fatalf("suppressed=%v: %d triggers after source/doubler/suppressor departures, want %d", suppressed, g.Stack.Size(), want)
		}
		for range want {
			e.resolveTopOfStack(g, &TurnLog{})
		}
		branch := e.Simulator().ResolvePriority(pending, policies())
		if g.Players[game.Player1].Life != 40+want || branch.Players[game.Player1].Life != 40+want ||
			pending.Players[game.Player1].Life != 40 || pending.Stack.Size() != 1 {
			t.Fatal("combined execution diverged across direct and Simulator/clone branches")
		}
	}
}
