package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// ReferenceUse identifies a typed consumer's operation on its subject.
type ReferenceUse uint8

// Reference uses are proved independently of the adapter requesting them.
const (
	ReferenceUseMutation ReferenceUse = iota + 1
	ReferenceUseCharacteristic
	ReferenceUseAttribution
	ReferenceUseCardAction
	ReferenceUsePlayerProjection
	ReferenceUseStackDisposition
	ReferenceUsePolicy
)

// DamageAttribution separates a resolving source from an original permanent.
type DamageAttribution uint8

// Damage attribution is unsupported unless its exact source use is proved.
const (
	DamageAttributionUnsupported DamageAttribution = iota
	DamageAttributionResolvingSource
	DamageAttributionOriginalObject
)

// Uses are derived from the owning typed consumers, not selected by an Adapter.
func referenceUses(reference CompiledReference, content AbilityContent) uint8 {
	var uses uint8
	add := func(use ReferenceUse) {
		// Live mutation exists only for a permanent subject; other domains
		// keep their own operation permissions.
		if use == ReferenceUseMutation && reference.Subject.domain != ReferenceSubjectPermanent &&
			reference.Subject.domain != ReferenceSubjectTargetChoice {
			return
		}
		uses |= 1 << (use - 1)
	}
	if reference.Subject.domain == ReferenceSubjectPolicy {
		add(ReferenceUsePolicy)
		return uses
	}
	if reference.Subject.domain == ReferenceSubjectPlayer || reference.Subject.domain == ReferenceSubjectPlayerGroup {
		add(ReferenceUsePlayerProjection)
	}
	for _, condition := range content.Conditions {
		if condition.HasSubjectReference && condition.SubjectRefID == reference.NodeID {
			add(ReferenceUseCharacteristic)
		}
	}
	for i, effect := range content.Effects {
		// A compound "gets ... and deals damage" clause names its damage source
		// through the parser's typed prior-subject context.
		if effect.Kind == EffectDealDamage && effect.Context == parser.EffectContextPriorSubject &&
			len(effect.SubjectReferences) == 0 && i > 0 && len(content.Effects[i-1].SubjectReferences) == 1 &&
			content.Effects[i-1].SubjectReferences[0].NodeID == reference.NodeID {
			add(ReferenceUseAttribution)
		}
		found := false
		for _, candidate := range effect.References {
			if candidate.NodeID == reference.NodeID {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		if effect.Amount.ReferenceNodeID == reference.NodeID && damageAmountObservesCharacteristic(effect.Amount.DynamicKind) {
			add(ReferenceUseCharacteristic)
		}
		switch effect.Kind {
		case EffectDealDamage:
			if len(effect.References) > 0 && effect.References[0].NodeID == reference.NodeID {
				add(ReferenceUseAttribution)
			} else {
				add(ReferenceUsePlayerProjection)
			}
		case EffectPut:
			if effect.ToZone == 0 {
				add(ReferenceUseMutation)
				break
			}
			add(ReferenceUseCardAction)
		case EffectReturn, EffectExile, EffectCast, EffectShuffle:
			add(ReferenceUseCardAction)
			if reference.Subject.lifetime == ReferenceLifetimeStackOccurrence {
				add(ReferenceUseStackDisposition)
			}
		case EffectCopyStackObject:
			add(ReferenceUsePolicy)
		case EffectEnterTapped:
			if effect.EntersWithCounters && reference.Binding == ReferenceBindingEventStackObject {
				add(ReferenceUsePolicy)
			} else {
				add(ReferenceUseMutation)
			}
		case EffectCreate:
			if effect.TokenCopyOfReference || effect.TokenCopyOfSource || effect.TokenCopyOfTarget {
				add(ReferenceUseCharacteristic)
			} else {
				add(ReferenceUsePolicy)
			}
		case EffectGain, EffectLose:
			if effect.LifeObject {
				add(ReferenceUsePlayerProjection)
			} else {
				add(ReferenceUseMutation)
			}
		case EffectDraw, EffectDiscard:
			add(ReferenceUsePlayerProjection)
		default:
			add(ReferenceUseMutation)
		}
	}
	return uses
}

func damageAmountObservesCharacteristic(kind DynamicAmountKind) bool {
	switch kind {
	case DynamicAmountSourcePower, DynamicAmountSourceToughness, DynamicAmountSourceManaValue:
		return true
	default:
		return false
	}
}

// SupportsUse checks the owned subject's permission for the requested operation.
func (reference CompiledReference) SupportsUse(use ReferenceUse) bool {
	return reference.SubjectSupported() && use >= ReferenceUseMutation &&
		use <= ReferenceUsePolicy && reference.Subject.uses&(1<<(use-1)) != 0
}

// ObjectProjectionDomain keeps the original object behind an owner/controller
// projection distinct from the resulting player. It grants no mutation subject.
func (reference CompiledReference) ObjectProjectionDomain() ReferenceSubjectDomain {
	if !reference.SubjectSupported() {
		return ReferenceSubjectUnknown
	}
	if reference.Subject.projectedDomain != ReferenceSubjectUnknown && reference.SupportsUse(ReferenceUsePlayerProjection) {
		return reference.Subject.projectedDomain
	}
	return reference.Subject.domain
}

// DamageAttribution returns the proven damage-source role.
func (reference CompiledReference) DamageAttribution() DamageAttribution {
	if !reference.SupportsUse(ReferenceUseAttribution) {
		return DamageAttributionUnsupported
	}

	if reference.Binding == ReferenceBindingSource &&
		(reference.Subject.domain == ReferenceSubjectCard ||
			reference.Subject.domain == ReferenceSubjectStackObject) {
		return DamageAttributionResolvingSource
	}
	if reference.Subject.domain == ReferenceSubjectPermanent {
		return DamageAttributionOriginalObject
	}
	return DamageAttributionUnsupported
}

// ExcludesPriorPermanentTargets proves an "another" declaration excludes its earlier
// declared permanent targets, not the resolving spell's nonexistent permanent.
func (target CompiledTarget) ExcludesPriorPermanentTargets() bool {
	if target.subjectScope == nil || target.subjectScope.owner == nil || !target.Exact ||
		!target.Selector.Another || target.Selector.Other ||
		target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
		targetSubjectDomain(target) != ReferenceSubjectPermanent {
		return false
	}
	owner := target.subjectScope.owner
	if owner.Source.kind != AbilitySpell {
		return false
	}
	found, prior := 0, 0
	for _, candidate := range owner.Targets {
		if candidate.Order == target.Order {
			found++
			continue
		}
		if found == 0 {
			if !candidate.Exact || candidate.Cardinality.Min != 1 || candidate.Cardinality.Max != 1 ||
				targetSubjectDomain(candidate) != ReferenceSubjectPermanent {
				return false
			}
			prior++
		}
	}
	return found == 1 && prior > 0
}

// EnclosingSpellTargetOccurrence returns the exact target shared by split bodies.
func (reference CompiledReference) EnclosingSpellTargetOccurrence() (int, bool) {
	if !reference.SubjectSupported() || reference.Binding != ReferenceBindingTarget ||
		reference.Subject.domain != ReferenceSubjectPermanent || reference.Subject.targetOccurrence != 0 {
		return 0, false
	}
	owner := reference.Subject.scope.owner
	if owner == nil || owner.Source.kind != AbilitySpell || len(owner.Targets) != 0 ||
		len(owner.Modes) != 0 || owner.Source.enclosingSpellTarget == nil ||
		reference.Subject.targetOrder != owner.Source.enclosingSpellTarget.Order {
		return 0, false
	}
	return 0, true
}
