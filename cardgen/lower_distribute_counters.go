package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerDistributeCountersSpell lowers a "Distribute N <kind> counters among
// <cardinality> target creatures" effect: a fixed (or X) total of counters split
// among the chosen targets, at least one each, the counter analog of divided
// damage. It emits one multi-target spec and a single Distribute AddCounter
// instruction addressing every permanent chosen for that spec. It fails closed
// for any non-controller, negated, referenced, conditional, or modal shape, an
// unplaceable counter kind, and an amount the distribution cannot represent.
func lowerDistributeCountersSpell(ctx contentCtx) (game.AbilityContent, bool) {
	// lowerPutEffectSpell — this function's only caller — is reached solely
	// through the EffectPut arm of lowerImmediateSingleEffectSpell, whose content
	// is always single-effect, so an effect count other than one is a dispatch
	// bug rather than an unsupported card.
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf(
			"lowerDistributeCountersSpell: reached with %d effects; lowerPutEffectSpell dispatches only single-effect content",
			len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if !effect.DistributeCounters {
		return game.AbilityContent{}, false
	}
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(effect.CounterKind) ||
		effect.CounterKind.PlayerOnly() {
		return game.AbilityContent{}, false
	}
	amount, capTotal, ok := distributeCountersAmount(effect.Amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	spec, ok := distributeCountersTargetSpec(ctx.content.Targets[0], capTotal)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{{
			Primitive: game.AddCounter{
				Amount:      amount,
				Object:      game.AllTargetPermanentsReference(0),
				CounterKind: effect.CounterKind,
				Distribute:  true,
			},
		}},
	}.Ability(), true
}

// distributeCountersAmount resolves the total a distribute-counters effect splits
// among its targets onto a runtime Quantity. It supports an exact fixed total of
// at least one and the spell's bare variable X ("Distribute X +1/+1 counters
// ..."), rejecting every dynamic or modified amount form. The returned capTotal
// bounds the chosen target count for a fixed total (each target receives at least
// one); it is zero for a variable X, whose runtime total is unknown at lowering.
func distributeCountersAmount(amount compiler.CompiledAmount) (game.Quantity, int, bool) {
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

// distributeCountersTargetSpec builds the multi-target spec a distribute-counters
// effect chooses among. The minimum is one (the controller must choose at least
// one target to receive the counters); the maximum is the wording's bound,
// further capped at the fixed total since each chosen target must receive at
// least one counter. A variable-X total passes capTotal == 0, leaving the bound
// at the wording's maximum; the runtime distributes the resolved X among the
// chosen targets. The effect-level byte-exact round-trip already restricts the
// clause to "target creatures[ you control]", so the spec is built directly from
// the cardinality rather than the target's own exactness flag, which the wider
// "one, two, or three" range does not set. It mirrors dividedDamageTargetSpec.
func distributeCountersTargetSpec(target compiler.CompiledTarget, capTotal int) (game.TargetSpec, bool) {
	maxTargets := target.Cardinality.Max
	if maxTargets < 1 {
		return game.TargetSpec{}, false
	}
	if capTotal >= 1 && maxTargets > capTotal {
		maxTargets = capTotal
	}
	if maxTargets < 1 {
		return game.TargetSpec{}, false
	}
	selection, ok := distributeCountersPermanentSelection(target.Selector)
	if !ok {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: maxTargets,
		Allow:      game.TargetAllowPermanent,
		Constraint: lowerFirst(target.Text),
	}
	if !selection.Empty() {
		spec.Selection = opt.Val(selection)
	}
	return spec, true
}

// distributeCountersPermanentSelection builds the runtime permanent filter a
// distribute-counters spell chooses among. It supports the plain "creature" noun
// optionally restricted to the controller or a counter qualifier — the selectors
// the distribute round-trip reconstructs byte-exactly. It fails closed for every
// other qualifier so an unsupported wording cannot reach a spec.
func distributeCountersPermanentSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Kind != compiler.SelectorCreature {
		return game.Selection{}, false
	}
	if selectorHasUnsupportedPermanentFilters(selector) ||
		selector.Tapped || selector.Untapped ||
		selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		selector.PowerLessThanSource || selector.PowerGreaterThanSource ||
		selector.TokenOnly || selector.NonToken ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.ExcludedKeyword != parser.KeywordUnknown ||
		len(selector.SubtypesAny()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.ExcludedColors()) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.ExcludedSubtypes()) != 0 {
		return game.Selection{}, false
	}
	if union := selector.RequiredTypesAny(); len(union) > 1 ||
		(len(union) == 1 && union[0] != types.Creature) {
		return game.Selection{}, false
	}
	selection := game.Selection{RequiredTypesAny: []types.Card{types.Creature}}
	applyCounterTargetSelection(&selection, selector)
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	default:
		return game.Selection{}, false
	}
	return selection, true
}
