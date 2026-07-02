package cardgen

import (
	"fmt"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerEachSourceDamageSpell lowers an "each <group> deals N damage to its
// controller/owner" effect ("Each creature deals 1 damage to its controller.")
// onto a GroupSourceDamage primitive: every member of the battlefield group is
// the damage source dealing the amount to the player who controls (or owns) it.
// It fails closed (ok=false) for every other shape, leaving the standard damage
// paths to handle their own effects.
func lowerEachSourceDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	if effect.DamageRecipient.EachSourceRole == parser.DamageRecipientReferenceNone ||
		effect.DamageRecipient.EachSourceRole == parser.DamageRecipientReferenceItself ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	amount, ok := groupDamageAmount(effect.Amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	group, ok := damageGroupRecipient(effect.DamageRecipient.EachSourceGroup)
	if !ok {
		return game.AbilityContent{}, false
	}
	primitive := game.GroupSourceDamage{
		Group:   group,
		Amount:  amount,
		ToOwner: effect.DamageRecipient.EachSourceRole == parser.DamageRecipientReferenceOwner,
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: primitive}},
	}.Ability(), true
}

// lowerEachSelfPowerDamageSpell lowers the group self-power damage shape in which
// every member of an "each <group>" subject deals damage to itself equal to its
// own power ("Each creature deals damage to itself equal to its power.", Wave of
// Reckoning; "Each tapped creature deals damage to itself equal to its power.",
// The Akroan War chapter III). The subject group is recorded as
// EachSourceDamageGroup with the per-member self recipient, and the per-member
// power is the amount, so the group reuses the same SelectionForSelector-backed
// damage group recipient as fixed group damage. The dealing entity must have a
// power, so the group is restricted to creatures. It fails closed (ok=false) for
// every other recipient role, amount, or shape, leaving the fixed/X and source-
// power damage paths unchanged.
func lowerEachSelfPowerDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	if effect.DamageRecipient.EachSourceRole != parser.DamageRecipientReferenceItself ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower ||
		effect.Negated ||
		effect.DamageRecipient.EachSourceGroup.Kind != compiler.SelectorCreature ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	group, ok := damageGroupRecipient(effect.DamageRecipient.EachSourceGroup)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: game.GroupSelfPowerDamage{Group: group}}},
	}.Ability(), true
}

func lowerGroupDamageSpell(
	_ string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	amount, amountOK := groupDamageAmountForContext(ctx, effect.Amount)
	if !amountOK ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed or X group damage amounts",
		)
	}
	if len(ctx.content.References) == 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	if _, ok := lowerDamageSourceReference(ctx.content.References[:1]); !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	sel := effect.Selector
	recipientSelectors := []compiler.CompiledSelector{sel}
	if len(effect.DamageRecipient.GroupSelectors) > 0 {
		recipientSelectors = effect.DamageRecipient.GroupSelectors
	}
	recipients := make([]game.DamageRecipient, 0, len(recipientSelectors))
	for _, recipientSel := range recipientSelectors {
		recipient, ok := groupDamageRecipientFor(recipientSel)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend does not support this group recipient",
			)
		}
		recipients = append(recipients, recipient)
	}
	if !effect.Exact || !exactDamageSourceSyntax(ctx.content.References) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	damageSourceRef := primaryDamageSource(ctx.content.References)
	instructions := make([]game.Instruction, 0, len(recipients))
	for _, recipient := range recipients {
		damage := game.Damage{
			Amount:       amount,
			Recipient:    recipient,
			DamageSource: damageSourceRef,
		}
		instructions = append(instructions, game.Instruction{Primitive: damage})
	}
	return game.Mode{
		Sequence: instructions,
	}.Ability(), nil
}

// groupDamageAmountForContext resolves a group-damage amount, additionally
// accepting a "that much"/"that many" triggering-event anaphor ("deals that much
// damage to each opponent.") and binding it to whichever event fired the
// enclosing triggered ability. The quantity is a single value shared by every
// recipient (CR 603.3e resolves the triggering event's quantity once), so it
// lowers to the same group-wide Quantity path as the other dynamic amounts.
// Outside a triggered context the anaphor has no source and stays rejected, so
// the helper falls back to groupDamageAmount.
func groupDamageAmountForContext(ctx contentCtx, amount compiler.CompiledAmount) (game.Quantity, bool) {
	if triggeringEventQuantityKind(amount.DynamicKind) {
		dynamic, ok := lowerTriggeringEventQuantityAmount(ctx, amount)
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	}
	return groupDamageAmount(amount)
}

// groupDamageAmount resolves the supported group-damage amounts onto a runtime
// Quantity: an exact fixed amount of at least one, the spell's X, or a dynamic
// count amount ("equal to the number of ..." / "where X is the number of ...").
// The executable backend deals the resolved amount to every member of each
// recipient group; a fixed or X amount needs no per-recipient computation, and a
// dynamic count amount is computed once against the battlefield and reused for
// every recipient. It fails closed for a zero or negative fixed amount and for
// every dynamic amount form the group path cannot reconstruct exactly, leaving
// those spells rejected.
func groupDamageAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	if amount.DynamicKind != compiler.DynamicAmountNone ||
		amount.DynamicForm != compiler.DynamicAmountFormNone {
		return groupDynamicDamageAmount(amount)
	}
	switch {
	case amount.Known:
		if amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(amount.Value), true
	case amount.VariableX:
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	default:
		return game.Quantity{}, false
	}
}

