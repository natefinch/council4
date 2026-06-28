package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerCantAttackSpell lowers the temporary combat-restriction effect
// "<targets> can't attack this turn." into one ApplyRule instruction per target
// slot, each placing an unconditional RuleEffectCantAttack restriction on a
// targeted creature for the turn (game.DurationThisTurn, removed during
// cleanup). It mirrors the sibling can't-block lowering, accepting the
// single-target and optional/plural multi-target cardinalities the parser
// recognizes; every other recipient, duration, condition, mode, or reference
// fails closed.
func lowerCantAttackSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return lowerTargetCombatRestriction(
		ctx,
		[]game.RuleEffect{{Kind: game.RuleEffectCantAttack}},
		"unsupported can't-attack effect",
		"the executable source backend supports only exact \"<targets> can't attack this turn.\"",
	)
}

// lowerCantAttackOrBlockSpell lowers the combined temporary combat-restriction
// effect "<targets> can't attack or block this turn." into one ApplyRule per
// target slot, each placing both a RuleEffectCantAttack and a RuleEffectCantBlock
// restriction on a targeted creature for the turn. It mirrors the can't-block
// lowering, accepting the single-target and optional/plural multi-target
// cardinalities; every other recipient, duration, condition, mode, or reference
// fails closed.
func lowerCantAttackOrBlockSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return lowerTargetCombatRestriction(
		ctx,
		[]game.RuleEffect{
			{Kind: game.RuleEffectCantAttack},
			{Kind: game.RuleEffectCantBlock},
		},
		"unsupported can't-attack-or-block effect",
		"the executable source backend supports only exact \"<targets> can't attack or block this turn.\"",
	)
}

// lowerTargetMustAttack lowers the temporary single-target forced-attack effect
// "<target> attacks this turn if able." into one ApplyRule per target slot, each
// placing a RuleEffectMustAttack requirement on a targeted creature for the turn.
// It reuses the per-target combat-rule shape of the can't-attack lowering; the
// group "<group> attack this turn if able." form is lowered separately by
// lowerGroupMustAttack. Every other recipient, duration, condition, mode, or
// reference fails closed.
func lowerTargetMustAttack(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return lowerTargetCombatRestriction(
		ctx,
		[]game.RuleEffect{{Kind: game.RuleEffectMustAttack}},
		"unsupported forced-attack effect",
		"the executable source backend supports only exact \"<target> attacks this turn if able.\"",
	)
}

// lowerTargetCombatRestriction lowers an exact "<targets> <combat predicate>
// this turn." resolving effect into one ApplyRule instruction per target slot,
// each applying ruleEffects to a targeted creature for the turn. It is the shared
// implementation behind the temporary single- and multi-target combat
// requirement and restriction lowerings (can't attack, can't attack or block,
// attacks this turn if able), which differ only in the rule effects placed on the
// targeted creature(s). Every recipient, duration, condition, mode, or reference
// outside the exact target-creature form fails closed with summary and detail.
func lowerTargetCombatRestriction(
	ctx contentCtx,
	ruleEffects []game.RuleEffect,
	summary string,
	detail string,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, summary, detail)
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
		effects := make([]game.RuleEffect, len(ruleEffects))
		copy(effects, ruleEffects)
		sequence = append(sequence, game.Instruction{
			Primitive: game.ApplyRule{
				Object:      opt.Val(game.TargetPermanentReference(i)),
				RuleEffects: effects,
				Duration:    game.DurationThisTurn,
			},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), nil
}
