package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

const maxVariableRemovalTargets = 20

// Keyword, actor, optionality and amount acceptance remain primitive-owned.
func plainTargetEffect(ctx contentCtx) bool {
	return len(ctx.content.Effects) == 1 &&
		ctx.content.Effects[0].Exact &&
		!ctx.content.Effects[0].Negated &&
		len(ctx.content.Targets) == 1 &&
		len(ctx.content.References) == 0 &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Modes) == 0
}

// targetCardinalitySpec is shared by permanent and card target projection.
// A printed upper-bound X needs an explicit event contract; CountEqualsX cannot
// enforce "up to X".
func targetCardinalitySpec(count compiler.TargetCardinality) (game.TargetSpec, bool) {
	if count.MaxX || count.Min < 0 || count.Max < 1 || count.Min > count.Max {
		return game.TargetSpec{}, false
	}
	if count.CountEqualsX {
		if count.Min != 0 || count.Max != 99 {
			return game.TargetSpec{}, false
		}
		return game.TargetSpec{
			MinTargets: 0, MaxTargets: maxVariableRemovalTargets, CountEqualsX: true,
		}, true
	}
	return game.TargetSpec{MinTargets: count.Min, MaxTargets: count.Max}, true
}

func targetCountDiagnostic(target compiler.CompiledTarget, detail string) *shared.Diagnostic {
	span := target.CountSpan
	if span == (shared.Span{}) {
		span = target.Span
	}
	return &shared.Diagnostic{
		Severity: shared.SeverityWarning, Summary: "unsupported target count parameter",
		Detail: detail, Span: span,
	}
}

func targetCountContentBlocker(ctx contentCtx) *shared.Diagnostic {
	hasCount := false
	for _, effect := range ctx.content.Effects {
		for _, target := range effect.Targets {
			count := target.Cardinality
			parameterized := count.CountEqualsX || count.MaxX
			groupAction := targetCardinalityIsUnbounded(target) &&
				(effect.Kind == compiler.EffectDestroy || effect.Kind == compiler.EffectExile ||
					effect.Kind == compiler.EffectTap ||
					effect.Kind == compiler.EffectReturn && effect.FromZone == zone.None)
			hasCount = hasCount || parameterized || groupAction
			if count.MaxX && ctx.triggerEvent != game.EventPermanentBecameMonstrous {
				return targetCountDiagnostic(target, "up-to-X target counts require a runtime maximum-count parameter; exactly X is not equivalent")
			}
			if count.CountEqualsX && (effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
				effect.Amount.DynamicForm != compiler.DynamicAmountFormNone) {
				return targetCountDiagnostic(target, "derived X target counts are not the spell or activation's chosen X")
			}
		}
	}
	if !hasCount {
		return nil
	}
	for i, target := range ctx.content.Targets {
		if target.Cardinality.Min == 1 && target.Cardinality.Max == 1 {
			continue
		}
		if i != len(ctx.content.Targets)-1 {
			return targetCountDiagnostic(target, "a variable or plural target spec must be last; later target addresses require target-spec identity")
		}
	}
	return nil
}

func selectedGroupTargetMode(
	target compiler.CompiledTarget,
	primitiveFactory func(game.ObjectReference) game.Primitive,
	acceptsSelectedGroup bool,
) (game.AbilityContent, bool) {
	unbounded := targetCardinalityIsUnbounded(target)
	if unbounded && !acceptsSelectedGroup {
		return game.AbilityContent{}, false
	}
	group := unbounded || target.Cardinality.CountEqualsX && acceptsSelectedGroup
	spec, ok := permanentTargetSpecAllowingUnbounded(target, group)
	if !ok {
		return game.AbilityContent{}, false
	}
	if group {
		return game.Mode{
			Targets: []game.TargetSpec{spec},
			Sequence: []game.Instruction{{
				Primitive: primitiveFactory(game.AllTargetPermanentsReference(0)),
			}},
		}.Ability(), true
	}
	sequence := make([]game.Instruction, 0, spec.MaxTargets)
	for i := range spec.MaxTargets {
		sequence = append(sequence, game.Instruction{Primitive: primitiveFactory(game.TargetPermanentReference(i))})
	}
	return game.Mode{Targets: []game.TargetSpec{spec}, Sequence: sequence}.Ability(), true
}
