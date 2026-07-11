package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// extraDrawPayLifeOrTopDrewKey gates the pay-life-or-return instruction on the
// controller having drawn the additional cards. It is local to a single
// resolution and never observed across abilities.
const extraDrawPayLifeOrTopDrewKey = game.ResultKey("sylvan-extra-draw-drew")

// lowerExtraDrawPayLifeOrTopTrigger lowers the parser-recognized draw-step body
// "you may draw N additional cards. If you do, choose M cards in your hand drawn
// this turn. For each of those cards, pay L life or put the card on top of your
// library." (Sylvan Library) into an optional Draw followed by a
// ChooseDrawnPayLifeOrTop gated on the draw. The compiler marks the body with a
// text-blind exact-sequence kind carrying the counts N, M, and L; this lowering
// switches on that kind and the trigger's semantic step, so it never inspects
// Oracle words. It fails closed unless the trigger is the controller's own draw
// step with no other content.
func lowerExtraDrawPayLifeOrTopTrigger(
	ability compiler.CompiledAbility,
	pattern *game.TriggerPattern,
	intervening opt.V[game.Condition],
) (game.TriggeredAbility, *shared.Diagnostic) {
	if pattern.Event != game.EventBeginningOfStep ||
		pattern.Step != game.StepDraw ||
		pattern.Controller != game.TriggerControllerYou ||
		intervening.Exists ||
		ability.Content.Unconsumed() ||
		ability.ExactSequenceDrawCount == 0 ||
		ability.ExactSequenceChooseCount == 0 {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"the extra-draw pay-life-or-top sequence requires the controller's own draw-step trigger with no other content",
		)
	}
	sequence := []game.Instruction{
		{
			Primitive: game.Draw{
				Player: game.ControllerReference(),
				Amount: game.Fixed(int(ability.ExactSequenceDrawCount)),
			},
			Optional:      true,
			PublishResult: extraDrawPayLifeOrTopDrewKey,
		},
		{
			Primitive: game.ChooseDrawnPayLifeOrTop{
				Player:      game.ControllerReference(),
				ChooseCount: int(ability.ExactSequenceChooseCount),
				LifeCost:    int(ability.ExactSequencePayLife),
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:      extraDrawPayLifeOrTopDrewKey,
				Accepted: game.TriTrue,
			}),
		},
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 game.TriggerAt,
			Pattern:              *pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Content: game.Mode{Text: ability.Text, Sequence: sequence}.Ability(),
	}, nil
}
