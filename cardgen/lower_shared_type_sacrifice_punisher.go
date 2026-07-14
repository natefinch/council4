package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// sharedTypeSacrificeLinkKey records the permanent the controller sacrifices so
// the each-opponent shared-card-type offer can read its card types through
// last-known information once it has left the battlefield. It is local to a
// single resolution and never observed across abilities.
const sharedTypeSacrificeLinkKey = game.LinkedKey("braids-sacrificed-permanent")

// sharedTypeSacrificeResultKey gates the punisher on the controller's optional
// sacrifice having succeeded ("If you do"). It is local to a single resolution.
const sharedTypeSacrificeResultKey = game.ResultKey("braids-sacrificed")

// lowerSharedTypeSacrificePunisherTrigger lowers the parser-recognized end-step
// body "you may sacrifice an artifact, creature, enchantment, land, or
// planeswalker. If you do, each opponent may sacrifice a permanent of their
// choice that shares a card type with it. For each opponent who doesn't, that
// player loses 2 life and you draw a card." (Braids, Arisen Nightmare) into an
// optional controller sacrifice of one of the five permanent types followed by a
// PunisherEachLoseLife gated on that sacrifice. The controller's sacrifice
// publishes the sacrificed permanent as a linked object, so the punisher's
// alternative sacrifice can require a permanent that shares a card type with it;
// each opponent who takes the life loss instead lets the controller draw a card.
// The compiler marks the body with a text-blind exact-sequence kind carrying no
// data; this lowering switches on that kind and the trigger's semantic step, so
// it never inspects Oracle words. It fails closed unless the trigger is the
// controller's own end step with no other content.
func lowerSharedTypeSacrificePunisherTrigger(
	ability compiler.CompiledAbility,
	pattern *game.TriggerPattern,
	intervening opt.V[game.Condition],
) (game.TriggeredAbility, *shared.Diagnostic) {
	if pattern.Event != game.EventBeginningOfStep ||
		pattern.Step != game.StepEnd ||
		pattern.Controller != game.TriggerControllerYou ||
		intervening.Exists ||
		ability.Content.Unconsumed() ||
		len(ability.Content.Effects) != 0 {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"the shared-type sacrifice punisher sequence requires the controller's own end-step trigger with no other content",
		)
	}
	sequence := []game.Instruction{
		{
			Primitive: game.SacrificePermanents{
				Player: game.ControllerReference(),
				Amount: game.Fixed(1),
				Selection: game.Selection{
					RequiredTypesAny: []types.Card{
						types.Artifact,
						types.Creature,
						types.Enchantment,
						types.Land,
						types.Planeswalker,
					},
				},
				PublishLinked: sharedTypeSacrificeLinkKey,
				// The controller may sacrifice a token (e.g. a Treasure), so bind the
				// sacrificed permanent by ObjectID to keep its card types readable
				// through last-known information for the each-opponent shared-type offer.
				PublishObjectBinding: true,
			},
			Optional:      true,
			PublishResult: sharedTypeSacrificeResultKey,
		},
		{
			Primitive: game.PunisherEachLoseLife{
				PlayerGroup:    game.OpponentsReference(),
				Amount:         game.Fixed(2),
				AllowSacrifice: true,
				SacrificeSelection: game.Selection{
					SharesCardTypeFromLinked: sharedTypeSacrificeLinkKey,
				},
				ControllerDrawEach: true,
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       sharedTypeSacrificeResultKey,
				Succeeded: game.TriTrue,
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