// groupDynamicDamageAmount resolves a dynamic group-damage amount ("deals X
// damage to each creature, where X is the number of creatures on the
// battlefield.", "Gates Ablaze deals X damage to each creature, where X is the
// number of Gates you control.", "Fanatic of Mogis deals damage to each
// opponent equal to your devotion to red.") onto a runtime Quantity. The amount
// is resolved once against the game state and dealt to every recipient, so it
// reuses lowerDynamicAmount. It accepts only group-wide amount kinds whose value
// is shared by every recipient (count selectors, devotion, domain, controller
// life, opponent count, and greatest-in-group); the per-object forms ("equal to
// its power", counters on a referenced object) need a per-object reference that
// has no single group-wide value, so they stay rejected and the group path
// remains fail-closed.
func groupDynamicDamageAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	if !groupWideDynamicAmountKind(amount.DynamicKind) {
		return game.Quantity{}, false
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

// groupWideDynamicAmountKind reports whether a dynamic amount kind resolves to a
// single game-state value that every member of a damage or life-loss group
// shares. These amounts are computed once and reused for every recipient. It
// fails closed for the per-object forms (source power/toughness/mana value and
// counters on a referenced object), which have no single group-wide value.
func groupWideDynamicAmountKind(kind compiler.DynamicAmountKind) bool {
	switch kind {
	case compiler.DynamicAmountCount,
		compiler.DynamicAmountControllerLife,
		compiler.DynamicAmountOpponentCount,
		compiler.DynamicAmountBasicLandTypes,
		compiler.DynamicAmountDevotion,
		compiler.DynamicAmountGreatestPower,
		compiler.DynamicAmountGreatestToughness,
		compiler.DynamicAmountGreatestManaValue,
		compiler.DynamicAmountTotalPower,
		compiler.DynamicAmountTotalToughness,
		compiler.DynamicAmountTotalManaValue,
		compiler.DynamicAmountColorCount:
		return true
	default:
		return false
	}
}

// groupDamageRecipientFor resolves one fixed group-damage recipient selector
// onto a runtime DamageRecipient: the all-players and opponents player groups,
// or a filtered battlefield permanent group. It fails closed for any selector
// the executable backend cannot damage as a group so unsupported recipients stay
// rejected.
func groupDamageRecipientFor(sel compiler.CompiledSelector) (game.DamageRecipient, bool) {
	return groupDamageRecipientForExcluding(sel, game.SourcePermanentReference())
}

// groupDamageRecipientForExcluding resolves one group-damage recipient selector
// like groupDamageRecipientFor, but excludes the supplied object from an "other"
// permanent group instead of the spell's own source permanent. It backs the
// source-power group damage shape ("Target creature you control deals damage
// equal to its power to each other creature and each opponent."), where "each
// other creature" excludes the dealing target rather than the spell.
func groupDamageRecipientForExcluding(sel compiler.CompiledSelector, exclude game.ObjectReference) (game.DamageRecipient, bool) {
	switch {
	case sel.Kind == compiler.SelectorOpponent && !sel.Other:
		return game.PlayerGroupDamageRecipient(game.OpponentsReference()), true
	case sel.Kind == compiler.SelectorPlayer && !sel.Other:
		return game.PlayerGroupDamageRecipient(game.AllPlayersReference()), true
	default:
		group, ok := damageGroupRecipientExcluding(sel, exclude)
		if !ok {
			return game.DamageRecipient{}, false
		}
		return game.GroupDamageRecipient(group), true
	}
}

// damageGroupRecipient maps a compiled group-damage recipient selector onto a
// battlefield group reference, excluding the spell's own source when the
// recipient is an "other" group. It mirrors the parser's
// exactGroupDamagePermanentRecipientText reconstruction so the executable
// backend and the exactness gate accept exactly the same filtered recipients.
func damageGroupRecipient(sel compiler.CompiledSelector) (game.GroupReference, bool) {
	return damageGroupRecipientExcluding(sel, game.SourcePermanentReference())
}

// damageGroupRecipientExcluding maps a compiled group-damage recipient selector
// onto a battlefield group reference like damageGroupRecipient, but excludes the
// supplied object from an "other" group instead of the spell's own source.
func damageGroupRecipientExcluding(sel compiler.CompiledSelector, exclude game.ObjectReference) (game.GroupReference, bool) {
	selection, ok := damageGroupSelection(sel)
	if !ok {
		return game.GroupReference{}, false
	}
	if sel.Other {
		return game.BattlefieldGroupExcluding(selection, exclude), true
	}
	return game.BattlefieldGroup(selection), true
}

// damageGroupSelection translates the supported filters of a group-damage
// recipient selector (controller, combat, tapped, single color, a subtype union,
// single excluded type, keyword, named card, mana-value/power/toughness
// comparison) onto a runtime Selection, failing closed for any selector field it
// cannot represent exactly so unsupported recipients stay rejected. The guard
// enforces the recipient-specific accept set (single-valued color/excluded-type,
// a subtype union, no supertype or color filter, a damageable kind) that no
// SelectionMask dimension expresses; the numeric mana-value/power/toughness
// comparison reaches the canonical projector, which honors a fixed bound and
// fails closed on a chosen-{X} bound; damageGroupSelectionMask drops the
// remaining canonical dimensions a damage group never carries.
func damageGroupSelection(sel compiler.CompiledSelector) (game.Selection, bool) {
	requiredTypes := sel.RequiredTypesAny()
	unionTypes := len(requiredTypes) == 2 &&
		requiredTypes[0] == types.Creature && requiredTypes[1] == types.Planeswalker
	if sel.All || sel.Another || sel.Zone != zone.None ||
		sel.Colorless || sel.Multicolored ||
		(len(requiredTypes) != 0 && !unionTypes) ||
		len(sel.Supertypes()) != 0 ||
		len(sel.ExcludedColors()) != 0 ||
		len(sel.ColorsAny()) > 1 ||
		len(sel.ExcludedTypes()) > 1 {
		return game.Selection{}, false
	}
	if (sel.Attacking && sel.Blocking) ||
		(sel.Tapped && sel.Untapped) ||
		((sel.Tapped || sel.Untapped) && (sel.Attacking || sel.Blocking)) {
		return game.Selection{}, false
	}
	_, hasNoun, ok := damageGroupRequiredType(sel.Kind)
	if !ok {
		return game.Selection{}, false
	}
	if !hasNoun && len(sel.SubtypesAny()) == 0 {
		return game.Selection{}, false
	}
	return SelectionForSelectorMasked(sel, damageGroupSelectionMask)
}

// damageGroupSelectionMask drops the canonical dimensions a group-damage
// recipient never carries: the self-exclusion, excluded supertype, kind-agnostic
// counter, "aren't of the chosen type" exclusion, conjunctive type set, and
// historic disjunction. It honors per-object token state ("each nontoken
// creature"), the excluded creature subtype ("each non-Dragon creature"), and
// the named-card filter ("each other creature you control named Charmed Stray").
// It fails closed on a source-relative power comparison: a damage group has no
// source permanent to compare against, so the predecessor projector rejected
// that filter rather than dropping it.
var damageGroupSelectionMask = SelectionMask{}.Ignoring(
	DimExcludeSource,
	DimExcludedSupertype,
	DimMatchAnyCounter,
	DimSubtypeChoiceExcluded,
	DimConjunctiveTypes,
	DimHistoric,
).Rejecting(
	DimPowerVsSource,
)

// damageGroupRequiredType reports the battlefield required card type for a
// group-damage recipient selector kind. hasNoun is false for an unqualified
// subtype recipient ("each Goblin"), whose required type is left unset; ok is
// false for selector kinds the executable backend cannot damage as a group.
func damageGroupRequiredType(kind compiler.SelectorKind) (cardType types.Card, hasNoun, ok bool) {
	switch kind {
	case compiler.SelectorCreature:
		return types.Creature, true, true
	case compiler.SelectorPlaneswalker:
		return types.Planeswalker, true, true
	case compiler.SelectorArtifact:
		return types.Artifact, true, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true, true
	case compiler.SelectorLand:
		return types.Land, true, true
	case compiler.SelectorPermanent:
		return "", true, true
	case compiler.SelectorUnknown:
		return "", false, true
	default:
		return "", false, false
	}
}

// lowerSingleTargetDamageAmount resolves the game.Quantity for a single-target
// deal-damage effect, supporting fixed values, triggering-event quantities, and
// other dynamic-amount kinds (and defaulting to DynamicAmountX when no amount is
// specified). It returns false when the dynamic amount is unsupported. It is
// shared by the single-target damage spell lowerers so they resolve the damage
// amount identically.
func lowerSingleTargetDamageAmount(ctx contentCtx, effect compiler.CompiledEffect) (game.Quantity, bool) {
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	var damageSource game.ObjectReference
	var sourceBound bool
	if len(ctx.content.References) > 0 {
		damageSource, sourceBound = lowerDamageSourceReference(ctx.content.References[:1])
	}
	switch {
	case effect.Amount.Known:
		amount = game.Fixed(effect.Amount.Value)
	case triggeringEventQuantityKind(effect.Amount.DynamicKind):
		dynamic, ok := lowerTriggeringEventQuantityAmount(ctx, effect.Amount)
		if !ok {
			return game.Quantity{}, false
		}
		amount = game.Dynamic(dynamic)
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		amountObject := game.SourcePermanentReference()
		if sourceBound {
			amountObject = damageSource
		}
		if obj, ok := lowerDamageAmountObject(effect.Amount, ctx.content.References); ok {
			amountObject = obj
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, amountObject)
		if !ok {
			return game.Quantity{}, false
		}
		amount = game.Dynamic(dynamic)
	default:
		// No amount override: the damage defaults to X (DynamicAmountX).
	}
	return amount, true
}

func lowerFixedDamageSpell(
	_ string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		!effect.Exact ||
		(effect.Context != parser.EffectContextSource &&
			effect.Context != parser.EffectContextReferencedObject &&
			effect.Context != parser.EffectContextPriorSubject) ||
		(effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone ||
		effect.Negated ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Targets[0].Cardinality.Min < 0 ||
		ctx.content.Targets[0].Cardinality.Min > 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	amount, amountOK := lowerSingleTargetDamageAmount(ctx, effect)
	if !amountOK {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	if effect.DamageRecipient.Reference != parser.DamageRecipientReferenceNone {
		return lowerReferencedPlayerDamageSpell(ctx, effect.DamageRecipient.Reference, amount)
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	// A target-controller rider ("... and B damage to that creature's
	// controller") contributes a second, target-bound reference for the rider
	// recipient. The damage-source exactness checks only validate the spell's
	// own source reference (references[0]), so exclude the trailing rider
	// reference from them; it is validated separately by the rider lowering.
	sourceReferences := ctx.content.References
	if _, ok := parser.TargetControllerDamageRider(effect.DamageRiders); ok {
		if len(sourceReferences) != 2 ||
			sourceReferences[1].Binding != compiler.ReferenceBindingTarget {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		sourceReferences = sourceReferences[:1]
	}
	if !ok ||
		!exactDamageSourceSyntax(sourceReferences) ||
		!exactDamageAmountReferences(effect.Amount, sourceReferences) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}

	damage := game.Damage{
		Amount:    amount,
		Recipient: game.AnyTargetDamageRecipient(0),
	}
	damage.DamageSource = primaryDamageSource(ctx.content.References)
	if !damage.DamageSource.Exists &&
		effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
	}
	instructions := []game.Instruction{{Primitive: damage}}
	// "deals A damage to <target> and B damage to you" appends a second Damage
	// instruction dealing the fixed rider amount to the source's own controller.
	if selfRider, ok := parser.SelfDamageRider(effect.DamageRiders); ok {
		if !effect.Amount.Known || selfRider.Value < 1 {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		rider := game.Damage{
			Amount:       game.Fixed(selfRider.Value),
			Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
			DamageSource: damage.DamageSource,
		}
		instructions = append(instructions, game.Instruction{Primitive: rider})
	}
	// "deals A damage to target creature and B damage to that creature's
	// controller/owner" appends a second Damage instruction dealing the fixed
	// rider amount to the primary target's controller or owner.
	if tcRider, ok := parser.TargetControllerDamageRider(effect.DamageRiders); ok {
		riderRecipient, ok := targetControllerRiderRecipient(
			ctx.content.Targets[0], tcRider.ReferenceRole)
		if !ok || !effect.Amount.Known || tcRider.Value < 1 {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		rider := game.Damage{
			Amount:       game.Fixed(tcRider.Value),
			Recipient:    game.PlayerDamageRecipient(riderRecipient),
			DamageSource: damage.DamageSource,
		}
		instructions = append(instructions, game.Instruction{Primitive: rider})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{target},
		Sequence: instructions,
	}.Ability(), nil
}

// lowerTwoTargetDamageSpell lowers a "<source> deals A damage to <target0> and B
// damage to <target1>" spell that names two independently chosen single targets,
// as in "Hungry Flames deals 3 damage to target creature and 2 damage to target
// player or planeswalker." Both amounts are fixed (>= 1); it emits one Damage
// instruction per target keyed to occurrence 0 and 1 respectively. The variable
// form "<source> deals X damage to any target and X damage to any other target,
// where X is ..." (The Brothers' War chapter III) shares one dynamic amount
// across both instructions. It fails closed for any shape outside those templates
// (a missing rider, an unrecognized amount, a non-single-target cardinality, or
// any condition, keyword, or mode).
func lowerTwoTargetDamageSpell(
	_ string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	secondRider, hasSecondRider := parser.SecondTargetDamageRider(effect.DamageRiders)
	if !effect.Exact ||
		!hasSecondRider ||
		(effect.Context != parser.EffectContextSource &&
			effect.Context != parser.EffectContextReferencedObject &&
			effect.Context != parser.EffectContextPriorSubject) ||
		effect.Negated ||
		len(ctx.content.Targets) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	// The rider amount is either a fixed value B (>= 1) shared with a known
	// primary amount A (>= 1), or the variable "X" that reuses the clause's
	// single dynamic amount for both targets (The Brothers' War chapter III).
	dynamicRider := secondRider.Dynamic
	if dynamicRider {
		if effect.Amount.DynamicKind == compiler.DynamicAmountNone {
			return unsupported()
		}
	} else if !effect.Amount.Known || effect.Amount.Value < 1 ||
		secondRider.Value < 1 {
		return unsupported()
	}
	for i := range ctx.content.Targets {
		if ctx.content.Targets[i].Cardinality.Max != 1 ||
			ctx.content.Targets[i].Cardinality.Min < 0 ||
			ctx.content.Targets[i].Cardinality.Min > 1 {
			return unsupported()
		}
	}
	target0, ok0 := damageTargetSpec(ctx.content.Targets[0])
	// "... and B damage to any other target" names a second "any target" slot
	// whose "other" qualifier requires a different object from the first slot.
	// damageTargetSpec rejects the bare "other"/"another" any-target form (its
	// single-target meaning excludes the source, not a prior target), so the
	// distinctness is applied here, scoped to this prior-target context.
	second := ctx.content.Targets[1]
	secondDistinct := false
	if second.Selector.Kind == compiler.SelectorAny &&
		(second.Selector.Other || second.Selector.Another) {
		second.Selector.Other = false
		second.Selector.Another = false
		secondDistinct = true
	}
	target1, ok1 := damageTargetSpec(second)
	sourceReferences := ctx.content.References
	if len(sourceReferences) > 1 {
		sourceReferences = sourceReferences[:1]
	}
	if !ok0 || !ok1 ||
		ctx.content.Targets[0].Selector.Other ||
		ctx.content.Targets[0].Selector.Another ||
		!exactDamageSourceSyntax(sourceReferences) ||
		!exactDamageAmountReferences(effect.Amount, sourceReferences) {
		return unsupported()
	}
	if secondDistinct {
		target1.DistinctFromPriorTargets = true
	}
	var damageSource opt.V[game.ObjectReference]
	if damageSourceIsSourcePermanent(sourceReferences) {
		damageSource = opt.Val(game.SourcePermanentReference())
	}
	primaryAmount := game.Fixed(effect.Amount.Value)
	riderAmount := game.Fixed(secondRider.Value)
	if dynamicRider {
		amountObject := game.SourcePermanentReference()
		if obj, ok := lowerDamageAmountObject(effect.Amount, ctx.content.References); ok {
			amountObject = obj
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, amountObject)
		if !ok {
			return unsupported()
		}
		primaryAmount = game.Dynamic(dynamic)
		riderAmount = game.Dynamic(dynamic)
	}
	primary := game.Damage{
		Amount:       primaryAmount,
		Recipient:    game.AnyTargetDamageRecipient(0),
		DamageSource: damageSource,
	}
	rider := game.Damage{
		Amount:       riderAmount,
		Recipient:    game.AnyTargetDamageRecipient(1),
		DamageSource: damageSource,
	}
	return game.Mode{
		Targets: []game.TargetSpec{target0, target1},
		Sequence: []game.Instruction{
			{Primitive: primary},
			{Primitive: rider},
		},
	}.Ability(), nil
}

// targetControllerRiderRecipient resolves the recipient player of a "... and B
// damage to that creature's controller/owner" rider: the controller or owner of
// the clause's sole permanent target (occurrence 0). It fails closed for a
// non-permanent target, leaving the rider rejected.
func targetControllerRiderRecipient(
	target compiler.CompiledTarget,
	kind parser.DamageRecipientReferenceKind,
) (game.PlayerReference, bool) {
	var object game.ObjectReference
	switch target.Selector.Kind {
	case compiler.SelectorArtifact, compiler.SelectorCreature, compiler.SelectorEnchantment,
		compiler.SelectorLand, compiler.SelectorPermanent, compiler.SelectorPlaneswalker,
		compiler.SelectorBattle:
		object = game.TargetPermanentReference(0)
	default:
		return game.PlayerReference{}, false
	}
	switch kind {
	case parser.DamageRecipientReferenceController:
		return game.ObjectControllerReference(object), true
	case parser.DamageRecipientReferenceOwner:
		return game.ObjectOwnerReference(object), true
	default:
		return game.PlayerReference{}, false
	}
}

// lowerReferencedPlayerDamageSpell lowers a damage effect whose recipient is the
// controller or owner of the prior removal target in an ordered sequence, as in
// "Destroy target land. Melt Terrain deals 2 damage to that land's controller."
// The inherited removal target arrives as the clause's sole target and the
// recipient reference ("that land's"/"its", controller/owner) binds to it. The
// damage instruction keeps the inherited target so the sequence machinery can
// rebase it; the recipient resolves to that target's controller or owner. It
// fails closed for any other shape.
func lowerReferencedPlayerDamageSpell(
	ctx contentCtx,
	recipientKind parser.DamageRecipientReferenceKind,
	amount game.Quantity,
) (game.AbilityContent, *shared.Diagnostic) {
	recipient, ok := referencedDamageRecipientPlayer(ctx, recipientKind)
	target, targetOK := removalTargetSpecForRecipient(ctx.content.Targets[0])
	if !ok || !targetOK || len(ctx.content.References) == 0 ||
		!exactDamageSourceSyntax(ctx.content.References[:1]) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	damage := game.Damage{
		Amount:       amount,
		Recipient:    game.PlayerDamageRecipient(recipient),
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

// referencedDamageRecipientPlayer resolves the recipient player for a damage
// effect aimed at the controller or owner of the inherited removal target. The
// recipient reference is the sole target-bound reference in the clause ("that
// land's"/"its"); its occurrence indexes the inherited target. The target's
// selector drives the object reference: a permanent target (including bare
// subtype and compound type leads) yields a permanent reference, a spell target
// a stack-object reference. It fails closed for any other shape.
func referencedDamageRecipientPlayer(
	ctx contentCtx,
	kind parser.DamageRecipientReferenceKind,
) (game.PlayerReference, bool) {
	if len(ctx.content.Targets) != 1 {
		return game.PlayerReference{}, false
	}
	var recipientRef *compiler.CompiledReference
	for i := range ctx.content.References {
		if ctx.content.References[i].Binding != compiler.ReferenceBindingTarget {
			continue
		}
		if recipientRef != nil {
			return game.PlayerReference{}, false
		}
		recipientRef = &ctx.content.References[i]
	}
	if recipientRef == nil || recipientRef.Occurrence < 0 {
		return game.PlayerReference{}, false
	}
	occ := recipientRef.Occurrence
	object, ok := inheritedRemovalTargetObjectRef(ctx.content.Targets[0], occ)
	if !ok {
		return game.PlayerReference{}, false
	}
	switch kind {
	case parser.DamageRecipientReferenceController:
		return game.ObjectControllerReference(object), true
	case parser.DamageRecipientReferenceOwner:
		return game.ObjectOwnerReference(object), true
	default:
		return game.PlayerReference{}, false
	}
}

// lowerControllerDamageSpell lowers a "deals N damage to you" effect, whose
// recipient is the source's own controller, as in "this creature deals 1 damage
// to you." or "Sell-Sword Brute deals 2 damage to you." The recipient binds to
// the resolving ability's controller; the amount is a fixed value or X. It emits
// one Damage instruction with a controller player recipient and no target spec,
// failing closed for any shape outside that exact template (a non-"you"
// recipient, a dynamic count amount, any target, condition, keyword, or mode).
func lowerControllerDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	assertUndividedRecipientDamageDispatch(ctx, parser.DamageRecipientReferenceYou)
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed or X damage to you",
		)
	}
	if !effect.Exact ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		(!effect.Amount.Known && !effect.Amount.VariableX) ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		return unsupported()
	}
	if !exactDamageSourceSyntax(ctx.content.References) {
		return unsupported()
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
	}
	damage := game.Damage{
		Amount:       amount,
		Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
		DamageSource: primaryDamageSource(ctx.content.References),
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), nil
}

// lowerEventPlayerDamageSpell lowers a "deals N damage to that player" effect
// whose recipient is the triggering event's player, as in "Whenever an opponent
// draws a card, this enchantment deals 1 damage to that player." (Underworld
// Dreams, Fate Unraveler, Megrim, Manabarbs). The "that player" reference binds
// to the event player (ReferenceBindingEventPlayer). The amount is a fixed
// value, X, or a resolvable dynamic amount: "deals damage equal to its power to
// that player" (Gleeful Arsonist) reads the source's power and deals it to the
// event player. It emits one Damage instruction with an event-player recipient
// and no target spec, failing closed for any shape outside that template (a
// recipient that is not the event-bound "that player", an unresolvable amount,
// any target, recipient selector, condition, keyword, or mode).
func lowerEventPlayerDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	assertUndividedRecipientDamageDispatch(ctx, parser.DamageRecipientReferenceThatPlayer)
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed, X, or source-power damage to that player",
		)
	}
	if !effect.Exact ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	// The event-player recipient ("that player") contributes its own reference;
	// exclude it so the damage-source and amount exactness checks see only the
	// source clause's references (the damage subject and the amount referent).
	sourceReferences := make([]compiler.CompiledReference, 0, len(ctx.content.References))
	for _, reference := range ctx.content.References {
		if reference.Kind == compiler.ReferenceThatPlayer &&
			reference.Binding == compiler.ReferenceBindingEventPlayer {
			continue
		}
		sourceReferences = append(sourceReferences, reference)
	}
	if !damageReferencesEventPlayer(ctx.content.References) ||
		len(sourceReferences) == 0 ||
		!exactDamageSourceSyntax(sourceReferences[:1]) {
		return unsupported()
	}
	amount, ok := eventPlayerDamageAmount(effect, sourceReferences)
	if !ok {
		return unsupported()
	}
	damage := game.Damage{
		Amount:       amount,
		Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
		DamageSource: primaryDamageSource(sourceReferences[:1]),
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), nil
}

// eventPlayerDamageAmount resolves the quantity an event-player damage effect
// deals. A known fixed value (>= 1) and bare X lower directly; a dynamic amount
// resolves through the source-power machinery, binding to the amount's own
// referent and validating its references exactly so "deals damage equal to its
// power to that player" reads the source's power. It fails closed for a
// non-positive fixed value, a non-X unknown amount, or any dynamic amount the
// shared resolver cannot represent exactly.
func eventPlayerDamageAmount(
	effect compiler.CompiledEffect,
	sourceReferences []compiler.CompiledReference,
) (game.Quantity, bool) {
	if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		amountObject, ok := lowerDamageAmountObject(effect.Amount, sourceReferences)
		if !ok {
			return game.Quantity{}, false
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, amountObject)
		if !ok || !exactDamageAmountReferences(effect.Amount, sourceReferences) {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	}
	if effect.Amount.Known {
		if effect.Amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(effect.Amount.Value), true
	}
	if !effect.Amount.VariableX {
		return game.Quantity{}, false
	}
	return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
}

// damageReferencesEventPlayer reports whether the references carry the
// event-bound "that player" recipient reference that backs "deals N damage to
// that player." The recipient reference is a ReferenceThatPlayer whose binding
// resolves to the triggering event's player (ReferenceBindingEventPlayer).
func damageReferencesEventPlayer(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind == compiler.ReferenceThatPlayer &&
			reference.Binding == compiler.ReferenceBindingEventPlayer {
			return true
		}
	}
	return false
}

// lowerEventRelatedPermanentDamageSpell lowers a "deals N damage to that
// creature" effect whose recipient is the triggering event's related combat
// permanent, as in "Whenever this creature blocks or becomes blocked by a
// creature, this creature deals 3 damage to that creature." (Inferno Elemental).
// The "that creature" reference binds to the event's related permanent
// (ReferenceBindingEventRelatedPermanent), the opposing combatant the runtime
// resolves through EventRelatedPermanentReference. The amount is a fixed value,
// X, or a resolvable dynamic amount. It emits one Damage instruction with the
// related-permanent recipient and no target spec, failing closed for any shape
// outside that template (a recipient that is not the event-bound "that creature",
// an unresolvable amount, any target, recipient selector, condition, keyword, or
// mode).
func lowerEventRelatedPermanentDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	assertUndividedRecipientDamageDispatch(ctx, parser.DamageRecipientReferenceThatCreature)
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed, X, or source-power damage to that creature",
		)
	}
	if !effect.Exact ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	// The related-permanent recipient ("that creature") contributes its own
	// reference; exclude it so the damage-source and amount exactness checks see
	// only the source clause's references (the damage subject and amount referent).
	sourceReferences := make([]compiler.CompiledReference, 0, len(ctx.content.References))
	for _, reference := range ctx.content.References {
		if reference.Kind == compiler.ReferenceThatObject &&
			reference.Binding == compiler.ReferenceBindingEventRelatedPermanent {
			continue
		}
		sourceReferences = append(sourceReferences, reference)
	}
	if !damageReferencesEventRelatedPermanent(ctx.content.References) ||
		len(sourceReferences) == 0 ||
		!exactDamageSourceSyntax(sourceReferences[:1]) {
		return unsupported()
	}
	amount, ok := eventPlayerDamageAmount(effect, sourceReferences)
	if !ok {
		return unsupported()
	}
	damage := game.Damage{
		Amount:       amount,
		Recipient:    game.ObjectDamageRecipient(game.EventRelatedPermanentReference()),
		DamageSource: primaryDamageSource(sourceReferences[:1]),
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), nil
}

