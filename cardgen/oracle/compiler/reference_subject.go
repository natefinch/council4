package compiler

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// ReferenceSubjectDomain identifies the kind of subject an owned proof permits.
type ReferenceSubjectDomain uint8

// Subject domains keep permanent, card, stack, player, and policy uses distinct.
const (
	ReferenceSubjectUnknown ReferenceSubjectDomain = iota
	ReferenceSubjectPermanent
	ReferenceSubjectCard
	ReferenceSubjectStackObject
	ReferenceSubjectPlayer
	ReferenceSubjectPlayerGroup
	ReferenceSubjectPolicy
	ReferenceSubjectTargetChoice
)

// ReferenceSubjectLifetime identifies the incarnation or product a proof binds.
type ReferenceSubjectLifetime uint8

// Subject lifetimes distinguish original objects from later cards and products.
const (
	ReferenceLifetimeUnknown ReferenceSubjectLifetime = iota
	ReferenceLifetimeOriginalObject
	ReferenceLifetimeCardIncarnation
	ReferenceLifetimeActualProduct
	ReferenceLifetimeStackOccurrence
	ReferenceLifetimePlayer
	ReferenceLifetimePolicy
)

// ReferenceSubjectProof is issued by binding, never inferred by a runtime adapter.
// Effect indices may change; parser producer ClauseIDs and reference identities may not.
type ReferenceSubjectProof struct {
	scope               *referenceSubjectScope
	domain              ReferenceSubjectDomain
	lifetime            ReferenceSubjectLifetime
	binding             ReferenceBinding
	kind                ReferenceKind
	pronoun             ReferencePronounKind
	libraryObservation  bool
	producerKind        EffectKind
	producerFromZone    zone.Type
	producerToZone      zone.Type
	producerCardSource  parser.EffectCardSourceKind
	producerExact       bool
	producerNegated     bool
	nodeID              int
	clauseID            int
	reachedClauseID     int
	reached             *CompiledReference
	occurrence          int
	cardIdentity        bool
	noun                parser.ObjectNoun
	targetProduct       bool
	enteredProduct      bool
	producerTarget      int
	producerTargetOrder shared.SourceOrder
	targetOrder         shared.SourceOrder
	targetSelector      SelectorKind
	targetZone          zone.Type
	targetMin           int
	targetMax           int
	targetOccurrence    int
	uses                uint8
	projectedDomain     ReferenceSubjectDomain
}

// SubjectSupported checks that the issued proof still matches the reference.
func (reference CompiledReference) SubjectSupported() bool {
	proof := reference.Subject
	return proof.scope != nil && proof.domain != ReferenceSubjectUnknown && proof.lifetime != ReferenceLifetimeUnknown &&
		proof.binding == reference.Binding && proof.clauseID == reference.ProducerClauseID &&
		proof.reachedClauseID == reference.ReachedCardProducerClauseID &&
		proof.kind == reference.Kind && proof.pronoun == reference.Pronoun &&
		proof.libraryObservation == reference.LibraryCardObservation &&
		proof.occurrence == reference.Occurrence && proof.cardIdentity == reference.CardIdentity &&
		proof.noun == reference.SubjectNoun && proof.nodeID == reference.NodeID
}

// SubjectDomain returns the proven domain, or Unknown for an invalid proof.
func (reference CompiledReference) SubjectDomain() ReferenceSubjectDomain {
	if !reference.SubjectSupported() {
		return ReferenceSubjectUnknown
	}
	return reference.Subject.domain
}

// SubjectLifetime returns the proven lifetime, or Unknown for an invalid proof.
func (reference CompiledReference) SubjectLifetime() ReferenceSubjectLifetime {
	if !reference.SubjectSupported() {
		return ReferenceLifetimeUnknown
	}
	return reference.Subject.lifetime
}

// ProducerTargetOccurrence returns the exact target producing this subject.
func (reference CompiledReference) ProducerTargetOccurrence() (int, bool) {
	if !reference.SubjectSupported() || !reference.Subject.targetProduct {
		return 0, false
	}
	return reference.Subject.producerTarget, true
}

