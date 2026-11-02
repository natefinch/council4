package rules

// Source review: Engine gap #3515. CR 2026-09-25 (MagicCompRules 20260925.txt,
// "effective as of September 25, 2026", verified local copy) 702.140a ("Mutate
// [cost]" makes the spell a mutating creature spell that "targets a non-Human
// creature with the same owner as this spell"), 702.140b (an illegal target as
// it begins resolving makes it resolve as an ordinary creature spell), 702.140c
// (with a legal target it merges with the target creature), 115.7a ("change the
// target(s)": only to another legal target; if none, the original is
// unchanged), 115.7d ("choose new targets": targets may stay unchanged; changed
// targets must be legal) and 707.10/707.10c (a copy keeps the original's
// targets and alternative cost, is owned by the player under whose control it
// was put on the stack, and may choose new targets like 115.7d). Pinned Oracle:
// Dirge Bat 4746828f-46ea-439f-88de-5e54cc536545 "Mutate {4}{B}{B} (If you cast
// this spell for its mutate cost, put it over or under target non-Human
// creature you own. ...) Flash Flying Whenever this creature mutates, destroy
// target creature or planeswalker an opponent controls." (Scryfall rulings: an
// empty list); Willbender 0aae277e-e58e-4115-b5fd-0459451e17ec "When this
// creature is turned face up, change the target of target spell or ability with
// a single target." (rulings 2018-03-16: you must change the target if
// possible, only to a legal target; with none, the target isn't changed);
// Aethersnatch "Gain control of target spell. You may choose new targets for
// it."

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// castDirgeBatForMutate casts a real Dirge Bat from Player1's hand for its
// mutate cost targeting target.
func castDirgeBatForMutate(t *testing.T, g *game.Game, engine *Engine, target *game.Permanent) *game.StackObject {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Dirge Bat"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 6)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastMutateSpell(cardID, target.ObjectID)) {
		t.Fatal("Dirge Bat mutate cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Mutate || obj.SourceID != cardID || obj.MutateTargetID != target.ObjectID {
		t.Fatalf("Dirge Bat stack object = %+v, want mutating spell targeting %d", obj, target.ObjectID)
	}
	return obj
}

func addHumanCreature(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Human Bystander",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Human},
	}})
}

func assertMutateTarget(t *testing.T, obj *game.StackObject, want *game.Permanent) {
	t.Helper()
	if obj.MutateTargetID != want.ObjectID || !slices.Equal(obj.Targets, []game.Target{game.PermanentTarget(want.ObjectID)}) {
		t.Fatalf("mutating spell target = %d %v, want %d", obj.MutateTargetID, obj.Targets, want.ObjectID)
	}
}

func assertOfferedSet(t *testing.T, requests []game.ChoiceRequest, want ...*game.Permanent) {
	t.Helper()
	offered := retargetOfferedTargets(requests)
	if len(offered) != len(want) {
		t.Fatalf("offered targets = %v, want %d options", offered, len(want))
	}
	for _, permanent := range want {
		option := []game.Target{game.PermanentTarget(permanent.ObjectID)}
		if !slices.ContainsFunc(offered, func(got []game.Target) bool { return slices.Equal(got, option) }) {
			t.Fatalf("offered targets = %v, missing %v", offered, option)
		}
	}
}

// assertMergedWith checks that the Dirge Bat card merged on top of merged and
// that untouched kept no merged components.
func assertMergedWith(t *testing.T, g *game.Game, batCardID id.ID, merged, untouched *game.Permanent) {
	t.Helper()
	if merged.CardInstanceID != batCardID || len(merged.MergedCards) != 1 {
		t.Fatalf("new target = %+v, want Dirge Bat merged on top", merged)
	}
	if len(untouched.MergedCards) != 0 || untouched.CardInstanceID == batCardID {
		t.Fatalf("original target = %+v, want unmerged", untouched)
	}
	assertEvent(t, g.Events, game.EventPermanentMutated, func(event game.Event) bool {
		return event.PermanentID == merged.ObjectID && event.CardID == batCardID
	})
	assertNoEvent(t, g.Events, game.EventPermanentMutated, func(event game.Event) bool {
		return event.PermanentID == untouched.ObjectID
	})
}

// TestWillbenderChangesMutatingSpellTargetToAnotherOwnedCreature proves a
// change-target effect offers only another non-Human creature with the spell's
// owner, including one that player owns but doesn't control, and the spell
// then merges with the new target (CR 702.140a,c, CR 115.7a).
func TestWillbenderChangesMutatingSpellTargetToAnotherOwnedCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	alternative := addCreaturePermanent(g, game.Player1)
	stolen := addCreaturePermanent(g, game.Player1)
	stolen.Controller = game.Player2
	addHumanCreature(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(bat.ID), want: []game.Target{game.PermanentTarget(alternative.ObjectID)}}
	turnUpWillbender(t, g, engine, agent)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertOfferedSet(t, agent.requests, alternative, stolen)
	assertMutateTarget(t, bat, alternative)
	assertEvent(t, g.Events, game.EventObjectBecameTarget, func(event game.Event) bool {
		return event.PermanentID == alternative.ObjectID && event.StackObjectID == bat.ID
	})
	clone := g.Clone()
	if cloned, ok := stackObjectByID(clone, bat.ID); !ok || cloned.MutateTargetID != alternative.ObjectID {
		t.Fatalf("cloned mutating spell = %+v, want new target %d", cloned, alternative.ObjectID)
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertMergedWith(t, g, batCardID, alternative, original)
}