// damageReferencesEventRelatedPermanent reports whether the references carry the
// event-bound "that creature" recipient reference that backs "deals N damage to
// that creature." The recipient reference is a ReferenceThatObject whose binding
// resolves to the triggering event's related combat permanent
// (ReferenceBindingEventRelatedPermanent).
func damageReferencesEventRelatedPermanent(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind == compiler.ReferenceThatObject &&
			reference.Binding == compiler.ReferenceBindingEventRelatedPermanent {
			return true
		}
	}
	return false
}

// returned spec is discarded (the removal clause already contributes it); the
// damage Mode only needs a valid, non-empty target spec so the sequence machinery
// rebases the recipient reference. A spell target yields the stack-spell spec;
// any other removal target yields the permanent spec. It fails closed otherwise.
func removalTargetSpecForRecipient(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if target.Selector.Kind == compiler.SelectorSpell {
		return stackSpellTargetSpec(target)
	}
	return permanentTargetSpec(target)
}

// inheritedTargetSubjectReference reports whether reference is a singular
// inherited back-reference to a prior effect's target, the antecedent subject of
// a follow-up clause. Both the pronoun "it" and the object demonstrative "that
// creature" name that antecedent ("... gets +N/+N until end of turn. It deals
// ..." vs "Put a +1/+1 counter on target creature you control. Then that creature
// deals ..."); the compiler binds either to the target with a non-negative
// occurrence index. It fails closed for any other reference kind or binding.
func inheritedTargetSubjectReference(reference compiler.CompiledReference) bool {
	isIt := reference.Kind == compiler.ReferencePronoun &&
		reference.Pronoun == compiler.ReferencePronounIt
	isThatCreature := reference.Kind == compiler.ReferenceThatObject
	return (isIt || isThatCreature) &&
		reference.Binding == compiler.ReferenceBindingTarget &&
		reference.Occurrence >= 0
}

// lowerInheritedPowerDamageSpell lowers an inherited "it deals damage equal to
// its power to <target>" effect, where "it" refers to a prior effect's target
// (e.g. Clear Shot / Rabid Gnaw: "Target creature you control gets +N/+N until
// end of turn. It deals damage equal to its power to target creature you don't
// control."). The damage source and the dynamic power amount both bind to the
// inherited antecedent target; the recipient is this effect's own target.
//
// This handles only the two-target inherited shape and fails closed (ok=false)
// otherwise, leaving lowerFixedDamageSpell's single-target form byte-identical.
func lowerInheritedPowerDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextReferencedObject ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower {
		return game.AbilityContent{}, false
	}
	if len(ctx.content.Targets) != 2 ||
		len(ctx.content.References) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	source := ctx.content.References[0]
	amountRef := ctx.content.References[1]
	if !inheritedTargetSubjectReference(source) ||
		source.Occurrence >= len(ctx.content.Targets) {
		return game.AbilityContent{}, false
	}
	if amountRef.Kind != compiler.ReferencePronoun ||
		amountRef.Binding != compiler.ReferenceBindingTarget ||
		amountRef.Occurrence != source.Occurrence ||
		amountRef.Span != effect.Amount.ReferenceSpan {
		return game.AbilityContent{}, false
	}
	sourceIdx := source.Occurrence
	recipientIdx := 0
	if sourceIdx == 0 {
		recipientIdx = 1
	}
	sourceRef, ok := lowerObjectReference(source, referenceLoweringContext{AllowTarget: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, sourceRef)
	if !ok {
		return game.AbilityContent{}, false
	}
	sourceSpec, ok := damageTargetSpec(ctx.content.Targets[sourceIdx])
	if !ok {
		return game.AbilityContent{}, false
	}
	recipientSpec, ok := damageTargetSpec(ctx.content.Targets[recipientIdx])
	if !ok {
		return game.AbilityContent{}, false
	}
	specs := make([]game.TargetSpec, 2)
	specs[sourceIdx] = sourceSpec
	specs[recipientIdx] = recipientSpec
	return game.Mode{
		Targets: specs,
		Sequence: []game.Instruction{{
			Primitive: game.Damage{
				Amount:       game.Dynamic(dynamic),
				Recipient:    game.AnyTargetDamageRecipient(recipientIdx),
				DamageSource: opt.Val(sourceRef),
			},
		}},
	}.Ability(), true
}

// lowerSourcePowerDamageSpell lowers the one-sided source-power damage effects in
// which a target creature deals damage equal to its own power. Two shapes are
// supported: the self form "Target creature deals damage to itself equal to its
// power." (one target that is both the damage source and the recipient) and the
// two-target form "Target creature you control deals damage equal to its power
// to target creature you don't control." (the first target deals, the second
// receives). The dealing creature is identified by the occurrence of the single
// "its power" reference; its power feeds the dynamic amount and it is the damage
// source so its keywords (deathtouch, lifelink) apply at resolution. This shape
// differs from lowerInheritedPowerDamageSpell (an inherited "it" subject carried
// from a prior effect) in that the dealing creature is the clause's own target.
// It fails closed (ok=false) for every other shape, leaving lowerFixedDamageSpell
// and its diagnostic unchanged.
func lowerSourcePowerDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextTarget ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.References) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	powerRef := ctx.content.References[0]
	if powerRef.Kind != compiler.ReferencePronoun ||
		powerRef.Pronoun != compiler.ReferencePronounIts ||
		powerRef.Binding != compiler.ReferenceBindingTarget ||
		powerRef.Occurrence < 0 ||
		powerRef.Occurrence >= len(ctx.content.Targets) ||
		powerRef.Span != effect.Amount.ReferenceSpan {
		return game.AbilityContent{}, false
	}
	sourceIdx := powerRef.Occurrence
	sourceRef := game.TargetPermanentReference(sourceIdx)
	dynamic, ok := lowerDynamicAmount(effect.Amount, sourceRef)
	if !ok {
		return game.AbilityContent{}, false
	}
	damage := game.Damage{
		Amount:       game.Dynamic(dynamic),
		DamageSource: opt.Val(sourceRef),
	}
	switch len(ctx.content.Targets) {
	case 1:
		if sourceIdx != 0 {
			return game.AbilityContent{}, false
		}
		spec, ok := damageTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, false
		}
		damage.Recipient = game.AnyTargetDamageRecipient(0)
		return game.Mode{
			Targets:  []game.TargetSpec{spec},
			Sequence: []game.Instruction{{Primitive: damage}},
		}.Ability(), true
	case 2:
		recipientIdx := 1 - sourceIdx
		sourceSpec, ok := damageTargetSpec(ctx.content.Targets[sourceIdx])
		if !ok {
			return game.AbilityContent{}, false
		}
		// "deals damage equal to its power to each of N other target creatures"
		// (Betrayal at the Vault): the recipient is a plural slot dealt the
		// dealer's power once per chosen target. Unroll one Damage instruction
		// per recipient slot, keyed past the single source slot.
		if ctx.content.Targets[recipientIdx].Cardinality.Max >= 2 {
			if sourceIdx != 0 {
				return game.AbilityContent{}, false
			}
			recipientSpec, ok := eachOfDamageTargetSpec(ctx.content.Targets[recipientIdx])
			if !ok {
				return game.AbilityContent{}, false
			}
			sequence := make([]game.Instruction, 0, recipientSpec.MaxTargets)
			for i := range recipientSpec.MaxTargets {
				sequence = append(sequence, game.Instruction{Primitive: game.Damage{
					Amount:       game.Dynamic(dynamic),
					Recipient:    game.AnyTargetDamageRecipient(1 + i),
					DamageSource: opt.Val(sourceRef),
				}})
			}
			return game.Mode{
				Targets:  []game.TargetSpec{sourceSpec, recipientSpec},
				Sequence: sequence,
			}.Ability(), true
		}
		recipientSpec, ok := damageTargetSpec(ctx.content.Targets[recipientIdx])
		if !ok {
			return game.AbilityContent{}, false
		}
		specs := make([]game.TargetSpec, 2)
		specs[sourceIdx] = sourceSpec
		specs[recipientIdx] = recipientSpec
		damage.Recipient = game.AnyTargetDamageRecipient(recipientIdx)
		return game.Mode{
			Targets:  specs,
			Sequence: []game.Instruction{{Primitive: damage}},
		}.Ability(), true
	default:
		return game.AbilityContent{}, false
	}
}

// lowerSourcePowerGroupDamageSpell lowers the source-power group damage shape in
// which a chosen target creature deals damage equal to its own power to a
// compound group of recipients ("Target creature you control deals damage equal
// to its power to each other creature and each opponent."). The single target is
// the damage source; its power feeds the dynamic amount and it is the damage
// source so its keywords (deathtouch, lifelink) apply. Each recipient group in
// the pair becomes its own Damage instruction, and an "each other creature"
// group excludes the dealing target rather than the spell's own source. It fails
// closed (ok=false) for every other shape, leaving the single-target source-power
// and group-damage paths unchanged.
func lowerSourcePowerGroupDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextTarget ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower ||
		len(effect.DamageRecipient.GroupSelectors) == 0 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	powerRef := ctx.content.References[0]
	if powerRef.Kind != compiler.ReferencePronoun ||
		powerRef.Pronoun != compiler.ReferencePronounIts ||
		powerRef.Binding != compiler.ReferenceBindingTarget ||
		powerRef.Occurrence != 0 ||
		powerRef.Span != effect.Amount.ReferenceSpan {
		return game.AbilityContent{}, false
	}
	sourceRef := game.TargetPermanentReference(0)
	dynamic, ok := lowerDynamicAmount(effect.Amount, sourceRef)
	if !ok {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	instructions := make([]game.Instruction, 0, len(effect.DamageRecipient.GroupSelectors))
	for _, sel := range effect.DamageRecipient.GroupSelectors {
		recipient, ok := groupDamageRecipientForExcluding(sel, sourceRef)
		if !ok {
			return game.AbilityContent{}, false
		}
		instructions = append(instructions, game.Instruction{
			Primitive: game.Damage{
				Amount:       game.Dynamic(dynamic),
				Recipient:    recipient,
				DamageSource: opt.Val(sourceRef),
			},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: instructions,
	}.Ability(), true
}

// lowerInheritedPowerGroupDamageSpell lowers the inherited-subject source-power
// group damage shape in which a creature carried from a prior clause ("it",
// bound to that clause's target) deals damage equal to its own power to a single
// group of recipients ("Choose target creature you control. It deals damage
// equal to its power to each other creature.", Nibelheim Aflame). The inherited
// target is the damage source: its power feeds the dynamic amount and it is the
// damage source so its keywords (deathtouch, lifelink) apply. The lone recipient
// group lives in effect.Selector (single-recipient damage, so
// DamageRecipientSelectors is empty), and an "each other creature" group
// excludes the dealing target rather than the spell's own source. It fails
// closed (ok=false) for every other shape, leaving the two-target inherited and
// dual-recipient source-power paths unchanged.
func lowerInheritedPowerGroupDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextReferencedObject ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		effect.DamageRecipient.Reference != parser.DamageRecipientReferenceNone ||
		len(effect.DamageRiders) != 0 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	source := ctx.content.References[0]
	amountRef := ctx.content.References[1]
	if source.Kind != compiler.ReferencePronoun ||
		source.Pronoun != compiler.ReferencePronounIt ||
		source.Binding != compiler.ReferenceBindingTarget ||
		source.Occurrence != 0 {
		return game.AbilityContent{}, false
	}
	if amountRef.Kind != compiler.ReferencePronoun ||
		amountRef.Pronoun != compiler.ReferencePronounIts ||
		amountRef.Binding != compiler.ReferenceBindingTarget ||
		amountRef.Occurrence != 0 ||
		amountRef.Span != effect.Amount.ReferenceSpan {
		return game.AbilityContent{}, false
	}
	sourceRef := game.TargetPermanentReference(0)
	dynamic, ok := lowerDynamicAmount(effect.Amount, sourceRef)
	if !ok {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	recipient, ok := groupDamageRecipientForExcluding(effect.Selector, sourceRef)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.Damage{
				Amount:       game.Dynamic(dynamic),
				Recipient:    recipient,
				DamageSource: opt.Val(sourceRef),
			},
		}},
	}.Ability(), true
}

// lowerEventPowerGroupDamageSpell lowers the triggered-ability payoff shape in
// which a permanent deals damage equal to a referenced object's power or
// toughness to a group of players ("Whenever another creature you control
// enters, it deals damage equal to its power to each opponent."). The amount
// reads the referenced object (the entering creature) once and the resolved
// value is dealt to every member of the recipient group, so it reuses
// lowerDynamicAmount with the amount's own referent. The damage source is the
// clause subject ("it" for the entering creature, "this creature" for the
// source permanent). It fails closed (ok=false) for every other shape, leaving
// the fixed/X group-damage and single-target source-power paths unchanged.
func lowerEventPowerGroupDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		(effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower &&
			effect.Amount.DynamicKind != compiler.DynamicAmountSourceToughness) ||
		effect.DamageRecipient.Reference != parser.DamageRecipientReferenceNone ||
		len(effect.DamageRiders) != 0 ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	if !exactDamageSourceSyntax(ctx.content.References[:1]) ||
		!exactDamageAmountReferences(effect.Amount, ctx.content.References) {
		return game.AbilityContent{}, false
	}
	amountObject, ok := lowerDamageAmountObject(effect.Amount, ctx.content.References)
	if !ok {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, amountObject)
	if !ok {
		return game.AbilityContent{}, false
	}
	recipient, ok := groupDamageRecipientFor(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	damageSourceRef := primaryDamageSource(ctx.content.References)
	if !damageSourceRef.Exists {
		damageSourceRef = opt.Val(game.SourcePermanentReference())
	}
	damage := game.Damage{
		Amount:       game.Dynamic(dynamic),
		Recipient:    recipient,
		DamageSource: damageSourceRef,
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), true
}