// ProducerTargetMatches checks the producing target's identity and selection.
func (reference CompiledReference) ProducerTargetMatches(target CompiledTarget) bool {
	return reference.SubjectSupported() && reference.Subject.targetProduct &&
		reference.Subject.scope == target.subjectScope &&
		reference.Subject.producerTargetOrder == target.Order &&
		reference.Subject.targetSelector == target.Selector.Kind &&
		reference.Subject.targetZone == target.Selector.Zone &&
		reference.Subject.targetMin == target.Cardinality.Min &&
		reference.Subject.targetMax == target.Cardinality.Max
}

// ReferenceSourceContext carries the compiler-owned source shell and scope.
type ReferenceSourceContext struct {
	scope                *referenceSubjectScope
	kind                 AbilityKind
	zone                 zone.Type
	event                TriggerEvent
	self                 bool
	known                bool
	cardCostAntecedent   bool
	eventTypes           []types.Card
	enclosingSpellTarget *CompiledTarget
}

func referenceSourceContext(kind AbilityKind, sourceZone zone.Type, trigger *CompiledTrigger) ReferenceSourceContext {
	source := ReferenceSourceContext{kind: kind, zone: sourceZone, known: kind != AbilityUnknown}
	if trigger != nil {
		source.event, source.self = trigger.Pattern.Event, trigger.Pattern.Source == TriggerSourceSelf
		source.eventTypes = append(slices.Clone(trigger.Pattern.SubjectSelection.RequiredTypes),
			trigger.Pattern.CardSelection.RequiredTypes...)
	}
	return source
}

// OriginalObjectSubject proves a battlefield source's original permanent.
func (source ReferenceSourceContext) OriginalObjectSubject() (CompiledReference, bool) {
	if source.scope == nil || !source.known || source.kind == AbilitySpell || source.zone != zone.Battlefield {
		return CompiledReference{}, false
	}
	return issuedIntrinsicSubject(ReferenceBindingSource, ReferenceSubjectPermanent, ReferenceLifetimeOriginalObject, source.scope), true
}

// AttachedObjectSubject proves an attached object from a battlefield source.
func (source ReferenceSourceContext) AttachedObjectSubject(effect CompiledEffect) (CompiledReference, bool) {
	if !effect.SubjectSourceAttached {
		return CompiledReference{}, false
	}
	if _, ok := source.OriginalObjectSubject(); !ok {
		return CompiledReference{}, false
	}
	return issuedIntrinsicSubject(ReferenceBindingSourceAttached, ReferenceSubjectPermanent, ReferenceLifetimeOriginalObject, source.scope), true
}

// CardSubject proves an exact off-battlefield source or self-death event card.
func (source ReferenceSourceContext) CardSubject() (CompiledReference, zone.Type, bool) {
	if source.scope == nil || !source.known || source.kind == AbilitySpell {
		return CompiledReference{}, zone.None, false
	}
	if source.self && (source.event == TriggerEventPermanentDied || source.event == TriggerEventPermanentSacrificed) {
		return issuedIntrinsicSubject(ReferenceBindingEventCard, ReferenceSubjectCard, ReferenceLifetimeCardIncarnation, source.scope), zone.Graveyard, true
	}
	if source.zone == zone.None || source.zone == zone.Battlefield {
		return CompiledReference{}, zone.None, false
	}
	return issuedIntrinsicSubject(ReferenceBindingSource, ReferenceSubjectCard, ReferenceLifetimeCardIncarnation, source.scope), source.zone, true
}

func issuedIntrinsicSubject(binding ReferenceBinding, domain ReferenceSubjectDomain, lifetime ReferenceSubjectLifetime, scope *referenceSubjectScope) CompiledReference {
	reference := CompiledReference{Binding: binding}
	reference.Subject = subjectProof(reference, domain, lifetime)
	reference.Subject.scope = scope
	return reference
}

func subjectProof(reference CompiledReference, domain ReferenceSubjectDomain, lifetime ReferenceSubjectLifetime) ReferenceSubjectProof {
	return ReferenceSubjectProof{
		domain: domain, lifetime: lifetime, binding: reference.Binding, nodeID: reference.NodeID,
		kind: reference.Kind, pronoun: reference.Pronoun, libraryObservation: reference.LibraryCardObservation,
		clauseID: reference.ProducerClauseID, occurrence: reference.Occurrence,
		reachedClauseID: reference.ReachedCardProducerClauseID,
		cardIdentity:    reference.CardIdentity, noun: reference.SubjectNoun,
	}
}

