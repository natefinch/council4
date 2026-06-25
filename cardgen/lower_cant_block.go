package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerCantBlockSpell lowers the temporary combat-restriction effect "<targets>
// can't block this turn." into one ApplyRule instruction per target slot, each
// placing an unconditional RuleEffectCantBlock restriction on a targeted
// creature for the turn (game.DurationThisTurn, removed during cleanup). It
// accepts the single-target form and the optional/plural multi-target
// cardinalities ("Up to three target creatures can't block this turn.") the
// parser recognizes; every other recipient, duration, condition, mode, or
// reference fails closed so the broader "can't block this turn" family stays
// faithful and bounded.
func lowerCantBlockSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported can't-block effect",
			"the executable source backend supports only exact \"<targets> can't block this turn.\"",
		)
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextTarget ||
		effect.Duration != compiler.DurationThisTurn ||
		ctx.optional ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Selector.Kind != compiler.SelectorCreature ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.ApplyRule{
				Object: opt.Val(game.TargetPermanentReference(i)),
				RuleEffects: []game.RuleEffect{
					{Kind: game.RuleEffectCantBlock},
				},
				Duration: game.DurationThisTurn,
			},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), nil
}