// lowerEachOfTargetsDamageSpell lowers "deals N damage to each of <cardinality>
// <targets>" effects, which deal the full amount to each of the chosen targets
// (unlike divided damage, which splits one total). It emits one Damage
// instruction per target slot, each addressing its own flat target index, the
// same per-slot pattern the multi-target pump path uses. Declined "up to N"
// slots leave fewer chosen targets and the runtime Damage no-ops on an
// unresolved target index, so only the chosen targets take damage. The amount is
// either an exact fixed value or the spell's bare variable X ("deals X damage to
// each of up to three targets", Jaya's Immolating Inferno, Fall of the Titans),
// dealt in full to each chosen target. The recipient may be an "any target" slot
// (permanent or player) or a creature target. It fails closed (ok=false) for
// every other dynamic amount form, divided damage, riders, or any other selector
// so the single-target path and its diagnostic stay unchanged.
func lowerEachOfTargetsDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	if len(ctx.content.Targets) != 1 {
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
		ctx.content.Targets[0].Cardinality.Max < 2 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 {
		return game.AbilityContent{}, false
	}
	spec, ok := eachOfDamageTargetSpec(ctx.content.Targets[0])
	if !ok || !exactDamageSourceSyntax(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	damageSourceRef := primaryDamageSource(ctx.content.References)
	sequence := make([]game.Instruction, 0, spec.MaxTargets)
	for i := range spec.MaxTargets {
		sequence = append(sequence, game.Instruction{Primitive: game.Damage{
			Amount:       amount,
			Recipient:    game.AnyTargetDamageRecipient(i),
			DamageSource: damageSourceRef,
		}})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{spec},
		Sequence: sequence,
	}.Ability(), true
}

// eachOfDamageAmount resolves the per-target amount an "each of N targets" damage
// effect deals to every chosen target. It supports an exact fixed value of at
// least one and the spell's bare variable X, returning each as a runtime
// Quantity. It fails closed for every dynamic, modified, or non-positive amount
// the each-of path cannot represent, so those wordings keep their diagnostics.
func eachOfDamageAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	if amount.DynamicKind != compiler.DynamicAmountNone ||
		amount.DynamicForm != compiler.DynamicAmountFormNone ||
		amount.Addend != 0 || amount.Multiplier != 0 {
		return game.Quantity{}, false
	}
	switch {
	case amount.Known:
		if amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(amount.Value), true
	case amount.VariableX:
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	default:
		return game.Quantity{}, false
	}
}

// eachOfDamageTargetSpec builds the multi-target spec an "each of N targets"
// damage effect chooses among, carrying the wording's own cardinality range so
// the plural ("two target creatures") and optional ("up to two target
// creatures") forms both lower. It supports the "any target" slot (permanent or
// player) and the creature target the parser marks exact, failing closed for
// every other selector.
func eachOfDamageTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !target.Exact || target.Cardinality.Max < 2 ||
		target.Cardinality.Min < 0 || target.Cardinality.Min > target.Cardinality.Max {
		return game.TargetSpec{}, false
	}
	switch target.Selector.Kind {
	case compiler.SelectorAny:
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
			MinTargets: target.Cardinality.Min,
			MaxTargets: target.Cardinality.Max,
			Constraint: target.Text,
			Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
		}, true
	case compiler.SelectorCreature:
		spec, ok := permanentTargetSpecWithCardinality(target)
		if !ok {
			return game.TargetSpec{}, false
		}
		spec.Constraint = target.Text
		return spec, true
	default:
		return game.TargetSpec{}, false
	}
}

// damageSourceIsSourcePermanent reports whether the damage subject is the source
// permanent itself, referenced as "this <object>" (ReferenceThisObject) or "it"
// (ReferencePronoun) bound to ReferenceBindingSource. Such damage must carry an
// explicit game.SourcePermanentReference() so the runtime attributes the source
// permanent's keywords (lifelink, deathtouch) via last-known information. The
// card-name form (ReferenceSelfName) is excluded because an instant/sorcery
// spell's source is the spell, not a permanent; its empty default is left
// unchanged.
func damageSourceIsSourcePermanent(references []compiler.CompiledReference) bool {
	if len(references) == 0 || references[0].Binding != compiler.ReferenceBindingSource {
		return false
	}
	switch references[0].Kind {
	case compiler.ReferenceThisObject:
		return true
	case compiler.ReferencePronoun:
		return references[0].Pronoun == compiler.ReferencePronounIt
	default:
		return false
	}
}

// primaryDamageSource resolves the optional DamageSource attribution shared by
// every primary deal-damage instruction. It is the lowering-side realization of
// the unified damage output's Source component (#1748): the same source-binding
// branch was previously rebuilt in each primary-damage lowering. When the damage
// subject is the triggering event's permanent ("it" bound to the event) it
// returns that event reference; when the subject is the ability's own source
// permanent ("this creature", or "it" bound to the source) it returns
// game.SourcePermanentReference() so the runtime attributes the source's
// last-known keywords (lifelink, deathtouch); otherwise it returns the zero opt
// and the runtime attributes the damage to the resolving source. Only the damage
// subject (references[0]) is inspected.
func primaryDamageSource(references []compiler.CompiledReference) opt.V[game.ObjectReference] {
	if len(references) > 0 {
		if source, bound := lowerDamageSourceReference(references[:1]); bound &&
			source.Kind() == game.ObjectReferenceEventPermanent {
			return opt.Val(source)
		}
	}
	if damageSourceIsSourcePermanent(references) {
		return opt.Val(game.SourcePermanentReference())
	}
	return opt.V[game.ObjectReference]{}
}

func exactDamageSourceSyntax(references []compiler.CompiledReference) bool {
	if len(references) == 0 {
		return false
	}
	reference := references[0]
	if reference.Kind == compiler.ReferencePronoun && reference.Pronoun == compiler.ReferencePronounIt {
		return reference.Binding == compiler.ReferenceBindingEventPermanent ||
			reference.Binding == compiler.ReferenceBindingSource
	}
	if reference.Kind == compiler.ReferenceThisObject {
		return reference.Binding == compiler.ReferenceBindingSource
	}
	return reference.Kind == compiler.ReferenceSelfName
}

// lowerSourcePowerModifyPTSpell lowers an exact until-end-of-turn power/toughness
// pump whose variable amount reads a permanent's power ("… gets +X/+X until end
// of turn, where X is its power."). The power referent ("its", "this creature's",
// or the card's own name) lowers to the permanent whose power supplies X, which
// the runtime snapshots when the spell or ability resolves (the pump appends a
// fixed-delta continuous effect, so reading the pumped object's own power does
// not feed back on itself).
//
// It handles three subject shapes:
//   - a single creature target ("Target creature gets +X/+X … where X is its
//     power." or "… where X is <this creature>'s power."), pumping the target
//     slot;
//   - the source permanent itself ("<Name>/This creature gets +X/+X … where X is
//     its power.", EffectContextSource), pumping the source; and
//   - the triggering permanent or a prior clause's target referenced by "it"
//     (EffectContextReferencedObject), pumping that permanent.
//
// Every other shape — riders, keyword grants, conditions, modes, plural or
// non-creature targets, or a reference set that is not exactly the power referent
// plus the single subject — returns ok=false so the caller falls through to the
// fail-closed diagnostic.
func lowerSourcePowerModifyPTSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!dynamicModifyPTFormValid(&effect) {
		return game.AbilityContent{}, false
	}
	powerReference, subjects, ok := sourcePowerReferences(&effect)
	if !ok {
		return game.AbilityContent{}, false
	}
	powerObject, ok := lowerObjectReference(powerReference, referenceLoweringContext{
		AllowSource: true,
		AllowTarget: true,
		AllowEvent:  true,
	})
	if !ok {
		return game.AbilityContent{}, false
	}
	pumped, targets, ok := sourcePowerPumpTarget(ctx, &effect, subjects)
	if !ok {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, powerObject)
	if !ok {
		return game.AbilityContent{}, false
	}
	var powerDelta, toughnessDelta game.Quantity
	switch effect.Amount.DynamicForm {
	case compiler.DynamicAmountWhereX:
		powerDelta = whereXSignedQuantity(&dynamic, effect.PowerDelta)
		toughnessDelta = whereXSignedQuantity(&dynamic, effect.ToughnessDelta)
	case compiler.DynamicAmountForEach:
		powerDelta = dynamicSignedQuantity(&dynamic, effect.PowerDelta)
		toughnessDelta = dynamicSignedQuantity(&dynamic, effect.ToughnessDelta)
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         pumped,
				PowerDelta:     powerDelta,
				ToughnessDelta: toughnessDelta,
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}

// lowerSharedCreatureTypePumpSpell lowers an exact until-end-of-turn pump whose
// "for each" amount counts the other creatures sharing a creature type with the
// affected permanent ("it gets +1/+0 until end of turn for each other attacking
// creature that shares a creature type with it.", Shared Animosity). The "with
// it" referent and the pumped subject both name the affected permanent, so the
// referent reference is split off by the amount's span and the pump addresses
// the remaining subject reference (the triggering attacker or the source) or a
// creature target slot. The shared-creature-type count is evaluated relative to
// the pumped permanent as the ability resolves, so the dynamic amount's count
// object is unused. It fails closed for any other shape so unsupported wordings
// stay rejected.
func lowerSharedCreatureTypePumpSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if effect.Amount.DynamicKind != compiler.DynamicAmountSharedCreatureTypeCount ||
		effect.Amount.DynamicForm != compiler.DynamicAmountForEach ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!dynamicModifyPTFormValid(&effect) {
		return game.AbilityContent{}, false
	}
	referent, subjects, ok := sourcePowerReferences(&effect)
	if !ok {
		return game.AbilityContent{}, false
	}
	switch referent.Binding {
	case compiler.ReferenceBindingEventPermanent,
		compiler.ReferenceBindingSource,
		compiler.ReferenceBindingTarget:
	default:
		return game.AbilityContent{}, false
	}
	pumped, targets, ok := sourcePowerPumpTarget(ctx, &effect, subjects)
	if !ok {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, pumped)
	if !ok {
		return game.AbilityContent{}, false
	}
	powerDelta := dynamicSignedQuantity(&dynamic, effect.PowerDelta)
	toughnessDelta := dynamicSignedQuantity(&dynamic, effect.ToughnessDelta)
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         pumped,
				PowerDelta:     powerDelta,
				ToughnessDelta: toughnessDelta,
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}

// and the target spec it declares, given the effect's non-power subject
// references. A single creature target with no subject reference pumps the target
// slot; no target with a single source, triggering-permanent, or prior-target
// subject reference pumps that permanent. The subject reference's binding must
// agree with the effect context so a mismatched reference set fails closed.
func sourcePowerPumpTarget(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
	subjects []compiler.CompiledReference,
) (game.ObjectReference, []game.TargetSpec, bool) {
	switch {
	case len(ctx.content.Targets) == 1 && len(subjects) == 0 &&
		effect.Context == parser.EffectContextTarget:
		target := ctx.content.Targets[0]
		if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
			target.Selector.Kind != compiler.SelectorCreature {
			return game.ObjectReference{}, nil, false
		}
		spec, ok := permanentTargetSpec(target)
		if !ok {
			return game.ObjectReference{}, nil, false
		}
		return game.TargetPermanentReference(0), []game.TargetSpec{spec}, true
	case len(ctx.content.Targets) == 0 && len(subjects) == 1:
		if !sourcePowerSubjectContextValid(subjects[0].Binding, effect.Context) {
			return game.ObjectReference{}, nil, false
		}
		object, ok := lowerObjectReference(subjects[0], referenceLoweringContext{
			AllowSource: true,
			AllowTarget: true,
			AllowEvent:  true,
		})
		if !ok {
			return game.ObjectReference{}, nil, false
		}
		return object, nil, true
	default:
		return game.ObjectReference{}, nil, false
	}
}

// sourcePowerSubjectContextValid pairs a subject reference's binding with the
// effect context the parser assigns its wording: the source permanent itself is
// EffectContextSource, while the triggering permanent or a prior clause's target
// addressed by "it" is EffectContextReferencedObject. Any other pairing fails
// closed.
func sourcePowerSubjectContextValid(
	binding compiler.ReferenceBinding,
	context parser.EffectContextKind,
) bool {
	switch binding {
	case compiler.ReferenceBindingSource:
		return context == parser.EffectContextSource
	case compiler.ReferenceBindingEventPermanent, compiler.ReferenceBindingTarget:
		return context == parser.EffectContextReferencedObject
	default:
		return false
	}
}

func lowerFixedModifyPTSpell(
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := &ctx.content.Effects[0]
	if effect.StaticSubject != compiler.StaticSubjectNone {
		return lowerFixedGroupModifyPTSpell(ctx, effect)
	}
	if content, ok := lowerSourcePowerModifyPTSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSharedCreatureTypePumpSpell(ctx); ok {
		return content, nil
	}
	if len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent {
		return lowerEventPermanentFixedModifyPT(ctx)
	}
	if len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		(ctx.content.References[0].Binding == compiler.ReferenceBindingSource ||
			ctx.content.References[0].Binding == compiler.ReferenceBindingTarget) {
		return lowerReferencedFixedModifyPT(ctx)
	}
	dynamicPT := effect.Amount.DynamicKind != compiler.DynamicAmountNone
	if !dynamicPT {
		if content, ok := lowerFixedModifyPTTargets(ctx); ok {
			return content, nil
		}
	}
	if content, ok := lowerReferencedAmountModifyPTTargetSpell(ctx); ok {
		return content, nil
	}
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Targets[0].Selector.Kind != compiler.SelectorCreature ||
		(!dynamicPT && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known)) ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!validModifyPTAmount(effect, len(ctx.content.References)) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	powerDelta := game.Fixed(compiledSignedAmountValue(effect.PowerDelta))
	toughnessDelta := game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta))
	if dynamicPT {
		var dynamic game.DynamicAmount
		var ok bool
		if triggeringEventQuantityKind(effect.Amount.DynamicKind) {
			dynamic, ok = lowerTriggeringEventQuantityAmount(ctx, effect.Amount)
		} else {
			dynamic, ok = lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		}
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
		switch effect.Amount.DynamicForm {
		case compiler.DynamicAmountWhereX:
			powerDelta = whereXSignedQuantity(&dynamic, effect.PowerDelta)
			toughnessDelta = whereXSignedQuantity(&dynamic, effect.ToughnessDelta)
		case compiler.DynamicAmountForEach:
			powerDelta = dynamicSignedQuantity(&dynamic, effect.PowerDelta)
			toughnessDelta = dynamicSignedQuantity(&dynamic, effect.ToughnessDelta)
		default:
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.ModifyPT{
					Object:         game.TargetPermanentReference(0),
					PowerDelta:     powerDelta,
					ToughnessDelta: toughnessDelta,
					Duration:       game.DurationUntilEndOfTurn,
				},
			},
		},
	}.Ability(), nil
}

// lowerFixedModifyPTTargets lowers an exact until-end-of-turn power/toughness
// pump whose single target slot may be single ("Target creature gets +1/+1
// until end of turn."), plural ("Two target creatures each get -1/-1 until end
// of turn."), or optional ("Up to one target creature gets -2/-2 until end of
// turn."), and whose selector may name a creature or a creature subtype ("Target
// Goblin you control gets +1/+0 until end of turn."). Each power/toughness side
// is either a fixed signed amount or the spell's variable "X" ("Target creature
// gets +X/+0 until end of turn." with {X} in the spell or activation cost), the
// latter lowering to the runtime's X amount. It emits one ModifyPT per target
// slot, each addressing its own slot, mirroring the multi-target permanent
// verbs. Declined "up to" slots leave fewer chosen targets and the runtime
// ModifyPT no-ops on an unresolved target index, so the spell pumps only the
// chosen targets. It returns ok=false for any shape outside this bounded set
// (rules-derived dynamic amounts, riders, "you may" optionality, or a
// non-creature selector) so callers fall back to the dynamic single-creature
// path and the fail-closed diagnostic.
func lowerFixedModifyPTTargets(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	effect := &ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!modifyPTSideResolved(effect.PowerDelta) ||
		!modifyPTSideResolved(effect.ToughnessDelta) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!validModifyPTAmount(effect, len(ctx.content.References)) {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if !pumpTargetSelector(target.Selector) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := permanentTargetSpecWithCardinality(target)
	if !ok || targetSpec.MaxTargets < 1 {
		return game.AbilityContent{}, false
	}
	powerDelta := modifyPTSideQuantity(effect.PowerDelta)
	toughnessDelta := modifyPTSideQuantity(effect.ToughnessDelta)
	sequence := make([]game.Instruction, 0, targetSpec.MaxTargets)
	for i := range targetSpec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.ModifyPT{
				Object:         game.TargetPermanentReference(i),
				PowerDelta:     powerDelta,
				ToughnessDelta: toughnessDelta,
				Duration:       game.DurationUntilEndOfTurn,
			},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

// modifyPTSideResolved reports whether one power/toughness delta side of a
// non-dynamic pump carries a value the backend can lower: a fixed signed amount
// ("+2", "-1") or the spell's variable "X" ("+X"). A side that is neither stays
// fail-closed.
func modifyPTSideResolved(side compiler.CompiledSignedAmount) bool {
	return side.Known || side.VariableX
}

// modifyPTSideQuantity lowers one power/toughness delta side of a non-dynamic
// pump. A fixed side becomes its signed integer; a variable "X" side becomes the
// runtime X amount ("+X" reads the spell or activation X, "-X" negates it).
func modifyPTSideQuantity(side compiler.CompiledSignedAmount) game.Quantity {
	if side.VariableX {
		multiplier := 1
		if side.Negative {
			multiplier = -1
		}
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: multiplier})
	}
	return game.Fixed(compiledSignedAmountValue(side))
}