func finalizeReferenceSubjects(content *AbilityContent, source ReferenceSourceContext) {
	scope := &referenceSubjectScope{owner: content}
	content.subjectScope, source.scope = scope, scope
	for i := range content.Targets {
		content.Targets[i].subjectScope = scope
	}
	for i := range content.Effects {
		content.Effects[i].subjectScope = scope
		for j := range content.Effects[i].Targets {
			content.Effects[i].Targets[j].subjectScope = scope
		}
		for j := range content.Effects[i].SubjectTargets {
			content.Effects[i].SubjectTargets[j].subjectScope = scope
		}
	}
	content.Source = source
	bindingTargets := content.Targets
	if source.kind == AbilitySpell && len(bindingTargets) == 0 && len(content.Modes) == 0 &&
		source.enclosingSpellTarget != nil {
		bindingTargets = []CompiledTarget{*source.enclosingSpellTarget}
		bindingTargets[0].subjectScope = scope
	}
	reconcileConditionSubjects(content)
	seen := make(map[int]int)
	for _, reference := range content.References {
		seen[reference.NodeID]++
	}
	for i := range content.References {
		reference := &content.References[i]
		if seen[reference.NodeID] != 1 {
			reference.Binding = ReferenceBindingUnsupported
			reference.Subject = ReferenceSubjectProof{}
			continue
		}
		issueReferenceSubject(reference, bindingTargets, content.Effects, source)
		reference.Subject.uses = referenceUses(*reference, *content)
		applyEffectReferenceBindings(content.Effects, []CompiledReference{*reference})
	}
	applyEffectReferenceBindings(content.Effects, content.References)
	for i := range content.Conditions {
		condition := &content.Conditions[i]
		if condition.ObjectReference != nil {
			reference := *condition.ObjectReference
			found := slices.IndexFunc(content.References, func(candidate CompiledReference) bool {
				return condition.HasSubjectReference && condition.SubjectRefID == reference.NodeID &&
					candidate.NodeID == reference.NodeID
			})
			switch {
			case found >= 0:
				reference = content.References[found]
			case !condition.HasSubjectReference && reference.Kind == ReferenceUnknown:
				issueReferenceSubject(&reference, bindingTargets, content.Effects, source)
			default:
				reference.Subject = ReferenceSubjectProof{}
			}
			condition.ObjectReference = &reference
			condition.ObjectBinding = reference.Binding
		} else if !condition.HasSubjectReference {
			reference := CompiledReference{Binding: condition.ObjectBinding}
			issueReferenceSubject(&reference, bindingTargets, content.Effects, source)
			if reference.SubjectSupported() {
				condition.ObjectReference = &reference
			}
		}
		if reference := condition.ObjectReference; reference != nil && reference.SubjectSupported() &&
			reference.Binding == ReferenceBindingTarget && condition.ObjectTarget == nil &&
			reference.Occurrence >= 0 && reference.Occurrence < len(bindingTargets) {
			target := bindingTargets[reference.Occurrence]
			condition.ObjectTarget = &target
		}
	}
	for i := range content.Modes {
		for j := range content.Modes[i].Content.References {
			reference := &content.Modes[i].Content.References[j]
			if reference.Binding != ReferenceBindingTarget || reference.Occurrence < 0 ||
				reference.Occurrence >= len(content.Targets) {
				continue
			}
			target := content.Targets[reference.Occurrence]
			local := slices.IndexFunc(content.Modes[i].Content.Targets, func(candidate CompiledTarget) bool {
				return candidate.Order == target.Order
			})
			if local >= 0 {
				reference.Occurrence = local
			}
		}
		finalizeReferenceSubjects(&content.Modes[i].Content, source)
	}
}

