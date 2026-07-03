package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// copyChainPaidResultKey wires the copy-chain resolution Pay instruction to the
// result-gated copy consequence: the affected target's controller may pay, and
// only if they pay may they copy the spell.
const copyChainPaidResultKey = game.ResultKey("copy-chain-paid")

// lowerResolvingCopyChain lowers the copy-chain family: a resolving spell that
// performs a base effect on one target, then lets the affected target's
// controller copy the spell with a new target, so the copy chains iteratively
// off each new target. Two forms are handled:
//
//	Return target creature to its owner's hand. Then that creature's controller
//	may pay {U}{U}. If the player does, they may copy this spell and may choose a
//	new target for that copy. (String of Disappearances — payment-gated)
//
//	Destroy target noncreature permanent. Then that permanent's controller may
//	copy this spell and may choose a new target for that copy. (Chain of Acid —
//	unconditional)
//
// The base effect lowers compositionally through the shared per-effect path and
// must yield exactly one target at slot 0, which is both the affected target
// whose controller becomes the copier and the target the copy's chooser is
// resolved against. The copy is a CopyStackObject over the resolving spell with
// MayChooseNewTargets set and Chooser bound to the affected target's controller
// (AffectedTargetControllerReference(0)), so the copier controls the copy and its
// own iterative offer chains off the copier's new target (CR 707.10a).
//
// For the payment-gated form the payment folds onto the copy effect as a
// MayPayThenIfDo mana payment whose payer is the affected target's controller,
// linked to an "If the player does" PriorInstructionAccepted gate; it lowers to a
// resolution Pay instruction publishing its result and a result-gated copy. The
// unconditional form has no payment or condition and lowers to the base effect
// followed by the optional copy. Every other shape fails closed.
func lowerResolvingCopyChain(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		syntax == nil {
		return game.AbilityContent{}, false
	}
	base := ctx.content.Effects[0]
	copyEffect := ctx.content.Effects[1]
	if copyEffect.Kind != compiler.EffectCopyStackObject ||
		!copyEffect.CopyMayChooseNewTargets ||
		copyEffect.Negated ||
		copyEffect.Amount.Known ||
		copyEffect.DelayedTiming != 0 ||
		copyEffect.Duration != compiler.DurationNone ||
		!copyEffectCopiesResolvingSpell(copyEffect) {
		return game.AbilityContent{}, false
	}
	if !base.Exact ||
		base.Negated ||
		base.DelayedTiming != 0 ||
		base.Payment.Form != "" {
		return game.AbilityContent{}, false
	}

	gated := copyEffect.Payment.Form == parser.EffectPaymentFormMayPayThenIfDo &&
		copyEffect.Payment.Payer == parser.EffectPaymentPayerAffectedTargetController
	var resolutionPayment game.ResolutionPayment
	if gated {
		payment := copyEffect.Payment
		if len(ctx.content.Conditions) != 1 {
			return game.AbilityContent{}, false
		}
		condition := ctx.content.Conditions[0]
		if len(payment.ManaCost) == 0 ||
			payment.AdditionalCost != nil ||
			manaCostHasVariableSymbol(payment.ManaCost) ||
			payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
			condition.Kind != compiler.ConditionIf ||
			condition.Predicate != compiler.ConditionPredicatePriorInstructionAccepted ||
			condition.NodeID != payment.SuccessConditionNodeID {
			return game.AbilityContent{}, false
		}
		payout, ok := controllerPaidResolutionPayment(cardName, payment)
		if !ok {
			return game.AbilityContent{}, false
		}
		payout.Payer = opt.Val(game.AffectedTargetControllerReference(0))
		resolutionPayment = payout
	} else if copyEffect.Payment.Form != "" ||
		len(ctx.content.Conditions) != 0 ||
		(copyEffect.Context != parser.EffectContextReferencedObjectController &&
			copyEffect.Context != parser.EffectContextReferencedPlayer) {
		// The unconditional form carries the copy in the base sentence with no
		// folded payment and no gate; its copy context names the affected
		// target's controller or the affected target player directly.
		return game.AbilityContent{}, false
	}

	// Lower the base effect standalone through the shared per-effect path so the
	// family stays text-blind: it fails closed on any base effect the backend
	// cannot already lower. The base must contribute exactly one target, at slot
	// 0, which anchors the copy's chooser.
	//
	// A controller-optional base ("You may tap or untap target creature", Chain
	// Stasis) is lowered without its optionality — the standalone optional-effect
	// path does not lower a lone "you may" resolving effect — and the produced
	// base instructions are then marked Optional so the spell's controller
	// decides whether to apply the base, matching the printed "You may".
	baseOptional := base.Optional
	base.Optional = false
	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	baseCtx := contextForEffect(ctx, &base)
	baseCtx.content.Conditions = nil
	baseContent, diag := lowerContent(cardName, baseCtx, &clauseSyntaxes[0])
	if diag != nil ||
		baseContent.IsModal() ||
		len(baseContent.SharedTargets) != 0 ||
		len(baseContent.Modes) != 1 ||
		len(baseContent.Modes[0].Sequence) == 0 ||
		len(baseContent.Modes[0].Targets) != 1 {
		return game.AbilityContent{}, false
	}
	baseTargets := baseContent.Modes[0].Targets
	baseSequence := baseContent.Modes[0].Sequence
	// The base effect may not carry its own result plumbing or optional envelope,
	// which would collide with the copy gate or the re-applied "you may".
	for i := range baseSequence {
		if baseSequence[i].Optional ||
			baseSequence[i].OptionalActor.Exists ||
			baseSequence[i].PublishResult != "" ||
			baseSequence[i].ResultGate.Exists {
			return game.AbilityContent{}, false
		}
	}
	if baseOptional {
		for i := range baseSequence {
			baseSequence[i].Optional = true
		}
	}

	chooser := opt.Val(game.AffectedTargetControllerReference(0))
	copyInstruction := game.Instruction{
		Primitive: game.CopyStackObject{
			Object:              game.ResolvingStackObjectReference(),
			MayChooseNewTargets: true,
			Chooser:             chooser,
		},
		Optional:      true,
		OptionalActor: chooser,
	}

	sequence := make([]game.Instruction, 0, len(baseSequence)+2)
	sequence = append(sequence, baseSequence...)
	if gated {
		copyInstruction.ResultGate = opt.Val(game.InstructionResultGate{
			Key:       copyChainPaidResultKey,
			Succeeded: game.TriTrue,
		})
		sequence = append(sequence, game.Instruction{
			Primitive:     game.Pay{Payment: resolutionPayment},
			PublishResult: copyChainPaidResultKey,
		})
	}
	sequence = append(sequence, copyInstruction)

	return game.Mode{
		Targets:  baseTargets,
		Sequence: sequence,
	}.Ability(), true
}

// copyEffectCopiesResolvingSpell reports whether a copy-stack-object effect
// copies the resolving spell itself ("copy this spell"), identified by a
// self-name or "this spell" reference among the effect's references. The
// copy-chain family always copies the resolving spell, never a targeted stack
// object.
func copyEffectCopiesResolvingSpell(effect compiler.CompiledEffect) bool {
	if len(effect.Targets) != 0 {
		return false
	}
	for i := range effect.References {
		switch effect.References[i].Kind {
		case compiler.ReferenceThisObject, compiler.ReferenceSelfName:
			return true
		default:
			// Not a self-copy reference; keep scanning.
		}
	}
	return false
}
