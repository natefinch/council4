package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Internal sequencing keys for the chosen-type library-top trigger. They are
// local to a single resolution and never observed across abilities.
const (
	heraldHornLookedKey   = game.LinkedKey("chosen-type-top")
	heraldHornRevealedKey = game.ResultKey("chosen-type-revealed")
)

// lowerChosenTypeLibraryTopTrigger lowers the parser-recognized upkeep sequence
// "look at the top card of your library; if it's a creature card of the chosen
// type, you may reveal it and put it into your hand" into its fixed instruction
// template. The compiler marks the body with a text-blind exact-sequence kind;
// this lowering switches on that kind and the trigger's semantic step, so it
// never inspects Oracle words. It fails closed unless the trigger is the
// controller's own upkeep with no other content.
func lowerChosenTypeLibraryTopTrigger(
	ability compiler.CompiledAbility,
	pattern *game.TriggerPattern,
	intervening opt.V[game.Condition],
) (game.TriggeredAbility, *shared.Diagnostic) {
	if pattern.Event != game.EventBeginningOfStep ||
		pattern.Step != game.StepUpkeep ||
		pattern.Controller != game.TriggerControllerYou ||
		ability.Optional ||
		intervening.Exists ||
		ability.Content.Unconsumed() {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			"unsupported phase/step trigger phrase effect",
			"the chosen-type library-top sequence requires a controller-upkeep trigger with no other content",
		)
	}
	lookedCard := game.CardReference{Kind: game.CardReferenceLinked, LinkID: string(heraldHornLookedKey)}
	sequence := []game.Instruction{
		{
			Primitive: game.LookAtLibraryTop{
				Player:        game.ControllerReference(),
				PublishLinked: heraldHornLookedKey,
			},
		},
		{
			Primitive: game.Reveal{Card: lookedCard},
			CardCondition: opt.Val(game.CardCondition{
				Card:              lookedCard,
				Types:             []types.Card{types.Creature},
				ChosenSubtypeFrom: game.EntryTypeChoiceKey,
			}),
			Optional:      true,
			PublishResult: heraldHornRevealedKey,
		},
		{
			Primitive: game.MoveCard{
				Card:        lookedCard,
				FromZone:    zone.Library,
				Destination: zone.Hand,
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       heraldHornRevealedKey,
				Accepted:  game.TriTrue,
				Succeeded: game.TriTrue,
			}),
		},
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:    game.TriggerAt,
			Pattern: *pattern,
		},
		Content: game.Mode{Text: ability.Text, Sequence: sequence}.Ability(),
	}, nil
}
