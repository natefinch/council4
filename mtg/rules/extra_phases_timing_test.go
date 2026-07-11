package rules

import (
	"testing"

	aureliacards "github.com/natefinch/council4/mtg/cards/a"
	sphinxcards "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// phaseOrder extracts the phases recorded by a turn, in the order they began,
// so a test can assert extra phases interleave at the correct boundary.
func phaseOrder(log TurnLog) []game.Phase {
	phases := make([]game.Phase, len(log.Phases))
	for i, entry := range log.Phases {
		phases[i] = entry.Phase
	}
	return phases
}

// stockLibrary gives a player enough basic lands to draw through a turn that
// may contain extra beginning phases (each with its own draw step) without
// decking out and ending the game mid-turn.
func stockLibrary(g *game.Game, playerID game.PlayerID, count int) {
	for range count {
		addCardToLibrary(g, playerID, basicLand())
	}
}

// addExtraCombatAtBeginningOfCombatCreature puts a creature onto controller's
// battlefield whose "at the beginning of combat" trigger queues one additional
// combat phase. MaxTriggersPerTurn is 1 so the trigger fires only during the
// turn's first combat phase and does not chain a new extra combat during the
// extra combat phase it creates. It stands in for an "after this phase, there
// is an additional combat phase" effect that resolves during combat (Éomer,
// Marshal of Rohan), letting a test assert the timing through the turn loop
// without depending on a specific attack being declared.
func addExtraCombatAtBeginningOfCombatCreature(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Extra Combat Source",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerAt,
				Pattern: game.TriggerPattern{
					Event:      game.EventBeginningOfStep,
					Controller: game.TriggerControllerYou,
					Step:       game.StepBeginningOfCombat,
				},
			},
			MaxTriggersPerTurn: 1,
			Content: game.Mode{
				Sequence: []game.Instruction{
					{Primitive: game.AddExtraPhases{Combat: true}},
				},
			}.Ability(),
		}},
	}})
}

// TestExtraCombatPhaseQueuedDuringCombatRunsBeforePostcombatMain proves the
// fix for #2837: an "additional phase after this phase" effect that resolves
// during the combat phase now runs immediately after combat, before the
// postcombat main phase — not drained at the end of the turn after postcombat
// main. It runs a full turn through runTurn and asserts, via the recorded phase
// order, that two combat phases occur before the first postcombat main phase.
func TestExtraCombatPhaseQueuedDuringCombatRunsBeforePostcombatMain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addExtraCombatAtBeginningOfCombatCreature(g, game.Player1)
	stockLibrary(g, game.Player1, 6)

	log := engine.runTurn(g, allFirstLegalAgents())

	phases := phaseOrder(log)
	postcombatIdx := -1
	for i, phase := range phases {
		if phase == game.PhasePostcombatMain {
			postcombatIdx = i
			break
		}
	}
	if postcombatIdx == -1 {
		t.Fatalf("no postcombat main phase in %v", phases)
	}

	combatsBeforePostcombat := 0
	for _, phase := range phases[:postcombatIdx] {
		if phase == game.PhaseCombat {
			combatsBeforePostcombat++
		}
	}
	if combatsBeforePostcombat != 2 {
		t.Fatalf("combat phases before postcombat main = %d, want 2 (the extra combat must run right after combat); phases = %v",
			combatsBeforePostcombat, phases)
	}
	for _, phase := range phases[postcombatIdx:] {
		if phase == game.PhaseCombat {
			t.Fatalf("a combat phase ran after postcombat main (mistimed extra phase); phases = %v", phases)
		}
	}
}

// TestSphinxExtraBeginningPhaseRunsAfterPostcombatMain fires Sphinx of the
// Second Sun's real trigger through the turn loop. Sphinx reads "At the
// beginning of each of your postcombat main phases, there is an additional
// beginning phase after this phase." (its Oracle text queues an additional
// beginning phase — not the combat+main of Aggravated Assault). Because the
// trigger genuinely resolves during the postcombat main phase, the extra phase
// must still run right after postcombat main, unchanged by the #2837 fix: the
// recorded phase order shows a beginning phase after the postcombat main phase
// and before the ending phase.
func TestSphinxExtraBeginningPhaseRunsAfterPostcombatMain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, sphinxcards.SphinxOfTheSecondSun())
	stockLibrary(g, game.Player1, 6)

	log := engine.runTurn(g, allFirstLegalAgents())

	phases := phaseOrder(log)
	postcombatIdx := -1
	for i, phase := range phases {
		if phase == game.PhasePostcombatMain {
			postcombatIdx = i
			break
		}
	}
	if postcombatIdx == -1 {
		t.Fatalf("no postcombat main phase in %v", phases)
	}
	// Exactly one beginning phase runs before combat (the normal one); the
	// Sphinx-created extra beginning phase must appear after postcombat main.
	beginningsBeforePostcombat := 0
	for _, phase := range phases[:postcombatIdx] {
		if phase == game.PhaseBeginning {
			beginningsBeforePostcombat++
		}
	}
	if beginningsBeforePostcombat != 1 {
		t.Fatalf("beginning phases before postcombat main = %d, want 1; phases = %v",
			beginningsBeforePostcombat, phases)
	}
	extraBeginningAfterPostcombat := false
	for _, phase := range phases[postcombatIdx+1:] {
		if phase == game.PhaseBeginning {
			extraBeginningAfterPostcombat = true
		}
	}
	if !extraBeginningAfterPostcombat {
		t.Fatalf("Sphinx's extra beginning phase did not run after postcombat main; phases = %v", phases)
	}
	if phases[len(phases)-1] != game.PhaseEnding {
		t.Fatalf("final phase = %v, want ending; phases = %v", phases[len(phases)-1], phases)
	}
}

// TestAureliaFirstAttackTriggerRunsExtraCombatOncePerTurn fires Aurelia, the
// Warleader's real "Whenever Aurelia attacks for the first time each turn, untap
// all creatures you control. After this phase, there is an additional combat
// phase." trigger through a full turn. Because the inline "for the first time
// each turn" qualifier caps the trigger at one firing per turn, Aurelia's first
// attack queues exactly one additional combat phase; her attack during that
// extra combat does not re-fire the trigger, so exactly two combat phases run
// (not an unbounded chain). Asserting two combats proves both that the
// first-attack trigger fired and that the extra combat phase actually ran.
func TestAureliaFirstAttackTriggerRunsExtraCombatOncePerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, aureliacards.AureliaTheWarleader())
	stockLibrary(g, game.Player1, 6)

	log := engine.runTurn(g, allFirstLegalAgents())

	phases := phaseOrder(log)
	combats := 0
	for _, phase := range phases {
		if phase == game.PhaseCombat {
			combats++
		}
	}
	if combats != 2 {
		t.Fatalf("combat phases = %d, want 2 (Aurelia's first attack queues exactly one extra combat, capped once per turn); phases = %v",
			combats, phases)
	}
}
