package rules

// Source review: Engine gap #3604. CR 2026-09-25 (MagicCompRules 20260925.txt,
// "effective as of September 25, 2026", verified local copy) 601.2a (a spell's
// continuous characteristic-modifying effects begin as it is put on the stack,
// before targets are announced in 601.2c), 702.16a (protection from a card type
// applies to sources not on the battlefield of that type), 702.16b (a permanent
// with protection can't be targeted by spells with the stated quality),
// 702.140a (a mutating creature spell targets a non-Human creature with the
// same owner), 702.140b (an illegal target as it begins resolving makes it
// resolve as a creature spell), 115.7a/d (changed or new targets must be
// legal) and 707.10c (a copy's new targets must be legal). Pinned Oracle: Dirge
// Bat 59677c76-8148-4b6f-95b0-0e3ccf137a3f "Mutate {4}{B}{B} ... Flash Flying
// ..."; Biotransference 75a787a3-5838-4121-bbaf-745d22c4b0bc "Creatures you
// control are artifacts in addition to their other types. The same is true for
// creature spells you control and creature cards you own that aren't on the
// battlefield. Whenever you cast an artifact spell, you lose 1 life and create
// a 2/2 black Necron Warrior artifact creature token."; Tel-Jilad Chosen
// eefea84f-d657-491d-b60d-63e6a61e9eb2 "Protection from artifacts"; Willbender
// 6e6a3551-250c-4563-8c30-6d8f49e6429f "When this creature is turned face up,
// change the target of target spell or ability with a single target."

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// drainBiotransferenceCastTrigger puts Biotransference's "whenever you cast an
// artifact spell" trigger on the stack and resolves it, returning the Necron
// Warrior token it creates.
func drainBiotransferenceCastTrigger(t *testing.T, g *game.Game, engine *Engine) *game.Permanent {
	t.Helper()
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("Biotransference did not trigger for the artifact Dirge Bat spell")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			return permanent
		}
	}
	t.Fatal("Biotransference did not create a Necron Warrior token")
	return nil
}

func removeBiotransference(t *testing.T, g *game.Game, biotransference *game.Permanent) {
	t.Helper()
	if !movePermanentToZone(g, biotransference, zone.Graveyard) {
		t.Fatal("failed to remove Biotransference")
	}
}

// TestMutateCastTargetsSeeArtifactSpellTypeForProtection proves cast legality
// reads the mutating spell's effective types: Biotransference makes Dirge Bat
// an artifact spell, so Tel-Jilad Chosen (protection from artifacts) isn't a
// legal mutate target from hand or the command zone, while it is without
// Biotransference (CR 601.2a, 601.2c, 702.16b, 702.140a).
func TestMutateCastTargetsSeeArtifactSpellTypeForProtection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	plain := addCreaturePermanent(g, game.Player1)
	chosen := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Tel-Jilad Chosen"))
	batID := addCardToHand(g, game.Player1, productionChoiceCard(t, "Dirge Bat"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 6)
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	g.Turn.PriorityPlayer = game.Player1
	fromHand := func(target *game.Permanent) action.Action { return action.CastMutateSpell(batID, target.ObjectID) }

	if !actionsContain(engine.legalActions(g, game.Player1), fromHand(chosen)) {
		t.Fatal("non-artifact Dirge Bat could not mutate onto Tel-Jilad Chosen")
	}

	biotransference := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	actions := engine.legalActions(g, game.Player1)
	if actionsContain(actions, fromHand(chosen)) {
		t.Fatal("artifact Dirge Bat spell was offered Tel-Jilad Chosen as a mutate target")
	}
	if !actionsContain(actions, fromHand(plain)) {
		t.Fatal("artifact Dirge Bat spell lost its unprotected mutate target")
	}
	if engine.canCastMutateSpell(g, game.Player1, batID, zone.Hand, chosen.ObjectID) {
		t.Fatal("canCastMutateSpell accepted Tel-Jilad Chosen for an artifact spell")
	}
	if engine.applyAction(g, game.Player1, fromHand(chosen)) {
		t.Fatal("artifact Dirge Bat spell was cast targeting Tel-Jilad Chosen")
	}
	if !g.Players[game.Player1].Hand.Contains(batID) || !g.Stack.IsEmpty() {
		t.Fatal("rejected mutate cast moved Dirge Bat")
	}

	player := g.Players[game.Player1]
	player.Hand.Remove(batID)
	player.CommandZone.Add(batID)
	trackCommanderID(g, game.Player1, batID)
	fromCommand := func(target *game.Permanent) action.Action {
		return action.CastMutateSpellFromZone(batID, zone.Command, target.ObjectID)
	}
	actions = engine.legalActions(g, game.Player1)
	if actionsContain(actions, fromCommand(chosen)) || !actionsContain(actions, fromCommand(plain)) {
		t.Fatal("command-zone mutate targets ignored the artifact spell type")
	}

	removeBiotransference(t, g, biotransference)
	if !actionsContain(engine.legalActions(g, game.Player1), fromCommand(chosen)) {
		t.Fatal("Dirge Bat stayed an artifact spell after Biotransference left")
	}
}

// TestWillbenderMutateRetargetSeesGainedArtifactSpellType proves changing a
// mutating spell's target rechecks protection against its current types: once
// Biotransference makes the Dirge Bat spell an artifact, Tel-Jilad Chosen isn't
// offered, and with no other legal target the target is unchanged (CR 115.7a,
// 702.16b).
func TestWillbenderMutateRetargetSeesGainedArtifactSpellType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	alternative := addCreaturePermanent(g, game.Player1)
	chosen := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Tel-Jilad Chosen"))
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(bat.ID), want: []game.Target{game.PermanentTarget(alternative.ObjectID)}}
	turnUpWillbender(t, g, engine, agent)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertOfferedSet(t, agent.requests, alternative)
	assertMutateTarget(t, bat, alternative)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertMergedWith(t, g, batCardID, alternative, original)
	if len(chosen.MergedCards) != 0 {
		t.Fatalf("Tel-Jilad Chosen = %+v, want unmerged", chosen)
	}
}

