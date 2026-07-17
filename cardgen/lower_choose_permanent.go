package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerChoosePermanentSpell lowers a typed resolution-time permanent choice.
// The choice publishes both identity and effective name for later instructions.
func lowerChoosePermanentSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported permanent choice",
			"the executable source backend supports an exact controller choice of one battlefield permanent",
		)
	}
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectChoosePermanent ||
		effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone {
		return unsupported()
	}
	selection, ok := SelectionForSelector(effect.Selector)
	if !ok || selection.Controller != game.ControllerYou {
		return unsupported()
	}
	player := game.ControllerReference()
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Choose{
			Choice: game.ResolutionChoice{
				Kind:            game.ResolutionChoicePermanent,
				PlayerReference: &player,
				Selection:       &selection,
				Prompt:          "Choose a permanent",
			},
			PublishChoice: game.ResolutionChosenPermanentChoiceKey,
		},
	}}}.Ability(), nil
}
