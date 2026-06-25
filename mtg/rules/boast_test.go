package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/opt"
)

// TestBoastActivationRequiresAttackAndOncePerTurn verifies the Boast keyword's
// two implied restrictions: the ability can be activated only if its source
// attacked this turn (the synthesized attacked-this-turn activation condition),
// and only once each turn (the OncePerTurn timing restriction).
func TestBoastActivationRequiresAttackAndOncePerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		Timing:              game.OncePerTurn,
		ActivationCondition: opt.Val(game.BoastActivationCondition()),
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	}))
	source.SummoningSick = false
	g.Turn.Phase = game.PhasePostcombatMain
	g.Turn.Step = game.StepNone
	act := action.ActivateAbility(source.ObjectID, 0, nil, 0)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Boast ability was legal before its source attacked this turn")
	}

	emitEvent(g, game.Event{
		Kind:           game.EventAttackerDeclared,
		SourceObjectID: source.ObjectID,
		PermanentID:    source.ObjectID,
		Controller:     game.Player1,
	})
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Boast ability was not legal after its source attacked this turn")
	}

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("Boast activation failed after the source attacked")
	}
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Boast ability was legal a second time in the same turn")
	}
}