// Condition binding may refine an event/target antecedent after the initial
// reference pass. Commit that decision once before issuing proof to every copy.
func reconcileConditionSubjects(content *AbilityContent) {
	for i := range content.References {
		reference := &content.References[i]
		var resolved *CompiledReference
		for _, condition := range content.Conditions {
			if !condition.HasSubjectReference || condition.SubjectRefID != reference.NodeID ||
				condition.ObjectReference == nil {
				continue
			}
			subject := condition.ObjectReference
			if subject.NodeID != reference.NodeID || subject.Kind != reference.Kind ||
				subject.Pronoun != reference.Pronoun || subject.CardIdentity != reference.CardIdentity ||
				subject.SubjectNoun != reference.SubjectNoun ||
				subject.LibraryCardObservation != reference.LibraryCardObservation ||
				subject.ReachedCardProducerClauseID != reference.ReachedCardProducerClauseID ||
				reference.ProducerClauseID > 0 && subject.ProducerClauseID != reference.ProducerClauseID ||
				resolved != nil && (resolved.Binding != subject.Binding ||
					resolved.Occurrence != subject.Occurrence || resolved.PriorInstruction != subject.PriorInstruction) {
				reference.Binding = ReferenceBindingUnsupported
				resolved = nil
				break
			}
			resolved = subject
		}
		if resolved != nil {
			*reference = *resolved
		}
	}
}