// TestWillbenderLeavesMutateTargetWhenOnlyProtectedAlternative proves an
// artifact mutating spell whose only other owned creature has protection from
// artifacts keeps its target (CR 115.7a, 702.16b).
func TestWillbenderLeavesMutateTargetWhenOnlyProtectedAlternative(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	chosen := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Tel-Jilad Chosen"))
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(bat.ID)}
	turnUpWillbender(t, g, engine, agent)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(agent.requests) != 0 {
		t.Fatalf("Willbender offered a protected new target: %+v", agent.requests)
	}
	assertMutateTarget(t, bat, original)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertMergedWith(t, g, batCardID, original, chosen)
}

// TestWillbenderMutateRetargetSeesLostArtifactSpellType proves the recheck
// also follows a lost type: Dirge Bat is cast as an artifact spell under
// Biotransference, and after Biotransference leaves, Tel-Jilad Chosen is a
// legal new target the spell then merges with (CR 115.7a, 702.140c).
func TestWillbenderMutateRetargetSeesLostArtifactSpellType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	chosen := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Tel-Jilad Chosen"))
	biotransference := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	bat := castDirgeBatForMutate(t, g, engine, original)
	batCardID := bat.SourceID
	necron := drainBiotransferenceCastTrigger(t, g, engine)
	removeBiotransference(t, g, biotransference)
	agent := &retargetRuleAgent{t: t, hostTarget: game.StackObjectTarget(bat.ID), want: []game.Target{game.PermanentTarget(chosen.ObjectID)}}
	turnUpWillbender(t, g, engine, agent)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	assertOfferedSet(t, agent.requests, chosen, necron)
	assertMutateTarget(t, bat, chosen)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	assertMergedWith(t, g, batCardID, chosen, original)
}

// TestMutatingSpellResolutionRechecksArtifactProtection proves the target is
// rechecked against the spell's types as it begins resolving: Tel-Jilad Chosen
// was legal when the non-artifact Dirge Bat was cast, but Biotransference then
// makes the spell an artifact, so it enters as an ordinary creature (CR 608.2b,
// 702.140b).
func TestMutatingSpellResolutionRechecksArtifactProtection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chosen := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Tel-Jilad Chosen"))
	bat := castDirgeBatForMutate(t, g, engine, chosen)
	batCardID := bat.SourceID
	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanent := permanentWithCardID(g, batCardID); permanent == nil || permanent == chosen || len(permanent.MergedCards) != 0 {
		t.Fatalf("Dirge Bat permanent = %+v, want an ordinary creature", permanent)
	}
	if len(chosen.MergedCards) != 0 || chosen.CardInstanceID == batCardID {
		t.Fatalf("Tel-Jilad Chosen = %+v, want unmerged", chosen)
	}
	assertNoEvent(t, g.Events, game.EventPermanentMutated, func(game.Event) bool { return true })
}

// TestMutatingSpellCopyTargetsSeeArtifactSpellType proves a copy's new mutate
// target and its resolution recheck use the copy's current types: under
// Biotransference Tel-Jilad Chosen isn't offered, after it leaves the copy may
// choose Tel-Jilad Chosen, and when Biotransference returns before the copy
// resolves the copy enters as a token creature (CR 707.10c, 702.140b).
func TestMutatingSpellCopyTargetsSeeArtifactSpellType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	original := addCreaturePermanent(g, game.Player1)
	chosen := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Tel-Jilad Chosen"))
	biotransference := addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	bat := castDirgeBatForMutate(t, g, engine, original)
	necron := drainBiotransferenceCastTrigger(t, g, engine)
	copyBat := func(agent *retargetRuleAgent) *game.StackObject {
		t.Helper()
		addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true},
			[]game.Target{game.StackObjectTarget(bat.ID)})
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
		copied, ok := g.Stack.Peek()
		if !ok || !copied.Copy || !copied.Mutate || copied.ID == bat.ID {
			t.Fatalf("top of stack = %+v, want mutating spell copy", copied)
		}
		return copied
	}

	keep := &retargetRuleAgent{t: t, want: []game.Target{game.PermanentTarget(original.ObjectID)}}
	first := copyBat(keep)
	assertOfferedSet(t, keep.requests, necron, original)
	assertMutateTarget(t, first, original)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(original.MergedCards) != 1 {
		t.Fatalf("first copy target = %+v, want merged token Dirge Bat", original)
	}

	removeBiotransference(t, g, biotransference)
	change := &retargetRuleAgent{t: t, want: []game.Target{game.PermanentTarget(chosen.ObjectID)}}
	second := copyBat(change)
	assertOfferedSet(t, change.requests, chosen, necron, original)
	assertMutateTarget(t, second, chosen)
	assertMutateTarget(t, bat, original)

	addCombatPermanent(g, game.Player1, productionChoiceCard(t, "Biotransference"))
	tokensBefore := countTokens(g)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(chosen.MergedCards) != 0 || chosen.Token {
		t.Fatalf("Tel-Jilad Chosen = %+v, want unmerged", chosen)
	}
	if countTokens(g) != tokensBefore+1 {
		t.Fatalf("token count = %d, want the copy to enter as a token creature", countTokens(g))
	}
}
