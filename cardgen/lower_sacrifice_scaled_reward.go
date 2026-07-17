package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// sacrificedCreatureLinkKey is the linked-object key under which an optional
// resolving sacrifice records the permanent it sacrificed, so a gated reward
// effect can read that creature's power/toughness/mana value through last-known
// information once it has left the battlefield.
const sacrificedCreatureLinkKey = game.LinkedKey("sacrificed-creature")

// lowerSacrificeReflexiveDamage lowers optional or mandatory creature sacrifices
// followed by reflexive source damage scaled to the sacrificed creature's power.
// The reflexive body may also create tokens, or conditionally draw the same amount
// based on the sacrificed creature's subtype. Its target is chosen only after the
// sacrifice resolves, at the proper CR 603.11 time.
func lowerSacrificeReflexiveDamage(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Modes) != 0 ||
		len(content.Targets) != 1 ||
		len(content.Keywords) != 0 ||
		len(content.Effects) < 2 ||
		len(content.Effects) > 3 ||
		len(content.Conditions) < 1 ||
		len(content.Conditions) > 2 {
		return game.AbilityContent{}, false
	}
	sacrifice, damage := &content.Effects[0], &content.Effects[1]
	if sacrifice.Kind != compiler.EffectSacrifice ||
		!sacrifice.Exact ||
		sacrifice.Negated ||
		sacrifice.Context != parser.EffectContextController ||
		!sacrifice.Amount.Known ||
		sacrifice.Amount.Value != 1 ||
		damage.Kind != compiler.EffectDealDamage ||
		damage.Optional ||
		!damage.Exact ||
		damage.Negated {
		return game.AbilityContent{}, false
	}
	if !effectAmountBindsPriorInstruction(damage, content.References, 0) {
		return game.AbilityContent{}, false
	}
	if damage.Amount.DynamicKind != compiler.DynamicAmountSourcePower {
		return game.AbilityContent{}, false
	}
	if !damageSourceReferencesSource(*damage) {
		return game.AbilityContent{}, false
	}
	target, ok := singleAnyTargetSpec(content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	var subtypeCondition compiler.CompiledCondition
	foundReflexive, foundSubtype := false, false
	for i := range content.Conditions {
		condition := content.Conditions[i]
		switch {
		case condition.Reflexive &&
			condition.Predicate == compiler.ConditionPredicatePriorInstructionAccepted:
			foundReflexive = true
		case condition.Predicate == compiler.ConditionPredicateObjectMatches &&
			condition.ObjectBinding == compiler.ReferenceBindingPriorInstructionResult &&
			len(content.Effects) == 3 &&
			condition.Order.Start > damage.VerbOrder.Start &&
			condition.Order.Start < content.Effects[2].VerbOrder.Start:
			subtypeCondition = condition
			foundSubtype = true
		default:
			return game.AbilityContent{}, false
		}
	}
	if !foundReflexive || !foundSubtype {
		if !foundReflexive || len(content.Conditions) != 1 {
			return game.AbilityContent{}, false
		}
	}
	sacrificeInstr, ok := lowerSequenceSacrificeInstruction(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrificePrimitive, ok := sacrificeInstr.Primitive.(game.SacrificePermanents)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrificePrimitive.PublishLinked = sacrificedCreatureLinkKey
	sacrificePrimitive.PublishObjectBinding = true
	sacrificeInstr.Primitive = sacrificePrimitive
	sacrificeInstr.Optional = sacrifice.Optional
	sacrificeInstr.PublishResult = optionalIfYouDoResultKey

	sharedAmount, hasSharedAmount := sacrificeScaledSharedAmount(content.Effects[1:])
	damageAmount, ok := sacrificeScaledRewardAmount(damage, sharedAmount, hasSharedAmount)
	if !ok {
		return game.AbilityContent{}, false
	}
	innerSequence := []game.Instruction{{
		Primitive: game.Damage{
			Amount:       damageAmount,
			Recipient:    game.AnyTargetDamageRecipient(0),
			DamageSource: opt.Val(game.SourcePermanentReference()),
		},
	}}
	if foundSubtype {
		if len(content.Effects) != 3 {
			return game.AbilityContent{}, false
		}
		draw := &content.Effects[2]
		if draw.Kind != compiler.EffectDraw ||
			draw.Optional ||
			!draw.Exact ||
			draw.Negated ||
			draw.Context != parser.EffectContextController {
			return game.AbilityContent{}, false
		}
		selection, ok := lowerConditionSelection(subtypeCondition.Selection)
		if !ok || len(selection.SubtypesAny) == 0 {
			return game.AbilityContent{}, false
		}
		if sacrificedCharacteristicKind(draw.Amount.DynamicKind) &&
			!effectAmountBindsPriorInstruction(draw, content.References, 0) {
			return game.AbilityContent{}, false
		}
		drawAmount, ok := sacrificeScaledRewardAmount(draw, sharedAmount, hasSharedAmount)
		if !ok {
			return game.AbilityContent{}, false
		}
		innerSequence = append(innerSequence, game.Instruction{
			Primitive: game.Draw{
				Player: game.ControllerReference(),
				Amount: drawAmount,
			},
			Condition: opt.Val(game.EffectCondition{
				Text:   subtypeCondition.Text,
				Object: game.LinkedObjectReference(string(sacrificedCreatureLinkKey)),
				Condition: opt.Val(game.Condition{
					Text: subtypeCondition.Text,
					Object: opt.Val(
						game.LinkedObjectReference(string(sacrificedCreatureLinkKey)),
					),
					ObjectMatches: opt.Val(selection),
				}),
			}),
		})
	} else if len(content.Effects) == 3 {
		create := &content.Effects[2]
		if create.Kind != compiler.EffectCreate {
			return game.AbilityContent{}, false
		}
		createCtx := contextForEffect(ctx, create)
		createCtx.content.Conditions = nil
		created, diagnostic := lowerCreateTokenSpell(createCtx)
		if diagnostic != nil ||
			len(created.Modes) != 1 ||
			len(created.Modes[0].Targets) != 0 ||
			len(created.Modes[0].Sequence) != 1 {
			return game.AbilityContent{}, false
		}
		innerSequence = append(innerSequence, created.Modes[0].Sequence[0])
	}
	inner := game.Mode{
		Targets:  []game.TargetSpec{target},
		Sequence: innerSequence,
	}.Ability()
	outer := game.Mode{Sequence: []game.Instruction{
		sacrificeInstr,
		{
			Primitive: game.CreateReflexiveTrigger{
				Trigger: game.ReflexiveTriggerDef{Content: inner},
			},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       optionalIfYouDoResultKey,
				Succeeded: game.TriTrue,
			}),
		},
	}}.Ability()
	err := game.ValidateInstructionSequence(
		outer.Modes[0].Sequence,
		outer.Modes[0].Targets,
	)
	return outer, err == nil
}

