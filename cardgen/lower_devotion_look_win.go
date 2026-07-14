package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerDevotionLookWinTrigger lowers the parser-recognized sequence "look at the
// top X cards of your library, where X is your devotion to <color>. Put up to
// one of them on top of your library and the rest on the bottom of your library
// in a random order. If X is greater than or equal to the number of cards in
// your library, you win the game." (Thassa's Oracle) into its fixed instruction
// template. The compiler marks the body with a text-blind exact-sequence kind
// and carries the devotion color; this lowering switches on that kind and
// consumes only that typed color, so it never inspects Oracle words. It fails
// closed unless the trigger carries no other content and the compiler recorded a
// devotion color.
//
// The win check is a separate instruction gated on its own live comparison
// rather than on the look succeeding: an empty library makes the look a no-op
// but the controller still wins when their devotion (X) is at least their
// library size (X >= 0 is always true, and 0 >= 0 holds). Both the look amount
// (X) and the win threshold (X) are the controller's live devotion measured as
// the instructions resolve, so the source leaving the battlefield before
// resolution correctly drops its own devotion contribution.
func lowerDevotionLookWinTrigger(
	ability compiler.CompiledAbility,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported triggered ability effect"
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern")
	}
	if ability.Optional || ability.Content.Unconsumed() || ability.ExactSequenceDevotionColor == "" {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the devotion look-win sequence requires a non-optional trigger with no other content and a recorded devotion color",
		)
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
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger condition")
	}
	devotion := game.DynamicAmount{
		Kind:   game.DynamicAmountDevotion,
		Colors: []color.Color{ability.ExactSequenceDevotionColor},
	}
	sequence := []game.Instruction{
		{
			Primitive: game.Dig{
				Player:      game.ControllerReference(),
				Look:        game.Dynamic(devotion),
				Take:        game.Fixed(1),
				TakeUpTo:    true,
				Destination: zone.Library,
				Remainder:   game.DigRemainderLibraryBottom,
			},
		},
		{
			Primitive: game.PlayerWinsGame{Player: game.ControllerReference()},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{
					Aggregates: []game.AggregateComparison{{
						Aggregate:   game.AggregateControllerLibrarySize,
						Op:          compare.LessOrEqual,
						ValueAmount: opt.Val(devotion),
					}},
				}),
			}),
		},
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:                 triggerType,
			Pattern:              pattern,
			InterveningIf:        interveningIfText(ability.Trigger),
			InterveningCondition: intervening,
		},
		Content: game.Mode{Text: ability.Text, Sequence: sequence}.Ability(),
	}, nil
}
