package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestTapStunCounterSequenceKeepsTargetTappedThroughNextUntapStep verifies the
// modern tap-and-stun sequence end to end: resolving "tap target and put a stun
// counter on it" leaves the target tapped with one stun counter, the stun
// counter (not the untap step) keeps it tapped while one counter is removed in
// its place, and the target untaps normally once the counter is gone.
func TestTapStunCounterSequenceKeepsTargetTappedThroughNextUntapStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Stunned Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.AddCounter{
			Amount:      game.Fixed(1),
			Object:      game.TargetPermanentReference(0),
			CounterKind: counter.Stun,
		}},
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if !target.Tapped {
		t.Fatal("tap-and-stun sequence did not tap its target")
	}
	if got := target.Counters.Get(counter.Stun); got != 1 {
		t.Fatalf("stun counters after resolution = %d, want 1", got)
	}

	// The target's controller takes their untap step: the target stays tapped
	// and one stun counter is removed in place of untapping.
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw One"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !target.Tapped {
		t.Fatal("target untapped while it still carried a stun counter")
	}
	if got := target.Counters.Get(counter.Stun); got != 0 {
		t.Fatalf("stun counters after untap step = %d, want 0", got)
	}

	// The following untap step untaps the target normally now that the stun
	// counter is gone.
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw Two"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if target.Tapped {
		t.Fatal("target did not untap after its stun counter was removed")
	}
}

// TestStunCounterRemovedOnlyWhenPermanentWouldUntap verifies that an already
// untapped permanent keeps its stun counters through the untap step, because a
// stun counter only replaces an actual untapping (CR 122.6f).
func TestStunCounterRemovedOnlyWhenPermanentWouldUntap(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Resting Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	permanent.Tapped = false
	permanent.Counters.Add(counter.Stun, 2)

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw One"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := permanent.Counters.Get(counter.Stun); got != 2 {
		t.Fatalf("stun counters on untapped permanent = %d, want 2", got)
	}
}