func effectAmountBindsPriorInstruction(
	effect *compiler.CompiledEffect,
	references []compiler.CompiledReference,
	prior int,
) bool {
	for i := range references {
		if references[i].NodeID == effect.Amount.ReferenceNodeID {
			return references[i].Binding == compiler.ReferenceBindingPriorInstructionResult &&
				references[i].PriorInstruction == prior
		}
	}
	return false
}

func damageSourceReferencesSource(effect compiler.CompiledEffect) bool {
	for i := range effect.SubjectReferences {
		if effect.SubjectReferences[i].Binding == compiler.ReferenceBindingSource {
			return true
		}
	}
	return false
}

func singleAnyTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !target.Exact ||
		target.Cardinality.Min != 1 ||
		target.Cardinality.Max != 1 ||
		target.Selector.Kind != compiler.SelectorAny ||
		selectorHasUnsupportedPermanentFilters(target.Selector) {
		return game.TargetSpec{}, false
	}
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
		Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
	}, true
}

// lowerOptionalSacrificeScaledReward lowers the "you may sacrifice another
// creature. If you do, <rewards>" family where each reward is a controller life
// gain or card draw scaled by the sacrificed creature's power, toughness, or
// mana value (Disciple of Freyalise's "you gain X life and draw X cards, where X
// is that creature's power."). The optional sacrifice publishes both its
// success (gating the rewards) and the sacrificed permanent as a linked object;
// each reward reads that permanent's characteristic via DynamicAmountObjectPower
// and friends. Any other shape — a non-creature sacrifice, an else branch, a
// reward that is neither a life gain nor a draw, or an unmodeled reference —
// fails closed.
func lowerOptionalSacrificeScaledReward(ctx contentCtx) (game.AbilityContent, bool) {
	content := ctx.content
	if len(content.Modes) != 0 ||
		len(content.Targets) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Effects) < 2 ||
		len(content.Conditions) != 1 {
		return game.AbilityContent{}, false
	}
	sacrifice := &content.Effects[0]
	if sacrifice.Kind != compiler.EffectSacrifice ||
		!sacrifice.Optional ||
		!sacrifice.Exact ||
		sacrifice.Negated ||
		sacrifice.Context != parser.EffectContextController ||
		!sacrifice.Amount.Known ||
		sacrifice.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	// The optionality must be exactly "you may X. If you do, <tail>" with no else
	// branch, so every reward is gated on the sacrifice having succeeded.
	plan, ok := planOptionalFlow(content)
	if !ok ||
		!plan.enabled ||
		plan.publishWithoutOptional ||
		plan.optionalIndex != 0 ||
		plan.gateIndex != 1 ||
		plan.elseIndex >= 0 {
		return game.AbilityContent{}, false
	}
	// Only the source self-reference and the "that creature" demonstratives that
	// name the sacrificed creature are permitted; any other reference denotes an
	// unmodeled object and fails closed.
	for i := range content.References {
		switch content.References[i].Kind {
		case compiler.ReferenceThisObject, compiler.ReferenceSelfName, compiler.ReferenceThatObject:
		default:
			return game.AbilityContent{}, false
		}
	}

	sharedAmount, hasShared := sacrificeScaledSharedAmount(content.Effects[1:])
	rewards := make([]game.Instruction, 0, len(content.Effects)-1)
	for i := 1; i < len(content.Effects); i++ {
		effect := &content.Effects[i]
		amount, ok := sacrificeScaledRewardAmount(effect, sharedAmount, hasShared)
		if !ok {
			return game.AbilityContent{}, false
		}
		primitive, ok := sacrificeScaledRewardPrimitive(effect, amount)
		if !ok {
			return game.AbilityContent{}, false
		}
		rewards = append(rewards, game.Instruction{
			Primitive: primitive,
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       optionalIfYouDoResultKey,
				Succeeded: game.TriTrue,
			}),
		})
	}

	sacrificeInstr, ok := lowerSequenceSacrificeInstruction(ctx)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrifice0, ok := sacrificeInstr.Primitive.(game.SacrificePermanents)
	if !ok {
		return game.AbilityContent{}, false
	}
	sacrifice0.PublishLinked = sacrificedCreatureLinkKey
	sacrificeInstr.Primitive = sacrifice0
	sacrificeInstr.Optional = true
	sacrificeInstr.PublishResult = optionalIfYouDoResultKey

	sequence := make([]game.Instruction, 0, len(rewards)+1)
	sequence = append(sequence, sacrificeInstr)
	sequence = append(sequence, rewards...)
	return game.Mode{Sequence: sequence}.Ability(), true
}