func issueReferenceSubject(reference *CompiledReference, targets []CompiledTarget, effects []CompiledEffect, source ReferenceSourceContext) {
	reference.Subject = ReferenceSubjectProof{}
	if source.kind == AbilitySpell && source.scope != nil && source.scope.owner != nil &&
		referenceIsOwnedSpellDamageSubject(*reference, *source.scope.owner) {
		reference.Binding = ReferenceBindingSource
	}
	if referenceIsOwnedCastCostPolicy(*reference, effects) ||
		referenceIsCorrelatedPlayerPolicy(*reference, effects) ||
		source.scope != nil && source.scope.owner != nil &&
			referenceIsOwnedPaymentDecision(*reference, *source.scope.owner, source) {
		reference.Subject = subjectProof(*reference, ReferenceSubjectPolicy, ReferenceLifetimePolicy)
		reference.Subject.scope = source.scope
		return
	}
	var domain ReferenceSubjectDomain
	var lifetime ReferenceSubjectLifetime
	projectedDomain := ReferenceSubjectUnknown
	switch reference.Binding {
	case ReferenceBindingSource:
		if !source.known {
			return
		}
		// A recurring trigger cannot start with its source in the graveyard
		// when discovery only proves a battlefield source.
		if source.kind == AbilityTriggered && source.event == TriggerEventBeginningOfStep &&
			source.zone == zone.Battlefield && len(effects) != 0 &&
			effects[0].FromZone == zone.Graveyard && effectOwnsReference(effects[0], *reference) {
			return
		}
		switch {
		case referenceIsOwnedCountPolicy(*reference, effects):
			domain, lifetime = ReferenceSubjectPolicy, ReferenceLifetimePolicy
		case referenceIsOwnedGroupPlayer(*reference, effects):
			domain, lifetime = ReferenceSubjectPlayerGroup, ReferenceLifetimePlayer
		case source.self && source.event == TriggerEventSpellCast &&
			reference.SubjectNoun == parser.ObjectNounSpell:
			domain, lifetime = ReferenceSubjectStackObject, ReferenceLifetimeStackOccurrence
		case reference.CardIdentity && source.self &&
			(source.event == TriggerEventPermanentDied || source.event == TriggerEventPermanentSacrificed):
			reference.Binding = ReferenceBindingEventCard
			domain, lifetime = ReferenceSubjectCard, ReferenceLifetimeCardIncarnation
		case source.kind == AbilitySpell:
			if reference.Kind != ReferenceSelfName && reference.Kind != ReferenceThisObject &&
				!sourceCardAntecedent(*reference, effects) && !referenceIsResolvingDamageSubject(*reference, effects) {
				return
			}
			if reference.SubjectNoun != parser.ObjectNounUnknown && reference.SubjectNoun != parser.ObjectNounSpell &&
				reference.SubjectNoun != parser.ObjectNounCard {
				return
			}
			domain, lifetime = ReferenceSubjectCard, ReferenceLifetimeStackOccurrence
		case source.zone != zone.Battlefield:
			if source.zone == zone.None || !reference.CardIdentity &&
				!sourceCardAntecedent(*reference, effects) && !source.cardCostAntecedent &&
				!recurringSourceCardReference(source, *reference, effects) {
				return
			}
			domain, lifetime = ReferenceSubjectCard, ReferenceLifetimeCardIncarnation
		case reference.CardIdentity:
			domain, lifetime = ReferenceSubjectCard, ReferenceLifetimeCardIncarnation
		default:
			domain, lifetime = ReferenceSubjectPermanent, ReferenceLifetimeOriginalObject
		}
	case ReferenceBindingEventPermanent:
		switch {
		case reference.SubjectNoun == parser.ObjectNounPlayer && referenceIsEventPlayerProjection(*reference, source, effects):
			projectedDomain = ReferenceSubjectPermanent
			domain, lifetime = ReferenceSubjectPlayer, ReferenceLifetimePlayer
		case reference.CardIdentity:
			reference.Binding = ReferenceBindingEventCard
			domain, lifetime = ReferenceSubjectCard, ReferenceLifetimeCardIncarnation
		default:
			domain, lifetime = ReferenceSubjectPermanent, ReferenceLifetimeOriginalObject
		}
	case ReferenceBindingSourceAttached, ReferenceBindingEventRelatedPermanent:
		domain, lifetime = ReferenceSubjectPermanent, ReferenceLifetimeOriginalObject
	case ReferenceBindingEventCard:
		domain, lifetime = ReferenceSubjectCard, ReferenceLifetimeCardIncarnation
	case ReferenceBindingEventStackObject:
		domain, lifetime = ReferenceSubjectStackObject, ReferenceLifetimeStackOccurrence
	case ReferenceBindingEventPlayer, ReferenceBindingLibraryOwner:
		domain, lifetime = ReferenceSubjectPlayer, ReferenceLifetimePlayer
	case ReferenceBindingCreatedToken:
		creates := 0
		for _, effect := range effects {
			if effect.Kind == EffectCreate && !effect.Negated &&
				(effect.Exact || effect.TokenCopyOfTarget || effect.TokenCopyOfReference || effect.TokenCopyOfSource) {
				creates++
			}
		}
		if creates != 1 {
			return
		}
		domain, lifetime = ReferenceSubjectPermanent, ReferenceLifetimeActualProduct
	case ReferenceBindingTarget:
		if reference.Occurrence < 0 || reference.Occurrence >= len(targets) {
			return
		}
		target := targets[reference.Occurrence]
		reference.Subject.targetOrder = target.Order
		domain = targetSubjectDomain(target)
		if domain == ReferenceSubjectUnknown || domain == ReferenceSubjectCard && target.Selector.Zone == zone.None {
			domain = ownedTargetSubjectDomain(target, effects)
		}
		switch domain {
		case ReferenceSubjectCard:
			lifetime = ReferenceLifetimeCardIncarnation
		case ReferenceSubjectStackObject:
			lifetime = ReferenceLifetimeStackOccurrence
		case ReferenceSubjectPlayer:
			lifetime = ReferenceLifetimePlayer
		case ReferenceSubjectPermanent, ReferenceSubjectTargetChoice:
			lifetime = ReferenceLifetimeOriginalObject
		default:
			return
		}
		if reference.CardIdentity && domain == ReferenceSubjectPermanent {
			domain, lifetime = ReferenceSubjectCard, ReferenceLifetimeCardIncarnation
		}
		if reference.SubjectNoun == parser.ObjectNounPlayer &&
			(domain == ReferenceSubjectPermanent || domain == ReferenceSubjectCard || domain == ReferenceSubjectStackObject) &&
			referenceIsTargetPlayerProjection(*reference, target, effects) {
			projectedDomain = domain
			domain, lifetime = ReferenceSubjectPlayer, ReferenceLifetimePlayer
		}
	case ReferenceBindingPriorInstructionResult:
		index := reference.PriorInstruction
		if index < 0 || index >= len(effects) {
			return
		}
		producer := effects[index]
		matches := 0
		for _, effect := range effects {
			if effect.ClauseID == producer.ClauseID {
				matches++
			}
		}
		if producer.ClauseID <= 0 || reference.ProducerClauseID > 0 && reference.ProducerClauseID != producer.ClauseID ||
			matches != 1 {
			return
		}
		reference.ProducerClauseID = producer.ClauseID
		switch {
		case reference.LibraryCardObservation:
			if !exactLibraryCardReferenceProducer(producer, effects) {
				return
			}
			domain = ReferenceSubjectCard
		case singularEnteredSubjectProducer(producer, effects):
			domain = ReferenceSubjectPermanent
			if reference.CardIdentity {
				domain = ReferenceSubjectCard
			}
		case searchedPermanentProduct(*reference, effects):
			domain = ReferenceSubjectPermanent
		default:
			switch producer.Kind {
			case EffectCreate, EffectBolster, EffectChoosePermanent, EffectSacrifice, EffectDestroy:
				domain = ReferenceSubjectPermanent
			case EffectManifest, EffectManifestDread:
				if !producer.Exact || producer.Negated {
					return
				}
				domain = ReferenceSubjectPermanent
			case EffectExile, EffectSearch, EffectDig, EffectMill, EffectReveal:
				if singularReferenceToPluralTopCards(*reference, producer) {
					return
				}
				domain = ReferenceSubjectCard
			default:
				return
			}
		}
		lifetime = ReferenceLifetimeActualProduct
	default:
		return
	}
	if !subjectNounCompatible(reference.SubjectNoun, domain) &&
		!referenceIsCardCharacteristicUse(*reference, domain, effects) &&
		!referenceIsCastEntryPolicy(*reference, domain, source, effects) {
		return
	}
	targetOrder := reference.Subject.targetOrder
	reference.Subject = subjectProof(*reference, domain, lifetime)
	reference.Subject.scope = source.scope
	reference.Subject.projectedDomain = projectedDomain
	reference.Subject.targetOrder = targetOrder
	if reference.Binding == ReferenceBindingPriorInstructionResult {
		stampSubjectProducer(&reference.Subject, effects[reference.PriorInstruction])
		reference.Subject.enteredProduct = singularEnteredSubjectProducer(effects[reference.PriorInstruction], effects)
	}
	if reference.Binding == ReferenceBindingPriorInstructionResult &&
		singularEnteredSubjectProducer(effects[reference.PriorInstruction], effects) &&
		len(effects[reference.PriorInstruction].Targets) == 1 {
		target := effects[reference.PriorInstruction].Targets[0]
		matches, index := 0, -1
		for i, candidate := range targets {
			if candidate.Order == target.Order {
				matches++
				index = i
			}
		}
		if matches == 1 {
			reference.Subject.targetProduct = true
			reference.Subject.producerTarget = index
			reference.Subject.producerTargetOrder = targets[index].Order
			stampSubjectTarget(&reference.Subject, targets[index])
		}
	}
	if reference.Binding == ReferenceBindingTarget {
		stampSubjectTarget(&reference.Subject, targets[reference.Occurrence])
		reference.Subject.targetOccurrence = reference.Occurrence
	}
	if reference.ReachedCardProducerClauseID > 0 && !issueCardLineageSubject(reference, effects) {
		reference.Subject = ReferenceSubjectProof{}
	}
}

