package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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
	assertDealDamageDispatch(ctx, true)
	amount, capTotal, amountOK := dividedDamageTotal(effect.Amount, ctx)
	if !effect.Exact ||
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

// dividedDamageTotal resolves the total a divided-damage effect splits among its
// targets onto a runtime Quantity. It supports an exact fixed total of at least
// one ("5 damage divided ..."), the spell's bare variable X ("X damage divided
// ..."), and a dynamic scalar total whose value the runtime computes once at
// resolution ("X damage divided ..., where X is the number of lands you control",
// Ureni; "X damage divided ..., where X is the number of age counters on it",
// Magmatic Core). It rejects every modified amount form (an addend or multiplier
// the divided path cannot represent). The returned capTotal is the fixed total
// used to bound the chosen target count (each target receives at least one); it
// is zero for any non-fixed total, whose runtime value is unknown at lowering and
// so imposes no static target cap.
func dividedDamageTotal(amount compiler.CompiledAmount, ctx contentCtx) (game.Quantity, int, bool) {
	if amount.DynamicKind == compiler.DynamicAmountNone &&
		amount.DynamicForm == compiler.DynamicAmountFormNone {
		if amount.Addend != 0 || amount.Multiplier != 0 {
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
	if !dividedDynamicTotalKind(amount.DynamicKind) {
		return game.Quantity{}, 0, false
	}
	object := game.SourcePermanentReference()
	if referent, ok := lowerDamageAmountObject(amount, ctx.content.References); ok {
		object = referent
	}
	dynamic, ok := lowerDynamicAmount(amount, object)
	if !ok {
		return game.Quantity{}, 0, false
	}
	return game.Dynamic(dynamic), 0, true
}

// dividedDynamicTotalKind reports whether a dynamic amount kind resolves to a
// single scalar the divided path can use as the whole-spell total computed once
// at resolution. It admits the group-wide kinds shared by a damage group (count
// selectors, devotion, domain, controller life, opponent count, greatest- and
// total-in-group) plus the per-object source characteristics (power, toughness,
// mana value, and counter count), each of which is a single value read from the
// spell's own source. Per-recipient or otherwise multi-valued forms have no
// single divided total, so they stay rejected and the divided path fails closed.
func dividedDynamicTotalKind(kind compiler.DynamicAmountKind) bool {
	if groupWideDynamicAmountKind(kind) {
		return true
	}
	switch kind {
	case compiler.DynamicAmountSourcePower,
		compiler.DynamicAmountSourceToughness,
		compiler.DynamicAmountSourceManaValue,
		compiler.DynamicAmountSourceCounterCount:
		return true
	default:
		return false
	}
}

// dividedDamageTargetSpec builds the multi-target spec a divided-damage effect
// chooses among. The minimum comes from the wording: "up to N" and "any number
// of" let the controller choose zero targets, while the enumerated "one or
// two"/"one, two, or three" forms require at least one. The maximum is the
// wording's bound, further capped at the total for a fixed total since each chosen
// target must receive at least one damage. A variable-X total passes capTotal == 0,
// leaving the bound at the wording's maximum; the runtime divides the resolved X
// among the chosen targets and is defensive when fewer than the target count are
// available. It supports only the "any target" and plain "creature" selectors the
// parser marks exact.
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
	// "up to N" and "any number of" cardinalities let the controller choose zero
	// targets, in which case no damage is dealt (CR 601.2d/608.2d, which apply the
	// post-War of the Spark rule change to both damage and counters); the
	// enumerated forms require at least one. Derive the minimum from the wording
	// rather than forcing at least one target.
	minTargets := min(target.Cardinality.Min, maxTargets)
	// "Any target" admits both permanents and players, a combination the
	// permanent selection does not model, so build it directly.
	if target.Selector.Kind == compiler.SelectorAny {
		if selectorHasCounterQualifier(target.Selector) ||
			selectorHasAttachmentQualifier(target.Selector) {
			return game.TargetSpec{}, false
		}
		return game.TargetSpec{
			MinTargets: minTargets,
			MaxTargets: maxTargets,
			Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
			Constraint: target.Text,
		}, true
	}
	selection, ok := dividedDamagePermanentSelection(target.Selector)
	if !ok {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: minTargets,
		MaxTargets: maxTargets,
		Allow:      game.TargetAllowPermanent,
		Constraint: target.Text,
	}
	if !selection.Empty() {
		spec.Selection = opt.Val(selection)
	}
	return spec, true
}

// dividedDamagePermanentSelection builds the runtime permanent filter a
// divided-damage spell chooses among. It supports the plain "creature" noun, the
// "creature and/or planeswalker" card-type union, a controller filter ("your
// opponents control"), color adjectives ("white and/or blue"), an
// attacking/blocking combat state, counter qualifier, and a single
// "with"/"without" keyword qualifier — the selectors the divided round-trip
// reconstructs. It fails closed for every subtype, supertype, tapped, or numeric
// qualifier it does not model.
func dividedDamagePermanentSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Kind != compiler.SelectorCreature {
		return game.Selection{}, false
	}
	if selectorHasUnsupportedPermanentFilters(selector) ||
		selector.Tapped || selector.Untapped ||
		selector.Another || selector.Other ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		selector.PowerLessThanSource || selector.PowerGreaterThanSource ||
		selector.TokenOnly || selector.NonToken ||
		len(selector.SubtypesAny()) != 0 ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.ExcludedColors()) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.ExcludedSubtypes()) != 0 {
		return game.Selection{}, false
	}
	required := []types.Card{types.Creature}
	if union := selector.RequiredTypesAny(); len(union) != 0 {
		if union[0] != types.Creature {
			return game.Selection{}, false
		}
		for _, cardType := range union {
			if cardType != types.Creature && cardType != types.Planeswalker {
				return game.Selection{}, false
			}
		}
		required = append([]types.Card(nil), union...)
	}
	selection := game.Selection{RequiredTypesAny: required}
	applyCounterTargetSelection(&selection, selector)
	applyAttachmentTargetSelection(&selection, selector)
	switch {
	case selector.Attacking && selector.Blocking:
		selection.CombatState = game.CombatStateAttackingOrBlocking
	case selector.Attacking:
		selection.CombatState = game.CombatStateAttacking
	case selector.Blocking:
		selection.CombatState = game.CombatStateBlocking
	default:
	}
	if selector.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	if selector.ExcludedKeyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.ExcludedKeyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.ExcludedKeyword = keyword
	}
	if colors := selector.ColorsAny(); len(colors) != 0 {
		selection.ColorsAny = append([]color.Color(nil), colors...)
	}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		selection.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		selection.Controller = game.ControllerNotYou
	default:
		// ControllerKind is a closed enum whose values are all handled above;
		// an unhandled value is an internal bug, not an unsupported card.
		panic(fmt.Sprintf("dividedDamagePermanentSelection: unhandled ControllerKind %v", selector.Controller))
	}
	return selection, true
}
