package cardgen

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerGroupDamageSpell(
	_ string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	amount, amountOK := groupDamageAmount(effect.Amount)
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		!amountOK ||
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
	damageSource, ok := lowerDamageSourceReference(ctx.content.References)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	sel := effect.Selector
	recipientSelectors := []compiler.CompiledSelector{sel}
	if len(effect.DamageRecipientSelectors) > 0 {
		recipientSelectors = effect.DamageRecipientSelectors
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
	var damageSourceRef opt.V[game.ObjectReference]
	if damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damageSourceRef = opt.Val(damageSource)
	} else if damageSourceIsSourcePermanent(ctx.content.References) {
		damageSourceRef = opt.Val(game.SourcePermanentReference())
	}
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

// groupDamageAmount resolves the supported group-damage amounts onto a runtime
// Quantity: an exact fixed amount of at least one, or the spell's X. The
// executable backend deals the resolved amount to every member of each
// recipient group, so a fixed or X amount needs no per-recipient computation. It
// fails closed for a zero or negative fixed amount and for every dynamic amount
// form (e.g. "where X is the number of ...") the group path cannot reconstruct
// exactly, leaving those spells rejected.
func groupDamageAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	if amount.DynamicKind != compiler.DynamicAmountNone ||
		amount.DynamicForm != compiler.DynamicAmountFormNone {
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

// groupDamageRecipientFor resolves one fixed group-damage recipient selector
// onto a runtime DamageRecipient: the all-players and opponents player groups,
// or a filtered battlefield permanent group. It fails closed for any selector
// the executable backend cannot damage as a group so unsupported recipients stay
// rejected.
func groupDamageRecipientFor(sel compiler.CompiledSelector) (game.DamageRecipient, bool) {
	switch {
	case sel.Kind == compiler.SelectorOpponent && !sel.Other:
		return game.PlayerGroupDamageRecipient(game.OpponentsReference()), true
	case sel.Kind == compiler.SelectorPlayer && !sel.Other:
		return game.PlayerGroupDamageRecipient(game.AllPlayersReference()), true
	default:
		group, ok := damageGroupRecipient(sel)
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
	selection, ok := damageGroupSelection(sel)
	if !ok {
		return game.GroupReference{}, false
	}
	if sel.Other {
		return game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()), true
	}
	return game.BattlefieldGroup(selection), true
}

// damageGroupSelection translates the supported filters of a group-damage
// recipient selector (controller, combat, tapped, single color/subtype/excluded
// type, keyword) onto a runtime Selection, failing closed for any selector field
// it cannot represent exactly so unsupported recipients stay rejected.
func damageGroupSelection(sel compiler.CompiledSelector) (game.Selection, bool) {
	if sel.All || sel.Another || sel.Zone != zone.None ||
		sel.MatchManaValue || sel.MatchPower || sel.MatchToughness ||
		sel.Colorless || sel.Multicolored ||
		len(sel.RequiredTypesAny()) != 0 ||
		len(sel.Supertypes()) != 0 ||
		len(sel.ExcludedColors()) != 0 ||
		len(sel.ColorsAny()) > 1 ||
		len(sel.SubtypesAny()) > 1 ||
		len(sel.ExcludedTypes()) > 1 {
		return game.Selection{}, false
	}
	if (sel.Attacking && sel.Blocking) ||
		(sel.Tapped && sel.Untapped) ||
		((sel.Tapped || sel.Untapped) && (sel.Attacking || sel.Blocking)) {
		return game.Selection{}, false
	}
	requiredType, hasNoun, ok := damageGroupRequiredType(sel.Kind)
	if !ok {
		return game.Selection{}, false
	}
	if !hasNoun && len(sel.SubtypesAny()) != 1 {
		return game.Selection{}, false
	}
	selection, ok := selectorCharacteristics(sel)
	if !ok {
		return game.Selection{}, false
	}
	if requiredType != "" {
		selection.RequiredTypes = []types.Card{requiredType}
	}
	switch sel.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		selection.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		selection.Controller = game.ControllerNotYou
	default:
		return game.Selection{}, false
	}
	switch {
	case sel.Attacking:
		selection.CombatState = game.CombatStateAttacking
	case sel.Blocking:
		selection.CombatState = game.CombatStateBlocking
	default:
	}
	switch {
	case sel.Tapped:
		selection.Tapped = game.TriTrue
	case sel.Untapped:
		selection.Tapped = game.TriFalse
	default:
	}
	return selection, true
}

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
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	var damageSource game.ObjectReference
	var sourceBound bool
	if len(ctx.content.References) > 0 {
		damageSource, sourceBound = lowerDamageSourceReference(ctx.content.References[:1])
	}
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
	} else if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		amountObject := game.SourcePermanentReference()
		if sourceBound {
			amountObject = damageSource
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, amountObject)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		amount = game.Dynamic(dynamic)
	}
	if effect.DamageRecipientReference != parser.DamageRecipientReferenceNone {
		return lowerReferencedPlayerDamageSpell(ctx, effect.DamageRecipientReference, amount, damageSource, sourceBound)
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok ||
		!exactDamageSourceSyntax(ctx.content.References) ||
		!exactDamageAmountReferences(effect.Amount, ctx.content.References) {
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
	if sourceBound && damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	} else if damageSourceIsSourcePermanent(ctx.content.References) ||
		effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
	}
	instructions := []game.Instruction{{Primitive: damage}}
	// "deals A damage to <target> and B damage to you" appends a second Damage
	// instruction dealing the fixed rider amount to the source's own controller.
	if effect.HasSelfDamageRider {
		if !effect.Amount.Known || effect.SelfDamageRiderValue < 1 {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		rider := game.Damage{
			Amount:       game.Fixed(effect.SelfDamageRiderValue),
			Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
			DamageSource: damage.DamageSource,
		}
		instructions = append(instructions, game.Instruction{Primitive: rider})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{target},
		Sequence: instructions,
	}.Ability(), nil
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
	damageSource game.ObjectReference,
	sourceBound bool,
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
		Amount:    amount,
		Recipient: game.PlayerDamageRecipient(recipient),
	}
	if sourceBound && damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	} else if damageSourceIsSourcePermanent(ctx.content.References[:1]) {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
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
// selector kind drives the object reference: a permanent target yields a
// permanent reference, a spell target a stack-object reference. It fails closed
// for any other shape.
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
	var object game.ObjectReference
	switch ctx.content.Targets[0].Selector.Kind {
	case compiler.SelectorArtifact, compiler.SelectorCreature, compiler.SelectorEnchantment,
		compiler.SelectorLand, compiler.SelectorPermanent, compiler.SelectorPlaneswalker,
		compiler.SelectorBattle:
		object = game.TargetPermanentReference(occ)
	case compiler.SelectorSpell:
		object = game.TargetStackObjectReference(occ)
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

// lowerControllerDamageSpell lowers a "deals N damage to you" effect, whose
// recipient is the source's own controller, as in "this creature deals 1 damage
// to you." or "Sell-Sword Brute deals 2 damage to you." The recipient binds to
// the resolving ability's controller; the amount is a fixed value or X. It emits
// one Damage instruction with a controller player recipient and no target spec,
// failing closed for any shape outside that exact template (a non-"you"
// recipient, a dynamic count amount, any target, condition, keyword, or mode).
func lowerControllerDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed or X damage to you",
		)
	}
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		effect.DamageRecipientReference != parser.DamageRecipientReferenceYou ||
		!effect.Exact ||
		effect.Negated ||
		effect.Divided ||
		len(ctx.content.Targets) != 0 ||
		len(effect.DamageRecipientSelectors) != 0 ||
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
		Amount:    amount,
		Recipient: game.PlayerDamageRecipient(game.ControllerReference()),
	}
	if damageSource, ok := lowerDamageSourceReference(ctx.content.References); ok &&
		damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	} else if damageSourceIsSourcePermanent(ctx.content.References) {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), nil
}

