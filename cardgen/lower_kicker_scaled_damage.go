package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// maxKickerScaledTargets bounds the target slots a Multikicker "another target
// for each time this spell was kicked" spell may choose. It equals one base
// target plus the runtime's maximum enumerated multikick count
// (maxLegalMultikickCount = 20 in the rules engine), so a spell kicked the most
// times the caster can pay for still has a slot for each of its 1 + kicks
// targets. Announcement over-generates every target count in [1, this] and the
// cast-time CountEqualsKickerPlusOne check binds the legal count for the chosen
// kicker count, mirroring the CountEqualsX any-target precedent.
const maxKickerScaledTargets = 1 + 20

// lowerKickerScaledEachTargetDamageSpell lowers a Multikicker damage spell that
// chooses one target plus another target for each time it was kicked, then deals
// its amount to each of them (Comet Storm: "Choose any target, then choose
// another target for each time this spell was kicked. Comet Storm deals X damage
// to each of them.").
//
// The parser folds the two-target preamble into a single any-target slot flagged
// KickerScaledCount and marks the "each of them" damage effect exact. This
// lowerer emits one variable-count any-target spec carrying
// CountEqualsKickerPlusOne, so cast-time validation requires exactly 1 + kicker
// targets, and a single EachTarget Damage primitive that deals the full amount
// independently to every still-legal chosen target at resolution.
//
// It fails closed (ok=false) for any content that is not exactly this shape — a
// non-kicker-scaled target, a wrong recipient, a dynamic/non-positive amount,
// conditions, modes, keywords, or a damage source it cannot represent — leaving
// such wordings to the remaining damage lowerers and their diagnostics.
func lowerKickerScaledEachTargetDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	if len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if !target.KickerScaledCount {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	amount, amountOK := eachOfDamageAmount(effect.Amount)
	if !effect.Exact ||
		effect.Negated ||
		!amountOK ||
		effect.DamageRecipient.Reference != parser.DamageRecipientReferenceNone ||
		(effect.Context != parser.EffectContextSource &&
			effect.Context != parser.EffectContextReferencedObject) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	spec, ok := kickerScaledDamageTargetSpec(target)
	if !ok || !exactDamageSourceSyntax(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{{Primitive: game.Damage{
			Amount:       amount,
			Recipient:    game.AnyTargetDamageRecipient(0),
			EachTarget:   true,
			DamageSource: primaryDamageSource(ctx.content.References),
		}}},
	}.Ability(), true
}

// kickerScaledDamageTargetSpec builds the variable-count target spec a
// kicker-scaled "each of them" damage spell chooses among. The wording's own
// slot is a single "any target" (permanent or player) with cardinality 1; this
// widens it to the [1, maxKickerScaledTargets] range and flags
// CountEqualsKickerPlusOne so action generation and cast-time validation bind
// the legal count to 1 + kicker. It supports only the unfiltered any-target slot,
// failing closed for every other selector.
func kickerScaledDamageTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if target.Selector.Kind != compiler.SelectorAny {
		return game.TargetSpec{}, false
	}
	if selectorHasUnsupportedPermanentFilters(target.Selector) ||
		selectorHasCounterQualifier(target.Selector) ||
		selectorHasAttachmentQualifier(target.Selector) ||
		len(target.Selector.SubtypesAny()) != 0 ||
		len(target.Selector.ColorsAny()) != 0 ||
		len(target.Selector.ExcludedTypes()) != 0 ||
		len(target.Selector.ExcludedColors()) != 0 ||
		len(target.Selector.Supertypes()) != 0 {
		return game.TargetSpec{}, false
	}
	return game.TargetSpec{
		MinTargets:               1,
		MaxTargets:               maxKickerScaledTargets,
		Constraint:               target.Text,
		Allow:                    game.TargetAllowPermanent | game.TargetAllowPlayer,
		CountEqualsKickerPlusOne: true,
	}, true
}
