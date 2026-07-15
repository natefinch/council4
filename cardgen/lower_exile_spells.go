package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerAnyNumberTargetExileSpells lowers the unbounded "Exile any number of
// target spells." body (Mindbreak Trap) into a single ExileTargetSpells over
// every chosen target spell. The unbounded "any number of" count (Min 0, Max 99)
// cannot unroll one instruction per sentinel slot, so it exiles the whole
// stack-object target spec through the all-target-stack-objects reference,
// mirroring the group-blink exile and any-number phase-out.
//
// Every bounded or single-target cardinality, and every non-controller, negated,
// optional, conditional, modal, keyword-carrying, reference-carrying, or
// otherwise decorated exile shape, fails closed so the single-target counter and
// permanent/zone exile paths keep ownership of them. The stack-spell selector is
// shared with the single-target counter path, so only spell targets — never
// permanents or abilities — reach the group exile.
func lowerAnyNumberTargetExileSpells(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 || len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if !targetCardinalityIsUnbounded(target) {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		effect.Optional ||
		!effect.Exact ||
		effect.Context != parser.EffectContextController ||
		effect.Duration != compiler.DurationNone ||
		effect.CardSource != parser.EffectCardSourceNone ||
		effect.FaceDown ||
		effect.CounterKindKnown ||
		effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := stackSpellTargetSpecUnbounded(target)
	if !ok || targetSpec.MaxTargets < 1 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ExileTargetSpells{Object: game.AllTargetStackObjectsReference(0)},
		}},
	}.Ability(), true
}
