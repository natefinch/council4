package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerCantBeSacrificedSpell lowers the reflexive shield "it can't be sacrificed
// this turn." (Slicer, Hired Muscle) into a single ApplyRule placing a
// RuleEffectCantBeSacrificed restriction on the ability's own source permanent
// for the turn (game.DurationThisTurn, removed during cleanup). The back-
// reference "it" names the source, so the rule effect is scoped with
// AffectedSource; the engine binds it to the source object at application time.
// Any other recipient, duration, condition, mode, keyword, target, or reference
// fails closed so the effect stays faithful and bounded.
func lowerCantBeSacrificedSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-be-sacrificed effect",
			"the executable source backend supports only exact \"it can't be sacrificed this turn.\" naming the source",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextReferencedObject ||
		effect.Duration != compiler.DurationThisTurn ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 ||
		!referencesSourceSelfOnly(ctx.content.References) {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyRule{
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCantBeSacrificed,
					AffectedSource: true,
				}},
				Duration: game.DurationThisTurn,
			},
		}},
	}.Ability(), nil
}