// pumpTargetSelector reports whether a fixed-pump target selector names a
// creature the executable backend can pump: a bare "creature" head noun or a
// creature subtype noun ("Goblin", "Elf", "Merfolk"). It fails closed for every
// other permanent kind so pumps stay restricted to creatures and creature
// subtypes, matching the real pump-spell population. permanentTargetSpecWith-
// Cardinality already rejects an Unknown selector that carries no subtype.
func pumpTargetSelector(selector compiler.CompiledSelector) bool {
	switch selector.Kind {
	case compiler.SelectorCreature:
		return true
	case compiler.SelectorUnknown:
		return len(selector.SubtypesAny()) > 0
	default:
		return false
	}
}

// lowerEventPermanentFixedModifyPT lowers an exact until-end-of-turn ModifyPT
// body whose sole non-target subject reference is
// ReferenceBindingEventPermanent. The text is "It gets <power>/<toughness>
// until end of turn." with either a fixed amount or a dynamic "where X is the
// number of …"/"… for each …" amount counted over the ability controller's
// permanents or cards. The object lowers to game.EventPermanentReference(),
// which identifies the permanent named by the triggering event.
func lowerEventPermanentFixedModifyPT(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact until-end-of-turn power/toughness changes to the triggering permanent",
		)
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Context != parser.EffectContextReferencedObject {
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return unsupported()
	}
	powerDelta, toughnessDelta, ok := referencedModifyPTQuantities(&effect, object)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         object,
				PowerDelta:     powerDelta,
				ToughnessDelta: toughnessDelta,
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

// lowerReferencedFixedModifyPT lowers an exact until-end-of-turn ModifyPT body
// whose sole subject reference is the source permanent itself ("This creature
// gets <p>/<t> until end of turn.", EffectContextSource) or a prior clause's
// target referenced by "it" in an ordered sequence ("… It gets <p>/<t> until
// end of turn.", EffectContextReferencedObject). The amount may be fixed or a
// dynamic "where X is the number of …"/"… for each …" amount counted over the
// ability controller's permanents or cards. The object lowers to
// game.SourcePermanentReference() or a target reference accordingly.
func lowerReferencedFixedModifyPT(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact until-end-of-turn power/toughness changes to the source or referenced permanent",
		)
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.References) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn {
		return unsupported()
	}
	binding := ctx.content.References[0].Binding
	switch {
	case binding == compiler.ReferenceBindingSource && effect.Context == parser.EffectContextSource:
	case binding == compiler.ReferenceBindingTarget && effect.Context == parser.EffectContextReferencedObject:
	default:
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource: true,
		AllowTarget: true,
	})
	if !ok {
		return unsupported()
	}
	powerDelta, toughnessDelta, ok := referencedModifyPTQuantities(&effect, object)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         object,
				PowerDelta:     powerDelta,
				ToughnessDelta: toughnessDelta,
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

// lowerDoublePTSpell lowers a power/toughness doubling effect ("double target
// creature's power until end of turn", Unleash Fury; "double the power and
// toughness of each creature you control until end of turn", Unnatural Growth)
// into an until-end-of-turn continuous effect whose DoublePower/DoubleToughness
// flags add each affected creature's own current value back into itself (CR
// 107.16). The same continuous-effect building block is used for both targeted
// and group recipients. Conditions, keyword riders, modes, and non-target
// references fail closed here; richer supported combinations are handled by
// sequence composition helpers.
func lowerDoublePTSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported double power/toughness spell",
			"the executable source backend supports only doubling the power and/or toughness of supported targets or creature groups until end of turn",
		)
	}
	effect := &ctx.content.Effects[0]
	if len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Negated {
		return unsupported()
	}
	duration, ok := temporaryContinuousDuration(effect.Duration)
	if !ok {
		return unsupported()
	}
	continuous := doublePTContinuousEffect(effect)
	continuousEffects := []game.ContinuousEffect{continuous}
	return continuousSubjectMode(
		ctx,
		effect,
		continuousEffects,
		duration,
		continuousSubjectOptions{
			AllowGroup:           true,
			AllowTarget:          true,
			AllowReferenceObject: true,
		},
		unsupported,
	)
}

func doublePTContinuousEffect(effect *compiler.CompiledEffect) game.ContinuousEffect {
	if !effect.DoublePower && !effect.DoubleToughness {
		panic("doublePTContinuousEffect called without doubled power or toughness")
	}
	return game.ContinuousEffect{
		Layer:           game.LayerPowerToughnessModify,
		DoublePower:     effect.DoublePower,
		DoubleToughness: effect.DoubleToughness,
	}
}

// lowerTemporaryDoublePTKeywordSpell composes two already-supported continuous
// building blocks for the Legion Leadership shape: one targeted continuous
// effect doubles power/toughness, and another grants keyword(s), all for the
// shared until-end-of-turn duration. The keyword clause may inherit the target
// either structurally ("target creature ... and gains ...") or through the
// singular back-reference ("it gains ...").
func lowerTemporaryDoublePTKeywordSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!temporaryKeywordTarget(ctx.content.Targets[0]) {
		return game.AbilityContent{}, false
	}
	var doubleEffect, keywordEffect compiler.CompiledEffect
	switch {
	case ctx.content.Effects[0].Kind == compiler.EffectDouble &&
		ctx.content.Effects[1].Kind == compiler.EffectGain:
		doubleEffect = ctx.content.Effects[0]
		keywordEffect = ctx.content.Effects[1]
	case ctx.content.Effects[0].Kind == compiler.EffectGain &&
		ctx.content.Effects[1].Kind == compiler.EffectDouble:
		keywordEffect = ctx.content.Effects[0]
		doubleEffect = ctx.content.Effects[1]
	default:
		return game.AbilityContent{}, false
	}
	if doubleEffect.Negated ||
		keywordEffect.Negated ||
		doubleEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		doubleEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		!keywordEffect.Exact {
		return game.AbilityContent{}, false
	}
	if !doubleEffect.Exact && doubleEffect.Span != keywordEffect.Span {
		return game.AbilityContent{}, false
	}
	switch len(ctx.content.References) {
	case 0:
	case 1:
		if ctx.content.References[0].Binding != compiler.ReferenceBindingTarget ||
			keywordEffect.Context != parser.EffectContextReferencedObject {
			return game.AbilityContent{}, false
		}
	default:
		return game.AbilityContent{}, false
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return game.AbilityContent{}, false
	}
	content, diag := continuousTargetMode(
		ctx.content.Targets[0],
		[]game.ContinuousEffect{
			doublePTContinuousEffect(&doubleEffect),
			{
				Layer:       game.LayerAbility,
				AddKeywords: keywords,
			},
		},
		game.DurationUntilEndOfTurn,
		func() (game.AbilityContent, *shared.Diagnostic) {
			return game.AbilityContent{}, nil
		},
	)
	if diag != nil {
		return game.AbilityContent{}, false
	}
	return content, true
}

// lowerDoubleCountersSpell lowers a counter-doubling effect ("Double the number
// of +1/+1 counters on this creature.", Mossborn Hydra; "Double the number of
// each kind of counter on target artifact, creature, or land.", Vorel of the
// Hull Clade). The doubled permanent is the source itself (self form) or the
// effect's single permanent target. A single named kind is doubled with a
// dynamic counter placement that adds counters equal to the object's current
// count of that kind (DynamicAmountObjectCounters); the all-kinds form emits a
// single AddCounter{AllKinds} whose runtime doubles every counter kind present.
// Conditions, keywords, modes, negation, and an unsupported single counter kind
// fail closed.
func lowerDoubleCountersSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported double counters spell",
			"the executable source backend supports doubling a supported counter kind or every kind of counter on the source or one permanent target",
		)
	}
	effect := &ctx.content.Effects[0]
	if len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Negated {
		return unsupported()
	}
	object, target, ok := doubleCountersObjectReference(ctx, effect)
	if !ok {
		return unsupported()
	}
	var primitive game.Primitive
	switch {
	case effect.DoubleCountersAllKinds:
		primitive = game.AddCounter{Object: object, AllKinds: true}
	case effect.DoubleCountersGroup:
		kind := effect.DoubleSourceCounterKind
		if !kind.Valid() || kind.PlayerOnly() || !compiler.CounterKindPlacementSupported(kind) {
			return unsupported()
		}
		group, ok := groupCounterRecipient(effect.Selector)
		if !ok {
			return unsupported()
		}
		primitive = game.AddCounter{Group: group, CounterKind: kind, DoubleKind: true}
	default:
		kind := effect.DoubleSourceCounterKind
		if !kind.Valid() || kind.PlayerOnly() || !compiler.CounterKindPlacementSupported(kind) {
			return unsupported()
		}
		primitive = game.AddCounter{
			Object:      object,
			CounterKind: kind,
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:        game.DynamicAmountObjectCounters,
				Object:      object,
				CounterKind: kind,
			}),
		}
	}
	mode := game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}
	if target.Exists {
		mode.Targets = []game.TargetSpec{target.Val}
	}
	return mode.Ability(), nil
}

// doubleCountersObjectReference resolves the permanent a counter-doubling effect
// doubles. The target form ("... on target <permanent>") binds to the effect's
// single permanent target, returned as the set target spec; the self form binds
// to the source permanent, allowing only a single source-bound self reference
// ("this creature"/"it"). Any other shape fails closed.
func doubleCountersObjectReference(ctx contentCtx, effect *compiler.CompiledEffect) (game.ObjectReference, opt.V[game.TargetSpec], bool) {
	if effect.DoubleCountersTarget {
		if len(ctx.content.Targets) != 1 || len(ctx.content.References) != 0 {
			return game.ObjectReference{}, opt.V[game.TargetSpec]{}, false
		}
		spec, ok := permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.ObjectReference{}, opt.V[game.TargetSpec]{}, false
		}
		return game.TargetPermanentReference(0), opt.Val(spec), true
	}
	if len(ctx.content.Targets) != 0 {
		return game.ObjectReference{}, opt.V[game.TargetSpec]{}, false
	}
	if len(ctx.content.References) == 1 {
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowSource: true,
			AllowEvent:  !ctx.sequenceClause || ctx.allowEventPronoun,
		})
		if !ok {
			return game.ObjectReference{}, opt.V[game.TargetSpec]{}, false
		}
		return object, opt.V[game.TargetSpec]{}, true
	}
	if len(ctx.content.References) != 0 {
		return game.ObjectReference{}, opt.V[game.TargetSpec]{}, false
	}
	return game.SourcePermanentReference(), opt.V[game.TargetSpec]{}, true
}

func lowerFixedGroupModifyPTSpell(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported group power/toughness spell",
			"the executable source backend supports exact fixed group power/toughness changes and linked all-creatures -X/-X until end of turn",
		)
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn {
		return unsupported()
	}
	group, ok := resolvingStaticSubjectGroup(effect)
	if !ok {
		return unsupported()
	}
	continuous, ok := groupModifyPTContinuousEffect(effect, group)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{continuous},
				Duration:          game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

// groupModifyPTContinuousEffect builds the LayerPowerToughnessModify continuous
// effect for a group power/toughness change over group. It models three shapes:
// a fixed delta ("+1/+1"), the linked all-creatures -X/-X form (the variable
// spell "X" applied to every creature), and a dynamic battlefield-counted
// amount ("+X/+X … where X is the number of creatures you control",
// "+X/+X … where X is the greatest power among creatures you control", or a
// "for each" multiplier). It returns ok=false for any amount the dynamic
// machinery cannot render, keeping inexpressible forms fail-closed.
func groupModifyPTContinuousEffect(
	effect *compiler.CompiledEffect,
	group game.GroupReference,
) (game.ContinuousEffect, bool) {
	continuous := game.ContinuousEffect{
		Layer: game.LayerPowerToughnessModify,
		Group: group,
	}
	variableX := effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		effect.PowerDelta.VariableX &&
		effect.ToughnessDelta.VariableX
	switch {
	case variableX:
		if effect.StaticSubject != compiler.StaticSubjectAllCreatures ||
			!effect.PowerDelta.Negative ||
			!effect.ToughnessDelta.Negative {
			return game.ContinuousEffect{}, false
		}
		dynamic := game.DynamicAmount{Kind: game.DynamicAmountX, Multiplier: -1}
		continuous.PowerDeltaDynamic = opt.Val(dynamic)
		continuous.ToughnessDeltaDynamic = opt.Val(dynamic)
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		power, toughness, ok := referencedModifyPTQuantities(effect, game.SourcePermanentReference())
		if !ok {
			return game.ContinuousEffect{}, false
		}
		continuous.PowerDelta = power.Value()
		continuous.ToughnessDelta = toughness.Value()
		continuous.PowerDeltaDynamic = power.DynamicAmount()
		continuous.ToughnessDeltaDynamic = toughness.DynamicAmount()
	default:
		if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
			return game.ContinuousEffect{}, false
		}
		continuous.PowerDelta = compiledSignedAmountValue(effect.PowerDelta)
		continuous.ToughnessDelta = compiledSignedAmountValue(effect.ToughnessDelta)
	}
	return continuous, true
}

func lowerTemporaryKeywordSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported temporary keyword spell",
			"the executable source backend supports only exact non-parameterized keyword grants to one target creature or permanent until end of turn",
		)
	}
	effect := ctx.content.Effects[0]
	if effect.Context == parser.EffectContextController &&
		effect.Duration == compiler.DurationUntilYourNextTurn {
		if len(ctx.content.Effects) != 1 ||
			len(ctx.content.Targets) != 0 ||
			len(ctx.content.References) != 0 ||
			len(ctx.content.Conditions) != 0 ||
			len(ctx.content.Modes) != 0 ||
			effect.Kind != compiler.EffectGain ||
			!effect.Exact ||
			effect.Negated ||
			len(ctx.content.Keywords) != 1 {
			return unsupported()
		}
		keyword := ctx.content.Keywords[0]
		protection := keyword.Protection
		if keyword.Kind != parser.KeywordProtection ||
			keyword.ParameterKind != parser.KeywordParameterProtection ||
			!keyword.ProtectionKnown ||
			!protection.Everything ||
			len(protection.FromColors) != 0 ||
			len(protection.FromTypes) != 0 ||
			len(protection.FromSubtypes) != 0 ||
			protection.Multicolored ||
			protection.Monocolored ||
			protection.EachColor {
			return unsupported()
		}
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.ApplyRule{
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectPlayerProtection,
						AffectedPlayer: game.PlayerYou,
						Protection:     protection,
					}},
					Duration: game.DurationUntilYourNextTurn,
				},
			}},
		}.Ability(), nil
	}
	if effect.StaticSubject != compiler.StaticSubjectNone {
		return lowerGroupTemporaryKeywordSpell(ctx, unsupported)
	}
	// The subject is either a single permanent target or a single referenced
	// object. continuousReferenceObject accepts every reference binding the
	// runtime's ApplyContinuous can resolve (source, source-attached, triggering
	// (related) event permanent, prior-instruction result, and a referenced-object
	// target back-reference), so "it gains <keyword>" binds whatever the trigger
	// or earlier clause named.
	targetSubject := len(ctx.content.Targets) == 1 &&
		len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextTarget &&
		temporaryKeywordTarget(ctx.content.Targets[0])
	referenceSubject := len(ctx.content.Targets) == 0 && len(ctx.content.References) == 1
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Kind != compiler.EffectGain ||
		!effect.Exact ||
		(!targetSubject && !referenceSubject) ||
		effect.Negated ||
		effect.StaticSubject != compiler.StaticSubjectNone {
		return unsupported()
	}
	duration, ok := temporaryContinuousDuration(effect.Duration)
	if !ok {
		return unsupported()
	}
	keywords, abilities, ok := partitionTemporaryKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	if effect.KeywordGrantChoice {
		return lowerTemporaryKeywordChoiceGrant(ctx, &effect, keywords, abilities, targetSubject, duration, unsupported)
	}
	continuousEffects := []game.ContinuousEffect{{
		Layer:        game.LayerAbility,
		AddKeywords:  keywords,
		AddAbilities: abilities,
	}}
	if targetSubject {
		return continuousTargetMode(ctx.content.Targets[0], continuousEffects, duration, unsupported)
	}
	object, ok := continuousReferenceObject(ctx.content.References[0], &effect, true)
	if !ok {
		return unsupported()
	}
	return continuousObjectMode(object, continuousEffects, duration), nil
}

