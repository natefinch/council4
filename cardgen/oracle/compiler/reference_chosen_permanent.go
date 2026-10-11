package compiler

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

type chosenPermanentNameProof struct {
	selector       CompiledSelector
	referenceOrder shared.SourceOrder
	producerOrder  shared.SourceOrder
	paymentOrder   shared.SourceOrder
	paymentForm    parser.EffectPaymentForm
	paymentPayer   parser.EffectPaymentPayerKind
	paymentNodeID  int
	paymentMana    cost.Mana
	paymentCost    *CompiledCost
	gates          []parser.ConditionOwnership
	gateOrders     []shared.SourceOrder
}

func chosenNameConsumer(reference CompiledReference, effect CompiledEffect) bool {
	if effect.Kind != EffectSearch || !effect.SearchSameNameAsChosenObject {
		return false
	}
	for _, candidate := range effect.References {
		if candidate.NodeID == reference.NodeID && candidate.Kind == ReferenceThatObject {
			return true
		}
	}
	return false
}

func bindChosenPermanentNameReference(reference *CompiledReference, effects []CompiledEffect) bool {
	owned := reference.ChoiceClauseID > 0 || slices.ContainsFunc(effects, func(effect CompiledEffect) bool {
		return chosenNameConsumer(*reference, effect)
	})
	if !owned {
		return false
	}
	reference.Binding = ReferenceBindingUnsupported
	if index, ok := chosenNameProducer(*reference, effects); ok {
		reference.Binding = ReferenceBindingPriorInstructionResult
		reference.PriorInstruction = index
	}
	return true
}

func chosenNameProducer(reference CompiledReference, effects []CompiledEffect) (int, bool) {
	if reference.ChoiceClauseID <= 0 || reference.Kind != ReferenceThatObject ||
		reference.Pronoun != ReferencePronounUnknown || reference.CardIdentity ||
		reference.LibraryCardObservation || reference.ReachedCardProducerClauseID != 0 ||
		reference.NodeID < 0 || reference.Order.End <= reference.Order.Start {
		return 0, false
	}
	index, choices := -1, 0
	for i, effect := range effects {
		if effect.ClauseID == reference.ChoiceClauseID {
			if index >= 0 {
				return 0, false
			}
			index = i
		}
		if effect.Kind == EffectChoosePermanent {
			choices++
		}
	}
	if index < 0 || choices != 1 {
		return 0, false
	}
	producer := effects[index]
	if !exactChosenNameProducer(producer, reference.SubjectNoun) ||
		producer.VerbOrder.End <= producer.VerbOrder.Start ||
		producer.VerbOrder.End > reference.Order.Start {
		return 0, false
	}
	for _, effect := range effects {
		if effect.VerbOrder.Start <= producer.VerbOrder.Start ||
			effect.VerbOrder.Start >= reference.Order.Start {
			continue
		}
		switch effect.Kind {
		case EffectBolster, EffectManifest, EffectManifestDread, EffectCloak,
			EffectPopulate, EffectChooseCreatureType:
			return 0, false
		default:
			continue
		}
	}
	return index, true
}

func exactChosenNameProducer(effect CompiledEffect, noun parser.ObjectNoun) bool {
	if effect.Kind != EffectChoosePermanent || !effect.Exact || effect.Optional || effect.Negated ||
		effect.DelayedTiming != 0 || effect.Duration != DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.CardSource != parser.EffectCardSourceNone || effect.FromZone != zone.None || effect.ToZone != zone.None ||
		len(effect.Targets) != 0 || len(effect.SubjectTargets) != 0 || len(effect.References) != 0 ||
		effect.Selector.Controller != ControllerYou ||
		(effect.Selector.Zone != zone.None && effect.Selector.Zone != zone.Battlefield) {
		return false
	}
	switch noun {
	case parser.ObjectNounPermanent:
		return targetSubjectDomain(CompiledTarget{Selector: effect.Selector}) == ReferenceSubjectPermanent
	case parser.ObjectNounLand:
		return effect.Selector.Kind == SelectorLand
	case parser.ObjectNounCreature:
		return effect.Selector.Kind == SelectorCreature
	case parser.ObjectNounArtifact:
		return effect.Selector.Kind == SelectorArtifact
	case parser.ObjectNounEnchantment:
		return effect.Selector.Kind == SelectorEnchantment
	case parser.ObjectNounPlaneswalker:
		return effect.Selector.Kind == SelectorPlaneswalker
	default:
		return false
	}
}

