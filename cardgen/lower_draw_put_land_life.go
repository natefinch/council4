package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	putLandFromHandLinkKey = game.LinkedKey("put-land-from-hand")
	putLandFromHandResult  = game.ResultKey("put-land-from-hand")
)

func lowerDrawPutLandSubtypeLifeTrigger(ability compiler.CompiledAbility) (game.TriggeredAbility, *shared.Diagnostic) {
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern")
	}
	if ability.Optional ||
		ability.Content.Unconsumed() ||
		ability.ExactSequencePutSubtype == "" ||
		ability.ExactSequenceLifeAmount == 0 {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability effect",
			"the draw-put-land-life sequence requires a subtype, positive life amount, and no other content")
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok {
		summary, detail := triggerPatternCapabilityDiagnostic(ability.Trigger)
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger kind")
	}
	put := game.PutFromHandChoice(
		game.ControllerReference(),
		game.Selection{RequiredTypes: []types.Card{types.Land}},
		game.Fixed(1),
		false,
		false,
		false,
	)
	put.Riders.PublishLinked = putLandFromHandLinkKey
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:    triggerType,
			Pattern: pattern,
		},
		Content: game.Mode{Text: ability.Text, Sequence: []game.Instruction{
			{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			{
				Primitive:     put,
				Optional:      true,
				PublishResult: putLandFromHandResult,
			},
			{
				Primitive: game.GainLife{
					Amount: game.Fixed(int(ability.ExactSequenceLifeAmount)),
					Player: game.ControllerReference(),
				},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       putLandFromHandResult,
					Succeeded: game.TriTrue,
				}),
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object: opt.Val(game.LinkedObjectReference(string(putLandFromHandLinkKey))),
					ObjectMatches: opt.Val(game.Selection{
						SubtypesAny: []types.Sub{ability.ExactSequencePutSubtype},
					}),
				})}),
			},
		}}.Ability(),
	}, nil
}
