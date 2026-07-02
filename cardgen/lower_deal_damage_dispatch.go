package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// This file centralizes the dispatch invariants that lowerDealDamageSpell
// guarantees before it delegates to any specific deal-damage lowerer. Every such
// lowerer (both the diagnostic-returning ones dispatched on a recipient reference
// and the trial ok-returning ones tried in sequence) has exactly one caller — a
// lowerDealDamageSpell branch — and lowerDealDamageSpell itself is reached only
// through the EffectDealDamage arm of lowerImmediateSingleEffectSpell's
// effect-kind switch. That content is always single-effect: every path into
// lowerImmediateSingleEffectSpell narrows to one effect (the len==1 gate in
// lowerSingleEffectSpell's caller, the RepeatBody==1 gate in lower_repeat, the
// len==1 gate for delayed effects, and contextForEffect narrowing each control
// sequence clause to a single effect).
//
// Because of that, an effect count other than one, an effect kind other than
// EffectDealDamage, a divided flag that does not match the branch, or a recipient
// reference that does not match the branch, is an internal dispatch bug rather
// than an unsupported card. The lowerers used to re-test these conditions inside
// their unsupported/recognition guards, which silently reported a would-be
// dispatch bug as a missing card feature; asserting them instead makes such a bug
// loud and keeps those guards limited to genuine capability gaps.

// assertDealDamageDispatch asserts the invariants shared by every deal-damage
// lowerer reached from lowerDealDamageSpell: the content is single-effect, that
// effect is an EffectDealDamage, and its divided flag matches the branch that
// dispatched here (true only for lowerDividedDamageSpell, which lowerDealDamageSpell
// routes first; false for every other lowerer).
func assertDealDamageDispatch(ctx contentCtx, divided bool) {
	if len(ctx.content.Effects) != 1 {
		panic(fmt.Sprintf(
			"deal-damage lowerer reached with %d effects; lowerDealDamageSpell dispatches only single-effect content",
			len(ctx.content.Effects)))
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDealDamage || effect.Divided != divided {
		panic(fmt.Sprintf(
			"deal-damage lowerer reached with kind=%v divided=%v; lowerDealDamageSpell dispatches here only for an EffectDealDamage with divided=%v",
			effect.Kind, effect.Divided, divided))
	}
}

// assertUndividedRecipientDamageDispatch asserts the invariants that hold when
// lowerDealDamageSpell delegates to a recipient-specific damage lowerer
// (lowerControllerDamageSpell, lowerEventPlayerDamageSpell,
// lowerEventRelatedPermanentDamageSpell, lowerAttackedDefenderDamageSpell): the
// shared undivided single-effect deal-damage invariants, plus a recipient
// reference matching the branch that dispatched here.
func assertUndividedRecipientDamageDispatch(
	ctx contentCtx,
	reference parser.DamageRecipientReferenceKind,
) {
	assertDealDamageDispatch(ctx, false)
	if ctx.content.Effects[0].DamageRecipient.Reference != reference {
		panic(fmt.Sprintf(
			"recipient damage lowerer reached with recipient=%v; lowerDealDamageSpell dispatches here only for recipient=%v",
			ctx.content.Effects[0].DamageRecipient.Reference, reference))
	}
}