// temporaryContinuousDuration maps a compiled effect duration to the runtime
// EffectDuration for a one-shot continuous effect (a keyword grant or loss, a
// power/toughness double, and the rest of the continuous family). Only the two
// bounded forms the ApplyContinuous machinery expires are realized: "until end of
// turn" and "until your next turn" (proven by the keyword-grant path, e.g. "It
// gains haste until your next turn." on Kardur's Vicious Return chapter III). Any
// other compiled duration fails closed.
func temporaryContinuousDuration(duration compiler.DurationKind) (game.EffectDuration, bool) {
	switch duration {
	case compiler.DurationUntilEndOfTurn:
		return game.DurationUntilEndOfTurn, true
	case compiler.DurationUntilYourNextTurn:
		return game.DurationUntilYourNextTurn, true
	default:
		return game.DurationUntilEndOfTurn, false
	}
}

// lowerGroupTemporaryKeywordSpell lowers a resolving keyword grant to a
// never-resolving creature or permanent group until end of turn ("Creatures you
// control gain trample until end of turn.") into a single game.ApplyContinuous
// over the affected battlefield group with a keyword layer. It fails closed for
// any group the executable backend cannot resolve (such as a color-filtered
// group), parameterized or unimplemented keywords, and any richer shape.
func lowerGroupTemporaryKeywordSpell(
	ctx contentCtx,
	unsupported func() (game.AbilityContent, *shared.Diagnostic),
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	_, _, hasCounterQualifier := effect.StaticSubjectCounter()
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		!counterQualifierReferencesOnly(ctx.content.References, hasCounterQualifier) ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Kind != compiler.EffectGain ||
		!effect.Exact ||
		effect.Negated ||
		effect.KeywordGrantChoice ||
		effect.Duration != compiler.DurationUntilEndOfTurn {
		return unsupported()
	}
	keywords, abilities, ok := partitionTemporaryKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	group, ok := resolvingStaticSubjectGroup(&effect)
	if !ok {
		return unsupported()
	}
	return continuousGroupMode(group, []game.ContinuousEffect{{
		Layer:        game.LayerAbility,
		AddKeywords:  keywords,
		AddAbilities: abilities,
	}}, game.DurationUntilEndOfTurn), nil
}

// lowerTemporaryKeywordLossSpell lowers a resolving keyword removal until end of
// turn ("Permanents your opponents control lose hexproof and indestructible
// until end of turn.", "Target creature loses flying until end of turn.") into a
// single game.ApplyContinuous whose ability layer removes the named keywords from
// the affected subject. It mirrors lowerTemporaryKeywordSpell's grant path but
// emits RemoveKeywords. The subject may be a never-resolving creature or
// permanent group, a single targeted permanent, a referenced object, or the
// source permanent. It fails closed for parameterized or quoted abilities (only
// simple keywords reduce to RemoveKeywords) and any richer shape.
func lowerTemporaryKeywordLossSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported keyword or ability loss",
			"the executable source backend supports only exact non-parameterized keyword removal from one target permanent or a controlled/opponent group until end of turn",
		)
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Kind != compiler.EffectLose ||
		!effect.Exact ||
		effect.Negated {
		return unsupported()
	}
	duration, ok := temporaryContinuousDuration(effect.Duration)
	if !ok {
		return unsupported()
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	continuous := game.ContinuousEffect{
		Layer:          game.LayerAbility,
		RemoveKeywords: keywords,
	}
	continuousEffects := []game.ContinuousEffect{continuous}
	return continuousSubjectMode(
		ctx,
		&effect,
		continuousEffects,
		duration,
		continuousSubjectOptions{
			AllowGroup:           true,
			AllowTarget:          true,
			AllowReferenceObject: true,
			SourceAsCard:         true,
		},
		unsupported,
	)
}

// lowerTemporaryPTKeywordSpell lowers the single-subject combined buff
// "<target creature(s)> get(s) +N/+N and gain <keyword(s)> until end of turn."
// into one game.ApplyContinuous per target slot, each carrying both a
// power/toughness layer and a keyword layer. The parser splits the body into a
// target EffectModifyPT and a prior-subject EffectGain sharing one span; both
// must be until-end-of-turn. The pump delta is a fixed +N/+N, the spell's "X",
// or a "for each <permanent>"/"where X is …" dynamic amount; the dynamic
// for-each clause loses its own "until end of turn" terminator to the shared
// gain clause and so reports inexact, which the identical-span coverage
// tolerates. The target slot may be single ("Target creature gets +1/+1 and
// gains trample…") or multi-cardinality ("Up to two target creatures each get
// +1/+1 and gain lifelink…"); a declined "up to" slot leaves an unresolved
// target index that the runtime ApplyContinuous no-ops, so only chosen
// creatures are buffed. It fails closed for any richer shape.
func lowerTemporaryPTKeywordSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectModifyPT ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!temporaryKeywordTarget(ctx.content.Targets[0]) {
		return game.AbilityContent{}, false
	}
	modifyEffect := ctx.content.Effects[0]
	keywordEffect := ctx.content.Effects[1]
	if modifyEffect.Span != keywordEffect.Span ||
		!temporaryPTKeywordPumpExact(modifyEffect) ||
		!keywordEffect.Exact ||
		modifyEffect.Negated ||
		keywordEffect.Negated ||
		modifyEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		modifyEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn {
		return game.AbilityContent{}, false
	}
	powerDelta, toughnessDelta, ok := temporaryPTKeywordDeltas(modifyEffect, keywordEffect)
	if !ok {
		return game.AbilityContent{}, false
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpecWithCardinality(ctx.content.Targets[0])
	if !ok || target.MaxTargets < 1 {
		return game.AbilityContent{}, false
	}
	sequence := make([]game.Instruction, 0, target.MaxTargets)
	for i := range target.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(i)),
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer:                 game.LayerPowerToughnessModify,
						PowerDelta:            powerDelta.Value(),
						ToughnessDelta:        toughnessDelta.Value(),
						PowerDeltaDynamic:     powerDelta.DynamicAmount(),
						ToughnessDeltaDynamic: toughnessDelta.DynamicAmount(),
					},
					{
						Layer:       game.LayerAbility,
						AddKeywords: keywords,
					},
				},
				Duration: game.DurationUntilEndOfTurn,
			},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{target},
		Sequence: sequence,
	}.Ability(), true
}

// temporaryPTKeywordDeltas computes the power and toughness deltas for the
// combined "<target> gets <deltas> and gains <keyword> until end of turn" pump.
// A fixed pump ("+1/+1") yields fixed deltas; a self-counted dynamic pump
// ("+X/+X … for each …") resolves through referencedModifyPTQuantities. The
// "+X/+X … where X is the number of basic land types …" domain form (The
// Weatherseed Treaty chapter III) is special: the trailing "where X is …"
// clause attaches to the gain effect, not the pump effect, so X is resolved
// from keywordEffect's dynamic amount and applied to the pump's variable sides.
// It returns ok=false for any shape the dynamic machinery cannot render.
func temporaryPTKeywordDeltas(modifyEffect, keywordEffect compiler.CompiledEffect) (power, toughness game.Quantity, ok bool) {
	if modifyEffect.Amount.DynamicKind != compiler.DynamicAmountNone {
		return referencedModifyPTQuantities(&modifyEffect, game.SourcePermanentReference())
	}
	if modifyEffect.PowerDelta.Known && modifyEffect.ToughnessDelta.Known {
		return game.Fixed(compiledSignedAmountValue(modifyEffect.PowerDelta)),
			game.Fixed(compiledSignedAmountValue(modifyEffect.ToughnessDelta)), true
	}
	if !modifyEffect.PowerDelta.VariableX && !modifyEffect.ToughnessDelta.VariableX {
		return game.Quantity{}, game.Quantity{}, false
	}
	if keywordEffect.Amount.DynamicForm != compiler.DynamicAmountWhereX ||
		keywordEffect.Amount.DynamicKind == compiler.DynamicAmountNone ||
		keywordEffect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return game.Quantity{}, game.Quantity{}, false
	}
	dynamic, ok := lowerDynamicAmount(keywordEffect.Amount, game.SourcePermanentReference())
	if !ok {
		return game.Quantity{}, game.Quantity{}, false
	}
	return whereXSignedQuantity(&dynamic, modifyEffect.PowerDelta),
		whereXSignedQuantity(&dynamic, modifyEffect.ToughnessDelta), true
}

// temporaryPTKeywordPumpExact reports whether the pump clause of a combined
// "<target> gets <deltas> and gains <keyword> until end of turn" buff covers its
// source exactly. A fixed or "where X is …" pump is recognized exact directly. A
// "for each <permanent>" pump reports inexact only because the shared trailing
// "until end of turn" terminator binds to the gain clause; the caller's
// identical-span check already proves the combined clause is fully consumed, so
// that dynamic for-each form counts as exact here too.
func temporaryPTKeywordPumpExact(modifyEffect compiler.CompiledEffect) bool {
	return modifyEffect.Exact ||
		modifyEffect.Amount.DynamicForm == compiler.DynamicAmountForEach
}

// temporaryKeywordTarget reports whether a permanent target is one the temporary
// keyword grant, loss, and combined +N/+N-and-gain lowerings can act on. It
// defers entirely to the canonical permanentTargetSpecWithCardinality, so any
// target that destroy, exile, or tap already accept — a bare subtype ("target
// Human"), a card type ("target artifact"), a color ("target black creature"),
// or a tapped/attacking qualifier — is accepted here too, including the optional
// and multi-target cardinalities those specs carry.
func temporaryKeywordTarget(target compiler.CompiledTarget) bool {
	spec, ok := permanentTargetSpecWithCardinality(target)
	return ok && spec.MaxTargets >= 1
}

// lowerGroupTemporaryPTKeywordSpell lowers the Overrun-style group buff
// "<group> get +N/+N and gain <keyword(s)> until end of turn." into a single
// game.ApplyContinuous over a battlefield group with both a power/toughness layer
// and a keyword layer. The parser splits the body into a group EffectModifyPT and
// a prior-subject EffectGain; both must be exact and until-end-of-turn with fixed
// deltas. It fails closed for any richer shape.
func lowerGroupTemporaryPTKeywordSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	// The two clauses appear in either order: "creatures you control get +N/+N
	// and gain trample …" (modify first) or "creatures you control gain trample
	// and get +N/+N …" (keyword first). The subject-bearing clause carries the
	// static group subject; the other inherits it as the prior subject.
	var modifyEffect, keywordEffect compiler.CompiledEffect
	switch {
	case ctx.content.Effects[0].Kind == compiler.EffectModifyPT &&
		ctx.content.Effects[1].Kind == compiler.EffectGain:
		modifyEffect = ctx.content.Effects[0]
		keywordEffect = ctx.content.Effects[1]
	case ctx.content.Effects[0].Kind == compiler.EffectGain &&
		ctx.content.Effects[1].Kind == compiler.EffectModifyPT:
		keywordEffect = ctx.content.Effects[0]
		modifyEffect = ctx.content.Effects[1]
	default:
		return game.AbilityContent{}, false
	}
	if !modifyEffect.Exact ||
		!keywordEffect.Exact ||
		modifyEffect.Negated ||
		keywordEffect.Negated ||
		modifyEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn {
		return game.AbilityContent{}, false
	}
	// Exactly one clause names the affected group; the other inherits it.
	subjectEffect := &modifyEffect
	if modifyEffect.StaticSubject == compiler.StaticSubjectNone {
		subjectEffect = &keywordEffect
	}
	if subjectEffect.StaticSubject == compiler.StaticSubjectNone ||
		(modifyEffect.StaticSubject != compiler.StaticSubjectNone &&
			keywordEffect.StaticSubject != compiler.StaticSubjectNone) {
		return game.AbilityContent{}, false
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return game.AbilityContent{}, false
	}
	group, ok := resolvingStaticSubjectGroup(subjectEffect)
	if !ok {
		return game.AbilityContent{}, false
	}
	modifyContinuous, ok := groupModifyPTContinuousEffect(&modifyEffect, group)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{
					modifyContinuous,
					{
						Layer:       game.LayerAbility,
						Group:       group,
						AddKeywords: keywords,
					},
				},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}

// selfSourceBounceReferences reports whether the references denote the source
// permanent returning itself, named either as "this <object>"
// (ReferenceThisObject, "Return this creature to its owner's hand.") or by the
// card's own name (ReferenceSelfName, "Return Selenia to its owner's hand."):
// the first reference is that source object and every reference binds to the
// source.
func selfSourceBounceReferences(references []compiler.CompiledReference) bool {
	if len(references) == 0 {
		return false
	}
	switch references[0].Kind {
	case compiler.ReferenceThisObject, compiler.ReferenceSelfName:
	default:
		return false
	}
	for i := range references {
		if references[i].Binding != compiler.ReferenceBindingSource {
			return false
		}
	}
	return true
}

// plainControllerBounceToHand reports whether the content is the shared
// precondition every battlefield bounce-to-hand scope requires before it
// inspects its own subject: a non-negated, non-optional, controller-context
// EffectReturn whose typed destination is the hand, carrying no attached
// conditions, keywords, or modes. The repeated destination (ToZone), context,
// and flag checks that each return scope used to spell out inline live here once,
// so a scope only adds its subject-specific checks (target cardinality, group
// selection, self-reference, or stack-target union) on top. Exactness is left to
// the caller because the spell-or-permanent target fold is recorded inexact while
// every other bounce scope requires an exact effect.
func plainControllerBounceToHand(ctx contentCtx) bool {
	effect := ctx.content.Effects[0]
	return !effect.Negated && !effect.Optional && !ctx.optional &&
		effect.Context == parser.EffectContextController &&
		effect.ToZone == zone.Hand &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0
}

func lowerFixedBounceSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) == 0 &&
		effect.Exact &&
		plainControllerBounceToHand(ctx) &&
		selfSourceBounceReferences(ctx.content.References) {
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
		if ok {
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Bounce{Object: object},
				}},
			}.Ability(), nil
		}
	}
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		!ctx.content.Effects[0].Exact ||
		!plainControllerBounceToHand(ctx) ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	target := ctx.content.Targets[0]
	targetSpec, ok := permanentTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Bounce{
					Object: object,
				},
			},
		},
	}.Ability(), nil
}

