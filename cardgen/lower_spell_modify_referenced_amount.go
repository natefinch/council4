package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerReferencedAmountModifyPTTargetSpell lowers an exact until-end-of-turn
// power/toughness pump of a single target creature whose magnitude is a dynamic
// amount counted over a referenced object: the source permanent ("Target
// creature gets +X/+X until end of turn, where X is the number of verse counters
// on this enchantment.", War Dance; "… for each +1/+1 counter on this creature.",
// Canopy Crawler) or the target itself ("Target creature gets +0/+X until end of
// turn, where X is its mana value.", Great Defender). It differs from the
// fixed-amount target pump only in that the dynamic amount names a referent the
// effect carries as a CompiledReference, which the exact-amount path
// (validModifyPTAmount) rejects because it forbids any reference. The pumped
// object is always the target slot; the dynamic count is anchored to the
// referent object. It returns ok=false for any shape outside this bounded set so
// the caller falls through to the fail-closed diagnostic.
func lowerReferencedAmountModifyPTTargetSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := &ctx.content.Effects[0]
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) == 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Context != parser.EffectContextTarget ||
		!referencedAmountModifyPTKind(effect.Amount.DynamicKind) {
		return game.AbilityContent{}, false
	}
	referent, ok := referencedAmountReferent(ctx.content.References, effect.Amount.ReferenceSpan)
	if !ok {
		return game.AbilityContent{}, false
	}
	if !referencedAmountReferentBindingValid(referent) {
		return game.AbilityContent{}, false
	}
	countObject, ok := lowerObjectReference(referent, referenceLoweringContext{
		AllowSource: true,
		AllowTarget: true,
	})
	if !ok {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	powerDelta, toughnessDelta, ok := referencedAmountModifyPTQuantities(effect, countObject)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         game.TargetPermanentReference(0),
				PowerDelta:     powerDelta,
				ToughnessDelta: toughnessDelta,
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}

// referencedAmountReferentBindingValid reports whether the referent the dynamic
// amount counts over is one whose runtime identity the target pump can resolve
// soundly. A source-bound referent is the source permanent itself. A
// target-bound referent is accepted only when it is a pronoun self reference
// ("its mana value", "counters on it") that names the very creature being
// pumped; a target-bound demonstrative ("that spell's mana value") is a
// different object the antecedent binder routed to the target slot, so it fails
// closed rather than reading the wrong object's characteristic.
func referencedAmountReferentBindingValid(referent compiler.CompiledReference) bool {
	switch referent.Binding {
	case compiler.ReferenceBindingSource:
		return true
	case compiler.ReferenceBindingTarget:
		return referent.Kind == compiler.ReferencePronoun
	default:
		return false
	}
}

// referencedAmountModifyPTQuantities computes the power and toughness deltas of a
// single-target pump whose dynamic magnitude is counted over countObject: a
// counter count ("the number of <kind> counters on this <permanent>", lowered
// through lowerDynamicAmount) or the referent's mana value or toughness ("its
// mana value", lowered through objectCharacteristicAmount). Each delta side is
// the dynamic amount when written as the variable "X" ("+X", "for each ...") and
// its fixed signed value otherwise ("+0" in "+0/+X"). It returns ok=false for
// the source-power form and any amount shape the dynamic machinery cannot model.
func referencedAmountModifyPTQuantities(
	effect *compiler.CompiledEffect,
	countObject game.ObjectReference,
) (power, toughness game.Quantity, ok bool) {
	if !dynamicModifyPTFormValid(effect) {
		return game.Quantity{}, game.Quantity{}, false
	}
	dynamic, ok := referencedAmountDynamic(effect.Amount, countObject)
	if !ok {
		return game.Quantity{}, game.Quantity{}, false
	}
	switch effect.Amount.DynamicForm {
	case compiler.DynamicAmountWhereX:
		return whereXSignedQuantity(&dynamic, effect.PowerDelta),
			whereXSignedQuantity(&dynamic, effect.ToughnessDelta), true
	case compiler.DynamicAmountForEach:
		return dynamicSignedQuantity(&dynamic, effect.PowerDelta),
			dynamicSignedQuantity(&dynamic, effect.ToughnessDelta), true
	default:
		return game.Quantity{}, game.Quantity{}, false
	}
}

// referencedAmountDynamic lowers a counter-count amount through lowerDynamicAmount
// and a mana-value or toughness amount through objectCharacteristicAmount, in
// both cases anchoring the count to countObject. Splitting on the amount kind
// keeps the object-characteristic kinds, which lowerDynamicAmount does not anchor
// to an arbitrary object, restricted to the referent-binding target pump.
func referencedAmountDynamic(
	amount compiler.CompiledAmount,
	countObject game.ObjectReference,
) (game.DynamicAmount, bool) {
	if amount.DynamicKind == compiler.DynamicAmountSourceCounterCount {
		return lowerDynamicAmount(amount, countObject)
	}
	return objectCharacteristicAmount(amount.DynamicKind, countObject)
}

// referencedAmountModifyPTKind reports whether a dynamic power/toughness amount
// kind is one counted over a referenced object that the referenced-amount target
// pump supports: a counter count, the referent's mana value, or its toughness.
// The source-power form ("its power") is excluded because the executable backend
// does not bind that referent, matching referencedModifyPTQuantities.
func referencedAmountModifyPTKind(kind compiler.DynamicAmountKind) bool {
	switch kind {
	case compiler.DynamicAmountSourceCounterCount,
		compiler.DynamicAmountSourceManaValue,
		compiler.DynamicAmountSourceToughness:
		return true
	default:
		return false
	}
}

// referencedAmountReferent splits the dynamic amount's referent reference (the
// one whose span names the counted object) from the remaining references. It
// returns ok=false unless exactly one reference matches the amount's referent
// span and every other reference binds the source permanent, so a benign self
// reference carried by an activation cost ("Sacrifice this enchantment:")
// composes while any reference that would change the pumped object's resolution
// fails closed.
func referencedAmountReferent(
	references []compiler.CompiledReference,
	referenceSpan shared.Span,
) (compiler.CompiledReference, bool) {
	var referent compiler.CompiledReference
	found := false
	for _, reference := range references {
		if reference.Span == referenceSpan {
			if found {
				return compiler.CompiledReference{}, false
			}
			referent = reference
			found = true
			continue
		}
		if reference.Binding != compiler.ReferenceBindingSource {
			return compiler.CompiledReference{}, false
		}
	}
	return referent, found
}