// LocalizeSubjectOccurrence changes only the adapter slot, preserving the
// compiler-proven target occurrence rather than rebinding its noun.
func (reference CompiledReference) LocalizeSubjectOccurrence(target CompiledTarget, occurrence int) (CompiledReference, bool) {
	if !reference.SubjectSupported() || reference.Binding != ReferenceBindingTarget ||
		occurrence < 0 || reference.Subject.scope != target.subjectScope ||
		reference.Subject.targetOrder != target.Order ||
		reference.Subject.targetSelector != target.Selector.Kind ||
		reference.Subject.targetZone != target.Selector.Zone ||
		reference.Subject.targetMin != target.Cardinality.Min ||
		reference.Subject.targetMax != target.Cardinality.Max ||
		target.Order == (shared.SourceOrder{}) {
		return CompiledReference{}, false
	}

	reference.Occurrence = occurrence
	reference.Subject.occurrence = occurrence
	return reference, true
}

func stampSubjectTarget(proof *ReferenceSubjectProof, target CompiledTarget) {
	proof.targetSelector, proof.targetZone = target.Selector.Kind, target.Selector.Zone
	proof.targetMin, proof.targetMax = target.Cardinality.Min, target.Cardinality.Max
}

func subjectNounCompatible(noun parser.ObjectNoun, domain ReferenceSubjectDomain) bool {
	if noun == parser.ObjectNounUnknown {
		return true
	}
	if noun == parser.ObjectNounCard {
		return domain == ReferenceSubjectCard
	}
	if noun == parser.ObjectNounPlayer {
		return domain == ReferenceSubjectPlayer || domain == ReferenceSubjectTargetChoice
	}
	if noun == parser.ObjectNounSpell {
		return domain == ReferenceSubjectStackObject ||
			domain == ReferenceSubjectCard
	}

	return domain == ReferenceSubjectPermanent || domain == ReferenceSubjectTargetChoice
}

