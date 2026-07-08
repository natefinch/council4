package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerTransformSelfSpell lowers "Transform this creature." (and the equivalent
// self-name form) into a game.Transform primitive that transforms the source
// permanent (CR 701.28), the shape carried by transforming double-faced cards
// whose front face activates or triggers its own transformation (Ulvenwald
// Captive, werewolves). The single reference must resolve to the source
// permanent; the targeted "Transform target creature." form and any other shape
// fail closed.
func lowerTransformSelfSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported transform effect",
			"the executable source backend supports only \"transform this <permanent>\" transforming the source",
		)
	}
	if effect.Negated ||
		ctx.optional ||
		effect.Optional ||
		!effect.Exact ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.References) != 1 {
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0],
		referenceLoweringContext{AllowSource: true})
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.Transform{Object: object},
		}},
	}.Ability(), nil
}