// TestWillbenderLeavesMutatingSpellTargetWithoutAlternative proves the
// mutating spell keeps its target when no other legal target exists: a Human,
// another player's creature and Willbender itself aren't legal (CR 115.7a).
func TestWillbenderLeavesMutatingSpellTargetWithoutAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	addHumanCreature(g, game.Player1)
	other := addCreaturePermanent(g, game.Player2)
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(bat.ID)}
	turnUpWillbender(t, g, engine, agent)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(agent.requests) != 0 {
		t.Fatalf("Willbender offered a change with no legal alternative: %+v", agent.requests)
	}
	assertMutateTarget(t, bat, original)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertMergedWith(t, g, batCardID, original, other)
}

// TestRetargetedMutatingSpellWithIllegalNewTargetEntersAsCreature proves the
// changed target is the one rechecked as the spell begins resolving: when it
// has left, the spell resolves as a creature rather than merging with the
// original target (CR 702.140b).
func TestRetargetedMutatingSpellWithIllegalNewTargetEntersAsCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	alternative := addCreaturePermanent(g, game.Player1)
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(bat.ID), want: []game.Target{game.PermanentTarget(alternative.ObjectID)}}
	turnUpWillbender(t, g, engine, agent)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertMutateTarget(t, bat, alternative)
	if !movePermanentToZone(g, alternative, zone.Graveyard) {
		t.Fatal("failed to remove the new mutate target")
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if permanent := permanentWithCardID(g, batCardID); permanent == nil || len(permanent.MergedCards) != 0 {
		t.Fatalf("Dirge Bat permanent = %+v, want ordinary creature", permanent)
	}
	if len(original.MergedCards) != 0 {
		t.Fatalf("original target = %+v, want unmerged", original)
	}
}

// TestAethersnatchNewMutateTargetUsesSpellOwner proves "choose new targets"
// after a control change still requires the spell owner's creature: the new
// controller's own creature isn't offered, and the spell merges with the
// owner's chosen creature under its new controller (CR 702.140a, CR 115.7d).
func TestAethersnatchNewMutateTargetUsesSpellOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	alternative := addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	pushRetargetHostCard(g, game.Player2, productionChoiceCard(t, "Aethersnatch"), game.StackObjectTarget(bat.ID))
	agent := &retargetRuleAgent{t: t, want: []game.Target{game.PermanentTarget(alternative.ObjectID)}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if bat.Controller != game.Player2 {
		t.Fatalf("Dirge Bat controller = %v, want Player2", bat.Controller)
	}
	assertOfferedSet(t, agent.requests, alternative, original)
	assertMutateTarget(t, bat, alternative)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertMergedWith(t, g, batCardID, alternative, original)
	if alternative.Owner != game.Player1 {
		t.Fatalf("merged owner = %v, want the spell owner Player1", alternative.Owner)
	}
}

// TestMutatingSpellCopyChoosesNewTargetAndMergesThere proves a copy of a
// mutating spell keeps its mutate target unless new targets are chosen, and a
// chosen new target is where the copy merges while the original merges with
// its own target (CR 707.10, CR 707.10c, CR 702.140c).
func TestMutatingSpellCopyChoosesNewTargetAndMergesThere(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	alternative := addCreaturePermanent(g, game.Player1)
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true},
		[]game.Target{game.StackObjectTarget(bat.ID)})
	agent := &retargetRuleAgent{t: t, want: []game.Target{game.PermanentTarget(alternative.ObjectID)}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertOfferedSet(t, agent.requests, alternative, original)
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || !copied.Mutate || copied.ID == bat.ID {
		t.Fatalf("top of stack = %+v, want mutating spell copy", copied)
	}
	assertMutateTarget(t, copied, alternative)
	assertMutateTarget(t, bat, original)

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if !alternative.Token || alternative.TokenDef == nil || len(alternative.MergedCards) != 1 {
		t.Fatalf("copy target = %+v, want token Dirge Bat merged on top", alternative)
	}
	if len(original.MergedCards) != 0 {
		t.Fatalf("original target merged with the copy: %+v", original)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if original.CardInstanceID != batCardID || len(original.MergedCards) != 1 {
		t.Fatalf("original target = %+v, want Dirge Bat card merged on top", original)
	}
}