func targetSubjectDomain(target CompiledTarget) ReferenceSubjectDomain {
	if target.Selector.Zone != zone.None && target.Selector.Zone != zone.Battlefield {
		return ReferenceSubjectCard
	}
	switch target.Selector.Kind {
	case SelectorUnknown:
		if len(target.Selector.SubtypesAny()) != 0 {
			return ReferenceSubjectPermanent
		}
		if len(target.Selector.Alternatives) != 0 {
			domain := ReferenceSubjectPermanent
			for _, alternative := range target.Selector.Alternatives {
				candidate := targetSubjectDomain(CompiledTarget{Selector: alternative})
				if candidate != domain {
					return ReferenceSubjectUnknown
				}
			}
			return domain
		}
		if len(target.Selector.RequiredTypesAny()) != 0 {
			permanent := true
			for _, kind := range target.Selector.RequiredTypesAny() {
				permanent = permanent && kind.IsPermanent()
			}
			if permanent {
				return ReferenceSubjectPermanent
			}
		}
		return ReferenceSubjectUnknown
	case SelectorAny:
		return ReferenceSubjectTargetChoice
	case SelectorPlayer, SelectorOpponent:
		return ReferenceSubjectPlayer
	case SelectorSpell, SelectorActivatedAbility, SelectorTriggeredAbility,
		SelectorActivatedOrTriggeredAbility, SelectorSpellActivatedOrTriggeredAbility,
		SelectorTriggeredAbilityOrSpell:
		return ReferenceSubjectStackObject
	case SelectorCard:
		return ReferenceSubjectCard
	default:
		return ReferenceSubjectPermanent
	}
}

// SubjectProducerMatches checks a subject's exact producing clause and scope.
func (reference CompiledReference) SubjectProducerMatches(effect CompiledEffect) bool {
	return reference.SubjectSupported() &&
		reference.Binding == ReferenceBindingPriorInstructionResult &&
		reference.ProducerClauseID > 0 &&
		reference.Subject.scope == effect.subjectScope &&
		reference.ProducerClauseID == effect.ClauseID &&
		reference.Subject.producerKind == effect.Kind &&
		reference.Subject.producerFromZone == effect.FromZone &&
		reference.Subject.producerToZone == effect.ToZone &&
		reference.Subject.producerCardSource == effect.CardSource &&
		reference.Subject.producerExact == effect.Exact &&
		reference.Subject.producerNegated == effect.Negated
}

func stampSubjectProducer(proof *ReferenceSubjectProof, effect CompiledEffect) {
	proof.producerKind, proof.producerFromZone, proof.producerToZone = effect.Kind, effect.FromZone, effect.ToZone
	proof.producerCardSource = effect.CardSource
	proof.producerExact, proof.producerNegated = effect.Exact, effect.Negated
}

func sourceCardAntecedent(reference CompiledReference, effects []CompiledEffect) bool {
	if reference.Kind != ReferencePronoun || reference.SubjectNoun != parser.ObjectNounUnknown ||
		(reference.Pronoun != ReferencePronounIt && reference.Pronoun != ReferencePronounIts) {
		return false
	}
	for _, effect := range effects {
		for _, candidate := range effect.References {
			if candidate.Binding == ReferenceBindingSource &&
				(candidate.CardIdentity || candidate.Kind == ReferenceSelfName || candidate.Kind == ReferenceThisObject) &&
				candidate.Order.End <= reference.Order.Start {
				return true
			}
		}
	}
	return false
}

// EnteredSubjectSupported reports whether the proof names an actual entered object.
func (reference CompiledReference) EnteredSubjectSupported() bool {
	return reference.SubjectSupported() && reference.Subject.enteredProduct &&
		reference.Subject.lifetime == ReferenceLifetimeActualProduct
}

type referenceSubjectScope struct {
	owner *AbilityContent
}

// OwnsSubject checks that the reference's proof belongs to this exact body.
func (content AbilityContent) OwnsSubject(reference CompiledReference) bool {
	return reference.SubjectSupported() && content.subjectScope != nil &&
		reference.Subject.scope == content.subjectScope
}