// A conditional choice may publish only inside the same governed scope as its
// consumer. A later independently checked condition cannot reuse a stale key.
func chosenNameGates(content *AbilityContent, producer int, reference CompiledReference) ([]parser.ConditionOwnership, bool) {
	var gates []parser.ConditionOwnership
	for _, condition := range content.Conditions {
		ownership := condition.Ownership
		if ownership.Scope == parser.ConditionScopeUnknown || ownership.Scope == parser.ConditionScopeUnsupported {
			return nil, false
		}
		if !slices.Contains(ownership.ClauseIDs, producer) {
			continue
		}
		if ownership.Scope != parser.ConditionScopeClause && ownership.Scope != parser.ConditionScopeGroup {
			return nil, false
		}
		for _, effect := range content.Effects {
			if chosenNameConsumer(reference, effect) && !slices.Contains(ownership.ClauseIDs, effect.ClauseID) &&
				!chosenNameFoldedPaymentScope(content, producer, condition) {
				return nil, false
			}
		}
		ownership.ClauseIDs = slices.Clone(ownership.ClauseIDs)
		ownership.ReferenceNodeIDs = slices.Clone(ownership.ReferenceNodeIDs)
		gates = append(gates, ownership)
	}
	return gates, true
}

func chosenNameFoldedPaymentScope(content *AbilityContent, producer int, condition CompiledCondition) bool {
	if len(content.Conditions) != 1 || condition.Kind != ConditionIf || condition.Negated ||
		condition.Predicate != ConditionPredicatePriorInstructionAccepted ||
		condition.Ownership.Scope != parser.ConditionScopeClause ||
		len(condition.Ownership.ClauseIDs) != 1 || condition.Ownership.ClauseIDs[0] != producer {
		return false
	}
	payments := 0
	for _, effect := range content.Effects {
		if effect.Payment.Form == parser.EffectPaymentFormUnknown {
			continue
		}
		payments++
		payment := effect.Payment
		if effect.ClauseID != producer || payment.Form != parser.EffectPaymentFormMayPayThenIfDo ||
			payment.Payer != parser.EffectPaymentPayerController ||
			payment.SuccessConditionNodeID != condition.NodeID {
			return false
		}
	}
	return payments == 1
}

func issueChosenPermanentNameSubject(reference *CompiledReference, source ReferenceSourceContext) {
	if source.scope == nil || source.scope.owner == nil ||
		reference.Binding != ReferenceBindingPriorInstructionResult {
		return
	}
	content := source.scope.owner
	if !slices.ContainsFunc(content.Effects, func(effect CompiledEffect) bool {
		return chosenNameConsumer(*reference, effect)
	}) {
		return
	}
	index, ok := chosenNameProducer(*reference, content.Effects)
	if !ok || reference.PriorInstruction != index ||
		reference.ProducerClauseID != 0 && reference.ProducerClauseID != reference.ChoiceClauseID {
		return
	}
	producer := content.Effects[index]
	gates, ok := chosenNameGates(content, producer.ClauseID, *reference)
	if !ok {
		return
	}
	reference.ProducerClauseID = producer.ClauseID
	reference.Subject = subjectProof(*reference, ReferenceSubjectPermanent, ReferenceLifetimeActualProduct)
	reference.Subject.scope = source.scope
	stampSubjectProducer(&reference.Subject, producer)
	reference.Subject.chosenName = &chosenPermanentNameProof{
		selector: cloneActualCardSelector(producer.Selector), referenceOrder: reference.Order,
		producerOrder: producer.VerbOrder, paymentOrder: producer.Payment.Order, gates: gates,
		paymentForm: producer.Payment.Form, paymentPayer: producer.Payment.Payer,
		paymentNodeID: producer.Payment.SuccessConditionNodeID,
		paymentMana:   slices.Clone(producer.Payment.ManaCost),
		paymentCost:   chosenNamePaymentCost(producer.Payment.AdditionalCost),
		gateOrders:    chosenNameGateOrders(content, producer.ClauseID),
	}
}

