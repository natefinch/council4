package cardgen

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func lowerFixedPermanentTargetSpell(
	ctx contentCtx,
	verb string,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
	acceptsSelectedGroup bool,
) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerMultiTargetPermanentSpell(ctx, primitiveFactory, acceptsSelectedGroup); ok {
		return content, nil
	}
	if len(ctx.content.Effects) == 1 &&
		!ctx.content.Effects[0].Optional &&
		!ctx.optional &&
		len(ctx.content.Conditions) == 0 &&
		referencesAreRedundantSoleTargetBackReferences(ctx.content.References, ctx.content.Targets) {
		if content, ok := lowerSinglePermanentTargetSpell(ctx, primitiveFactory); ok {
			return content, nil
		}
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported "+strings.ToLower(verb)+" spell",
		"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
	)
}

// Effect-count, optional-envelope, condition and reference policies belong to
// the caller; this frame consumes the common single-permanent action shape.
func lowerSinglePermanentTargetSpell(
	ctx contentCtx,
	primitiveFactory func(game.ObjectReference) game.Primitive,
) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 ||
		!targetCardinalityIsOne(ctx.content.Targets[0]) ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: primitiveFactory(game.TargetPermanentReference(0)),
		}},
	}.Ability(), true
}