// removalTargetSpecForRecipient rebuilds the inherited removal target's spec for
// the recipient-damage clause. In the ordered-sequence shared-target path the
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

// lowerDividedDamageSpell lowers a "deals N damage divided as you choose among
// <cardinality> <targets>" effect: a fixed total split among the chosen targets,
// at least one to each at resolution (CR 601.2d). It emits one multi-target spec
// and a single Divided Damage instruction whose recipient addresses that spec.
// It fails closed for any shape the executable backend cannot represent exactly.
func lowerDividedDamageSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		!effect.Exact ||
		!effect.Divided ||
		(effect.Context != parser.EffectContextSource &&
			effect.Context != parser.EffectContextReferencedObject &&
			effect.Context != parser.EffectContextPriorSubject) ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported divided damage spell",
			"the executable source backend supports only an exact fixed total divided among one supported multi-target spec",
		)
	}
	total := effect.Amount.Value
	target, ok := dividedDamageTargetSpec(ctx.content.Targets[0], total)
	if !ok ||
		!exactDamageSourceSyntax(ctx.content.References) ||
		!exactDamageAmountReferences(effect.Amount, ctx.content.References) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported divided damage spell",
			"the executable source backend supports only an exact fixed total divided among one supported multi-target spec",
		)
	}
	damage := game.Damage{
		Amount:    game.Fixed(total),
		Recipient: game.AnyTargetDamageRecipient(0),
		Divided:   true,
	}
	var damageSource game.ObjectReference
	var sourceBound bool
	if len(ctx.content.References) > 0 {
		damageSource, sourceBound = lowerDamageSourceReference(ctx.content.References[:1])
	}
	if sourceBound && damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	} else if damageSourceIsSourcePermanent(ctx.content.References) {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
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

// dividedDamageTargetSpec builds the multi-target spec a divided-damage effect
// chooses among. The minimum is one (a divided spell must have at least one
// target); the maximum is the smaller of the wording's bound and the total,
// since each chosen target must receive at least one damage. It supports only
// the "any target" and plain "creature" selectors the parser marks exact.
func dividedDamageTargetSpec(target compiler.CompiledTarget, total int) (game.TargetSpec, bool) {
	if !target.Exact && target.Cardinality.Max < 1 {
		return game.TargetSpec{}, false
	}
	maxTargets := target.Cardinality.Max
	if maxTargets < 1 || maxTargets > total {
		maxTargets = total
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
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
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
	if len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDealDamage ||
		!effect.Exact ||
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
	if source.Kind != compiler.ReferencePronoun ||
		source.Pronoun != compiler.ReferencePronounIt ||
		source.Binding != compiler.ReferenceBindingTarget ||
		source.Occurrence < 0 ||
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
	if len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDealDamage ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextTarget ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourcePower ||
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

// lowerEachOfTargetsDamageSpell lowers "deals N damage to each of <cardinality>
// <targets>" effects, which deal the full fixed amount to each of the chosen
// targets (unlike divided damage, which splits one total). It emits one Damage
// instruction per target slot, each addressing its own flat target index, the
// same per-slot pattern the multi-target pump path uses. Declined "up to N"
// slots leave fewer chosen targets and the runtime Damage no-ops on an
// unresolved target index, so only the chosen targets take damage. The recipient
// may be an "any target" slot (permanent or player) or a creature target. It
// fails closed (ok=false) for dynamic amounts, divided damage, riders, or any
// other selector so the single-target path and its diagnostic stay unchanged.
func lowerEachOfTargetsDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 || len(ctx.content.Targets) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDealDamage ||
		!effect.Exact ||
		effect.Negated ||
		effect.Divided ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.DamageRecipientReference != parser.DamageRecipientReferenceNone ||
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
	var damageSource game.ObjectReference
	var sourceBound bool
	if len(ctx.content.References) > 0 {
		damageSource, sourceBound = lowerDamageSourceReference(ctx.content.References[:1])
	}
	var damageSourceRef opt.V[game.ObjectReference]
	if sourceBound && damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damageSourceRef = opt.Val(damageSource)
	} else if damageSourceIsSourcePermanent(ctx.content.References) {
		damageSourceRef = opt.Val(game.SourcePermanentReference())
	}
	sequence := make([]game.Instruction, 0, spec.MaxTargets)
	for i := range spec.MaxTargets {
		sequence = append(sequence, game.Instruction{Primitive: game.Damage{
			Amount:       game.Fixed(effect.Amount.Value),
			Recipient:    game.AnyTargetDamageRecipient(i),
			DamageSource: damageSourceRef,
		}})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{spec},
		Sequence: sequence,
	}.Ability(), true
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

func lowerFixedModifyPTSpell(
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := &ctx.content.Effects[0]
	if effect.StaticSubject != compiler.StaticSubjectNone {
		return lowerFixedGroupModifyPTSpell(ctx, effect)
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
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
		switch effect.Amount.DynamicForm {
		case compiler.DynamicAmountWhereX:
			powerDelta = whereXSignedQuantity(dynamic, effect.PowerDelta)
			toughnessDelta = whereXSignedQuantity(dynamic, effect.ToughnessDelta)
		case compiler.DynamicAmountForEach:
			powerDelta = dynamicSignedQuantity(dynamic, effect.PowerDelta)
			toughnessDelta = dynamicSignedQuantity(dynamic, effect.ToughnessDelta)
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

// lowerFixedModifyPTTargets lowers an exact fixed until-end-of-turn power/
// toughness pump whose single target slot may be single ("Target creature gets
// +1/+1 until end of turn."), plural ("Two target creatures each get -1/-1
// until end of turn."), or optional ("Up to one target creature gets -2/-2
// until end of turn."), and whose selector may name a creature or a creature
// subtype ("Target Goblin you control gets +1/+0 until end of turn."). It emits
// one ModifyPT per target slot, each addressing its own slot, mirroring the
// multi-target permanent verbs. Declined "up to" slots leave fewer chosen
// targets and the runtime ModifyPT no-ops on an unresolved target index, so the
// spell pumps only the chosen targets. It returns ok=false for any shape outside
// this bounded set (dynamic amounts, riders, "you may" optionality, or a
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
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
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
	powerDelta := game.Fixed(compiledSignedAmountValue(effect.PowerDelta))
	toughnessDelta := game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta))
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

func lowerFixedGroupModifyPTSpell(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported group power/toughness spell",
			"the executable source backend supports only exact fixed supported group power/toughness changes until end of turn",
		)
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known {
		return unsupported()
	}
	group, ok := resolvingStaticSubjectGroup(effect)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:          game.LayerPowerToughnessModify,
					Group:          group,
					PowerDelta:     compiledSignedAmountValue(effect.PowerDelta),
					ToughnessDelta: compiledSignedAmountValue(effect.ToughnessDelta),
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
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
	if effect.StaticSubject != compiler.StaticSubjectNone {
		return lowerGroupTemporaryKeywordSpell(ctx, unsupported)
	}
	referencedObject := len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingTarget &&
		effect.Context == parser.EffectContextReferencedObject
	sourceSubject := len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingSource &&
		effect.Context == parser.EffectContextSource
	targetSubject := len(ctx.content.Targets) == 1 &&
		len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextTarget &&
		temporaryKeywordTarget(ctx.content.Targets[0])
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Kind != compiler.EffectGain ||
		!effect.Exact ||
		(!targetSubject && !referencedObject && !sourceSubject) ||
		effect.Negated ||
		effect.StaticSubject != compiler.StaticSubjectNone ||
		effect.Duration != compiler.DurationUntilEndOfTurn {
		return unsupported()
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	var object game.ObjectReference
	var target opt.V[game.TargetSpec]
	switch {
	case targetSubject:
		spec, ok := permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		target = opt.Val(spec)
		object = game.TargetPermanentReference(0)
	case sourceSubject:
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowSource:      true,
			SourceCardObject: true,
		})
		if !ok {
			return unsupported()
		}
	default:
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
		if !ok {
			return unsupported()
		}
	}
	mode := game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(object),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					AddKeywords: keywords,
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}
	if target.Exists {
		mode.Targets = []game.TargetSpec{target.Val}
	}
	return mode.Ability(), nil
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
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Kind != compiler.EffectGain ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn {
		return unsupported()
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	group, ok := resolvingStaticSubjectGroup(&effect)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					Group:       group,
					AddKeywords: keywords,
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

// lowerTemporaryPTKeywordSpell lowers the single-subject combined buff
// "<target creature(s)> get(s) +N/+N and gain <keyword(s)> until end of turn."
// into one game.ApplyContinuous per target slot, each carrying both a
// power/toughness layer and a keyword layer. The parser splits the body into a
// target EffectModifyPT and a prior-subject EffectGain sharing one span; both
// must be exact and until-end-of-turn with fixed deltas. The target slot may be
// single ("Target creature gets +1/+1 and gains trample…") or multi-cardinality
// ("Up to two target creatures each get +1/+1 and gain lifelink…"); a declined
// "up to" slot leaves an unresolved target index that the runtime
// ApplyContinuous no-ops, so only chosen creatures are buffed. It fails closed
// for any richer shape.
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
		!modifyEffect.Exact ||
		!keywordEffect.Exact ||
		modifyEffect.Negated ||
		keywordEffect.Negated ||
		modifyEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		modifyEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		modifyEffect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!modifyEffect.PowerDelta.Known ||
		!modifyEffect.ToughnessDelta.Known {
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
						Layer:          game.LayerPowerToughnessModify,
						PowerDelta:     compiledSignedAmountValue(modifyEffect.PowerDelta),
						ToughnessDelta: compiledSignedAmountValue(modifyEffect.ToughnessDelta),
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

func temporaryKeywordTarget(target compiler.CompiledTarget) bool {
	return target.Selector.Kind == compiler.SelectorCreature ||
		target.Selector.Kind == compiler.SelectorPermanent
}

// lowerGroupTemporaryPTKeywordSpell lowers the Overrun-style group buff
// "<group> get +N/+N and gain <keyword(s)> until end of turn." into a single
// game.ApplyContinuous over a battlefield group with both a power/toughness layer
// and a keyword layer. The parser splits the body into a group EffectModifyPT and
// a prior-subject EffectGain; both must be exact and until-end-of-turn with fixed
// deltas. It fails closed for any richer shape.
func lowerGroupTemporaryPTKeywordSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectModifyPT ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	modifyEffect := ctx.content.Effects[0]
	keywordEffect := ctx.content.Effects[1]
	if !modifyEffect.Exact ||
		!keywordEffect.Exact ||
		modifyEffect.Negated ||
		keywordEffect.Negated ||
		modifyEffect.StaticSubject == compiler.StaticSubjectNone ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.Context != parser.EffectContextPriorSubject ||
		modifyEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		modifyEffect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!modifyEffect.PowerDelta.Known ||
		!modifyEffect.ToughnessDelta.Known {
		return game.AbilityContent{}, false
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return game.AbilityContent{}, false
	}
	group, ok := resolvingStaticSubjectGroup(&modifyEffect)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer:          game.LayerPowerToughnessModify,
						Group:          group,
						PowerDelta:     compiledSignedAmountValue(modifyEffect.PowerDelta),
						ToughnessDelta: compiledSignedAmountValue(modifyEffect.ToughnessDelta),
					},
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
// permanent returning itself ("Return this creature to its owner's hand."): the
// first reference is the source object and every reference binds to the source.
func selfSourceBounceReferences(references []compiler.CompiledReference) bool {
	if len(references) == 0 || references[0].Kind != compiler.ReferenceThisObject {
		return false
	}
	for i := range references {
		if references[i].Binding != compiler.ReferenceBindingSource {
			return false
		}
	}
	return true
}

func lowerFixedBounceSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Targets) == 0 &&
		!effect.Negated && !effect.Optional && effect.Exact && !ctx.optional &&
		effect.Context == parser.EffectContextController &&
		effect.ToZone == zone.Hand &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0 &&
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
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].Optional ||
		!ctx.content.Effects[0].Exact ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		ctx.content.Effects[0].ToZone != zone.Hand ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
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

