package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerPayHandSizeOrCantAttackTrigger lowers the parser-recognized punisher body
// "that opponent may pay {X}, where X is the number of cards in their hand. If
// they don't, they can't attack you this combat." (Champions of Minas Tirith)
// into a single PlayerMayPayGenericOrRule instruction. The compiler marks the
// body with a text-blind exact-sequence kind; this lowering switches on that
// kind and the trigger's semantic step, so it never inspects Oracle words. It
// fails closed unless the trigger is each opponent's beginning of combat gated
// by an intervening-if condition, with no other content.
func lowerPayHandSizeOrCantAttackTrigger(
	ability compiler.CompiledAbility,
	pattern *game.TriggerPattern,
	intervening opt.V[game.Condition],
) (game.TriggeredAbility, *shared.Diagnostic) {
	if pattern.Event != game.EventBeginningOfStep ||
		pattern.Step != game.StepBeginningOfCombat ||
		pattern.Controller != game.TriggerControllerOpponent ||
		ability.Optional ||
		!intervening.Exists ||
		ability.Content.Unconsumed() {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"the pay-or-can't-attack sequence requires an opponent beginning-of-combat trigger gated by an intervening-if condition with no other content",
		)
	}
	handSize := game.EventPlayerReference()
	handSelection := game.Selection{}
	instruction := game.Instruction{
		Primitive: game.PlayerMayPayGenericOrRule{
			Player: game.EventPlayerReference(),
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:      game.DynamicAmountCountCardsInZone,
				Player:    &handSize,
				CardZone:  zone.Hand,
				Selection: &handSelection,
			}),
			RuleEffects: []game.RuleEffect{{
				Kind:                      game.RuleEffectCantAttack,
				AffectedPlayerRef:         game.EventPlayerReference(),
				DefendingPlayer:           game.PlayerYou,
				DefendingPlayerDirectOnly: true,
			}},
			Duration: game.DurationUntilEndOfCombat,
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
		Content: game.Mode{Text: ability.Text, Sequence: []game.Instruction{instruction}}.Ability(),
	}, nil
}
