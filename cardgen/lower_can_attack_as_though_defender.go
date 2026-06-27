package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerCanAttackAsThoughDefenderSpell lowers the temporary combat permission
// "<source> can attack this turn as though it didn't have defender." into an
// ApplyRule instruction that places a RuleEffectCanAttackAsThoughDefender
// permission on the source creature for the turn (game.DurationThisTurn, removed
// during cleanup). Only the source subject the parser recognizes is accepted:
// the source itself ("This creature ...", an activated or triggered self grant)
// or a prior-subject sequence clause inheriting the source. Every other
// recipient, duration, condition, mode, keyword, or reference fails closed so
// the permission stays self-scoped and bounded.
func lowerCanAttackAsThoughDefenderSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can-attack-as-though-defender effect",
			"the executable source backend supports only exact \"<source> can attack this turn as though it didn't have defender.\"",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Duration != compiler.DurationThisTurn ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 1 {
		return unsupported()
	}
	if effect.Context != parser.EffectContextSource &&
		effect.Context != parser.EffectContextPriorSubject {
		return unsupported()
	}
	if ctx.content.References[0].Binding != compiler.ReferenceBindingSource {
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyRule{
				Object: opt.Val(object),
				RuleEffects: []game.RuleEffect{
					{Kind: game.RuleEffectCanAttackAsThoughDefender},
				},
				Duration: game.DurationThisTurn,
			},
		}},
	}.Ability(), nil
}