// lowerMultiTargetBounceSpell lowers a battlefield bounce whose single permanent
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
		effect.Negated ||
		effect.Optional ||
		!effect.Exact ||
		ctx.optional ||
		effect.Context != parser.EffectContextController ||
		effect.ToZone != zone.Hand ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
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
		effect.Negated ||
		effect.Optional ||
		!effect.Exact ||
		ctx.optional ||
		effect.Context != parser.EffectContextController ||
		effect.ToZone != zone.Hand ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
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
		effect.Negated ||
		effect.Optional ||
		!effect.Exact ||
		ctx.optional ||
		effect.Context != parser.EffectContextController ||
		effect.ToZone != zone.Hand ||
		effect.Selector.All ||
		effect.Selector.Controller != compiler.ControllerYou ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
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

// bounceDestinationPossessiveReferencesOnly reports whether every reference is
// the plural "their" possessive pronoun that names the bounce destination
// ("their owners' hands"). The compiler cannot bind a possessive pronoun to a
// multi-target permanent so it leaves it ambiguous; the mass group bounce
// addresses the group directly rather than through the reference, so the
// destination possessive is the only reference the lowering tolerates. Any
// other reference fails closed.
func bounceDestinationPossessiveReferencesOnly(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun ||
			reference.Pronoun != compiler.ReferencePronounTheir ||
			reference.Binding != compiler.ReferenceBindingAmbiguous {
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
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].Optional ||
		!ctx.content.Effects[0].Exact ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
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
		(len(ctx.content.References) != 0 && !hasEventPlayerRef && !hasReferencedControllerRef) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	amount, ok := cardCountQuantity(effect.Amount, allowDynamic)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	if len(ctx.content.Targets) == 0 && len(ctx.content.References) == 0 {
		switch effect.Context {
		case parser.EffectContextEachOpponent:
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
		}
	}
	switch {
	case hasEventPlayerRef && len(ctx.content.Targets) == 0 &&
		(effect.Context == parser.EffectContextEventPlayer || effect.Context == parser.EffectContextReferencedPlayer) &&
		effect.Amount.Known:
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
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(amount, playerRef),
			},
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
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 {
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
