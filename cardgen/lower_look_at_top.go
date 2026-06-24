package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Internal sequencing key for the conditional look-at-top reveal trigger. It is
// local to a single resolution and never observed across abilities.
const lookAtTopLookedKey = game.LinkedKey("look-at-top-card")

// lowerConditionalLookAtTopTrigger lowers the parser-recognized sequence "look
// at the top card of your library; if it's a card of one of the recorded types,
// you may reveal it and put it into your hand; if you don't, you may put it into
// your graveyard" into its fixed instruction template. The compiler marks the
// body with a text-blind exact-sequence kind and carries the disjunctive card
// types; this lowering switches on that kind and consumes only those typed
// values, so it never inspects Oracle words. It fails closed unless the trigger
// carries no other content and the compiler recorded at least one card type.
func lowerConditionalLookAtTopTrigger(
	ability compiler.CompiledAbility,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported triggered ability effect"
	if ability.Trigger == nil {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend requires a semantic trigger pattern")
	}
	if ability.Optional || ability.Content.Unconsumed() || len(ability.ExactSequenceLookAtTopTypes) == 0 {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the conditional look-at-top reveal sequence requires a non-optional trigger with no other content and at least one recorded card type",
		)
	}
	pattern, ok := lowerTriggerPattern(&ability.Trigger.Pattern)
	if !ok {
		summary, detail := triggerPatternCapabilityDiagnostic(ability.Trigger)
		return game.TriggeredAbility{}, executableDiagnostic(ability, summary, detail)
	}
	triggerType, ok := lowerTriggerKind(ability.Trigger.Pattern.Kind)
	if !ok || triggerType == game.TriggerAt {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger kind")
	}
	intervening, ok := lowerAtInterveningCondition(ability.Trigger)
	if !ok {
		return game.TriggeredAbility{}, executableDiagnostic(ability, "unsupported triggered ability",
			"the executable source backend does not support this semantic trigger condition")
	}
	lookedCard := game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(lookAtTopLookedKey)}
	sequence := []game.Instruction{
		{
			Primitive: game.LookAtLibraryTop{
				Player:        game.ControllerReference(),
				PublishLinked: lookAtTopLookedKey,
			},
		},
		{
			Primitive: game.ConditionalDestinationPlace{
				Card:     lookedCard,
				FromZone: zone.Library,
				CardCondition: opt.Val(game.CardSelection{
					Card: lookedCard,
					Selection: game.Selection{
						RequiredTypesAny: ability.ExactSequenceLookAtTopTypes,
					},
				}),
				Then:         zone.Hand,
				ThenReveal:   true,
				Else:         zone.Graveyard,
				ElseOptional: true,
			},
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