// lowerSpellBounce lowers "Return target spell to its owner's hand." and the
// compound "Return target spell or <permanent filter> to its owner's hand."
// (Sink into Stupor, Venser, Unsubstantiate, Press the Enemy). The parser folds
// the compound into a spell selector marked inexact, carrying the permanent
// side's card-type filter (e.g. "creature", "nonland permanent") and shared
// controller. The spell-only form is an exact spell selector with no folded
// permanent filter. It emits a combined stack-object/permanent target so the
// chosen target may be either the spell on the stack or a matching permanent,
// and one Bounce per slot returns whichever was chosen to its owner's hand.
func lowerSpellBounce(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectReturn ||
		!plainControllerBounceToHand(ctx) {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if target.Selector.Kind != compiler.SelectorSpell ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	spec, ok := spellBounceTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	instructions := make([]game.Instruction, 0, spec.MaxTargets)
	for i := 0; i < spec.MaxTargets; i++ {
		instructions = append(instructions, game.Instruction{
			Primitive: game.Bounce{Object: game.TargetObjectReference(i)},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{spec},
		Sequence: instructions,
	}.Ability(), true
}

// spellBounceTargetSpec builds the combined target spec for a spell bounce. An
// exact spell selector ("target spell") accepts only stack spells; an inexact
// one is the "spell or <permanent>" fold whose RequiredTypesAny/ExcludedTypes
// describe the permanent alternative, so it additionally accepts permanents
// matching that type filter. Only the controller filter and a permanent-side
// card-type filter are expressible; any other selector qualifier fails closed.
func spellBounceTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	selector := target.Selector
	if target.Cardinality.Max != 1 || target.Cardinality.Min < 0 || target.Cardinality.Min > 1 {
		return game.TargetSpec{}, false
	}
	if selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped ||
		selector.Colorless || selector.Multicolored ||
		selector.BasicLandType || selectorHasCounterQualifier(selector) ||
		selectorHasAttachmentQualifier(selector) ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.ExcludedKeyword != parser.KeywordUnknown ||
		selector.Zone != zone.None ||
		len(selector.Alternatives) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.SubtypesAny()) != 0 ||
		len(selector.ExcludedSubtypes()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedColors()) != 0 {
		return game.TargetSpec{}, false
	}
	required := selector.RequiredTypesAny()
	excluded := selector.ExcludedTypes()
	predicate := game.TargetPredicate{
		StackObjectKinds: []game.StackObjectKind{game.StackSpell},
	}
	allow := game.TargetAllowStackObject
	var selection game.Selection
	hasPermanentSide := false
	if target.Exact {
		if len(required) != 0 || len(excluded) != 0 {
			return game.TargetSpec{}, false
		}
	} else {
		allow |= game.TargetAllowPermanent
		hasPermanentSide = true
		if len(required) > 0 {
			selection.RequiredTypesAny = append([]types.Card(nil), required...)
		}
		if len(excluded) > 0 {
			selection.ExcludedTypes = append([]types.Card(nil), excluded...)
		}
	}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		predicate.Controller = game.ControllerYou
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		predicate.Controller = game.ControllerOpponent
		selection.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		predicate.Controller = game.ControllerNotYou
		selection.Controller = game.ControllerNotYou
	default:
		// ControllerKind is a closed const-iota enum whose only values
		// (ControllerAny, ControllerYou, ControllerOpponent, ControllerNotYou)
		// are all handled above; an unhandled value is an internal compiler bug,
		// not an unsupported card.
		panic(fmt.Sprintf("spellBounceTargetSpec: unhandled ControllerKind %v", selector.Controller))
	}
	spec := game.TargetSpec{
		MinTargets: target.Cardinality.Min,
		MaxTargets: target.Cardinality.Max,
		Allow:      allow,
		Predicate:  predicate,
		Constraint: lowerFirst(target.Text),
	}
	if hasPermanentSide && !selection.Empty() {
		spec.Selection = opt.Val(selection)
	}
	return spec, true
}

// target has a plural ("Return two target creatures to their owners' hands."),
// optional-plural ("Return up to two target creatures to their owners' hands."),
// or optional-singular ("Return up to one target creature to its owner's hand.")
// cardinality. It emits one multi-target spec carrying the chosen
// MinTargets/MaxTargets range and one Bounce instruction per slot, each
// addressing its target index. The possessive destination clause ("their
// owners'" or, for the optional-singular form, "its owner's") names where the
// permanents go, not the bounced object, so each slot bounces its own target
// permanent. Declined "up to" slots leave fewer chosen targets and the runtime
// Bounce primitive no-ops on an unresolved target index, so the spell returns
// only the chosen targets. It returns ok=false for the fixed single-target
// "Return target <permanent> to its owner's hand." form (cardinality exactly
// one) so that path stays on lowerFixedBounceSpell.
func lowerMultiTargetBounceSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	effect := ctx.content.Effects[0]
	if targetCardinalityIsOne(target) ||
		!effect.Exact ||
		!plainControllerBounceToHand(ctx) ||
		!bounceDestinationPronounReferencesOnly(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	return multiTargetPermanentMode(target, func(object game.ObjectReference) game.Primitive {
		return game.Bounce{Object: object}
	})
}

// lowerDualTargetBounceSpell lowers the dual-target battlefield bounce "Return
// target <A> and target <B> to their owners' hands." (e.g. Aether Tradewinds,
// Peel from Reality, Churning Eddy) to a Mode carrying two single-target specs
// in Oracle order and one Bounce per slot, each addressing its own target index.
// The two targets carry independent selectors ("creature you control" and
// "creature you don't control", or unrelated types like "creature" and "land")
// that a single multi-target range cannot express, so each slot bounces its own
// target permanent. The plural possessive destination ("their owners' hands")
// names where the permanents go, not a bounced object, and the compiler records
// it as a single destination pronoun reference; that pronoun is the only
// reference the lowering tolerates. It returns ok=false for every other return
// wording (single, multi-slot, mass, controlled-choice, self) so those paths are
// untouched.
func lowerDualTargetBounceSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 2 ||
		!effect.Exact ||
		!plainControllerBounceToHand(ctx) ||
		!choiceBounceDestinationReferencesOnly(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	specs := make([]game.TargetSpec, 0, len(ctx.content.Targets))
	sequence := make([]game.Instruction, 0, len(ctx.content.Targets))
	for i := range ctx.content.Targets {
		target := ctx.content.Targets[i]
		if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
			return game.AbilityContent{}, false
		}
		spec, ok := permanentTargetSpec(target)
		if !ok {
			return game.AbilityContent{}, false
		}
		specs = append(specs, spec)
		sequence = append(sequence, game.Instruction{
			Primitive: game.Bounce{Object: game.TargetPermanentReference(i)},
		})
	}
	return game.Mode{Targets: specs, Sequence: sequence}.Ability(), true
}

// lowerSelfAndTargetBounceSpell lowers the source-and-target battlefield bounce
// "Return this creature and (another) target <permanent> to their owners'
// hands." (e.g. Wizard Mentor, Coastal Wizard, Snow Hound, Lady Sun) to a Mode
// carrying the one target's single-target spec and two Bounce instructions: the
// source first, then the chosen target. It is the self sibling of
// lowerDualTargetBounceSpell, where one of the two returned permanents is the
// ability's own source named by its card name or "this <type>" rather than a
// second target. The compound clause folds the effect inexact, so this path
// tolerates an inexact return where the dual-target path requires exactness. The
// references are the source object plus the plural possessive destination pronoun
// ("their owners' hands"); both are tolerated and any other reference fails
// closed. It returns ok=false for every other return wording so the single, dual,
// multi-slot, mass, controlled, and pure-self bounce paths are untouched.
func lowerSelfAndTargetBounceSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 1 ||
		!plainControllerBounceToHand(ctx) ||
		!selfAndTargetBounceReferences(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return game.AbilityContent{}, false
	}
	spec, ok := permanentTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	source, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{spec},
		Sequence: []game.Instruction{
			{Primitive: game.Bounce{Object: source}},
			{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
		},
	}.Ability(), true
}

// selfAndTargetBounceReferences reports whether the references are exactly the
// source object returning itself, named by the card's own name
// (ReferenceSelfName, "Return Lady Sun and ...") or "this <object>"
// (ReferenceThisObject, "Return this creature and ..."), followed only by the
// destination possessive pronouns ("their"/"its" in "their owners' hands"). The
// source object is reference zero; every other reference must be that destination
// possessive. Any other reference fails closed.
func selfAndTargetBounceReferences(references []compiler.CompiledReference) bool {
	if len(references) == 0 {
		return false
	}
	switch references[0].Kind {
	case compiler.ReferenceThisObject, compiler.ReferenceSelfName:
	default:
		return false
	}
	if references[0].Binding != compiler.ReferenceBindingSource {
		return false
	}
	for i := 1; i < len(references); i++ {
		if references[i].Kind != compiler.ReferencePronoun {
			return false
		}
		switch references[i].Pronoun {
		case compiler.ReferencePronounTheir, compiler.ReferencePronounIts:
		default:
			return false
		}
	}
	return true
}

// lowerControlledBounceSpell lowers the controlled-choice battlefield bounce
// "Return a/an/another <permanent> you control to its owner's hand." to a Bounce
// whose resolving controller chooses one permanent they control matching the
// effect's selector. It carries no target — the parser records the chosen
// permanent as the effect's Selector rather than a target — so the runtime makes
// the choice at resolution. It returns ok=false for every other return wording
// (mass "all", "each", targeted, self) so those paths are untouched, and fails
// closed unless the selector is the representable "you control" relation.
func lowerControlledBounceSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 0 ||
		!effect.Exact ||
		!plainControllerBounceToHand(ctx) ||
		effect.Selector.All ||
		controlledBounceCount(effect) >= 2 ||
		effect.Selector.Controller != compiler.ControllerYou ||
		!choiceBounceDestinationReferencesOnly(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	selection, ok := massGroupSelection(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	selection.ExcludeSource = effect.Selector.Other || effect.Selector.Another
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.Bounce{
				ControlledChoice: true,
				Amount:           game.Fixed(1),
				Group:            game.BattlefieldGroup(selection),
			},
		}},
	}.Ability(), true
}

// controlledBounceCount returns the fixed count a controlled-choice bounce
// returns, or zero when the effect carries no known count of two or more. The
// singular "Return a/an/another <permanent> you control" form compiles to a
// count of one (or an unset amount); only the multi-count "Return <N> <group>"
// form records a known value of two or more, so it routes to
// lowerControlledCountBounceSpell while the singular form stays on
// lowerControlledBounceSpell.
func controlledBounceCount(effect compiler.CompiledEffect) int {
	if !effect.Amount.Known || effect.Amount.Value < 2 {
		return 0
	}
	return effect.Amount.Value
}

// lowerControlledCountBounceSpell lowers the fixed-count controlled-choice
// battlefield bounce "Return <N> <group> you control to their owner's hand."
// (Dust Elemental, Khalni Gem) to a Bounce whose resolving controller chooses
// exactly N permanents they control matching the effect's selector. It is the
// multi-count sibling of lowerControlledBounceSpell: the only difference is the
// returned count, which it reads from the effect's compiled amount rather than
// hardcoding one. It returns ok=false for every other return wording so the
// singular controlled, targeted, and mass bounce paths are untouched.
func lowerControlledCountBounceSpell(ctx contentCtx) (game.AbilityContent, bool) {
	effect := ctx.content.Effects[0]
	count := controlledBounceCount(effect)
	if len(ctx.content.Targets) != 0 ||
		!effect.Exact ||
		!plainControllerBounceToHand(ctx) ||
		effect.Selector.All ||
		count < 2 ||
		effect.Selector.Controller != compiler.ControllerYou ||
		!choiceBounceDestinationReferencesOnly(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	selection, ok := massGroupSelection(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	selection.ExcludeSource = effect.Selector.Other || effect.Selector.Another
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.Bounce{
				ControlledChoice: true,
				Amount:           game.Fixed(count),
				Group:            game.BattlefieldGroup(selection),
			},
		}},
	}.Ability(), true
}

// choiceBounceDestinationReferencesOnly reports whether every reference is the
// possessive pronoun that names the controlled-choice bounce destination
// ("its"/"it"/"their" in "to its owner's hand"). Unlike the targeted multi-bounce
// the chosen permanent is the effect's selector, not a referenced object, so the
// destination possessive is the only reference the lowering tolerates. Its
// binding is irrelevant — the destination is always the bounced card's owner's
// hand — so a triggered "When this <permanent> enters" body, where the compiler
// binds "its" to the triggering permanent, is accepted alongside the spell body
// where it stays ambiguous. Any non-pronoun reference fails closed.
func choiceBounceDestinationReferencesOnly(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun {
			return false
		}
		switch reference.Pronoun {
		case compiler.ReferencePronounTheir,
			compiler.ReferencePronounIts,
			compiler.ReferencePronounIt:
		default:
			return false
		}
	}
	return true
}

// bounceDestinationPronounReferencesOnly reports whether every reference is the
// possessive pronoun that names the bounce destination, either the plural "their"
// ("to their owners' hands", used by a multi-target plural bounce) or the
// singular "its"/"it" ("to its owner's hand", used by the "up to one target ..."
// optional single-slot bounce). For an optional ("up to one") or plural target
// the compiler cannot bind the possessive to the permanent, so it leaves it
// ambiguous; the multi-target bounce addresses each slot by index rather than
// through the reference, so the destination possessive is the only reference the
// lowering tolerates. Any other reference fails closed.
func bounceDestinationPronounReferencesOnly(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun ||
			reference.Binding != compiler.ReferenceBindingAmbiguous {
			return false
		}
		switch reference.Pronoun {
		case compiler.ReferencePronounTheir,
			compiler.ReferencePronounIts,
			compiler.ReferencePronounIt:
		default:
			return false
		}
	}
	return true
}

func lowerFixedPermanentTargetSpell(
	ctx contentCtx,
	verb string,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerMultiTargetPermanentSpell(ctx, primitiveFactory); ok {
		return content, nil
	}
	if !matchesExactSinglePermanentTargetSpell(ctx) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(game.TargetPermanentReference(0)),
			},
		},
	}.Ability(), nil
}

// lowerBecomeCopyContent lowers an activated/resolving become-a-copy effect
// ("This land becomes a copy of target land, except it has this ability.",
// Thespian's Stage; "... until end of turn.", Mirage Mirror) into a BecomeCopy
// primitive acting on the source permanent and copying the single target. The
// source permanent is implicit, so the clause's source back-reference is ignored
// here rather than rejected by the fixed-single-target lowering's modifier guard.
func lowerBecomeCopyContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) != 1 || !targetCardinalityIsOne(ctx.content.Targets[0]) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported become-a-copy effect",
			"the executable source backend supports only a become-a-copy effect with one target permanent",
		)
	}
	target, ok := becomeCopyTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported become-a-copy effect",
			"the executable source backend supports only a become-a-copy effect with a supported target permanent",
		)
	}
	var keywords []game.Keyword
	for _, keyword := range effect.BecomeCopyAddKeywords {
		runtime, ok := runtimeKeyword(keyword)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported become-a-copy effect",
				"the executable source backend does not support the copiable keyword rider",
			)
		}
		keywords = append(keywords, runtime)
	}
	primitive := game.BecomeCopy{
		Object:             target.object,
		Card:               target.card,
		UntilEndOfTurn:     effect.BecomeCopyUntilEndOfTurn,
		RetainsThisAbility: effect.BecomeCopyRetainsThisAbility,
		AddKeywords:        keywords,
	}
	return game.Mode{
		Targets:  []game.TargetSpec{target.spec},
		Sequence: []game.Instruction{{Primitive: primitive}},
	}.Ability(), nil
}