// sacrificeScaledSharedAmount resolves the lone "where X is that creature's
// <characteristic>" reward definition so sibling bare-X reward clauses can adopt
// it. It returns false when no reward clause carries such a definition.
func sacrificeScaledSharedAmount(rewards []compiler.CompiledEffect) (game.Quantity, bool) {
	for i := range rewards {
		amount := rewards[i].Amount
		if amount.DynamicForm != compiler.DynamicAmountWhereX {
			continue
		}
		if quantity, ok := sacrificedCharacteristicAmount(amount); ok {
			return quantity, true
		}
	}
	return game.Quantity{}, false
}

// sacrificeScaledRewardAmount resolves one reward clause's quantity: its own
// "that creature's <characteristic>" definition, or the shared definition when
// the clause carries only the bare variable X.
func sacrificeScaledRewardAmount(
	effect *compiler.CompiledEffect,
	shared game.Quantity,
	hasShared bool,
) (game.Quantity, bool) {
	amount := effect.Amount
	if sacrificedCharacteristicKind(amount.DynamicKind) {
		return sacrificedCharacteristicAmount(amount)
	}
	if amount.VariableX && amount.DynamicKind == compiler.DynamicAmountNone && hasShared {
		return shared, true
	}
	return game.Quantity{}, false
}

// sacrificedCharacteristicAmount lowers a "that creature's power/toughness/mana
// value" amount that reads the sacrificed creature through the linked object the
// sacrifice published. It fails closed for any other dynamic kind.
func sacrificedCharacteristicAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	if !sacrificedCharacteristicKind(amount.DynamicKind) {
		return game.Quantity{}, false
	}
	dynamic, ok := lowerDynamicAmount(amount, game.LinkedObjectReference(string(sacrificedCreatureLinkKey)))
	if !ok {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

// sacrificedCharacteristicKind reports whether a dynamic-amount kind reads a
// referenced object's power, toughness, or mana value — the characteristics a
// sacrificed creature can scale a reward by.
func sacrificedCharacteristicKind(kind compiler.DynamicAmountKind) bool {
	switch kind {
	case compiler.DynamicAmountSourcePower,
		compiler.DynamicAmountSourceToughness,
		compiler.DynamicAmountSourceManaValue:
		return true
	default:
		return false
	}
}

// sacrificeScaledRewardPrimitive builds a reward clause's primitive: a
// controller life gain or card draw of the resolved quantity. Any other effect
// kind, recipient, or modifier fails closed.
func sacrificeScaledRewardPrimitive(
	effect *compiler.CompiledEffect,
	amount game.Quantity,
) (game.Primitive, bool) {
	if effect.Context != parser.EffectContextController ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional {
		return nil, false
	}
	switch effect.Kind {
	case compiler.EffectGain:
		if !effect.LifeObject {
			return nil, false
		}
		return game.GainLife{Player: game.ControllerReference(), Amount: amount}, true
	case compiler.EffectDraw:
		return game.Draw{Player: game.ControllerReference(), Amount: amount}, true
	default:
		return nil, false
	}
}
