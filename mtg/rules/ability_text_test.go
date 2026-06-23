package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestRecordActionSourceCapturesAbilityText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, activatedAbilityPermanent(&game.ActivatedAbility{
		Text:            "{T}: You gain 1 life.",
		AdditionalCosts: cost.Tap,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	}))
	source.SummoningSick = false
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	actionLog := &ActionLog{Player: game.Player1}
	recordActionSource(g, game.Player1, actionLog, action.ActivateAbility(source.ObjectID, 0, nil, 0))

	if actionLog.AbilityText != "{T}: You gain 1 life." {
		t.Fatalf("AbilityText = %q, want the ability's printed text", actionLog.AbilityText)
	}
	if actionLog.ManaAbility {
		t.Fatal("non-mana activated ability flagged as a mana ability")
	}
}

func TestRecordActionSourceFallsBackToOracleText(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := activatedAbilityPermanent(&game.ActivatedAbility{
		AdditionalCosts: cost.Tap,
		Content: game.Mode{
			Sequence: []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}},
		}.Ability(),
	})
	def.OracleText = "{T}: You gain 1 life."
	source := addCombatPermanent(g, game.Player1, def)
	source.SummoningSick = false
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	actionLog := &ActionLog{Player: game.Player1}
	recordActionSource(g, game.Player1, actionLog, action.ActivateAbility(source.ObjectID, 0, nil, 0))

	if actionLog.AbilityText != "{T}: You gain 1 life." {
		t.Fatalf("AbilityText = %q, want the source's oracle text when the ability carries none", actionLog.AbilityText)
	}
}
