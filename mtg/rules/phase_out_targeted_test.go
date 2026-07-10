package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// targetedPhaseOut bundles the game, engine, resolving stack object, and chosen
// target permanents for a targeted phase-out resolution.
type targetedPhaseOut struct {
	game       *game.Game
	engine     *Engine
	obj        *game.StackObject
	permanents []*game.Permanent
}

// newTargetedPhaseOutGame builds a game with count tapped creatures controlled
// by Player1 and a stack object whose chosen targets are those creatures, so the
// targeted phase-out references (TargetPermanentReference /
// AllTargetPermanentsReference) resolve the way the cardgen phase-out lowerers
// wire them for "Target creature phases out." and "Any number of target ...
// permanents you control phase out.".
func newTargetedPhaseOutGame(t *testing.T, count int) targetedPhaseOut {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanents := make([]*game.Permanent, 0, count)
	targets := make([]game.Target, 0, count)
	for range count {
		permanent := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  "Phased Creature",
			Types: []types.Card{types.Creature},
		}})
		permanent.Tapped = true
		permanents = append(permanents, permanent)
		targets = append(targets, game.PermanentTarget(permanent.ObjectID))
	}
	obj := &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Phase Out Spell"}}),
		Controller:   game.Player1,
		Targets:      targets,
		TargetCounts: []int{count},
	}
	return targetedPhaseOut{game: g, engine: engine, obj: obj, permanents: permanents}
}

func assertPhasedOutFor(t *testing.T, permanents []*game.Permanent, player game.PlayerID) {
	t.Helper()
	for _, permanent := range permanents {
		if !permanent.PhasedOut || !permanent.PhaseInScheduled || permanent.PhasedOutFor != player {
			t.Fatalf("permanent %d phase state = %+v, want phased out for %v", permanent.ObjectID, permanent, player)
		}
	}
}

// TestAnyNumberTargetPhaseOutPhasesEveryChosenPermanent covers the unbounded
// "Any number of target nonland permanents you control phase out." lowering
// (Clever Concealment): a single PhaseOut over AllTargetPermanentsReference phases
// out every chosen target, and all of them phase back in before their
// controller's next untap step.
func TestAnyNumberTargetPhaseOutPhasesEveryChosenPermanent(t *testing.T) {
	env := newTargetedPhaseOutGame(t, 3)
	g, engine, obj, permanents := env.game, env.engine, env.obj, env.permanents

	resolveInstruction(engine, g, obj, game.PhaseOut{
		Object: game.AllTargetPermanentsReference(0),
	}, nil)
	assertPhasedOutFor(t, permanents, game.Player1)

	g.Turn.ActivePlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, permanent := range permanents {
		if permanent.PhasedOut {
			t.Fatalf("permanent %d did not phase in at controller's next untap", permanent.ObjectID)
		}
		if permanent.Tapped {
			t.Fatalf("phased-in permanent %d controlled by the active player did not untap", permanent.ObjectID)
		}
	}
}

// TestAnyNumberTargetPhaseOutAllowsZeroTargets covers the "any number of target"
// minimum of zero: with no chosen targets the single all-targets PhaseOut resolves
// to no permanents and does nothing rather than panicking or phasing an unrelated
// permanent.
func TestAnyNumberTargetPhaseOutAllowsZeroTargets(t *testing.T) {
	env := newTargetedPhaseOutGame(t, 0)
	g, engine, obj := env.game, env.engine, env.obj
	bystander := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bystander",
		Types: []types.Card{types.Creature},
	}})

	startEvents := len(g.Events)
	resolveInstruction(engine, g, obj, game.PhaseOut{
		Object: game.AllTargetPermanentsReference(0),
	}, nil)

	if bystander.PhasedOut {
		t.Fatal("zero-target phase out phased out an unrelated permanent")
	}
	if len(g.Events) != startEvents {
		t.Fatalf("zero-target phase out emitted %d events, want none", len(g.Events)-startEvents)
	}
}

// TestSingleTargetPhaseOutPhasesChosenPermanent covers the exact single-target
// "Target creature phases out." lowering (Reality Ripple, Vodalian Illusionist)
// and the resolved slot of an "up to one target ... phases out." unroll: the
// chosen permanent phases out and phases back in at its controller's next untap.
func TestSingleTargetPhaseOutPhasesChosenPermanent(t *testing.T) {
	env := newTargetedPhaseOutGame(t, 1)
	g, engine, obj, permanents := env.game, env.engine, env.obj, env.permanents

	resolveInstruction(engine, g, obj, game.PhaseOut{
		Object: game.TargetPermanentReference(0),
	}, nil)
	assertPhasedOutFor(t, permanents, game.Player1)

	g.Turn.ActivePlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if permanents[0].PhasedOut {
		t.Fatal("single-target phased-out permanent did not phase in at controller's next untap")
	}
	if permanents[0].Tapped {
		t.Fatal("phased-in permanent controlled by the active player did not untap")
	}
}

// TestUpToOneTargetPhaseOutWithNoTargetChosenIsNoOp covers the empty slot of an
// "up to one target ... phases out." unroll (Talon Gates of Madara's phase-out
// clause): the per-slot TargetPermanentReference resolves to nothing when no
// target was chosen, so the instruction does nothing.
func TestUpToOneTargetPhaseOutWithNoTargetChosenIsNoOp(t *testing.T) {
	env := newTargetedPhaseOutGame(t, 0)
	g, engine, obj := env.game, env.engine, env.obj
	bystander := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bystander",
		Types: []types.Card{types.Creature},
	}})

	startEvents := len(g.Events)
	resolveInstruction(engine, g, obj, game.PhaseOut{
		Object: game.TargetPermanentReference(0),
	}, nil)

	if bystander.PhasedOut {
		t.Fatal("phase out with no chosen target phased out an unrelated permanent")
	}
	if len(g.Events) != startEvents {
		t.Fatalf("phase out with no chosen target emitted %d events, want none", len(g.Events)-startEvents)
	}
}