func chosenPermanentNameSubjectValid(reference CompiledReference) bool {
	proof := reference.Subject
	if proof.chosenName == nil || proof.scope == nil || proof.scope.owner == nil ||
		reference.ProducerClauseID != reference.ChoiceClauseID ||
		reference.Order != proof.chosenName.referenceOrder {
		return false
	}
	content := proof.scope.owner
	matches := 0
	for _, canonical := range content.References {
		if canonical.NodeID != reference.NodeID {
			continue
		}
		matches++
		if canonical.Subject.scope != proof.scope || canonical.Subject.chosenName != proof.chosenName ||
			canonical.ChoiceClauseID != reference.ChoiceClauseID || canonical.ProducerClauseID != reference.ProducerClauseID ||
			canonical.Order != reference.Order || canonical.Kind != reference.Kind || canonical.SubjectNoun != reference.SubjectNoun ||
			canonical.Pronoun != reference.Pronoun || canonical.Occurrence != reference.Occurrence ||
			canonical.Binding != reference.Binding || canonical.CardIdentity != reference.CardIdentity ||
			canonical.LibraryCardObservation != reference.LibraryCardObservation ||
			canonical.ReachedCardProducerClauseID != reference.ReachedCardProducerClauseID {
			return false
		}
	}
	if matches != 1 || !slices.ContainsFunc(content.Effects, func(effect CompiledEffect) bool {
		return chosenNameConsumer(reference, effect)
	}) {
		return false
	}
	index, ok := chosenNameProducer(reference, content.Effects)
	if !ok {
		return false
	}
	producer := content.Effects[index]
	gates, ok := chosenNameGates(content, producer.ClauseID, reference)
	return ok && producer.subjectScope == proof.scope && producer.VerbOrder == proof.chosenName.producerOrder &&
		producer.Payment.Order == proof.chosenName.paymentOrder &&
		producer.Payment.Form == proof.chosenName.paymentForm && producer.Payment.Payer == proof.chosenName.paymentPayer &&
		producer.Payment.SuccessConditionNodeID == proof.chosenName.paymentNodeID &&
		slices.Equal(producer.Payment.ManaCost, proof.chosenName.paymentMana) &&
		reflect.DeepEqual(chosenNamePaymentCost(producer.Payment.AdditionalCost), proof.chosenName.paymentCost) &&
		reflect.DeepEqual(producer.Selector, proof.chosenName.selector) &&
		reflect.DeepEqual(gates, proof.chosenName.gates) &&
		slices.Equal(chosenNameGateOrders(content, producer.ClauseID), proof.chosenName.gateOrders)
}

func chosenPermanentNameProducerMatches(reference CompiledReference, effect CompiledEffect) bool {
	return reference.SubjectSupported() && reference.Binding == ReferenceBindingPriorInstructionResult &&
		reference.Subject.scope == effect.subjectScope && reference.ChoiceClauseID == effect.ClauseID &&
		exactChosenNameProducer(effect, reference.SubjectNoun) &&
		effect.VerbOrder == reference.Subject.chosenName.producerOrder &&
		chosenNameProjectedPaymentMatches(reference, effect.Payment) &&
		reflect.DeepEqual(effect.Selector, reference.Subject.chosenName.selector)
}

func chosenNameProjectedPaymentMatches(reference CompiledReference, payment CompiledEffectPayment) bool {
	proof := reference.Subject.chosenName
	if payment.Order == proof.paymentOrder && payment.Form == proof.paymentForm &&
		payment.Payer == proof.paymentPayer && payment.SuccessConditionNodeID == proof.paymentNodeID &&
		slices.Equal(payment.ManaCost, proof.paymentMana) &&
		reflect.DeepEqual(chosenNamePaymentCost(payment.AdditionalCost), proof.paymentCost) {
		return true
	}
	if !reflect.DeepEqual(payment, CompiledEffectPayment{}) {
		return false
	}
	// The paid-body adapter externalizes this envelope into Pay and gates every
	// consequence. SubjectSupported still revalidates the original canonical
	// payment and scope; only that exact folded envelope may be absent here.
	content := reference.Subject.scope.owner
	return len(content.Conditions) == 1 &&
		chosenNameFoldedPaymentScope(content, reference.ChoiceClauseID, content.Conditions[0])
}

func chosenNamePaymentCost(original *CompiledCost) *CompiledCost {
	if original == nil {
		return nil
	}
	frozen := *original
	frozen.Span, frozen.Text = shared.Span{}, ""
	frozen.Components = slices.Clone(original.Components)
	for i := range frozen.Components {
		component := &frozen.Components[i]
		component.Span, component.Text = shared.Span{}, ""
		component.Symbol, component.Amount, component.Object = "", "", ""
		component.SubtypesAny = slices.Clone(component.SubtypesAny)
		if component.PaidSubject != nil {
			subject := *component.PaidSubject
			component.PaidSubject = &subject
		}
	}
	return &frozen
}

func chosenNameGateOrders(content *AbilityContent, producer int) []shared.SourceOrder {
	var orders []shared.SourceOrder
	for _, condition := range content.Conditions {
		if slices.Contains(condition.Ownership.ClauseIDs, producer) {
			orders = append(orders, condition.Order)
		}
	}
	return orders
}

func resolveChosenPermanentSearchNames(content *AbilityContent) {
	for i := range content.Effects {
		effect := &content.Effects[i]
		if effect.Kind != EffectSearch || !effect.SearchSameNameAsChosenObject {
			continue
		}
		effect.SearchNameFromChosenPermanent = false
		subjects := 0
		for _, reference := range effect.References {
			if reference.Kind != ReferenceThatObject {
				continue
			}
			subjects++
			index, ok := chosenNameProducer(reference, content.Effects)
			if ok && content.OwnsSubject(reference) && reference.SubjectProducerMatches(content.Effects[index]) {
				effect.SearchNameFromChosenPermanent = true
			}
		}
		if subjects != 1 || !effect.SearchNameFromChosenPermanent {
			effect.SearchNameFromChosenPermanent = false
			effect.UnsupportedDetail = "same-name search reference does not bind to a prior permanent choice"
		}
	}
}
