package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestTapDownSequenceKeepsTargetTappedThroughNextUntapStep verifies the
// tap-down (stun) sequence end to end: resolving "tap target, then skip its
// next untap" leaves the target tapped and exerted, the target stays tapped
// through its controller's next untap step (shedding the exert), and it untaps
// on the following untap step.
func TestTapDownSequenceKeepsTargetTappedThroughNextUntapStep(t *testing.T) {
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
		{Primitive: game.SkipNextUntap{Object: game.TargetPermanentReference(0)}},
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if !target.Tapped {
		t.Fatal("tap-down sequence did not tap its target")
	}
	if !target.Exerted {
		t.Fatal("tap-down sequence did not mark its target to skip its next untap")
	}

	// The target's controller takes their untap step: the target stays tapped
	// and sheds the exert flag.
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw One"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !target.Tapped {
		t.Fatal("target untapped during its skipped untap step")
	}
	if target.Exerted {
		t.Fatal("skip-next-untap flag was not cleared after the skipped untap step")
	}

	// The following untap step untaps the target normally.
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw Two"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if target.Tapped {
		t.Fatal("target did not untap on the untap step after its skipped one")
	}
}

// TestMultiTargetTapStunSequenceStunsEachChosenTarget verifies the multi-target
// tap-stun sequence ("Tap up to two target creatures. Those creatures don't
// untap during their controller's next untap step.") taps and exerts every
// chosen target slot, and that each stays tapped through its controller's next
// untap step.
func TestMultiTargetTapStunSequenceStunsEachChosenTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Stunned Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	second := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Stunned Elk",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.Tap{Object: game.TargetPermanentReference(1)}},
		{Primitive: game.SkipNextUntap{Object: game.TargetPermanentReference(0)}},
		{Primitive: game.SkipNextUntap{Object: game.TargetPermanentReference(1)}},
	}, []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	for i, c := range []*game.Permanent{first, second} {
		if !c.Tapped {
			t.Fatalf("target %d was not tapped by the multi-target tap-stun", i)
		}
		if !c.Exerted {
			t.Fatalf("target %d was not marked to skip its next untap", i)
		}
	}

	// The targets' controller takes their untap step: both stay tapped and shed
	// the exert flag.
	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw One"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	for i, c := range []*game.Permanent{first, second} {
		if !c.Tapped {
			t.Fatalf("target %d untapped during its skipped untap step", i)
		}
		if c.Exerted {
			t.Fatalf("target %d did not shed its skip-next-untap flag", i)
		}
	}

	// The following untap step untaps both targets normally.
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Draw Two"}})
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	for i, c := range []*game.Permanent{first, second} {
		if c.Tapped {
			t.Fatalf("target %d did not untap on the untap step after its skipped one", i)
		}
	}
}