// lowerBecomeTypeContent lowers a targeted continuous type-adding effect
// ("Target permanent becomes an artifact in addition to its other types until
// end of turn.", Liquimetal Torque, Liquimetal Coating) into an ApplyContinuous
// at LayerType that adds the parser-recognized card types to the single target
// permanent until end of turn. Only the additive until-end-of-turn form reaches
// here; any other shape (multiple targets, missing duration, riders) fails
// closed.
func lowerBecomeTypeContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported type-change effect",
			"the executable source backend supports only a target permanent gaining card types until end of turn",
		)
	}
	if !effect.BecomeTypeUntilEndOfTurn ||
		len(effect.BecomeTypeAddTypes) == 0 ||
		effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if len(ctx.content.Targets) != 1 || !targetCardinalityIsOne(ctx.content.Targets[0]) {
		return unsupported()
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	continuousEffects := []game.ContinuousEffect{{
		Layer:    game.LayerType,
		AddTypes: append([]types.Card(nil), effect.BecomeTypeAddTypes...),
	}}
	if len(effect.BecomeTypeAddColors) != 0 {
		continuousEffects = append(continuousEffects, game.ContinuousEffect{
			Layer:     game.LayerColor,
			AddColors: append([]color.Color(nil), effect.BecomeTypeAddColors...),
		})
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object:            opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: continuousEffects,
				Duration:          game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

// becomeCopyTarget pairs a become-a-copy target spec with the reference the
// BecomeCopy primitive uses to find the copied object: object for a battlefield
// permanent, card for a card in a non-battlefield zone. Exactly one is set.
type becomeCopyTarget struct {
	spec   game.TargetSpec
	object game.ObjectReference
	card   game.CardReference
}

// becomeCopyTargetSpec builds the target spec and copy reference for a
// become-a-copy effect. A battlefield permanent target (Thespian's Stage) copies
// the permanent referenced by object; a permanent card in the controller's
// graveyard (Shifting Woodland) copies the card referenced by card.
func becomeCopyTargetSpec(target compiler.CompiledTarget) (becomeCopyTarget, bool) {
	if spec, ok := cardInZoneTargetSpec(target, zone.Graveyard); ok {
		return becomeCopyTarget{spec: spec, card: game.CardReference{Kind: game.CardReferenceTarget}}, true
	}
	if spec, ok := permanentTargetSpec(target); ok {
		return becomeCopyTarget{spec: spec, object: game.TargetPermanentReference(0)}, true
	}
	return becomeCopyTarget{}, false
}

func matchesExactSinglePermanentTargetSpell(ctx contentCtx) bool {
	return hasExactSinglePermanentTarget(ctx.content) &&
		hasExactControllerEffect(ctx.content) &&
		hasNoFixedPermanentSpellModifiers(ctx)
}

func hasExactSinglePermanentTarget(content compiler.AbilityContent) bool {
	return len(content.Targets) == 1 && targetCardinalityIsOne(content.Targets[0])
}

func hasExactControllerEffect(content compiler.AbilityContent) bool {
	if len(content.Effects) != 1 {
		return false
	}
	effect := content.Effects[0]
	return !effect.Negated &&
		!effect.Optional &&
		effect.Exact &&
		effect.Context == parser.EffectContextController
}

func hasNoFixedPermanentSpellModifiers(ctx contentCtx) bool {
	return !ctx.optional &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0 &&
		referencesAreRedundantSoleTargetBackReferences(ctx.content.References)
}

// referencesAreRedundantSoleTargetBackReferences reports whether every reference
// (if any) is a demonstrative object reference bound to the spell's sole target,
// e.g. "exile that creature" or "destroy it" naming the permanent the spell
// already targets. Such a reference is redundant with the target, so the fixed
// single-target lowering can ignore it and act on the target directly. This
// enables a sequence clause that removes the prior clause's target ("Tap target
// creature. ... exile that creature."), where the back-reference is materialized
// as the clause's inherited target. Possessive pronouns (its/their) name a
// different object (the target's controller/owner) and so are rejected.
func referencesAreRedundantSoleTargetBackReferences(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Binding != compiler.ReferenceBindingTarget || reference.Occurrence != 0 {
			return false
		}
		switch reference.Kind {
		case compiler.ReferenceThatObject, compiler.ReferenceThisObject:
		case compiler.ReferencePronoun:
			if reference.Pronoun != compiler.ReferencePronounIt {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func lowerFixedCardCountPlayerSpell(
	ctx contentCtx,
	_ *parser.Ability,
	controllerVerb string,
	targetVerb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
	groupPrimitiveFactory func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	// Allow a single EventPlayer reference for "They {verb} N card(s)." bodies;
	// reject all other non-zero-reference forms.
	hasEventPlayerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPlayer
	hasReferencedControllerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingTarget &&
		effect.Context == parser.EffectContextReferencedObjectController
	// "That permanent's controller mills N." / "Its controller mills N." on a
	// triggered ability binds the recipient to the controller of the triggering
	// event permanent (Mesmeric Orb, Chronic Flooding, Riddlekeeper).
	hasEventPermanentControllerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent &&
		effect.Context == parser.EffectContextReferencedObjectController
	// "Defending player mills N." on a combat trigger binds the recipient to the
	// attacked player carried by the triggering attack/blocked event (Flint
	// Golem, Nemesis of Reason). The triggering source is implicit, so the
	// content carries no reference; the defending player is named by context.
	hasDefendingPlayerContext := len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextDefendingPlayer
	// "That player discards/mills N." after an ordered-sequence clause that
	// targeted a player or permanent inherits its subject from that antecedent
	// target: a player target denotes that player directly, a permanent target
	// denotes its controller. This mirrors the life-then "That player loses N
	// life." shape (Ozai's Cruelty, Immersturm Skullcairn, Recoil, Dinrova
	// Horror). The single "that player" reference is bound to the inherited
	// target.
	hasThatPlayerRef := len(ctx.content.References) == 1 &&
		effect.Context == parser.EffectContextReferencedPlayer &&
		hasThatPlayerTargetReference(ctx.content.References)
	// "Target player mills X cards, where X is the number of charge counters on
	// this artifact." (Grindclock, Font of Progress) counts a counter kind on the
	// ability's own source. The self-counter amount carries a lone source
	// reference; cardCountQuantityForContext resolves the count against the source
	// permanent, so the recipient still derives from the target player.
	hasSourceCounterRef := effect.Amount.DynamicKind == compiler.DynamicAmountSourceCounterCount &&
		singleSelfReference(ctx.content.References)
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.Selector.Kind != compiler.SelectorCard ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		(len(ctx.content.References) != 0 && !hasEventPlayerRef && !hasReferencedControllerRef &&
			!hasEventPermanentControllerRef && !hasThatPlayerRef && !hasSourceCounterRef) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	// The half-library amount ("mills half their library, rounded up/down") is
	// half the milling player's library, so its dynamic count names the resolved
	// recipient rather than a fixed or source-derived number. Defer building the
	// amount until the recipient playerRef is chosen below; the generic
	// card-count quantity helper cannot lower it (it is neither a triggering-event
	// nor a selector count) and would fail the clause closed.
	halfLibrary := effect.Amount.DynamicKind == compiler.DynamicAmountHalfPlayerLibrary
	var amount game.Quantity
	if !halfLibrary {
		resolved, ok := cardCountQuantityForContext(ctx, effect.Amount, allowDynamic)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		amount = resolved
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	if len(ctx.content.Targets) == 0 && len(ctx.content.References) == 0 && !halfLibrary {
		switch effect.Context {
		case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.OpponentsReference()),
				}},
			}.Ability(), nil
		case parser.EffectContextEachPlayer:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.AllPlayersReference()),
				}},
			}.Ability(), nil
		default:
			// Non-"each player/opponent" contexts fall through to the
			// single-recipient handling below.
		}
	}
	switch {
	case hasEventPlayerRef && len(ctx.content.Targets) == 0 &&
		(effect.Context == parser.EffectContextEventPlayer || effect.Context == parser.EffectContextReferencedPlayer):
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 0 &&
		!hasEventPlayerRef &&
		effect.Context == parser.EffectContextController:
	case hasReferencedControllerRef && len(ctx.content.Targets) == 1 && effect.Amount.Known:
		ref, ok := referencedControllerPlayerRef(ctx)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = ref
	case hasEventPermanentControllerRef && len(ctx.content.Targets) == 0:
		playerRef = game.ObjectControllerReference(game.EventPermanentReference())
	case hasThatPlayerRef && len(ctx.content.Targets) == 1:
		ref, ok := referencedThatPlayerRef(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = ref
	case hasDefendingPlayerContext && len(ctx.content.Targets) == 0:
		playerRef = game.DefendingPlayerReference()
	case len(ctx.content.Targets) == 1 &&
		!hasEventPlayerRef &&
		(effect.Context == parser.EffectContextTarget || effect.Context == parser.EffectContextPriorSubject):
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = game.TargetPlayerReference(0)
		targets = []game.TargetSpec{targetSpec}
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	if halfLibrary {
		// Count the resolved recipient's library and halve it as the effect
		// resolves, rounding up or down per the recognized "rounded up"/"rounded
		// down" word. The empty selection matches every card, so the count is the
		// whole library size before halving.
		recipient := playerRef
		amount = game.Dynamic(game.DynamicAmount{
			Kind:      game.DynamicAmountCountCardsInZone,
			Player:    &recipient,
			CardZone:  zone.Library,
			Selection: &game.Selection{},
			Divisor:   2,
			RoundUp:   effect.Amount.RoundUp,
		})
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(amount, playerRef),
			},
		},
	}.Ability(), nil
}

// discardSelectorImposesCardFilter reports whether the discard clause's selector
// carries a typed card filter that the plain game.Discard primitive cannot
// express ("a creature card", "a land card", "a nonland card", a color/subtype/
// keyword/mana-value filter). The bare "discard a card" selector (a generic card
// with no filter) returns false so it stays on the plain controlled-discard path.
func discardSelectorImposesCardFilter(selector compiler.CompiledSelector) bool {
	return selector.Kind != compiler.SelectorCard ||
		len(selector.ExcludedTypes()) > 0 ||
		len(selector.RequiredTypesAny()) > 0 ||
		len(selector.Supertypes()) > 0 ||
		len(selector.SubtypesAny()) > 0 ||
		len(selector.ColorsAny()) > 0 ||
		len(selector.ExcludedColors()) > 0 ||
		selector.Colorless ||
		selector.Multicolored ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.MatchManaValue
}

// lowerFilteredControllerDiscard lowers the controller's own single-card filtered
// self-discard ("Discard a creature card.", "Discard a nonland card.", and the
// optional "You may discard a creature card. If you do, <Y>." X-action) to a
// game.ChooseDiscardFromHand whose Selection carries the typed card filter. The
// runtime has the controller choose one matching card from their own hand and
// discard it; ChooseDiscardFromHand already filters its candidate pool by the
// Selection. The plain unfiltered "discard a card" stays on the existing
// game.Discard path (this returns ok=false for it). It is text-blind, reading
// only the typed effect/selector fields, and fails closed for any non-controller
// subject, any target/reference/condition/keyword/mode, a random or
// multi-card discard, or a selector cardSelectionForSelector cannot express, so
// every shape it does not fully model stays unsupported.
func lowerFilteredControllerDiscard(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDiscard ||
		effect.DiscardEntireHand ||
		effect.HandDiscard.AtRandom ||
		effect.HasUnrecognizedSibling ||
		effect.RequiresOrderedLowering ||
		effect.UnsupportedDetail != "" ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	if !effect.Amount.Known ||
		effect.Amount.RangeKnown ||
		effect.Amount.VariableX ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		effect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	if !discardSelectorImposesCardFilter(effect.Selector) {
		return game.AbilityContent{}, false
	}
	selection, ok := cardSelectionForSelector(effect.Selector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ChooseDiscardFromHand{
				Player:    game.ControllerReference(),
				Selection: selection,
			},
		}},
	}.Ability(), true
}

// lowerDiscardEntireHandSpell lowers a "discard their hand" clause to a
// game.Discard with EntireHand set. It supports the controller, each-player,
// each-opponent, and single-target-player subjects recognized by the parser.
func lowerDiscardEntireHandSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported discard spell",
			"the executable source backend supports only exact discard-their-hand by the controller, each player, each opponent, or one target player",
		)
	}
	if effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	switch effect.Context {
	case parser.EffectContextController:
		if len(ctx.content.Targets) != 0 {
			return unsupported()
		}
		return discardEntireHandAbility(game.Discard{EntireHand: true, Player: game.ControllerReference()}, nil)
	case parser.EffectContextEachPlayer:
		if len(ctx.content.Targets) != 0 {
			return unsupported()
		}
		return discardEntireHandAbility(game.Discard{EntireHand: true, PlayerGroup: game.AllPlayersReference()}, nil)
	case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
		if len(ctx.content.Targets) != 0 {
			return unsupported()
		}
		return discardEntireHandAbility(game.Discard{EntireHand: true, PlayerGroup: game.OpponentsReference()}, nil)
	case parser.EffectContextTarget:
		if len(ctx.content.Targets) != 1 {
			return unsupported()
		}
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		return discardEntireHandAbility(
			game.Discard{EntireHand: true, Player: game.TargetPlayerReference(0)},
			[]game.TargetSpec{targetSpec},
		)
	default:
		return unsupported()
	}
}

func discardEntireHandAbility(discard game.Discard, targets []game.TargetSpec) (game.AbilityContent, *shared.Diagnostic) {
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{Primitive: discard},
		},
	}.Ability(), nil
}

func lowerFixedControllerSpell(
	ctx contentCtx,
	_ *parser.Ability,
	verb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	// A scry/surveil verb always acts on the resolving controller, so when it
	// trails a prior-subject, target, or source clause ("Target creature gets
	// +1/+0…, then scry 1." / "~ gets +1/+0…, then scry 1.") its inherited
	// subject target and references denote that earlier object and are simply
	// ignored here. Conditions, keywords, and nested modes must still be consumed.
	inheritedSubject := effect.Context == parser.EffectContextPriorSubject ||
		effect.Context == parser.EffectContextTarget ||
		effect.Context == parser.EffectContextSource
	acceptedContext := controllerActionContext(effect.Context) || inheritedSubject
	unconsumed := ctx.content.Unconsumed()
	if inheritedSubject {
		unconsumed = len(ctx.content.Conditions) != 0 ||
			len(ctx.content.Keywords) != 0 ||
			len(ctx.content.Modes) != 0
	}
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Exact ||
		effect.Negated ||
		!acceptedContext ||
		unconsumed ||
		(!inheritedSubject && len(ctx.content.References) != 0) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	amount, ok := controllerActionQuantity(effect.Amount, allowDynamic)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(amount, game.ControllerReference()),
			},
		},
	}.Ability(), nil
}

// cardCountQuantityForContext resolves a draw/discard/mill card-count amount,
// additionally accepting a "that many" triggering-event anaphor ("that player
// mills that many cards.") and binding it to whichever event fired the enclosing
// triggered ability (the life lost on a life-loss trigger, the damage dealt on a
// combat-damage trigger, and so on). Outside a triggered context, or when the
// caller forbids dynamic amounts, the anaphor has no source and the helper falls
// back to cardCountQuantity, leaving every other amount form unchanged.
func cardCountQuantityForContext(ctx contentCtx, amount compiler.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if allowDynamic && !amount.Known && triggeringEventQuantityKind(amount.DynamicKind) {
		dynamic, ok := lowerTriggeringEventQuantityAmount(ctx, amount)
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	}
	return cardCountQuantity(amount, allowDynamic)
}

func cardCountQuantity(amount compiler.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), amount.VariableX
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

func controllerActionContext(context parser.EffectContextKind) bool {
	switch context {
	case parser.EffectContextController, parser.EffectContextPriorSubject:
		return true
	default:
		return false
	}
}

func controllerActionQuantity(amount compiler.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), amount.VariableX
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}
