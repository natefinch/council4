package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerDividedDamageSpell lowers a "deals N damage divided as you choose among
// <cardinality> <targets>" effect: a total split among the chosen targets, at
// least one to each at resolution (CR 601.2d). It emits one multi-target spec
// and a single Divided Damage instruction whose recipient addresses that spec.
// The total is either a fixed amount or the spell's variable X; it fails closed
// for any shape the executable backend cannot represent exactly.
func lowerDividedDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	amount, capTotal, amountOK := dividedDamageAmount(effect.Amount)
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		!effect.Exact ||
		!effect.Divided ||
		(effect.Context != parser.EffectContextSource &&
			effect.Context != parser.EffectContextReferencedObject &&
			effect.Context != parser.EffectContextPriorSubject) ||
		!amountOK ||
		effect.Negated ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported divided damage spell",
			"the executable source backend supports only an exact fixed or X total divided among one supported multi-target spec",
		)
	}
	target, ok := dividedDamageTargetSpec(ctx.content.Targets[0], capTotal)
	if !ok ||
		!exactDamageSourceSyntax(ctx.content.References) ||
		!exactDamageAmountReferences(effect.Amount, ctx.content.References) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported divided damage spell",
			"the executable source backend supports only an exact fixed or X total divided among one supported multi-target spec",
		)
	}
	damage := game.Damage{
		Amount:       amount,
		Recipient:    game.AnyTargetDamageRecipient(0),
		Divided:      true,
		DamageSource: primaryDamageSource(ctx.content.References),
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{
				Primitive: damage,
			},
		},
	}.Ability(), nil
}

// dividedDamageAmount resolves the total a divided-damage effect splits among its
// targets onto a runtime Quantity. It supports an exact fixed total of at least
// one and the spell's bare variable X ("X damage divided as you choose among
// ..."), rejecting every dynamic or modified amount form the divided path cannot
// represent. The returned capTotal is the fixed total used to bound the chosen
// target count (each target receives at least one); it is zero for a variable X,
// whose runtime total is unknown at lowering and so imposes no static target cap.
func dividedDamageAmount(amount compiler.CompiledAmount) (game.Quantity, int, bool) {
	if amount.DynamicKind != compiler.DynamicAmountNone ||
		amount.DynamicForm != compiler.DynamicAmountFormNone ||
		amount.Addend != 0 || amount.Multiplier != 0 {
		return game.Quantity{}, 0, false
	}
	switch {
	case amount.Known:
		if amount.Value < 1 {
			return game.Quantity{}, 0, false
		}
		return game.Fixed(amount.Value), amount.Value, true
	case amount.VariableX:
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), 0, true
	default:
		return game.Quantity{}, 0, false
	}
}

// dividedDamageTargetSpec builds the multi-target spec a divided-damage effect
// chooses among. The minimum is one (a divided spell must have at least one
// target); the maximum is the wording's bound, further capped at the total for a
// fixed total since each chosen target must receive at least one damage. A
// variable-X total passes capTotal == 0, leaving the bound at the wording's
// maximum; the runtime divides the resolved X among the chosen targets and is
// defensive when fewer than the target count are available. It supports only the
// "any target" and plain "creature" selectors the parser marks exact.
func dividedDamageTargetSpec(target compiler.CompiledTarget, capTotal int) (game.TargetSpec, bool) {
	if !target.Exact && target.Cardinality.Max < 1 {
		return game.TargetSpec{}, false
	}
	maxTargets := target.Cardinality.Max
	if maxTargets < 1 {
		if capTotal < 1 {
			return game.TargetSpec{}, false
		}
		maxTargets = capTotal
	} else if capTotal >= 1 && maxTargets > capTotal {
		maxTargets = capTotal
	}
	if maxTargets < 1 {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: maxTargets,
		Constraint: target.Text,
	}
	switch target.Selector.Kind {
	case compiler.SelectorAny:
		spec.Allow = game.TargetAllowPermanent | game.TargetAllowPlayer
	case compiler.SelectorCreature:
		if selectorHasUnsupportedPermanentFilters(target.Selector) ||
			len(target.Selector.SubtypesAny()) != 0 ||
			len(target.Selector.ColorsAny()) != 0 ||
			len(target.Selector.ExcludedTypes()) != 0 ||
			len(target.Selector.ExcludedColors()) != 0 ||
			len(target.Selector.Supertypes()) != 0 ||
			target.Selector.Attacking || target.Selector.Blocking ||
			target.Selector.Tapped || target.Selector.Untapped {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPermanent
		spec.Selection = opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}})
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}
